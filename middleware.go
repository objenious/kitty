package kitty

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
)

// nopMiddleware is the default middleware, and does nothing.
func nopMiddleware(e endpoint.Endpoint) endpoint.Endpoint {
	return e
}

// Middlewares defines the list of endpoint middlewares to be added to all endpoints.
func (s *Server) Middlewares(m ...endpoint.Middleware) *Server {
	s.middleware = func(next endpoint.Endpoint) endpoint.Endpoint {
		for i := len(m) - 1; i >= 0; i-- {
			next = m[i](next)
		}
		return next
	}
	return s
}

// LogOption is a LogEndpoint middleware option.
type LogOption int

const (
	// LogRequest logs the request.
	LogRequest LogOption = iota
	// LogResponse logs the response.
	LogResponse
	// LogErrors logs the request in case of an error.
	LogErrors
)

// LogEndpoint creates a middleware that logs Endpoint calls.
// If LogRequest is specified, the endpoint request will be logged before the endpoint is called.
// If LogResponse is specified, the endpoint response will be logged after.
// If LogErrors is specified, the endpoint request will be logged if the endpoint returns an error.
// With LogResponse and LogErrors, the endpoint duration and result HTTP status code will be added to logs.
func LogEndpoint(fields ...LogOption) endpoint.Middleware {
	opts := map[LogOption]bool{}
	for _, f := range fields {
		opts[f] = true
	}
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			if opts[LogRequest] {
				_ = LogMessage(ctx, fmt.Sprintf("request: %+v", request))
			}
			start := time.Now()
			response, err = e(ctx, request)
			code := http.StatusOK
			if err != nil {
				code = http.StatusInternalServerError
				if sc, ok := err.(kithttp.StatusCoder); ok {
					code = sc.StatusCode()
				}
			}
			switch {
			case opts[LogResponse]:
				_ = LogMessage(ctx, fmt.Sprintf("response: %+v", response), "status", code, "duration", time.Since(start))
			case opts[LogErrors] && err != nil:
				_ = LogMessage(ctx, fmt.Sprintf("request: %+v", request), "error", err, "status", code, "duration", time.Since(start))
			default:
				return
			}
			return
		}
	}
}
