package router

import (
	"fmt"
	"net/http"
	"strings"
)

// Handler is returned based on HTTP method and amount of path params.
// Example: GET /$table/$id, or GET /$table
// map holds handlers for an appropriate http method and amount of parameters separated by slash
type Router struct {
	getRoutes    map[int]http.HandlerFunc
	postRoutes   map[int]http.HandlerFunc
	putRoutes    map[int]http.HandlerFunc
	deleteRoutes map[int]http.HandlerFunc
}

func New() *Router {
	return &Router{
		getRoutes:    make(map[int]http.HandlerFunc),
		postRoutes:   make(map[int]http.HandlerFunc),
		putRoutes:    make(map[int]http.HandlerFunc),
		deleteRoutes: make(map[int]http.HandlerFunc),
	}
}

// RegisterRoute registers a route depending on amount of params in URL
// Example: 1 param - /$table?limit=5&offset=7, 2 params - /$table/$id
func (rt *Router) RegisterRoute(httpMethod string, paramsAmount int, hf http.HandlerFunc) error {
	switch httpMethod {
	case "GET":
		rt.getRoutes[paramsAmount] = hf

	case "POST":
		rt.postRoutes[paramsAmount] = hf

	case "PUT":
		rt.putRoutes[paramsAmount] = hf

	case "DELETE":
		rt.deleteRoutes[paramsAmount] = hf

	default:
		return fmt.Errorf("unsupported http method: %s", httpMethod)
	}

	return nil
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handler http.HandlerFunc
	paramsAmount := resolveParamsAmount(r.URL.Path)

	switch r.Method {
	case "GET":
		h, ok := rt.getRoutes[paramsAmount]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		handler = h

	case "POST":
		h, ok := rt.postRoutes[paramsAmount]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		handler = h

	case "PUT":
		h, ok := rt.putRoutes[paramsAmount]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		handler = h

	case "DELETE":
		h, ok := rt.deleteRoutes[paramsAmount]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		handler = h
	default:
		w.WriteHeader(http.StatusNotImplemented)
	}

	handler(w, r)
}

func resolveParamsAmount(urlPath string) int {
	if urlPath == "/" {
		return 0
	}

	split := strings.Split(urlPath, "/")
	// we do -1 because URL starts with slash and there is a redundant part
	return len(split) - 1
}
