package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/portfolio-sim/market-data-service/config"
	"github.com/portfolio-sim/market-data-service/database"
	loggingpkg "github.com/portfolio-sim/market-data-service/logging"
	"github.com/portfolio-sim/market-data-service/normalizer"
	"github.com/portfolio-sim/market-data-service/providers"
	"github.com/portfolio-sim/market-data-service/redis"
	"github.com/portfolio-sim/market-data-service/sse"
	"github.com/portfolio-sim/market-data-service/storage"
)

type MarketDataService struct {
	cfg             *config.Config
	db              *database.Postgres
	redis           *redis.Client
	storage         *storage.Storage
	logClient       *loggingpkg.Client
	sseMgr          *sse.Manager
	normalizer      *normalizer.Normalizer
	providers       []providers.Provider
	fetcherRegistry sync.Map
}

func NewMarketDataService(cfg *config.Config) *MarketDataService {
	logURL := os.Getenv("LOGGING_SERVICE_URL")
	if logURL == "" {
		logURL = "http://main-api:8080/api/logs"
	}
	logClient := loggingpkg.NewClient("market-data-service", logURL)

	db, err := database.NewPostgres(database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		MaxConns: cfg.Database.MaxConns,
	})
	if err != nil {
		logClient.Error(context.Background(), fmt.Sprintf("failed to connect to database: %v", err))
		os.Exit(1)
	}

	redisClient, err := redis.NewClient(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password)
	if err != nil {
		logClient.Error(context.Background(), fmt.Sprintf("failed to connect to redis: %v", err))
		os.Exit(1)
	}

	return &MarketDataService{
		cfg:        cfg,
		db:         db,
		redis:      redisClient,
		storage:    storage.NewStorage(db.Pool()),
		logClient:  logClient,
		sseMgr:     sse.NewManager(redisClient),
		normalizer: normalizer.NewNormalizer(),
	}
}

func (s *MarketDataService) Start() error {
	s.setupProviders()
	go s.handleBackfillQueue()
	go s.handleTickerSubscribe()
	go s.startHTTPServer()

	s.logClient.Info(context.Background(), "market data service starting")
	s.runPriceFetchers()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	s.logClient.Info(context.Background(), "shutting down")
	s.redis.Close()
	s.db.Close()
	return nil
}

func (s *MarketDataService) startHTTPServer() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	http.HandleFunc("/api/questrade/oauth/save", loggingMiddleware(s.logClient, s.handleSaveQuestradeOAuth))
	http.HandleFunc("/api/tickers/search", loggingMiddleware(s.logClient, s.handleSearchTickers))
	http.HandleFunc("/api/tickers/", loggingMiddleware(s.logClient, s.handleTickerRequest))

	addr := fmt.Sprintf(":%d", s.cfg.Server.HTTPPort)
	s.logClient.InfoWithMeta(context.Background(), "HTTP server starting", map[string]interface{}{"addr": addr})
	if err := http.ListenAndServe(addr, nil); err != nil {
		s.logClient.ErrorWithMeta(context.Background(), "HTTP server error", map[string]interface{}{"error": err.Error()})
	}
}

