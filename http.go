package kitty

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
)

// HTTPTransport defines a HTTP transport for a kitty Server.
type HTTPTransport struct {
	cfg Config

	opts []kithttp.ServerOption

	endpoints []*httpendpoint

	httpmiddleware func(http.Handler) http.Handler
	mux            Router
	svr            *http.Server

	liveness  http.HandlerFunc
	readiness http.HandlerFunc
}

var _ Transport = &HTTPTransport{}

// nopHTTPMiddleWare is the default HTTP middleware, and does nothing.
func nopHTTPMiddleWare(h http.Handler) http.Handler {
	return h
}

// HTTPMiddlewares defines the list of HTTP middlewares to be added to all HTTP handlers.
func (t *HTTPTransport) HTTPMiddlewares(m ...func(http.Handler) http.Handler) *HTTPTransport {
	t.httpmiddleware = func(next http.Handler) http.Handler {
		for i := len(m) - 1; i >= 0; i-- {
			next = m[i](next)
		}
		return next
	}
	return t
}

// NewHTTPTransport creates a new HTTP transport, based on the specified config.
func NewHTTPTransport(cfg Config) *HTTPTransport {
	t := &HTTPTransport{
		cfg:            DefaultConfig,
		httpmiddleware: nopHTTPMiddleWare,
		mux:            StdlibRouter(),
		liveness:       defaultHealthcheck,
		readiness:      defaultHealthcheck,
	}
	if cfg.HTTPPort > 0 {
		t.cfg.HTTPPort = cfg.HTTPPort
	}
	if cfg.LivenessCheckPath != "" {
		t.cfg.LivenessCheckPath = cfg.LivenessCheckPath
	}
	if cfg.ReadinessCheckPath != "" {
		t.cfg.ReadinessCheckPath = cfg.ReadinessCheckPath
	}
	t.cfg.EnablePProf = cfg.EnablePProf
	return t
}

// ServeHTTP implements http.Handler.
func (t *HTTPTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.httpmiddleware(t.mux).ServeHTTP(w, r)
}

// RegisterEndpoints registers all configured endpoints, wraps them with the m middleware.
func (t *HTTPTransport) RegisterEndpoints(m endpoint.Middleware) error {
	opts := []kithttp.ServerOption{
		kithttp.ServerBefore(kithttp.PopulateRequestContext),
	}
	opts = append(opts, t.opts...)

	// register endpoints
	for _, ep := range t.endpoints {
		t.mux.Handle(ep.method, ep.path,
			kithttp.NewServer(
				m(ep.endpoint),
				ep.decoder,
				ep.encoder,
				append(opts, ep.options...)...))
	}

	// register health handlers
	t.mux.Handle("GET", t.cfg.LivenessCheckPath, t.liveness)
	t.mux.Handle("GET", t.cfg.ReadinessCheckPath, t.readiness)

	// register pprof handlers
	registerPProf(t.cfg, t.mux)
	return nil
}

var httpLogkeys = map[string]interface{}{
	"http-method":            kithttp.ContextKeyRequestMethod,
	"http-uri":               kithttp.ContextKeyRequestURI,
	"http-path":              kithttp.ContextKeyRequestPath,
	"http-proto":             kithttp.ContextKeyRequestProto,
	"http-requesthost":       kithttp.ContextKeyRequestHost,
	"http-remote-addr":       kithttp.ContextKeyRequestRemoteAddr,
	"http-x-forwarded-for":   kithttp.ContextKeyRequestXForwardedFor,
	"http-x-forwarded-proto": kithttp.ContextKeyRequestXForwardedProto,
	"http-user-agent":        kithttp.ContextKeyRequestUserAgent,
	"http-x-request-id":      kithttp.ContextKeyRequestXRequestID,
}

// LogKeys returns the list of name key to context key mappings
func (t *HTTPTransport) LogKeys() map[string]interface{} {
	return httpLogkeys
}

// Start starts the HTTP server.
func (t *HTTPTransport) Start(ctx context.Context) error {
	t.svr = &http.Server{
		Handler: t,
		Addr:    fmt.Sprintf(":%d", t.cfg.HTTPPort),
	}
	_ = LogMessage(ctx, fmt.Sprintf("Listening on port: %d", t.cfg.HTTPPort))
	err := t.svr.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown shutdowns the HTTP server
func (t *HTTPTransport) Shutdown(ctx context.Context) error {
	return t.svr.Shutdown(ctx)
}

func registerPProf(cfg Config, mux Router) {
	if !cfg.EnablePProf {
		return
	}
	mux.Handle("GET", "/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("GET", "/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("GET", "/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("GET", "/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("GET", "/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	mux.Handle("GET", "/debug/pprof/goroutine", pprof.Handler("goroutine"))
	mux.Handle("GET", "/debug/pprof/heap", pprof.Handler("heap"))
	mux.Handle("GET", "/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	mux.Handle("GET", "/debug/pprof/block", pprof.Handler("block"))
}
