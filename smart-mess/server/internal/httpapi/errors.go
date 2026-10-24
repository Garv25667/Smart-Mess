package httpapi

import (
	"encoding/json"
	"net/http"
)

// Error codes per ARCHITECTURE.md section 9.
const (
	ErrValidation           = "VALIDATION_ERROR"
	ErrEmailDomain          = "EMAIL_DOMAIN_NOT_ALLOWED"
	ErrOTPInvalid           = "OTP_INVALID"
	ErrOTPRateLimited       = "OTP_RATE_LIMITED"
	ErrUnauthenticated      = "UNAUTHENTICATED"
	ErrForbidden            = "FORBIDDEN"
	ErrNotFound             = "NOT_FOUND"
	ErrHostelGroupSet       = "HOSTEL_GROUP_ALREADY_SET"
	ErrMealNotPublished     = "MEAL_NOT_PUBLISHED"
	ErrReviewWindowClosed   = "REVIEW_WINDOW_CLOSED"
	ErrReviewAlready        = "REVIEW_ALREADY_SUBMITTED"
	ErrMenuDateInPast       = "MENU_DATE_IN_PAST"
)

type apiError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type errorBody struct {
	Error apiError `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, code, message string) {
	writeErrFields(w, status, code, message, nil)
}

func writeErrFields(w http.ResponseWriter, status int, code, message string, fields map[string]string) {
	writeJSON(w, status, errorBody{Error: apiError{Code: code, Message: message, Fields: fields}})
}
