package kitty

import kithttp "github.com/go-kit/kit/transport/http"

// Options defines the list of go-kit http.ServerOption to be added to all endpoints.
func (t *HTTPTransport) Options(opts ...kithttp.ServerOption) *HTTPTransport {
	t.opts = opts
	return t
}
