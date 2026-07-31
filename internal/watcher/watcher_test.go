package watcher_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/marcuwynu23/dataserve/internal/config"
	"github.com/marcuwynu23/dataserve/internal/loader"
	"github.com/marcuwynu23/dataserve/internal/metrics"
	"github.com/marcuwynu23/dataserve/internal/parser"
	jsonparser "github.com/marcuwynu23/dataserve/internal/parser/json"
	"github.com/marcuwynu23/dataserve/internal/service"
	"github.com/marcuwynu23/dataserve/internal/storage"
	"github.com/marcuwynu23/dataserve/internal/watcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestService(t *testing.T, cfg *config.Config) (*service.Service, *storage.Storage) {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	reg := parser.NewRegistry()
	reg.Register(jsonparser.New())

	ld := loader.New(reg, logger)
	st := storage.New()
	svc := service.New(cfg, st, ld, metrics.New(), reg, logger)
	return svc, st
}

func eventually(t *testing.T, timeout time.Duration, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}

func startWatcher(t *testing.T, ctx context.Context, cfg *config.Config, svc *service.Service) {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	reg := parser.NewRegistry()
	reg.Register(jsonparser.New())

	w, err := watcher.New(cfg, svc, reg, logger)
	require.NoError(t, err)

	errCh := make(chan error, 1)
	go func() { errCh <- w.Start(ctx) }()
	t.Cleanup(func() { require.NoError(t, <-errCh) })
}

func TestWatcher_DirectoryCreateReloadRemove(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{DataDir: dir, Host: "127.0.0.1", Port: 0}
	svc, st := newTestService(t, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startWatcher(t, ctx, cfg, svc)

	// Give the watcher goroutine time to establish the watch before
	// creating files, otherwise the Create event can be missed.
	time.Sleep(300 * time.Millisecond)

	path := filepath.Join(dir, "users.json")

	require.NoError(t, os.WriteFile(path, []byte(`[{"id":1}]`), 0o644))
	eventually(t, 5*time.Second, func() bool {
		_, ok := st.Get("users")
		return ok
	})

	require.NoError(t, os.WriteFile(path, []byte(`[{"id":2}]`), 0o644))
	eventually(t, 5*time.Second, func() bool {
		d, ok := st.Get("users")
		if !ok {
			return false
		}
		arr := d.Data.([]any)
		return len(arr) == 1 && arr[0].(map[string]any)["id"].(json.Number).String() == "2"
	})

	require.NoError(t, os.Remove(path))
	eventually(t, 5*time.Second, func() bool {
		_, ok := st.Get("users")
		return !ok
	})
}

func TestWatcher_SingleFileReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "users.json")
	require.NoError(t, os.WriteFile(path, []byte(`[{"id":1}]`), 0o644))

	cfg := &config.Config{FilePath: path, Endpoint: "/users", Host: "127.0.0.1", Port: 0}
	svc, st := newTestService(t, cfg)
	require.NoError(t, svc.LoadInitial())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startWatcher(t, ctx, cfg, svc)

	time.Sleep(300 * time.Millisecond)

	require.NoError(t, os.WriteFile(path, []byte(`[{"id":2}]`), 0o644))
	eventually(t, 5*time.Second, func() bool {
		d, ok := st.Get("users")
		if !ok {
			return false
		}
		arr := d.Data.([]any)
		return len(arr) == 1 && arr[0].(map[string]any)["id"].(json.Number).String() == "2"
	})
}

func TestWatcher_IgnoresUnsupportedFiles(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{DataDir: dir, Host: "127.0.0.1", Port: 0}
	svc, st := newTestService(t, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startWatcher(t, ctx, cfg, svc)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("hi"), 0o644))
	time.Sleep(300 * time.Millisecond)
	assert.Empty(t, st.Names())
}
