package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/portfolio-sim/market-data-service/config"
	"github.com/portfolio-sim/market-data-service/logging"
	"github.com/portfolio-sim/market-data-service/storage"
)

const questradeOAuthURL = "https://login.questrade.com/oauth2/token"

var questradeRefreshMu sync.Mutex

type questradeTokenStore interface {
	GetQuestradeTokens(ctx context.Context) (*storage.QuestradeTokens, error)
	UpdateQuestradeTokens(ctx context.Context, accessToken, refreshToken, apiServer string, expiresIn int) error
}

type QuestradeProvider struct {
	cfg           config.QuestradeConfig
	client        *http.Client
	oauthURL      string
	rateLimiter   *RateLimiter
	logClient     *logging.Client
	backendURL    string
	internalToken string
	token         string
	baseURL       string
	tokenExpiresAt time.Time
	symbolIDMu    sync.RWMutex
	symbolIDCache map[string]int
}

func NewQuestradeProvider(cfg config.QuestradeConfig, backendURL, internalToken string, logClient *logging.Client) *QuestradeProvider {
	return &QuestradeProvider{
		cfg:           cfg,
		client:        &http.Client{Timeout: 30 * time.Second},
		oauthURL:      questradeOAuthURL,
		rateLimiter:   NewRateLimiter(20, 15000),
		logClient:     logClient,
		backendURL:    backendURL,
		internalToken: internalToken,
		symbolIDCache: map[string]int{},
	}
}

func (p *QuestradeProvider) Name() string {
	return "questrade"
}

func (p *QuestradeProvider) logRequest(method, url string, headers http.Header) {
	if p.logClient == nil {
		return
	}
	url = redactQuestradeURL(url)
	p.logClient.InfoWithMeta(nil, "Questrade Request: "+method+" "+url, map[string]interface{}{
		"method":          method,
		"url":             url,
		"provider":        "questrade",
		"base_url":        p.baseURL,
		"has_token":       p.token != "",
		"request_headers": logging.RedactHeaders(headers),
	})
}

func (p *QuestradeProvider) logResponse(method, url string, status int, headers http.Header, body []byte, err error) {
	if p.logClient == nil {
		return
	}
	url = redactQuestradeURL(url)
	body = redactQuestradeBody(url, body)
	level := "INFO"
	if status >= 400 || err != nil {
		level = "ERROR"
	}
	msg := fmt.Sprintf("Questrade %s %s → %d (%d bytes)", method, url, status, len(body))
	if err != nil {
		msg = fmt.Sprintf("Questrade %s %s → ERROR: %v", method, url, err)
	}
	meta := map[string]interface{}{
		"method":           method,
		"url":              url,
		"status":           status,
		"body_length":      len(body),
		"provider":         "questrade",
		"base_url":         p.baseURL,
		"has_token":        p.token != "",
		"response_headers": logging.RedactHeaders(headers),
	}
	if len(body) > 0 {
		meta["body"] = logging.SanitizeBody(headers.Get("Content-Type"), body)
	}
	if err != nil {
		meta["error"] = err.Error()
		meta["error_type"] = fmt.Sprintf("%T", err)
	}
	p.logClient.EmitWithMetadata(nil, level, msg, meta)
}

func (p *QuestradeProvider) logError(method, url string, err error) {
	if p.logClient == nil {
		return
	}
	url = redactQuestradeURL(url)
	p.logClient.EmitWithMetadata(nil, "ERROR", "Questrade Error: "+method+" "+url, map[string]interface{}{
		"method":      method,
		"url":         url,
		"error":       err.Error(),
		"error_type":  fmt.Sprintf("%T", err),
		"provider":    "questrade",
		"base_url":    p.baseURL,
		"has_token":   p.token != "",
		"status":      0,
		"body_length": 0,
	})
}

func redactQuestradeURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	query := parsed.Query()
	if _, ok := query["refresh_token"]; ok {
		query.Set("refresh_token", "REDACTED")
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func redactQuestradeBody(rawURL string, body []byte) []byte {
	if !strings.Contains(rawURL, "/oauth2/token") || len(body) == 0 {
		return body
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return body
	}

	if _, ok := payload["access_token"]; ok {
		payload["access_token"] = "REDACTED"
	}
	if _, ok := payload["refresh_token"]; ok {
		payload["refresh_token"] = "REDACTED"
	}

	redacted, err := json.Marshal(payload)
	if err != nil {
		return body
	}
	return redacted
}

func (p *QuestradeProvider) refreshToken(force bool) error {
	if p.backendURL == "" || p.internalToken == "" {
		p.logClient.ErrorWithMeta(nil, "Questrade token refresh skipped: backend not configured", map[string]interface{}{
			"backend_url_empty":     p.backendURL == "",
			"internal_token_empty": p.internalToken == "",
			"provider":              "questrade",
		})
		errMsg := "questrade: backend not configured for token refresh"
		if p.backendURL == "" {
			errMsg += " (BACKEND_URL is empty)"
		}
		if p.internalToken == "" {
			errMsg += " (INTERNAL_API_TOKEN not set)"
		}
		return fmt.Errorf("%s", errMsg)
	}

	questradeRefreshMu.Lock()
	defer questradeRefreshMu.Unlock()

	tokens, err := p.getTokensFromBackend()
	if err != nil {
		return fmt.Errorf("questrade: failed to get tokens from backend: %w", err)
	}

	if !force && tokens.AccessToken != "" && tokens.APIServer != "" && time.Now().Before(tokens.ExpiresAt) {
		p.token = tokens.AccessToken
		p.baseURL = tokens.APIServer
		p.tokenExpiresAt = tokens.ExpiresAt
		return nil
	}

	if tokens.RefreshToken == "" {
		return fmt.Errorf("no refresh token available from backend")
	}

	tokenURL := fmt.Sprintf("%s?grant_type=refresh_token&refresh_token=%s",
		p.oauthURL,
		url.QueryEscape(tokens.RefreshToken))
	req, err := http.NewRequest(http.MethodGet, tokenURL, nil)
	if err != nil {
		p.logError("GET", tokenURL, err)
		return err
	}
	req.Header.Set("Accept", "application/json")
	p.logRequest("GET", tokenURL, req.Header)
	resp, err := p.client.Do(req)
	if err != nil {
		p.logError("GET", tokenURL, err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		p.logError("GET", tokenURL, err)
		return err
	}
	p.logResponse("GET", tokenURL, resp.StatusCode, resp.Header, body, nil)

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("questrade token refresh failed: %d, body: %s", resp.StatusCode, string(body))
		p.logResponse("GET", tokenURL, resp.StatusCode, resp.Header, body, err)
		return err
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		APIServer    string `json:"api_server"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}
	if result.AccessToken == "" {
		return fmt.Errorf("questrade token refresh missing access token")
	}
	if result.RefreshToken == "" {
		result.RefreshToken = tokens.RefreshToken
	}
	if result.APIServer == "" {
		result.APIServer = tokens.APIServer
	}
	if result.APIServer == "" {
		return fmt.Errorf("questrade token refresh missing api_server")
	}

	p.token = result.AccessToken
	p.baseURL = result.APIServer
	p.tokenExpiresAt = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)

	return nil
}

type backendTokens struct {
	AccessToken  string
	RefreshToken string
	APIServer    string
	ExpiresAt    time.Time
}

func (p *QuestradeProvider) getTokensFromBackend() (*backendTokens, error) {
	url := p.backendURL + "/internal/providers/questrade/tokens"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Internal-API-Token", p.internalToken)
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("backend request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("questrade tokens not configured in backend")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("backend returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		APIServer    string `json:"api_server"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &backendTokens{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		APIServer:    result.APIServer,
		ExpiresAt:    time.Now().Add(25 * time.Minute),
	}, nil
}

func (p *QuestradeProvider) doRequest(endpoint string) ([]byte, error) {
	p.rateLimiter.Wait()

	reqURL := strings.TrimSuffix(p.baseURL, "/") + endpoint
	if p.token == "" || p.baseURL == "" || (!p.tokenExpiresAt.IsZero() && !time.Now().Before(p.tokenExpiresAt)) {
		if err := p.refreshToken(false); err != nil {
			p.logError("GET", reqURL, err)
			return nil, err
		}
	}

	reqURL = strings.TrimSuffix(p.baseURL, "/") + endpoint
	if p.logClient != nil {
		p.logClient.InfoWithMeta(nil, "Questrade API Request", map[string]interface{}{
			"method":       "GET",
			"url":          reqURL,
			"base_url":     p.baseURL,
			"has_token":    p.token != "",
			"token_length": len(p.token),
			"provider":     "questrade",
		})
	}
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		p.logError("GET", reqURL, err)
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Accept", "application/json")
	p.logRequest("GET", reqURL, req.Header)
	if p.logClient != nil {
		p.logClient.InfoWithMeta(nil, "Questrade API Request Details", map[string]interface{}{
			"method":          "GET",
			"url":             redactQuestradeURL(reqURL),
			"has_auth":        p.token != "",
			"token_length":    len(p.token),
			"provider":        "questrade",
			"request_headers": logging.RedactHeaders(req.Header),
		})
	}

resp, err := p.client.Do(req)
	if err != nil {
		p.logError("GET", reqURL, err)
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	p.logResponse("GET", reqURL, resp.StatusCode, resp.Header, body, nil)

	if resp.StatusCode == http.StatusUnauthorized {
		if err := p.refreshToken(true); err != nil {
			p.logError("GET", reqURL, err)
			return nil, err
		}
		reqURL = strings.TrimSuffix(p.baseURL, "/") + endpoint
		req, err = http.NewRequest("GET", reqURL, nil)
		if err != nil {
			p.logError("GET", reqURL, err)
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+p.token)
		req.Header.Set("Accept", "application/json")
		p.logRequest("GET", reqURL, req.Header)
		resp, err = p.client.Do(req)
		if err != nil {
			p.logError("GET", reqURL, err)
			return nil, err
		}
		defer resp.Body.Close()
		body, _ = io.ReadAll(resp.Body)
		p.logResponse("GET", reqURL, resp.StatusCode, resp.Header, body, nil)

		if resp.StatusCode == http.StatusUnauthorized {
			p.logError("GET", reqURL, fmt.Errorf("questrade token refresh succeeded but retry still unauthorized"))
			return nil, fmt.Errorf("questrade token refresh failed: retry returned 401")
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			err := fmt.Errorf("questrade rate limit exceeded")
			p.logResponse("GET", reqURL, resp.StatusCode, resp.Header, body, err)
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			bodyPreview := string(body)
			if len(bodyPreview) > 200 {
				bodyPreview = bodyPreview[:200] + "..."
			}
			if p.logClient != nil {
				p.logClient.ErrorWithMeta(nil, "Questrade API Error Response", map[string]interface{}{
					"method":           "GET",
					"url":              redactQuestradeURL(reqURL),
					"status":           resp.StatusCode,
					"body":             logging.SanitizeBody(resp.Header.Get("Content-Type"), []byte(bodyPreview)),
					"body_length":      len(body),
					"provider":         "questrade",
					"response_headers": logging.RedactHeaders(resp.Header),
				})
			}
			err := fmt.Errorf("questrade request failed: %d, body: %s", resp.StatusCode, bodyPreview)
			p.logResponse("GET", reqURL, resp.StatusCode, resp.Header, body, err)
			return nil, err
		}

		return body, nil
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		err := fmt.Errorf("questrade rate limit exceeded")
		p.logResponse("GET", reqURL, resp.StatusCode, resp.Header, body, err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		bodyPreview := string(body)
		if len(bodyPreview) > 200 {
			bodyPreview = bodyPreview[:200] + "..."
		}
		if p.logClient != nil {
			p.logClient.ErrorWithMeta(nil, "Questrade API Error Response", map[string]interface{}{
				"method":           "GET",
				"url":              redactQuestradeURL(reqURL),
				"status":           resp.StatusCode,
				"body":             logging.SanitizeBody(resp.Header.Get("Content-Type"), []byte(bodyPreview)),
				"body_length":      len(body),
				"provider":         "questrade",
				"response_headers": logging.RedactHeaders(resp.Header),
			})
		}
		err := fmt.Errorf("questrade request failed: %d, body: %s", resp.StatusCode, bodyPreview)
		p.logResponse("GET", reqURL, resp.StatusCode, resp.Header, body, err)
		return nil, err
	}

	return body, nil
}

type QuestradeQuote struct {
	Symbol         string  `json:"symbol"`
	SymbolID       int     `json:"symbolId"`
	Tier           string  `json:"tier"`
	BidPrice       float64 `json:"bidPrice"`
	BidSize        int     `json:"bidSize"`
	AskPrice       float64 `json:"askPrice"`
	AskSize        int     `json:"askSize"`
	LastTradePrice float64 `json:"lastTradePrice"`
	LastTradeSize  int     `json:"lastTradeSize"`
	LastTradeTime  string  `json:"lastTradeTime"`
	LastTradeTick  string  `json:"lastTradeTick"`
	Change         float64 `json:"change"`
	ChangePct      float64 `json:"changePercent"`
	Volume         int64   `json:"volume"`
	OpenPrice      float64 `json:"openPrice"`
	HighPrice      float64 `json:"highPrice"`
	LowPrice       float64 `json:"lowPrice"`
	Delay          int     `json:"delay"`
	IsHalted       bool    `json:"isHalted"`
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
	Symbol          string `json:"symbol"`
	SymbolID        int    `json:"symbolId"`
	Description     string `json:"description"`
	SecurityType    string `json:"securityType"`
	ListingExchange string `json:"listingExchange"`
	Currency        string `json:"currency"`
	IsQuotable      bool   `json:"isQuotable"`
	IsTradable      bool   `json:"isTradable"`
}

type QuestradeOption struct {
	Symbol             string  `json:"symbol"`
	SymbolID           int     `json:"symbolId"`
	ExpirationDate     string  `json:"expirationDate"`
	Strike             float64 `json:"strike"`
	CallPut            string  `json:"callPut"`
	BidPrice           float64 `json:"bidPrice"`
	AskPrice           float64 `json:"askPrice"`
	LastTradePrice     float64 `json:"lastTradePrice"`
	Volume             int     `json:"volume"`
	OpenInterest       int     `json:"openInterest"`
	Description        string  `json:"description"`
	IntrinsicValue     float64 `json:"intrinsicValue"`
	OptionOpenInterest int     `json:"optionOpenInterest"`
	Delta              float64 `json:"delta"`
	Gamma              float64 `json:"gamma"`
	Theta              float64 `json:"theta"`
	Vega               float64 `json:"vega"`
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
	qtInterval := questradeCandleInterval(interval)
	endpoint := fmt.Sprintf("/v1/markets/candles/%s?startTime=%s&endTime=%s&interval=%s",
		symbolID,
		url.QueryEscape(startTime.Format(time.RFC3339)),
		url.QueryEscape(endTime.Format(time.RFC3339)),
		url.QueryEscape(qtInterval))

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

func (p *QuestradeProvider) SearchTickers(prefix string) ([]TickerSearchResult, error) {
	symbols, err := p.FetchSymbolSearch(prefix)
	if err != nil {
		return nil, err
	}
	results := make([]TickerSearchResult, 0, len(symbols))
	for _, s := range symbols {
		if s.IsTradable {
			results = append(results, TickerSearchResult{
				Symbol:   s.Symbol,
				Name:     s.Description,
				Exchange: s.ListingExchange,
				Type:     s.SecurityType,
				SymbolID: s.SymbolID,
			})
		}
	}
	return results, nil
}

func (p *QuestradeProvider) FetchSymbolSearch(prefix string) ([]QuestradeSymbol, error) {
	endpoint := "/v1/symbols/search?prefix=" + url.QueryEscape(prefix)
	body, err := p.doRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var result struct {
		Symbols []struct {
			Symbol          string `json:"symbol"`
			SymbolID        int    `json:"symbolId"`
			Description     string `json:"description"`
			SecurityType    string `json:"securityType"`
			ListingExchange string `json:"listingExchange"`
			IsQuotable      bool   `json:"isQuotable"`
			IsTradable      bool   `json:"isTradable"`
			Currency        string `json:"currency"`
		} `json:"symbols"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	symbols := make([]QuestradeSymbol, 0, len(result.Symbols))
	for _, s := range result.Symbols {
		symbols = append(symbols, QuestradeSymbol{
			Symbol:          s.Symbol,
			SymbolID:        s.SymbolID,
			Description:     s.Description,
			SecurityType:    s.SecurityType,
			ListingExchange: s.ListingExchange,
			Currency:        s.Currency,
			IsQuotable:      s.IsQuotable,
			IsTradable:      s.IsTradable,
		})
	}
	return symbols, nil
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
	symbolID, err := p.GetSymbolID(ticker)
	if err != nil {
		return nil, err
	}
	quote, err := p.FetchQuote(strconv.Itoa(symbolID))
	if err != nil {
		return nil, err
	}

	price := quote.LastTradePrice
	if price == 0 {
		switch {
		case quote.BidPrice > 0 && quote.AskPrice > 0:
			price = (quote.BidPrice + quote.AskPrice) / 2
		case quote.BidPrice > 0:
			price = quote.BidPrice
		case quote.AskPrice > 0:
			price = quote.AskPrice
		}
	}

	timestamp := time.Now()
	if quote.LastTradeTime != "" {
		if parsed, err := ParseQuestradeTimestamp(quote.LastTradeTime); err == nil && !parsed.IsZero() {
			timestamp = parsed
		}
	}

	return &Price{
		Ticker:    quote.Symbol,
		Price:     price,
		Change:    quote.Change,
		ChangePct: quote.ChangePct,
		DayOpen:   quote.OpenPrice,
		Bid:       quote.BidPrice,
		Ask:       quote.AskPrice,
		Volume:    quote.Volume,
		Source:    p.Name(),
		Timestamp: timestamp,
	}, nil
}

func (p *QuestradeProvider) FetchOptionChain(ticker string) ([]*OptionChain, error) {
	symbolID, err := p.GetSymbolID(ticker)
	if err != nil {
		return nil, err
	}

	body, err := p.doRequest("/v1/symbols/" + strconv.Itoa(symbolID) + "/options")
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
	candles, err := p.FetchCandlesBySymbol(ticker, time.Now().Add(-24*time.Hour), time.Now(), interval)
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
	wanted := strings.ToUpper(strings.TrimSpace(ticker))
	if wanted == "" {
		return 0, fmt.Errorf("symbol not found: %s", ticker)
	}

	p.symbolIDMu.RLock()
	if symbolID, ok := p.symbolIDCache[wanted]; ok {
		p.symbolIDMu.RUnlock()
		return symbolID, nil
	}
	p.symbolIDMu.RUnlock()

	symbols, err := p.FetchSymbolSearch(ticker)
	if err != nil {
		return 0, err
	}

	for _, s := range symbols {
		if strings.ToUpper(strings.TrimSpace(s.Symbol)) == wanted {
			p.symbolIDMu.Lock()
			if p.symbolIDCache == nil {
				p.symbolIDCache = map[string]int{}
			}
			p.symbolIDCache[wanted] = s.SymbolID
			p.symbolIDMu.Unlock()
			return s.SymbolID, nil
		}
	}
	return 0, fmt.Errorf("symbol not found: %s", ticker)
}

func (p *QuestradeProvider) FetchCompanyProfile(ticker string) (*CompanyProfile, error) {
	symbolID, err := p.GetSymbolID(ticker)
	if err != nil {
		return nil, err
	}
	sym, err := p.FetchSymbol(strconv.Itoa(symbolID))
	if err != nil {
		return nil, err
	}
	return &CompanyProfile{
		Symbol:    sym.Symbol,
		Name:      sym.Description,
		Exchange:  sym.ListingExchange,
		Sector:    "",
		Industry:  "",
		Price:     0,
		MarketCap: 0,
	}, nil
}

func (p *QuestradeProvider) FetchFinancialRatios(ticker string) ([]*FinancialRatio, error) {
	return []*FinancialRatio{}, nil
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

func questradeCandleInterval(interval string) string {
	switch strings.ToLower(strings.TrimSpace(interval)) {
	case "1m", "1min", "1minute":
		return "OneMinute"
	case "5m", "5min", "5minute":
		return "FiveMinutes"
	case "15m", "15min", "15minute":
		return "FifteenMinutes"
	case "30m", "30min", "30minute":
		return "ThirtyMinutes"
	case "1h", "1hr", "1hour":
		return "OneHour"
	case "1d", "1day", "daily":
		return "OneDay"
	default:
		return "OneMinute"
	}
}

func (p *QuestradeProvider) FetchCandlesBySymbol(ticker string, startTime, endTime time.Time, interval string) ([]QuestradeCandle, error) {
	symbolID, err := p.GetSymbolID(ticker)
	if err != nil {
		return nil, err
	}
	return p.FetchCandles(strconv.Itoa(symbolID), startTime, endTime, interval)
}
