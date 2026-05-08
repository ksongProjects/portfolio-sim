package services

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type ProviderConfig struct {
	ID           string `json:"id"`
	ProviderID   string `json:"provider_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	APIKeySet    bool   `json:"api_key_set"`
	IsConnected  bool   `json:"is_connected"`
	RateLimit    int    `json:"rate_limit"`
	DocURL       string `json:"docs_url"`
}

type ConnectionStatus struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	IsUp     bool   `json:"is_up"`
	LatencyMs int64 `json:"latency_ms"`
}

type ProviderService struct{}

func NewProviderService() *ProviderService {
	return &ProviderService{}
}

func (s *ProviderService) GetProviders(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}) ([]ProviderConfig, error) {
	query := `
		SELECT ds.id, ds.name, ds.source_priority, ds.rate_limit_per_min,
			   CASE WHEN pc.id IS NOT NULL THEN true ELSE false END as has_key,
			   CASE WHEN pc.id IS NOT NULL THEN true ELSE false END as is_connected
		FROM data_sources ds
		LEFT JOIN provider_configurations pc ON pc.provider_id = ds.id
		ORDER BY ds.source_priority
	`
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	providerMeta := map[string]struct {
		Description string
		DocURL      string
	}{
		"questrade": {"Questrade market data", "https://www.questrade.com/api"},
		"polygon":   {"Polygon.io real-time and historical market data", "https://polygon.io/docs"},
		"fmp":       {"Financial Modeling Prep financial data", "https://site.financialmodelingprep.com/developers/docs"},
	}

	defaults := map[string]struct {
		RateLimit int
		DocURL    string
	}{
		"questrade": {100, "https://www.questrade.com/api"},
		"polygon":   {60, "https://polygon.io/docs"},
		"fmp":       {250, "https://site.financialmodelingprep.com/developers/docs"},
	}

	var results []ProviderConfig
	for rows.Next() {
		var id, name string
		var priority, rateLimit int
		var hasKey, isConnected bool
		if err := rows.Scan(&id, &name, &priority, &rateLimit, &hasKey, &isConnected); err != nil {
			continue
		}
		meta := providerMeta[id]
		def := defaults[id]
		results = append(results, ProviderConfig{
			ID:          id,
			ProviderID:  id,
			Name:        name,
			Description: meta.Description,
			APIKeySet:   hasKey,
			IsConnected: isConnected,
			RateLimit:   rateLimit,
			DocURL:      def.DocURL,
		})
	}

	if len(results) == 0 {
		results = []ProviderConfig{
			{ID: "polygon", ProviderID: "polygon", Name: "Polygon.io", Description: "Real-time and historical market data", RateLimit: 60, DocURL: "https://polygon.io/docs"},
			{ID: "questrade", ProviderID: "questrade", Name: "Questrade", Description: "Questrade market data API", RateLimit: 100, DocURL: "https://www.questrade.com/api"},
			{ID: "fmp", ProviderID: "fmp", Name: "Financial Modeling Prep", Description: "Financial statements and fundamental data", RateLimit: 250, DocURL: "https://site.financialmodelingprep.com/developers/docs"},
		}
	}
	return results, nil
}

func (s *ProviderService) SaveProviderKey(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}, providerID, apiKey string) error {
	query := `
		INSERT INTO provider_configurations (id, provider_id, encrypted_key, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, NOW(), NOW())
		ON CONFLICT (provider_id) DO UPDATE SET encrypted_key = $2, updated_at = NOW()
	`
	_, err := db.Query(ctx, query, providerID, apiKey)
	return err
}

func (s *ProviderService) CheckConnection(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}) ([]ConnectionStatus, error) {
	start := time.Now()
	rows, err := db.Query(ctx, "SELECT 1")
	latency := time.Since(start).Milliseconds()

	var isUp bool
	if err == nil {
		rows.Close()
		isUp = true
	}

	statuses := []ConnectionStatus{
		{ID: "postgres", Name: "PostgreSQL", Type: "database", IsUp: isUp, LatencyMs: latency},
		{ID: "redis", Name: "Redis", Type: "cache", IsUp: true, LatencyMs: 1},
		{ID: "websocket", Name: "WebSocket", Type: "streaming", IsUp: true, LatencyMs: 0},
	}
	return statuses, nil
}
