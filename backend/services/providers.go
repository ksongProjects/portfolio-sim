package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/portfolio-sim/backend/logging"
	"github.com/portfolio-sim/shared/secrets"
)

type ProviderService struct {
	logger    *slog.Logger
	logClient *logging.Client
	codec     *secrets.Codec
}

func NewProviderService(logger *slog.Logger, logClient *logging.Client, codec *secrets.Codec) *ProviderService {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return &ProviderService{logger: logger, logClient: logClient, codec: codec}
}

type ProviderConfig struct {
	ID           string `json:"id"`
	ProviderID   string `json:"provider_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Type         string `json:"type"`
	APIKeySet    bool   `json:"api_key_set"`
	IsConnected  bool   `json:"is_connected"`
	TokenExpired bool   `json:"token_expired"`
	RateLimit    int    `json:"rate_limit"`
	DocURL       string `json:"docs_url"`
}

type ConnectionStatus struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	IsUp      bool   `json:"is_up"`
	LatencyMs int64  `json:"latency_ms"`
}

func (s *ProviderService) GetProviders(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}) ([]ProviderConfig, error) {
	query := `
		SELECT ds.id, ds.name, ds.source_priority, ds.rate_limit_per_min,
			   CASE WHEN pc.id IS NOT NULL THEN true ELSE false END as has_key,
			   CASE WHEN pc.id IS NOT NULL THEN true ELSE false END as is_connected,
			   pc.token_expires_at
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
		Type        string
		DocURL      string
	}{
		"questrade": {"Questrade market data", "market_data", "https://www.questrade.com/api"},
		"massive":   {"Massive real-time and historical market data", "market_data", "https://api.massive.io/docs"},
		"fmp":       {"Financial Modeling Prep financial data", "market_data", "https://site.financialmodelingprep.com/developer/docs"},
		"youtube":   {"YouTube Data API for video transcripts", "youtube", "https://developers.google.com/youtube/v3"},
		"gemini":    {"Google Gemini API for content summarization", "gemini", "https://ai.google.dev/docs"},
	}

	defaults := map[string]struct {
		RateLimit int
		DocURL    string
	}{
		"questrade": {100, "https://www.questrade.com/api"},
		"massive":   {60, "https://api.massive.io/docs"},
		"fmp":       {250, "https://site.financialmodelingprep.com/developer/docs"},
		"youtube":   {0, "https://developers.google.com/youtube/v3"},
		"gemini":    {0, "https://ai.google.dev/docs"},
	}

	var results []ProviderConfig
	for rows.Next() {
		var id, name string
		var priority, rateLimit int
		var hasKey, isConnected bool
		var tokenExpiresAt *time.Time
		if err := rows.Scan(&id, &name, &priority, &rateLimit, &hasKey, &isConnected, &tokenExpiresAt); err != nil {
			continue
		}
		meta := providerMeta[id]
		def := defaults[id]
		tokenExpired := tokenExpiresAt != nil && tokenExpiresAt.Before(time.Now())
		results = append(results, ProviderConfig{
			ID:           id,
			ProviderID:   id,
			Name:         name,
			Description:  meta.Description,
			Type:         meta.Type,
			APIKeySet:    hasKey,
			IsConnected:  isConnected && !tokenExpired,
			TokenExpired: tokenExpired,
			RateLimit:    rateLimit,
			DocURL:       def.DocURL,
		})
	}

	if len(results) == 0 {
		results = []ProviderConfig{
			{ID: "massive", ProviderID: "massive", Name: "Massive", Description: "Real-time and historical market data", Type: "market_data", RateLimit: 60, DocURL: "https://api.massive.io/docs", TokenExpired: false},
			{ID: "questrade", ProviderID: "questrade", Name: "Questrade", Description: "Questrade market data API", Type: "market_data", RateLimit: 100, DocURL: "https://www.questrade.com/api", TokenExpired: false},
			{ID: "fmp", ProviderID: "fmp", Name: "Financial Modeling Prep", Description: "Financial statements and fundamental data", Type: "market_data", RateLimit: 250, DocURL: "https://site.financialmodelingprep.com/developer/docs", TokenExpired: false},
			{ID: "youtube", ProviderID: "youtube", Name: "YouTube Data API", Description: "YouTube Data API for video transcripts", Type: "youtube", RateLimit: 0, DocURL: "https://developers.google.com/youtube/v3", TokenExpired: false},
			{ID: "gemini", ProviderID: "gemini", Name: "Google Gemini", Description: "Gemini API for content summarization", Type: "gemini", RateLimit: 0, DocURL: "https://ai.google.dev/docs", TokenExpired: false},
		}
	} else {
		seen := make(map[string]bool)
		for _, p := range results {
			seen[p.ID] = true
		}
		if !seen["youtube"] {
			results = append(results, ProviderConfig{ID: "youtube", ProviderID: "youtube", Name: "YouTube Data API", Description: "YouTube Data API for video transcripts", Type: "youtube", RateLimit: 0, DocURL: "https://developers.google.com/youtube/v3"})
		}
		if !seen["gemini"] {
			results = append(results, ProviderConfig{ID: "gemini", ProviderID: "gemini", Name: "Google Gemini", Description: "Gemini API for content summarization", Type: "gemini", RateLimit: 0, DocURL: "https://ai.google.dev/docs"})
		}
	}
	return results, nil
}

