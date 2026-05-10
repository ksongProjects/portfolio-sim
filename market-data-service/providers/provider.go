package providers

import (
	"time"
)

type Price struct {
	Ticker    string
	Price     float64
	Bid       float64
	Ask       float64
	Volume    int64
	Source    string
	Timestamp time.Time
}

type OptionChain struct {
	Ticker       string
	Expiration   time.Time
	Strike       float64
	OptionType   string
	Bid          float64
	Ask          float64
	Delta        float64
	Gamma        float64
	Theta        float64
	Vega         float64
	ImpliedVol   float64
	Volume       int64
	OpenInterest int64
	Source       string
	Timestamp    time.Time
}

type IntradayBar struct {
	Ticker    string
	Interval  string
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    int64
	Timestamp time.Time
}

type Provider interface {
	FetchPrice(ticker string) (*Price, error)
	FetchOptionChain(ticker string) ([]*OptionChain, error)
	FetchIntradayBars(ticker string, interval string) ([]*IntradayBar, error)
	SearchTickers(prefix string) ([]TickerSearchResult, error)
	GetSymbolID(ticker string) (int, error)
	FetchCompanyProfile(ticker string) (*CompanyProfile, error)
	FetchFinancialRatios(ticker string) ([]*FinancialRatio, error)
	Name() string
}

type APIKeyResolver func() (string, error)

type CompanyProfile struct {
	Symbol      string  `json:"symbol"`
	Name        string  `json:"name"`
	Exchange    string  `json:"exchange"`
	Sector      string  `json:"sector"`
	Industry    string  `json:"industry"`
	CEO         string  `json:"ceo"`
	Website     string  `json:"website"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	MarketCap   float64 `json:"marketCap"`
	PeRatio     float64 `json:"pe"`
	Eps         float64 `json:"eps"`
	DivYield    float64 `json:"dividendYield"`
	Week52High  float64 `json:"year52High"`
	Week52Low   float64 `json:"year52Low"`
	AvgVolume   int64   `json:"avgVolume"`
}

type FinancialRatio struct {
	Label       string `json:"label"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

type TickerSearchResult struct {
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Exchange string `json:"exchange"`
	Type     string `json:"type"`
	SymbolID int    `json:"symbolId"`
}
