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
	"time"

	"github.com/portfolio-sim/backend/logging"
)

type TickerService struct {
	marketDataURL string
	client        *http.Client
	logger        *slog.Logger
	logClient     *logging.Client
}

func NewTickerService(marketDataURL string, logClient *logging.Client) *TickerService {
	if marketDataURL == "" {
		marketDataURL = "http://localhost:8081"
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	return &TickerService{
		marketDataURL: marketDataURL,
		client:        &http.Client{Timeout: 10 * time.Second},
		logger:        logger,
		logClient:     logClient,
	}
}

type TickerSearchResult struct {
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	Exchange  string  `json:"exchange"`
	Sector    string  `json:"sector"`
	Price     float64 `json:"price"`
	Change    float64 `json:"change"`
	ChangePct float64 `json:"changePct"`
}

type TickerDetails struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	Exchange      string  `json:"exchange"`
	Sector        string  `json:"sector"`
	Industry      string  `json:"industry"`
	Price         float64 `json:"price"`
	Change        float64 `json:"change"`
	ChangePct     float64 `json:"changePct"`
	Volume        int64   `json:"volume"`
	AvgVolume     int64   `json:"avgVolume"`
	MarketCap     float64 `json:"marketCap"`
	PeRatio       float64 `json:"peRatio"`
	Eps           float64 `json:"eps"`
	DividendYield float64 `json:"dividendYield"`
	Week52High    float64 `json:"week52High"`
	Week52Low     float64 `json:"week52Low"`
}

type IntradayBar struct {
	Timestamp string  `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    int64   `json:"volume"`
}

type FinancialRatio struct {
	Label       string `json:"label"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

func (s *TickerService) doGet(ctx context.Context, rawURL string, meta map[string]interface{}) ([]byte, int, http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, 0, nil, err
	}

	if s.logClient != nil {
		requestMeta := map[string]interface{}{
			"type":            "outbound_request",
			"target":          "market-data-service",
			"method":          http.MethodGet,
			"url":             rawURL,
			"request_headers": map[string][]string{},
		}
		for key, value := range meta {
			requestMeta[key] = value
		}
		s.logClient.InfoWithMeta(ctx, "Market Data Request", requestMeta)
	}

	start := time.Now()
	resp, err := s.client.Do(req)
	duration := time.Since(start)
	if err != nil {
		if s.logClient != nil {
			errorMeta := map[string]interface{}{
				"type":        "outbound_response_error",
				"target":      "market-data-service",
				"method":      http.MethodGet,
				"url":         rawURL,
				"duration_ms": duration.Milliseconds(),
				"error":       err.Error(),
			}
			for key, value := range meta {
				errorMeta[key] = value
			}
			s.logClient.ErrorWithMeta(ctx, "Market Data Request Error", errorMeta)
		}
		return nil, 0, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if s.logClient != nil {
			errorMeta := map[string]interface{}{
				"type":        "outbound_response_error",
				"target":      "market-data-service",
				"method":      http.MethodGet,
				"url":         rawURL,
				"status":      resp.StatusCode,
				"duration_ms": duration.Milliseconds(),
				"error":       err.Error(),
			}
			for key, value := range meta {
				errorMeta[key] = value
			}
			s.logClient.ErrorWithMeta(ctx, "Market Data Response Read Error", errorMeta)
		}
		return nil, resp.StatusCode, resp.Header, err
	}

	if s.logClient != nil {
		responseMeta := map[string]interface{}{
			"type":               "outbound_response",
			"target":             "market-data-service",
			"method":             http.MethodGet,
			"url":                rawURL,
			"status":             resp.StatusCode,
			"duration_ms":        duration.Milliseconds(),
			"response_headers":   map[string][]string(resp.Header),
			"response_body":      string(body),
			"response_body_size": len(body),
		}
		for key, value := range meta {
			responseMeta[key] = value
		}

		switch {
		case resp.StatusCode >= 500:
			s.logClient.ErrorWithMeta(ctx, "Market Data Response Error", responseMeta)
		case resp.StatusCode >= 400:
			s.logClient.WarnWithMeta(ctx, "Market Data Response Warning", responseMeta)
		default:
			s.logClient.InfoWithMeta(ctx, "Market Data Response", responseMeta)
		}
	}

	return body, resp.StatusCode, resp.Header, nil
}