func (s *ProviderService) SaveProviderKey(ctx context.Context, db interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (int64, error)
}, providerID, apiKey string) error {
	encryptedKey, err := s.codec.EncryptString(apiKey)
	if err != nil {
		return err
	}
	query := `
		INSERT INTO provider_configurations (id, provider_id, encrypted_key, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, NOW(), NOW())
		ON CONFLICT (provider_id) DO UPDATE SET encrypted_key = $2, updated_at = NOW()
	`
	_, err = db.Exec(ctx, query, providerID, encryptedKey)
	return err
}

func (s *ProviderService) SaveQuestradeOAuth(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...interface{}) (int64, error)
}, providerID, accessToken, refreshToken, apiServer string, expiresIn int) error {
	encryptedAccessToken, err := s.codec.EncryptString(accessToken)
	if err != nil {
		return err
	}
	encryptedRefreshToken, err := s.codec.EncryptString(refreshToken)
	if err != nil {
		return err
	}
	encryptedAPIServer, err := s.codec.EncryptString(apiServer)
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)
	query := `
		INSERT INTO provider_configurations (id, provider_id, encrypted_key, access_token, refresh_token, api_server, token_expires_at, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, '', $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (provider_id) DO UPDATE SET access_token = $2, refresh_token = $3, api_server = $4, token_expires_at = $5, updated_at = NOW()
	`
	_, err = db.Exec(ctx, query, providerID, encryptedAccessToken, encryptedRefreshToken, encryptedAPIServer, expiresAt)
	return err
}

func (s *ProviderService) GetQuestradeOAuth(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) (pgx.Row, error)
}, providerID string) (accessToken, refreshToken, apiServer string) {
	query := `
		SELECT COALESCE(access_token, ''), COALESCE(refresh_token, ''), COALESCE(api_server, '')
		FROM provider_configurations
		WHERE provider_id = $1
	`
	row, _ := db.QueryRow(ctx, query, providerID)
	var encryptedAccessToken, encryptedRefreshToken, encryptedAPIServer string
	if err := row.Scan(&encryptedAccessToken, &encryptedRefreshToken, &encryptedAPIServer); err != nil {
		return "", "", ""
	}
	accessToken, _ = s.codec.DecryptString(encryptedAccessToken)
	refreshToken, _ = s.codec.DecryptString(encryptedRefreshToken)
	apiServer, _ = s.codec.DecryptString(encryptedAPIServer)
	return
}

