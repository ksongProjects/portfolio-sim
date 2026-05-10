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

type FMPProvider struct {
	cfg            config.FMPConfig
	client         *http.Client
	logClient      *logging.Client
	apiKeyResolver APIKeyResolver
}

func NewFMPProvider(cfg config.FMPConfig, logClient *logging.Client, apiKeyResolver APIKeyResolver) *FMPProvider {
	return &FMPProvider{
		cfg:            cfg,
		client:         &http.Client{Timeout: 10 * time.Second},
		logClient:      logClient,
		apiKeyResolver: apiKeyResolver,
	}
}

func (p *FMPProvider) Name() string {
	return "fmp"
}

func (p *FMPProvider) GetSymbolID(ticker string) (int, error) {
	return 0, fmt.Errorf("fmp does not use symbol IDs")
}

func (p *FMPProvider) FetchCompanyProfile(ticker string) (*CompanyProfile, error) {
	url := fmt.Sprintf("https://financialmodelingprep.com/api/v3/profile/%s", ticker)
	body, status, _, err := p.get(url)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("fmp profile failed: %d - %s", status, string(body))
	}
	var result []struct {
		Symbol      string  `json:"symbol"`
		CompanyName string  `json:"companyName"`
		Exchange    string  `json:"exchange"`
		Sector      string  `json:"sector"`
		Industry    string  `json:"industry"`
		CEO         string  `json:"ceo"`
		Website     string  `json:"website"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		MarketCap   float64 `json:"mktCap"`
		Pe          float64 `json:"pe"`
		Eps         float64 `json:"eps"`
		DivYield    float64 `json:"dividendYield"`
		Year52High  float64 `json:"year52High"`
		Year52Low   float64 `json:"year52Low"`
		AvgVolume   int64   `json:"volAvg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no profile for %s", ticker)
	}
	r := result[0]
	return &CompanyProfile{
		Symbol:      r.Symbol,
		Name:        r.CompanyName,
		Exchange:    r.Exchange,
		Sector:      r.Sector,
		Industry:    r.Industry,
		CEO:         r.CEO,
		Website:     r.Website,
		Description: r.Description,
		Price:       r.Price,
		MarketCap:   r.MarketCap,
		PeRatio:     r.Pe,
		Eps:         r.Eps,
		DivYield:    r.DivYield,
		Week52High:  r.Year52High,
		Week52Low:   r.Year52Low,
		AvgVolume:   r.AvgVolume,
	}, nil
}

func (p *FMPProvider) FetchFinancialRatios(ticker string) ([]*FinancialRatio, error) {
	url := fmt.Sprintf("https://financialmodelingprep.com/api/v3/ratios/%s", ticker)
	body, status, _, err := p.get(url)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("fmp ratios failed: %d - %s", status, string(body))
	}
	var result []struct {
		Symbol                   string  `json:"symbol"`
		PE                       float64 `json:"peRatio"`
		PEG                      float64 `json:"pegRatio"`
		PB                       float64 `json:"pbRatio"`
		PS                       float64 `json:"pcfRatio"`
		DividendYield            float64 `json:"dividendYield"`
		DebtToEquity             float64 `json:"debtToEquity"`
		CurrentRatio             float64 `json:"currentRatio"`
		QuickRatio               float64 `json:"cashRatio"`
		ReturnOnEquity           float64 `json:"returnOnEquity"`
		ReturnOnAssets           float64 `json:"returnOnAssets"`
		ProfitMargin             float64 `json:"profitMargin"`
		OperatingMargin          float64 `json:"operatingMargin"`
		GrossMargin              float64 `json:"grossProfitMargin"`
		AssetTurnover            float64 `json:"assetTurnover"`
		InventoryTurnover        float64 `json:"inventoryTurnover"`
		ReceivableTurnover       float64 `json:"receivableTurnover"`
		PriceToBook              float64 `json:"priceToBookRatio"`
		PriceToEarnings          float64 `json:"priceEarningsRatio"`
		PriceToFreeCashFlow      float64 `json:"priceToFreeCashFlowsRatio"`
		PriceToOperatingCashFlow float64 `json:"priceToOperatingCashFlowsRatio"`
		PriceToSales             float64 `json:"priceToSalesRatio"`
		EVToEBITDA               float64 `json:"enterpriseValueMultiple"`
		EVToRevenue              float64 `json:"evToRevenue"`
		EVToOperatingIncome      float64 `json:"evToOperatingIncome"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return []*FinancialRatio{}, nil
	}
	r := result[0]
	ratios := []*FinancialRatio{
		{Label: "P/E Ratio", Value: fmt.Sprintf("%.2f", r.PE), Description: "Price to earnings ratio"},
		{Label: "PEG Ratio", Value: fmt.Sprintf("%.2f", r.PEG), Description: "Price to earnings growth"},
		{Label: "P/B Ratio", Value: fmt.Sprintf("%.2f", r.PB), Description: "Price to book value"},
		{Label: "Dividend Yield", Value: fmt.Sprintf("%.2f%%", r.DividendYield*100), Description: "Annual dividend yield"},
		{Label: "Debt/Equity", Value: fmt.Sprintf("%.2f", r.DebtToEquity), Description: "Debt to equity ratio"},
		{Label: "Current Ratio", Value: fmt.Sprintf("%.2f", r.CurrentRatio), Description: "Current liquidity ratio"},
		{Label: "Quick Ratio", Value: fmt.Sprintf("%.2f", r.QuickRatio), Description: "Cash to current liabilities"},
		{Label: "ROE", Value: fmt.Sprintf("%.2f%%", r.ReturnOnEquity*100), Description: "Return on equity"},
		{Label: "ROA", Value: fmt.Sprintf("%.2f%%", r.ReturnOnAssets*100), Description: "Return on assets"},
		{Label: "Profit Margin", Value: fmt.Sprintf("%.2f%%", r.ProfitMargin*100), Description: "Net profit margin"},
		{Label: "Op. Margin", Value: fmt.Sprintf("%.2f%%", r.OperatingMargin*100), Description: "Operating margin"},
		{Label: "Gross Margin", Value: fmt.Sprintf("%.2f%%", r.GrossMargin*100), Description: "Gross profit margin"},
		{Label: "EV/EBITDA", Value: fmt.Sprintf("%.2f", r.EVToEBITDA), Description: "Enterprise value to EBITDA"},
		{Label: "EV/Revenue", Value: fmt.Sprintf("%.2f", r.EVToRevenue), Description: "Enterprise value to revenue"},
	}
	return ratios, nil
}

func (p *FMPProvider) FetchPrice(ticker string) (*Price, error) {
	url := fmt.Sprintf("https://financialmodelingprep.com/api/v3/quote/%s", ticker)
	body, status, _, err := p.get(url)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("fmp request failed: %d - %s", status, string(body))
	}

	var result []struct {
		Symbol    string  `json:"symbol"`
		Price     float64 `json:"price"`
		Bid       float64 `json:"bid"`
		Ask       float64 `json:"ask"`
		Volume    int64   `json:"volume"`
		Timestamp int64   `json:"timestamp"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no quote for %s", ticker)
	}

	q := result[0]
	return &Price{
		Ticker:    q.Symbol,
		Price:     q.Price,
		Bid:       q.Bid,
		Ask:       q.Ask,
		Volume:    q.Volume,
		Source:    p.Name(),
		Timestamp: time.UnixMilli(q.Timestamp),
	}, nil
}

