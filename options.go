package kitty

import "github.com/go-kit/kit/transport/http"

// HTTPOptions defines the list of go-kit http.ServerOption to be added to all endpoints.
func (s *Server) HTTPOptions(opts ...http.ServerOption) *Server {
	s.opts = opts
	return s
}
