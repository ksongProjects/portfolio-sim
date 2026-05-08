package models

import (
	"time"

	"github.com/google/uuid"
)

type Log struct {
	ID        uuid.UUID       `json:"id"`
	Timestamp time.Time      `json:"timestamp"`
	Level     string          `json:"level"`
	Service   string          `json:"service"`
	Component string          `json:"component,omitempty"`
	Message   string          `json:"message"`
	Metadata  map[string]any  `json:"metadata,omitempty"`
	TraceID   string          `json:"trace_id,omitempty"`
	SpanID    string          `json:"span_id,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

type EmitLogRequest struct {
	Level     string         `json:"level"`
	Service   string         `json:"service"`
	Component string         `json:"component,omitempty"`
	Message   string         `json:"message"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	TraceID   string         `json:"trace_id,omitempty"`
	SpanID    string         `json:"span_id,omitempty"`
}

type EmitLogResponse struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
}