package router

import (
	"log"
	"net/http"
	"strings"
)

// Handler is returned based on HTTP method and amount of path params.
// Example: GET /$table/$id, or GET /$table
type Router struct {
	getRoutes map[int]http.HandlerFunc
}

func New() *Router {
	return &Router{getRoutes: make(map[string]*route)}
}

func (rt *Router) RegisterRoute(path string, handler http.HandlerFunc, methods ...string) {
	if path == "" {
		return
	}

	route := &route{}
	if path == "/" {
		route.path = path
	} else {
		route.path = usePlaceholders(path)
	}

	route.methods = methods
	route.handler = handler

	rt.getRoutes[path] = route
}

func usePlaceholders(path string) string {
	pathParts := strings.Split(path, "/")

	for _, p := range pathParts {
		if strings.HasPrefix(p, "{") || strings.HasPrefix(p, "}") {
			p = placeholder
		}
	}

	return strings.Join(pathParts, "/")
}

func (rt *Router) getHandler(path string) http.HandlerFunc {
	path = usePlaceholders(path)
	route, ok := rt.getRoutes[path]
	if !ok {
		log.Printf("path not registered %s", path)
		return nil
	}

	return route.handler
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := rt.getHandler(r.URL.Path)
	if handler == nil {
		w.WriteHeader(http.StatusNotFound)
	}

	handler(w, r)
}
