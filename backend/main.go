package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/portfolio-sim/backend/database"
	"github.com/portfolio-sim/backend/logging"
	"github.com/portfolio-sim/backend/redis"
	"github.com/portfolio-sim/backend/services"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
}

type ServerConfig struct {
	HTTPPort       int
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	ShutdownTime   time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	MaxConns int
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			HTTPPort:     getEnvInt("SERVER_HTTP_PORT", 8080),
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			ShutdownTime: getEnvDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DATABASE_HOST", "localhost"),
			Port:     getEnvInt("DATABASE_PORT", 5432),
			User:     getEnv("DATABASE_USER", "postgres"),
			Password: getEnv("DATABASE_PASSWORD", "postgres"),
			DBName:   getEnv("DATABASE_NAME", "portfolio_sim"),
			MaxConns: getEnvInt("DATABASE_MAX_CONNS", 20),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
		},
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if dur, err := time.ParseDuration(val); err == nil {
			return dur
		}
	}
	return defaultVal
}

type Server struct {
	cfg          *Config
	logger       *slog.Logger
	middleware   *logging.Middleware
	db           *database.Postgres
	redisClient  *redis.Client
	obsService   *services.ObservabilityService
	portfolioSvc *services.PortfolioService
	providerSvc  *services.ProviderService
	tickerSvc    *services.TickerService
}

