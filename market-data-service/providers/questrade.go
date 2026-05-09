package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/portfolio-sim/market-data-service/config"
	"github.com/portfolio-sim/market-data-service/storage"
)

type QuestradeProvider struct {
	cfg         config.QuestradeConfig
	client      *http.Client
	baseURL     string
	token       string
	rateLimiter *RateLimiter
	storage     *storage.Storage
}

func NewQuestradeProvider(cfg config.QuestradeConfig, storage *storage.Storage) *QuestradeProvider {
	return &QuestradeProvider{
		cfg:         cfg,
		client:      &http.Client{Timeout: 30 * time.Second},
		baseURL:     cfg.APIURL,
		rateLimiter: NewRateLimiter(20, 15000),
		storage:     storage,
	}
}

func (p *QuestradeProvider) Name() string {
	return "questrade"
}

func (p *QuestradeProvider) refreshToken() error {
	if p.storage == nil {
		return fmt.Errorf("no storage configured")
	}

	tokens, err := p.storage.GetQuestradeTokens(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get tokens from storage: %w", err)
	}
	if tokens.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	tokenURL := fmt.Sprintf("%s?grant_type=refresh_token&refresh_token=%s",
		tokens.APIServer, url.QueryEscape(tokens.RefreshToken))
	resp, err := http.Get(tokenURL)
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
		APIServer    string `json:"api_server"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	p.token = result.AccessToken
	if result.APIServer != "" {
		p.baseURL = result.APIServer
	}

	_ = p.storage.UpdateQuestradeTokens(context.Background(), result.AccessToken, result.RefreshToken, result.APIServer, result.ExpiresIn)

	return nil
}

func (p *QuestradeProvider) doRequest(endpoint string) ([]byte, error) {
	p.rateLimiter.Wait()

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

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("questrade rate limit exceeded")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("questrade request failed: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

type QuestradeQuote struct {
	Symbol            string  `json:"symbol"`
	SymbolID          int     `json:"symbolId"`
	Tier              string  `json:"tier"`
	BidPrice          float64 `json:"bidPrice"`
	BidSize           int     `json:"bidSize"`
	AskPrice          float64 `json:"askPrice"`
	AskSize           int     `json:"askSize"`
	LastTradePrice    float64 `json:"lastTradePrice"`
	LastTradeSize     int     `json:"lastTradeSize"`
	LastTradeTick     string  `json:"lastTradeTick"`
	Volume            int64   `json:"volume"`
	OpenPrice         float64 `json:"openPrice"`
	HighPrice         float64 `json:"highPrice"`
	LowPrice          float64 `json:"lowPrice"`
	Delay             bool    `json:"delay"`
	IsHalted          bool    `json:"isHalted"`
}

type QuestradeCandle struct {
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
	Open   float64   `json:"open"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Close  float64   `json:"close"`
	Volume int64     `json:"volume"`
}

type QuestradeSymbol struct {
	Symbol     string `json:"symbol"`
	SymbolID   int    `json:"symbolId"`
	BoardID    string `json:"boardId"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Currency   string `json:"currency"`
	Decimals   int    `json:"decimals"`
	Active     bool   `json:"isActive"`
}

type QuestradeOption struct {
	Symbol            string  `json:"symbol"`
	SymbolID          int     `json:"symbolId"`
	ExpirationDate    string  `json:"expirationDate"`
	Strike            float64 `json:"strike"`
	CallPut           string  `json:"callPut"`
	BidPrice          float64 `json:"bidPrice"`
	AskPrice          float64 `json:"askPrice"`
	LastTradePrice    float64 `json:"lastTradePrice"`
	Volume            int     `json:"volume"`
	OpenInterest      int     `json:"openInterest"`
	Description       string  `json:"description"`
	IntrinsicValue    float64 `json:"intrinsicValue"`
	OptionOpenInterest int    `json:"optionOpenInterest"`
	Delta             float64 `json:"delta"`
	Gamma             float64 `json:"gamma"`
	Theta             float64 `json:"theta"`
	Vega              float64 `json:"vega"`
}

func (p *QuestradeProvider) FetchMarkets() ([]string, error) {
	body, err := p.doRequest("/v1/markets")
	if err != nil {
		return nil, err
	}

	var result struct {
		Markets []string `json:"markets"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.Markets, nil
}