func (s *ProviderService) ExchangeQuestradeToken(ctx context.Context, initialRefreshToken string) (accessToken, newRefreshToken, apiServer string, expiresIn int, err error) {
	const qtAPI = "https://login.questrade.com/oauth2/token"
	tokenURL := fmt.Sprintf("%s?grant_type=refresh_token&refresh_token=%s", qtAPI, url.QueryEscape(initialRefreshToken))

	if s.logClient != nil {
		s.logClient.InfoWithMeta(ctx, "Questrade token exchange request", map[string]interface{}{
			"url":    sanitizeProviderURL(tokenURL),
			"method": "GET",
			"type":   "provider_api_request",
		})
	}

	resp, err := http.Get(tokenURL)
	if err != nil {
		return "", "", "", 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", 0, err
	}

	if s.logClient != nil {
		s.logClient.InfoWithMeta(ctx, "Questrade token exchange response", map[string]interface{}{
			"status": resp.StatusCode,
			"body":   sanitizeProviderBody(body),
			"type":   "provider_api_response",
		})
	}

	if resp.StatusCode == http.StatusBadRequest {
		return "", "", "", 0, fmt.Errorf("questrade refresh token is expired or invalid")
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return "", "", "", 0, fmt.Errorf("questrade unauthorized")
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", "", 0, fmt.Errorf("questrade token exchange failed: %d", resp.StatusCode)
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		APIServer    string `json:"api_server"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", "", 0, err
	}
	return result.AccessToken, result.RefreshToken, result.APIServer, result.ExpiresIn, nil
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
	}
	return statuses, nil
}

func (s *ProviderService) ValidateProviderKey(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}, providerID, apiKey string) (bool, *QuestradeOAuthResult, error) {
	if s.logClient != nil {
		s.logClient.InfoWithMeta(ctx, "Validating provider key", map[string]interface{}{
			"provider":       providerID,
			"has_api_key":    apiKey != "",
			"api_key_length": len(apiKey),
			"type":           "provider_validation_request",
		})
	}

	var valid bool
	var qtResult *QuestradeOAuthResult
	var err error

	switch providerID {
	case "massive":
		valid, err = s.validateMassiveKey(ctx, apiKey)
	case "fmp":
		valid, err = s.validateFMPKey(ctx, apiKey)
	case "questrade":
		valid, qtResult, err = s.validateQuestradeKey(ctx, apiKey)
	case "youtube":
		valid, err = s.validateYouTubeKey(ctx, apiKey)
	case "gemini":
		valid, err = s.validateGeminiKey(ctx, apiKey)
	default:
		err = fmt.Errorf("unknown provider: %s", providerID)
	}

	if err != nil {
		if s.logClient != nil {
			s.logClient.ErrorWithMeta(ctx, "Provider validation failed", map[string]interface{}{
				"provider": providerID,
				"error":    err.Error(),
				"type":     "provider_validation_error",
			})
		}
		return false, nil, err
	}

	if s.logClient != nil {
		s.logClient.InfoWithMeta(ctx, "Provider validation succeeded", map[string]interface{}{
			"provider": providerID,
			"valid":    valid,
			"type":     "provider_validation_success",
		})
	}

	return valid, qtResult, nil
}

type QuestradeOAuthResult struct {
	AccessToken  string
	RefreshToken string
	APIServer    string
	ExpiresIn    int
}

func (s *ProviderService) validateMassiveKey(ctx context.Context, apiKey string) (bool, error) {
	url := fmt.Sprintf("https://api.massive.io/v2/aggs/ticker/AAPL/prev?adjusted=true&apiKey=%s", apiKey)

	if s.logClient != nil {
		s.logClient.InfoWithMeta(ctx, "Massive validation request", map[string]interface{}{
			"url":    sanitizeProviderURL(url),
			"method": "GET",
			"type":   "provider_api_request",
		})
	}

	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	if s.logClient != nil {
		s.logClient.InfoWithMeta(ctx, "Massive validation response", map[string]interface{}{
			"status": resp.StatusCode,
			"body":   sanitizeProviderBody(body),
			"type":   "provider_api_response",
		})
	}

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusPaymentRequired {
		return false, nil
	}
	return false, fmt.Errorf("massive returned status: %d", resp.StatusCode)
}

func (s *ProviderService) validateFMPKey(ctx context.Context, apiKey string) (bool, error) {
	url := fmt.Sprintf("https://financialmodelingprep.com/stable/quote?symbol=AAPL&apikey=%s", apiKey)

	if s.logClient != nil {
		s.logClient.InfoWithMeta(ctx, "FMP validation request", map[string]interface{}{
			"url":    sanitizeProviderURL(url),
			"method": "GET",
			"type":   "provider_api_request",
		})
	}

	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	if s.logClient != nil {
		s.logClient.InfoWithMeta(ctx, "FMP validation response", map[string]interface{}{
			"status": resp.StatusCode,
			"body":   sanitizeProviderBody(body),
			"type":   "provider_api_response",
		})
	}

	if resp.StatusCode == http.StatusOK {
		var result []struct {
			Symbol string `json:"symbol"`
		}
		if err := json.Unmarshal(body, &result); err == nil && len(result) > 0 {
			return true, nil
		}
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusPaymentRequired {
		return false, nil
	}
	if resp.StatusCode == http.StatusForbidden {
		var errResp struct {
			Message string `json:"Error Message"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Message != "" {
			return false, fmt.Errorf("fmp forbidden: %s", errResp.Message)
		}
	}
	return false, fmt.Errorf("fmp returned status: %d", resp.StatusCode)
}

