package middleware

import (
	"net/http"
	"strings"

	"qodex/internal/auth"
)

// Auth returns middleware that enforces access code authentication.
// It allows auth endpoints and static files through, blocks API calls without a valid session.
// On each valid API call, the session's last activity time is extended.
func Auth(mgr *auth.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// Always allow auth endpoints
			if strings.HasPrefix(path, "/api/auth/") {
				next.ServeHTTP(w, r)
				return
			}

			// Allow static files (frontend)
			if !strings.HasPrefix(path, "/api/") {
				next.ServeHTTP(w, r)
				return
			}

			// Check session cookie for API calls
			cookie, err := r.Cookie("qodex_session")
			if err != nil || !mgr.ValidSession(cookie.Value) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Extend session on activity
			mgr.TouchSession(cookie.Value)

			next.ServeHTTP(w, r)
		})
	}
}
