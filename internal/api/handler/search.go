package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"qodex/internal/indexer"
	"qodex/pkg/models"
)

// Search handles GET /api/search.
func Search(idx indexer.Indexer, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "query parameter 'q' is required",
			})
			return
		}

		limit := 20
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		results, err := idx.Search(r.Context(), query, limit)
		if err != nil {
			logger.Error("search failed", "error", err, "query", query)
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "search failed",
			})
			return
		}

		resp := models.SearchResponse{
			Query:   query,
			Results: results,
			Total:   len(results),
		}

		writeJSON(w, http.StatusOK, resp)
	}
}
