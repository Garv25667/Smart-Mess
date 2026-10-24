package httpapi

import (
	"net/http"
	"strings"
	"time"

	"smartmess/internal/rules"
	"smartmess/internal/store"
)

type createReviewBody struct {
	Variant string `json:"variant"`
	Stars   any    `json:"stars"`
	Comment string `json:"comment"`
}

// POST /api/meals/{mealId}/reviews — one review per student per meal, inside the window.
func (a *App) createReview(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	if u.Role != store.UserRoleStudent {
		writeErr(w, http.StatusForbidden, ErrForbidden, "Only students submit reviews.")
		return
	}
	mealID, err := parseIDParam(r, "mealId")
	if err != nil {
		writeErr(w, http.StatusNotFound, ErrNotFound, "Meal not found.")
		return
	}
	meal, err := a.Q.GetMealByID(r.Context(), uid(mealID))
	if err != nil {
		writeErr(w, http.StatusNotFound, ErrNotFound, "Meal not found.")
		return
	}
	// Check 1: the student's group must equal the meal's group.
	if !u.HostelGroup.Valid || meal.HostelGroup != u.HostelGroup.HostelGroup {
		writeErr(w, http.StatusForbidden, ErrForbidden, "This meal is not in your hostel group.")
		return
	}
	// Check 2: meal is published.
	if meal.Status.MealStatus != store.MealStatusPublished {
		writeErr(w, http.StatusConflict, ErrMealNotPublished, "The menu for this meal is not published yet.")
		return
	}

	var body createReviewBody
	if err := decodeJSON(r, &body); err != nil {
		writeErrFields(w, http.StatusBadRequest, ErrValidation, "Invalid request body.", map[string]string{"stars": "Stars must be 1 to 5."})
		return
	}

	// Schedule window for this meal and service date.
	schedRows, err := a.Q.GetMealSchedules(r.Context(), meal.HostelGroup)
	if err != nil {
		serverError(w, err)
		return
	}
	scheds := toRulesSchedules(schedRows)
	window := findMealWindow(scheds, string(meal.MealType), fromDate(meal.ServiceDate))
	if window == nil {
		writeErr(w, http.StatusInternalServerError, "INTERNAL", "Meal schedule missing.")
		return
	}

	// Check 5: stars an integer 1..5; variant regular or jain.
	fields := map[string]string{}
	stars, hasStars := body.Stars.(float64)
	if !hasStars || stars != float64(int(stars)) || stars < 1 || stars > 5 {
		fields["stars"] = "Stars must be a whole number from 1 to 5."
	}
	if !isValidDietPref(body.Variant) {
		fields["variant"] = "Choose regular or jain."
	}
	comment := strings.TrimSpace(body.Comment)
	if len([]rune(comment)) > 500 {
		fields["comment"] = "Comments are at most 500 characters."
	}
	if len(fields) > 0 {
		writeErrFields(w, http.StatusBadRequest, ErrValidation, "Fix the highlighted fields.", fields)
		return
	}

	now := a.Clock.Now()
	// Check 3: inside the meal window (C6).
	if state := rules.ReviewState(window.Start, window.End, false, now); state != "open" {
		if state == "not_yet_open" {
			writeErr(w, http.StatusConflict, ErrReviewWindowClosed, "Reviews for this meal are not open yet.")
			return
		}
		writeErr(w, http.StatusConflict, ErrReviewWindowClosed, "Reviews for this meal have closed.")
		return
	}

	// Check 4: no review exists for (meal, user).
	if _, err := a.Q.GetReviewByMealAndUser(r.Context(), store.GetReviewByMealAndUserParams{
		MealID: uid(mealID),
		UserID: u.ID,
	}); err == nil {
		writeErr(w, http.StatusConflict, ErrReviewAlready, "You already reviewed this meal.")
		return
	}

	var commentArg *string
	if comment != "" {
		c := comment
		commentArg = &c
	}
	review, err := a.Q.CreateReview(r.Context(), store.CreateReviewParams{
		MealID:  uid(mealID),
		UserID:  u.ID,
		Variant: store.DietPref(body.Variant),
		Stars:   int16(stars),
		Comment: commentArg,
	})
	if err != nil {
		if isUniqueViolation(err) {
			writeErr(w, http.StatusConflict, ErrReviewAlready, "You already reviewed this meal.")
			return
		}
		serverError(w, err)
		return
	}

	created := fromTstamp(review.CreatedAt)
	if created.IsZero() {
		created = time.Now()
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"review": reviewJSON(review, created),
	})
}

// reviewJSON renders a review row with API field names.
func reviewJSON(rv store.Reviews, createdAt time.Time) map[string]any {
	return map[string]any{
		"id":        fromUUID(rv.ID),
		"mealId":    fromUUID(rv.MealID),
		"variant":   string(rv.Variant),
		"stars":     rv.Stars,
		"comment":   derefString(rv.Comment),
		"createdAt": createdAt,
	}
}
