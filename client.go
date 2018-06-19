package kitty

import (
	"context"
	"net/http"
	"net/url"

	httptransport "github.com/go-kit/kit/transport/http"
)

// Client is a wrapper above the go-kit http client.
// It maps HTTP errors to Go errors.
// As the mapped error implements StatusCode, the returned status code will also be used
// as the status code returned by a go-kit HTTP endpoint.
// When using the backoff middleware, only 429 & 5XX errors trigger a retry.
type Client struct {
	*httptransport.Client
}

// NewClient creates a kitty client.
func NewClient(
	method string,
	tgt *url.URL,
	enc httptransport.EncodeRequestFunc,
	dec httptransport.DecodeResponseFunc,
	options ...httptransport.ClientOption,
) *Client {
	return &Client{Client: httptransport.NewClient(method, tgt, enc, makeDecodeResponseFunc(dec), options...)}
}

// makeDecodeResponseFunc maps HTTP errors to Go errors.
func makeDecodeResponseFunc(fn httptransport.DecodeResponseFunc) httptransport.DecodeResponseFunc {
	return func(ctx context.Context, resp *http.Response) (interface{}, error) {
		if err := HTTPError(resp); err != nil {
			return nil, err
		}
		return fn(ctx, resp)
	}
}
