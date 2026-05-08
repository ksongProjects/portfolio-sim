package logging

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	pool *pgxpool.Pool
}

func NewClient(pool *pgxpool.Pool) *Client {
	return &Client{pool: pool}
}

func (c *Client) Log(ctx context.Context, entry LogEntry) error {
	_, err := c.pool.Exec(ctx, `
		INSERT INTO logs (timestamp, level, service, component, message, metadata, trace_id, span_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, entry.Timestamp, entry.Level, entry.Service, entry.Component, entry.Message, entry.Metadata, entry.TraceID, entry.SpanID)
	return err
}

func (c *Client) LogAsync(entry LogEntry) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = c.Log(ctx, entry)
	}()
}
