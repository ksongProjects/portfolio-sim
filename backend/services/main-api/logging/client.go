package logging

import "context"

type LogEmitter interface {
	Emit(ctx context.Context, level, msg string) error
}

type NoOpLogger struct{}

func NewNoOpLogger() *NoOpLogger {
	return &NoOpLogger{}
}

func (l *NoOpLogger) Emit(ctx context.Context, level, msg string) error {
	return nil
}

type Client struct {
	addr string
}

func NewClient(addr string) (*Client, error) {
	return &Client{addr: addr}, nil
}

func (c *Client) Emit(ctx context.Context, level, msg string) error {
	return nil
}