func NewServer(cfg *Config) (*Server, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	db, err := database.NewPostgresWithLogger(database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		MaxConns: cfg.Database.MaxConns,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	logURL := os.Getenv("LOGGING_SERVICE_URL")
	if logURL == "" {
		logURL = "http://backend:8080/api/logs"
	}
	loggingMiddleware := logging.NewMiddleware("main-api", logURL, logger)

	redisClient, err := redis.NewClient(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password)
	if err != nil {
		return nil, fmt.Errorf("connect redis: %w", err)
	}

	return &Server{
		cfg:          cfg,
		logger:       logger,
		middleware:   loggingMiddleware,
		db:           db,
		redisClient:  redisClient,
		obsService:   services.NewObservabilityService(),
		portfolioSvc: services.NewPortfolioService(),
		providerSvc:  services.NewProviderService(logger),
		tickerSvc:    services.NewTickerService(os.Getenv("MARKET_DATA_SERVICE_URL")),
	}, nil
}

func (s *Server) Start() error {
	corsMiddleware := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			h.ServeHTTP(w, r)
		})
	}

	http.HandleFunc("GET /health", s.handleHealth)
	http.HandleFunc("POST /api/logs", s.handleIngestLog)
	http.HandleFunc("GET /api/observability/services", s.middleware.WrapHandlerFunc(s.handleGetServices))
	http.HandleFunc("GET /api/observability/logs", s.middleware.WrapHandlerFunc(s.handleGetLogs))
	http.HandleFunc("GET /api/portfolio/positions", s.middleware.WrapHandlerFunc(s.handleGetPositions))
	http.HandleFunc("GET /api/portfolio/summary", s.middleware.WrapHandlerFunc(s.handleGetPortfolioSummary))
	http.HandleFunc("GET /api/market/indices", s.middleware.WrapHandlerFunc(s.handleGetMarketIndices))
	http.HandleFunc("GET /api/news", s.middleware.WrapHandlerFunc(s.handleGetNews))
	http.HandleFunc("GET /api/strategies", s.middleware.WrapHandlerFunc(s.handleGetStrategies))
	http.HandleFunc("GET /api/signals", s.middleware.WrapHandlerFunc(s.handleGetSignals))
	http.HandleFunc("GET /api/providers", s.middleware.WrapHandlerFunc(s.handleGetProviders))
	http.HandleFunc("PUT /api/providers", s.middleware.WrapHandlerFunc(s.handleUpdateProvider))
	http.HandleFunc("POST /api/providers/validate", s.middleware.WrapHandlerFunc(s.handleValidateProvider))
	http.HandleFunc("PUT /api/providers/questrade/oauth", s.middleware.WrapHandlerFunc(s.handleSaveQuestradeOAuth))
	http.HandleFunc("GET /api/providers/questrade/oauth", s.middleware.WrapHandlerFunc(s.handleGetQuestradeOAuth))
	http.HandleFunc("GET /api/connections", s.middleware.WrapHandlerFunc(s.handleGetConnections))
	http.HandleFunc("GET /api/rss-feeds", s.middleware.WrapHandlerFunc(s.handleGetRSSFeeds))
	http.HandleFunc("POST /api/rss-feeds", s.middleware.WrapHandlerFunc(s.handleAddRSSFeed))
	http.HandleFunc("DELETE /api/rss-feeds", s.middleware.WrapHandlerFunc(s.handleDeleteRSSFeed))
	http.HandleFunc("GET /api/tickers/search", s.middleware.WrapHandlerFunc(s.handleSearchTickers))
	http.HandleFunc("GET /api/tickers/", s.middleware.WrapHandlerFunc(s.handleGetTickerDetails))
	http.HandleFunc("GET /api/stream/market", s.handleMarketStream)

	s.logger.Info("main api server starting", "port", s.cfg.Server.HTTPPort)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.cfg.Server.HTTPPort),
		Handler:      corsMiddleware(http.DefaultServeMux),
		ReadTimeout:  s.cfg.Server.ReadTimeout,
		WriteTimeout: s.cfg.Server.WriteTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("http server error", "error", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	s.logger.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Server.ShutdownTime)
	defer cancel()
	srv.Shutdown(ctx)
	s.db.Close()
	return nil
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleGetServices(w http.ResponseWriter, r *http.Request) {
	services := s.obsService.CheckServices(r.Context())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

func (s *Server) handleIngestLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var entry struct {
		ID        string                 `json:"id"`
		Timestamp string                 `json:"timestamp"`
		Level     string                 `json:"level"`
		Service   string                 `json:"service"`
		Component string                 `json:"component,omitempty"`
		Message   string                 `json:"message"`
		Metadata  map[string]interface{} `json:"metadata,omitempty"`
		TraceID   string                 `json:"trace_id,omitempty"`
		SpanID    string                 `json:"span_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if entry.Level == "" || entry.Service == "" || entry.Message == "" {
		http.Error(w, "level, service, and message are required", http.StatusBadRequest)
		return
	}

	metadataJSON, _ := json.Marshal(entry.Metadata)
	_, err := s.db.Exec(r.Context(), `
		INSERT INTO logs (id, timestamp, level, service, component, message, metadata, trace_id, span_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, entry.ID, entry.Timestamp, entry.Level, entry.Service, entry.Component, entry.Message, metadataJSON, entry.TraceID, entry.SpanID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 500 {
			limit = l
		}
	}
	level := r.URL.Query().Get("level")
	service := r.URL.Query().Get("service")

	query := `
		SELECT id, timestamp, level, service, component, message, metadata, trace_id, span_id
		FROM logs
		WHERE ($1 = '' OR level = $1)
		  AND ($2 = '' OR service = $2)
		ORDER BY timestamp DESC
		LIMIT $3
	`
	rows, err := s.db.Query(r.Context(), query, level, service, limit)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	defer rows.Close()

	type LogEntry struct {
		ID        string                 `json:"id"`
		Timestamp string                 `json:"timestamp"`
		Level     string                 `json:"level"`
		Service   string                 `json:"service"`
		Component string                 `json:"component,omitempty"`
		Message   string                 `json:"message"`
		Metadata  map[string]interface{} `json:"metadata,omitempty"`
		TraceID   string                 `json:"trace_id,omitempty"`
		SpanID    string                 `json:"span_id,omitempty"`
	}

	logs := []LogEntry{}
	for rows.Next() {
		var le LogEntry
		var metadata []byte
		if err := rows.Scan(&le.ID, &le.Timestamp, &le.Level, &le.Service, &le.Component, &le.Message, &metadata, &le.TraceID, &le.SpanID); err != nil {
			continue
		}
		if metadata != nil {
			json.Unmarshal(metadata, &le.Metadata)
		}
		logs = append(logs, le)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func (s *Server) handleMarketStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	symbols := r.URL.Query()["symbols"]
	streamStreams := make([]string, 0)
	if len(symbols) > 0 {
		for _, sym := range symbols {
			streamStreams = append(streamStreams, fmt.Sprintf("stream:market:ticks:%s", sym))
		}
	} else {
		streamStreams = []string{"stream:market:ticks:*"}
	}

	lastIDs := make(map[string]string)
	for _, stream := range streamStreams {
		lastIDs[stream] = "$"
	}

	s.logger.Info("SSE market stream connected", "symbols", symbols)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("SSE market stream disconnected")
			return
		default:
		}

		streams := make([]string, 0, len(streamStreams)*2)
		for stream := range lastIDs {
			streams = append(streams, stream, lastIDs[stream])
		}

		results, err := s.redisClient.XRead(ctx, streams, 10)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		for _, result := range results {
			for _, msg := range result.Messages {
				lastIDs[result.Stream] = msg.ID
				if data, ok := msg.Values["data"].(string); ok {
					fmt.Fprintf(w, "data: %s\n\n", data)
					flusher.Flush()
				}
			}
		}
	}
}

