package main

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/portfolio-sim/market-data-service/providers"
	"github.com/portfolio-sim/market-data-service/storage"
)

type marketDataOperation string

const (
	operationSearch      marketDataOperation = "search"
	operationQuote       marketDataOperation = "quote"
	operationIntraday    marketDataOperation = "intraday"
	operationOptionChain marketDataOperation = "option_chain"
	operationProfile     marketDataOperation = "profile"
	operationRatios      marketDataOperation = "ratios"
)

type providerDescriptor struct {
	name      string
	build     func() providers.Provider
	supported map[marketDataOperation]bool
	priority  map[marketDataOperation]int
}

type activeProvider struct {
	provider   providers.Provider
	descriptor providerDescriptor
}

type tickerDetailsResponse struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	Exchange      string  `json:"exchange"`
	Sector        string  `json:"sector"`
	Industry      string  `json:"industry"`
	Price         float64 `json:"price"`
	Change        float64 `json:"change"`
	ChangePct     float64 `json:"changePct"`
	Volume        int64   `json:"volume"`
	AvgVolume     int64   `json:"avgVolume"`
	MarketCap     float64 `json:"marketCap"`
	PeRatio       float64 `json:"peRatio"`
	Eps           float64 `json:"eps"`
	DividendYield float64 `json:"dividendYield"`
	Week52High    float64 `json:"week52High"`
	Week52Low     float64 `json:"week52Low"`
}

func (d providerDescriptor) supports(operation marketDataOperation) bool {
	return d.supported[operation]
}

func (d providerDescriptor) priorityFor(operation marketDataOperation) int {
	priority, ok := d.priority[operation]
	if !ok {
		return 999
	}

	return priority
}

func (s *MarketDataService) hasQuestradeProvider(ctx context.Context) bool {
	isValidated, err := s.storage.IsProviderValidated(ctx, "questrade")
	if err != nil || !isValidated {
		return false
	}

	tokens, err := s.storage.GetQuestradeTokens(ctx)
	if err != nil || tokens == nil {
		return false
	}

	return (tokens.AccessToken != "" && tokens.APIServer != "") || tokens.RefreshToken != ""
}

func (s *MarketDataService) providerDescriptors(ctx context.Context) []providerDescriptor {
	descriptors := make([]providerDescriptor, 0, 3)

	if s.hasQuestradeProvider(ctx) {
		descriptors = append(descriptors, providerDescriptor{
			name: "questrade",
			build: func() providers.Provider {
				return providers.NewQuestradeProvider(s.cfg.Questrade, s.storage, s.logClient)
			},
			supported: map[marketDataOperation]bool{
				operationSearch:      true,
				operationQuote:       true,
				operationIntraday:    true,
				operationOptionChain: true,
				operationProfile:     true,
			},
			priority: map[marketDataOperation]int{
				operationSearch:      1,
				operationQuote:       1,
				operationIntraday:    1,
				operationOptionChain: 1,
				operationProfile:     2,
			},
		})
	}

	if isValidated, err := s.storage.IsProviderValidated(ctx, "massive"); err == nil && isValidated {
		if apiKey, keyErr := s.providerAPIKey(ctx, "massive"); keyErr == nil && apiKey != "" {
			descriptors = append(descriptors, providerDescriptor{
				name:  "massive",
				build: func() providers.Provider { return s.newMassiveProvider() },
				supported: map[marketDataOperation]bool{
					operationSearch:   true,
					operationQuote:    true,
					operationIntraday: true,
				},
				priority: map[marketDataOperation]int{
					operationSearch:   2,
					operationQuote:    2,
					operationIntraday: 2,
				},
			})
		}
	}

	if isValidated, err := s.storage.IsProviderValidated(ctx, "fmp"); err == nil && isValidated {
		if apiKey, keyErr := s.providerAPIKey(ctx, "fmp"); keyErr == nil && apiKey != "" {
			descriptors = append(descriptors, providerDescriptor{
				name:  "fmp",
				build: func() providers.Provider { return s.newFMPProvider() },
				supported: map[marketDataOperation]bool{
					operationSearch:   true,
					operationQuote:    true,
					operationIntraday: true,
					operationProfile:  true,
					operationRatios:   true,
				},
				priority: map[marketDataOperation]int{
					operationSearch:   3,
					operationQuote:    3,
					operationIntraday: 3,
					operationProfile:  1,
					operationRatios:   1,
				},
			})
		}
	}

	return descriptors
}

func (s *MarketDataService) providersForOperation(ctx context.Context, operation marketDataOperation) []providers.Provider {
	descriptors := s.providerDescriptors(ctx)
	filtered := make([]activeProvider, 0, len(descriptors))

	for _, descriptor := range descriptors {
		if !descriptor.supports(operation) {
			continue
		}

		filtered = append(filtered, activeProvider{
			provider:   descriptor.build(),
			descriptor: descriptor,
		})
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].descriptor.priorityFor(operation) < filtered[j].descriptor.priorityFor(operation)
	})

	result := make([]providers.Provider, 0, len(filtered))
	for _, provider := range filtered {
		result = append(result, provider.provider)
	}

	return result
}

