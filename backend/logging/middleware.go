package logging

import (
	"context"
	"net/http"
)

func WrapHandlerFunc(next func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next(w, r)
	})
}

func Log(ctx context.Context, client *Client, level, msg string, meta map[string]interface{}) {
	if meta == nil {
		meta = map[string]interface{}{}
	}
	switch level {
	case "ERROR":
		client.ErrorWithMeta(ctx, msg, meta)
	case "WARN":
		client.WarnWithMeta(ctx, msg, meta)
	case "DEBUG":
		client.DebugWithMeta(ctx, msg, meta)
	default:
		client.InfoWithMeta(ctx, msg, meta)
	}
}

func Info(ctx context.Context, client *Client, msg string) {
	client.Info(ctx, msg)
}

func Error(ctx context.Context, client *Client, msg string, meta map[string]interface{}) {
	client.ErrorWithMeta(ctx, msg, meta)
}

func Warn(ctx context.Context, client *Client, msg string, meta map[string]interface{}) {
	client.WarnWithMeta(ctx, msg, meta)
}