func (p *FMPProvider) logRequest(method, rawURL string, headers http.Header, body []byte) {
	if p.logClient == nil {
		return
	}
	safeURL := logging.SanitizeURL(rawURL)
	meta := map[string]interface{}{
		"method":          method,
		"url":             safeURL,
		"provider":        p.Name(),
		"request_headers": logging.RedactHeaders(headers),
	}
	if len(body) > 0 {
		meta["body"] = logging.SanitizeBody(headers.Get("Content-Type"), body)
		meta["body_size"] = len(body)
	}
	p.logClient.InfoWithMeta(nil, "FMP Request: "+method+" "+safeURL, meta)
}

func (p *FMPProvider) logResponse(method, rawURL string, status int, headers http.Header, body []byte, duration time.Duration) {
	if p.logClient == nil {
		return
	}
	level := "INFO"
	if status >= 400 {
		level = "ERROR"
	}
	safeURL := logging.SanitizeURL(rawURL)
	p.logClient.EmitWithMetadata(nil, level, "FMP Response: "+method+" "+safeURL+" -> "+fmt.Sprintf("%d", status), map[string]interface{}{
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

func (p *FMPProvider) logResponseError(method, rawURL string, headers http.Header, status int, errMsg string, duration time.Duration) {
	if p.logClient == nil {
		return
	}
	safeURL := logging.SanitizeURL(rawURL)
	p.logClient.ErrorWithMeta(nil, "FMP Error: "+method+" "+safeURL, map[string]interface{}{
		"method":          method,
		"url":             safeURL,
		"status":          status,
		"provider":        p.Name(),
		"request_headers": logging.RedactHeaders(headers),
		"duration_ms":     duration.Milliseconds(),
		"error":           errMsg,
	})
}

func (p *FMPProvider) apiKey() (string, error) {
	if p.apiKeyResolver != nil {
		apiKey, err := p.apiKeyResolver()
		if err == nil && apiKey != "" {
			return apiKey, nil
		}
		if p.cfg.APIKey == "" && err != nil {
			return "", err
		}
	}
	if p.cfg.APIKey == "" {
		return "", fmt.Errorf("fmp API key not configured")
	}
	return p.cfg.APIKey, nil
}

func (p *FMPProvider) get(rawURL string) ([]byte, int, http.Header, error) {
	apiKey, err := p.apiKey()
	if err != nil {
		return nil, 0, nil, err
	}
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, 0, nil, err
	}
	req.Header.Set("apikey", apiKey)

	p.logRequest(http.MethodGet, rawURL, req.Header, nil)
	start := time.Now()
	resp, err := p.client.Do(req)
	duration := time.Since(start)
	if err != nil {
		p.logResponseError(http.MethodGet, rawURL, req.Header, 0, err.Error(), duration)
		return nil, 0, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		p.logResponseError(http.MethodGet, rawURL, req.Header, resp.StatusCode, err.Error(), duration)
		return nil, resp.StatusCode, resp.Header, err
	}

	p.logResponse(http.MethodGet, rawURL, resp.StatusCode, resp.Header, body, duration)
	return body, resp.StatusCode, resp.Header, nil
}

func (p *FMPProvider) FetchOptionChain(ticker string) ([]*OptionChain, error) {
	return nil, fmt.Errorf("fmp does not support option chains via this endpoint")
}

func (p *FMPProvider) FetchIntradayBars(ticker string, interval string) ([]*IntradayBar, error) {
	url := fmt.Sprintf("https://financialmodelingprep.com/api/v3/historical-chart/%s/%s", interval, ticker)
	body, status, _, err := p.get(url)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("fmp intraday failed: %d - %s", status, string(body))
	}

	var result []struct {
		Date   string  `json:"date"`
		Open   float64 `json:"open"`
		High   float64 `json:"high"`
		Low    float64 `json:"low"`
		Close  float64 `json:"close"`
		Volume int64   `json:"volume"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	bars := make([]*IntradayBar, 0, len(result))
	for _, r := range result {
		ts, _ := time.Parse("2006-01-02 15:04:05", r.Date)
		bars = append(bars, &IntradayBar{
			Ticker:    ticker,
			Interval:  interval,
			Open:      r.Open,
			High:      r.High,
			Low:       r.Low,
			Close:     r.Close,
			Volume:    r.Volume,
			Timestamp: ts,
		})
	}
	return bars, nil
}

func (p *FMPProvider) SearchTickers(prefix string) ([]TickerSearchResult, error) {
	searchURL := fmt.Sprintf("https://financialmodelingprep.com/api/v3/search?query=%s&limit=20", url.QueryEscape(prefix))
	body, status, _, err := p.get(searchURL)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("fmp search failed: %d - %s", status, string(body))
	}

	var result []struct {
		Symbol   string `json:"symbol"`
		Name     string `json:"name"`
		Exchange string `json:"exchangeShortName"`
		Type     string `json:"type"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	results := make([]TickerSearchResult, 0, len(result))
	for _, t := range result {
		results = append(results, TickerSearchResult{
			Symbol:   t.Symbol,
			Name:     t.Name,
			Exchange: t.Exchange,
			Type:     t.Type,
		})
	}
	return results, nil
}

type IncomeStatement struct {
	Symbol           string  `json:"symbol"`
	Revenue          float64 `json:"revenue"`
	NetIncome        float64 `json:"netIncome"`
	GrossProfit      float64 `json:"grossProfit"`
	EBITDA           float64 `json:"ebitda"`
	Eps              float64 `json:"eps"`
	Period           string  `json:"period"`
	FiscalDateEnding string  `json:"fiscalDateEnding"`
}

type BalanceSheet struct {
	Symbol           string  `json:"symbol"`
	TotalAssets      float64 `json:"totalAssets"`
	TotalLiabilities float64 `json:"totalLiabilities"`
	Equity           float64 `json:"totalEquity"`
	Period           string  `json:"period"`
	FiscalDateEnding string  `json:"fiscalDateEnding"`
}

type CashFlow struct {
	Symbol            string  `json:"symbol"`
	OperatingCashFlow float64 `json:"operatingCashFlow"`
	InvestingCashFlow float64 `json:"investingCashFlow"`
	FinancingCashFlow float64 `json:"financingCashFlow"`
	FreeCashFlow      float64 `json:"freeCashFlow"`
	Period            string  `json:"period"`
	FiscalDateEnding  string  `json:"fiscalDateEnding"`
}

func (p *FMPProvider) FetchIncomeStatement(ticker string) ([]*IncomeStatement, error) {
	url := fmt.Sprintf("https://financialmodelingprep.com/api/v3/income-statement/%s?apikey=%s", ticker, p.cfg.APIKey)
	body, status, _, err := p.get(url)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("fmp income statement failed: %d - %s", status, string(body))
	}

	var result []*IncomeStatement
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (p *FMPProvider) FetchBalanceSheet(ticker string) ([]*BalanceSheet, error) {
	url := fmt.Sprintf("https://financialmodelingprep.com/api/v3/balance-sheet-statement/%s?apikey=%s", ticker, p.cfg.APIKey)
	body, status, _, err := p.get(url)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("fmp balance sheet failed: %d - %s", status, string(body))
	}

	var result []*BalanceSheet
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (p *FMPProvider) FetchCashFlow(ticker string) ([]*CashFlow, error) {
	url := fmt.Sprintf("https://financialmodelingprep.com/api/v3/cash-flow-statement/%s?apikey=%s", ticker, p.cfg.APIKey)
	body, status, _, err := p.get(url)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("fmp cash flow failed: %d - %s", status, string(body))
	}

	var result []*CashFlow
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}
