package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/marcuwynu23/dataflute/internal/api/response"
	"github.com/marcuwynu23/dataflute/internal/service"
	"github.com/marcuwynu23/dataflute/internal/utils"
)

// DatasetHandler serves individual datasets.
type DatasetHandler struct {
	service *service.Service
}

// NewDatasetHandler returns a DatasetHandler for the given service.
func NewDatasetHandler(s *service.Service) *DatasetHandler {
	return &DatasetHandler{service: s}
}

// Get serves the dataset named by the URL route parameter, applying the
// query features from the request URL.
func (h *DatasetHandler) Get(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	ds, ok := h.service.Dataset(name)
	if !ok {
		response.Error(w, http.StatusNotFound, "dataset not found")
		return
	}

	data, err := utils.ApplyQuery(ds.Data, r.URL.Query())
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid query")
		return
	}
	response.JSON(w, http.StatusOK, data, h.service.Pretty())
}
