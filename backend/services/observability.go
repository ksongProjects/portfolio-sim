package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type ServiceHealth struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Uptime   string `json:"uptime"`
	LastCheck string `json:"last_check"`
}

type ObservabilityService struct {
	checks     []ServiceCheck
	tracker    map[string]*serviceTracker
	mu         sync.Mutex
}

type serviceTracker struct {
	name         string
	uptimeCount  int
	totalChecks  int
	startTime    time.Time
}

type ServiceCheck struct {
	Name         string
	HealthURL    string
	Timeout      time.Duration
}

func NewObservabilityService() *ObservabilityService {
	return &ObservabilityService{
		checks: []ServiceCheck{
			{Name: "main-api", HealthURL: "http://main-api:8080/health", Timeout: 500 * time.Millisecond},
			{Name: "market-data-service", HealthURL: "http://market-data-service:8080/health", Timeout: 500 * time.Millisecond},
			{Name: "news-feed-service", HealthURL: "http://news-feed-service:8080/health", Timeout: 500 * time.Millisecond},
			{Name: "analyst-service", HealthURL: "http://analyst-service:8080/health", Timeout: 500 * time.Millisecond},
		},
		tracker: make(map[string]*serviceTracker),
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

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, check.HealthURL, nil)
			if err == nil {
				client := &http.Client{Timeout: check.Timeout}
				resp, err := client.Do(req)
				if err == nil {
					defer resp.Body.Close()
					if resp.StatusCode == http.StatusOK {
						body, _ := io.ReadAll(resp.Body)
						bodyStr := string(body)
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