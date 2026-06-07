package middleware

import (
	"net/http"
	"strings"

	"github.com/timoruohomaki/open311-to-Go/pkg/httputil"
)

// ContentTypeMiddleware handles content negotiation based on Accept header
func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acceptHeader := r.Header.Get("Accept")

		// Default to JSON if no Accept header
		if acceptHeader == "" {
			r.Header.Set("Accept", "application/json")
		} else if strings.Contains(acceptHeader, "application/xml") {
			r.Header.Set("Accept", "application/xml")
		} else {
			r.Header.Set("Accept", "application/json")
		}

		// Check and parse Content-Type for requests with body
		contentType := r.Header.Get("Content-Type")
		if r.Method == "POST" || r.Method == "PUT" {
			if contentType == "" {
				_ = httputil.SendError(w, r, http.StatusBadRequest, "Content-Type header is required")
				return
			}

			if !strings.Contains(contentType, "application/json") && !strings.Contains(contentType, "application/xml") {
				_ = httputil.SendError(w, r, http.StatusUnsupportedMediaType, "Content-Type must be application/json or application/xml")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
