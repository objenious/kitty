package circuitbreaker

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/objenious/kitty"
	"github.com/sony/gobreaker"
)

type cbResponse struct {
	err error
	res interface{}
}

// NewCircuitBreaker creates a circuit breaker middleware, based on github.com/sony/gobreaker.
// CircuitBreaker will only trigger on retryable errors (see kitty.IsRetryable).
func NewCircuitBreaker(cb *gobreaker.CircuitBreaker) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			res, err := cb.Execute(func() (interface{}, error) {
				res, err := next(ctx, request)
				if kitty.IsRetryable(err) {
					return res, err
				}
				return cbResponse{res: res, err: err}, nil
			})
			if err == gobreaker.ErrOpenState || err == gobreaker.ErrTooManyRequests {
				return nil, kitty.Retryable(err)
			}
			if cbres, ok := res.(cbResponse); ok {
				return cbres.res, cbres.err
			}
			return res, err
		}
	}
}
