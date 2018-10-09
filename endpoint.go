package kitty

import (
	"context"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
)

// httpendpoint encapsulates everything required to build
// an endpoint hosted on a kit server.
type httpendpoint struct {
	method, path string
	endpoint     endpoint.Endpoint
	decoder      kithttp.DecodeRequestFunc
	encoder      kithttp.EncodeResponseFunc
	options      []kithttp.ServerOption
}

// HTTPEndpointOption is an option for an HTTP endpoint
type HTTPEndpointOption func(*httpendpoint) *httpendpoint

// Endpoint registers an endpoint to a kitty.HTTPTransport.
// Unless specified, NopRequestDecoder will decode the request (and do nothing),
// and EncodeJSONResponse will encode the response.
func (t *HTTPTransport) Endpoint(method, path string, ep endpoint.Endpoint, opts ...HTTPEndpointOption) *HTTPTransport {
	e := &httpendpoint{
		method:   method,
		path:     path,
		endpoint: ep,
		decoder:  kithttp.NopRequestDecoder,
	}
	for _, opt := range opts {
		e = opt(e)
	}
	t.endpoints = append(t.endpoints, e)
	return t
}

type decoderError struct {
	error
}

func (e decoderError) StatusCode() int {
	if err, ok := e.error.(kithttp.StatusCoder); ok {
		return err.StatusCode()
	}
	return http.StatusBadRequest
}

// Decoder defines the request decoder for a HTTP endpoint.
// If none is provided, NopRequestDecoder is used.
func Decoder(dec kithttp.DecodeRequestFunc) HTTPEndpointOption {
	return func(e *httpendpoint) *httpendpoint {
		e.decoder = func(ctx context.Context, r *http.Request) (interface{}, error) {
			request, err := dec(ctx, r)
			if err != nil {
				return nil, decoderError{error: err}
			}
			return request, nil
		}
		return e
	}
}

// Encoder defines the response encoder for a HTTP endpoint.
// If none is provided, EncodeJSONResponse is used.
func Encoder(enc kithttp.EncodeResponseFunc) HTTPEndpointOption {
	return func(e *httpendpoint) *httpendpoint {
		e.encoder = enc
		return e
	}
}

// ServerOptions defines a liste of go-kit ServerOption to be used by a HTTP endpoint.
func ServerOptions(opts ...kithttp.ServerOption) HTTPEndpointOption {
	return func(e *httpendpoint) *httpendpoint {
		e.options = opts
		return e
	}
}
