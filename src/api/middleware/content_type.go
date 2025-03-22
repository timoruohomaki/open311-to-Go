package middleware

import (
	"net/http"
	"strings"
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
				http.Error(w, "Content-Type header is required", http.StatusBadRequest)
				return
			}

			if !strings.Contains(contentType, "application/json") && !strings.Contains(contentType, "application/xml") {
				http.Error(w, "Content-Type must be application/json or application/xml", http.StatusUnsupportedMediaType)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
