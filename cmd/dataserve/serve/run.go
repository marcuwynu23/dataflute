package serve

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/marcuwynu23/dataserve/internal/api/router"
	"github.com/marcuwynu23/dataserve/internal/config"
	"github.com/marcuwynu23/dataserve/internal/loader"
	"github.com/marcuwynu23/dataserve/internal/metrics"
	"github.com/marcuwynu23/dataserve/internal/parser"
	"github.com/marcuwynu23/dataserve/internal/parser/csv"
	"github.com/marcuwynu23/dataserve/internal/parser/ini"
	"github.com/marcuwynu23/dataserve/internal/parser/json"
	"github.com/marcuwynu23/dataserve/internal/parser/toml"
	"github.com/marcuwynu23/dataserve/internal/parser/tsv"
	"github.com/marcuwynu23/dataserve/internal/parser/xml"
	"github.com/marcuwynu23/dataserve/internal/parser/yaml"
	"github.com/marcuwynu23/dataserve/internal/service"
	"github.com/marcuwynu23/dataserve/internal/storage"
	"github.com/marcuwynu23/dataserve/internal/watcher"
)

const (
	requestTimeout    = 30 * time.Second
	shutdownTimeout   = 10 * time.Second
	readHeaderTimeout = 5 * time.Second
)

// run wires the application together and starts the HTTP server,
// blocking until the context is cancelled or the server fails.
func run(ctx context.Context, cfg *config.Config) error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if err := resolvePaths(cfg); err != nil {
		return err
	}

	registry := parser.NewRegistry()
	registry.Register(json.New())
	registry.Register(csv.New())
	registry.Register(tsv.New())
	registry.Register(xml.New())
	registry.Register(yaml.New())
	registry.Register(toml.New())
	registry.Register(ini.New())

	ld := loader.New(registry, logger)
	st := storage.New()
	met := metrics.New()
	svc := service.New(cfg, st, ld, met, registry, logger)

	if err := svc.LoadInitial(); err != nil {
		return fmt.Errorf("failed to load data: %w", err)
	}

	w, err := watcher.New(cfg, svc, registry, logger)
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	go func() {
		if err := w.Start(ctx); err != nil {
			logger.Error("file watcher stopped", "error", err)
		}
	}()

	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           router.New(svc, logger, requestTimeout),
		ReadHeaderTimeout: readHeaderTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server starting",
			"addr", cfg.Addr(),
			"mode", cfg.Mode(),
			"data", dataSource(cfg),
			"formats", svc.Formats(),
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
	}

	logger.Info("shutdown initiated")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}
	logger.Info("shutdown complete")
	return nil
}

// resolvePaths converts data paths to absolute paths so the file watcher
// can reliably compare event paths against the configured target.
func resolvePaths(cfg *config.Config) error {
	if cfg.DataDir != "" {
		abs, err := filepath.Abs(cfg.DataDir)
		if err != nil {
			return fmt.Errorf("resolve data directory: %w", err)
		}
		cfg.DataDir = abs
	}
	if cfg.FilePath != "" {
		abs, err := filepath.Abs(cfg.FilePath)
		if err != nil {
			return fmt.Errorf("resolve data file: %w", err)
		}
		cfg.FilePath = abs
	}
	return nil
}

func dataSource(cfg *config.Config) string {
	if cfg.Mode() == config.ModeSingle {
		return cfg.FilePath
	}
	return cfg.DataDir
}