func loggingMiddleware(logClient *loggingpkg.Client, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := r.Context()
		reqBody, readErr := io.ReadAll(r.Body)
		if readErr != nil {
			reqBody = nil
		}
		r.Body = io.NopCloser(bytes.NewReader(reqBody))

		requestMeta := map[string]interface{}{
			"method":          r.Method,
			"path":            r.URL.Path,
			"query":           loggingpkg.SanitizeQuery(r.URL.RawQuery),
			"type":            "api_request",
			"request_headers": loggingpkg.RedactHeaders(r.Header),
			"remote_addr":     r.RemoteAddr,
			"content_length":  r.ContentLength,
		}
		if len(reqBody) > 0 {
			requestMeta["request_body"] = loggingpkg.SanitizeBody(r.Header.Get("Content-Type"), reqBody)
			requestMeta["request_body_size"] = len(reqBody)
		}
		if readErr != nil {
			requestMeta["request_body_error"] = readErr.Error()
		}
		logClient.InfoWithMeta(ctx, "API Request", requestMeta)

		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next(wrapper, r)

		duration := time.Since(start)
		meta := map[string]interface{}{
			"method":           r.Method,
			"path":             r.URL.Path,
			"query":            loggingpkg.SanitizeQuery(r.URL.RawQuery),
			"status":           wrapper.statusCode,
			"duration_ms":      duration.Milliseconds(),
			"type":             "api_response",
			"response_headers": loggingpkg.RedactHeaders(wrapper.Header()),
		}
		if len(wrapper.body) > 0 {
			meta["response_body"] = loggingpkg.SanitizeBody(wrapper.Header().Get("Content-Type"), wrapper.body)
			meta["response_body_size"] = len(wrapper.body)
		}

		if wrapper.statusCode >= 500 {
			logClient.ErrorWithMeta(ctx, "API Response Error", meta)
		} else if wrapper.statusCode >= 400 {
			logClient.WarnWithMeta(ctx, "API Response Warning", meta)
		} else {
			logClient.InfoWithMeta(ctx, "API Response", meta)
		}
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body = append(rw.body, b...)
	return rw.ResponseWriter.Write(b)
}

