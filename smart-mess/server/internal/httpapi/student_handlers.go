package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"smartmess/internal/config"
	"smartmess/internal/rules"
	"smartmess/internal/store"
)

// mealWindow holds the absolute start and end of a meal window.
type mealWindow struct {
	MealType string
	Start    time.Time
	End      time.Time
}

// findMealWindow locates the window for a meal's schedule on its service date.
func findMealWindow(scheds []rules.Schedule, mealType string, serviceDate time.Time) *mealWindow {
	dayKind := rules.DayKind(serviceDate)
	for _, s := range scheds {
		if s.MealType == mealType && s.DayKind == dayKind {
			start, end := rules.MealWindow(s, serviceDate)
			return &mealWindow{MealType: s.MealType, Start: start, End: end}
		}
	}
	return nil
}

// currentWindow returns the current meal's window, or nil when none is open.
func currentWindow(scheds []rules.Schedule, now time.Time) *mealWindow {
	s := rules.CurrentMeal(scheds, now)
	if s == nil {
		return nil
	}
	start, end := rules.MealWindow(*s, now)
	return &mealWindow{MealType: s.MealType, Start: start, End: end}
}

// GET /api/messes — all messes of the user's group with crowd state (A1).
func (a *App) listMesses(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	if !u.HostelGroup.Valid {
		writeErr(w, http.StatusForbidden, ErrForbidden, "Choose your hostel group first.")
		return
	}
	now := a.Clock.Now()
	messes, err := a.Q.ListMessesByGroup(r.Context(), u.HostelGroup.HostelGroup)
	if err != nil {
		serverError(w, err)
		return
	}
	schedRows, err := a.Q.GetMealSchedules(r.Context(), u.HostelGroup.HostelGroup)
	if err != nil {
		serverError(w, err)
		return
	}
	scheds := toRulesSchedules(schedRows)
	current := currentWindow(scheds, now)
	nextOpening := rules.NextOpening(scheds, now)

	cards := make([]map[string]any, 0, len(messes))
	for _, mess := range messes {
		card := map[string]any{"id": mess.ID, "code": mess.Code, "name": mess.Name}

		var lastUpdate *time.Time
		var level string
		if row, err := a.Q.GetLatestCrowdStatus(r.Context(), mess.ID); err == nil && row.UpdatedAt.Valid {
			lastUpdate = &row.UpdatedAt.Time
			level = string(row.Level)
		}
		card["crowd"] = crowdStateJSON(now, current, lastUpdate, level, nextOpening)
		cards = append(cards, card)
	}
	writeJSON(w, http.StatusOK, map[string]any{"messes": cards, "serverTime": now})
}

// crowdStateJSON builds the crowd object per ARCHITECTURE.md section 6.
func crowdStateJSON(now time.Time, current *mealWindow, lastUpdate *time.Time, level string, nextOpening *time.Time) map[string]any {
	var cs rules.CrowdState
	if current == nil {
		cs = rules.ComputeCrowdState(now, time.Time{}, time.Time{}, nil, "", config.StaleAfterMinutes, nextOpening)
	} else {
		cs = rules.ComputeCrowdState(now, current.Start, current.End, lastUpdate, level, config.StaleAfterMinutes, nextOpening)
	}
	m := map[string]any{
		"state":       cs.State,
		"level":       nil,
		"updatedAt":   nil,
		"ageMinutes":  nil,
		"nextOpening": nil,
	}
	if cs.Level != "" {
		m["level"] = cs.Level
	}
	if cs.UpdatedAt != nil {
		m["updatedAt"] = cs.UpdatedAt
	}
	if cs.AgeMinutes != nil {
		m["ageMinutes"] = *cs.AgeMinutes
	}
	if cs.NextOpening != nil {
		m["nextOpening"] = cs.NextOpening
	}
	return m
}

// GET /api/meals?date=YYYY-MM-DD — published meals for the user's group on a date.
func (a *App) listMeals(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	if u.Role != store.UserRoleStudent {
		writeErr(w, http.StatusForbidden, ErrForbidden, "Only students use this endpoint.")
		return
	}
	if !u.HostelGroup.Valid {
		writeErr(w, http.StatusForbidden, ErrForbidden, "Choose your hostel group first.")
		return
	}
	now := a.Clock.Now()
	var date time.Time
	if dateStr := r.URL.Query().Get("date"); dateStr == "" {
		date = todayInKolkata(now)
	} else {
		var err error
		if date, err = parseKolkataDate(dateStr); err != nil {
			writeErrFields(w, http.StatusBadRequest, ErrValidation, "Use date=YYYY-MM-DD.", map[string]string{"date": "Use the YYYY-MM-DD format."})
			return
		}
	}

	meals, err := a.Q.GetPublishedMealsByDate(r.Context(), store.GetPublishedMealsByDateParams{
		HostelGroup: u.HostelGroup.HostelGroup,
		ServiceDate: toPgDate(date),
	})
	if err != nil {
		serverError(w, err)
		return
	}
	schedRows, err := a.Q.GetMealSchedules(r.Context(), u.HostelGroup.HostelGroup)
	if err != nil {
		serverError(w, err)
		return
	}
	scheds := toRulesSchedules(schedRows)

	out := make([]map[string]any, 0, len(meals))
	for _, meal := range meals {
		out = append(out, a.mealJSON(r, meal, scheds, date, now, u))
	}
	writeJSON(w, http.StatusOK, map[string]any{"meals": out, "serverTime": now, "date": date.Format("2006-01-02")})
}

