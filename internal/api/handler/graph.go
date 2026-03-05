package handler

import (
	"log/slog"
	"net/http"

	"qodex/internal/service"
)

// Graph handles GET /api/graph.
func Graph(svc *service.IngestService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := svc.GraphData()
		writeJSON(w, http.StatusOK, data)
	}
}
