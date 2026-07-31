package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/marcuwynu23/dataflute/internal/config"
	"github.com/marcuwynu23/dataflute/internal/loader"
	"github.com/marcuwynu23/dataflute/internal/metrics"
	"github.com/marcuwynu23/dataflute/internal/models"
	"github.com/marcuwynu23/dataflute/internal/parser"
	jsonparser "github.com/marcuwynu23/dataflute/internal/parser/json"
	"github.com/marcuwynu23/dataflute/internal/service"
	"github.com/marcuwynu23/dataflute/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestService(t *testing.T) *service.Service {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	reg := parser.NewRegistry()
	reg.Register(jsonparser.New())

	ld := loader.New(reg, logger)
	st := storage.New()
	st.Upsert("users", &models.Dataset{
		Name: "users",
		Data: []any{map[string]any{"id": json.Number("1"), "name": "John"}},
	})

	cfg := &config.Config{DataDir: ".", Host: "127.0.0.1", Port: 0}
	return service.New(cfg, st, ld, metrics.New(), reg, logger)
}

func doGet(t *testing.T, h http.HandlerFunc, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func doChiGet(t *testing.T, h http.HandlerFunc, path string) *httptest.ResponseRecorder {
	t.Helper()
	// Route through a minimal chi router so chi.URLParam works.
	r := chi.NewRouter()
	r.Get("/{name}", h)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestDatasetHandler_Get(t *testing.T) {
	h := NewDatasetHandler(newTestService(t))

	rec := doChiGet(t, h.Get, "/users")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `[{"id":1,"name":"John"}]`, rec.Body.String())

	rec = doChiGet(t, h.Get, "/missing")
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.JSONEq(t, `{"error":"dataset not found"}`, rec.Body.String())
}

func TestDatasetHandler_InvalidQuery(t *testing.T) {
	h := NewDatasetHandler(newTestService(t))
	rec := doChiGet(t, h.Get, "/users?page=abc")
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.JSONEq(t, `{"error":"invalid query"}`, rec.Body.String())
}

func TestMetadataHandler_Get(t *testing.T) {
	h := NewMetadataHandler(newTestService(t))
	rec := doGet(t, h.Get, "/")
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "directory", body["mode"])
	assert.Equal(t, []any{"/users"}, body["available"])
}

func TestHealthHandler_Get(t *testing.T) {
	h := NewHealthHandler()
	rec := doGet(t, h.Get, "/health")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"status":"ok"}`, rec.Body.String())
}

func TestMetricsHandler_Get(t *testing.T) {
	h := NewMetricsHandler(newTestService(t))
	rec := doGet(t, h.Get, "/metrics")
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, float64(1), body["datasets"])
	assert.NotEmpty(t, body["uptime"])
}
