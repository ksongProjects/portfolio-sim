package logging

import (
	"context"
	"encoding/json"
	"time"
)

type LogWriter struct {
	emitter LogEmitter
	service string
	level   string
}

func NewLogWriter(emitter LogEmitter, service, level string) *LogWriter {
	return &LogWriter{
		emitter: emitter,
		service: service,
		level:   level,
	}
}

func (w *LogWriter) Write(p []byte) (n int, err error) {
	msg := string(p)
	if msg == "" {
		return 0, nil
	}

	metadata := map[string]interface{}{}
	if len(p) > 0 {
		if json.Valid(p) {
			var m map[string]interface{}
			if err := json.Unmarshal(p, &m); err == nil {
				metadata = m
			}
		}
	}

	ctx := context.Background()
	err = w.emitter.Emit(ctx, w.level, w.service, "", msg, metadata, "", "")
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

type writerCtxKey struct{}

func WithWriter(ctx context.Context, w *LogWriter) context.Context {
	return context.WithValue(ctx, writerCtxKey{}, w)
}

func WriterFromContext(ctx context.Context) *LogWriter {
	if w, ok := ctx.Value(writerCtxKey{}).(*LogWriter); ok {
		return w
	}
	return nil
}

func StripTimestamp(s string) string {
	if len(s) > 30 && s[0] == '[' {
		if idx := time.Now().Format("2006-01-02T15:04:05"); len(s) > len(idx)+2 {
			start := 0
			for i, c := range s {
				if c == ']' {
					start = i + 2
					break
				}
			}
			if start > 0 && start < len(s) {
				return s[start:]
			}
		}
	}
	return s
}