func (s *Server) handleGetPositions(w http.ResponseWriter, r *http.Request) {
	portfolioID := r.URL.Query().Get("portfolio_id")
	if portfolioID == "" || portfolioID == "default" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]services.Position{})
		return
	}
	positions, err := s.portfolioSvc.GetPositions(r.Context(), s.db, portfolioID)
	if err != nil {
		positions = []services.Position{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(positions)
}

func (s *Server) handleGetPortfolioSummary(w http.ResponseWriter, r *http.Request) {
	portfolioID := r.URL.Query().Get("portfolio_id")
	if portfolioID == "" || portfolioID == "default" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&services.PortfolioSummary{})
		return
	}
	summary, err := s.portfolioSvc.GetPortfolioSummary(r.Context(), s.db, portfolioID)
	if err != nil || summary == nil {
		summary = &services.PortfolioSummary{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (s *Server) handleGetMarketIndices(w http.ResponseWriter, r *http.Request) {
	indices, _ := s.portfolioSvc.GetMarketIndices(r.Context(), s.db)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(indices)
}

func (s *Server) handleGetNews(w http.ResponseWriter, r *http.Request) {
	limit := r.URL.Query().Get("limit")
	l := 20
	if limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil && parsed > 0 {
			l = parsed
		}
	}
	articles, _ := s.portfolioSvc.GetNewsArticles(r.Context(), s.db, l)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(articles)
}

func (s *Server) handleGetStrategies(w http.ResponseWriter, r *http.Request) {
	strategies, _ := s.portfolioSvc.GetStrategies(r.Context(), s.db)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(strategies)
}

func (s *Server) handleGetSignals(w http.ResponseWriter, r *http.Request) {
	limit := r.URL.Query().Get("limit")
	l := 10
	if limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil && parsed > 0 {
			l = parsed
		}
	}
	signals, _ := s.portfolioSvc.GetSignals(r.Context(), s.db, l)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(signals)
}

func (s *Server) handleGetProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := s.providerSvc.GetProviders(r.Context(), s.db)
	if err != nil || providers == nil || len(providers) == 0 {
		providers = []services.ProviderConfig{
			{ID: "polygon", ProviderID: "polygon", Name: "Polygon.io", Description: "Real-time and historical market data", Type: "market_data", RateLimit: 60, DocURL: "https://polygon.io/docs", TokenExpired: false},
			{ID: "questrade", ProviderID: "questrade", Name: "Questrade", Description: "Questrade market data API", Type: "market_data", RateLimit: 100, DocURL: "https://www.questrade.com/api", TokenExpired: false},
			{ID: "fmp", ProviderID: "fmp", Name: "Financial Modeling Prep", Description: "Financial statements and fundamental data", Type: "market_data", RateLimit: 250, DocURL: "https://site.financialmodelingprep.com/developers/docs", TokenExpired: false},
			{ID: "youtube", ProviderID: "youtube", Name: "YouTube Data API", Description: "YouTube Data API for video transcripts", Type: "youtube", RateLimit: 0, DocURL: "https://developers.google.com/youtube/v3", TokenExpired: false},
			{ID: "gemini", ProviderID: "gemini", Name: "Google Gemini", Description: "Gemini API for content summarization", Type: "gemini", RateLimit: 0, DocURL: "https://ai.google.dev/docs", TokenExpired: false},
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(providers)
}

func (s *Server) handleUpdateProvider(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ProviderID string `json:"provider_id"`
		APIKey     string `json:"api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.ProviderID == "" || req.APIKey == "" {
		http.Error(w, "provider_id and api_key are required", http.StatusBadRequest)
		return
	}
	if err := s.providerSvc.SaveProviderKey(r.Context(), s.db, req.ProviderID, req.APIKey); err != nil {
		http.Error(w, "failed to save", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "saved"})
}

func (s *Server) handleGetConnections(w http.ResponseWriter, r *http.Request) {
	statuses, _ := s.providerSvc.CheckConnection(r.Context(), s.db)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statuses)
}

func (s *Server) handleValidateProvider(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ProviderID string `json:"provider_id"`
		APIKey     string `json:"api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.ProviderID == "" || req.APIKey == "" {
		http.Error(w, "provider_id and api_key are required", http.StatusBadRequest)
		return
	}

	s.logger.Info("validating provider key", "provider", req.ProviderID, "has_key", len(req.APIKey) > 0)
	s.middleware.LogInfo(r.Context(), "validating provider key", map[string]interface{}{"provider": req.ProviderID, "has_key": len(req.APIKey) > 0})

	valid, err := s.providerSvc.ValidateProviderKey(r.Context(), s.db, req.ProviderID, req.APIKey)
	if err != nil {
		s.logger.Error("provider validation failed", "provider", req.ProviderID, "error", err)
		s.middleware.LogError(r.Context(), "provider validation failed", map[string]interface{}{"provider": req.ProviderID, "error": err.Error()})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"valid": false, "error": err.Error()})
		return
	}

	s.logger.Info("provider validation succeeded", "provider", req.ProviderID, "valid", valid)
	s.middleware.LogInfo(r.Context(), "provider validation succeeded", map[string]interface{}{"provider": req.ProviderID, "valid": valid})

	if valid && req.ProviderID == "questrade" {
		s.logger.Info("exchanging questrade OAuth token", "provider", req.ProviderID)
		s.middleware.LogInfo(r.Context(), "exchanging questrade OAuth token", map[string]interface{}{"provider": req.ProviderID})
		accessToken, refreshToken, apiServer, expiresIn, err := s.providerSvc.ExchangeQuestradeToken(req.APIKey)
		if err == nil && refreshToken != "" {
			s.logger.Info("saving questrade OAuth tokens", "provider", req.ProviderID, "expires_in", expiresIn)
			s.middleware.LogInfo(r.Context(), "saving questrade OAuth tokens", map[string]interface{}{"provider": req.ProviderID, "expires_in": expiresIn})
			s.providerSvc.SaveQuestradeOAuth(r.Context(), s.db, req.ProviderID, accessToken, refreshToken, apiServer, expiresIn)
		} else if err != nil {
			s.logger.Error("questrade token exchange failed", "provider", req.ProviderID, "error", err)
			s.middleware.LogError(r.Context(), "questrade token exchange failed", map[string]interface{}{"provider": req.ProviderID, "error": err.Error()})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"valid": valid})
}

