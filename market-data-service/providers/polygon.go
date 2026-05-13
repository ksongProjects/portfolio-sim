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

type MassiveProvider struct {
	cfg       config.MassiveConfig
	logClient *logging.Client
}

func NewMassiveProvider(cfg config.MassiveConfig, logClient *logging.Client) *MassiveProvider {
	return &MassiveProvider{
		cfg:       cfg,
		logClient: logClient,
	}
}

func (p *MassiveProvider) Name() string {
	return "massive"
}

func (p *MassiveProvider) GetSymbolID(ticker string) (int, error) {
	return 0, fmt.Errorf("massive does not use symbol IDs")
}

func (p *MassiveProvider) FetchCompanyProfile(ticker string) (*CompanyProfile, error) {
	return nil, fmt.Errorf("massive does not support company profile")
}

func (p *MassiveProvider) FetchFinancialRatios(ticker string) ([]*FinancialRatio, error) {
	return nil, fmt.Errorf("massive does not support financial ratios")
}

func (p *MassiveProvider) FetchPrice(ticker string) (*Price, error) {
	apiKey := p.cfg.APIKey
	if apiKey == "" {
		return nil, fmt.Errorf("massive API key not configured")
	}

	url := fmt.Sprintf("https://api.massive.io/v2/aggs/ticker/%s/prev?adjusted=true&apiKey=%s", ticker, apiKey)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("massive request failed: %d", resp.StatusCode)
	}

	var result struct {
		Results []struct {
			Open      float64 `json:"o"`
			High      float64 `json:"h"`
			Low       float64 `json:"l"`
			Close     float64 `json:"c"`
			Volume    float64 `json:"v"`
			Timestamp int64   `json:"t"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if len(result.Results) == 0 {
		return nil, fmt.Errorf("no price data for %s", ticker)
	}

	r := result.Results[0]
	change := r.Close - r.Open
	var changePct float64
	if r.Open != 0 {
		changePct = (change / r.Open) * 100
	}
	return &Price{
		Ticker:    ticker,
		Price:     r.Close,
		Change:    change,
		ChangePct: changePct,
		Volume:    int64(r.Volume),
		Source:    p.Name(),
		Timestamp: time.UnixMilli(r.Timestamp),
	}, nil
}

func (p *MassiveProvider) FetchOptionChain(ticker string) ([]*OptionChain, error) {
	return nil, fmt.Errorf("massive does not support option chains via this endpoint")
}

func (p *MassiveProvider) FetchIntradayBars(ticker string, interval string) ([]*IntradayBar, error) {
	apiKey := p.cfg.APIKey
	if apiKey == "" {
		return nil, fmt.Errorf("massive API key not configured")
	}

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

	url := fmt.Sprintf("https://api.massive.io/v2/aggs/ticker/%s/range/%s/%s/%s/%s?adjusted=true&apiKey=%s",
		ticker, multiplier, timespan, from, to, apiKey)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("massive request failed: %d", resp.StatusCode)
	}

	var result struct {
		Results []struct {
			Open      float64 `json:"o"`
			High      float64 `json:"h"`
			Low       float64 `json:"l"`
			Close     float64 `json:"c"`
			Volume    float64 `json:"v"`
			Timestamp int64   `json:"t"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	bars := make([]*IntradayBar, 0, len(result.Results))
	for _, b := range result.Results {
		bars = append(bars, &IntradayBar{
			Ticker:    ticker,
			Interval:  interval,
			Open:      b.Open,
			High:      b.High,
			Low:      b.Low,
			Close:     b.Close,
			Volume:    int64(b.Volume),
			Timestamp: time.UnixMilli(b.Timestamp),
		})
	}
	return bars, nil
}

func (p *MassiveProvider) SearchTickers(prefix string) ([]TickerSearchResult, error) {
	apiKey := p.cfg.APIKey
	if apiKey == "" {
		return nil, fmt.Errorf("massive API key not configured")
	}

	searchURL := fmt.Sprintf("https://api.massive.io/v3/reference/tickers?search=%s&active=true&apiKey=%s", url.QueryEscape(prefix), apiKey)

	req, err := http.NewRequest(http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("massive search failed: %d", resp.StatusCode)
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