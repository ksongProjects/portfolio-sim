package logging

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type LogEmitter interface {
	Emit(ctx context.Context, level string, msg string) error
}

type Client struct {
	serviceName string
}

func NewClient(serviceName string) *Client {
	return &Client{serviceName: serviceName}
}

func (c *Client) Emit(ctx context.Context, level string, msg string) error {
	entry := LogEntry{
		ID:        uuid.New(),
		Timestamp: time.Now().UTC(),
		Level:     level,
		Service:   c.serviceName,
		Message:   msg,
	}
	return c.send(ctx, entry)
}

func (c *Client) send(ctx context.Context, entry LogEntry) error {
	return nil
}
