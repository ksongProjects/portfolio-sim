package logging

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type ResponseWriter struct {
	http.ResponseWriter
	StatusCode int
	Size       int
}

func (rw *ResponseWriter) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.Size += n
	return n, err
}

type Middleware struct {
	client  *Client
	logger  *slog.Logger
	service string
}

func NewMiddleware(service string, logURL string, logger *slog.Logger) *Middleware {
	return &Middleware{
		client:  NewClient(service, logURL),
		logger:  logger,
		service: service,
	}
}

func (m *Middleware) WrapHandlerFunc(next func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := r.Context()

		rw := &ResponseWriter{ResponseWriter: w, StatusCode: http.StatusOK}

		m.client.InfoWithMeta(ctx, "Request started", map[string]interface{}{
			"method":     r.Method,
			"path":       r.URL.Path,
			"query":      r.URL.RawQuery,
			"remote_addr": r.RemoteAddr,
		})

		next(w, r.WithContext(ctx))

		duration := time.Since(start)

		meta := map[string]interface{}{
			"method":      r.Method,
			"path":        r.URL.Path,
			"status":     rw.StatusCode,
			"duration_ms": duration.Milliseconds(),
			"size_bytes":  rw.Size,
		}

		if rw.StatusCode >= 500 {
			m.client.ErrorWithMeta(ctx, "Request error", meta)
		} else if rw.StatusCode >= 400 {
			m.client.WarnWithMeta(ctx, "Request warning", meta)
		} else {
			m.client.InfoWithMeta(ctx, "Request completed", meta)
		}
	})
}

func (m *Middleware) Log(ctx context.Context, level, msg string, meta map[string]interface{}) {
	if meta == nil {
		meta = map[string]interface{}{}
	}
	switch level {
	case "ERROR":
		m.client.ErrorWithMeta(ctx, msg, meta)
	case "WARN":
		m.client.WarnWithMeta(ctx, msg, meta)
	case "DEBUG":
		m.client.DebugWithMeta(ctx, msg, meta)
	default:
		m.client.InfoWithMeta(ctx, msg, meta)
	}
}

func (m *Middleware) Info(ctx context.Context, msg string) {
	m.client.Info(ctx, msg)
}

func (m *Middleware) Error(ctx context.Context, msg string, meta map[string]interface{}) {
	m.client.ErrorWithMeta(ctx, msg, meta)
}

func (m *Middleware) Warn(ctx context.Context, msg string, meta map[string]interface{}) {
	m.client.WarnWithMeta(ctx, msg, meta)
}