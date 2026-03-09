package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"qodex/internal/service"
)

type timelineRequest struct {
	Files []string `json:"files"`
}

// Timeline handles POST /api/graph/timeline.
// Returns a filtered graph containing only the specified files and their interconnecting links.
func Timeline(svc *service.IngestService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req timelineRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		data := svc.FilteredGraph(req.Files)
		writeJSON(w, http.StatusOK, data)
	}
}
