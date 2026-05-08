package logging

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type LogEntry struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	TraceID   string    `json:"trace_id"`
}

type LogEmitter interface {
	Emit(ctx context.Context, level string, msg string) error
}

type Client struct {
	conn   *grpc.ClientConn
	client LoggingServiceClient
}

type LoggingServiceClient interface {
	EmitLog(ctx context.Context, in *LogRequest, opts ...grpc.CallOption) (*LogResponse, error)
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:   conn,
		client: NewLoggingServiceClient(conn),
	}, nil
}

func (c *Client) Emit(ctx context.Context, level string, msg string) error {
	_, err := c.client.EmitLog(ctx, &LogRequest{
		Level:   level,
		Message: msg,
	})
	return err
}

func (c *Client) Close() error {
	return c.conn.Close()
}

type NoOpLogger struct{}

func NewNoOpLogger() *NoOpLogger {
	return &NoOpLogger{}
}

func (l *NoOpLogger) Emit(ctx context.Context, level string, msg string) error {
	return nil
}