package handler

import (
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"strings"
)

type clientEvent struct {
	Action string `json:"action"`
	Target string `json:"target,omitempty"`
	Value  string `json:"value,omitempty"`
}

type eventsRequest struct {
	Events []clientEvent `json:"events"`
}

// Events handles POST /api/events — logs frontend UI events.
func Events(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req eventsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		sessionID := "-"
		if cookie, err := r.Cookie("qodex_session"); err == nil && cookie.Value != "" {
			v := cookie.Value
			if len(v) > 8 {
				v = v[:8]
			}
			sessionID = v
		}

		ip := eventClientIP(r)
		ua := r.UserAgent()

		for _, evt := range req.Events {
			logger.Info("ui_event",
				"action", evt.Action,
				"target", evt.Target,
				"value", evt.Value,
				"ip", ip,
				"session", sessionID,
				"user_agent", ua,
			)
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func eventClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if ip := strings.TrimSpace(strings.Split(xff, ",")[0]); ip != "" {
			return ip
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
