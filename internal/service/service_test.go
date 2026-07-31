package service

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/marcuwynu23/dataserve/internal/config"
	"github.com/marcuwynu23/dataserve/internal/loader"
	"github.com/marcuwynu23/dataserve/internal/metrics"
	"github.com/marcuwynu23/dataserve/internal/parser"
	jsonparser "github.com/marcuwynu23/dataserve/internal/parser/json"
	"github.com/marcuwynu23/dataserve/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newService(t *testing.T, cfg *config.Config) *Service {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	reg := parser.NewRegistry()
	reg.Register(jsonparser.New())

	st := storage.New()
	return New(cfg, st, loader.New(reg, logger), metrics.New(), reg, logger)
}

func TestLoadInitial_DirectoryMode(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "users.json"), []byte(`[{"id":1}]`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "bad.json"), []byte(`{`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "health.json"), []byte(`[]`), 0o644))

	svc := newService(t, &config.Config{DataDir: dir, Host: "127.0.0.1", Port: 0})
	require.NoError(t, svc.LoadInitial())

	_, ok := svc.Dataset("users")
	assert.True(t, ok)
	// Reserved name is skipped; invalid file is skipped.
	_, ok = svc.Dataset("health")
	assert.False(t, ok)
	assert.Equal(t, []string{"/users"}, svc.AvailableEndpoints())
}

func TestLoadInitial_SingleFileMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "users.json")
	require.NoError(t, os.WriteFile(path, []byte(`[{"id":1}]`), 0o644))

	cfg := &config.Config{FilePath: path, Endpoint: "/employees", Host: "127.0.0.1", Port: 0}
	svc := newService(t, cfg)
	require.NoError(t, svc.LoadInitial())

	_, ok := svc.Dataset("employees")
	assert.True(t, ok)
	_, ok = svc.Dataset("users")
	assert.False(t, ok)
}

func TestLoadInitial_SingleFileModeError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	require.NoError(t, os.WriteFile(path, []byte(`{`), 0o644))

	cfg := &config.Config{FilePath: path, Endpoint: "/bad", Host: "127.0.0.1", Port: 0}
	svc := newService(t, cfg)
	require.Error(t, svc.LoadInitial())
}

func TestReloadFile_UpdatesDataset(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "users.json")
	require.NoError(t, os.WriteFile(path, []byte(`[{"id":1}]`), 0o644))

	svc := newService(t, &config.Config{DataDir: dir, Host: "127.0.0.1", Port: 0})
	require.NoError(t, svc.LoadInitial())

	require.NoError(t, os.WriteFile(path, []byte(`[{"id":2}]`), 0o644))
	require.NoError(t, svc.ReloadFile(path))

	d, ok := svc.Dataset("users")
	require.True(t, ok)
	items := d.Data.([]any)
	assert.Len(t, items, 1)
	assert.Equal(t, float64(2), toFloat(items[0].(map[string]any)["id"]))
	assert.Equal(t, int64(1), svc.MetricsSnapshot().Reloads)
}

func TestReloadFile_Error(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "users.json")
	require.NoError(t, os.WriteFile(path, []byte(`[]`), 0o644))

	svc := newService(t, &config.Config{DataDir: dir, Host: "127.0.0.1", Port: 0})
	require.NoError(t, svc.LoadInitial())

	require.NoError(t, os.WriteFile(path, []byte(`{`), 0o644))
	require.Error(t, svc.ReloadFile(path))
}

func TestRemoveFile_DirectoryMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "users.json")
	require.NoError(t, os.WriteFile(path, []byte(`[]`), 0o644))

	svc := newService(t, &config.Config{DataDir: dir, Host: "127.0.0.1", Port: 0})
	require.NoError(t, svc.LoadInitial())

	svc.RemoveFile(path)
	_, ok := svc.Dataset("users")
	assert.False(t, ok)
}

func TestRemoveFile_SingleModeNoop(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "users.json")
	require.NoError(t, os.WriteFile(path, []byte(`[]`), 0o644))

	cfg := &config.Config{FilePath: path, Endpoint: "/employees", Host: "127.0.0.1", Port: 0}
	svc := newService(t, cfg)
	require.NoError(t, svc.LoadInitial())

	svc.RemoveFile(path)
	_, ok := svc.Dataset("employees")
	assert.True(t, ok)
}

func TestServiceAccessors(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "users.json"), []byte(`[]`), 0o644))

	cfg := &config.Config{DataDir: dir, Host: "127.0.0.1", Port: 0, Pretty: true}
	svc := newService(t, cfg)
	require.NoError(t, svc.LoadInitial())

	assert.Equal(t, "directory", svc.Mode())
	assert.Equal(t, []string{".json"}, svc.Formats())
	assert.True(t, svc.Pretty())
	assert.NotNil(t, svc.Metrics())
	assert.Equal(t, 1, svc.MetricsSnapshot().Datasets)
}

func toFloat(v any) float64 {
	switch t := v.(type) {
	case json.Number:
		f, _ := t.Float64()
		return f
	case float64:
		return t
	}
	return 0
}
