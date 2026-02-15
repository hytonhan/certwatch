package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func LoggingMiddlewarefunc(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := uuid.NewString()

			ctx := context.WithValue(r.Context(), "request_id", requestID)
			r = r.WithContext(ctx)

			// TODO: statuscode needs a wrapper
			// rw := NewResponseWriter(w)

			next.ServeHTTP(w, r)

			logger.Info("http_request",
				"event_type", "http_access",
				"request_id", requestID,
				"remote_ip", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.Path,
				// "status_code", w.Header().s,
				"latency_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}
