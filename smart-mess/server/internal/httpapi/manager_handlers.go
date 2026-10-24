package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"smartmess/internal/store"
)

// channelItem is one menu item as sent by the manager editor.
type channelItem struct {
	Name      string `json:"name"`
	DietMenu  string `json:"dietMenu"`
	SortOrder int32  `json:"sortOrder"`
}

func mealRowJSON(meal store.Meals) map[string]any {
	publishedAt := any(nil)
	if meal.PublishedAt.Valid {
		publishedAt = meal.PublishedAt.Time
	}
	return map[string]any{
		"id":          fromUUID(meal.ID),
		"hostelGroup": string(meal.HostelGroup),
		"serviceDate": fromDate(meal.ServiceDate).Format("2006-01-02"),
		"mealType":    string(meal.MealType),
		"status":      string(meal.Status.MealStatus),
		"publishedAt": publishedAt,
	}
}

// PUT /api/manager/messes/{messId}/crowd — mess_manager only, assigned messes only.
func (a *App) updateCrowd(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	messID, err := parseIDParam(r, "messId")
	if err != nil {
		writeErr(w, http.StatusNotFound, ErrNotFound, "Mess not found.")
		return
	}
	mess, err := a.Q.GetMessByID(r.Context(), uid(messID))
	if err != nil {
		writeErr(w, http.StatusNotFound, ErrNotFound, "Mess not found.")
		return
	}
	assigned, err := a.Q.IsMessAssignedToManager(r.Context(), store.IsMessAssignedToManagerParams{
		UserID: u.ID,
		MessID: mess.ID,
	})
	if err != nil {
		serverError(w, err)
		return
	}
	if !assigned {
		writeErr(w, http.StatusForbidden, ErrForbidden, "This mess is not assigned to you.")
		return
	}

	var body struct {
		Level string `json:"level"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErrFields(w, http.StatusBadRequest, ErrValidation, "Invalid request body.", map[string]string{"level": "Choose low, medium or high."})
		return
	}
	var level store.CrowdLevel
	switch body.Level {
	case "low":
		level = store.CrowdLevelLow
	case "medium":
		level = store.CrowdLevelMedium
	case "high":
		level = store.CrowdLevelHigh
	default:
		writeErrFields(w, http.StatusBadRequest, ErrValidation, "Level must be low, medium or high.", map[string]string{"level": "Choose low, medium or high."})
		return
	}

	row, err := a.Q.CreateCrowdStatus(r.Context(), store.CreateCrowdStatusParams{
		MessID: mess.ID,
		Level: level,
		UpdatedBy: u.ID,
	})
	if err != nil {
		serverError(w, err)
		return
	}
	updatedAt := row.UpdatedAt
	if !updatedAt.Valid {
		updatedAt.Time = time.Now()
		updatedAt.Valid = true
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"crowd": map[string]any{
			"messId":    mess.ID,
			"messCode":  mess.Code,
			"level":     string(row.Level),
			"state":     "fresh",
			"updatedAt": updatedAt.Time,
			"serverTime": a.Clock.Now(),
		},
	})
}

// GET /api/manager/meals?date=YYYY-MM-DD — draft and published meals for the manager's group.
func (a *App) managerMeals(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	if !u.HostelGroup.Valid {
		writeErr(w, http.StatusForbidden, ErrForbidden, "Your account has no hostel group.")
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
	meals, err := a.Q.GetAllMealsByDate(r.Context(), store.GetAllMealsByDateParams{
		HostelGroup: u.HostelGroup.HostelGroup,
		ServiceDate: toPgDate(date),
	})
	if err != nil {
		serverError(w, err)
		return
	}
	out := make([]map[string]any, 0, len(meals))
	for _, meal := range meals {
		items, err := a.Q.GetMealItems(r.Context(), meal.ID)
		if err != nil {
			serverError(w, err)
			return
		}
		itemList := make([]map[string]any, 0, len(items))
		for _, it := range items {
			itemList = append(itemList, map[string]any{
				"id":        it.ID,
				"dishId":    it.DishID,
				"name":      it.DishName,
				"dietMenu":  string(it.DietMenu.DietMenu),
				"sortOrder": it.SortOrder,
			})
		}
		m := mealRowJSON(meal)
		m["items"] = itemList
		out = append(out, m)
	}
	writeJSON(w, http.StatusOK, map[string]any{"meals": out, "serverTime": now, "date": date.Format("2006-01-02")})
}

// GET /api/manager/default-items?mealType= — defaults to prefill the editor.
func (a *App) managerDefaultItems(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	if !u.HostelGroup.Valid {
		writeErr(w, http.StatusForbidden, ErrForbidden, "Your account has no hostel group.")
		return
	}
	mealType := r.URL.Query().Get("mealType")
	if !isValidMealType(mealType) {
		writeErrFields(w, http.StatusBadRequest, ErrValidation, " mealType must be breakfast, lunch, snacks or dinner.", map[string]string{"mealType": "Choose breakfast, lunch, snacks or dinner."})
		return
	}
	items, err := a.Q.GetDefaultItems(r.Context(), store.GetDefaultItemsParams{
		HostelGroup: u.HostelGroup.HostelGroup,
		MealType:    store.MealType(mealType),
	})
	if err != nil {
		serverError(w, err)
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, it := range items {
		out = append(out, map[string]any{"name": it.Name, "sortOrder": it.SortOrder})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out, "serverTime": a.Clock.Now()})
}

// PUT /api/manager/meals/{date}/{mealType} — replace a meal's items.
func (a *App) saveMeal(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	if !u.HostelGroup.Valid {
		writeErr(w, http.StatusForbidden, ErrForbidden, "Your account has no hostel group.")
		return
	}
	now := a.Clock.Now()
	date, err := parseKolkataDate(chi.URLParam(r, "date"))
	if err != nil {
		writeErrFields(w, http.StatusBadRequest, ErrValidation, "Date must be YYYY-MM-DD.", map[string]string{"date": "Use the YYYY-MM-DD format."})
		return
	}
	mealType := chi.URLParam(r, "mealType")
	if !isValidMealType(mealType) {
		writeErrFields(w, http.StatusBadRequest, ErrValidation, "Unknown meal type.", map[string]string{"mealType": "Choose breakfast, lunch, snacks or dinner."})
		return
	}
	// A10: menu edits are allowed for today and future dates only.
	if date.Before(todayInKolkata(now)) {
		writeErr(w, http.StatusConflict, ErrMenuDateInPast, "Menus can only be edited for today and future dates.")
		return
	}

	var body struct {
		Items []channelItem `json:"items"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, ErrValidation, "Invalid request body.")
		return
	}
	fields := map[string]string{}
	for i, it := range body.Items {
		if it.Name == "" || len(it.Name) > 200 {
			fields["items."+itoa(i)+".name"] = "Dish names must be 1 to 200 characters."
		}
		switch it.DietMenu {
		case "regular", "jain", "both":
		default:
			fields["items."+itoa(i)+".dietMenu"] = "Choose regular, jain or both."
		}
	}
	if len(fields) > 0 {
		writeErrFields(w, http.StatusBadRequest, ErrValidation, "Fix the highlighted fields.", fields)
		return
	}

	tx, err := a.Pool.Begin(r.Context())
	if err != nil {
		serverError(w, err)
		return
	}
	defer tx.Rollback(r.Context())
	qtx := a.Q.WithTx(tx)

	meal, err := qtx.UpsertMeal(r.Context(), store.UpsertMealParams{
		HostelGroup: u.HostelGroup.HostelGroup,
		ServiceDate: toPgDate(date),
		MealType:    store.MealType(mealType),
	})
	if err != nil {
		serverError(w, err)
		return
	}
	if err := qtx.DeleteMealItems(r.Context(), meal.ID); err != nil {
		serverError(w, err)
		return
	}
	saved := make([]map[string]any, 0, len(body.Items))
	for _, it := range body.Items {
		dish, err := qtx.GetOrCreateDish(r.Context(), store.GetOrCreateDishParams{
			HostelGroup: u.HostelGroup.HostelGroup,
			Name:        it.Name,
		})
		if err != nil {
			serverError(w, err)
			return
		}
		if _, err := qtx.CreateMealItem(r.Context(), store.CreateMealItemParams{
			MealID:    meal.ID,
			DishID:    dish.ID,
			DietMenu: store.NullDietMenu{DietMenu: store.DietMenu(it.DietMenu), Valid: true},
			SortOrder: &it.SortOrder,
		}); err != nil {
			serverError(w, err)
			return
		}
		saved = append(saved, map[string]any{"name": dish.Name, "dietMenu": it.DietMenu, "sortOrder": it.SortOrder})
	}
	if err := tx.Commit(r.Context()); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"meal": mealRowJSON(meal), "items": saved, "serverTime": now})
}

