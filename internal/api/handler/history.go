package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"qodex/internal/service"
	"qodex/pkg/models"
)

// History handles GET /api/history.
func History(svc *service.IngestService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 50
		if v := r.URL.Query().Get("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				limit = n
			}
		}

		commits, err := svc.CommitHistory(limit)
		if err != nil {
			logger.Warn("commit history failed", "error", err)
			writeJSON(w, http.StatusOK, models.HistoryResponse{})
			return
		}

		repoURL := strings.TrimSuffix(svc.CurrentRepoURL(), ".git")

		writeJSON(w, http.StatusOK, models.HistoryResponse{
			RepoURL: repoURL,
			Commits: commits,
		})
	}
}
