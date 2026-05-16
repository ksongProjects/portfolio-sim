package providers

import (
	"context"
	"fmt"
	"strings"
	"time"

	massive "github.com/massive-com/client-go/v3/rest"
	"github.com/massive-com/client-go/v3/rest/gen"
	"github.com/portfolio-sim/market-data-service/config"
	"github.com/portfolio-sim/market-data-service/logging"
)

type MassiveProvider struct {
	cfg            config.MassiveConfig
	logClient      *logging.Client
	apiKeyResolver APIKeyResolver
}

func NewMassiveProvider(cfg config.MassiveConfig, logClient *logging.Client, apiKeyResolver APIKeyResolver) *MassiveProvider {
	return &MassiveProvider{
		cfg:            cfg,
		logClient:      logClient,
		apiKeyResolver: apiKeyResolver,
	}
}

func (p *MassiveProvider) Name() string {
	return "massive"
}

func (p *MassiveProvider) GetSymbolID(ticker string) (int, error) {
	return 0, fmt.Errorf("massive does not use symbol IDs")
}

func (p *MassiveProvider) FetchCompanyProfile(ticker string) (*CompanyProfile, error) {
	client, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := client.GetTickerWithResponse(context.Background(), ticker, nil)
	if err != nil {
		return nil, err
	}
	if err := massive.CheckResponse(resp); err != nil {
		return nil, fmt.Errorf("massive profile failed: %w", err)
	}
	if resp.JSON200 == nil || resp.JSON200.Results == nil {
		return nil, fmt.Errorf("no profile for %s", ticker)
	}

	result := resp.JSON200.Results
	profile := &CompanyProfile{
		Symbol: result.Ticker,
		Name:   result.Name,
	}
	if result.PrimaryExchange != nil {
		profile.Exchange = *result.PrimaryExchange
	}
	if result.SicDescription != nil {
		profile.Industry = *result.SicDescription
	}
	if result.Description != nil {
		profile.Description = *result.Description
	}
	if result.HomepageUrl != nil {
		profile.Website = *result.HomepageUrl
	}
	if result.MarketCap != nil {
		profile.MarketCap = *result.MarketCap
	}

	ratios, err := p.fetchLatestRatiosSnapshot(client, ticker)
	if err != nil {
		if p.logClient != nil {
			p.logClient.WarnWithMeta(nil, "massive profile ratio enrichment failed", map[string]interface{}{
				"provider": p.Name(),
				"symbol":   ticker,
				"error":    err.Error(),
			})
		}
		return profile, nil
	}

	applyMassiveRatiosToProfile(profile, ratios)
	return profile, nil
}

func (p *MassiveProvider) FetchFinancialRatios(ticker string) ([]*FinancialRatio, error) {
	client, err := p.client()
	if err != nil {
		return nil, err
	}

	ratios, err := p.fetchLatestRatiosSnapshot(client, ticker)
	if err != nil {
		return nil, err
	}
	if ratios == nil {
		return []*FinancialRatio{}, nil
	}

	result := make([]*FinancialRatio, 0, 10)
	appendRatioValue(&result, "P/E Ratio", ratios.PriceToEarnings, false, "Price to earnings ratio")
	appendRatioValue(&result, "P/B Ratio", ratios.PriceToBook, false, "Price to book value")
	appendRatioValue(&result, "P/S Ratio", ratios.PriceToSales, false, "Price to sales ratio")
	appendRatioValue(&result, "Dividend Yield", ratios.DividendYield, true, "Annual dividend yield")
	appendRatioValue(&result, "Debt/Equity", ratios.DebtToEquity, false, "Debt to equity ratio")
	appendRatioValue(&result, "Current Ratio", ratios.Current, false, "Current liquidity ratio")
	appendRatioValue(&result, "Quick Ratio", ratios.Quick, false, "Immediate liquidity ratio")
	appendRatioValue(&result, "ROE", ratios.ReturnOnEquity, true, "Return on equity")
	appendRatioValue(&result, "ROA", ratios.ReturnOnAssets, true, "Return on assets")
	appendRatioValue(&result, "EV/EBITDA", ratios.EvToEbitda, false, "Enterprise value to EBITDA")
	appendRatioValue(&result, "EV/Revenue", ratios.EvToSales, false, "Enterprise value to revenue")
	appendRatioValue(&result, "P/Cash Flow", ratios.PriceToCashFlow, false, "Price to operating cash flow")
	appendRatioValue(&result, "P/Free Cash Flow", ratios.PriceToFreeCashFlow, false, "Price to free cash flow")

	return result, nil
}

func (p *MassiveProvider) apiKey() (string, error) {
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
		return "", fmt.Errorf("massive API key not configured")
	}
	return p.cfg.APIKey, nil
}

