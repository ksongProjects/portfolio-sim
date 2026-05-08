package logging

type LogEntry struct {
	Level     string
	Service   string
	Component string
	Message   string
	TraceID   string
	SpanID    string
}