func (s *MarketDataService) providerForOperation(ctx context.Context, operation marketDataOperation, requestedSource string) providers.Provider {
	providersForOp := s.providersForOperation(ctx, operation)
	if requestedSource != "" {
		for _, provider := range providersForOp {
			if provider.Name() == strings.ToLower(strings.TrimSpace(requestedSource)) {
				return provider
			}
		}
	}

	if len(providersForOp) == 0 {
		return nil
	}

	return providersForOp[0]
}

func (s *MarketDataService) preferredProviderNameForOperation(ctx context.Context, operation marketDataOperation) string {
	provider := s.providerForOperation(ctx, operation, "")
	if provider == nil {
		return ""
	}

	return provider.Name()
}

func backfillOperation(dataType string) marketDataOperation {
	switch dataType {
	case "intraday_bars":
		return operationIntraday
	case "option_chain":
		return operationOptionChain
	default:
		return ""
	}
}

func mergeSearchResult(existing providers.TickerSearchResult, incoming providers.TickerSearchResult) providers.TickerSearchResult {
	if existing.Name == "" {
		existing.Name = incoming.Name
	}
	if existing.Exchange == "" {
		existing.Exchange = incoming.Exchange
	}
	if existing.Type == "" {
		existing.Type = incoming.Type
	}
	if existing.SymbolID == 0 {
		existing.SymbolID = incoming.SymbolID
	}
	if existing.Price == 0 && incoming.Price != 0 {
		existing.Price = incoming.Price
		existing.Change = incoming.Change
		existing.ChangePct = incoming.ChangePct
	}

	return existing
}

func storageSearchToProviderSearch(result storage.TickerSearchResult) providers.TickerSearchResult {
	return providers.TickerSearchResult{
		Symbol:    result.Symbol,
		Name:      result.Name,
		Exchange:  result.Exchange,
		Type:      result.Type,
		SymbolID:  result.SymbolID,
		Price:     result.Price,
		Change:    result.Change,
		ChangePct: result.ChangePct,
	}
}

func (s *MarketDataService) searchTickersComposite(ctx context.Context, query string) ([]providers.TickerSearchResult, []string) {
	merged := make(map[string]providers.TickerSearchResult)
	order := make([]string, 0, 20)

	dbResults, err := s.storage.SearchTickers(ctx, query)
	if err == nil {
		for _, result := range dbResults {
			converted := storageSearchToProviderSearch(result)
			if _, seen := merged[converted.Symbol]; !seen {
				order = append(order, converted.Symbol)
			}
			merged[converted.Symbol] = mergeSearchResult(merged[converted.Symbol], converted)
		}
	}

	providerErrors := make([]string, 0)
	for _, provider := range s.providersForOperation(ctx, operationSearch) {
		results, err := provider.SearchTickers(query)
		if err != nil {
			s.logClient.WarnWithMeta(ctx, "search failed for provider", map[string]interface{}{
				"provider": provider.Name(),
				"query":    query,
				"error":    err.Error(),
			})
			providerErrors = append(providerErrors, provider.Name()+": "+err.Error())
			continue
		}

		s.logClient.InfoWithMeta(ctx, "search succeeded for provider", map[string]interface{}{
			"provider": provider.Name(),
			"query":    query,
			"count":    len(results),
		})

		for _, result := range results {
			if result.Symbol == "" {
				continue
			}

			if _, seen := merged[result.Symbol]; !seen {
				order = append(order, result.Symbol)
			}
			merged[result.Symbol] = mergeSearchResult(merged[result.Symbol], result)

			_ = s.storage.UpsertTickerFromSearch(ctx, result.Symbol, result.Name, result.Exchange, result.Type)
		}
	}

	enriched := make([]providers.TickerSearchResult, 0, len(order))
	for _, symbol := range order {
		result := merged[symbol]
		if result.Symbol == "" {
			continue
		}

		if result.Price == 0 {
			for _, provider := range s.providersForOperation(ctx, operationQuote) {
				price, err := provider.FetchPrice(result.Symbol)
				if err != nil {
					continue
				}

				result.Price = price.Price
				result.Change = price.Change
				result.ChangePct = price.ChangePct
				break
			}
		}

		enriched = append(enriched, result)
	}

	return enriched, providerErrors
}

func newTickerDetailsResponse(symbol string, cached *storage.TickerDetails) *tickerDetailsResponse {
	response := &tickerDetailsResponse{Symbol: symbol}
	if cached == nil {
		return response
	}

	response.Symbol = cached.Symbol
	response.Name = cached.Name
	response.Exchange = cached.Exchange
	response.Price = cached.Price
	response.Change = cached.Change
	response.ChangePct = cached.ChangePct
	response.Volume = cached.Volume
	response.MarketCap = cached.MarketCap
	return response
}

