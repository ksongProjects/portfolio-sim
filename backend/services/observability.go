package services

import (
	"context"
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
	checks []ServiceCheck
}

type ServiceCheck struct {
	Name         string
	HealthURL    string
	Timeout      time.Duration
}

func NewObservabilityService() *ObservabilityService {
	return &ObservabilityService{
		checks: []ServiceCheck{
			{Name: "main-api", HealthURL: "http://localhost:8080/health", Timeout: 500 * time.Millisecond},
			{Name: "market-data-service", HealthURL: "http://market-data-service:8080/health", Timeout: 500 * time.Millisecond},
			{Name: "news-feed-service", HealthURL: "http://news-feed-service:8080/health", Timeout: 500 * time.Millisecond},
			{Name: "analyst-service", HealthURL: "http://analyst-service:8080/health", Timeout: 500 * time.Millisecond},
		},
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
			status := "error"
			uptime := "0%"

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, check.HealthURL, nil)
			if err == nil {
				client := &http.Client{Timeout: check.Timeout}
				resp, err := client.Do(req)
				if err == nil {
					defer resp.Body.Close()
					if resp.StatusCode == http.StatusOK {
						body, _ := io.ReadAll(resp.Body)
						if len(body) > 0 && (string(body) == "ok" || string(body) == `"ok"` || contains(body, `"status":"ok"`)) {
							status = "healthy"
							uptime = "100%"
						} else {
							status = "healthy"
							uptime = "100%"
						}
					} else {
						status = "warning"
						uptime = "50%"
					}
				}
			}

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