func (p *MassiveProvider) client() (*massive.Client, error) {
	apiKey, err := p.apiKey()
	if err != nil {
		return nil, err
	}

	return massive.NewWithOptions(
		apiKey,
		massive.WithTrace(false),
		massive.WithPagination(false),
	), nil
}

func (p *MassiveProvider) FetchPrice(ticker string) (*Price, error) {
	client, err := p.client()
	if err != nil {
		return nil, err
	}

	resp, err := client.GetStocksSnapshotTickerWithResponse(context.Background(), ticker)
	if err != nil {
		return nil, err
	}
	if err := massive.CheckResponse(resp); err != nil {
		return nil, fmt.Errorf("massive request failed: %w", err)
	}

	if resp.JSON200 == nil || resp.JSON200.Ticker == nil {
		return nil, fmt.Errorf("no price data for %s", ticker)
	}

	snapshot := resp.JSON200.Ticker
	price := 0.0
	bid := 0.0
	ask := 0.0
	volume := int64(0)
	timestamp := time.Now()

	if snapshot.Min != nil {
		price = snapshot.Min.C
		volume = int64(snapshot.Min.Av)
		timestamp = time.UnixMilli(int64(snapshot.Min.Timestamp))
	}
	if price == 0 && snapshot.LastTrade != nil {
		price = snapshot.LastTrade.BidPrice
		timestamp = time.Unix(0, int64(snapshot.LastTrade.Timestamp))
	}
	if price == 0 && snapshot.PrevDay != nil {
		price = snapshot.PrevDay.C
		volume = int64(snapshot.PrevDay.V)
	}
	if snapshot.LastQuote != nil {
		bid = snapshot.LastQuote.BidPrice
		ask = snapshot.LastQuote.AskPrice
	}
	if price == 0 {
		return nil, fmt.Errorf("no price data for %s", ticker)
	}

	change := 0.0
	changePct := 0.0
	if snapshot.TodaysChange != nil {
		change = *snapshot.TodaysChange
	}
	if snapshot.TodaysChangePerc != nil {
		changePct = *snapshot.TodaysChangePerc
	}

	return &Price{
		Ticker:    ticker,
		Price:     price,
		Change:    change,
		ChangePct: changePct,
		DayOpen:   0,
		Bid:       bid,
		Ask:       ask,
		Volume:    volume,
		Source:    p.Name(),
		Timestamp: timestamp,
	}, nil
}

func (p *MassiveProvider) FetchOptionChain(ticker string) ([]*OptionChain, error) {
	return nil, fmt.Errorf("massive does not support option chains via this endpoint")
}

func (p *MassiveProvider) FetchIntradayBars(ticker string, interval string) ([]*IntradayBar, error) {
	client, err := p.client()
	if err != nil {
		return nil, err
	}

	multiplier, timespan := massiveAggregateResolution(interval)

	from := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	to := time.Now().Format("2006-01-02")
	params := &gen.GetStocksAggregatesParams{
		Adjusted: massive.Ptr(true),
		Sort:     "asc",
		Limit:    massive.Ptr(5000),
	}

	resp, err := client.GetStocksAggregatesWithResponse(
		context.Background(),
		ticker,
		multiplier,
		gen.GetStocksAggregatesParamsTimespan(timespan),
		from,
		to,
		params,
	)
	if err != nil {
		return nil, err
	}
	if err := massive.CheckResponse(resp); err != nil {
		return nil, fmt.Errorf("massive request failed: %w", err)
	}
	if resp.JSON200 == nil || resp.JSON200.Results == nil {
		return nil, fmt.Errorf("no intraday data for %s", ticker)
	}

	bars := make([]*IntradayBar, 0, len(*resp.JSON200.Results))
	for _, b := range *resp.JSON200.Results {
		bars = append(bars, &IntradayBar{
			Ticker:    ticker,
			Interval:  interval,
			Open:      b.O,
			High:      b.H,
			Low:       b.L,
			Close:     b.C,
			Volume:    int64(b.V),
			Timestamp: time.UnixMilli(int64(b.Timestamp)),
		})
	}
	return bars, nil
}

