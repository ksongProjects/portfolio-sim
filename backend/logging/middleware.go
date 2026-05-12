package logging

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func Middleware(client *Client) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !shouldLog(r.URL.Path) || client == nil {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			captureBody := shouldCaptureBody(r.URL.Path)
			var reqBody []byte
			if captureBody {
				reqBody, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewReader(reqBody))
			}

			reqMeta := map[string]interface{}{
				"type":            "api_request",
				"method":          r.Method,
				"path":            r.URL.Path,
				"query":           sanitizeURLQuery(r.URL.RawQuery),
				"request_headers": redactHeaders(r.Header),
				"remote_addr":     r.RemoteAddr,
				"route":           getActionFromPath(r.URL.Path),
			}
			if len(reqBody) > 0 {
				reqMeta["request_body"] = sanitizeBody(r.Header.Get("Content-Type"), reqBody)
				reqMeta["request_body_size"] = len(reqBody)
			} else {
				reqMeta["request_body_skipped"] = true
			}
			client.log(r.Context(), "INFO", "API Request", reqMeta)

			wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK, captureBody: captureBody}
			next.ServeHTTP(w, r)

			respMeta := map[string]interface{}{
				"type":               "api_response",
				"method":             r.Method,
				"path":               r.URL.Path,
				"query":              sanitizeURLQuery(r.URL.RawQuery),
				"status":             wrapper.statusCode,
				"duration_ms":        time.Since(start).Milliseconds(),
				"response_headers":   redactHeaders(wrapper.Header()),
				"response_body_size": wrapper.size,
				"route":              getActionFromPath(r.URL.Path),
			}
			if captureBody && len(wrapper.body) > 0 {
				respMeta["response_body"] = sanitizeBody(wrapper.Header().Get("Content-Type"), wrapper.body)
			} else {
				respMeta["response_body_skipped"] = true
			}

			level := "INFO"
			if wrapper.statusCode >= 500 {
				level = "ERROR"
			} else if wrapper.statusCode >= 400 {
				level = "WARN"
			}
			client.log(r.Context(), level, "API Response", respMeta)
		})
	}
}

func randomID() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = "abcdefghijklmnopqrstuvwxyz0123456789"[time.Now().UnixNano()%36]
	}
	return string(b)
}

var skipPaths = map[string]bool{
	"/api/logs":                   true,
	"/api/observability/logs":     true,
	"/api/observability/services": true,
}

func shouldLog(path string) bool {
	return !skipPaths[path]
}

func shouldCaptureBody(path string) bool {
	return path != "/api/logs"
}

type responseWriter struct {
	http.ResponseWriter
	statusCode  int
	body        []byte
	size        int
	captureBody bool
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.size += len(b)
	if rw.captureBody {
		rw.body = append(rw.body, b...)
	}
	return rw.ResponseWriter.Write(b)
}

func getActionFromPath(path string) string {
	if strings.HasPrefix(path, "/api/observability/") {
		return "observability"
	}
	if strings.HasPrefix(path, "/api/portfolio/") {
		return "portfolio"
	}
	if strings.HasPrefix(path, "/api/market/") {
		return "market"
	}
	if strings.HasPrefix(path, "/api/news") {
		return "news"
	}
	if strings.HasPrefix(path, "/api/strategies") {
		return "strategies"
	}
	if strings.HasPrefix(path, "/api/signals") {
		return "signals"
	}
	if strings.HasPrefix(path, "/api/notifications") {
		return "notifications"
	}
	if strings.HasPrefix(path, "/api/providers") {
		return "providers"
	}
	if strings.HasPrefix(path, "/api/connections") {
		return "connections"
	}
	if strings.HasPrefix(path, "/api/rss-feeds") {
		return "rss-feeds"
	}
	if strings.HasPrefix(path, "/api/tickers/") {
		return "tickers"
	}
	return "api"
}

func sanitizeURLQuery(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}

	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return rawQuery
	}

	for key := range values {
		lowerKey := strings.ToLower(key)
		if lowerKey == "access_token" || lowerKey == "refresh_token" || lowerKey == "apikey" || lowerKey == "api_key" || lowerKey == "key" || lowerKey == "token" {
			values.Set(key, "REDACTED")
		}
	}
	return values.Encode()
}

func redactHeaders(headers http.Header) map[string][]string {
	if headers == nil {
		return nil
	}

	redacted := make(map[string][]string, len(headers))
	for key, values := range headers {
		lowerKey := strings.ToLower(key)
		if lowerKey == "authorization" || lowerKey == "cookie" || lowerKey == "set-cookie" || lowerKey == "x-api-key" || lowerKey == "proxy-authorization" {
			redacted[key] = []string{"REDACTED"}
			continue
		}

		copied := make([]string, len(values))
		copy(copied, values)
		redacted[key] = copied
	}
	return redacted
}

func sanitizeBody(contentType string, body []byte) interface{} {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return ""
	}

	if strings.Contains(strings.ToLower(contentType), "json") || trimmed[0] == '{' || trimmed[0] == '[' {
		var payload interface{}
		if err := json.Unmarshal(trimmed, &payload); err == nil {
			return redactJSONValue(payload)
		}
	}

	return string(body)
}

func redactJSONValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		redacted := make(map[string]interface{}, len(typed))
		for key, nested := range typed {
			lowerKey := strings.ToLower(key)
			if lowerKey == "access_token" || lowerKey == "refresh_token" || lowerKey == "authorization" || lowerKey == "api_key" || lowerKey == "apikey" || lowerKey == "encrypted_key" || lowerKey == "password" || lowerKey == "token" {
				redacted[key] = "REDACTED"
				continue
			}
			redacted[key] = redactJSONValue(nested)
		}
		return redacted
	case []interface{}:
		redacted := make([]interface{}, len(typed))
		for i, nested := range typed {
			redacted[i] = redactJSONValue(nested)
		}
		return redacted
	default:
		return typed
	}
}
