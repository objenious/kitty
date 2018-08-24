package backoff

import (
	"context"

	"github.com/cenkalti/backoff"
	"github.com/go-kit/kit/endpoint"
	"github.com/objenious/kitty"
)

// NewBackoff creates an exponential backoff middleware, based on github.com/cenkalti/backoff.
// Retries will be attempted if the returned error implements is retryable (see kitty.IsRetryable).
func NewBackoff(bo backoff.BackOff) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, finalerr error) {
			err := backoff.Retry(func() error {
				response, finalerr = next(ctx, request)
				if kitty.IsRetryable(finalerr) {
					return finalerr
				}
				return nil
			}, backoff.NewExponentialBackOff())

			if err != nil {
				finalerr = err
			}
			return
		}
	}
}
