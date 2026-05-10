package providers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/portfolio-sim/market-data-service/config"
	"github.com/portfolio-sim/market-data-service/logging"
)

type PolygonProvider struct {
	cfg       config.PolygonConfig
	client    *http.Client
	logClient *logging.Client
}

func NewPolygonProvider(cfg config.PolygonConfig, logClient *logging.Client) *PolygonProvider {
	return &PolygonProvider{
		cfg:       cfg,
		client:    &http.Client{Timeout: 10 * time.Second},
		logClient: logClient,
	}
}

func (p *PolygonProvider) Name() string {
	return "polygon"
}

func (p *PolygonProvider) GetSymbolID(ticker string) (int, error) {
	return 0, fmt.Errorf("polygon does not use symbol IDs")
}

func (p *PolygonProvider) FetchCompanyProfile(ticker string) (*CompanyProfile, error) {
	return nil, fmt.Errorf("polygon does not support company profile")
}

func (p *PolygonProvider) FetchFinancialRatios(ticker string) ([]*FinancialRatio, error) {
	return nil, fmt.Errorf("polygon does not support financial ratios")
}

func (p *PolygonProvider) logRequest(method, rawURL string, headers http.Header) {
	if p.logClient == nil {
		return
	}
	safeURL := logging.SanitizeURL(rawURL)
	p.logClient.InfoWithMeta(nil, "Polygon Request: "+method+" "+safeURL, map[string]interface{}{
		"method":          method,
		"url":             safeURL,
		"provider":        p.Name(),
		"request_headers": logging.RedactHeaders(headers),
	})
}

func (p *PolygonProvider) logResponse(method, rawURL string, status int, headers http.Header, body []byte, duration time.Duration) {
	if p.logClient == nil {
		return
	}
	level := "INFO"
	if status >= 400 {
		level = "ERROR"
	}
	safeURL := logging.SanitizeURL(rawURL)
	p.logClient.EmitWithMetadata(nil, level, "Polygon Response: "+method+" "+safeURL+" -> "+fmt.Sprintf("%d", status), map[string]interface{}{
		"method":           method,
		"url":              safeURL,
		"status":           status,
		"provider":         p.Name(),
		"duration_ms":      duration.Milliseconds(),
		"response_headers": logging.RedactHeaders(headers),
		"body":             logging.SanitizeBody(headers.Get("Content-Type"), body),
		"body_size":        len(body),
	})
}

func (p *PolygonProvider) logError(method, rawURL string, headers http.Header, errMsg string, duration time.Duration) {
	if p.logClient == nil {
		return
	}
	safeURL := logging.SanitizeURL(rawURL)
	p.logClient.ErrorWithMeta(nil, "Polygon Error: "+method+" "+safeURL, map[string]interface{}{
		"method":          method,
		"url":             safeURL,
		"provider":        p.Name(),
		"request_headers": logging.RedactHeaders(headers),
		"duration_ms":     duration.Milliseconds(),
		"error":           errMsg,
	})
}

func (p *PolygonProvider) get(rawURL string) ([]byte, int, http.Header, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, 0, nil, err
	}

	p.logRequest(http.MethodGet, rawURL, req.Header)
	start := time.Now()
	resp, err := p.client.Do(req)
	duration := time.Since(start)
	if err != nil {
		p.logError(http.MethodGet, rawURL, req.Header, err.Error(), duration)
		return nil, 0, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		p.logError(http.MethodGet, rawURL, req.Header, err.Error(), duration)
		return nil, resp.StatusCode, resp.Header, err
	}

	p.logResponse(http.MethodGet, rawURL, resp.StatusCode, resp.Header, body, duration)
	return body, resp.StatusCode, resp.Header, nil
}

func (p *PolygonProvider) FetchPrice(ticker string) (*Price, error) {
	url := fmt.Sprintf("https://api.polygon.io/v2/aggs/ticker/%s/prev?adjusted=true&apiKey=%s", ticker, p.cfg.APIKey)
	body, status, _, err := p.get(url)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("polygon request failed: %d - %s", status, string(body))
	}

	var result struct {
		Results []struct {
			Ticker    string  `json:"T"`
			Open      float64 `json:"o"`
			High      float64 `json:"h"`
			Low       float64 `json:"l"`
			Close     float64 `json:"c"`
			Volume    int64   `json:"v"`
			Timestamp int64   `json:"t"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if len(result.Results) == 0 {
		return nil, fmt.Errorf("no data for %s", ticker)
	}

	r := result.Results[0]
	return &Price{
		Ticker:    ticker,
		Price:     r.Close,
		Volume:    r.Volume,
		Source:    p.Name(),
		Timestamp: time.UnixMilli(r.Timestamp),
	}, nil
}

func (p *PolygonProvider) FetchOptionChain(ticker string) ([]*OptionChain, error) {
	return nil, fmt.Errorf("polygon does not support option chains via this endpoint")
}

func (p *PolygonProvider) FetchIntradayBars(ticker string, interval string) ([]*IntradayBar, error) {
	multiplier := "1"
	timespan := interval
	switch interval {
	case "1m":
		timespan = "minute"
	case "5m":
		timespan = "minute"
		multiplier = "5"
	case "1h":
		timespan = "hour"
	default:
		timespan = "minute"
	}

	from := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	to := time.Now().Format("2006-01-02")

	url := fmt.Sprintf("https://api.polygon.io/v2/aggs/ticker/%s/range/%s/%s/%s/%s?adjusted=true&apiKey=%s",
		ticker, multiplier, timespan, from, to, p.cfg.APIKey)
	body, status, _, err := p.get(url)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("polygon request failed: %d - %s", status, string(body))
	}

	var result struct {
		Results []struct {
			Open      float64 `json:"o"`
			High      float64 `json:"h"`
			Low       float64 `json:"l"`
			Close     float64 `json:"c"`
			Volume    int64   `json:"v"`
			Timestamp int64   `json:"t"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	bars := make([]*IntradayBar, len(result.Results))
	for i, b := range result.Results {
		bars[i] = &IntradayBar{
			Ticker:    ticker,
			Interval:  interval,
			Open:      b.Open,
			High:      b.High,
			Low:       b.Low,
			Close:     b.Close,
			Volume:    b.Volume,
			Timestamp: time.UnixMilli(b.Timestamp),
		}
	}
	return bars, nil
}

func (p *PolygonProvider) SearchTickers(prefix string) ([]TickerSearchResult, error) {
	url := fmt.Sprintf("https://api.polygon.io/v3/reference/tickers?search=%s&active=true&apiKey=%s", url.QueryEscape(prefix), p.cfg.APIKey)
	body, status, _, err := p.get(url)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("polygon search failed: %d - %s", status, string(body))
	}

	var result struct {
		Results []struct {
			Ticker   string `json:"ticker"`
			Name     string `json:"name"`
			Exchange string `json:"exchange"`
			Type     string `json:"type"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	results := make([]TickerSearchResult, 0, len(result.Results))
	for _, t := range result.Results {
		results = append(results, TickerSearchResult{
			Symbol:   t.Ticker,
			Name:     t.Name,
			Exchange: t.Exchange,
			Type:     t.Type,
		})
	}
	return results, nil
}