func (s *MarketDataService) handleSaveQuestradeOAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
		APIServer    string `json:"api_server"`
		AccessToken  string `json:"access_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		http.Error(w, "refresh_token is required", http.StatusBadRequest)
		return
	}

	_ = s.storage.UpdateQuestradeTokens(r.Context(), req.AccessToken, req.RefreshToken, req.APIServer, req.ExpiresIn)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "saved"})
}

func (s *MarketDataService) setupProviders() {
	s.providers = append(s.providers, providers.NewQuestradeProvider(s.cfg.Questrade, s.storage, s.logClient))
	if s.cfg.Polygon.APIKey != "" {
		s.providers = append(s.providers, providers.NewPolygonProvider(s.cfg.Polygon, s.logClient))
	}
	if s.cfg.FMP.APIKey != "" {
		s.providers = append(s.providers, providers.NewFMPProvider(s.cfg.FMP, s.logClient))
	}
}

func (s *MarketDataService) fetcherKey(ticker string, provider providers.Provider) string {
	return ticker + ":" + provider.Name()
}

func (s *MarketDataService) registerFetcher(ticker string, provider providers.Provider) bool {
	key := s.fetcherKey(ticker, provider)
	_, loaded := s.fetcherRegistry.LoadOrStore(key, true)
	return !loaded
}

func (s *MarketDataService) unregisterFetcher(ticker string, provider providers.Provider) {
	key := s.fetcherKey(ticker, provider)
	s.fetcherRegistry.Delete(key)
}

func (s *MarketDataService) runPriceFetchers() {
	tickerSymbols := s.storage.GetActiveTickers(context.Background())
	if len(tickerSymbols) == 0 {
		s.logClient.Info(context.Background(), "no tickers to fetch")
		return
	}
	for _, ticker := range tickerSymbols {
		for _, provider := range s.providers {
			if s.registerFetcher(ticker, provider) {
				go s.fetchPriceLoop(ticker, provider)
			}
		}
	}
}

func (s *MarketDataService) fetchPriceLoop(ticker string, provider providers.Provider) {
	if !s.registerFetcher(ticker, provider) {
		s.logClient.InfoWithMeta(context.Background(), "fetcher already running, skipping", map[string]interface{}{"ticker": ticker, "provider": provider.Name()})
		return
	}
	defer s.unregisterFetcher(ticker, provider)

	tickerID, err := s.storage.GetTickerID(context.Background(), ticker)
	if err != nil {
		s.logClient.Warn(context.Background(), fmt.Sprintf("ticker not found: %s", ticker))
		return
	}

	ticker = provider.Name() + ":" + ticker

	for {
		price, err := provider.FetchPrice(ticker)
		if err != nil {
			s.logClient.ErrorWithMeta(context.Background(), "failed to fetch price", map[string]interface{}{"ticker": ticker, "provider": provider.Name(), "error": err.Error()})
			time.Sleep(5 * time.Second)
			continue
		}

		normPrice, err := s.normalizer.NormalizePrice(price, tickerID)
		if err != nil {
			s.logClient.Error(context.Background(), fmt.Sprintf("failed to normalize price: %v", err))
			continue
		}

		if err := s.storage.UpsertNormalizedPrice(context.Background(), normPrice.TickerID, normPrice.Price, normPrice.Bid, normPrice.Ask, normPrice.Volume, normPrice.SourceID, normPrice.Timestamp); err != nil {
			s.logClient.ErrorWithMeta(context.Background(), "failed to store price", map[string]interface{}{"error": err.Error()})
		}

		s.sseMgr.PublishTick(ticker, normPrice)

		time.Sleep(time.Second)
	}
}

func (s *MarketDataService) handleTickerSubscribe() {
	for {
		result, err := s.redis.BRPop(context.Background(), 0, "queue:ticker:subscribe")
		if err != nil {
			continue
		}
		if len(result) < 2 {
			continue
		}

		var req struct {
			Symbol string `json:"symbol"`
			Action string `json:"action"`
		}
		if err := json.Unmarshal([]byte(result[1]), &req); err != nil {
			s.logClient.ErrorWithMeta(context.Background(), "failed to unmarshal ticker subscription", map[string]interface{}{"error": err.Error()})
			continue
		}

		tickerID, err := s.storage.GetTickerID(context.Background(), req.Symbol)
		if err != nil {
			s.logClient.Warn(context.Background(), fmt.Sprintf("ticker not found: %s", req.Symbol))
			continue
		}
		_ = tickerID // validated that ticker exists

		if req.Action == "subscribe" {
			for _, provider := range s.providers {
				if s.registerFetcher(req.Symbol, provider) {
					go s.fetchPriceLoop(req.Symbol, provider)
					s.logClient.InfoWithMeta(context.Background(), "started price fetcher for ticker", map[string]interface{}{"ticker": req.Symbol, "provider": provider.Name()})
				} else {
					s.logClient.InfoWithMeta(context.Background(), "fetcher already running, skipping", map[string]interface{}{"ticker": req.Symbol, "provider": provider.Name()})
				}
			}
		}
	}
}

func (s *MarketDataService) handleBackfillQueue() {
	for {
		result, err := s.redis.BRPop(context.Background(), 5, "queue:backfill")
		if err != nil {
			continue
		}
		if len(result) < 2 {
			continue
		}

		var req sse.BackfillRequest
		if err := json.Unmarshal([]byte(result[1]), &req); err != nil {
			s.logClient.ErrorWithMeta(context.Background(), "failed to unmarshal backfill request", map[string]interface{}{"error": err.Error()})
			continue
		}

		go s.processBackfill(&req)
	}
}

func (s *MarketDataService) processBackfill(req *sse.BackfillRequest) {
	s.logClient.InfoWithMeta(context.Background(), "processing backfill", map[string]interface{}{"ticker": req.Ticker, "data_type": req.DataType})

	tickerID, err := s.storage.GetTickerID(context.Background(), req.Ticker)
	if err != nil {
		s.logClient.ErrorWithMeta(context.Background(), "ticker not found", map[string]interface{}{"ticker": req.Ticker})
		return
	}

	var provider providers.Provider
	switch req.Source {
	case "polygon":
		provider = providers.NewPolygonProvider(s.cfg.Polygon, s.logClient)
	case "fmp":
		provider = providers.NewFMPProvider(s.cfg.FMP, s.logClient)
	default:
		provider = providers.NewQuestradeProvider(s.cfg.Questrade, s.storage, s.logClient)
	}

	switch req.DataType {
	case "intraday_bars":
		bars, err := provider.FetchIntradayBars(req.Ticker, req.Interval)
		if err != nil {
			s.logClient.ErrorWithMeta(context.Background(), "failed to fetch intraday bars", map[string]interface{}{"error": err.Error()})
			return
		}
		for _, bar := range bars {
			normBar, err := s.normalizer.NormalizeIntradayBar(bar, tickerID)
			if err != nil {
				continue
			}
			s.storage.InsertIntradayBar(context.Background(), normBar.TickerID, normBar.Interval, normBar.Open, normBar.High, normBar.Low, normBar.Close, normBar.Volume, normBar.Timestamp)
		}
	case "option_chain":
		chains, err := provider.FetchOptionChain(req.Ticker)
		if err != nil {
			s.logClient.ErrorWithMeta(context.Background(), "failed to fetch option chain", map[string]interface{}{"error": err.Error()})
			return
		}
		records := make([]*storage.OptionChainRecord, len(chains))
		for i, chain := range chains {
			normChain, err := s.normalizer.NormalizeOptionChain(chain, tickerID)
			if err != nil {
				continue
			}
			records[i] = &storage.OptionChainRecord{
				Expiration:   normChain.Expiration,
				Strike:       normChain.Strike,
				OptionType:   normChain.OptionType,
				Bid:          normChain.Bid,
				Ask:          normChain.Ask,
				Delta:        normChain.Delta,
				Gamma:        normChain.Gamma,
				Theta:        normChain.Theta,
				Vega:         normChain.Vega,
				ImpliedVol:   normChain.ImpliedVol,
				Volume:       normChain.Volume,
				OpenInterest: normChain.OpenInterest,
				Timestamp:    normChain.Timestamp,
			}
		}
		s.storage.InsertOptionChain(context.Background(), tickerID, req.Source, records)
	}

	s.logClient.InfoWithMeta(context.Background(), "backfill complete", map[string]interface{}{"ticker": req.Ticker})
}

func (s *MarketDataService) handleSearchTickers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "q parameter required", http.StatusBadRequest)
		return
	}

	dbResults, err := s.storage.SearchTickers(r.Context(), query)
	if err == nil && len(dbResults) > 0 {
		s.logClient.InfoWithMeta(r.Context(), "search returning DB results", map[string]interface{}{
			"query": query,
			"count": len(dbResults),
		})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(dbResults)
		return
	}

	var allResults []providers.TickerSearchResult
	var providerErrors []string
	successCount := 0
	for _, provider := range s.providers {
		results, err := provider.SearchTickers(query)
		if err != nil {
			s.logClient.WarnWithMeta(context.Background(), "search failed for provider", map[string]interface{}{"provider": provider.Name(), "error": err.Error()})
			providerErrors = append(providerErrors, fmt.Sprintf("%s: %v", provider.Name(), err))
			continue
		}
		successCount++
		s.logClient.InfoWithMeta(context.Background(), "search succeeded for provider", map[string]interface{}{
			"provider": provider.Name(),
			"query":    query,
			"count":    len(results),
		})
		allResults = append(allResults, results...)

		for _, result := range results {
			s.storage.UpsertTickerFromSearch(r.Context(), result.Symbol, result.Name, result.Exchange, result.Type)
		}
	}

	if successCount == 0 && len(providerErrors) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{"error": strings.Join(providerErrors, "; ")})
		return
	}

	if allResults == nil {
		allResults = []providers.TickerSearchResult{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allResults)
}

func (s *MarketDataService) handleTickerRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case strings.HasSuffix(path, "/details"):
		s.handleTickerDetails(w, r)
	case strings.HasSuffix(path, "/intraday"):
		s.handleIntradayBars(w, r)
	case strings.HasSuffix(path, "/ratios"):
		s.handleFinancialRatios(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *MarketDataService) handleTickerDetails(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasSuffix(path, "/details") {
		http.NotFound(w, r)
		return
	}
	symbol := strings.TrimSuffix(strings.TrimPrefix(path, "/api/tickers/"), "/details")
	if symbol == "" {
		http.Error(w, "symbol required", http.StatusBadRequest)
		return
	}
	details, err := s.storage.GetTickerDetails(r.Context(), symbol)
	if err == nil && details != nil && details.Price > 0 {
		stale, _ := s.storage.IsTickerDataStale(r.Context(), symbol, 24*time.Hour)
		if !stale {
			s.logClient.InfoWithMeta(r.Context(), "returning cached ticker details", map[string]interface{}{"symbol": symbol})
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(details)
			return
		}
		s.logClient.InfoWithMeta(r.Context(), "ticker data stale, refreshing from provider", map[string]interface{}{"symbol": symbol})
	}
	newDetails := s.fetchTickerDetailsFromProvider(symbol)
	if newDetails == nil {
		if details != nil && details.Price > 0 {
			s.logClient.WarnWithMeta(r.Context(), "provider fetch failed, returning stale data", map[string]interface{}{"symbol": symbol})
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(details)
			return
		}
		s.logClient.ErrorWithMeta(r.Context(), "get ticker details failed", map[string]interface{}{"symbol": symbol})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "ticker not found"})
		return
	}
	s.storage.UpdateTickerPrice(r.Context(), symbol, newDetails.Price, newDetails.Change, newDetails.ChangePct, newDetails.Volume, newDetails.MarketCap)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newDetails)
}

func (s *MarketDataService) fetchTickerDetailsFromProvider(symbol string) *storage.TickerDetails {
	for _, provider := range s.providers {
		profile, err := provider.FetchCompanyProfile(symbol)
		if err != nil {
			s.logClient.ErrorWithMeta(context.Background(), "FetchCompanyProfile failed", map[string]interface{}{"provider": provider.Name(), "symbol": symbol, "error": err.Error()})
			continue
		}
		var price float64
		var volume int64

		priceData, err := provider.FetchPrice(symbol)
		if err == nil {
			price = priceData.Price
			volume = priceData.Volume
		}

		return &storage.TickerDetails{
			Symbol:    profile.Symbol,
			Name:      profile.Name,
			Exchange:  profile.Exchange,
			Price:     price,
			Volume:    volume,
			MarketCap: profile.MarketCap,
		}
	}
	return nil
}

func (s *MarketDataService) handleIntradayBars(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasSuffix(path, "/intraday") {
		http.NotFound(w, r)
		return
	}
	symbol := strings.TrimSuffix(strings.TrimPrefix(path, "/api/tickers/"), "/intraday")
	if symbol == "" {
		http.Error(w, "symbol required", http.StatusBadRequest)
		return
	}
	bars, err := s.storage.GetIntradayBars(r.Context(), symbol, 100)
	if err != nil {
		s.logClient.ErrorWithMeta(r.Context(), "get intraday bars failed", map[string]interface{}{"symbol": symbol, "error": err.Error()})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	if len(bars) == 0 {
		s.logClient.InfoWithMeta(r.Context(), "no intraday data for ticker, triggering backfill", map[string]interface{}{"symbol": symbol})
		s.triggerBackfill(symbol, "intraday_bars", "1min")
		s.triggerTickerSubscribe(symbol)
		bars = []storage.IntradayBarRecord{}
	} else {
		oldest := bars[len(bars)-1].Timestamp
		if time.Since(oldest) > 5*time.Minute {
			s.logClient.InfoWithMeta(r.Context(), "intraday data stale for ticker, triggering refresh", map[string]interface{}{"symbol": symbol})
			s.triggerBackfill(symbol, "intraday_bars", "1min")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bars)
}

func (s *MarketDataService) triggerBackfill(ticker, dataType, interval string) {
	req := sse.BackfillRequest{
		Ticker:   ticker,
		DataType: dataType,
		Interval: interval,
		Source:   "polygon",
	}
	data, _ := json.Marshal(req)
	s.redis.LPush(context.Background(), "queue:backfill", string(data))
}

func (s *MarketDataService) triggerTickerSubscribe(symbol string) {
	req := map[string]string{"symbol": symbol, "action": "subscribe"}
	data, _ := json.Marshal(req)
	s.redis.LPush(context.Background(), "queue:ticker:subscribe", string(data))
}

func (s *MarketDataService) handleFinancialRatios(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasSuffix(path, "/ratios") {
		http.NotFound(w, r)
		return
	}
	symbol := strings.TrimSuffix(strings.TrimPrefix(path, "/api/tickers/"), "/ratios")
	if symbol == "" {
		http.Error(w, "symbol required", http.StatusBadRequest)
		return
	}
	ratios, err := s.storage.GetFinancialRatios(r.Context(), symbol)
	if err != nil || ratios == nil {
		ratios = []storage.FinancialRatioRecord{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ratios)
}

func main() {
	cfg := config.Load()
	service := NewMarketDataService(cfg)
	if err := service.Start(); err != nil {
		os.Exit(1)
	}
}
