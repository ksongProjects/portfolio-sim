package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/portfolio-sim/backend/services/market-data-service/config"
	"github.com/portfolio-sim/backend/services/market-data-service/database"
	loggingpkg "github.com/portfolio-sim/backend/services/market-data-service/logging"
	"github.com/portfolio-sim/backend/services/market-data-service/normalizer"
	"github.com/portfolio-sim/backend/services/market-data-service/providers"
	"github.com/portfolio-sim/backend/services/market-data-service/redis"
	"github.com/portfolio-sim/backend/services/market-data-service/sse"
	"github.com/portfolio-sim/backend/services/market-data-service/storage"
)

type MarketDataService struct {
	cfg       *config.Config
	db        *database.Postgres
	redis     *redis.Client
	logger    *slog.Logger
	storage   *storage.Storage
	logClient *loggingpkg.Client
	sseMgr    *sse.Manager
	normalizer *normalizer.Normalizer
	providers []providers.Provider
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

	return &MarketDataService{
		cfg:       cfg,
		db:        db,
		redis:     redisClient,
		logger:    logger,
		storage:   storage.NewStorage(db.Pool()),
		logClient: loggingpkg.NewClient("market-data-service"),
		sseMgr:    sse.NewManager(redisClient),
		normalizer: normalizer.NewNormalizer(),
	}
}

func (s *MarketDataService) Start() error {
	s.setupProviders()
	go s.handleBackfillQueue()

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

func (s *MarketDataService) setupProviders() {
	if s.cfg.Questrade.APIKey != "" {
		s.providers = append(s.providers, providers.NewQuestradeProvider(s.cfg.Questrade))
	}
	if s.cfg.Polygon.APIKey != "" {
		s.providers = append(s.providers, providers.NewPolygonProvider(s.cfg.Polygon))
	}
	if s.cfg.FMP.APIKey != "" {
		s.providers = append(s.providers, providers.NewFMPProvider(s.cfg.FMP))
	}
}

func (s *MarketDataService) runPriceFetchers() {
	tickers := []string{"AAPL", "MSFT", "GOOGL", "AMZN", "TSLA"}

	for _, ticker := range tickers {
		for _, provider := range s.providers {
			go s.fetchPriceLoop(ticker, provider)
		}
	}
}

func (s *MarketDataService) fetchPriceLoop(ticker string, provider providers.Provider) {
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
		provider = providers.NewQuestradeProvider(s.cfg.Questrade)
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

func main() {
	cfg := config.Load()
	service := NewMarketDataService(cfg)
	if err := service.Start(); err != nil {
		os.Exit(1)
	}
}
