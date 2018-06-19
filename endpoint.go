package kitty

import (
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

// httpendpoint encapsulates everything required to build
// an endpoint hosted on a kit server.
type httpendpoint struct {
	methods, paths []string
	endpoint       endpoint.Endpoint
	decoder        httptransport.DecodeRequestFunc
	encoder        httptransport.EncodeResponseFunc
	options        []httptransport.ServerOption
}

// HTTPEndpointOption is an option for an HTTP endpoint
type HTTPEndpointOption func(*httpendpoint) *httpendpoint

// HTTPEndpoint registers an endpoint to a kitty.Server.
// Unless specified, the endpoint will use the POST method,
// NopRequestDecoder will decode the request (and do nothing),
// and EncodeJSONResponse will encode the response.
func (s *Server) HTTPEndpoint(ep endpoint.Endpoint, opts ...HTTPEndpointOption) *Server {
	e := &httpendpoint{
		methods:  []string{"POST"},
		endpoint: ep,
		decoder:  httptransport.NopRequestDecoder,
		encoder:  httptransport.EncodeJSONResponse,
	}
	for _, opt := range opts {
		e = opt(e)
	}
	s.endpoints = append(s.endpoints, e)
	return s
}

// Method defines the method(s) used to call the endpoint.
func Method(methods ...string) HTTPEndpointOption {
	return func(e *httpendpoint) *httpendpoint {
		e.methods = methods
		return e
	}
}

// Path defines the path(s) on which the endpoint will use.
func Path(paths ...string) HTTPEndpointOption {
	return func(e *httpendpoint) *httpendpoint {
		e.paths = paths
		return e
	}
}

// Decoder defines the request decoder for an endpoint.
// If none is provided, NopRequestDecoder is used.
func Decoder(dec httptransport.DecodeRequestFunc) HTTPEndpointOption {
	return func(e *httpendpoint) *httpendpoint {
		e.decoder = dec
		return e
	}
}

// Encoder defines the response encoder for an endpoint.
// If none is provided, EncodeJSONResponse is used.
func Encoder(enc httptransport.EncodeResponseFunc) HTTPEndpointOption {
	return func(e *httpendpoint) *httpendpoint {
		e.encoder = enc
		return e
	}
}

// ServerOptions defines a liste of go-kit ServerOption to be used by the endpoint.
func ServerOptions(opts ...httptransport.ServerOption) HTTPEndpointOption {
	return func(e *httpendpoint) *httpendpoint {
		e.options = opts
		return e
	}
}
