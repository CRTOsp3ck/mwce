// internal/middleware/logging.go

package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

// NewLoggerMiddleware creates a new logger middleware
func NewLoggerMiddleware(logger zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Get request ID from the request context
			requestID := middleware.GetReqID(r.Context())

			// Create request start time
			start := time.Now()

			// Process the request
			defer func() {
				// Log the request after it's completed
				logger.Info().
					Str("request_id", requestID).
					Str("remote_addr", r.RemoteAddr).
					Str("method", r.Method).
					Str("url", r.URL.String()).
					Int("status", ww.Status()).
					Int("bytes", ww.BytesWritten()).
					Dur("elapsed", time.Since(start)).
					Msg("request completed")
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