// mealJSON renders one meal with items, ratings and review state.
func (a *App) mealJSON(r *http.Request, meal store.Meals, scheds []rules.Schedule, date, now time.Time, u store.Users) map[string]any {
	ctx := r.Context()
	items, err := a.Q.GetMealItems(ctx, meal.ID)
	if err != nil {
		items = nil
	}
	itemList := make([]map[string]any, 0, len(items))
	for _, it := range items {
		itemList = append(itemList, map[string]any{
			"id":        it.ID,
			"name":      it.DishName,
			"dietMenu":  string(it.DietMenu.DietMenu),
			"sortOrder": it.SortOrder,
		})
	}

	ratings := map[string]any{
		"regular": map[string]any{"avg": nil, "count": 0},
		"jain":    map[string]any{"avg": nil, "count": 0},
	}
	if summary, err := a.Q.GetMealRatingSummary(ctx, meal.ID); err == nil {
		for _, s := range summary {
			key := map[store.DietPref]string{store.DietPrefRegular: "regular", store.DietPrefJain: "jain"}[s.Variant]
			avg := any(nil)
			if f8, err := s.AvgStars.Float64Value(); err == nil && f8.Valid {
				avg = f8.Float64
			}
			ratings[key] = map[string]any{"avg": avg, "count": s.ReviewCount}
		}
	}

	hasReviewed := false
	if _, err := a.Q.GetReviewByMealAndUser(ctx, store.GetReviewByMealAndUserParams{
		MealID: meal.ID,
		UserID: u.ID,
	}); err == nil {
		hasReviewed = true
	}

	window := findMealWindow(scheds, string(meal.MealType), date)
	var startsAt, endsAt any
	reviewState := "not_yet_open"
	if window != nil {
		startsAt, endsAt = window.Start, window.End
		reviewState = rules.ReviewState(window.Start, window.End, hasReviewed, now)
	}

	// Items visible to a student: their preference or "both" (section 6).
	visible := make([]map[string]any, 0, len(itemList))
	for _, it := range itemList {
		if dm, _ := it["dietMenu"].(string); dm == "both" || dm == string(u.DietPref.DietPref) {
			visible = append(visible, it)
		}
	}

	return map[string]any{
		"id":          meal.ID,
		"mealType":    string(meal.MealType),
		"status":      string(meal.Status.MealStatus),
		"serviceDate": date.Format("2006-01-02"),
		"startsAt":    startsAt,
		"endsAt":      endsAt,
		"items":       visible,
		"ratings":     ratings,
		"reviewState": reviewState,
		"hasReviewed": hasReviewed,
		"serverTime":  now,
	}
}

// GET /api/meals/{mealId}/reviews — non-hidden comments with stars, newest first.
func (a *App) listMealReviews(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	mealID, err := uuid.Parse(chi.URLParam(r, "mealId"))
	if err != nil {
		writeErr(w, http.StatusNotFound, ErrNotFound, "Meal not found.")
		return
	}
	meal, err := a.Q.GetMealByID(r.Context(), uid(mealID))
	if err != nil {
		writeErr(w, http.StatusNotFound, ErrNotFound, "Meal not found.")
		return
	}
	if !u.HostelGroup.Valid || meal.HostelGroup != u.HostelGroup.HostelGroup {
		writeErr(w, http.StatusForbidden, ErrForbidden, "This meal is not in your hostel group.")
		return
	}

	variant := r.URL.Query().Get("variant")
	if variant == "" {
		variant = string(u.DietPref.DietPref)
	}
	if variant != "regular" && variant != "jain" {
		writeErrFields(w, http.StatusBadRequest, ErrValidation, "Choose regular or jain.", map[string]string{"variant": "Choose regular or jain."})
		return
	}
	limit := int32(20)
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.ParseInt(l, 10, 64); err == nil && n >= 1 && n <= 100 {
			limit = int32(n)
		}
	}
	offset := int32(0)
	if c := r.URL.Query().Get("cursor"); c != "" {
		if n, err := strconv.ParseInt(c, 10, 64); err == nil && n > 0 {
			offset = int32(n)
		}
	}

	rows, err := a.Q.GetVisibleReviews(r.Context(), store.GetVisibleReviewsParams{
		MealID:  uid(mealID),
		Variant: store.DietPref(variant),
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		serverError(w, err)
		return
	}
	reviews := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		reviews = append(reviews, map[string]any{
			"id":          fromUUID(row.ID),
			"variant":     string(row.Variant),
			"stars":       row.Stars,
			"comment":     derefString(row.Comment),
			"createdAt":   row.CreatedAt.Time,
			"maskedEmail": maskedEmail(row.UserEmail),
		})
	}
	var nextCursor any
	if int32(len(rows)) == limit {
		nextCursor = strconv.FormatInt(int64(offset+limit), 10)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"reviews":    reviews,
		"variant":    variant,
		"nextCursor": nextCursor,
		"serverTime": a.Clock.Now(),
	})
}
