package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// Logger returns middleware that logs HTTP requests with session, IP, and browser info.
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			next.ServeHTTP(rw, r)

			// Extract session ID from cookie (truncated for log readability)
			sessionID := "-"
			if cookie, err := r.Cookie("qodex_session"); err == nil && cookie.Value != "" {
				v := cookie.Value
				if len(v) > 8 {
					v = v[:8]
				}
				sessionID = v
			}

			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"status", rw.status,
				"duration", time.Since(start),
				"size", rw.size,
				"ip", clientIP(r),
				"session", sessionID,
				"user_agent", r.UserAgent(),
				"referer", r.Referer(),
			)
		})
	}
}

// clientIP extracts the client IP, checking X-Forwarded-For and X-Real-IP headers first.
func clientIP(r *http.Request) string {
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
