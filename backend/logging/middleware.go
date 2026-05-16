package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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
			requestID := randomID()
			captureBody := shouldCaptureBody(r.URL.Path)
			var reqBody []byte
			if captureBody {
				reqBody, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewReader(reqBody))
			}

			route := getRoute(r)
			r = r.WithContext(WithRoute(r.Context(), route))
			operation := getOperationName(r.Method, r.URL.Path)
			contextMeta := getRequestContext(r, reqBody)

			reqMeta := map[string]interface{}{
				"type":            "api_request",
				"request_id":      requestID,
				"operation":       operation,
				"method":          r.Method,
				"path":            r.URL.Path,
				"query":           sanitizeURLQuery(r.URL.RawQuery),
				"request_headers": redactHeaders(r.Header),
				"remote_addr":     r.RemoteAddr,
				"route":           route,
			}
			for key, value := range contextMeta {
				reqMeta[key] = value
			}
			if len(reqBody) > 0 {
				reqMeta["request_body"] = sanitizeBody(r.Header.Get("Content-Type"), reqBody)
				reqMeta["request_body_size"] = len(reqBody)
			} else {
				reqMeta["request_body_skipped"] = true
			}
			client.log(r.Context(), "INFO", formatRequestMessage(operation, r.Method, r.URL.Path, contextMeta), reqMeta)

			wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK, captureBody: captureBody}
			next.ServeHTTP(w, r)

			durationMs := time.Since(start).Milliseconds()
			respMeta := map[string]interface{}{
				"type":               "api_response",
				"request_id":         requestID,
				"operation":          operation,
				"method":             r.Method,
				"path":               r.URL.Path,
				"query":              sanitizeURLQuery(r.URL.RawQuery),
				"status":             wrapper.statusCode,
				"duration_ms":        durationMs,
				"response_headers":   redactHeaders(wrapper.Header()),
				"response_body_size": wrapper.size,
				"route":              route,
			}
			for key, value := range contextMeta {
				respMeta[key] = value
			}
			if wrapper.statusCode >= 400 || captureBody && len(wrapper.body) > 0 {
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
			client.log(r.Context(), level, formatResponseMessage(operation, r.Method, r.URL.Path, wrapper.statusCode, durationMs, contextMeta), respMeta)
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
	"/api/connections":            true,
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

func getRoute(r *http.Request) string {
	if route := getRouteFromHeaders(r.Header); route != "" {
		return route
	}
	return getActionFromPath(r.URL.Path)
}

func getRouteFromHeaders(headers http.Header) string {
	for _, value := range []string{
		headers.Get("X-Frontend-Route"),
		headers.Get("Next-Url"),
		headers.Get("Referer"),
	} {
		if route := normalizeRouteHeaderValue(value); route != "" {
			return route
		}
	}

	return ""
}

func normalizeRouteHeaderValue(value string) string {
	path := headerValueToPath(value)
	if path == "" {
		return ""
	}

	route := getRouteFromPath(path)
	if route == "api" {
		return ""
	}

	return route
}

func headerValueToPath(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		u, err := url.Parse(trimmed)
		if err != nil {
			return ""
		}

		return u.Path
	}

	if strings.HasPrefix(trimmed, "/") {
		if index := strings.Index(trimmed, "?"); index >= 0 {
			return trimmed[:index]
		}

		return trimmed
	}

	return "/" + trimmed
}

func getRouteFromPath(path string) string {
	switch {
	case path == "/dashboard" || strings.HasPrefix(path, "/dashboard/"):
		return "dashboard"
	case path == "/portfolio" || strings.HasPrefix(path, "/portfolio/"):
		return "portfolio"
	case path == "/news-feed" || strings.HasPrefix(path, "/news-feed/"):
		return "news"
	case path == "/strategy" || strings.HasPrefix(path, "/strategy/"):
		return "strategy"
	case path == "/observability" || strings.HasPrefix(path, "/observability/"):
		return "observability"
	case path == "/settings" || strings.HasPrefix(path, "/settings/"):
		return "settings"
	case strings.HasPrefix(path, "/ticker/"):
		return "portfolio"
	}
	return "api"
}

func getActionFromPath(path string) string {
	if strings.HasPrefix(path, "/api/observability/") {
		return "observability"
	}
	if strings.HasPrefix(path, "/api/portfolio/") {
		return "portfolio"
	}
	if strings.HasPrefix(path, "/api/market/") || strings.HasPrefix(path, "/api/stream/market") {
		return "market"
	}
	if strings.HasPrefix(path, "/api/news") {
		return "news"
	}
	if strings.HasPrefix(path, "/api/strategies") {
		return "strategy"
	}
	if strings.HasPrefix(path, "/api/signals") {
		return "strategy"
	}
	if strings.HasPrefix(path, "/api/notifications") {
		return "notifications"
	}
	if strings.HasPrefix(path, "/api/providers") {
		return "settings"
	}
	if strings.HasPrefix(path, "/api/settings/") {
		return "settings"
	}
	if strings.HasPrefix(path, "/api/connections") {
		return "settings"
	}
	if strings.HasPrefix(path, "/api/rss-feeds") {
		return "news"
	}
	if strings.HasPrefix(path, "/api/channels") || strings.HasPrefix(path, "/api/videos") {
		return "news"
	}
	if strings.HasPrefix(path, "/api/tickers/") {
		return "portfolio"
	}
	return "api"
}

