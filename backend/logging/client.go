package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type LogEntry struct {
	ID        string                 `json:"id"`
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Service   string                 `json:"service"`
	Component string                 `json:"component,omitempty"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
	SpanID    string                 `json:"span_id,omitempty"`
}

type Client struct {
	serviceName string
	logURL     string
	client     *http.Client
}

func NewClient(serviceName string, logURL string) *Client {
	if logURL == "" {
		logURL = "http://logging-service:9090/api/logs"
	}
	return &Client{
		serviceName: serviceName,
		logURL:      logURL,
		client:      &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) Emit(ctx context.Context, level string, msg string) error {
	entry := LogEntry{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level,
		Service:   c.serviceName,
		Message:   msg,
	}
	return c.send(ctx, entry)
}

func (c *Client) EmitWithMetadata(ctx context.Context, level string, msg string, metadata map[string]interface{}) error {
	entry := LogEntry{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level,
		Service:   c.serviceName,
		Message:   msg,
		Metadata:  metadata,
	}
	return c.send(ctx, entry)
}

func (c *Client) send(ctx context.Context, entry LogEntry) error {
	body, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.logURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("logging server returned %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) Info(ctx context.Context, msg string) error {
	return c.Emit(ctx, "INFO", msg)
}

func (c *Client) Error(ctx context.Context, msg string) error {
	return c.Emit(ctx, "ERROR", msg)
}

func (c *Client) Warn(ctx context.Context, msg string) error {
	return c.Emit(ctx, "WARN", msg)
}

func (c *Client) Debug(ctx context.Context, msg string) error {
	return c.Emit(ctx, "DEBUG", msg)
}

func (c *Client) InfoWithMeta(ctx context.Context, msg string, meta map[string]interface{}) error {
	return c.EmitWithMetadata(ctx, "INFO", msg, meta)
}

func (c *Client) ErrorWithMeta(ctx context.Context, msg string, meta map[string]interface{}) error {
	return c.EmitWithMetadata(ctx, "ERROR", msg, meta)
}

func (c *Client) WarnWithMeta(ctx context.Context, msg string, meta map[string]interface{}) error {
	return c.EmitWithMetadata(ctx, "WARN", msg, meta)
}

func (c *Client) DebugWithMeta(ctx context.Context, msg string, meta map[string]interface{}) error {
	return c.EmitWithMetadata(ctx, "DEBUG", msg, meta)
}