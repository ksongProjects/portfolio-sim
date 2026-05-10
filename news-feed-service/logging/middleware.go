package logging

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type ResponseWriter struct {
	http.ResponseWriter
	StatusCode int
	Size       int
}

func (rw *ResponseWriter) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.Size += n
	return n, err
}

func LoggingMiddleware(handler http.Handler, client *Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := r.Context()

		client.InfoWithMeta(ctx, "API Request", map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
			"query":  r.URL.RawQuery,
			"type":   "api_request",
		})

		wrapper := &ResponseWriter{ResponseWriter: w, StatusCode: http.StatusOK}
		handler.ServeHTTP(wrapper, r.WithContext(ctx))

		duration := time.Since(start)
		meta := map[string]interface{}{
			"method":      r.Method,
			"path":        r.URL.Path,
			"status":      wrapper.StatusCode,
			"duration_ms": duration.Milliseconds(),
			"size_bytes":  wrapper.Size,
			"type":        "api_response",
		}

		if wrapper.StatusCode >= 500 {
			client.ErrorWithMeta(ctx, "API Response Error", meta)
		} else if wrapper.StatusCode >= 400 {
			client.WarnWithMeta(ctx, "API Response Warning", meta)
		} else {
			client.InfoWithMeta(ctx, "API Response", meta)
		}
	})
}

func LoggingMiddlewareFunc(next func(http.ResponseWriter, *http.Request), client *Client) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := r.Context()

		client.InfoWithMeta(ctx, "API Request", map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
			"query":  r.URL.RawQuery,
			"type":   "api_request",
		})

		wrapper := &ResponseWriter{ResponseWriter: w, StatusCode: http.StatusOK}
		next(wrapper, r.WithContext(ctx))

		duration := time.Since(start)
		meta := map[string]interface{}{
			"method":      r.Method,
			"path":        r.URL.Path,
			"status":      wrapper.StatusCode,
			"duration_ms": duration.Milliseconds(),
			"size_bytes":  wrapper.Size,
			"type":        "api_response",
		}

		if wrapper.StatusCode >= 500 {
			client.ErrorWithMeta(ctx, "API Response Error", meta)
		} else if wrapper.StatusCode >= 400 {
			client.WarnWithMeta(ctx, "API Response Warning", meta)
		} else {
			client.InfoWithMeta(ctx, "API Response", meta)
		}
	})
}

func LogDBQuery(client *Client, sql string, duration time.Duration, err error) {
	ctx := context.Background()
	meta := map[string]interface{}{
		"sql":          sql,
		"duration_ms":  duration.Milliseconds(),
		"type":         "db_query",
	}
	if err != nil {
		meta["error"] = err.Error()
		client.ErrorWithMeta(ctx, "DB Query Error", meta)
	} else {
		client.InfoWithMeta(ctx, "DB Query", meta)
	}
}

func LogRedisCommand(client *Client, command string, duration time.Duration, err error) {
	ctx := context.Background()
	meta := map[string]interface{}{
		"command":      command,
		"duration_ms": duration.Milliseconds(),
		"type":        "redis_command",
	}
	if err != nil {
		meta["error"] = err.Error()
		client.ErrorWithMeta(ctx, "Redis Command Error", meta)
	} else {
		client.InfoWithMeta(ctx, "Redis Command", meta)
	}
}

func LogInfo(client *Client, msg string, meta map[string]interface{}) {
	if meta == nil {
		meta = map[string]interface{}{}
	}
	meta["type"] = "app_event"
	client.InfoWithMeta(context.Background(), msg, meta)
}

func LogError(client *Client, msg string, meta map[string]interface{}) {
	if meta == nil {
		meta = map[string]interface{}{}
	}
	meta["type"] = "app_event"
	client.ErrorWithMeta(context.Background(), msg, meta)
}

func LogWarn(client *Client, msg string, meta map[string]interface{}) {
	if meta == nil {
		meta = map[string]interface{}{}
	}
	meta["type"] = "app_event"
	client.WarnWithMeta(context.Background(), msg, meta)
}

func LogDebug(client *Client, msg string, meta map[string]interface{}) {
	if meta == nil {
		meta = map[string]interface{}{}
	}
	meta["type"] = "app_event"
	client.DebugWithMeta(context.Background(), msg, meta)
}

func LogNavigation(client *Client, page string) {
	client.InfoWithMeta(context.Background(), "Navigation", map[string]interface{}{
		"page": page,
		"type": "navigation",
	})
}

func LogAPICall(client *Client, method, path string, duration time.Duration, status int, err error) {
	ctx := context.Background()
	meta := map[string]interface{}{
		"method":      method,
		"path":        path,
		"duration_ms": duration.Milliseconds(),
		"status":      status,
		"type":        "api_call",
	}
	if err != nil {
		meta["error"] = err.Error()
		client.ErrorWithMeta(ctx, "API Call Error", meta)
	} else {
		client.InfoWithMeta(ctx, fmt.Sprintf("API Call: %s %s %d", method, path, status), meta)
	}
}

func LogAction(client *Client, action string, meta map[string]interface{}) {
	if meta == nil {
		meta = map[string]interface{}{}
	}
	meta["action"] = action
	meta["type"] = "user_action"
	client.InfoWithMeta(context.Background(), "User Action: "+action, meta)
}