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
	cfg       config.FMPConfig
	client    *http.Client
	logClient *logging.Client
}

func NewFMPProvider(cfg config.FMPConfig, logClient *logging.Client) *FMPProvider {
	return &FMPProvider{
		cfg:       cfg,
		client:    &http.Client{Timeout: 10 * time.Second},
		logClient: logClient,
	}
}

func (p *FMPProvider) Name() string {
	return "fmp"
}

func (p *FMPProvider) FetchPrice(ticker string) (*Price, error) {
	url := fmt.Sprintf("https://financialmodelingprep.com/api/v3/quote/%s?apikey=%s", ticker, p.cfg.APIKey)

	p.logRequest("GET", url, nil)

	resp, err := p.client.Get(url)
	if err != nil {
		p.logResponseError("GET", url, 0, err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		p.logResponseError("GET", url, resp.StatusCode, err.Error())
		return nil, err
	}

	p.logResponse("GET", url, resp.StatusCode, body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fmp request failed: %d - %s", resp.StatusCode, string(body))
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

func (p *FMPProvider) logRequest(method, url string, body []byte) {
	if p.logClient == nil {
		return
	}
	p.logClient.InfoWithMeta(nil, "FMP Request: "+method+" "+url, map[string]interface{}{
		"method": method,
		"url":    url,
		"body":   string(body),
	})
}

func (p *FMPProvider) logResponse(method, url string, status int, body []byte) {
	if p.logClient == nil {
		return
	}
	level := "INFO"
	if status >= 400 {
		level = "ERROR"
	}
	p.logClient.EmitWithMetadata(nil, level, "FMP Response: "+method+" "+url+" -> "+fmt.Sprintf("%d", status), map[string]interface{}{
		"method":  method,
		"url":     url,
		"status":  status,
		"body":    string(body),
	})
}

func (p *FMPProvider) logResponseError(method, url string, status int, errMsg string) {
	if p.logClient == nil {
		return
	}
	p.logClient.ErrorWithMeta(nil, "FMP Error: "+method+" "+url, map[string]interface{}{
		"method": method,
		"url":    url,
		"status": status,
		"error":  errMsg,
	})
}

func (p *FMPProvider) FetchOptionChain(ticker string) ([]*OptionChain, error) {
	return nil, fmt.Errorf("fmp does not support option chains via this endpoint")
}

func (p *FMPProvider) FetchIntradayBars(ticker string, interval string) ([]*IntradayBar, error) {
	return nil, fmt.Errorf("fmp does not support intraday bars via this endpoint")
}

func (p *FMPProvider) SearchTickers(prefix string) ([]TickerSearchResult, error) {
	searchURL := fmt.Sprintf("https://financialmodelingprep.com/api/v3/search?query=%s&limit=20&apikey=%s", url.QueryEscape(prefix), p.cfg.APIKey)
	resp, err := p.client.Get(searchURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fmp search failed: %d", resp.StatusCode)
	}

	var result []struct {
		Symbol   string `json:"symbol"`
		Name     string `json:"name"`
		Exchange string `json:"exchangeShortName"`
		Type     string `json:"type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
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
	resp, err := p.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result []*IncomeStatement
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (p *FMPProvider) FetchBalanceSheet(ticker string) ([]*BalanceSheet, error) {
	url := fmt.Sprintf("https://financialmodelingprep.com/api/v3/balance-sheet-statement/%s?apikey=%s", ticker, p.cfg.APIKey)
	resp, err := p.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result []*BalanceSheet
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (p *FMPProvider) FetchCashFlow(ticker string) ([]*CashFlow, error) {
	url := fmt.Sprintf("https://financialmodelingprep.com/api/v3/cash-flow-statement/%s?apikey=%s", ticker, p.cfg.APIKey)
	resp, err := p.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result []*CashFlow
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}