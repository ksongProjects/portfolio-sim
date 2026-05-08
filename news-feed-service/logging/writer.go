package logging

import (
	"sync"
	"time"
)

type LogWriter struct {
	client   *Client
}

func NewLogWriter(client *Client) *LogWriter {
	return &LogWriter{client: client}
}

func (w *LogWriter) Write(p []byte) (n int, err error) {
	entry := LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     "INFO",
		Service:   "news-feed-service",
		Message:   string(p),
	}
	w.client.LogAsync(entry)
	return len(p), nil
}

type safeWriter struct {
	mu    sync.Mutex
	w     *LogWriter
}

func (sw *safeWriter) Write(p []byte) (n int, err error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.w.Write(p)
}
