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

	// exit chan for graceful shutdown
	exit chan chan error
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
		exit:       make(chan chan error),
		logger:     &nopLogger{},
		middleware: nopMiddleware,
	}
}

// Run starts the server.
func (s *Server) Run(ctx context.Context) error {
	ctx = s.addLoggerToContext(ctx)
	for _, t := range s.transports {
		if err := t.RegisterEndpoints(s.middleware, s.addLoggerToContext); err != nil {
			return err
		}
	}
	for _, t := range s.transports {
		go func(t Transport) {
			if err := t.Start(ctx); err != nil {
				_ = s.logger.Log("msg", fmt.Sprintf("Shutting down due to server error: %s", err))
				_ = s.stop()
			}
		}(t)
	}

	go func() {
		exit := <-s.exit
		for _, fn := range s.shutdown {
			fn()
		}
		var err error
		for _, t := range s.transports {
			if eerr := t.Shutdown(ctx); eerr != nil {
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
	}
	return s.stop()
}

func (s *Server) stop() error {
	ch := make(chan error)
	s.exit <- ch
	return <-ch
}
