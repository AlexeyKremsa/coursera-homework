package router

import (
	"net/http"
	"strings"
)

const matchAny = "anyRoute"

type Route struct {
	methods   map[string]bool
	url       string
	handler   http.Handler
	subroutes map[string]*Route
}

type Router struct {
	routes map[string]*Route
}

func New() *Router {
	return &Router{routes: make(map[string]*Route)}
}

func Init(r *Router) {
	r.AddRoute().Path("/").Method("GET")
}

func (r *Router) AddRoute() *Route {
	return &Route{}
}

func (r *Route) Method(methods ...string) *Route {
	for _, val := range methods {
		r.methods[val] = true
	}

	return r
}

func (r *Route) Path(path string) *Route {
	if path == "" {
		return r
	}

	pathParts := strings.Split(path, "/")

	for _, p := range pathParts {
		if strings.HasPrefix(p, "{") || strings.HasPrefix(p, "}") {
			p = matchAny
		}

	}

	return nil
}
