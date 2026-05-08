package providers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/portfolio-sim/market-data-service/config"
)

type FMPProvider struct {
	cfg    config.FMPConfig
	client *http.Client
}

func NewFMPProvider(cfg config.FMPConfig) *FMPProvider {
	return &FMPProvider{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *FMPProvider) Name() string {
	return "fmp"
}

func (p *FMPProvider) FetchPrice(ticker string) (*Price, error) {
	url := fmt.Sprintf("https://financialmodelingprep.com/api/v3/quote/%s?apikey=%s", ticker, p.cfg.APIKey)
	resp, err := p.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fmp request failed: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result []struct {
		Symbol       string  `json:"symbol"`
		Price        float64 `json:"price"`
		Bid          float64 `json:"bid"`
		Ask          float64 `json:"ask"`
		Volume       int64   `json:"volume"`
		Timestamp    int64   `json:"timestamp"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no quote for %s", ticker)
	}

	q := result[0]
	return &Price{
		Ticker:   q.Symbol,
		Price:    q.Price,
		Bid:      q.Bid,
		Ask:      q.Ask,
		Volume:   q.Volume,
		Source:   p.Name(),
		Timestamp: time.UnixMilli(q.Timestamp),
	}, nil
}

func (p *FMPProvider) FetchOptionChain(ticker string) ([]*OptionChain, error) {
	return nil, fmt.Errorf("fmp does not support option chains via this endpoint")
}

func (p *FMPProvider) FetchIntradayBars(ticker string, interval string) ([]*IntradayBar, error) {
	return nil, fmt.Errorf("fmp does not support intraday bars via this endpoint")
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
	Symbol           string  `json:"symbol"`
	OperatingCashFlow float64 `json:"operatingCashFlow"`
	InvestingCashFlow float64 `json:"investingCashFlow"`
	FinancingCashFlow float64 `json:"financingCashFlow"`
	FreeCashFlow     float64 `json:"freeCashFlow"`
	Period           string  `json:"period"`
	FiscalDateEnding string  `json:"fiscalDateEnding"`
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