func getOperationName(method, path string) string {
	switch {
	case method == http.MethodGet && path == "/api/portfolio/positions":
		return "fetch portfolio positions"
	case method == http.MethodPost && path == "/api/portfolio/positions":
		return "add portfolio position"
	case method == http.MethodGet && path == "/api/portfolio/summary":
		return "fetch portfolio summary"
	case method == http.MethodGet && path == "/api/market/indices":
		return "fetch market indices"
	case method == http.MethodGet && path == "/api/settings/market-indices":
		return "fetch market index settings"
	case method == http.MethodPut && path == "/api/settings/market-indices":
		return "save market index settings"
	case method == http.MethodGet && path == "/api/news":
		return "fetch news feed"
	case method == http.MethodGet && path == "/api/strategies":
		return "fetch strategies"
	case method == http.MethodGet && path == "/api/signals":
		return "fetch signals"
	case method == http.MethodGet && path == "/api/notifications":
		return "fetch notifications"
	case method == http.MethodPost && path == "/api/notifications/dismiss":
		return "dismiss notification"
	case method == http.MethodGet && path == "/api/providers":
		return "fetch providers"
	case method == http.MethodPut && path == "/api/providers":
		return "save provider key"
	case method == http.MethodPost && path == "/api/providers/validate":
		return "validate provider key"
	case method == http.MethodPost && path == "/api/providers/questrade/oauth":
		return "save questrade oauth"
	case method == http.MethodGet && path == "/api/providers/questrade/oauth":
		return "fetch questrade oauth"
	case method == http.MethodPost && path == "/api/providers/questrade/refresh":
		return "refresh questrade token"
	case method == http.MethodGet && path == "/api/rss-feeds":
		return "fetch rss feeds"
	case method == http.MethodPost && path == "/api/rss-feeds":
		return "add rss feed"
	case method == http.MethodDelete && path == "/api/rss-feeds":
		return "delete rss feed"
	case method == http.MethodPost && path == "/api/rss-feeds/scrape":
		return "trigger rss scrape"
	case method == http.MethodGet && path == "/api/tickers/search":
		return "search tickers"
	case method == http.MethodGet && strings.HasPrefix(path, "/api/tickers/"):
		return "fetch ticker details"
	case method == http.MethodGet && path == "/api/channels":
		return "fetch youtube channels"
	case method == http.MethodGet && path == "/api/channels/search":
		return "search youtube channels"
	case method == http.MethodGet && path == "/api/videos/latest":
		return "fetch latest videos"
	case method == http.MethodGet && path == "/api/videos":
		return "fetch stored videos"
	case method == http.MethodPost && path == "/api/videos/analyze":
		return "analyze video"
	case method == http.MethodPost && path == "/api/videos/summarize":
		return "summarize videos"
	default:
		return fmt.Sprintf("%s %s", method, path)
	}
}

func formatRequestMessage(operation, method, path string, meta map[string]interface{}) string {
	return fmt.Sprintf("%s request: %s %s%s", operation, method, path, formatContextSuffix(meta))
}

func formatResponseMessage(operation, method, path string, status int, durationMs int64, meta map[string]interface{}) string {
	return fmt.Sprintf("%s response: %s %s -> %d (%dms)%s", operation, method, path, status, durationMs, formatContextSuffix(meta))
}

func formatContextSuffix(meta map[string]interface{}) string {
	parts := make([]string, 0, 4)

	for _, key := range []string{"provider", "symbol", "channel_id", "video_id", "portfolio_id", "notification_id", "feed_id", "query_term"} {
		if value, ok := meta[key]; ok {
			parts = append(parts, fmt.Sprintf("%s=%v", key, value))
		}
	}

	if len(parts) == 0 {
		return ""
	}

	return " [" + strings.Join(parts, " ") + "]"
}

func getRequestContext(r *http.Request, reqBody []byte) map[string]interface{} {
	meta := map[string]interface{}{}

	query := r.URL.Query()
	if provider := firstNonEmpty(query.Get("provider"), extractBodyString(reqBody, "provider_id")); provider != "" {
		meta["provider"] = provider
	}
	if symbol := firstNonEmpty(query.Get("symbol"), query.Get("q"), extractTickerSymbol(r.URL.Path)); symbol != "" {
		if strings.HasPrefix(r.URL.Path, "/api/tickers/search") || strings.HasPrefix(r.URL.Path, "/api/channels/search") {
			meta["query_term"] = symbol
		} else {
			meta["symbol"] = symbol
		}
	}
	if portfolioID := firstNonEmpty(query.Get("portfolio_id"), extractBodyString(reqBody, "portfolio_id")); portfolioID != "" {
		meta["portfolio_id"] = portfolioID
	}
	if channelID := firstNonEmpty(query.Get("channel_id"), extractBodyString(reqBody, "channel_id")); channelID != "" {
		meta["channel_id"] = channelID
	}
	if videoID := extractBodyString(reqBody, "video_id"); videoID != "" {
		meta["video_id"] = videoID
	}
	if feedID := query.Get("id"); feedID != "" {
		meta["feed_id"] = feedID
	}
	if notificationID := extractBodyString(reqBody, "id"); notificationID != "" && strings.HasPrefix(r.URL.Path, "/api/notifications/") {
		meta["notification_id"] = notificationID
	}

	return meta
}

func extractTickerSymbol(path string) string {
	if !strings.HasPrefix(path, "/api/tickers/") {
		return ""
	}

	trimmed := strings.TrimPrefix(path, "/api/tickers/")
	parts := strings.Split(trimmed, "/")
	if len(parts) == 0 || parts[0] == "" || parts[0] == "search" {
		return ""
	}

	return parts[0]
}

func extractBodyString(body []byte, key string) string {
	if len(body) == 0 {
		return ""
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}

	value, ok := payload[key]
	if !ok || value == nil {
		return ""
	}

	switch typed := value.(type) {
	case string:
		return typed
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return fmt.Sprint(typed)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
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
