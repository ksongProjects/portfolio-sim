package logging

import "context"

type Writer struct {
	client LogEmitter
}

func NewWriter(client LogEmitter) *Writer {
	return &Writer{client: client}
}

func (w *Writer) Write(p []byte) (n int, err error) {
	w.client.Emit(context.Background(), "INFO", string(p))
	return len(p), nil
}