func mergeProfileIntoTickerDetails(target *tickerDetailsResponse, profile *providers.CompanyProfile) {
	if profile == nil {
		return
	}

	if target.Symbol == "" {
		target.Symbol = profile.Symbol
	}
	if target.Name == "" {
		target.Name = profile.Name
	}
	if target.Exchange == "" {
		target.Exchange = profile.Exchange
	}
	if target.Sector == "" {
		target.Sector = profile.Sector
	}
	if target.Industry == "" {
		target.Industry = profile.Industry
	}
	if target.MarketCap == 0 {
		target.MarketCap = profile.MarketCap
	}
	if target.PeRatio == 0 {
		target.PeRatio = profile.PeRatio
	}
	if target.Eps == 0 {
		target.Eps = profile.Eps
	}
	if target.DividendYield == 0 {
		target.DividendYield = profile.DivYield
	}
	if target.Week52High == 0 {
		target.Week52High = profile.Week52High
	}
	if target.Week52Low == 0 {
		target.Week52Low = profile.Week52Low
	}
	if target.AvgVolume == 0 {
		target.AvgVolume = profile.AvgVolume
	}
}

func tickerDetailsNeedProfileFill(details *tickerDetailsResponse) bool {
	return details.Name == "" ||
		details.Exchange == "" ||
		details.Sector == "" ||
		details.Industry == "" ||
		details.MarketCap == 0 ||
		details.PeRatio == 0 ||
		details.Eps == 0 ||
		details.Week52High == 0 ||
		details.Week52Low == 0 ||
		details.AvgVolume == 0
}

func mergePriceIntoTickerDetails(target *tickerDetailsResponse, price *providers.Price) {
	if price == nil {
		return
	}

	target.Price = price.Price
	target.Change = price.Change
	target.ChangePct = price.ChangePct
	target.Volume = price.Volume
	if target.Symbol == "" {
		target.Symbol = price.Ticker
	}
}

func hasTickerDetailsData(details *tickerDetailsResponse) bool {
	if details == nil {
		return false
	}

	return details.Symbol != "" ||
		details.Name != "" ||
		details.Price != 0 ||
		details.Volume != 0 ||
		details.MarketCap != 0
}

func (s *MarketDataService) fetchTickerDetailsComposite(ctx context.Context, symbol string, cached *storage.TickerDetails, refreshQuote bool) (*tickerDetailsResponse, string) {
	details := newTickerDetailsResponse(symbol, cached)

	for _, provider := range s.providersForOperation(ctx, operationProfile) {
		profile, err := provider.FetchCompanyProfile(symbol)
		if err != nil {
			s.logClient.WarnWithMeta(ctx, "profile fetch failed", map[string]interface{}{
				"provider": provider.Name(),
				"symbol":   symbol,
				"error":    err.Error(),
			})
			continue
		}

		mergeProfileIntoTickerDetails(details, profile)
		if !tickerDetailsNeedProfileFill(details) {
			break
		}
	}

	quoteSource := ""
	if refreshQuote || cached == nil || cached.Price == 0 {
		for _, provider := range s.providersForOperation(ctx, operationQuote) {
			price, err := provider.FetchPrice(symbol)
			if err != nil {
				s.logClient.WarnWithMeta(ctx, "quote fetch failed", map[string]interface{}{
					"provider": provider.Name(),
					"symbol":   symbol,
					"error":    err.Error(),
				})
				continue
			}

			mergePriceIntoTickerDetails(details, price)
			quoteSource = price.Source
			break
		}
	}

	if !hasTickerDetailsData(details) {
		return nil, ""
	}

	if details.Symbol == "" {
		details.Symbol = symbol
	}

	return details, quoteSource
}

func providerRatiosToStorageRatios(ratios []*providers.FinancialRatio) []storage.FinancialRatioRecord {
	records := make([]storage.FinancialRatioRecord, 0, len(ratios))
	for _, ratio := range ratios {
		if ratio == nil {
			continue
		}

		records = append(records, storage.FinancialRatioRecord{
			Label:       ratio.Label,
			Value:       ratio.Value,
			Description: ratio.Description,
		})
	}

	return records
}

func (s *MarketDataService) fetchRatiosComposite(ctx context.Context, symbol string) ([]storage.FinancialRatioRecord, string) {
	for _, provider := range s.providersForOperation(ctx, operationRatios) {
		ratios, err := provider.FetchFinancialRatios(symbol)
		if err != nil {
			s.logClient.WarnWithMeta(ctx, "ratios fetch failed", map[string]interface{}{
				"provider": provider.Name(),
				"symbol":   symbol,
				"error":    err.Error(),
			})
			continue
		}

		records := providerRatiosToStorageRatios(ratios)
		if len(records) == 0 {
			continue
		}

		return records, provider.Name()
	}

	return nil, ""
}

func staleRatioTimestamp() time.Time {
	return time.Now().Add(-24 * time.Hour)
}
