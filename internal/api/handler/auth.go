package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"qodex/internal/auth"
)

// AuthStatus handles GET /api/auth/status.
func AuthStatus(mgr *auth.Manager, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"setup": mgr.IsSetup()})
	}
}

// AuthSetup handles POST /api/auth/setup.
func AuthSetup(mgr *auth.Manager, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		if err := mgr.Setup(req.Code); err != nil {
			logger.Warn("auth setup failed", "error", err)
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		// Auto-login after setup
		token, err := mgr.Verify(req.Code)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "qodex_session",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// AuthKeepalive handles POST /api/auth/keepalive.
func AuthKeepalive(mgr *auth.Manager, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("qodex_session")
		if err != nil || !mgr.ValidSession(cookie.Value) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		mgr.TouchSession(cookie.Value)
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// AuthVerify handles POST /api/auth/verify.
func AuthVerify(mgr *auth.Manager, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		token, err := mgr.Verify(req.Code)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid access code"})
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "qodex_session",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