// POST /api/manager/meals/{date}/{mealType}/publish
func (a *App) publishMeal(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	if !u.HostelGroup.Valid {
		writeErr(w, http.StatusForbidden, ErrForbidden, "Your account has no hostel group.")
		return
	}
	date, err := parseKolkataDate(chi.URLParam(r, "date"))
	if err != nil {
		writeErrFields(w, http.StatusBadRequest, ErrValidation, "Date must be YYYY-MM-DD.", map[string]string{"date": "Use the YYYY-MM-DD format."})
		return
	}
	mealType := chi.URLParam(r, "mealType")
	if !isValidMealType(mealType) {
		writeErr(w, http.StatusNotFound, ErrNotFound, "Unknown meal type.")
		return
	}
	meal, err := a.Q.UpsertMeal(r.Context(), store.UpsertMealParams{
		HostelGroup: u.HostelGroup.HostelGroup,
		ServiceDate: toPgDate(date),
		MealType:    store.MealType(mealType),
	})
	if err != nil {
		serverError(w, err)
		return
	}
	if meal.HostelGroup != u.HostelGroup.HostelGroup {
		writeErr(w, http.StatusForbidden, ErrForbidden, "This meal is not in your group.")
		return
	}
	published, err := a.Q.PublishMeal(r.Context(), meal.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"meal": mealRowJSON(published), "serverTime": a.Clock.Now()})
}

