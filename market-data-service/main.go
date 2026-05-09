package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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
	cfg            *config.Config
	db             *database.Postgres
	redis          *redis.Client
	logger         *slog.Logger
	storage        *storage.Storage
	logClient      *loggingpkg.Client
	sseMgr         *sse.Manager
	normalizer     *normalizer.Normalizer
	providers      []providers.Provider
	fetcherRegistry sync.Map
}

func NewMarketDataService(cfg *config.Config) *MarketDataService {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	db, err := database.NewPostgres(database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		MaxConns: cfg.Database.MaxConns,
	})
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	redisClient, err := redis.NewClient(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password)
	if err != nil {
		logger.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}

	logURL := os.Getenv("LOGGING_SERVICE_URL")
	if logURL == "" {
		logURL = "http://backend:8080/api/logs"
	}
	return &MarketDataService{
		cfg:       cfg,
		db:        db,
		redis:     redisClient,
		logger:    logger,
		storage:   storage.NewStorage(db.Pool()),
		logClient: loggingpkg.NewClient("market-data-service", logURL),
		sseMgr:    sse.NewManager(redisClient),
		normalizer: normalizer.NewNormalizer(),
	}
}

func (s *MarketDataService) Start() error {
	s.setupProviders()
	go s.handleBackfillQueue()
	go s.handleTickerSubscribe()
	go s.startHTTPServer()

	s.logger.Info("market data service starting")
	s.runPriceFetchers()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	s.logger.Info("shutting down")
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
	http.HandleFunc("/api/tickers/", loggingMiddleware(s.logClient, s.handleTickerDetails))
	http.HandleFunc("/api/tickers/", loggingMiddleware(s.logClient, s.handleIntradayBars))
	http.HandleFunc("/api/tickers/", loggingMiddleware(s.logClient, s.handleFinancialRatios))

	addr := fmt.Sprintf(":%d", s.cfg.Server.HTTPPort)
	s.logger.Info("HTTP server starting", "addr", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		s.logger.Error("HTTP server error", "error", err)
	}
}

func loggingMiddleware(logClient *loggingpkg.Client, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := r.Context()

		logClient.InfoWithMeta(ctx, fmt.Sprintf("API Request: %s %s", r.Method, r.URL.Path), map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
			"query":  r.URL.RawQuery,
		})

		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next(wrapper, r)

		duration := time.Since(start)
		logClient.InfoWithMeta(ctx, fmt.Sprintf("API Response: %s %s %d", r.Method, r.URL.Path, wrapper.statusCode), map[string]interface{}{
			"method":   r.Method,
			"path":     r.URL.Path,
			"status":   wrapper.statusCode,
			"duration": duration.Milliseconds(),
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
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
	s.providers = append(s.providers, providers.NewQuestradeProvider(s.cfg.Questrade, s.storage))
	if s.cfg.Polygon.APIKey != "" {
		s.providers = append(s.providers, providers.NewPolygonProvider(s.cfg.Polygon))
	}
	if s.cfg.FMP.APIKey != "" {
		s.providers = append(s.providers, providers.NewFMPProvider(s.cfg.FMP))
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
		s.logger.Info("no tickers to fetch")
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
		s.logger.Info("fetcher already running, skipping", "ticker", ticker, "provider", provider.Name())
		return
	}
	defer s.unregisterFetcher(ticker, provider)

	tickerID, err := s.storage.GetTickerID(context.Background(), ticker)
	if err != nil {
		s.logger.Warn("ticker not found", "ticker", ticker)
		return
	}

	ticker = provider.Name() + ":" + ticker

	for {
		price, err := provider.FetchPrice(ticker)
		if err != nil {
			s.logger.Error("failed to fetch price", "ticker", ticker, "provider", provider.Name(), "error", err)
			time.Sleep(5 * time.Second)
			continue
		}

		normPrice, err := s.normalizer.NormalizePrice(price, tickerID)
		if err != nil {
			s.logger.Error("failed to normalize price", "error", err)
			continue
		}

		if err := s.storage.UpsertNormalizedPrice(context.Background(), normPrice.TickerID, normPrice.Price, normPrice.Bid, normPrice.Ask, normPrice.Volume, normPrice.SourceID, normPrice.Timestamp); err != nil {
			s.logger.Error("failed to store price", "error", err)
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
			s.logger.Error("failed to unmarshal ticker subscription", "error", err)
			continue
		}

		tickerID, err := s.storage.GetTickerID(context.Background(), req.Symbol)
		if err != nil {
			s.logger.Warn("ticker not found", "ticker", req.Symbol)
			continue
		}
		_ = tickerID // validated that ticker exists

		if req.Action == "subscribe" {
			for _, provider := range s.providers {
				if s.registerFetcher(req.Symbol, provider) {
					go s.fetchPriceLoop(req.Symbol, provider)
					s.logger.Info("started price fetcher for ticker", "ticker", req.Symbol, "provider", provider.Name())
				} else {
					s.logger.Info("fetcher already running, skipping", "ticker", req.Symbol, "provider", provider.Name())
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
			s.logger.Error("failed to unmarshal backfill request", "error", err)
			continue
		}

		go s.processBackfill(&req)
	}
}

func (s *MarketDataService) processBackfill(req *sse.BackfillRequest) {
	s.logger.Info("processing backfill", "ticker", req.Ticker, "data_type", req.DataType)

	tickerID, err := s.storage.GetTickerID(context.Background(), req.Ticker)
	if err != nil {
		s.logger.Error("ticker not found", "ticker", req.Ticker)
		return
	}

	var provider providers.Provider
	switch req.Source {
	case "polygon":
		provider = providers.NewPolygonProvider(s.cfg.Polygon)
	case "fmp":
		provider = providers.NewFMPProvider(s.cfg.FMP)
	default:
		provider = providers.NewQuestradeProvider(s.cfg.Questrade, s.storage)
	}

	switch req.DataType {
	case "intraday_bars":
		bars, err := provider.FetchIntradayBars(req.Ticker, req.Interval)
		if err != nil {
			s.logger.Error("failed to fetch intraday bars", "error", err)
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
			s.logger.Error("failed to fetch option chain", "error", err)
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

	s.logger.Info("backfill complete", "ticker", req.Ticker)
}

func (s *MarketDataService) handleSearchTickers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "q parameter required", http.StatusBadRequest)
		return
	}
	results, err := s.storage.SearchTickers(r.Context(), query)
	if err != nil {
		s.logger.Error("search tickers failed", "error", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	if results == nil {
		results = []storage.TickerSearchResult{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
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
	if err != nil {
		s.logger.Error("get ticker details failed", "symbol", symbol, "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "ticker not found"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details)
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
		s.logger.Error("get intraday bars failed", "symbol", symbol, "error", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	if len(bars) == 0 {
		s.logger.Info("no intraday data for ticker, triggering backfill", "symbol", symbol)
		s.triggerBackfill(symbol, "intraday_bars", "1min")
		s.triggerTickerSubscribe(symbol)
		bars = []storage.IntradayBarRecord{}
	} else {
		oldest := bars[len(bars)-1].Timestamp
		if time.Since(oldest) > 5*time.Minute {
			s.logger.Info("intraday data stale for ticker, triggering refresh", "symbol", symbol)
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
