package api

import (
	"log/slog"
	"net/http"

	"qodex/internal/api/handler"
	"qodex/internal/api/middleware"
	"qodex/internal/auth"
	"qodex/internal/config"
	"qodex/internal/indexer"
	"qodex/internal/service"
)

// NewRouter creates the HTTP handler with all routes and middleware.
// If authMgr is non-nil, access code protection is enabled.
func NewRouter(
	cfg *config.Config,
	logger *slog.Logger,
	ingestSvc *service.IngestService,
	idx indexer.Indexer,
	authMgr *auth.Manager,
) http.Handler {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("POST /api/ingest", handler.Ingest(ingestSvc, logger))
	mux.HandleFunc("GET /api/graph", handler.Graph(ingestSvc, logger))
	mux.HandleFunc("GET /api/tree", handler.Tree(ingestSvc, logger))
	mux.HandleFunc("GET /api/search", handler.Search(idx, logger))
	mux.HandleFunc("POST /api/graph/timeline", handler.Timeline(ingestSvc, logger))
	mux.HandleFunc("GET /api/repos", handler.Repos(ingestSvc, logger))
	mux.HandleFunc("GET /api/history", handler.History(ingestSvc, logger))
	mux.HandleFunc("GET /api/file", handler.File(ingestSvc, logger))

	// Auth routes (always registered, only enforced when authMgr is set)
	if authMgr != nil {
		mux.HandleFunc("GET /api/auth/status", handler.AuthStatus(authMgr, logger))
		mux.HandleFunc("POST /api/auth/setup", handler.AuthSetup(authMgr, logger))
		mux.HandleFunc("POST /api/auth/verify", handler.AuthVerify(authMgr, logger))
		mux.HandleFunc("POST /api/auth/keepalive", handler.AuthKeepalive(authMgr, logger))
	}

	// Serve frontend static files
	fs := http.FileServer(http.Dir(cfg.Frontend.StaticDir))
	mux.Handle("/", fs)

	// Apply middleware stack: Recovery → Logger → CORS → (Auth)
	var h http.Handler = mux
	if authMgr != nil {
		h = middleware.Auth(authMgr)(h)
	}
	h = middleware.CORS(&cfg.CORS)(h)
	h = middleware.Logger(logger)(h)
	h = middleware.Recovery(logger)(h)

	return h
}
