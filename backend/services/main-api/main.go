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
	"syscall"
	"time"

	"github.com/portfolio-sim/backend/internal/config"
	"github.com/portfolio-sim/backend/internal/database"
	"github.com/portfolio-sim/backend/services/main-api/handlers"
	"github.com/portfolio-sim/backend/services/main-api/logging"
	"github.com/portfolio-sim/backend/services/main-api/redis"
	"github.com/portfolio-sim/backend/services/main-api/sse"
)

type Server struct {
	cfg       *config.Config
	db        *database.Postgres
	redis     *redis.Client
	logger    *slog.Logger
	sseMgr    *sse.Manager
	logClient logging.LogEmitter
}

func NewServer(cfg *config.Config) *Server {
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

	sseMgr := sse.NewManager(redisClient)

	return &Server{
		cfg:       cfg,
		db:        db,
		redis:     redisClient,
		logger:    logger,
		sseMgr:    sseMgr,
		logClient: logging.NewNoOpLogger(),
	}
}

func (s *Server) Start() error {
	http.HandleFunc("GET /health", s.handleHealth)
	http.HandleFunc("GET /stream", s.handleSSE)
	http.HandleFunc("GET /api/portfolios", withAuth(handlers.HandleListPortfolios))
	http.HandleFunc("POST /api/portfolios", withAuth(handlers.HandleCreatePortfolio))
	http.HandleFunc("GET /api/watchlists", withAuth(handlers.HandleListWatchlists))
	http.HandleFunc("POST /api/watchlists", withAuth(handlers.HandleCreateWatchlist))
	http.HandleFunc("GET /api/tickers", withAuth(handlers.HandleListTickers))
	http.HandleFunc("GET /api/jobs", withAuth(handlers.HandleListJobs))
	http.HandleFunc("POST /api/jobs", withAuth(handlers.HandleCreateJob))

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
	s.redis.Close()
	s.db.Close()
	return nil
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	clerkID := validateToken(r)
	if clerkID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	channelsParam := query.Get("channels")
	channels := sse.ParseChannels(channelsParam)

	if len(channels) == 0 {
		http.Error(w, "channels required", http.StatusBadRequest)
		return
	}

	client := &sse.Client{ID: fmt.Sprintf("%d", time.Now().UnixNano()), Channels: channels, Send: make(chan []byte, 256)}
	s.sseMgr.Register(client)
	defer s.sseMgr.Unregister(client)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, _ := w.(http.Flusher)

	for {
		select {
		case data := <-client.Send:
			w.Write(data)
			flusher.Flush()
		case <-time.After(30 * time.Second):
			fmt.Fprintf(w, "event: heartbeat\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func validateToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}

func withAuth(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		clerkID := validateToken(r)
		if clerkID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "clerk_id", clerkID)
		handler(w, r.WithContext(ctx))
	}
}

func main() {
	cfg := config.Load()
	server := NewServer(cfg)
	if err := server.Start(); err != nil {
		os.Exit(1)
	}
}