func (p *QuestradeProvider) FetchQuote(symbolID string) (*QuestradeQuote, error) {
	body, err := p.doRequest("/v1/markets/quotes/" + symbolID)
	if err != nil {
		return nil, err
	}

	var result struct {
		Quotes []QuestradeQuote `json:"quotes"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if len(result.Quotes) == 0 {
		return nil, fmt.Errorf("no quote found for symbol ID %s", symbolID)
	}
	return &result.Quotes[0], nil
}

func (p *QuestradeProvider) FetchQuotesByIDs(ids []string) ([]QuestradeQuote, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("no symbol IDs provided")
	}
	body, err := p.doRequest("/v1/markets/quotes?ids=" + url.QueryEscape(ids[0]))
	if err != nil {
		return nil, err
	}

	var result struct {
		Quotes []QuestradeQuote `json:"quotes"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.Quotes, nil
}

func (p *QuestradeProvider) FetchCandles(symbolID string, startTime, endTime time.Time, interval string) ([]QuestradeCandle, error) {
	endpoint := fmt.Sprintf("/v1/markets/candles/%s?startTime=%s&endTime=%s&interval=%s",
		symbolID,
		url.QueryEscape(startTime.Format(time.RFC3339)),
		url.QueryEscape(endTime.Format(time.RFC3339)),
		url.QueryEscape(interval))

	body, err := p.doRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var result struct {
		Candles []struct {
			Start  string  `json:"start"`
			End    string  `json:"end"`
			Open   float64 `json:"open"`
			High   float64 `json:"high"`
			Low    float64 `json:"low"`
			Close  float64 `json:"close"`
			Volume int64   `json:"volume"`
		} `json:"candles"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	candles := make([]QuestradeCandle, len(result.Candles))
	for i, c := range result.Candles {
		start, _ := time.Parse(time.RFC3339, c.Start)
		end, _ := time.Parse(time.RFC3339, c.End)
		candles[i] = QuestradeCandle{
			Start:  start,
			End:    end,
			Open:   c.Open,
			High:   c.High,
			Low:    c.Low,
			Close:  c.Close,
			Volume: c.Volume,
		}
	}
	return candles, nil
}

func (p *QuestradeProvider) FetchSymbol(symbolID string) (*QuestradeSymbol, error) {
	body, err := p.doRequest("/v1/symbols/" + symbolID)
	if err != nil {
		return nil, err
	}

	var result struct {
		Symbols []QuestradeSymbol `json:"symbols"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if len(result.Symbols) == 0 {
		return nil, fmt.Errorf("symbol not found: %s", symbolID)
	}
	return &result.Symbols[0], nil
}

func (p *QuestradeProvider) FetchSymbolSearch(prefix string) ([]QuestradeSymbol, error) {
	body, err := p.doRequest("/v1/symbols/search?prefix=" + url.QueryEscape(prefix))
	if err != nil {
		return nil, err
	}

	var result struct {
		Symbols []QuestradeSymbol `json:"symbols"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.Symbols, nil
}

func (p *QuestradeProvider) FetchSymbolOptions(symbolID string) ([]QuestradeOption, error) {
	body, err := p.doRequest("/v1/symbols/" + symbolID + "/options")
	if err != nil {
	 return nil, err
	}

	var result struct {
		Options []QuestradeOption `json:"optionSymbols"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.Options, nil
}

func (p *QuestradeProvider) FetchQuoteStrategies(strategy string) error {
	body, err := p.doRequest("/v1/markets/quotes/strategies?strategy=" + url.QueryEscape(strategy))
	if err != nil {
		return err
	}
	fmt.Println(string(body))
	return nil
}

func (p *QuestradeProvider) FetchQuoteOptions(symbolIDs []string) error {
	if len(symbolIDs) == 0 {
		return fmt.Errorf("no symbol IDs provided")
	}
	body, err := p.doRequest("/v1/markets/quotes/options?ids=" + url.QueryEscape(symbolIDs[0]))
	if err != nil {
		return err
	}
	fmt.Println(string(body))
	return nil
}

func (p *QuestradeProvider) FetchPrice(ticker string) (*Price, error) {
	body, err := p.doRequest("/v1/markets/quotes/" + ticker)
	if err != nil {
		return nil, err
	}

	var result struct {
		Quotes []struct {
			Symbol     string  `json:"symbol"`
			LastPrice  float64 `json:"lastTradePrice"`
			Bid        float64 `json:"bidPrice"`
			Ask        float64 `json:"askPrice"`
			Volume     int64   `json:"volume"`
			Timestamp  int64   `json:"lastTradeTime"`
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
		Ticker:    q.Symbol,
		Price:     q.LastPrice,
		Bid:       q.Bid,
		Ask:       q.Ask,
		Volume:    q.Volume,
		Source:    p.Name(),
		Timestamp: time.UnixMilli(q.Timestamp),
	}, nil
}

func (p *QuestradeProvider) FetchOptionChain(ticker string) ([]*OptionChain, error) {
	body, err := p.doRequest("/v1/symbols/" + ticker + "/options")
	if err != nil {
		return nil, err
	}

	var result struct {
		Options []struct {
			Symbol         string  `json:"symbol"`
			ExpirationDate string  `json:"expirationDate"`
			Strike         float64 `json:"strike"`
			CallPut        string  `json:"callPut"`
			Bid            float64 `json:"bidPrice"`
			Ask            float64 `json:"askPrice"`
			LastTradePrice float64 `json:"lastTradePrice"`
			Volume         int64   `json:"volume"`
			OpenInterest   int64   `json:"openInterest"`
			Delta          float64 `json:"delta"`
			Gamma          float64 `json:"gamma"`
			Theta          float64 `json:"theta"`
			Vega           float64 `json:"vega"`
		} `json:"optionSymbols"`
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
			OptionType:   o.CallPut,
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
	candles, err := p.FetchCandles(ticker, time.Now().Add(-24*time.Hour), time.Now(), interval)
	if err != nil {
		return nil, err
	}

	bars := make([]*IntradayBar, len(candles))
	for i, c := range candles {
		bars[i] = &IntradayBar{
			Ticker:    ticker,
			Interval:  interval,
			Open:      c.Open,
			High:      c.High,
			Low:       c.Low,
			Close:     c.Close,
			Volume:    c.Volume,
			Timestamp: c.Start,
		}
	}
	return bars, nil
}

func (p *QuestradeProvider) GetSymbolID(ticker string) (int, error) {
	body, err := p.doRequest("/v1/symbols/search?prefix=" + url.QueryEscape(ticker))
	if err != nil {
		return 0, err
	}

	var result struct {
		Symbols []struct {
			SymbolID int    `json:"symbolId"`
			Symbol   string `json:"symbol"`
		} `json:"symbols"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}

	for _, s := range result.Symbols {
		if s.Symbol == ticker {
			return s.SymbolID, nil
		}
	}
	if len(result.Symbols) > 0 {
		return result.Symbols[0].SymbolID, nil
	}
	return 0, fmt.Errorf("symbol not found: %s", ticker)
}

func (p *QuestradeProvider) GetRateLimitStatus() (remaining int, reset int64, err error) {
	secCount := p.rateLimiter.getSecCount()
	hourCount := p.rateLimiter.getHourCount()
	remainingSec := 20 - secCount
	remainingHour := 15000 - hourCount
	if remainingSec < remainingHour {
		remaining = remainingSec
	} else {
		remaining = remainingHour
	}
	reset = p.rateLimiter.getResetTime().Unix()
	return remaining, reset, nil
}

func (r *RateLimiter) getSecCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.secCount
}

func (r *RateLimiter) getHourCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.hourCount
}

func (r *RateLimiter) getResetTime() time.Time {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.hourReset.Before(r.secReset) {
		return r.hourReset
	}
	return r.secReset
}

func ParseQuestradeTimestamp(ts string) (time.Time, error) {
	if ts == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse("2006-01-02T15:04:05.000000-07:00", ts)
	if err != nil {
		return time.Parse(time.RFC3339, ts)
	}
	return t, err
}

func (p *QuestradeProvider) FetchCandlesBySymbol(ticker string, startTime, endTime time.Time, interval string) ([]QuestradeCandle, error) {
	symbolID, err := p.GetSymbolID(ticker)
	if err != nil {
		return nil, err
	}
	return p.FetchCandles(strconv.Itoa(symbolID), startTime, endTime, interval)
}