package router_test

import (
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const usersJSON = `[
  {"id":1,"name":"John","country":"PH","score":10},
  {"id":2,"name":"Jane","country":"US","score":20},
  {"id":3,"name":"Zoe","country":"PH","score":5}
]`

func newTestServer(t *testing.T, cfg *config.Config) http.Handler {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	reg := parser.NewRegistry()
	reg.Register(jsonparser.New())

	ld := loader.New(reg, logger)
	st := storage.New()
	met := metrics.New()
	svc := service.New(cfg, st, ld, met, reg, logger)
	require.NoError(t, svc.LoadInitial())
	return router.New(svc, logger, 5*time.Second)
}

func directoryServer(t *testing.T) (http.Handler, *config.Config) {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "users.json"), []byte(usersJSON), 0o644))

	cfg := &config.Config{
		DataDir: dir,
		Host:    "127.0.0.1",
		Port:    0,
	}
	return newTestServer(t, cfg), cfg
}

func doRequest(t *testing.T, h http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) any {
	t.Helper()
	dec := json.NewDecoder(rec.Body)
	dec.UseNumber()
	var v any
	require.NoError(t, dec.Decode(&v))
	return v
}

func TestGetMetadata(t *testing.T) {
	h, _ := directoryServer(t)
	rec := doRequest(t, h, http.MethodGet, "/")
	assert.Equal(t, http.StatusOK, rec.Code)

	body := decodeBody(t, rec).(map[string]any)
	assert.Equal(t, "directory", body["mode"])
	assert.Equal(t, []any{".json"}, body["formats"])
	assert.Equal(t, []any{"/users"}, body["available"])
}

func TestGetHealth(t *testing.T) {
	h, _ := directoryServer(t)
	rec := doRequest(t, h, http.MethodGet, "/health")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"status":"ok"}`, rec.Body.String())
}

func TestGetMetrics(t *testing.T) {
	h, _ := directoryServer(t)
	rec := doRequest(t, h, http.MethodGet, "/metrics")
	assert.Equal(t, http.StatusOK, rec.Code)

	body := decodeBody(t, rec).(map[string]any)
	assert.Equal(t, json.Number("1"), body["datasets"])
	assert.Equal(t, json.Number("1"), body["requests"])
	assert.NotEmpty(t, body["uptime"])
}

func TestGetDataset(t *testing.T) {
	h, _ := directoryServer(t)
	rec := doRequest(t, h, http.MethodGet, "/users")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, usersJSON, rec.Body.String())
}

func TestGetDatasetFilter(t *testing.T) {
	h, _ := directoryServer(t)
	rec := doRequest(t, h, http.MethodGet, "/users?country=PH")
	assert.Equal(t, http.StatusOK, rec.Code)

	items := decodeBody(t, rec).([]any)
	assert.Len(t, items, 2)
}

func TestGetDatasetSortDesc(t *testing.T) {
	h, _ := directoryServer(t)
	rec := doRequest(t, h, http.MethodGet, "/users?sort=-score")
	assert.Equal(t, http.StatusOK, rec.Code)

	items := decodeBody(t, rec).([]any)
	assert.Equal(t, "Jane", items[0].(map[string]any)["name"])
}

func TestGetDatasetPagination(t *testing.T) {
	h, _ := directoryServer(t)
	rec := doRequest(t, h, http.MethodGet, "/users?page=2&limit=1")
	assert.Equal(t, http.StatusOK, rec.Code)

	items := decodeBody(t, rec).([]any)
	require.Len(t, items, 1)
	assert.Equal(t, "Jane", items[0].(map[string]any)["name"])
}

func TestGetDatasetFields(t *testing.T) {
	h, _ := directoryServer(t)
	rec := doRequest(t, h, http.MethodGet, "/users?fields=name")
	assert.Equal(t, http.StatusOK, rec.Code)

	items := decodeBody(t, rec).([]any)
	obj := items[0].(map[string]any)
	assert.Equal(t, "John", obj["name"])
	assert.NotContains(t, obj, "country")
}

func TestGetDatasetSearch(t *testing.T) {
	h, _ := directoryServer(t)
	rec := doRequest(t, h, http.MethodGet, "/users?q=zoe")
	assert.Equal(t, http.StatusOK, rec.Code)

	items := decodeBody(t, rec).([]any)
	assert.Len(t, items, 1)
}

func TestGetDatasetNotFound(t *testing.T) {
	h, _ := directoryServer(t)
	rec := doRequest(t, h, http.MethodGet, "/nope")
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.JSONEq(t, `{"error":"dataset not found"}`, rec.Body.String())
}

func TestMethodNotAllowed(t *testing.T) {
	h, _ := directoryServer(t)
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		rec := doRequest(t, h, method, "/users")
		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code, method)
		assert.JSONEq(t, `{"error":"method not allowed"}`, rec.Body.String())
		assert.Contains(t, rec.Header().Get("Allow"), "GET")
	}
}

func TestInvalidQuery(t *testing.T) {
	h, _ := directoryServer(t)
	rec := doRequest(t, h, http.MethodGet, "/users?page=abc")
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.JSONEq(t, `{"error":"invalid query"}`, rec.Body.String())
}

func TestSingleFileMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "employees.json")
	require.NoError(t, os.WriteFile(path, []byte(`[{"id":1}]`), 0o644))

	cfg := &config.Config{
		FilePath: path,
		Endpoint: "/employees",
		Host:     "127.0.0.1",
		Port:     0,
	}
	h := newTestServer(t, cfg)

	rec := doRequest(t, h, http.MethodGet, "/employees")
	assert.Equal(t, http.StatusOK, rec.Code)

	rec = doRequest(t, h, http.MethodGet, "/")
	body := decodeBody(t, rec).(map[string]any)
	assert.Equal(t, "single", body["mode"])
	assert.Equal(t, []any{"/employees"}, body["available"])
}

func TestPrettyOutput(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "users.json"), []byte(usersJSON), 0o644))

	cfg := &config.Config{DataDir: dir, Host: "127.0.0.1", Port: 0, Pretty: true}
	h := newTestServer(t, cfg)

	rec := doRequest(t, h, http.MethodGet, "/users")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "\n  ")
}
