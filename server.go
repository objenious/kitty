package kitty

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
)

// Server defines a kitty server.
type Server struct {
	cfg Config

	opts []httptransport.ServerOption

	middleware endpoint.Middleware
	endpoints  []*httpendpoint

	httpmiddleware func(http.Handler) http.Handler
	mux            Router
	svr            *http.Server

	liveness  http.HandlerFunc
	readiness http.HandlerFunc

	shutdown []func()

	logkeys []string
	logger  log.Logger

	// exit chan for graceful shutdown
	exit chan chan error
}

type contextKey int

const (
	// context key for logger
	logKey contextKey = iota
)

// NewServer creates a kitty server.
func NewServer() *Server {
	return &Server{
		cfg:            DefaultConfig,
		exit:           make(chan chan error),
		logger:         &nopLogger{},
		middleware:     nopMiddleware,
		httpmiddleware: nopHTTPMiddleWare,
		mux:            StdlibRouter(),
		liveness:       defaultHealthcheck,
		readiness:      defaultHealthcheck,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.httpmiddleware(s.mux).ServeHTTP(w, r)
}

// Run starts the server.
func (s *Server) Run(ctx context.Context) error {
	if err := s.register(); err != nil {
		return err
	}
	if err := s.start(ctx); err != nil {
		return err
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-ch:
		s.logger.Log("msg", "received signal", "signal", sig)
	case <-ctx.Done():
		s.logger.Log("msg", "canceled context")
	}
	return s.stop()
}

func (s *Server) register() error {
	opts := []httptransport.ServerOption{
		httptransport.ServerBefore(httptransport.PopulateRequestContext),
		httptransport.ServerBefore(s.addLoggerToContext),
	}
	opts = append(opts, s.opts...)

	// register endpoints
	for _, ep := range s.endpoints {
		if len(ep.methods) == 0 {
			return errors.New("missing methods in handler")
		}
		if len(ep.paths) == 0 {
			return errors.New("missing paths in handler")
		}
		s.mux.Handle(ep.methods, ep.paths,
			httptransport.NewServer(
				s.middleware(ep.endpoint),
				ep.decoder,
				ep.encoder,
				append(opts, ep.options...)...))
	}

	// register health handlers
	s.mux.Handle([]string{http.MethodGet}, []string{s.cfg.LivenessCheckPath}, http.HandlerFunc(s.liveness))
	s.mux.Handle([]string{http.MethodGet}, []string{s.cfg.ReadinessCheckPath}, http.HandlerFunc(s.readiness))

	// register pprof handlers
	registerPProf(s.cfg, s.mux)
	return nil
}

func (s *Server) start(ctx context.Context) error {
	s.svr = &http.Server{
		Handler: s,
		Addr:    fmt.Sprintf(":%d", s.cfg.HTTPPort),
	}
	go func() {
		err := s.svr.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			s.logger.Log("msg", fmt.Sprintf("Shutting down due to server error: %s", err))
			s.stop()
		}
	}()

	s.logger.Log("msg", fmt.Sprintf("Listening on port: %d", s.cfg.HTTPPort))

	go func() {
		exit := <-s.exit
		for _, fn := range s.shutdown {
			fn()
		}
		exit <- s.svr.Shutdown(ctx)
	}()

	return nil
}

func (s *Server) stop() error {
	ch := make(chan error)
	s.exit <- ch
	return <-ch
}

func registerPProf(cfg Config, mux Router) {
	if !cfg.EnablePProf {
		return
	}
	mux.Handle([]string{http.MethodGet}, []string{"/debug/pprof/"}, http.HandlerFunc(pprof.Index))
	mux.Handle([]string{http.MethodGet}, []string{"/debug/pprof/cmdline"}, http.HandlerFunc(pprof.Cmdline))
	mux.Handle([]string{http.MethodGet}, []string{"/debug/pprof/profile"}, http.HandlerFunc(pprof.Profile))
	mux.Handle([]string{http.MethodGet}, []string{"/debug/pprof/symbol"}, http.HandlerFunc(pprof.Symbol))
	mux.Handle([]string{http.MethodGet}, []string{"/debug/pprof/trace"}, http.HandlerFunc(pprof.Trace))
	mux.Handle([]string{http.MethodGet}, []string{"/debug/pprof/goroutine"}, pprof.Handler("goroutine"))
	mux.Handle([]string{http.MethodGet}, []string{"/debug/pprof/heap"}, pprof.Handler("heap"))
	mux.Handle([]string{http.MethodGet}, []string{"/debug/pprof/threadcreate"}, pprof.Handler("threadcreate"))
	mux.Handle([]string{http.MethodGet}, []string{"/debug/pprof/block"}, pprof.Handler("block"))
}
