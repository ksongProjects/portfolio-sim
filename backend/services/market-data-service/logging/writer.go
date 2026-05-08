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
	w.client.Emit(context.Background(), "INFO", string(p))
	return len(p), nil
}
