package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestAPIKeyMiddleware(t *testing.T) {
	handler := APIKeyMiddleware([]string{"secret1", "secret2"})(okHandler())

	cases := []struct {
		name   string
		method string
		key    string
		want   int
	}{
		{"GET is public", http.MethodGet, "", http.StatusOK},
		{"HEAD is public", http.MethodHead, "", http.StatusOK},
		{"POST without key rejected", http.MethodPost, "", http.StatusUnauthorized},
		{"POST with invalid key rejected", http.MethodPost, "nope", http.StatusUnauthorized},
		{"POST with valid key passes", http.MethodPost, "secret1", http.StatusOK},
		{"PUT with valid key passes", http.MethodPut, "secret2", http.StatusOK},
		{"DELETE without key rejected", http.MethodDelete, "", http.StatusUnauthorized},
		{"DELETE with valid key passes", http.MethodDelete, "secret1", http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/api/v1/services", nil)
			if tc.key != "" {
				req.Header.Set("X-API-Key", tc.key)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			assert.Equal(t, tc.want, rec.Code)
		})
	}
}

func TestAPIKeyMiddlewareDisabledWhenNoKeys(t *testing.T) {
	handler := APIKeyMiddleware(nil)(okHandler())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/services", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// With no configured keys, write auth is disabled and the request passes.
	assert.Equal(t, http.StatusOK, rec.Code)
}
