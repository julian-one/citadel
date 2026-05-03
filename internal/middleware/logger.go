package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logger returns middleware that logs each request after the handler completes,
// including method, path, status, duration, and a generated request ID.
func Logger(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			requestID := uuid.New().String()
			requestLogger := logger.With(slog.String("request_id", requestID))
			w.Header().Set("X-Request-ID", requestID)

			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)

			requestLogger.Info("http request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", GetClientIP(r),
				"status", rw.status,
				"duration", time.Since(start),
			)
		})
	}
}
