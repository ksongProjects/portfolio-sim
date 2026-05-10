package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type TickerService struct {
	marketDataURL string
	client        *http.Client
	logger        *slog.Logger
}

func NewTickerService(marketDataURL string) *TickerService {
	if marketDataURL == "" {
		marketDataURL = "http://localhost:8081"
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	return &TickerService{
		marketDataURL: marketDataURL,
		client:        &http.Client{Timeout: 10 * time.Second},
		logger:        logger,
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

func (s *TickerService) SearchTickers(ctx context.Context, query string) ([]TickerSearchResult, error) {
	url := fmt.Sprintf("%s/api/tickers/search?q=%s", s.marketDataURL, query)
	s.logger.Info("Calling market-data-service", "url", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("SearchTickers request failed", "error", err, "url", url)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("SearchTickers failed to read body", "error", err)
		return nil, err
	}

	s.logger.Info("SearchTickers response", "status", resp.StatusCode, "body_size", len(body), "url", url)

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("SearchTickers failed", "status", resp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("search failed: %d - %s", resp.StatusCode, string(body))
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("GetTickerDetails request failed", "error", err, "symbol", symbol)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("GetTickerDetails failed to read body", "error", err, "symbol", symbol)
		return nil, err
	}

	s.logger.Info("GetTickerDetails response", "status", resp.StatusCode, "body_size", len(body), "symbol", symbol)

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("GetTickerDetails failed", "status", resp.StatusCode, "body", string(body), "symbol", symbol)
		return nil, fmt.Errorf("details failed: %d - %s", resp.StatusCode, string(body))
	}

	var details TickerDetails
	if err := json.Unmarshal(body, &details); err != nil {
		s.logger.Error("GetTickerDetails failed to unmarshal", "error", err, "body", string(body))
		return nil, err
	}
	return &details, nil
}

func (s *TickerService) GetIntradayBars(ctx context.Context, symbol string) ([]IntradayBar, error) {
	url := fmt.Sprintf("%s/api/tickers/%s/intraday", s.marketDataURL, symbol)
	s.logger.Info("Fetching intraday bars", "url", url, "symbol", symbol)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("GetIntradayBars request failed", "error", err, "symbol", symbol)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("GetIntradayBars failed to read body", "error", err)
		return nil, err
	}

	s.logger.Info("GetIntradayBars response", "status", resp.StatusCode, "body_size", len(body), "symbol", symbol)

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("GetIntradayBars failed", "status", resp.StatusCode, "body", string(body), "symbol", symbol)
		return nil, fmt.Errorf("intraday bars failed: %d - %s", resp.StatusCode, string(body))
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("GetFinancialRatios request failed", "error", err, "symbol", symbol)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("GetFinancialRatios failed to read body", "error", err)
		return nil, err
	}

	s.logger.Info("GetFinancialRatios response", "status", resp.StatusCode, "body_size", len(body), "symbol", symbol)

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("GetFinancialRatios failed", "status", resp.StatusCode, "body", string(body), "symbol", symbol)
		return nil, fmt.Errorf("financial ratios failed: %d - %s", resp.StatusCode, string(body))
	}

	var ratios []FinancialRatio
	if err := json.Unmarshal(body, &ratios); err != nil {
		s.logger.Error("GetFinancialRatios failed to unmarshal", "error", err, "body", string(body))
		return nil, err
	}
	return ratios, nil
}