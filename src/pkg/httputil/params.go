package httputil

import (
	"net/http"

	"github.com/timoruohomaki/open311-to-Go/pkg/router"
)

// GetPathParam retrieves a path parameter from the request context
func GetPathParam(r *http.Request, name string) string {
	value := r.Context().Value(router.PathParamKey(name))
	if value == nil {
		return ""
	}
	return value.(string)
}
