package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"qodex/internal/service"
	"qodex/pkg/models"
)

// Ingest handles POST /api/ingest.
func Ingest(svc *service.IngestService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.IngestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("failed to decode ingest request", "error", err)
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid request body",
			})
			return
		}

		// Add timeout for the entire ingest operation
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
		defer cancel()

		resp, err := svc.Ingest(ctx, &req)
		if err != nil {
			logger.Error("ingest failed", "error", err, "url", req.URL)
			if resp != nil {
				writeJSON(w, http.StatusInternalServerError, resp)
			} else {
				writeJSON(w, http.StatusBadRequest, map[string]string{
					"error": err.Error(),
				})
			}
			return
		}

		writeJSON(w, http.StatusOK, resp)
	}
}