// GET /api/manager/reviews?mealId= — all reviews including hidden.
func (a *App) managerReviews(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	if !u.HostelGroup.Valid {
		writeErr(w, http.StatusForbidden, ErrForbidden, "Your account has no hostel group.")
		return
	}
	mealIDStr := r.URL.Query().Get("mealId")
	mealID, err := uuid.Parse(mealIDStr)
	if err != nil {
		writeErrFields(w, http.StatusBadRequest, ErrValidation, " mealId must be a UUID.", map[string]string{"mealId": "Missing meal id."})
		return
	}
	meal, err := a.Q.GetMealByID(r.Context(), uid(mealID))
	if err != nil {
		writeErr(w, http.StatusNotFound, ErrNotFound, "Meal not found.")
		return
	}
	if meal.HostelGroup != u.HostelGroup.HostelGroup {
		writeErr(w, http.StatusForbidden, ErrForbidden, "This meal is not in your group.")
		return
	}
	rows, err := a.Q.GetAllReviewsForMeal(r.Context(), uid(mealID))
	if err != nil {
		serverError(w, err)
		return
	}
	reviews := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		var comment any
		if row.Comment != nil {
			comment = *row.Comment
		}
		reviews = append(reviews, map[string]any{
			"id":          fromUUID(row.ID),
			"variant":     string(row.Variant),
			"stars":       row.Stars,
			"comment":     comment,
			"hidden":      row.Hidden != nil && *row.Hidden,
			"createdAt":   row.CreatedAt.Time,
			"userEmail":   row.UserEmail,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"reviews":    reviews,
		"meal":       mealRowJSON(meal),
		"serverTime": a.Clock.Now(),
	})
}

// POST /api/manager/reviews/{reviewId}/hide and /unhide.
func (a *App) moderateReview(hidden bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := userFromCtx(r)
		if !u.HostelGroup.Valid {
			writeErr(w, http.StatusForbidden, ErrForbidden, "Your account has no hostel group.")
			return
		}
		reviewID, err := parseIDParam(r, "reviewId")
		if err != nil {
			writeErr(w, http.StatusNotFound, ErrNotFound, "Review not found.")
			return
		}
		review, err := a.Q.GetReviewByID(r.Context(), uid(reviewID))
		if err != nil {
			writeErr(w, http.StatusNotFound, ErrNotFound, "Review not found.")
			return
		}
		meal, err := a.Q.GetMealByID(r.Context(), review.MealID)
		if err != nil || meal.HostelGroup != u.HostelGroup.HostelGroup {
			writeErr(w, http.StatusForbidden, ErrForbidden, "This review is not in your group.")
			return
		}
		if hidden {
			if err := a.Q.HideReview(r.Context(), review.ID); err != nil {
				serverError(w, err)
				return
			}
		} else {
			if err := a.Q.UnhideReview(r.Context(), review.ID); err != nil {
				serverError(w, err)
				return
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"review":     map[string]any{"id": review.ID, "hidden": hidden},
			"serverTime": a.Clock.Now(),
		})
	}
}

func itoa(i int) string {
	return strconv.Itoa(i)
}
