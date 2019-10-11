package kitty

import (
	"context"
	"net/http"
	"net/url"

	kithttp "github.com/go-kit/kit/transport/http"
)

// Client is a wrapper above the go-kit http client.
// It maps HTTP errors to Go errors.
// As the mapped error implements StatusCode, the returned status code will also be used
// as the status code returned by a go-kit HTTP endpoint.
// When using the backoff middleware, only 429 & 5XX errors trigger a retry.
type Client struct {
	*kithttp.Client
}

// NewClient creates a kitty client.
func NewClient(
	method string,
	tgt *url.URL,
	enc kithttp.EncodeRequestFunc,
	dec kithttp.DecodeResponseFunc,
	options ...kithttp.ClientOption,
) *Client {
	return &Client{Client: kithttp.NewClient(method, tgt, enc, makeDecodeResponseFunc(dec), options...)}
}

// NewClientWithError creates a kitty client that decode error.
func NewClientWithError(
	method string,
	tgt *url.URL,
	enc kithttp.EncodeRequestFunc,
	dec kithttp.DecodeResponseFunc,
	options ...kithttp.ClientOption,
) *Client {
	return &Client{Client: kithttp.NewClient(method, tgt, enc, dec, options...)}
}

// makeDecodeResponseFunc maps HTTP errors to Go errors.
func makeDecodeResponseFunc(fn kithttp.DecodeResponseFunc) kithttp.DecodeResponseFunc {
	return func(ctx context.Context, resp *http.Response) (interface{}, error) {
		if err := HTTPError(resp); err != nil {
			return nil, err
		}
		return fn(ctx, resp)
	}
}