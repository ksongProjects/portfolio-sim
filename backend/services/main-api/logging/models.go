package logging

import (
	"time"
)

type LogEntry struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	TraceID   string    `json:"trace_id"`
}

type LogRequest struct {
	Level   string `protobuf:"bytes,1,opt,name=level,proto3"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3"`
}

type LogResponse struct {
}