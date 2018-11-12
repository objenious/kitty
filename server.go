package kitty

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

// Server defines a kitty server.
type Server struct {
	middleware endpoint.Middleware
	shutdown   []func()

	logkeys []string
	logger  log.Logger

	transports []Transport
}

type contextKey int

const (
	// context key for logger
	logKey contextKey = iota
)

// NewServer creates a kitty server.
func NewServer(t ...Transport) *Server {
	return &Server{
		transports: t,
		logger:     &nopLogger{},
		middleware: nopMiddleware,
	}
}

// Run starts the server.
func (s *Server) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stop := make(chan error)
	exit := make(chan error)
	ctx = s.addLoggerToContext(ctx, nil)
	for _, t := range s.transports {
		m := s.addLoggerToContextMiddleware(s.middleware, t)
		if err := t.RegisterEndpoints(m); err != nil {
			return err
		}
	}
	for _, t := range s.transports {
		go func(t Transport) {
			if err := t.Start(ctx); err != nil {
				_ = s.logger.Log("msg", fmt.Sprintf("Shutting down due to server error: %s", err))
				stop <- err
			}
		}(t)
	}

	go func() {
		err := <-stop
		for _, fn := range s.shutdown {
			fn()
		}
		for _, t := range s.transports {
			if eerr := t.Shutdown(ctx); eerr != nil && err == nil {
				err = eerr
			}
		}
		exit <- err
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-ch:
		_ = s.logger.Log("msg", "received signal", "signal", sig)
	case <-ctx.Done():
		_ = s.logger.Log("msg", "canceled context")
	case err := <-exit:
		return err
	}
	stop <- nil
	return <-exit
}
