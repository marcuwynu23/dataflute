package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/marcuwynu23/dataserve/internal/api/response"
)

// Timeout aborts requests that exceed the given duration with a 504
// response. The deadline is propagated through the request context so
// handlers can cancel their work.
//
// Because the handler runs in a separate goroutine, panics are recovered
// there and re-panicked on the calling goroutine so that the outer
// Recovery middleware can still convert them into 500 responses.
func Timeout(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()

			done := make(chan struct{})
			var recovered any
			go func() {
				defer func() {
					if rec := recover(); rec != nil {
						recovered = rec
					}
					close(done)
				}()
				next.ServeHTTP(w, r.WithContext(ctx))
			}()

			select {
			case <-done:
				if recovered != nil {
					panic(recovered)
				}
			case <-ctx.Done():
				if rc := http.NewResponseController(w); rc != nil {
					_ = rc.SetWriteDeadline(time.Now())
				}
				response.Error(w, http.StatusGatewayTimeout, "request timeout")
			}
		})
	}
}
