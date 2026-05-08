package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type LogEntry struct {
	ID        uuid.UUID       `json:"id"`
	Timestamp time.Time       `json:"timestamp"`
	Level     string          `json:"level"`
	Service   string          `json:"service"`
	Component string          `json:"component,omitempty"`
	Message   string          `json:"message"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
	TraceID   string          `json:"trace_id,omitempty"`
	SpanID    string          `json:"span_id,omitempty"`
}
