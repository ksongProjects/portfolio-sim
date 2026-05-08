package logging

import (
	"context"
	"time"
)

type LogWriter struct {
	emitter LogEmitter
	service string
}

func NewLogWriter(emitter LogEmitter, service string) *LogWriter {
	return &LogWriter{
		emitter: emitter,
		service: service,
	}
}

func (w *LogWriter) Write(p []byte) (n int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	entry := LogEntry{
		Level:     "info",
		Message:   string(p),
		Timestamp: time.Now(),
		Service:   w.service,
	}

	_ = w.emitter.Emit(ctx, entry.Level, entry.Message)
	return len(p), nil
}