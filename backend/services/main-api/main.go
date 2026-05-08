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

	"github.com/portfolio-sim/backend/services/main-api/config"
	"github.com/portfolio-sim/backend/services/main-api/database"
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

	logClient := logging.NewNoOpLogger()
	if cfg.GRPCPort > 0 {
		addr := fmt.Sprintf("localhost:%d", cfg.GRPCPort)
		if client, err := logging.NewClient(addr); err == nil {
			logClient = client
		}
	}

	return &Server{
		cfg:       cfg,
		db:        db,
		redis:     redisClient,
		logger:    logger,
		sseMgr:    sseMgr,
		logClient: logClient,
	}
}

func (s *Server) Start() error {
	http.HandleFunc("GET /health", s.handleHealth)
	http.HandleFunc("GET /stream", s.handleSSE)
	http.HandleFunc("GET /api/portfolios", withAuth(s.handleListPortfolios))
	http.HandleFunc("POST /api/portfolios", withAuth(s.handleCreatePortfolio))
	http.HandleFunc("GET /api/portfolios/", withAuth(s.handleGetPortfolio))
	http.HandleFunc("PUT /api/portfolios/", withAuth(s.handleUpdatePortfolio))
	http.HandleFunc("DELETE /api/portfolios/", withAuth(s.handleDeletePortfolio))
	http.HandleFunc("GET /api/watchlists", withAuth(s.handleListWatchlists))
	http.HandleFunc("POST /api/watchlists", withAuth(s.handleCreateWatchlist))
	http.HandleFunc("GET /api/watchlists/", withAuth(s.handleGetWatchlist))
	http.HandleFunc("PUT /api/watchlists/", withAuth(s.handleUpdateWatchlist))
	http.HandleFunc("DELETE /api/watchlists/", withAuth(s.handleDeleteWatchlist))
	http.HandleFunc("POST /api/watchlists/", withAuth(s.handleAddTickerToWatchlist))
	http.HandleFunc("DELETE /api/watchlists/", withAuth(s.handleRemoveTickerFromWatchlist))
	http.HandleFunc("GET /api/tickers", withAuth(s.handleListTickers))
	http.HandleFunc("GET /api/tickers/", withAuth(s.handleGetTicker))
	http.HandleFunc("GET /api/jobs", withAuth(s.handleListJobs))
	http.HandleFunc("POST /api/jobs", withAuth(s.handleCreateJob))
	http.HandleFunc("GET /api/jobs/", withAuth(s.handleGetJob))
	http.HandleFunc("POST /api/jobs/", withAuth(s.handleCancelJob))

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
	srv := NewServer(cfg)
	if err := srv.Start(); err != nil {
		os.Exit(1)
	}
}

