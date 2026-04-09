package api

import (
	"context"
	"net/http"
	"strings"
)

// pathParamContextKey is the context key for path parameters.
type pathParamContextKey struct{}

// Router is an HTTP router that supports method-based routing and path parameters.
type Router struct {
	middleware []MiddlewareFunc
	prefix     string
	routes     []route
	handlers   map[string]http.Handler
}

type route struct {
	method  string
	pattern string
	handler http.HandlerFunc
	segments []string // Pattern segments for matching
	hasWildcard bool
}

// NewRouter creates a new router.
func NewRouter() *Router {
	return &Router{
		handlers: make(map[string]http.Handler),
		routes:   make([]route, 0),
	}
}

// ServeHTTP implements http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.serveHTTP(w, req)
	})
	for i := len(r.middleware) - 1; i >= 0; i-- {
		handler = r.middleware[i](handler)
	}

	handler.ServeHTTP(w, req)
}

// serveHTTP is the internal handler that finds and executes the registered handler.
func (r *Router) serveHTTP(w http.ResponseWriter, req *http.Request) {
	fullPath := r.prefix + req.URL.Path

	// Check for exact method+path match first
	key := req.Method + ":" + fullPath
	if h, ok := r.handlers[key]; ok {
		h.ServeHTTP(w, req)
		return
	}

	// Check pattern routes
	for _, route := range r.routes {
		if route.method != req.Method {
			continue
		}

		if params, ok := r.matchRoute(route, fullPath); ok {
			// Store params in request context
			req = req.WithContext(setPathParams(req.Context(), params))
			route.handler(w, req)
			return
		}
	}

	// No match found
	http.NotFound(w, req)
}

// matchRoute checks if the path matches the route pattern.
// Returns path parameters and whether it matched.
func (r *Router) matchRoute(route route, path string) (map[string]string, bool) {
	patternSegments := route.segments
	pathSegments := strings.Split(strings.Trim(path, "/"), "/")

	if len(patternSegments) != len(pathSegments) {
		return nil, false
	}

	params := make(map[string]string)

	for i, patSeg := range patternSegments {
		if strings.HasPrefix(patSeg, "{") && strings.HasSuffix(patSeg, "}") {
			// Wildcard parameter
			paramName := patSeg[1 : len(patSeg)-1]
			params[paramName] = pathSegments[i]
		} else if patSeg != pathSegments[i] {
			// Exact mismatch
			return nil, false
		}
	}

	return params, true
}

// Use adds middleware to the router.
func (r *Router) Use(middleware ...MiddlewareFunc) {
	r.middleware = append(r.middleware, middleware...)
}

// HandleFunc registers a handler function.
func (r *Router) HandleFunc(method, path string, handler http.HandlerFunc) {
	// Check if pattern has wildcards
	hasWildcard := strings.Contains(path, "{")
	segments := strings.Split(strings.Trim(path, "/"), "/")

	pattern := method + ":" + r.prefix + path
	r.handlers[pattern] = handler

	if hasWildcard {
		r.routes = append(r.routes, route{
			method:      method,
			pattern:     r.prefix + path,
			handler:     handler,
			segments:    segments,
			hasWildcard: true,
		})
	}
}

// GET registers a GET handler.
func (r *Router) GET(path string, handler http.HandlerFunc) {
	r.HandleFunc(http.MethodGet, path, handler)
}

// POST registers a POST handler.
func (r *Router) POST(path string, handler http.HandlerFunc) {
	r.HandleFunc(http.MethodPost, path, handler)
}

// PUT registers a PUT handler.
func (r *Router) PUT(path string, handler http.HandlerFunc) {
	r.HandleFunc(http.MethodPut, path, handler)
}

// DELETE registers a DELETE handler.
func (r *Router) DELETE(path string, handler http.HandlerFunc) {
	r.HandleFunc(http.MethodDelete, path, handler)
}

// PATCH registers a PATCH handler.
func (r *Router) PATCH(path string, handler http.HandlerFunc) {
	r.HandleFunc(http.MethodPatch, path, handler)
}

// OPTIONS registers an OPTIONS handler.
func (r *Router) OPTIONS(path string, handler http.HandlerFunc) {
	r.HandleFunc(http.MethodOptions, path, handler)
}

// PathParamContext is the context key for path parameters.
type PathParamContext struct{}

// setPathParams stores path parameters in context.
func setPathParams(ctx context.Context, params map[string]string) context.Context {
	return context.WithValue(ctx, PathParamContext{}, params)
}

// GetPathParam retrieves a path parameter from the request.
func GetPathParam(r *http.Request, key string) string {
	params, ok := r.Context().Value(PathParamContext{}).(map[string]string)
	if !ok {
		return ""
	}
	return params[key]
}

// Group creates a new router with a prefix.
func (r *Router) Group(prefix string) *Router {
	return &Router{
		middleware: r.middleware,
		prefix:     r.prefix + prefix,
		handlers:   r.handlers,
	}
}

// Handle registers a handler for all methods.
func (r *Router) Handle(path string, handler http.Handler) {
	r.handlers[r.prefix+path] = handler
}
