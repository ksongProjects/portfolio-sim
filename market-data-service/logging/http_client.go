package logging

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"
)

type LoggingClient struct {
	client      *http.Client
	logClient   *Client
	serviceName string
}

func NewLoggingClient(logClient *Client, serviceName string) *LoggingClient {
	return &LoggingClient{
		client:      &http.Client{Timeout: 10 * time.Second},
		logClient:   logClient,
		serviceName: serviceName,
	}
}

func (c *LoggingClient) Get(ctx context.Context, url string, meta map[string]interface{}) (*http.Response, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}

	c.logRequest(ctx, http.MethodGet, url, nil, meta)

	start := time.Now()
	resp, err := c.client.Do(req)
	duration := time.Since(start)

	if err != nil {
		c.logResponseError(ctx, http.MethodGet, url, 0, nil, err.Error(), duration, meta)
		return nil, nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logResponseError(ctx, http.MethodGet, url, resp.StatusCode, nil, err.Error(), duration, meta)
		return resp, nil, err
	}
	resp.Body.Close()

	c.logResponse(ctx, http.MethodGet, url, resp.StatusCode, body, duration, meta)

	if resp.StatusCode >= 400 {
		return resp, body, nil
	}

	return resp, body, nil
}

func (c *LoggingClient) Post(ctx context.Context, url string, body []byte, meta map[string]interface{}) (*http.Response, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	c.logRequest(ctx, http.MethodPost, url, body, meta)

	start := time.Now()
	resp, err := c.client.Do(req)
	duration := time.Since(start)

	if err != nil {
		c.logResponseError(ctx, http.MethodPost, url, 0, nil, err.Error(), duration, meta)
		return nil, nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logResponseError(ctx, http.MethodPost, url, resp.StatusCode, nil, err.Error(), duration, meta)
		return resp, nil, err
	}
	resp.Body.Close()

	c.logResponse(ctx, http.MethodPost, url, resp.StatusCode, respBody, duration, meta)

	if resp.StatusCode >= 400 {
		return resp, respBody, nil
	}

	return resp, respBody, nil
}

func (c *LoggingClient) logRequest(ctx context.Context, method, url string, body []byte, meta map[string]interface{}) {
	metadata := map[string]interface{}{
		"type":         "outbound_request",
		"method":       method,
		"url":          url,
		"service":      c.serviceName,
	}
	if body != nil && len(body) > 0 {
		metadata["body_size"] = len(body)
	}
	if meta != nil {
		for k, v := range meta {
			metadata[k] = v
		}
	}
	c.logClient.InfoWithMeta(ctx, "HTTP Request: "+method+" "+url, metadata)
}

func (c *LoggingClient) logResponse(ctx context.Context, method, url string, status int, body []byte, duration time.Duration, meta map[string]interface{}) {
	level := "INFO"
	if status >= 500 {
		level = "ERROR"
	} else if status >= 400 {
		level = "WARN"
	}

	metadata := map[string]interface{}{
		"type":         "outbound_response",
		"method":       method,
		"url":          url,
		"status":       status,
		"duration_ms":  duration.Milliseconds(),
		"body_size":    len(body),
		"service":      c.serviceName,
	}
	if body != nil && len(body) < 500 {
		metadata["body"] = string(body)
	}
	if status >= 400 {
		metadata["error"] = true
	}
	if meta != nil {
		for k, v := range meta {
			metadata[k] = v
		}
	}

	c.logClient.EmitWithMetadata(ctx, level, "HTTP Response: "+method+" "+url+" -> "+string(rune(status)), metadata)
}

func (c *LoggingClient) logResponseError(ctx context.Context, method, url string, status int, body []byte, errMsg string, duration time.Duration, meta map[string]interface{}) {
	metadata := map[string]interface{}{
		"type":        "outbound_response_error",
		"method":      method,
		"url":         url,
		"status":      status,
		"duration_ms": duration.Milliseconds(),
		"error":       errMsg,
		"service":     c.serviceName,
	}
	if meta != nil {
		for k, v := range meta {
			metadata[k] = v
		}
	}
	c.logClient.ErrorWithMeta(ctx, "HTTP Response Error: "+method+" "+url, metadata)
}