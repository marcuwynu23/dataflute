package handlers

import (
	"net/http"

	"github.com/marcuwynu23/dataflute/internal/api/response"
	"github.com/marcuwynu23/dataflute/internal/service"
)

// MetricsHandler serves the GET /metrics endpoint.
type MetricsHandler struct {
	service *service.Service
}

// NewMetricsHandler returns a MetricsHandler for the given service.
func NewMetricsHandler(s *service.Service) *MetricsHandler {
	return &MetricsHandler{service: s}
}

// Get writes runtime statistics as JSON.
func (h *MetricsHandler) Get(w http.ResponseWriter, _ *http.Request) {
	snap := h.service.MetricsSnapshot()
	resp := map[string]any{
		"uptime":   snap.Uptime.String(),
		"requests": snap.Requests,
		"reloads":  snap.Reloads,
		"datasets": snap.Datasets,
	}
	response.JSON(w, http.StatusOK, resp, false)
}
