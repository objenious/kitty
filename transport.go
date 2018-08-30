package kitty

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// Transport is the interface all transports must implement.
type Transport interface {
	// RegisterEndpoints registers all endpoints, and injects the AddLoggerToContextFn call to configure the logger.
	RegisterEndpoints(m endpoint.Middleware, fn AddLoggerToContextFn) error
	// Start starts the transport.
	Start(ctx context.Context) error
	// Shutdown shutdowns the transport.
	Shutdown(ctx context.Context) error
}
