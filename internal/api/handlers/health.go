package handlers

import (
	"net/http"

	"github.com/marcuwynu23/dataserve/internal/api/response"
)

// HealthHandler serves the GET /health endpoint.
type HealthHandler struct{}

// NewHealthHandler returns a new HealthHandler.
func NewHealthHandler() *HealthHandler { return &HealthHandler{} }

// Get writes a liveness probe response.
func (*HealthHandler) Get(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"}, false)
}
