package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type TickerService struct {
	marketDataURL string
	client        *http.Client
}

func NewTickerService(marketDataURL string) *TickerService {
	if marketDataURL == "" {
		marketDataURL = "http://localhost:8081"
	}
	return &TickerService{
		marketDataURL: marketDataURL,
		client:        &http.Client{Timeout: 10 * time.Second},
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed: %d", resp.StatusCode)
	}

	var results []TickerSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}
	return results, nil
}

func (s *TickerService) GetTickerDetails(ctx context.Context, symbol string) (*TickerDetails, error) {
	url := fmt.Sprintf("%s/api/tickers/%s/details", s.marketDataURL, symbol)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("details failed: %d", resp.StatusCode)
	}

	var details TickerDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, err
	}
	return &details, nil
}

func (s *TickerService) GetIntradayBars(ctx context.Context, symbol string) ([]IntradayBar, error) {
	url := fmt.Sprintf("%s/api/tickers/%s/intraday", s.marketDataURL, symbol)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("intraday bars failed: %d", resp.StatusCode)
	}

	var bars []IntradayBar
	if err := json.NewDecoder(resp.Body).Decode(&bars); err != nil {
		return nil, err
	}
	return bars, nil
}

func (s *TickerService) GetFinancialRatios(ctx context.Context, symbol string) ([]FinancialRatio, error) {
	url := fmt.Sprintf("%s/api/tickers/%s/ratios", s.marketDataURL, symbol)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("financial ratios failed: %d", resp.StatusCode)
	}

	var ratios []FinancialRatio
	if err := json.NewDecoder(resp.Body).Decode(&ratios); err != nil {
		return nil, err
	}
	return ratios, nil
}