package handler

import (
	"log/slog"
	"net/http"

	"qodex/internal/service"
)

// Repos handles GET /api/repos.
func Repos(svc *service.IngestService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := svc.ListRepos()
		writeJSON(w, http.StatusOK, data)
	}
}