func (s *Server) handleGetRSSFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := s.providerSvc.GetRSSFeeds(r.Context(), s.db)
	if err != nil {
		feeds = []services.RSSFeed{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feeds)
}

func (s *Server) handleAddRSSFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.URL == "" {
		http.Error(w, "name and url are required", http.StatusBadRequest)
		return
	}
	if err := s.providerSvc.AddRSSFeed(r.Context(), s.db, req.Name, req.URL); err != nil {
		http.Error(w, "failed to add feed", http.StatusInternalServerError)
		return
	}
	feeds, _ := s.providerSvc.GetRSSFeeds(r.Context(), s.db)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feeds)
}

func (s *Server) handleDeleteRSSFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	feedID := r.URL.Query().Get("id")
	if feedID == "" {
		http.Error(w, "feed id is required", http.StatusBadRequest)
		return
	}
	if err := s.providerSvc.DeleteRSSFeed(r.Context(), s.db, feedID); err != nil {
		http.Error(w, "failed to delete feed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (s *Server) handleGetQuestradeOAuth(w http.ResponseWriter, r *http.Request) {
	accessToken, refreshToken, apiServer := s.providerSvc.GetQuestradeOAuth(r.Context(), s.db, "questrade")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"api_server":    apiServer,
	})
}

func (s *Server) handleSaveQuestradeOAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		APIServer    string `json:"api_server"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.RefreshToken == "" {
		http.Error(w, "refresh_token is required", http.StatusBadRequest)
		return
	}
	if err := s.providerSvc.SaveQuestradeOAuth(r.Context(), s.db, "questrade", req.AccessToken, req.RefreshToken, req.APIServer, 3600); err != nil {
		http.Error(w, "failed to save", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "saved"})
}

func (s *Server) handleSearchTickers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "q parameter required", http.StatusBadRequest)
		return
	}
	results, err := s.tickerSvc.SearchTickers(r.Context(), query)
	if err != nil {
		s.logger.Error("search tickers failed", "error", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (s *Server) handleGetTickerDetails(w http.ResponseWriter, r *http.Request) {
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

	details, err := s.tickerSvc.GetTickerDetails(r.Context(), symbol)
	if err != nil {
		s.logger.Error("get ticker details failed", "symbol", symbol, "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "ticker not found"})
		return
	}

	intraday, _ := s.tickerSvc.GetIntradayBars(r.Context(), symbol)
	ratios, _ := s.tickerSvc.GetFinancialRatios(r.Context(), symbol)

	response := map[string]interface{}{
		"symbol":      details.Symbol,
		"name":        details.Name,
		"exchange":    details.Exchange,
		"sector":      details.Sector,
		"industry":    details.Industry,
		"price":       details.Price,
		"change":      details.Change,
		"changePct":   details.ChangePct,
		"volume":      details.Volume,
		"avgVolume":   details.AvgVolume,
		"marketCap":   details.MarketCap,
		"peRatio":     details.PeRatio,
		"eps":         details.Eps,
		"dividendYield": details.DividendYield,
		"week52High":  details.Week52High,
		"week52Low":   details.Week52Low,
	}

	if len(intraday) > 0 {
		response["intraday"] = intraday
	}
	if len(ratios) > 0 {
		response["ratios"] = ratios
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	cfg := Load()
	srv, err := NewServer(cfg)
	if err != nil {
		os.Exit(1)
	}
	if err := srv.Start(); err != nil {
		os.Exit(1)
	}
}
