package kitty

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

// nopLogger is the default logger and does nothing.
type nopLogger struct{}

func (l *nopLogger) Log(keyvals ...interface{}) error { return nil }

// Logger sets the logger.
func (s *Server) Logger(l log.Logger) *Server {
	s.logger = l
	return s
}

// LogContext defines the list of keys to add to all log lines.
// Keys may vary depending on transport.
// Available keys for the http transport are : http-method, http-uri, http-path, http-proto, http-requesthost,
// http-remote-addr, http-x-forwarded-for, http-x-forwarded-proto, http-user-agent and http-x-request-id.
func (s *Server) LogContext(keys ...string) *Server {
	s.logkeys = keys
	return s
}

func (s *Server) addLoggerToContext(ctx context.Context, keys map[string]interface{}) context.Context {
	l := s.logger
	if keys != nil {
		for _, k := range s.logkeys {
			if val, ok := ctx.Value(keys[k]).(string); ok && val != "" {
				l = log.With(l, k, val)
			}
		}
	}
	return context.WithValue(ctx, logKey, l)
}

func (s *Server) addLoggerToContextMiddleware(m endpoint.Middleware, t Transport) endpoint.Middleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		e = m(e)
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			return e(s.addLoggerToContext(ctx, t.LogKeys()), request)
		}
	}
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
