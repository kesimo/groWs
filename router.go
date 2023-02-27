package groWs

import (
	"net/http"
)

type HandlerFunc func(http.ResponseWriter, *http.Request)

type Route struct {
	// Name of the route
	Path string
	// Handler function
	Handler ClientHandler
}

type Router struct {
	// internal list of routes
	routes []*Route
}

func NewRouter() *Router {
	return &Router{
		routes: make([]*Route, 0),
	}
}

// AddRoute adds a route to the router
func (r *Router) AddRoute(path string, handler ClientHandler) {
	r.routes = append(r.routes, &Route{
		Path:    path,
		Handler: handler,
	})
}

// GetRoutes returns all routes
func (r *Router) GetRoutes() []*Route {
	return r.routes
}