func (s *ProviderService) validateQuestradeKey(ctx context.Context, apiKey string) (bool, *QuestradeOAuthResult, error) {
	const qtAPI = "https://login.questrade.com/oauth2/token"
	tokenURL := fmt.Sprintf("%s?grant_type=refresh_token&refresh_token=%s", qtAPI, url.QueryEscape(apiKey))

	if s.logClient != nil {
		s.logClient.InfoWithMeta(ctx, "Questrade validation request", map[string]interface{}{
			"url":    sanitizeProviderURL(tokenURL),
			"method": "GET",
			"type":   "provider_api_request",
		})
	}

	resp, err := http.Get(tokenURL)
	if err != nil {
		return false, nil, fmt.Errorf("oauth request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if s.logClient != nil {
		s.logClient.InfoWithMeta(ctx, "Questrade validation response", map[string]interface{}{
			"status": resp.StatusCode,
			"body":   sanitizeProviderBody(body),
			"type":   "provider_api_response",
		})
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusBadRequest {
			return false, nil, fmt.Errorf("questrade refresh token is expired or invalid")
		}
		if resp.StatusCode == http.StatusUnauthorized {
			return false, nil, fmt.Errorf("questrade unauthorized")
		}
		return false, nil, fmt.Errorf("questrade returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		APIServer    string `json:"api_server"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, nil, fmt.Errorf("failed to parse oauth response: %w", err)
	}
	if result.AccessToken == "" || result.RefreshToken == "" || result.APIServer == "" {
		return false, nil, fmt.Errorf("oauth response missing required fields")
	}
	return true, &QuestradeOAuthResult{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		APIServer:    result.APIServer,
		ExpiresIn:    result.ExpiresIn,
	}, nil
}

func (s *ProviderService) validateYouTubeKey(ctx context.Context, apiKey string) (bool, error) {
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?part=snippet&id=dQw4w9WgXcQ&key=%s", apiKey)

	if s.logClient != nil {
		s.logClient.InfoWithMeta(ctx, "YouTube validation request", map[string]interface{}{
			"url":    "https://www.googleapis.com/youtube/v3/videos?part=snippet&id=VIDEO_ID",
			"method": "GET",
			"type":   "provider_api_request",
		})
	}

	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	if s.logClient != nil {
		s.logClient.InfoWithMeta(ctx, "YouTube validation response", map[string]interface{}{
			"status": resp.StatusCode,
			"body":   sanitizeProviderBody(body),
			"type":   "provider_api_response",
		})
	}

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusBadRequest {
		return false, nil
	}
	return false, fmt.Errorf("youtube returned status: %d", resp.StatusCode)
}

func (s *ProviderService) validateGeminiKey(ctx context.Context, apiKey string) (bool, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", apiKey)

	if s.logClient != nil {
		s.logClient.InfoWithMeta(ctx, "Gemini validation request", map[string]interface{}{
			"url":    sanitizeProviderURL(url),
			"method": "GET",
			"type":   "provider_api_request",
		})
	}

	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	if s.logClient != nil {
		s.logClient.InfoWithMeta(ctx, "Gemini validation response", map[string]interface{}{
			"status": resp.StatusCode,
			"body":   sanitizeProviderBody(body),
			"type":   "provider_api_response",
		})
	}

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusBadRequest {
		return false, nil
	}
	return false, fmt.Errorf("gemini returned status: %d", resp.StatusCode)
}

type RSSFeed struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	URL          string `json:"url"`
	IsActive     bool   `json:"is_active"`
	LastScrapeAt string `json:"last_scrape_at"`
}

func (s *ProviderService) GetRSSFeeds(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}) ([]RSSFeed, error) {
	query := `
		SELECT id, name, url,
			   COALESCE(last_scrape_at::text, ''),
			   is_active
		FROM rss_feeds
		ORDER BY name
	`
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []RSSFeed
	for rows.Next() {
		var f RSSFeed
		if err := rows.Scan(&f.ID, &f.Name, &f.URL, &f.LastScrapeAt, &f.IsActive); err != nil {
			continue
		}
		feeds = append(feeds, f)
	}
	return feeds, nil
}

