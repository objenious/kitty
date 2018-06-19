package kitty

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
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

// nopHTTPMiddleWare is the default HTTP middleware, and does nothing.
func nopHTTPMiddleWare(h http.Handler) http.Handler {
	return h
}

// HTTPMiddlewares defines the list of HTTP middlewares to be added to all HTTP handlers.
func (s *Server) HTTPMiddlewares(m ...func(http.Handler) http.Handler) *Server {
	s.httpmiddleware = func(next http.Handler) http.Handler {
		for i := len(m) - 1; i >= 0; i-- {
			next = m[i](next)
		}
		return next
	}
	return s
}

// LogEndpoint creates a middleware that logs Endpoint calls.
// If "request" is specified, the enpoint request will be logged before the endpoint.
// If "response" is specified, the endpoint response will be logged after.
// In all cases, the duration and HTTP status code will be logged.
func LogEndpoint(fields ...string) endpoint.Middleware {
	var logrequest, logresponse bool
	for _, f := range fields {
		switch f {
		case "request":
			logrequest = true
		case "response":
			logresponse = true
		}
	}
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			l := Logger(ctx)
			if logrequest {
				l.Log("msg", fmt.Sprintf("request: %+v", request))
			}
			start := time.Now()
			response, err = e(ctx, request)
			code := http.StatusInternalServerError
			if sc, ok := err.(httptransport.StatusCoder); ok {
				code = sc.StatusCode()
			}
			var msg string
			if logresponse {
				msg = fmt.Sprintf("response: %+v", response)
			} else {
				msg = "response"
			}
			l.Log("msg", msg, "status", code, "duration", time.Since(start))
			return
		}
	}
}
