package logging

import (
	"context"
	"time"

	"github.com/portfolio-sim/backend/internal/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client LoggingServiceClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:   conn,
		client: NewLoggingServiceClient(conn),
	}, nil
}

func (c *Client) Emit(ctx context.Context, level, msg string, metadata map[string]interface{}) error {
	entry := &models.LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     level,
		Service:   "portfolio-sim",
		Message:   msg,
		Metadata:  metadata,
	}
	_, err := c.client.Emit(ctx, entry)
	return err
}

func (c *Client) Close() error {
	return c.conn.Close()
}

type EmitResponse struct {
}

type LoggingServiceClient interface {
	Emit(ctx context.Context, in *models.LogEntry, opts ...grpc.CallOption) (*EmitResponse, error)
}

func NewLoggingServiceClient(cc grpc.ClientConnInterface) LoggingServiceClient {
	return &loggingServiceClient{cc: cc}
}

type loggingServiceClient struct {
	cc grpc.ClientConnInterface
}

func (c *loggingServiceClient) Emit(ctx context.Context, in *models.LogEntry, opts ...grpc.CallOption) (*EmitResponse, error) {
	return &EmitResponse{}, nil
}