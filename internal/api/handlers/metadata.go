package handlers

import (
	"net/http"

	"github.com/marcuwynu23/dataflute/internal/api/response"
	"github.com/marcuwynu23/dataflute/internal/service"
)

// MetadataHandler serves the GET / endpoint.
type MetadataHandler struct {
	service *service.Service
}

// NewMetadataHandler returns a MetadataHandler for the given service.
func NewMetadataHandler(s *service.Service) *MetadataHandler {
	return &MetadataHandler{service: s}
}

// Get writes API metadata as JSON.
func (h *MetadataHandler) Get(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"mode":      h.service.Mode(),
		"formats":   h.service.Formats(),
		"available": h.service.AvailableEndpoints(),
	}
	response.JSON(w, http.StatusOK, resp, h.service.Pretty())
}
