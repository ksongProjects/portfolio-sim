package logging

import (
	"context"
	"sync"
	"time"
)

type Writer struct {
	client  *Client
	mu      sync.Mutex
	service string
}

func NewWriter(client *Client, service string) *Writer {
	return &Writer{
		client:  client,
		service: service,
	}
}

func (w *Writer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msg := string(p)
	metadata := map[string]interface{}{"bytes_written": len(p)}

	if err := w.client.Emit(ctx, "INFO", msg, metadata); err != nil {
		return 0, err
	}
	return len(p), nil
}