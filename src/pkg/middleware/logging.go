package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/timoruohomaki/open311-to-Go/pkg/logger"
)

// LoggingMiddleware logs HTTP requests in Apache Combined Log Format
func LoggingMiddleware(accessLog logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := newResponseWriter(w)
			next.ServeHTTP(rw, r)

			// Apache Combined Log Format:
			// %h %l %u %t "%r" %>s %b "%{Referer}i" "%{User-Agent}i"
			remoteHost := r.RemoteAddr
			logname := "-" // Not used
			user := "-"    // Not used unless you have auth
			timestamp := time.Now().Format("[02/Jan/2006:15:04:05 -0700]")
			requestLine := fmt.Sprintf("%s %s %s", r.Method, r.RequestURI, r.Proto)
			status := rw.statusCode
			size := rw.size
			referer := r.Referer()
			userAgent := r.UserAgent()

			accessLog.Infof("%s %s %s %s \"%s\" %d %d \"%s\" \"%s\"",
				remoteHost, logname, user, timestamp, requestLine, status, size, referer, userAgent)
		})
	}
}

// responseWriter is a wrapper for http.ResponseWriter that captures status code and size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK, 0}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}
