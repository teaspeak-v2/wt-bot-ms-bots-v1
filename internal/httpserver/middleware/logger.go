package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(ww, r)
			if logger != nil {
				logger.Info("request",
					"method", r.Method,
					"path", r.URL.Path,
					"status", ww.statusCode,
					"duration", time.Since(start).String(),
					"request_id", RequestIDFromContext(r.Context()),
				)
			}
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