func (s *Server) handleListPortfolios(w http.ResponseWriter, r *http.Request) {
	clerkID := r.Context().Value("clerk_id").(string)

	query := `SELECT id, clerk_id, name, description, created_at, updated_at FROM portfolios WHERE clerk_id = $1`
	rows, err := s.db.Pool().Query(r.Context(), query, clerkID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var portfolios []map[string]interface{}
	for rows.Next() {
		var p struct {
			ID, ClerkID, Name, Description string
			CreatedAt, UpdatedAt           string
		}
		if err := rows.Scan(&p.ID, &p.ClerkID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			continue
		}
		portfolios = append(portfolios, map[string]interface{}{
			"id":          p.ID,
			"clerk_id":    p.ClerkID,
			"name":        p.Name,
			"description": p.Description,
			"created_at":  p.CreatedAt,
			"updated_at":  p.UpdatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(portfolios)
}

func (s *Server) handleCreatePortfolio(w http.ResponseWriter, r *http.Request) {
	clerkID := r.Context().Value("clerk_id").(string)

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `INSERT INTO portfolios (clerk_id, name, description) VALUES ($1, $2, $3)
		RETURNING id, clerk_id, name, description, created_at, updated_at`

	var p struct {
		ID, ClerkID, Name, Description string
		CreatedAt, UpdatedAt           string
	}
	err := s.db.Pool().QueryRow(r.Context(), query, clerkID, input.Name, input.Description).
		Scan(&p.ID, &p.ClerkID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          p.ID,
		"clerk_id":    p.ClerkID,
		"name":        p.Name,
		"description": p.Description,
		"created_at":  p.CreatedAt,
		"updated_at":  p.UpdatedAt,
	})
}

func (s *Server) handleGetPortfolio(w http.ResponseWriter, r *http.Request) {
	clerkID := r.Context().Value("clerk_id").(string)
	id := extractPathID(r.URL.Path, "/api/portfolios/")

	query := `SELECT id, clerk_id, name, description, created_at, updated_at FROM portfolios WHERE id = $1 AND clerk_id = $2`
	var p struct {
		ID, ClerkID, Name, Description string
		CreatedAt, UpdatedAt           string
	}
	err := s.db.Pool().QueryRow(r.Context(), query, id, clerkID).
		Scan(&p.ID, &p.ClerkID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          p.ID,
		"clerk_id":    p.ClerkID,
		"name":        p.Name,
		"description": p.Description,
		"created_at":  p.CreatedAt,
		"updated_at":  p.UpdatedAt,
		"positions":   []interface{}{},
	})
}

func (s *Server) handleUpdatePortfolio(w http.ResponseWriter, r *http.Request) {
	clerkID := r.Context().Value("clerk_id").(string)
	id := extractPathID(r.URL.Path, "/api/portfolios/")

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `UPDATE portfolios SET name = $1, description = $2 WHERE id = $3 AND clerk_id = $4
		RETURNING id, clerk_id, name, description, created_at, updated_at`

	var p struct {
		ID, ClerkID, Name, Description string
		CreatedAt, UpdatedAt           string
	}
	err := s.db.Pool().QueryRow(r.Context(), query, input.Name, input.Description, id, clerkID).
		Scan(&p.ID, &p.ClerkID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          p.ID,
		"clerk_id":    p.ClerkID,
		"name":        p.Name,
		"description": p.Description,
		"created_at":  p.CreatedAt,
		"updated_at":  p.UpdatedAt,
	})
}

func (s *Server) handleDeletePortfolio(w http.ResponseWriter, r *http.Request) {
	clerkID := r.Context().Value("clerk_id").(string)
	id := extractPathID(r.URL.Path, "/api/portfolios/")

	query := `DELETE FROM portfolios WHERE id = $1 AND clerk_id = $2`
	result, err := s.db.Pool().Exec(r.Context(), query, id, clerkID)
	if err != nil || result.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListWatchlists(w http.ResponseWriter, r *http.Request) {
	clerkID := r.Context().Value("clerk_id").(string)

	query := `SELECT id, clerk_id, name, created_at, updated_at FROM watchlists WHERE clerk_id = $1`
	rows, err := s.db.Pool().Query(r.Context(), query, clerkID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var watchlists []map[string]interface{}
	for rows.Next() {
		var wl struct {
			ID, ClerkID, Name string
			CreatedAt, UpdatedAt string
		}
		if err := rows.Scan(&wl.ID, &wl.ClerkID, &wl.Name, &wl.CreatedAt, &wl.UpdatedAt); err != nil {
			continue
		}
		watchlists = append(watchlists, map[string]interface{}{
			"id":         wl.ID,
			"clerk_id":   wl.ClerkID,
			"name":       wl.Name,
			"created_at": wl.CreatedAt,
			"updated_at": wl.UpdatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(watchlists)
}

func (s *Server) handleCreateWatchlist(w http.ResponseWriter, r *http.Request) {
	clerkID := r.Context().Value("clerk_id").(string)

	var input struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `INSERT INTO watchlists (clerk_id, name) VALUES ($1, $2)
		RETURNING id, clerk_id, name, created_at, updated_at`

	var wl struct {
		ID, ClerkID, Name string
		CreatedAt, UpdatedAt string
	}
	err := s.db.Pool().QueryRow(r.Context(), query, clerkID, input.Name).
		Scan(&wl.ID, &wl.ClerkID, &wl.Name, &wl.CreatedAt, &wl.UpdatedAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         wl.ID,
		"clerk_id":   wl.ClerkID,
		"name":       wl.Name,
		"created_at": wl.CreatedAt,
		"updated_at": wl.UpdatedAt,
	})
}

func (s *Server) handleGetWatchlist(w http.ResponseWriter, r *http.Request) {
	clerkID := r.Context().Value("clerk_id").(string)
	id := extractPathID(r.URL.Path, "/api/watchlists/")

	query := `SELECT id, clerk_id, name, created_at, updated_at FROM watchlists WHERE id = $1 AND clerk_id = $2`
	var wl struct {
		ID, ClerkID, Name string
		CreatedAt, UpdatedAt string
	}
	err := s.db.Pool().QueryRow(r.Context(), query, id, clerkID).
		Scan(&wl.ID, &wl.ClerkID, &wl.Name, &wl.CreatedAt, &wl.UpdatedAt)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         wl.ID,
		"clerk_id":   wl.ClerkID,
		"name":       wl.Name,
		"created_at": wl.CreatedAt,
		"updated_at": wl.UpdatedAt,
		"tickers":    []interface{}{},
	})
}

func (s *Server) handleUpdateWatchlist(w http.ResponseWriter, r *http.Request) {
	clerkID := r.Context().Value("clerk_id").(string)
	id := extractPathID(r.URL.Path, "/api/watchlists/")

	var input struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `UPDATE watchlists SET name = $1 WHERE id = $2 AND clerk_id = $3
		RETURNING id, clerk_id, name, created_at, updated_at`

	var wl struct {
		ID, ClerkID, Name string
		CreatedAt, UpdatedAt string
	}
	err := s.db.Pool().QueryRow(r.Context(), query, input.Name, id, clerkID).
		Scan(&wl.ID, &wl.ClerkID, &wl.Name, &wl.CreatedAt, &wl.UpdatedAt)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         wl.ID,
		"clerk_id":   wl.ClerkID,
		"name":       wl.Name,
		"created_at": wl.CreatedAt,
		"updated_at": wl.UpdatedAt,
	})
}

func (s *Server) handleDeleteWatchlist(w http.ResponseWriter, r *http.Request) {
	clerkID := r.Context().Value("clerk_id").(string)
	id := extractPathID(r.URL.Path, "/api/watchlists/")

	query := `DELETE FROM watchlists WHERE id = $1 AND clerk_id = $2`
	result, err := s.db.Pool().Exec(r.Context(), query, id, clerkID)
	if err != nil || result.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAddTickerToWatchlist(w http.ResponseWriter, r *http.Request) {
	clerkID := r.Context().Value("clerk_id").(string)
	id := extractPathID(r.URL.Path, "/api/watchlists/")

	if !strings.HasSuffix(r.URL.Path, "/tickers") {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var input struct {
		TickerID string `json:"ticker_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `INSERT INTO watchlist_tickers (watchlist_id, ticker_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING RETURNING id, watchlist_id, ticker_id, added_at`

	var wt struct {
		ID, WatchlistID, TickerID string
		AddedAt string
	}
	err := s.db.Pool().QueryRow(r.Context(), query, id, input.TickerID).
		Scan(&wt.ID, &wt.WatchlistID, &wt.TickerID, &wt.AddedAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":           wt.ID,
		"watchlist_id": wt.WatchlistID,
		"ticker_id":    wt.TickerID,
		"added_at":     wt.AddedAt,
	})
}

func (s *Server) handleRemoveTickerFromWatchlist(w http.ResponseWriter, r *http.Request) {
	clerkID := r.Context().Value("clerk_id").(string)
	id := extractPathID(r.URL.Path, "/api/watchlists/")

	tickerID := r.URL.Query().Get("ticker_id")
	if tickerID == "" {
		http.Error(w, "ticker_id required", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM watchlist_tickers WHERE watchlist_id = $1 AND ticker_id = $2`
	result, err := s.db.Pool().Exec(r.Context(), query, id, tickerID)
	if err != nil || result.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListTickers(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	search := r.URL.Query().Get("search")

	var query string
	var args []interface{}

	if symbol != "" {
		query = `SELECT id, symbol, name, exchange, last_price, updated_at FROM tickers WHERE symbol = $1`
		args = append(args, strings.ToUpper(symbol))
	} else if search != "" {
		query = `SELECT id, symbol, name, exchange, last_price, updated_at FROM tickers
			WHERE symbol ILIKE $1 OR name ILIKE $1 LIMIT 50`
		args = append(args, "%"+search+"%")
	} else {
		query = `SELECT id, symbol, name, exchange, last_price, updated_at FROM tickers LIMIT 50`
	}

	rows, err := s.db.Pool().Query(r.Context(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tickers []map[string]interface{}
	for rows.Next() {
		var t struct {
			ID, Symbol, Name, Exchange string
			LastPrice float64
			UpdatedAt string
		}
		if err := rows.Scan(&t.ID, &t.Symbol, &t.Name, &t.Exchange, &t.LastPrice, &t.UpdatedAt); err != nil {
			continue
		}
		tickers = append(tickers, map[string]interface{}{
			"id":         t.ID,
			"symbol":     t.Symbol,
			"name":       t.Name,
			"exchange":   t.Exchange,
			"last_price": t.LastPrice,
			"updated_at": t.UpdatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tickers)
}

func (s *Server) handleGetTicker(w http.ResponseWriter, r *http.Request) {
	id := extractPathID(r.URL.Path, "/api/tickers/")

	query := `SELECT id, symbol, name, exchange, last_price, updated_at FROM tickers WHERE id = $1`
	var t struct {
		ID, Symbol, Name, Exchange string
		LastPrice float64
		UpdatedAt string
	}
	err := s.db.Pool().QueryRow(r.Context(), query, id).
		Scan(&t.ID, &t.Symbol, &t.Name, &t.Exchange, &t.LastPrice, &t.UpdatedAt)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         t.ID,
		"symbol":     t.Symbol,
		"name":       t.Name,
		"exchange":   t.Exchange,
		"last_price": t.LastPrice,
		"updated_at": t.UpdatedAt,
	})
}

func (s *Server) handleListJobs(w http.ResponseWriter, r *http.Request) {
	clerkID := r.Context().Value("clerk_id").(string)

	query := `SELECT id, clerk_id, type, status, payload, result, error, created_at, updated_at
		FROM jobs WHERE clerk_id = $1 ORDER BY created_at DESC LIMIT 50`

	rows, err := s.db.Pool().Query(r.Context(), query, clerkID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var jobs []map[string]interface{}
	for rows.Next() {
		var j struct {
			ID, ClerkID, Type, Status, Payload, Result, Error string
			CreatedAt, UpdatedAt string
		}
		if err := rows.Scan(&j.ID, &j.ClerkID, &j.Type, &j.Status, &j.Payload, &j.Result, &j.Error, &j.CreatedAt, &j.UpdatedAt); err != nil {
			continue
		}
		jobs = append(jobs, map[string]interface{}{
			"id":         j.ID,
			"clerk_id":   j.ClerkID,
			"type":       j.Type,
			"status":     j.Status,
			"payload":    j.Payload,
			"result":     j.Result,
			"error":      j.Error,
			"created_at": j.CreatedAt,
			"updated_at": j.UpdatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

func (s *Server) handleCreateJob(w http.ResponseWriter, r *http.Request) {
	clerkID := r.Context().Value("clerk_id").(string)

	var input struct {
		Type    string `json:"type"`
		Payload string `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `INSERT INTO jobs (clerk_id, type, status, payload) VALUES ($1, $2, 'pending', $3)
		RETURNING id, clerk_id, type, status, payload, result, error, created_at, updated_at`

	var j struct {
		ID, ClerkID, Type, Status, Payload, Result, Error string
		CreatedAt, UpdatedAt string
	}
	err := s.db.Pool().QueryRow(r.Context(), query, clerkID, input.Type, input.Payload).
		Scan(&j.ID, &j.ClerkID, &j.Type, &j.Status, &j.Payload, &j.Result, &j.Error, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.redis.LPush(r.Context(), "jobs:pending", j.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         j.ID,
		"clerk_id":   j.ClerkID,
		"type":       j.Type,
		"status":     j.Status,
		"payload":    j.Payload,
		"result":     j.Result,
		"error":      j.Error,
		"created_at": j.CreatedAt,
		"updated_at": j.UpdatedAt,
	})
}

func (s *Server) handleGetJob(w http.ResponseWriter, r *http.Request) {
	clerkID := r.Context().Value("clerk_id").(string)
	id := extractPathID(r.URL.Path, "/api/jobs/")

	query := `SELECT id, clerk_id, type, status, payload, result, error, created_at, updated_at
		FROM jobs WHERE id = $1 AND clerk_id = $2`
	var j struct {
		ID, ClerkID, Type, Status, Payload, Result, Error string
		CreatedAt, UpdatedAt string
	}
	err := s.db.Pool().QueryRow(r.Context(), query, id, clerkID).
		Scan(&j.ID, &j.ClerkID, &j.Type, &j.Status, &j.Payload, &j.Result, &j.Error, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         j.ID,
		"clerk_id":   j.ClerkID,
		"type":       j.Type,
		"status":     j.Status,
		"payload":    j.Payload,
		"result":     j.Result,
		"error":      j.Error,
		"created_at": j.CreatedAt,
		"updated_at": j.UpdatedAt,
	})
}

func (s *Server) handleCancelJob(w http.ResponseWriter, r *http.Request) {
	clerkID := r.Context().Value("clerk_id").(string)
	id := extractPathID(r.URL.Path, "/api/jobs/")

	if !strings.HasSuffix(r.URL.Path, "/cancel") {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	query := `UPDATE jobs SET status = 'cancelled' WHERE id = $1 AND clerk_id = $2 AND status = 'pending'
		RETURNING id, clerk_id, type, status, payload, result, error, created_at, updated_at`

	var j struct {
		ID, ClerkID, Type, Status, Payload, Result, Error string
		CreatedAt, UpdatedAt string
	}
	err := s.db.Pool().QueryRow(r.Context(), query, id, clerkID).
		Scan(&j.ID, &j.ClerkID, &j.Type, &j.Status, &j.Payload, &j.Result, &j.Error, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		http.Error(w, "not found or cannot cancel", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         j.ID,
		"clerk_id":   j.ClerkID,
		"type":       j.Type,
		"status":     j.Status,
		"payload":    j.Payload,
		"result":     j.Result,
		"error":      j.Error,
		"created_at": j.CreatedAt,
		"updated_at": j.UpdatedAt,
	})
}

func extractPathID(path, base string) string {
	id := strings.TrimPrefix(path, base)
	return strings.Split(id, "/")[0]
}