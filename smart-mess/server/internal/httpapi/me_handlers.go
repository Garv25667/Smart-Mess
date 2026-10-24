package httpapi

import (
	"net/http"
	

	"smartmess/internal/store"
)

// GET /api/me
func (a *App) me(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	writeJSON(w, http.StatusOK, map[string]any{
		"user":       a.userToJSON(r, u),
		"serverTime": a.Clock.Now(),
	})
}

type preferencesBody struct {
	DietPref      *string `json:"dietPref"`
	CvdMode       *string `json:"cvdMode"`
	TextScale     *float64 `json:"textScale"`
	Theme         *string `json:"theme"`
	ReducedMotion *bool   `json:"reducedMotion"`
}

// PATCH /api/me/preferences
func (a *App) updatePreferences(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	var body preferencesBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, ErrValidation, "Invalid request body.")
		return
	}

	fields := map[string]string{}
	params := store.UpdatePreferencesParams{ID: u.ID}

	if body.DietPref != nil {
		if *body.DietPref != "regular" && *body.DietPref != "jain" {
			fields["dietPref"] = "Choose regular or jain."
		} else {
			var v store.NullDietPref
			v.DietPref, v.Valid = store.DietPref(*body.DietPref), true
			params.DietPref = v
		}
	}
	if body.CvdMode != nil {
		switch *body.CvdMode {
		case "off", "unsure", "protan", "deutan", "tritan", "mono":
			var v store.NullCvdMode
			v.CvdMode, v.Valid = store.CvdMode(*body.CvdMode), true
			params.CvdMode = v
		default:
			fields["cvdMode"] = "Choose a valid color vision mode."
		}
	}
	if body.TextScale != nil {
		if *body.TextScale < 1.0 || *body.TextScale > 2.0 {
			fields["textScale"] = "Text size must be between 1.0 and 2.0."
		} else {
			params.TextScale = floatToNumeric(*body.TextScale)
		}
	}
	if body.Theme != nil {
		switch *body.Theme {
		case "system", "light", "dark", "high_contrast":
			var v store.NullThemePref
			v.ThemePref, v.Valid = store.ThemePref(*body.Theme), true
			params.Theme = v
		default:
			fields["theme"] = "Choose a valid theme."
		}
	}
	if body.ReducedMotion != nil {
		rm := *body.ReducedMotion
		params.ReducedMotion = &rm
	}

	if len(fields) > 0 {
		writeErrFields(w, http.StatusBadRequest, ErrValidation, "Fix the highlighted fields.", fields)
		return
	}

	updated, err := a.Q.UpdatePreferences(r.Context(), params)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": a.userToJSON(r, updated)})
}

type hostelGroupBody struct {
	HostelGroup string `json:"hostelGroup"`
}

// POST /api/me/hostel-group — set once by a student.
func (a *App) setHostelGroup(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	if u.Role != store.UserRoleStudent {
		writeErr(w, http.StatusForbidden, ErrForbidden, "Only students choose a hostel group.")
		return
	}
	var body hostelGroupBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, ErrValidation, "Invalid request body.")
		return
	}
	var hg store.HostelGroup
	switch body.HostelGroup {
	case "girls":
		hg = store.HostelGroupGirls
	case "boys":
		hg = store.HostelGroupBoys
	default:
		writeErrFields(w, http.StatusBadRequest, ErrValidation, "Choose girls or boys.", map[string]string{"hostelGroup": "Choose girls or boys."})
		return
	}
	updated, err := a.Q.SetHostelGroup(r.Context(), store.SetHostelGroupParams{ID: u.ID, HostelGroup: store.NullHostelGroup{HostelGroup: hg, Valid: true}})
	if err != nil {
		// The WHERE clause only matches when hostel_group IS NULL,
		// so "no row" here means it was already chosen.
		writeErr(w, http.StatusConflict, ErrHostelGroupSet, "Your hostel group is already set and cannot be changed.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": a.userToJSON(r, updated)})
}
