package integration

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/marcuwynu23/dataserve/internal/api/router"
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

// sampleDataDir resolves the repository's sample datasets.
func sampleDataDir(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", "pkg", "data"))
	require.NoError(t, err)
	return abs
}

func buildApp(t *testing.T, cfg *config.Config) (*service.Service, http.Handler) {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	reg := parser.NewRegistry()
	reg.Register(jsonparser.New())

	ld := loader.New(reg, logger)
	st := storage.New()
	met := metrics.New()
	svc := service.New(cfg, st, ld, met, reg, logger)
	require.NoError(t, svc.LoadInitial())
	return svc, router.New(svc, logger, 10*time.Second)
}

func doGet(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestEndToEndDirectoryMode(t *testing.T) {
	cfg := &config.Config{DataDir: sampleDataDir(t), Host: "127.0.0.1", Port: 0}
	_, h := buildApp(t, cfg)

	// Metadata lists the three sample datasets.
	rec := doGet(t, h, "/")
	require.Equal(t, http.StatusOK, rec.Code)
	var meta struct {
		Mode      string   `json:"mode"`
		Formats   []string `json:"formats"`
		Available []string `json:"available"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &meta))
	assert.Equal(t, "directory", meta.Mode)
	assert.Equal(t, []string{".json"}, meta.Formats)
	assert.Equal(t, []string{"/config", "/products", "/users"}, meta.Available)

	// Object dataset.
	rec = doGet(t, h, "/config")
	require.Equal(t, http.StatusOK, rec.Code)
	var cfgJSON map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &cfgJSON))
	assert.Equal(t, "1.0.0", cfgJSON["version"])

	// Array dataset with query features.
	rec = doGet(t, h, "/users?country=PH&sort=-id&fields=id,name&limit=2")
	require.Equal(t, http.StatusOK, rec.Code)
	var users []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &users))
	require.Len(t, users, 2)
	assert.Equal(t, float64(7), users[0]["id"])

	// Array dataset with filtering.
	rec = doGet(t, h, "/products?category=electronics")
	require.Equal(t, http.StatusOK, rec.Code)
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &users))
	assert.Len(t, users, 3)
}

func TestEndToEndSingleFileMode(t *testing.T) {
	usersPath := filepath.Join(sampleDataDir(t), "users.json")
	cfg := &config.Config{
		FilePath: usersPath,
		Endpoint: "/employees",
		Host:     "127.0.0.1",
		Port:     0,
	}
	svc, h := buildApp(t, cfg)

	rec := doGet(t, h, "/employees")
	require.Equal(t, http.StatusOK, rec.Code)
	var employees []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &employees))
	assert.NotEmpty(t, employees)

	rec = doGet(t, h, "/users")
	assert.Equal(t, http.StatusNotFound, rec.Code)

	assert.Equal(t, "single", svc.Mode())
}

func TestEndToEndWatcherReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "notes.json")
	require.NoError(t, os.WriteFile(path, []byte(`[{"id":1}]`), 0o644))

	cfg := &config.Config{DataDir: dir, Host: "127.0.0.1", Port: 0}
	svc, h := buildApp(t, cfg)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	reg := parser.NewRegistry()
	reg.Register(jsonparser.New())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w, err := watcher.New(cfg, svc, reg, logger)
	require.NoError(t, err)
	go func() { _ = w.Start(ctx) }()

	// Give the watcher time to establish its watch before rewriting.
	time.Sleep(300 * time.Millisecond)

	rec := doGet(t, h, "/notes")
	require.Equal(t, http.StatusOK, rec.Code)

	require.NoError(t, os.WriteFile(path, []byte(`[{"id":2}]`), 0o644))

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		rec := doGet(t, h, "/notes")
		if rec.Code == http.StatusOK {
			var notes []map[string]any
			_ = json.Unmarshal(rec.Body.Bytes(), &notes)
			if len(notes) == 1 && notes[0]["id"] == float64(2) {
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("dataset was not reloaded in time")
}
