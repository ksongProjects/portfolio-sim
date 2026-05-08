package providers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/portfolio-sim/market-data-service/config"
)

type PolygonProvider struct {
	cfg    config.PolygonConfig
	client *http.Client
}

func NewPolygonProvider(cfg config.PolygonConfig) *PolygonProvider {
	return &PolygonProvider{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *PolygonProvider) Name() string {
	return "polygon"
}

func (p *PolygonProvider) FetchPrice(ticker string) (*Price, error) {
	url := fmt.Sprintf("https://api.polygon.io/v2/aggs/ticker/%s/prev?adjusted=true&apiKey=%s", ticker, p.cfg.APIKey)
	resp, err := p.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("polygon request failed: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Results []struct {
			Ticker   string  `json:"T"`
			Open     float64 `json:"o"`
			High     float64 `json:"h"`
			Low      float64 `json:"l"`
			Close    float64 `json:"c"`
			Volume   int64   `json:"v"`
			Timestamp int64  `json:"t"`
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
		Ticker:   ticker,
		Price:    r.Close,
		Volume:   r.Volume,
		Source:   p.Name(),
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

	resp, err := p.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("polygon request failed: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
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
