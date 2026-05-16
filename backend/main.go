package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/portfolio-sim/backend/database"
	"github.com/portfolio-sim/backend/logging"
	"github.com/portfolio-sim/backend/redis"
	"github.com/portfolio-sim/backend/services"
	"github.com/portfolio-sim/shared/secrets"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
}

type ServerConfig struct {
	HTTPPort     int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	ShutdownTime time.Duration
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
	logClient    *logging.Client
	db           *database.Postgres
	redisClient  *redis.Client
	obsService   *services.ObservabilityService
	portfolioSvc *services.PortfolioService
	providerSvc  *services.ProviderService
	tickerSvc    *services.TickerService
	newsFeedSvc  *services.NewsFeedService
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

	redisClient, err := redis.NewClient(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password)
	if err != nil {
		return nil, fmt.Errorf("connect redis: %w", err)
	}

	logURL := os.Getenv("LOGGING_SERVICE_URL")
	if logURL == "" {
		logURL = "http://main-api:8080/api/logs"
	}
	logClient := logging.NewClient("backend", logURL)

	providerSecret := os.Getenv("PROVIDER_SECRET_KEY")
	if providerSecret == "" {
		providerSecret = cfg.Database.Password
	}
	secretCodec, err := secrets.NewCodec(providerSecret)
	if err != nil {
		return nil, fmt.Errorf("init provider secret codec: %w", err)
	}

	return &Server{
		cfg:          cfg,
		logger:       logger,
		logClient:    logClient,
		db:           db,
		redisClient:  redisClient,
		obsService:   services.NewObservabilityService(),
		portfolioSvc: services.NewPortfolioService(),
		providerSvc:  services.NewProviderService(logger, logClient, secretCodec),
		tickerSvc:    services.NewTickerService(os.Getenv("MARKET_DATA_SERVICE_URL"), logClient),
		newsFeedSvc:  services.NewNewsFeedService(os.Getenv("NEWS_FEED_SERVICE_URL"), logClient),
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

	lm := logging.Middleware(s.logClient)

	http.HandleFunc("GET /health", s.handleHealth)
	http.HandleFunc("POST /api/logs", s.handleIngestLog)
	http.HandleFunc("GET /api/observability/services", lm(http.HandlerFunc(s.handleGetServices)))
	http.HandleFunc("GET /api/observability/logs", lm(http.HandlerFunc(s.handleGetLogs)))
	http.HandleFunc("GET /api/portfolio/positions", lm(http.HandlerFunc(s.handleGetPositions)))
	http.HandleFunc("POST /api/portfolio/positions", lm(http.HandlerFunc(s.handleAddPosition)))
	http.HandleFunc("GET /api/portfolio/summary", lm(http.HandlerFunc(s.handleGetPortfolioSummary)))
	http.HandleFunc("GET /api/market/indices", lm(http.HandlerFunc(s.handleGetMarketIndices)))
	http.HandleFunc("GET /api/settings/market-indices", lm(http.HandlerFunc(s.handleGetMarketIndexSettings)))
	http.HandleFunc("PUT /api/settings/market-indices", lm(http.HandlerFunc(s.handleUpdateMarketIndexSettings)))
	http.HandleFunc("GET /api/news", lm(http.HandlerFunc(s.handleGetNews)))
	http.HandleFunc("GET /api/strategies", lm(http.HandlerFunc(s.handleGetStrategies)))
	http.HandleFunc("GET /api/signals", lm(http.HandlerFunc(s.handleGetSignals)))
	http.HandleFunc("GET /api/notifications", lm(http.HandlerFunc(s.handleGetNotifications)))
	http.HandleFunc("POST /api/notifications/dismiss", lm(http.HandlerFunc(s.handleDismissNotification)))
	http.HandleFunc("GET /api/providers", lm(http.HandlerFunc(s.handleGetProviders)))
	http.HandleFunc("PUT /api/providers", lm(http.HandlerFunc(s.handleUpdateProvider)))
	http.HandleFunc("GET /internal/providers/", lm(http.HandlerFunc(s.handleGetInternalProviderKey)))
	http.HandleFunc("POST /api/providers/validate", lm(http.HandlerFunc(s.handleValidateProvider)))
	http.HandleFunc("POST /api/providers/questrade/oauth", lm(http.HandlerFunc(s.handleSaveQuestradeOAuth)))
	http.HandleFunc("GET /api/providers/questrade/oauth", lm(http.HandlerFunc(s.handleGetQuestradeOAuth)))
	http.HandleFunc("POST /api/providers/questrade/refresh", lm(http.HandlerFunc(s.handleRefreshQuestradeToken)))
	http.HandleFunc("DELETE /api/providers", lm(http.HandlerFunc(s.handleDeleteProvider)))
	http.HandleFunc("GET /api/connections", lm(http.HandlerFunc(s.handleGetConnections)))
	http.HandleFunc("GET /api/rss-feeds", lm(http.HandlerFunc(s.handleGetRSSFeeds)))
	http.HandleFunc("POST /api/rss-feeds", lm(http.HandlerFunc(s.handleAddRSSFeed)))
	http.HandleFunc("DELETE /api/rss-feeds", lm(http.HandlerFunc(s.handleDeleteRSSFeed)))
	http.HandleFunc("POST /api/rss-feeds/scrape", lm(http.HandlerFunc(s.handleScrapeRSSFeeds)))
	http.HandleFunc("GET /api/tickers/search", lm(http.HandlerFunc(s.handleSearchTickers)))
	http.HandleFunc("GET /api/tickers/", lm(http.HandlerFunc(s.handleGetTickerDetails)))
	http.HandleFunc("GET /api/channels", lm(http.HandlerFunc(s.handleGetChannels)))
	http.HandleFunc("GET /api/videos/latest", lm(http.HandlerFunc(s.handleGetLatestVideos)))
	http.HandleFunc("GET /api/videos", lm(http.HandlerFunc(s.handleGetStoredVideos)))
	http.HandleFunc("POST /api/videos/analyze", lm(http.HandlerFunc(s.handleAnalyzeVideo)))
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
		Route     string                 `json:"route,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		s.logger.Error("handleIngestLog: failed to decode", "error", err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if entry.Level == "" || entry.Service == "" || entry.Message == "" {
		s.logger.Error("handleIngestLog: missing required fields", "level", entry.Level, "service", entry.Service, "message", entry.Message)
		http.Error(w, "level, service, and message are required", http.StatusBadRequest)
		return
	}

	metadataJSON, err := json.Marshal(entry.Metadata)
	if err != nil {
		s.logger.Warn("log metadata marshaling failed", "error", err)
		metadataJSON = []byte(`{}`)
	}
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	} else if _, err := uuid.Parse(entry.ID); err != nil {
		entry.ID = uuid.New().String()
	}
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}
	_, err = s.db.Exec(r.Context(), `
		INSERT INTO logs (id, timestamp, level, service, component, message, metadata, trace_id, span_id, route)
		VALUES ($1, $2::timestamptz, $3, $4, $5, $6, $7, $8, $9, $10)
	`, entry.ID, entry.Timestamp, entry.Level, entry.Service, entry.Component, entry.Message, metadataJSON, entry.TraceID, entry.SpanID, entry.Route)
	if err != nil {
		s.logger.Error("log insert failed", "error", err, "service", entry.Service, "id", entry.ID)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	level := r.URL.Query().Get("level")
	service := r.URL.Query().Get("service")
	minutesStr := r.URL.Query().Get("minutes")
	minutes := 60
	if minutesStr != "" {
		if m, err := strconv.Atoi(minutesStr); err == nil && m > 0 {
			minutes = m
		}
	}

	route := r.URL.Query()["route"]

	s.logger.Info("handleGetLogs: fetching logs", "limit", limit, "level", level, "service", service, "minutes", minutes, "route", route)

	var query string
	var args []interface{}
	if len(route) > 0 {
		query = `
			SELECT id, timestamp::text, level, service, component, message, metadata, trace_id, span_id, COALESCE(route, '')
			FROM logs
			WHERE ($1 = '' OR level = $1)
			  AND ($2 = '' OR service = $2)
			  AND COALESCE(route, '') = ANY($5)
			  AND timestamp >= NOW() - INTERVAL '1 minute' * $3
			ORDER BY timestamp DESC
			LIMIT $4
		`
		args = []interface{}{level, service, minutes, limit, route}
	} else {
		query = `
			SELECT id, timestamp::text, level, service, component, message, metadata, trace_id, span_id, COALESCE(route, '')
			FROM logs
			WHERE ($1 = '' OR level = $1)
			  AND ($2 = '' OR service = $2)
			  AND timestamp >= NOW() - INTERVAL '1 minute' * $3
			ORDER BY timestamp DESC
			LIMIT $4
		`
		args = []interface{}{level, service, minutes, limit}
	}

	rows, err := s.db.Query(r.Context(), query, args...)
	if err != nil {
		s.logger.Error("handleGetLogs: query failed", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to fetch logs"})
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
		Route     string                 `json:"route,omitempty"`
	}

	logs := []LogEntry{}
	for rows.Next() {
		var le LogEntry
		var metadata []byte
		if err := rows.Scan(&le.ID, &le.Timestamp, &le.Level, &le.Service, &le.Component, &le.Message, &metadata, &le.TraceID, &le.SpanID, &le.Route); err != nil {
			s.logger.Error("handleGetLogs: scan failed", "error", err)
			continue
		}
		if metadata != nil {
			json.Unmarshal(metadata, &le.Metadata)
			if le.Route == "" && le.Metadata != nil {
				if r, ok := le.Metadata["route"].(string); ok {
					le.Route = r
				}
			}
		}
		logs = append(logs, le)
	}

	s.logger.Info("handleGetLogs: returning logs", "count", len(logs))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func (s *Server) handleGetNotifications(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query(r.Context(), `
		SELECT l.id, l.timestamp::text, l.level, l.service, l.message,
		       CASE WHEN d.log_id IS NULL THEN false ELSE true END as dismissed
		FROM logs l
		LEFT JOIN notification_dismissals d ON l.id::text = d.log_id
		WHERE l.level IN ('WARN', 'ERROR', 'FATAL')
		  AND l.timestamp >= NOW() - INTERVAL '24 hours'
		  AND d.log_id IS NULL
		ORDER BY l.timestamp DESC
		LIMIT 20
	`)
	if err != nil {
		s.logger.Error("handleGetNotifications: query failed", "error", err)
		http.Error(w, "failed to fetch notifications", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Notification struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Message   string `json:"message"`
		Type      string `json:"type"`
		Timestamp string `json:"timestamp"`
		Read      bool   `json:"read"`
	}

	notifications := []Notification{}
	for rows.Next() {
		var id, timestamp, level, service, message string
		var dismissed bool
		if err := rows.Scan(&id, &timestamp, &level, &service, &message, &dismissed); err != nil {
			s.logger.Error("handleGetNotifications: scan failed", "error", err)
			continue
		}

		notificationType := "warning"
		if level == "ERROR" || level == "FATAL" {
			notificationType = "error"
		}

		serviceTitle := strings.ReplaceAll(service, "-", " ")
		if serviceTitle == "" {
			serviceTitle = "system"
		}
		title := strings.ToUpper(level[:1]) + strings.ToLower(level[1:]) + " from " + serviceTitle

		notifications = append(notifications, Notification{
			ID:        id,
			Title:     title,
			Message:   message,
			Type:      notificationType,
			Timestamp: timestamp,
			Read:      dismissed,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}

func (s *Server) handleDismissNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	_, err := s.db.Exec(r.Context(),
		`INSERT INTO notification_dismissals (log_id) VALUES ($1) ON CONFLICT (log_id) DO NOTHING`,
		req.ID)
	if err != nil {
		s.logger.Error("handleDismissNotification: insert failed", "error", err)
		http.Error(w, "failed to dismiss notification", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

	symbols := uniqueSymbols(r.URL.Query()["symbols"])
	streamStreams := make([]string, 0)
	if len(symbols) > 0 {
		for _, sym := range symbols {
			s.queueTickerSubscribe(ctx, sym)
		}
		defer func() {
			for _, sym := range symbols {
				s.queueTickerAction(context.Background(), sym, "unsubscribe")
			}
		}()
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

func uniqueSymbols(symbols []string) []string {
	seen := make(map[string]struct{}, len(symbols))
	unique := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		symbol = strings.ToUpper(strings.TrimSpace(symbol))
		if symbol == "" {
			continue
		}
		if _, ok := seen[symbol]; ok {
			continue
		}
		seen[symbol] = struct{}{}
		unique = append(unique, symbol)
	}
	return unique
}

func (s *Server) queueTickerSubscribe(ctx context.Context, symbol string) {
	s.queueTickerAction(ctx, symbol, "subscribe")
}

func (s *Server) queueTickerAction(ctx context.Context, symbol, action string) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return
	}
	action = strings.ToLower(strings.TrimSpace(action))
	if action == "" {
		return
	}

	payload, err := json.Marshal(map[string]string{
		"symbol": symbol,
		"action": action,
	})
	if err != nil {
		s.logger.Warn("marshal ticker action failed", "symbol", symbol, "action", action, "error", err)
		return
	}

	if err := s.redisClient.LPush(ctx, "queue:ticker:subscribe", string(payload)); err != nil {
		s.logger.Warn("enqueue ticker action failed", "symbol", symbol, "action", action, "error", err)
	}
}

func (s *Server) fetchMarketIndexDetails(ctx context.Context, symbol string) (*services.TickerDetails, error) {
	return s.tickerSvc.GetTickerQuote(ctx, symbol)
}

func (s *Server) hydrateMissingMarketIndices(ctx context.Context, indices []services.MarketIndex) []services.MarketIndex {
	hydrated := make([]services.MarketIndex, len(indices))
	copy(hydrated, indices)

	for i := range hydrated {
		if hydrated[i].Price > 0 {
			continue
		}

		details, err := s.fetchMarketIndexDetails(ctx, hydrated[i].Symbol)
		if err != nil {
			s.logger.Warn("market index fallback fetch failed", "symbol", hydrated[i].Symbol, "error", err)
			continue
		}
		if details == nil || details.Price <= 0 {
			continue
		}

		hydrated[i].Price = details.Price
		hydrated[i].Change = details.Change
		hydrated[i].ChangePct = details.ChangePct
		if strings.TrimSpace(hydrated[i].Name) == "" {
			hydrated[i].Name = details.Name
		}
	}

	return hydrated
}

func (s *Server) handleGetPositions(w http.ResponseWriter, r *http.Request) {
	portfolioID := r.URL.Query().Get("portfolio_id")
	if portfolioID == "" || portfolioID == "default" {
		portfolioID = "00000000-0000-0000-0000-000000000001"
	}
	positions, err := s.portfolioSvc.GetPositions(r.Context(), s.db, portfolioID)
	if err != nil {
		positions = []services.Position{}
	}
	safePercent := func(numerator, denominator float64) float64 {
		if denominator == 0 {
			return 0
		}
		return (numerator / denominator) * 100
	}
	if len(positions) > 0 {
		var tickerIDs []string
		for _, p := range positions {
			tickerIDs = append(tickerIDs, p.TickerID)
		}
		prices, _ := s.portfolioSvc.GetLatestPrices(r.Context(), s.db, tickerIDs)
		for i := range positions {
			if price, ok := prices[positions[i].TickerID]; ok {
				positions[i].CurrentPrice = price.Price
				if price.ChangePct != 0 {
					positions[i].DayChange = positions[i].CurrentPrice * (price.ChangePct / 100)
					positions[i].DayChangePct = price.ChangePct
				} else {
					positions[i].DayChange = price.Change
					positions[i].DayChangePct = safePercent(price.Change, price.Price-price.Change)
				}
			}
			if positions[i].CurrentPrice == 0 {
				positions[i].CurrentPrice = positions[i].AvgCost
			}
			positions[i].CurrentValue = positions[i].Quantity * positions[i].CurrentPrice
			positions[i].TotalGain = positions[i].CurrentValue - (positions[i].Quantity * positions[i].AvgCost)
			positions[i].TotalGainPct = safePercent(positions[i].TotalGain, positions[i].Quantity*positions[i].AvgCost)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(positions)
}

func (s *Server) handleAddPosition(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PortfolioID string  `json:"portfolio_id"`
		Symbol      string  `json:"symbol"`
		Shares      float64 `json:"shares"`
		Price       float64 `json:"price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	portfolioID := req.PortfolioID
	if portfolioID == "" || portfolioID == "default" {
		portfolioID = "00000000-0000-0000-0000-000000000001"
	}
	err := s.portfolioSvc.AddPosition(r.Context(), s.db, portfolioID, req.Symbol, req.Shares, req.Price)
	if err != nil {
		s.logger.Error("add position failed", "symbol", req.Symbol, "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "added"})
}

func (s *Server) handleGetPortfolioSummary(w http.ResponseWriter, r *http.Request) {
	portfolioID := r.URL.Query().Get("portfolio_id")
	if portfolioID == "" || portfolioID == "default" {
		portfolioID = "00000000-0000-0000-0000-000000000001"
	}
	summary, err := s.portfolioSvc.GetPortfolioSummary(r.Context(), s.db, portfolioID)
	if err != nil || summary == nil {
		summary = &services.PortfolioSummary{}
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(summary); err != nil {
		s.logger.Error("encode portfolio summary failed", "portfolio_id", portfolioID, "error", err)
		http.Error(w, "failed to encode portfolio summary", http.StatusInternalServerError)
	}
}

func (s *Server) handleGetMarketIndices(w http.ResponseWriter, r *http.Request) {
	indices, _ := s.portfolioSvc.GetMarketIndices(r.Context(), s.db)
	indices = s.hydrateMissingMarketIndices(r.Context(), indices)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(indices)
}

func (s *Server) handleGetMarketIndexSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := s.portfolioSvc.GetMarketIndexSettings(r.Context(), s.db)
	if err != nil {
		s.logger.Error("get market index settings failed", "error", err)
		http.Error(w, "failed to fetch market index settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

func (s *Server) handleUpdateMarketIndexSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var settings []services.MarketIndexSetting
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := s.portfolioSvc.SaveMarketIndexSettings(r.Context(), s.db, settings); err != nil {
		s.logger.Error("save market index settings failed", "error", err)
		http.Error(w, "failed to save market index settings", http.StatusInternalServerError)
		return
	}

	savedSettings, err := s.portfolioSvc.GetMarketIndexSettings(r.Context(), s.db)
	if err != nil {
		s.logger.Error("reload market index settings failed", "error", err)
		http.Error(w, "failed to reload market index settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(savedSettings)
}

func (s *Server) handleGetNews(w http.ResponseWriter, r *http.Request) {
	articles, _ := s.portfolioSvc.GetNewsArticles(r.Context(), s.db)
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
			{ID: "massive", ProviderID: "massive", Name: "Massive", Description: "Real-time and historical market data", Type: "market_data", RateLimit: 5, DocURL: "https://massive.com/docs/rest/quickstart", TokenExpired: false},
			{ID: "questrade", ProviderID: "questrade", Name: "Questrade", Description: "Questrade market data API", Type: "market_data", RateLimit: 100, DocURL: "https://www.questrade.com/api", TokenExpired: false},
			{ID: "fmp", ProviderID: "fmp", Name: "Financial Modeling Prep", Description: "Financial statements and fundamental data", Type: "market_data", RateLimit: 250, DocURL: "https://site.financialmodelingprep.com/developer/docs", TokenExpired: false},
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
		ProviderID   string `json:"provider_id"`
		APIKey       string `json:"api_key"`
		Validated    bool   `json:"validated"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		APIServer    string `json:"api_server"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.ProviderID == "" || req.APIKey == "" {
		http.Error(w, "provider_id and api_key are required", http.StatusBadRequest)
		return
	}

	if req.ProviderID == "questrade" && req.Validated {
		if req.RefreshToken == "" || req.AccessToken == "" || req.APIServer == "" {
			http.Error(w, "validated questrade save requires oauth payload", http.StatusBadRequest)
			return
		}
		if err := s.providerSvc.SaveQuestradeOAuth(r.Context(), s.db, req.ProviderID, req.AccessToken, req.RefreshToken, req.APIServer, req.ExpiresIn); err != nil {
			http.Error(w, "failed to save", http.StatusInternalServerError)
			return
		}
	} else {
		if err := s.providerSvc.SaveProviderKey(r.Context(), s.db, req.ProviderID, req.APIKey); err != nil {
			http.Error(w, "failed to save", http.StatusInternalServerError)
			return
		}
		if req.Validated {
			if err := s.providerSvc.UpdateProviderValidationState(r.Context(), s.db, req.ProviderID, true, ""); err != nil {
				http.Error(w, "failed to save", http.StatusInternalServerError)
				return
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "saved"})
}

func (s *Server) requireInternalAPI(w http.ResponseWriter, r *http.Request) bool {
	expected := strings.TrimSpace(os.Getenv("INTERNAL_API_TOKEN"))
	if expected == "" {
		s.logger.Error("internal api token not configured")
		http.Error(w, "internal api unavailable", http.StatusServiceUnavailable)
		return false
	}
	if r.Header.Get("X-Internal-API-Token") != expected {
		http.Error(w, "forbidden", http.StatusForbidden)
		return false
	}
	return true
}

func (s *Server) handleGetInternalProviderKey(w http.ResponseWriter, r *http.Request) {
	if !s.requireInternalAPI(w, r) {
		return
	}

	providerID := strings.Trim(strings.TrimPrefix(r.URL.Path, "/internal/providers/"), "/")
	if providerID == "" || strings.Contains(providerID, "/") {
		http.Error(w, "provider_id required", http.StatusBadRequest)
		return
	}

	key, err := s.providerSvc.GetDecryptedProviderKey(r.Context(), s.db, providerID)
	if err != nil {
		s.logger.Error("get provider key failed", "provider", providerID, "error", err)
		http.Error(w, "failed to get key", http.StatusInternalServerError)
		return
	}
	if key == "" {
		http.Error(w, "key not found or not validated", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"api_key": key})
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

	valid, qtResult, err := s.providerSvc.ValidateProviderKey(r.Context(), s.db, req.ProviderID, req.APIKey)
	if err != nil {
		s.logger.Error("provider validation failed", "provider", req.ProviderID, "error", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"valid": false, "error": err.Error()})
		return
	}

	if !valid {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid": false,
			"error": "invalid credentials or no entitlement",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"valid":    true,
		"save_key": req.APIKey,
	}
	if req.ProviderID == "questrade" && qtResult != nil {
		response["save_key"] = qtResult.RefreshToken
		response["access_token"] = qtResult.AccessToken
		response["refresh_token"] = qtResult.RefreshToken
		response["api_server"] = qtResult.APIServer
		response["expires_in"] = qtResult.ExpiresIn
		if s.logClient != nil {
			s.logClient.InfoWithMeta(r.Context(), fmt.Sprintf("Questrade OAuth tokens validated: %s", req.ProviderID), map[string]interface{}{
				"provider":          req.ProviderID,
				"api_server":        qtResult.APIServer,
				"expires_in":        qtResult.ExpiresIn,
				"has_access_token":  qtResult.AccessToken != "",
				"has_refresh_token": qtResult.RefreshToken != "",
				"type":              "questrade_oauth_result",
			})
		}
	}
	json.NewEncoder(w).Encode(response)
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

func (s *Server) handleScrapeRSSFeeds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.logClient != nil {
		s.logClient.InfoWithMeta(r.Context(), "RSS scrape request", map[string]interface{}{
			"type":    "outbound_request",
			"target":  "news-feed-service",
			"method":  "POST",
			"url":     "http://news-feed-service:8080/api/scrape",
			"service": "backend",
		})
	}

	resp, err := http.Post("http://news-feed-service:8080/api/scrape", "application/json", nil)
	if err != nil {
		if s.logClient != nil {
			s.logClient.ErrorWithMeta(r.Context(), "RSS scrape failed", map[string]interface{}{
				"type":  "outbound_request_error",
				"error": err.Error(),
			})
		}
		http.Error(w, "failed to trigger scrape", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if s.logClient != nil {
		s.logClient.InfoWithMeta(r.Context(), "RSS scrape response", map[string]interface{}{
			"type":          "outbound_response",
			"status":        resp.StatusCode,
			"body":          string(body),
			"response_size": len(body),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleGetQuestradeOAuth(w http.ResponseWriter, r *http.Request) {
	accessToken, refreshToken, apiServer, err := s.providerSvc.GetQuestradeOAuth(r.Context(), s.db, "questrade")
	if err != nil {
		s.logger.Error("get questrade oauth failed", "error", err)
		http.Error(w, "failed to get oauth tokens", http.StatusInternalServerError)
		return
	}
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
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.RefreshToken == "" {
		http.Error(w, "refresh_token is required", http.StatusBadRequest)
		return
	}
	expiresIn := req.ExpiresIn
	if expiresIn == 0 {
		expiresIn = 3600
	}
	if err := s.providerSvc.SaveQuestradeOAuth(r.Context(), s.db, "questrade", req.AccessToken, req.RefreshToken, req.APIServer, expiresIn); err != nil {
		http.Error(w, "failed to save", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "saved"})
}

func (s *Server) handleRefreshQuestradeToken(w http.ResponseWriter, r *http.Request) {
	writeRefreshError := func(status int, message string) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(map[string]string{"error": message})
	}

	if r.Method != http.MethodPost {
		writeRefreshError(http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	_, refreshToken, _, err := s.providerSvc.GetQuestradeOAuth(r.Context(), s.db, "questrade")
	if err != nil || refreshToken == "" {
		message := "no refresh token available"
		if err != nil {
			message = err.Error()
		}
		if updateErr := s.providerSvc.UpdateProviderValidationState(r.Context(), s.db, "questrade", false, message); updateErr != nil {
			s.logger.Error("failed to record questrade refresh failure", "error", updateErr)
		}
		writeRefreshError(http.StatusBadRequest, message)
		return
	}
	s.logger.Info("refreshing questrade token")
	newAccessToken, newRefreshToken, newAPIServer, expiresIn, err := s.providerSvc.ExchangeQuestradeToken(r.Context(), refreshToken)
	if err != nil {
		s.logger.Error("failed to refresh questrade token", "error", err)
		message := "failed to refresh token: " + err.Error()
		if updateErr := s.providerSvc.UpdateProviderValidationState(r.Context(), s.db, "questrade", false, message); updateErr != nil {
			s.logger.Error("failed to record questrade refresh failure", "error", updateErr)
		}
		writeRefreshError(http.StatusInternalServerError, message)
		return
	}
	if err := s.providerSvc.SaveQuestradeOAuth(r.Context(), s.db, "questrade", newAccessToken, newRefreshToken, newAPIServer, expiresIn); err != nil {
		writeRefreshError(http.StatusInternalServerError, "failed to save new token")
		return
	}
	s.logger.Info("questrade token refreshed successfully")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "refreshed"})
}

func (s *Server) handleDeleteProvider(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	providerID := r.URL.Query().Get("id")
	if providerID == "" {
		http.Error(w, "provider id required", http.StatusBadRequest)
		return
	}

	if err := s.providerSvc.DeleteProviderConfig(r.Context(), s.db, providerID); err != nil {
		s.logger.Error("failed to delete provider config", "provider_id", providerID, "error", err)
		http.Error(w, "failed to delete provider", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
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
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{"error": "ticker search unavailable"})
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

	intradayRange := r.URL.Query().Get("range")
	intraday, _ := s.tickerSvc.GetIntradayBars(r.Context(), symbol, "", intradayRange)
	ratios, _ := s.tickerSvc.GetFinancialRatios(r.Context(), symbol)

	response := map[string]interface{}{
		"symbol":        details.Symbol,
		"name":          details.Name,
		"exchange":      details.Exchange,
		"sector":        details.Sector,
		"industry":      details.Industry,
		"price":         details.Price,
		"change":        details.Change,
		"changePct":     details.ChangePct,
		"volume":        details.Volume,
		"avgVolume":     details.AvgVolume,
		"marketCap":     details.MarketCap,
		"peRatio":       details.PeRatio,
		"eps":           details.Eps,
		"dividendYield": details.DividendYield,
		"week52High":    details.Week52High,
		"week52Low":     details.Week52Low,
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

func (s *Server) handleGetChannels(w http.ResponseWriter, r *http.Request) {
	data, err := s.newsFeedSvc.GetChannels(r.Context())
	if err != nil {
		s.logger.Error("get channels failed", "error", err)
		http.Error(w, "failed to get channels", http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (s *Server) handleGetLatestVideos(w http.ResponseWriter, r *http.Request) {
	channelID := r.URL.Query().Get("channel_id")
	if channelID == "" {
		http.Error(w, "channel_id required", http.StatusBadRequest)
		return
	}
	data, err := s.newsFeedSvc.GetLatestVideos(r.Context(), channelID)
	if err != nil {
		s.logger.Error("get latest videos failed", "error", err)
		http.Error(w, "failed to get videos", http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (s *Server) handleGetStoredVideos(w http.ResponseWriter, r *http.Request) {
	data, err := s.newsFeedSvc.GetStoredVideos(r.Context())
	if err != nil {
		s.logger.Error("get stored videos failed", "error", err)
		http.Error(w, "failed to get videos", http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (s *Server) handleAnalyzeVideo(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		VideoID string `json:"video_id"`
		Title   string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := s.newsFeedSvc.AnalyzeVideo(r.Context(), req.VideoID, req.Title); err != nil {
		s.logger.Error("analyze video failed", "error", err)
		http.Error(w, "failed to analyze video", http.StatusBadGateway)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
