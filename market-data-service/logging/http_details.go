package logging

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

func SanitizeURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	parsed.RawQuery = SanitizeQuery(parsed.RawQuery)
	return parsed.String()
}

func SanitizeQuery(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}

	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return rawQuery
	}

	for key := range query {
		lowerKey := strings.ToLower(key)
		if lowerKey == "access_token" || lowerKey == "refresh_token" || lowerKey == "apikey" || lowerKey == "api_key" || lowerKey == "key" || lowerKey == "token" {
			query.Set(key, "REDACTED")
		}
	}

	return query.Encode()
}

func RedactHeaders(headers http.Header) map[string][]string {
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

func SanitizeBody(contentType string, body []byte) interface{} {
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
