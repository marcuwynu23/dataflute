package router

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/marcuwynu23/dataserve/internal/api/handlers"
	"github.com/marcuwynu23/dataserve/internal/api/response"
	"github.com/marcuwynu23/dataserve/internal/middleware"
	"github.com/marcuwynu23/dataserve/internal/service"
)

// New builds the HTTP handler for the application, wiring up middleware
// and routes. It is dependency-injected so it can be reused by tests
// without starting a real server.
func New(svc *service.Service, logger *slog.Logger, requestTimeout time.Duration) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recovery(logger))
	r.Use(middleware.RequestLogging(logger, svc.Metrics()))
	r.Use(middleware.CORS())
	r.Use(middleware.Timeout(requestTimeout))
	r.Use(middleware.Compress())

	r.Get("/", handlers.NewMetadataHandler(svc).Get)
	r.Get("/health", handlers.NewHealthHandler().Get)
	r.Get("/metrics", handlers.NewMetricsHandler(svc).Get)
	r.Get("/{name}", handlers.NewDatasetHandler(svc).Get)

	r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		response.Error(w, http.StatusNotFound, "dataset not found")
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Allow", "GET, OPTIONS")
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
	})

	return r
}
