package kitty

import (
	"context"
	"net/http"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
)

// nopLogger is the default logger and does nothing.
type nopLogger struct{}

func (l *nopLogger) Log(keyvals ...interface{}) error { return nil }

var logkeys = map[string]interface{}{
	"http-method":            httptransport.ContextKeyRequestMethod,
	"http-uri":               httptransport.ContextKeyRequestURI,
	"http-path":              httptransport.ContextKeyRequestPath,
	"http-proto":             httptransport.ContextKeyRequestProto,
	"http-requesthost":       httptransport.ContextKeyRequestHost,
	"http-remote-addr":       httptransport.ContextKeyRequestRemoteAddr,
	"http-x-forwarded-for":   httptransport.ContextKeyRequestXForwardedFor,
	"http-x-forwarded-proto": httptransport.ContextKeyRequestXForwardedProto,
	"http-user-agent":        httptransport.ContextKeyRequestUserAgent,
	"http-x-request-id":      httptransport.ContextKeyRequestXRequestID,
}

// Logger sets the logger.
func (s *Server) Logger(l log.Logger) *Server {
	s.logger = l
	return s
}

// LogContext defines the list of keys to add to all log lines.
// Available keys are : http-method, http-uri, http-path, http-proto, http-requesthost, http-remote-addr,
// http-x-forwarded-for, http-x-forwarded-proto, http-user-agent and http-x-request-id.
func (s *Server) LogContext(keys ...string) *Server {
	s.logkeys = keys
	return s
}

func (s *Server) addLoggerToContext(ctx context.Context, _ *http.Request) context.Context {
	l := s.logger
	for _, k := range s.logkeys {
		if val, ok := ctx.Value(logkeys[k]).(string); ok && val != "" {
			l = log.With(l, k, val)
		}
	}
	return context.WithValue(ctx, logKey, l)
}

// Logger will return the logger that has been injected into the context by the kitty
// server. This function can only be called from an endpoint.
func Logger(ctx context.Context) log.Logger {
	return ctx.Value(logKey).(log.Logger)
}

// LogMessage will log a message.
// This function can only be called from an endpoint.
func LogMessage(ctx context.Context, msg string, keyvals ...interface{}) error {
	l := Logger(ctx)
	keyvals = append(keyvals, "msg", msg)
	return l.Log(keyvals...)
}
