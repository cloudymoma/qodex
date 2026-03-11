package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"gopkg.in/natefinch/lumberjack.v2"

	"qodex/internal/api"
	"qodex/internal/auth"
	"qodex/internal/config"
	"qodex/internal/graph"
	"qodex/internal/indexer"
	"qodex/internal/parser"
	"qodex/internal/repository"
	"qodex/internal/service"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	accessCode := flag.Bool("accesscode", false, "Enable access code protection")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load("conf.yaml")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Ensure directories exist
	dirs := []string{
		cfg.Data.Dir,
		cfg.Logging.Dir,
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	// Setup logger with rotation
	logger := setupLogger(cfg.Logging)

	logger.Debug("configuration loaded",
		"data_dir", cfg.Data.Dir,
		"log_dir", cfg.Logging.Dir,
		"log_level", cfg.Logging.Level,
		"log_format", cfg.Logging.Format,
		"rotation_max_size_mb", cfg.Logging.Rotation.MaxSizeMB,
		"rotation_max_age_days", cfg.Logging.Rotation.MaxAgeDays,
		"rotation_max_backups", cfg.Logging.Rotation.MaxBackups,
		"rotation_compress", cfg.Logging.Rotation.Compress,
	)

	// Initialize dependencies
	logger.Debug("initializing dependencies")

	repo := repository.NewGitRepository(logger)
	psr := parser.NewRegistry(cfg.Parser, logger)
	idx, err := indexer.NewBleveIndexer(cfg.Indexer, logger)
	if err != nil {
		return fmt.Errorf("create indexer: %w", err)
	}
	defer idx.Close()

	builder := graph.NewBuilder()
	ingestSvc := service.NewIngestService(cfg, repo, psr, idx, builder, logger)

	logger.Debug("all dependencies initialized")

	// Setup auth if enabled
	var authMgr *auth.Manager
	if *accessCode {
		authMgr = auth.NewManager()
		logger.Info("access code protection enabled")
	}

	// Setup HTTP server
	router := api.NewRouter(cfg, logger, ingestSvc, idx, authMgr)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("starting server",
			"addr", server.Addr,
			"data_dir", cfg.Data.Dir,
			"static_dir", cfg.Frontend.StaticDir,
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for interrupt or server error
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case sig := <-stop:
		logger.Info("received signal, shutting down", "signal", sig)
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	logger.Info("server stopped gracefully")
	return nil
}

func setupLogger(cfg config.LoggingConfig) *slog.Logger {
	level := parseLogLevel(cfg.Level)
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: level <= slog.LevelDebug,
	}

	// Setup log rotation writer
	logFile := &lumberjack.Logger{
		Filename:   filepath.Join(cfg.Dir, "qodex.log"),
		MaxSize:    cfg.Rotation.MaxSizeMB,
		MaxAge:     cfg.Rotation.MaxAgeDays,
		MaxBackups: cfg.Rotation.MaxBackups,
		Compress:   cfg.Rotation.Compress,
		LocalTime:  true,
	}

	// Write to both stdout and rotating log file
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(multiWriter, opts)
	} else {
		handler = slog.NewTextHandler(multiWriter, opts)
	}

	return slog.New(handler)
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "trace":
		return slog.Level(-8) // custom trace level below debug
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
