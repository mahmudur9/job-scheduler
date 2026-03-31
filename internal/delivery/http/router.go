package http

import (
	"fmt"
	"net/http"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request)

type Router struct {
	routes map[string]map[string]HandlerFunc
}

func NewRouter() *Router {
	return &Router{
		routes: make(map[string]map[string]HandlerFunc),
	}
}

func (r *Router) addRoute(method string, path string, handler HandlerFunc) {
	if r.routes[path] == nil {
		r.routes[path] = make(map[string]HandlerFunc)
	}
	r.routes[path][method] = handler
}

func (r *Router) GET(path string, handler HandlerFunc) {
	r.addRoute("GET", path, handler)
}

func (r *Router) POST(path string, handler HandlerFunc) {
	r.addRoute("POST", path, handler)
}

func (r *Router) PUT(path string, handler HandlerFunc) {
	r.addRoute("PUT", path, handler)
}

func (r *Router) DELETE(path string, handler HandlerFunc) {
	r.addRoute("DELETE", path, handler)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	method := req.Method
	path := req.URL.Path

	if methods, ok := r.routes[path]; ok {
		if handler, ok := methods[method]; ok {
			handler(w, req)
			return
		}

		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = fmt.Fprintf(w, "405 Method Not Allowed")
		return
	}
	w.WriteHeader(http.StatusNotFound)
	_, _ = fmt.Fprintf(w, "404 Not Found")
}
