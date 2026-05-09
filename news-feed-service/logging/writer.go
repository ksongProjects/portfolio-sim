package logging

import (
	"context"
)

type LogWriter struct {
	client *Client
}

func NewLogWriter(client *Client) *LogWriter {
	return &LogWriter{client: client}
}

func (w *LogWriter) Write(p []byte) (n int, err error) {
	ctx := context.Background()
	_ = w.client.Emit(ctx, "INFO", string(p))
	return len(p), nil
}

type safeWriter struct {
	mu *LogWriter
}

func (sw *safeWriter) Write(p []byte) (n int, err error) {
	return sw.mu.Write(p)
}