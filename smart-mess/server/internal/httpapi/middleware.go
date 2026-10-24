package httpapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"smartmess/internal/auth"
	"smartmess/internal/store"
)

type ctxKey string

const userCtxKey ctxKey = "user"

// userFromCtx returns the DB user loaded by SessionMiddleware.
func userFromCtx(r *http.Request) store.Users {
	u, _ := r.Context().Value(userCtxKey).(store.Users)
	return u
}

// userToJSON renders a user with camelCase API field names.
// For mess managers it also includes their assigned messes.
func (a *App) userToJSON(r *http.Request, u store.Users) map[string]any {
	var hostelGroup any
	if u.HostelGroup.Valid {
		hostelGroup = string(u.HostelGroup.HostelGroup)
	}
	var lastLoginAt any
	if u.LastLoginAt.Valid {
		lastLoginAt = u.LastLoginAt.Time
	}
	m := map[string]any{
		"id":            fromUUID(u.ID),
		"email":         u.Email,
		"role":          string(u.Role),
		"hostelGroup":   hostelGroup,
		"dietPref":      string(u.DietPref.DietPref),
		"cvdMode":       string(u.CvdMode.CvdMode),
		"textScale":     textScaleValue(u.TextScale),
		"theme":         string(u.Theme.ThemePref),
		"reducedMotion": derefBool(u.ReducedMotion),
		"createdAt":     u.CreatedAt.Time,
		"lastLoginAt":   lastLoginAt,
	}
	if u.Role == store.UserRoleMessManager {
		messes, err := a.Q.GetManagerMesses(r.Context(), u.ID)
		if err == nil {
			items := make([]map[string]any, 0, len(messes))
			for _, mess := range messes {
				items = append(items, map[string]any{
					"id": fromUUID(mess.ID), "code": mess.Code, "name": mess.Name,
				})
			}
			m["messes"] = items
		}
	}
	return m
}

// sessionUser loads a fresh DB user for the session cookie's user id.
func (a *App) sessionUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetSessionCookie(r)
		if err != nil {
			writeErr(w, http.StatusUnauthorized, ErrUnauthenticated, "Sign in to continue.")
			return
		}
		claims, err := auth.ParseJWT(token, a.Cfg.JWTSecret)
		if err != nil {
			writeErr(w, http.StatusUnauthorized, ErrUnauthenticated, "Sign in to continue.")
			return
		}
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			writeErr(w, http.StatusUnauthorized, ErrUnauthenticated, "Sign in to continue.")
			return
		}
		u, err := a.Q.GetUserByID(r.Context(), uid(userID))
		if err != nil {
			writeErr(w, http.StatusUnauthorized, ErrUnauthenticated, "Sign in to continue.")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userCtxKey, u)))
	})
}

// requireRole rejects users whose role is not in the list.
func (a *App) requireRole(roles ...store.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := userFromCtx(r)
			allowed := false
			for _, role := range roles {
				if u.Role == role {
					allowed = true
					break
				}
			}
			if !allowed {
				writeErr(w, http.StatusForbidden, ErrForbidden, "You do not have access to this resource.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
