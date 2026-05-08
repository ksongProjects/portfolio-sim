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
	"syscall"
	"time"

	"github.com/portfolio-sim/backend/database"
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
	cfg           *Config
	logger        *slog.Logger
	db            *database.Postgres
	obsService    *services.ObservabilityService
	portfolioSvc  *services.PortfolioService
	providerSvc   *services.ProviderService
}

func NewServer(cfg *Config) (*Server, error) {
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
		return nil, fmt.Errorf("connect database: %w", err)
	}

	return &Server{
		cfg:          cfg,
		logger:       logger,
		db:           db,
		obsService:   services.NewObservabilityService(),
		portfolioSvc: services.NewPortfolioService(),
		providerSvc:  services.NewProviderService(),
	}, nil
}

func (s *Server) Start() error {
	http.HandleFunc("GET /health", s.handleHealth)
	http.HandleFunc("GET /api/observability/services", s.handleGetServices)
	http.HandleFunc("GET /api/observability/logs", s.handleGetLogs)
	http.HandleFunc("GET /api/portfolio/positions", s.handleGetPositions)
	http.HandleFunc("GET /api/portfolio/summary", s.handleGetPortfolioSummary)
	http.HandleFunc("GET /api/market/indices", s.handleGetMarketIndices)
	http.HandleFunc("GET /api/news", s.handleGetNews)
	http.HandleFunc("GET /api/strategies", s.handleGetStrategies)
	http.HandleFunc("GET /api/signals", s.handleGetSignals)
	http.HandleFunc("GET /api/providers", s.handleGetProviders)
	http.HandleFunc("PUT /api/providers", s.handleUpdateProvider)
	http.HandleFunc("GET /api/connections", s.handleGetConnections)

	s.logger.Info("main api server starting", "port", s.cfg.Server.HTTPPort)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.cfg.Server.HTTPPort),
		Handler:      http.DefaultServeMux,
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

func (s *Server) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	loggingURL := "http://localhost:9090/api/logs"
	limit := r.URL.Query().Get("limit")
	if limit != "" {
		loggingURL += "?limit=" + limit
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, loggingURL, nil)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func (s *Server) handleGetPositions(w http.ResponseWriter, r *http.Request) {
	portfolioID := r.URL.Query().Get("portfolio_id")
	if portfolioID == "" {
		portfolioID = "default"
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
	if portfolioID == "" {
		portfolioID = "default"
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
	if err != nil || providers == nil {
		providers = []services.ProviderConfig{}
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
