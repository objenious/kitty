package kitty

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// Transport is the interface all transports must implement.
type Transport interface {
	// RegisterEndpoints registers all endpoints.
	RegisterEndpoints(m endpoint.Middleware, fn addLoggerToContextFn) error
	// Start starts the transport.
	Start(ctx context.Context) error
	// Shutdown shutdowns the transport.
	Shutdown(ctx context.Context) error
}