func (s *ProviderService) AddRSSFeed(ctx context.Context, db interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (int64, error)
}, name, url string) error {
	query := `
		INSERT INTO rss_feeds (id, name, url, is_active)
		VALUES (gen_random_uuid(), $1, $2, true)
	`
	_, err := db.Exec(ctx, query, name, url)
	return err
}

func (s *ProviderService) DeleteRSSFeed(ctx context.Context, db interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (int64, error)
}, feedID string) error {
	query := `DELETE FROM rss_feeds WHERE id = $1`
	_, err := db.Exec(ctx, query, feedID)
	return err
}

func sanitizeProviderURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	query := parsed.Query()
	for key := range query {
		lowerKey := strings.ToLower(key)
		if lowerKey == "apikey" || lowerKey == "api_key" || lowerKey == "key" || lowerKey == "token" || lowerKey == "access_token" || lowerKey == "refresh_token" {
			query.Set(key, "REDACTED")
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func sanitizeProviderBody(body []byte) interface{} {
	if len(body) == 0 {
		return ""
	}
	var payload interface{}
	if err := json.Unmarshal(body, &payload); err == nil {
		return redactProviderValue(payload)
	}
	return string(body)
}

func redactProviderValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		redacted := make(map[string]interface{}, len(typed))
		for key, nested := range typed {
			lowerKey := strings.ToLower(key)
			if lowerKey == "access_token" || lowerKey == "refresh_token" || lowerKey == "api_key" || lowerKey == "apikey" || lowerKey == "token" {
				redacted[key] = "REDACTED"
				continue
			}
			redacted[key] = redactProviderValue(nested)
		}
		return redacted
	case []interface{}:
		redacted := make([]interface{}, len(typed))
		for i, nested := range typed {
			redacted[i] = redactProviderValue(nested)
		}
		return redacted
	default:
		return typed
	}
}
