package logging

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type LogEmitter interface {
	Emit(ctx context.Context, level, service, component, message string, metadata map[string]interface{}, traceID, spanID string) error
}

type Client struct {
	emitter LogEmitter
}

func NewClient(e LogEmitter) *Client {
	return &Client{emitter: e}
}

func (c *Client) Emit(ctx context.Context, level, service, component, message string, metadata map[string]interface{}, traceID, spanID string) error {
	return c.emitter.Emit(ctx, level, service, component, message, metadata, traceID, spanID)
}

func (c *Client) Debug(ctx context.Context, service, message string) error {
	return c.Emit(ctx, "DEBUG", service, "", message, nil, "", "")
}

func (c *Client) Info(ctx context.Context, service, message string) error {
	return c.Emit(ctx, "INFO", service, "", message, nil, "", "")
}

func (c *Client) Warn(ctx context.Context, service, message string) error {
	return c.Emit(ctx, "WARN", service, "", message, nil, "", "")
}

func (c *Client) Error(ctx context.Context, service, message string) error {
	return c.Emit(ctx, "ERROR", service, "", message, nil, "", "")
}

func (c *Client) Fatal(ctx context.Context, service, message string) error {
	return c.Emit(ctx, "FATAL", service, "", message, nil, "", "")
}

type NoOpLogger struct{}

func NewNoOpLogger() *NoOpLogger {
	return &NoOpLogger{}
}

func (l *NoOpLogger) Emit(ctx context.Context, level, service, component, message string, metadata map[string]interface{}, traceID, spanID string) error {
	return nil
}

type Logger struct {
	client *Client
}

func NewLogger(e LogEmitter) *Logger {
	return &Logger{client: NewClient(e)}
}

func (l *Logger) Debug(ctx context.Context, service, message string) error {
	return l.client.Debug(ctx, service, message)
}

func (l *Logger) Info(ctx context.Context, service, message string) error {
	return l.client.Info(ctx, service, message)
}

func (l *Logger) Warn(ctx context.Context, service, message string) error {
	return l.client.Warn(ctx, service, message)
}

func (l *Logger) Error(ctx context.Context, service, message string) error {
	return l.client.Error(ctx, service, message)
}

func (l *Logger) Fatal(ctx context.Context, service, message string) error {
	return l.client.Fatal(ctx, service, message)
}

func (l *Logger) WithTrace(traceID, spanID string) *TraceLogger {
	return &TraceLogger{logger: l, traceID: traceID, spanID: spanID}
}

type TraceLogger struct {
	logger  *Logger
	traceID string
	spanID  string
}

func (t *TraceLogger) Debug(ctx context.Context, service, message string) error {
	return t.logger.client.Emit(ctx, "DEBUG", service, "", message, nil, t.traceID, t.spanID)
}

func (t *TraceLogger) Info(ctx context.Context, service, message string) error {
	return t.logger.client.Emit(ctx, "INFO", service, "", message, nil, t.traceID, t.spanID)
}

func (t *TraceLogger) Warn(ctx context.Context, service, message string) error {
	return t.logger.client.Emit(ctx, "WARN", service, "", message, nil, t.traceID, t.spanID)
}

func (t *TraceLogger) Error(ctx context.Context, service, message string) error {
	return t.logger.client.Emit(ctx, "ERROR", service, "", message, nil, t.traceID, t.spanID)
}

func (t *TraceLogger) Fatal(ctx context.Context, service, message string) error {
	return t.logger.client.Emit(ctx, "FATAL", service, "", message, nil, t.traceID, t.spanID)
}

type logEntry struct {
	ID        uuid.UUID
	Timestamp time.Time
	Level     string
	Service   string
	Component string
	Message   string
	Metadata  map[string]interface{}
	TraceID   string
	SpanID    string
}

func fmtEntry(entry *logEntry) string {
	return fmt.Sprintf("[%s] %s %s.%s: %s | trace=%s span=%s",
		entry.Timestamp.Format(time.RFC3339),
		entry.Level,
		entry.Service,
		entry.Component,
		entry.Message,
		entry.TraceID,
		entry.SpanID,
	)
}