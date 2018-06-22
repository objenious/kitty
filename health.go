package kitty

import (
	"io"
	"net/http"
)

// Liveness defines the liveness handler.
func (s *Server) Liveness(h http.HandlerFunc) *Server {
	s.liveness = h
	return s
}

// Readiness defines the readiness handler.
func (s *Server) Readiness(h http.HandlerFunc) *Server {
	s.readiness = h
	return s
}

// defaultHealthcheck is a default health handler that returns a 200 status and an "OK" body
func defaultHealthcheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "OK")
}
