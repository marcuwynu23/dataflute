package middleware

import (
	"log/slog"
	"net/http"

	"github.com/marcuwynu23/dataflute/internal/api/response"
)

// Recovery recovers from panics in downstream handlers, logs them, and
// returns a JSON 500 response.
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic recovered",
						"method", r.Method,
						"path", r.URL.Path,
						"panic", rec,
					)
					response.Error(w, http.StatusInternalServerError, "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
