package middleware

import (
	"net/http"
	"time"

	"github.com/timoruohomaki/open311-to-Go/pkg/logger"
)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a custom response writer to capture status code
			rw := newResponseWriter(w)

			// Process request
			next.ServeHTTP(rw, r)

			// Log request details
			log.Infof("%s %s %s %d %s %s",
				r.Method,
				r.RequestURI,
				r.RemoteAddr,
				rw.statusCode,
				time.Since(start),
				r.Header.Get("Accept"),
			)
		})
	}
}

// responseWriter is a wrapper for http.ResponseWriter that captures status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