func (s *TickerService) SearchTickers(ctx context.Context, query string) ([]TickerSearchResult, error) {
	url := fmt.Sprintf("%s/api/tickers/search?q=%s", s.marketDataURL, url.QueryEscape(query))
	s.logger.Info("Calling market-data-service", "url", url)

	body, status, _, err := s.doGet(ctx, url, map[string]interface{}{"query": query, "operation": "search_tickers"})
	if err != nil {
		s.logger.Error("SearchTickers request failed", "error", err, "url", url)
		return nil, err
	}

	s.logger.Info("SearchTickers response", "status", status, "body_size", len(body), "url", url)

	if status != http.StatusOK {
		s.logger.Error("SearchTickers failed", "status", status, "body", string(body))
		return nil, fmt.Errorf("search failed: %d - %s", status, string(body))
	}

	var results []TickerSearchResult
	if err := json.Unmarshal(body, &results); err != nil {
		s.logger.Error("SearchTickers failed to unmarshal", "error", err, "body", string(body))
		return nil, err
	}
	return results, nil
}

func (s *TickerService) GetTickerDetails(ctx context.Context, symbol string) (*TickerDetails, error) {
	url := fmt.Sprintf("%s/api/tickers/%s/details", s.marketDataURL, symbol)
	s.logger.Info("Fetching ticker details", "url", url, "symbol", symbol)

	body, status, _, err := s.doGet(ctx, url, map[string]interface{}{"symbol": symbol, "operation": "get_ticker_details"})
	if err != nil {
		s.logger.Error("GetTickerDetails request failed", "error", err, "symbol", symbol)
		return nil, err
	}

	s.logger.Info("GetTickerDetails response", "status", status, "body_size", len(body), "symbol", symbol)

	if status != http.StatusOK {
		s.logger.Error("GetTickerDetails failed", "status", status, "body", string(body), "symbol", symbol)
		return nil, fmt.Errorf("details failed: %d - %s", status, string(body))
	}

	var details TickerDetails
	if err := json.Unmarshal(body, &details); err != nil {
		s.logger.Error("GetTickerDetails failed to unmarshal", "error", err, "body", string(body))
		return nil, err
	}
	return &details, nil
}

func (s *TickerService) GetIntradayBars(ctx context.Context, symbol string, interval string) ([]IntradayBar, error) {
	param := ""
	if interval != "" && interval != "1min" {
		param = "?interval=" + interval
	}
	url := fmt.Sprintf("%s/api/tickers/%s/intraday%s", s.marketDataURL, symbol, param)
	s.logger.Info("Fetching intraday bars", "url", url, "symbol", symbol, "interval", interval)

	body, status, _, err := s.doGet(ctx, url, map[string]interface{}{"symbol": symbol, "operation": "get_intraday_bars"})
	if err != nil {
		s.logger.Error("GetIntradayBars request failed", "error", err, "symbol", symbol)
		return nil, err
	}

	s.logger.Info("GetIntradayBars response", "status", status, "body_size", len(body), "symbol", symbol)

	if status != http.StatusOK {
		s.logger.Error("GetIntradayBars failed", "status", status, "body", string(body), "symbol", symbol)
		return nil, fmt.Errorf("intraday bars failed: %d - %s", status, string(body))
	}

	var bars []IntradayBar
	if err := json.Unmarshal(body, &bars); err != nil {
		s.logger.Error("GetIntradayBars failed to unmarshal", "error", err, "body", string(body))
		return nil, err
	}
	return bars, nil
}

func (s *TickerService) GetFinancialRatios(ctx context.Context, symbol string) ([]FinancialRatio, error) {
	url := fmt.Sprintf("%s/api/tickers/%s/ratios", s.marketDataURL, symbol)
	s.logger.Info("Fetching financial ratios", "url", url, "symbol", symbol)

	body, status, _, err := s.doGet(ctx, url, map[string]interface{}{"symbol": symbol, "operation": "get_financial_ratios"})
	if err != nil {
		s.logger.Error("GetFinancialRatios request failed", "error", err, "symbol", symbol)
		return nil, err
	}

	s.logger.Info("GetFinancialRatios response", "status", status, "body_size", len(body), "symbol", symbol)

	if status != http.StatusOK {
		s.logger.Error("GetFinancialRatios failed", "status", status, "body", string(body), "symbol", symbol)
		return nil, fmt.Errorf("financial ratios failed: %d - %s", status, string(body))
	}

	var ratios []FinancialRatio
	if err := json.Unmarshal(body, &ratios); err != nil {
		s.logger.Error("GetFinancialRatios failed to unmarshal", "error", err, "body", string(body))
		return nil, err
	}
	return ratios, nil
}
