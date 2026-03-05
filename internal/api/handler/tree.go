package handler

import (
	"log/slog"
	"net/http"

	"qodex/internal/service"
)

// Tree handles GET /api/tree.
func Tree(svc *service.IngestService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := svc.TreeData()
		writeJSON(w, http.StatusOK, data)
	}
}
