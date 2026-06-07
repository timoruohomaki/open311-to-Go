package middleware

import (
	"net/http"
	"strings"

	"github.com/timoruohomaki/open311-to-Go/pkg/httputil"
)

// APIKeyMiddleware enforces X-API-Key authentication on write requests
// (POST/PUT/PATCH/DELETE). Read requests (GET/HEAD/OPTIONS) are always public,
// matching the Open311 model where service/request reads are open.
//
// If allowedKeys is empty, authentication is disabled and all requests pass —
// the caller should warn when starting in that mode.
func APIKeyMiddleware(allowedKeys []string) func(http.Handler) http.Handler {
	keySet := make(map[string]struct{}, len(allowedKeys))
	for _, k := range allowedKeys {
		if k = strings.TrimSpace(k); k != "" {
			keySet[k] = struct{}{}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(keySet) == 0 || !isWriteMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get("X-API-Key")
			if _, ok := keySet[key]; key == "" || !ok {
				_ = httputil.SendError(w, r, http.StatusUnauthorized, "missing or invalid API key")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isWriteMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}
