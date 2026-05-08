package logging

import (
	"time"

	"github.com/google/uuid"
)

type LogEntry struct {
	ID        uuid.UUID
	Timestamp time.Time
	Level     string
	Service   string
	Component string
	Message   string
	Metadata  map[string]interface{}
	TraceID   string
	SpanID    string
	CreatedAt time.Time
}