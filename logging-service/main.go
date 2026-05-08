package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/portfolio-sim/logging-service/config"
	"github.com/portfolio-sim/logging-service/database"
	"github.com/portfolio-sim/logging-service/redis"
)

type LogEntry struct {
	ID        string                 `json:"id"`
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Service   string                 `json:"service"`
	Component string                 `json:"component"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata"`
	TraceID   string                 `json:"trace_id"`
	SpanID    string                 `json:"span_id"`
}

type EmitLogRequest struct {
	Level     string                 `json:"level"`
	Service   string                 `json:"service"`
	Component string                 `json:"component"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata"`
	TraceID   string                 `json:"trace_id"`
	SpanID    string                 `json:"span_id"`
}

type LoggingService struct {
	db    *database.Postgres
	redis *redis.Client
}

func NewLoggingService(db *database.Postgres, r *redis.Client) *LoggingService {
	return &LoggingService{db: db, redis: r}
}

func (s *LoggingService) HandleEmitLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req EmitLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Level == "" || req.Service == "" || req.Message == "" {
		http.Error(w, "level, service, and message are required", http.StatusBadRequest)
		return
	}

	validLevels := map[string]bool{"DEBUG": true, "INFO": true, "WARN": true, "ERROR": true, "FATAL": true}
	if !validLevels[req.Level] {
		http.Error(w, "invalid log level", http.StatusBadRequest)
		return
	}

	id := uuid.New()
	timestamp := time.Now().UTC()

	metadataJSON, _ := json.Marshal(req.Metadata)

	_, err := s.db.Pool().Exec(r.Context(), `
		INSERT INTO logs (id, timestamp, level, service, component, message, metadata, trace_id, span_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, id, timestamp, req.Level, req.Service, req.Component, req.Message, metadataJSON, req.TraceID, req.SpanID)

	if err != nil {
		log.Printf("Failed to insert log: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	logJSON, _ := json.Marshal(LogEntry{
		ID:        id.String(),
		Timestamp: timestamp.Format(time.RFC3339Nano),
		Level:     req.Level,
		Service:   req.Service,
		Component: req.Component,
		Message:   req.Message,
		Metadata:  req.Metadata,
		TraceID:   req.TraceID,
		SpanID:    req.SpanID,
	})
	s.redis.Publish(r.Context(), "logs:"+req.Service, string(logJSON))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": id.String()})
}

func (s *LoggingService) HandleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	cfg := config.Load()

	db, err := database.NewPostgres(database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		MaxConns: cfg.Database.MaxConns,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	redisClient, err := redis.NewClient(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	service := NewLoggingService(db, redisClient)

	http.HandleFunc("GET /health", service.HandleHealth)
	http.HandleFunc("POST /api/logs", service.HandleEmitLog)

	addr := cfg.Server.Addr()
	log.Printf("Logging service listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