func (p *MassiveProvider) SearchTickers(prefix string) ([]TickerSearchResult, error) {
	client, err := p.client()
	if err != nil {
		return nil, err
	}

	params := &gen.ListTickersParams{
		Search: massive.String(prefix),
		Active: massive.Bool(true),
		Limit:  massive.Int(20),
		Market: massive.Ptr(gen.ListTickersParamsMarket("stocks")),
	}

	resp, err := client.ListTickersWithResponse(context.Background(), params)
	if err != nil {
		return nil, err
	}
	if err := massive.CheckResponse(resp); err != nil {
		return nil, fmt.Errorf("massive search failed: %w", err)
	}
	if resp.JSON200 == nil || resp.JSON200.Results == nil {
		return []TickerSearchResult{}, nil
	}

	results := make([]TickerSearchResult, 0, len(*resp.JSON200.Results))
	for _, t := range *resp.JSON200.Results {
		exchange := ""
		if t.PrimaryExchange != nil {
			exchange = *t.PrimaryExchange
		}
		tickerType := ""
		if t.Type != nil {
			tickerType = *t.Type
		}

		results = append(results, TickerSearchResult{
			Symbol:   t.Ticker,
			Name:     t.Name,
			Exchange: exchange,
			Type:     tickerType,
		})
	}
	return results, nil
}

type massiveRatiosSnapshot struct {
	AverageVolume       *float64
	Current             *float64
	DebtToEquity        *float64
	DividendYield       *float64
	EarningsPerShare    *float64
	EvToEbitda          *float64
	EvToSales           *float64
	MarketCap           *float64
	PriceToBook         *float64
	PriceToCashFlow     *float64
	PriceToEarnings     *float64
	PriceToFreeCashFlow *float64
	PriceToSales        *float64
	Quick               *float64
	ReturnOnAssets      *float64
	ReturnOnEquity      *float64
}

func (p *MassiveProvider) fetchLatestRatiosSnapshot(client *massive.Client, ticker string) (*massiveRatiosSnapshot, error) {
	params := &gen.GetStocksFinancialsV1RatiosParams{
		Ticker: massive.String(ticker),
		Limit:  massive.Int(1),
		Sort:   massive.String("date.desc"),
	}

	resp, err := client.GetStocksFinancialsV1RatiosWithResponse(context.Background(), params)
	if err != nil {
		return nil, err
	}
	if err := massive.CheckResponse(resp); err != nil {
		return nil, fmt.Errorf("massive ratios failed: %w", err)
	}
	if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
		return nil, nil
	}

	result := resp.JSON200.Results[0]
	return &massiveRatiosSnapshot{
		AverageVolume:       result.AverageVolume,
		Current:             result.Current,
		DebtToEquity:        result.DebtToEquity,
		DividendYield:       result.DividendYield,
		EarningsPerShare:    result.EarningsPerShare,
		EvToEbitda:          result.EvToEbitda,
		EvToSales:           result.EvToSales,
		MarketCap:           result.MarketCap,
		PriceToBook:         result.PriceToBook,
		PriceToCashFlow:     result.PriceToCashFlow,
		PriceToEarnings:     result.PriceToEarnings,
		PriceToFreeCashFlow: result.PriceToFreeCashFlow,
		PriceToSales:        result.PriceToSales,
		Quick:               result.Quick,
		ReturnOnAssets:      result.ReturnOnAssets,
		ReturnOnEquity:      result.ReturnOnEquity,
	}, nil
}

func applyMassiveRatiosToProfile(profile *CompanyProfile, ratios *massiveRatiosSnapshot) {
	if profile == nil || ratios == nil {
		return
	}

	if profile.MarketCap == 0 && ratios.MarketCap != nil {
		profile.MarketCap = *ratios.MarketCap
	}
	if profile.PeRatio == 0 && ratios.PriceToEarnings != nil {
		profile.PeRatio = *ratios.PriceToEarnings
	}
	if profile.Eps == 0 && ratios.EarningsPerShare != nil {
		profile.Eps = *ratios.EarningsPerShare
	}
	if profile.DivYield == 0 && ratios.DividendYield != nil {
		profile.DivYield = *ratios.DividendYield
	}
	if profile.AvgVolume == 0 && ratios.AverageVolume != nil {
		profile.AvgVolume = int64(*ratios.AverageVolume)
	}
}

func appendRatioValue(target *[]*FinancialRatio, label string, value *float64, percent bool, description string) {
	if target == nil || value == nil {
		return
	}

	formatted := fmt.Sprintf("%.2f", *value)
	if percent {
		formatted = fmt.Sprintf("%.2f%%", *value*100)
	}

	*target = append(*target, &FinancialRatio{
		Label:       label,
		Value:       formatted,
		Description: description,
	})
}

func massiveAggregateResolution(interval string) (int, string) {
	switch strings.ToLower(strings.TrimSpace(interval)) {
	case "1m", "1min", "1minute":
		return 1, "minute"
	case "5m", "5min", "5minute":
		return 5, "minute"
	case "15m", "15min", "15minute":
		return 15, "minute"
	case "30m", "30min", "30minute":
		return 30, "minute"
	case "1h", "1hr", "1hour":
		return 1, "hour"
	default:
		return 1, "minute"
	}
}
