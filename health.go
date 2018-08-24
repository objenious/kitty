package kitty

import (
	"io"
	"net/http"
)

// Liveness defines the liveness handler.
func (t *HTTPTransport) Liveness(h http.HandlerFunc) *HTTPTransport {
	t.liveness = h
	return t
}

// Readiness defines the readiness handler.
func (t *HTTPTransport) Readiness(h http.HandlerFunc) *HTTPTransport {
	t.readiness = h
	return t
}

// defaultHealthcheck is a default health handler that returns a 200 status and an "OK" body
func defaultHealthcheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "OK")
}
