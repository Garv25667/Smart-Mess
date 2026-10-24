package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"smartmess/internal/auth"
	"smartmess/internal/config"
	"smartmess/internal/store"
)

type requestOTPBody struct {
	Email string `json:"email"`
}

// POST /api/auth/request-otp
func (a *App) requestOTP(w http.ResponseWriter, r *http.Request) {
	var body requestOTPBody
	if err := decodeJSON(r, &body); err != nil {
		writeErrFields(w, http.StatusBadRequest, ErrValidation, "Invalid request body.", map[string]string{"email": "Enter your college email."})
		return
	}
	email := auth.NormalizeEmail(body.Email)
	if err := auth.ValidateEmailDomain(email, a.Cfg.AllowedEmailDomain); err != nil {
		writeErr(w, http.StatusBadRequest, ErrEmailDomain, "Use your IIIT Pune college email address.")
		return
	}

	// Max 5 code requests per email per hour.
	recent, err := a.Q.CountRecentOTPs(r.Context(), email)
	if err == nil && recent >= config.OTPMaxRequestsPerHr {
		writeErr(w, http.StatusTooManyRequests, ErrOTPRateLimited, "Too many code requests. Try again later.")
		return
	}

	code, err := auth.GenerateOTPCode()
	if err != nil {
		serverError(w, err)
		return
	}
	now := a.Clock.Now()
	if _, err := a.Q.CreateOTP(r.Context(), store.CreateOTPParams{
		Btrim:     email,
		CodeHash:  auth.HashOTP(code, a.Cfg.OTPSecret),
		ExpiresAt: tstamp(now.Add(config.OTPExpiryMinutes * time.Minute)),
	}); err != nil {
		serverError(w, err)
		return
	}

	if err := a.Mailer.SendOTP(email, code); err != nil {
		log.Printf("otp send error to=%s err=%v", email, err)
	}
	writeJSON(w, http.StatusNoContent, nil)
}

type verifyOTPBody struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

// POST /api/auth/verify-otp
func (a *App) verifyOTP(w http.ResponseWriter, r *http.Request) {
	var body verifyOTPBody
	if err := decodeJSON(r, &body); err != nil {
		writeErrFields(w, http.StatusBadRequest, ErrValidation, "Invalid request body.", map[string]string{"code": "Enter the 6-digit code."})
		return
	}
	email := auth.NormalizeEmail(body.Email)
	otp, err := a.Q.GetActiveOTP(r.Context(), email)
	if err != nil {
		writeErr(w, http.StatusBadRequest, ErrOTPInvalid, "That code is not valid. Request a new code.")
		return
	}
	if otp.Attempts != nil && *otp.Attempts >= config.OTPMaxAttempts {
		writeErr(w, http.StatusTooManyRequests, ErrOTPRateLimited, "Too many wrong attempts. Request a new code.")
		return
	}

	if !auth.VerifyOTPHash(body.Code, otp.CodeHash, a.Cfg.OTPSecret) {
		_ = a.Q.IncrementOTPAttempts(r.Context(), otp.ID)
		attempts := int32(1)
		if otp.Attempts != nil {
			attempts = *otp.Attempts + 1
		}
		if attempts >= config.OTPMaxAttempts {
			_ = a.Q.ConsumeOTP(r.Context(), otp.ID)
			writeErr(w, http.StatusTooManyRequests, ErrOTPRateLimited, "Too many wrong attempts. Request a new code.")
			return
		}
		writeErr(w, http.StatusBadRequest, ErrOTPInvalid, "That code is not valid. Check it and try again.")
		return
	}
	_ = a.Q.ConsumeOTP(r.Context(), otp.ID)

	u, err := a.Q.GetUserByEmail(r.Context(), email)
	if err != nil {
		// First sign-in creates a student account (managers are pre-seeded).
		u, err = a.Q.CreateUser(r.Context(), store.CreateUserParams{
			Btrim:       email,
			Role:        store.UserRoleStudent,
			HostelGroup: store.NullHostelGroup{},
		})
		if err != nil {
			serverError(w, err)
			return
		}
	}
	_ = a.Q.UpdateLastLogin(r.Context(), u.ID)

	token, err := auth.CreateJWT(u.ID.String(), string(u.Role), u.Email, a.Cfg.JWTSecret, config.SessionMaxAge)
	if err != nil {
		serverError(w, err)
		return
	}
	auth.SetSessionCookie(w, token, config.SessionMaxAge, a.Cfg.AppEnv == "production")
	writeJSON(w, http.StatusOK, map[string]any{"user": a.userToJSON(r, u)})
}

// POST /api/auth/logout
func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	auth.ClearSessionCookie(w)
	writeJSON(w, http.StatusNoContent, nil)
}

func decodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("empty body")
	}
	return json.NewDecoder(r.Body).Decode(dst)
}

func serverError(w http.ResponseWriter, err error) {
	if errors.Is(err, context.DeadlineExceeded) {
		writeErr(w, http.StatusServiceUnavailable, "SERVER_TIMEOUT", "The server is busy. Try again.")
		return
	}
	log.Printf("server error: %v", err)
	writeErr(w, http.StatusInternalServerError, "INTERNAL", "Something went wrong. Try again.")
}
