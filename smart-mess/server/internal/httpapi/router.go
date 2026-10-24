package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"

	"smartmess/internal/config"
)

// securityHeaders adds the basic header set from ARCHITECTURE.md section 11.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

// Register mounts every route under /api on the given router.
func Register(r chi.Router, app *App) {
	r.Use(securityHeaders)
	r.Use(corsMiddleware(app.Cfg))

	r.Route("/api", func(api chi.Router) {
		api.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
		})

		// Auth: rate limited per IP.
		api.Group(func(auth chi.Router) {
			auth.Use(httprate.LimitByIP(120, time.Minute))
			auth.With(httprate.LimitByIP(10, time.Hour)).Post("/auth/request-otp", app.requestOTP)
			auth.With(httprate.LimitByIP(30, time.Minute)).Post("/auth/verify-otp", app.verifyOTP)
			auth.Post("/auth/logout", app.logout)
		})

		api.Group(func(authed chi.Router) {
			authed.Use(app.sessionUser)

			authed.Get("/me", app.me)
			authed.Patch("/me/preferences", app.updatePreferences)
			authed.Post("/me/hostel-group", app.setHostelGroup)

			authed.Get("/messes", app.listMesses)
			authed.Get("/meals", app.listMeals)
			authed.Get("/meals/{mealId}/reviews", app.listMealReviews)

			// Review submission: moderate per-IP limit (per-user rule is 1/meal in DB).
			authed.With(httprate.LimitByIP(20, time.Minute)).Post("/meals/{mealId}/reviews", app.createReview)

			authed.Route("/manager", func(mgr chi.Router) {
				mgr.Group(func(m chi.Router) {
					m.Use(app.requireRole("main_manager"))
					m.Get("/meals", app.managerMeals)
					m.Get("/default-items", app.managerDefaultItems)
					m.Put("/meals/{date}/{mealType}", app.saveMeal)
					m.Post("/meals/{date}/{mealType}/publish", app.publishMeal)
					m.Get("/reviews", app.managerReviews)
					m.Post("/reviews/{reviewId}/hide", app.moderateReview(true))
					m.Post("/reviews/{reviewId}/unhide", app.moderateReview(false))
				})

				mgr.Group(func(m chi.Router) {
					m.Use(app.requireRole("mess_manager"))
					m.Put("/messes/{messId}/crowd", app.updateCrowd)
				})
			})
		})
	})
}

// corsMiddleware allows only the configured client origin with credentials.
func corsMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	origin := cfg.ClientOrigin
	return func(next http.Handler) http.Handler {
		allow := strings.HasPrefix(origin, "http")
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if allow {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				w.Header().Set("Vary", "Origin")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
