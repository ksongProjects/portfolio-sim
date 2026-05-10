package services

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/portfolio-sim/backend/logging"
)

type ServiceHealth struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Uptime    string `json:"uptime"`
	LastCheck string `json:"last_check"`
}

type ObservabilityService struct {
	checks     []ServiceCheck
	tracker    map[string]*serviceTracker
	mu         sync.Mutex
	logClient  *logging.Client
}

type serviceTracker struct {
	name        string
	uptimeCount int
	totalChecks int
	startTime   time.Time
}

type ServiceCheck struct {
	Name      string
	HealthURL string
	Timeout   time.Duration
}

func NewObservabilityService(logClient *logging.Client) *ObservabilityService {
	return &ObservabilityService{
		checks: []ServiceCheck{
			{Name: "main-api", HealthURL: "http://main-api:8080/health", Timeout: 500 * time.Millisecond},
			{Name: "market-data-service", HealthURL: "http://market-data-service:8080/health", Timeout: 500 * time.Millisecond},
			{Name: "news-feed-service", HealthURL: "http://news-feed-service:8080/health", Timeout: 500 * time.Millisecond},
			{Name: "analyst-service", HealthURL: "http://analyst-service:8080/health", Timeout: 500 * time.Millisecond},
		},
		tracker:   make(map[string]*serviceTracker),
		logClient: logClient,
	}
}

func (s *ObservabilityService) CheckServices(ctx context.Context) []ServiceHealth {
	results := make([]ServiceHealth, len(s.checks))
	now := time.Now()
	var wg sync.WaitGroup

	for i, check := range s.checks {
		wg.Add(1)
		go func(idx int, check ServiceCheck) {
			defer wg.Done()

			s.mu.Lock()
			if s.tracker[check.Name] == nil {
				s.tracker[check.Name] = &serviceTracker{
					name:      check.Name,
					startTime: now,
				}
			}
			tracker := s.tracker[check.Name]
			tracker.totalChecks++
			s.mu.Unlock()

			status := "error"
			healthy := false

			hostPort := check.HealthURL[strings.Index(check.HealthURL, "://")+3:]
			hostPort = strings.Split(hostPort, "/")[0]

			if s.logClient != nil {
				s.logClient.InfoWithMeta(ctx, "Health check request", map[string]interface{}{
					"type":    "health_check_request",
					"target":  check.Name,
					"url":     check.HealthURL,
					"timeout": check.Timeout.String(),
				})
			}

			dialer := &net.Dialer{Timeout: check.Timeout}
			conn, err := dialer.DialContext(ctx, "tcp", hostPort)
			if err == nil {
				conn.Close()
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, check.HealthURL, nil)
				if err == nil {
					client := &http.Client{Timeout: check.Timeout}
					resp, err := client.Do(req)
					if err == nil {
						defer resp.Body.Close()
						body, _ := io.ReadAll(resp.Body)
						bodyStr := string(body)

						if s.logClient != nil {
							s.logClient.InfoWithMeta(ctx, "Health check response", map[string]interface{}{
								"type":   "health_check_response",
								"target": check.Name,
								"status": resp.StatusCode,
								"body":   bodyStr,
							})
						}

						if resp.StatusCode == http.StatusOK {
							if bodyStr == "ok" || bodyStr == `"ok"` || contains(body, `"status":"ok"`) {
								status = "healthy"
								healthy = true
							} else {
								status = "healthy"
								healthy = true
							}
						} else {
							status = "warning"
						}
					} else {
						if s.logClient != nil {
							s.logClient.ErrorWithMeta(ctx, "Health check failed", map[string]interface{}{
								"type":  "health_check_error",
								"error": err.Error(),
							})
						}
						status = "unreachable"
					}
				}
			} else {
				if s.logClient != nil {
					s.logClient.ErrorWithMeta(ctx, "Health check connection failed", map[string]interface{}{
						"type":  "health_check_error",
						"error": err.Error(),
					})
				}
			}

			if healthy {
				s.mu.Lock()
				tracker.uptimeCount++
				s.mu.Unlock()
			}

			s.mu.Lock()
			uptime := "0%"
			if tracker.totalChecks > 0 {
				uptimePct := float64(tracker.uptimeCount) / float64(tracker.totalChecks) * 100
				uptime = fmt.Sprintf("%.0f%%", uptimePct)
			}
			s.mu.Unlock()

			results[idx] = ServiceHealth{
				Name:      check.Name,
				Status:    status,
				Uptime:    uptime,
				LastCheck: formatTimeAgo(now),
			}
		}(i, check)
	}

	wg.Wait()
	return results
}

func contains(b []byte, substr string) bool {
	return len(b) >= len(substr) && (string(b) == substr || len(b) > 0 && (string(b[:len(substr)]) == substr || len(b) > len(substr) && string(b[len(b)-len(substr):]) == substr))
}

func formatTimeAgo(t time.Time) string {
	return "Just now"
}