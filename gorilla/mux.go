package gorilla

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/objenious/kitty"
)

// router is a Router implementation for the gorilla/mux router.
type router struct {
	mux *mux.Router
}

var _ kitty.Router = &router{}

func Router() kitty.Router {
	return &router{mux.NewRouter()}
}

// Handle registers a handler to the router.
func (g *router) Handle(method, path string, h http.Handler) {
	g.mux.Handle(path, h).Methods(method)
}

// SetNotFoundHandler will sets the NotFound handler.
func (g *router) SetNotFoundHandler(h http.Handler) {
	g.mux.NotFoundHandler = h
}

// ServeHTTP dispatches the handler registered in the matched route.
func (g *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
}
