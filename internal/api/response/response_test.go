package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	JSON(rec, http.StatusOK, map[string]any{"a": 1}, false)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"a":1}`, rec.Body.String())
}

func TestJSONPretty(t *testing.T) {
	rec := httptest.NewRecorder()
	JSON(rec, http.StatusOK, map[string]any{"a": 1}, true)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "\n  ")
}

func TestError(t *testing.T) {
	rec := httptest.NewRecorder()
	Error(rec, http.StatusNotFound, "dataset not found")

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.JSONEq(t, `{"error":"dataset not found"}`, rec.Body.String())
}

func TestJSONMarshalFailure(t *testing.T) {
	rec := httptest.NewRecorder()
	JSON(rec, http.StatusOK, make(chan int), false)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.JSONEq(t, `{"error":"failed to encode response"}`, rec.Body.String())
}

func TestJSONNullValue(t *testing.T) {
	rec := httptest.NewRecorder()
	JSON(rec, http.StatusOK, nil, false)
	assert.JSONEq(t, `null`, rec.Body.String())
}
