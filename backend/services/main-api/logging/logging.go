package logging

import (
	"context"
)

type LogEmitter interface {
	Emit(ctx context.Context, level string, msg string) error
}

type NoOpLogger struct{}

func NewNoOpLogger() *NoOpLogger {
	return &NoOpLogger{}
}

func (l *NoOpLogger) Emit(ctx context.Context, level string, msg string) error {
	return nil
}
