package kitty

import (
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

// httpendpoint encapsulates everything required to build
// an endpoint hosted on a kit server.
type httpendpoint struct {
	method, path string
	endpoint     endpoint.Endpoint
	decoder      httptransport.DecodeRequestFunc
	encoder      httptransport.EncodeResponseFunc
	options      []httptransport.ServerOption
}

// HTTPEndpointOption is an option for an HTTP endpoint
type HTTPEndpointOption func(*httpendpoint) *httpendpoint

// HTTPEndpoint registers an endpoint to a kitty.Server.
// Unless specified, the endpoint will use the POST method,
// NopRequestDecoder will decode the request (and do nothing),
// and EncodeJSONResponse will encode the response.
func (s *Server) HTTPEndpoint(method, path string, ep endpoint.Endpoint, opts ...HTTPEndpointOption) *Server {
	e := &httpendpoint{
		method:   method,
		path:     path,
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
