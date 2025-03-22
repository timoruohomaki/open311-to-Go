package router

import (
	"context"
	"net/http"
	"strings"
)

// Route defines an HTTP route
type Route struct {
	Method  string
	Pattern string
	Handler http.HandlerFunc
}

// Router implements http.Handler and manages routes
type Router struct {
	routes     []Route
	middleware []func(http.Handler) http.Handler
}

// New creates a new router
func New() *Router {
	return &Router{
		routes:     []Route{},
		middleware: []func(http.Handler) http.Handler{},
	}
}

// Use adds middleware to the router
func (r *Router) Use(middleware func(http.Handler) http.Handler) {
	r.middleware = append(r.middleware, middleware)
}

// Handle adds a route to the router
func (r *Router) Handle(method, pattern string, handler http.HandlerFunc) {
	r.routes = append(r.routes, Route{
		Method:  method,
		Pattern: pattern,
		Handler: handler,
	})
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Find matching route
		for _, route := range r.routes {
			// Check HTTP method
			if route.Method != req.Method {
				continue
			}

			// Check if route pattern matches, with params extraction
			params, ok := matchRoute(route.Pattern, req.URL.Path)
			if !ok {
				continue
			}

			// Set path parameters in request context
			ctx := req.Context()
			for key, value := range params {
				ctx = context.WithValue(ctx, PathParamKey(key), value)
			}

			// Execute handler with updated context
			route.Handler.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		// No route matched
		http.NotFound(w, req)
	})

	// Apply middleware in reverse order (so the first in the list is the outermost)
	for i := len(r.middleware) - 1; i >= 0; i-- {
		handler = r.middleware[i](handler)
	}

	handler.ServeHTTP(w, req)
}

// PathParamKey type for context keys
type PathParamKey string

// matchRoute checks if a URL path matches a route pattern and extracts parameters
func matchRoute(pattern, path string) (map[string]string, bool) {
	// Split pattern and path into segments
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	// Quick length check
	if len(patternParts) != len(pathParts) {
		return nil, false
	}

	// Check each segment and extract parameters
	params := make(map[string]string)
	for i, patternPart := range patternParts {
		// Parameter segment {name}
		if len(patternPart) > 2 && patternPart[0] == '{' && patternPart[len(patternPart)-1] == '}' {
			// Extract parameter name
			paramName := patternPart[1 : len(patternPart)-1]
			params[paramName] = pathParts[i]
		} else if patternPart != pathParts[i] {
			// Static segment doesn't match
			return nil, false
		}
	}

	return params, true
}
