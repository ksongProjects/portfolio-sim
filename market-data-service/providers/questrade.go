package providers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/portfolio-sim/market-data-service/config"
)

type QuestradeProvider struct {
	cfg     config.QuestradeConfig
	client  *http.Client
	baseURL string
	token   string
}

func NewQuestradeProvider(cfg config.QuestradeConfig) *QuestradeProvider {
	return &QuestradeProvider{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *QuestradeProvider) Name() string {
	return "questrade"
}

func (p *QuestradeProvider) refreshToken() error {
	if p.cfg.RefreshToken == "" {
		return nil
	}

	values := url.Values{}
	values.Set("refresh_token", p.cfg.RefreshToken)
	values.Set("grant_type", "refresh_token")

	resp, err := p.client.PostForm(p.baseURL+"/oauth2/token", values)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("questrade token refresh failed: %d", resp.StatusCode)
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	p.token = result.AccessToken
	p.cfg.RefreshToken = result.RefreshToken
	return nil
}

func (p *QuestradeProvider) doRequest(endpoint string) ([]byte, error) {
	if p.token == "" {
		if err := p.refreshToken(); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest("GET", p.baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.token)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		if err := p.refreshToken(); err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+p.token)
		resp, err = p.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("questrade request failed: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (p *QuestradeProvider) FetchPrice(ticker string) (*Price, error) {
	body, err := p.doRequest("/v1/markets/quotes/" + ticker)
	if err != nil {
		return nil, err
	}

	var result struct {
		Quotes []struct {
			Symbol     string  `json:"symbol"`
			LastPrice  float64 `json:"lastPrice"`
			Bid        float64 `json:"bidPrice"`
			Ask        float64 `json:"askPrice"`
			Volume     int64   `json:"volume"`
			Timestamp  int64   `json:"timestamp"`
		} `json:"quotes"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if len(result.Quotes) == 0 {
		return nil, fmt.Errorf("no quote for %s", ticker)
	}

	q := result.Quotes[0]
	return &Price{
		Ticker:   q.Symbol,
		Price:    q.LastPrice,
		Bid:      q.Bid,
		Ask:      q.Ask,
		Volume:   q.Volume,
		Source:   p.Name(),
		Timestamp: time.UnixMilli(q.Timestamp),
	}, nil
}

func (p *QuestradeProvider) FetchOptionChain(ticker string) ([]*OptionChain, error) {
	body, err := p.doRequest("/v1/markets/options/" + ticker)
	if err != nil {
		return nil, err
	}

	var result struct {
		Options []struct {
			Symbol         string  `json:"symbol"`
			ExpirationDate string  `json:"expirationDate"`
			Strike         float64 `json:"strikePrice"`
			OptionType     string  `json:"optionType"`
			Bid            float64 `json:"bidPrice"`
			Ask            float64 `json:"askPrice"`
			Delta          float64 `json:"delta"`
			Gamma          float64 `json:"gamma"`
			Theta          float64 `json:"theta"`
			Vega           float64 `json:"vega"`
			Volume         int64   `json:"volume"`
			OpenInterest   int64   `json:"openInterest"`
		} `json:"options"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	chains := make([]*OptionChain, len(result.Options))
	for i, o := range result.Options {
		exp, _ := time.Parse("2006-01-02", o.ExpirationDate)
		chains[i] = &OptionChain{
			Ticker:       o.Symbol,
			Expiration:   exp,
			Strike:       o.Strike,
			OptionType:   o.OptionType,
			Bid:          o.Bid,
			Ask:          o.Ask,
			Delta:        o.Delta,
			Gamma:        o.Gamma,
			Theta:        o.Theta,
			Vega:         o.Vega,
			Volume:       o.Volume,
			OpenInterest: o.OpenInterest,
			Source:       p.Name(),
			Timestamp:    time.Now(),
		}
	}
	return chains, nil
}

func (p *QuestradeProvider) FetchIntradayBars(ticker string, interval string) ([]*IntradayBar, error) {
	body, err := p.doRequest("/v1/markets/bars/" + ticker + "?interval=" + interval)
	if err != nil {
		return nil, err
	}

	var result struct {
		Bars []struct {
			StartTime  int64   `json:"startTime"`
			Open       float64 `json:"open"`
			High       float64 `json:"high"`
			Low        float64 `json:"low"`
			Close      float64 `json:"close"`
			Volume     int64   `json:"volume"`
		} `json:"bars"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	bars := make([]*IntradayBar, len(result.Bars))
	for i, b := range result.Bars {
		bars[i] = &IntradayBar{
			Ticker:    ticker,
			Interval:  interval,
			Open:      b.Open,
			High:      b.High,
			Low:       b.Low,
			Close:     b.Close,
			Volume:    b.Volume,
			Timestamp: time.UnixMilli(b.StartTime),
		}
	}
	return bars, nil
}
