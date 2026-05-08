package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/portfolio-sim/logging-service/config"
	"github.com/portfolio-sim/logging-service/database"
	"github.com/portfolio-sim/logging-service/logging"
	"github.com/portfolio-sim/logging-service/models"
	"github.com/portfolio-sim/logging-service/redis"
	"github.com/portfolio-sim/logging-service/sse"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LoggingServiceServer struct {
	models.UnimplementedLoggingServiceServer
	db     *database.Postgres
	redis  *redis.Client
	sseMgr *sse.Manager
}

func NewLoggingServiceServer(db *database.Postgres, r *redis.Client, sseMgr *sse.Manager) *LoggingServiceServer {
	return &LoggingServiceServer{
		db:     db,
		redis:  r,
		sseMgr: sseMgr,
	}
}

func (s *LoggingServiceServer) EmitLog(ctx context.Context, req *models.EmitLogRequest) (*models.EmitLogResponse, error) {
	if req.Level == "" || req.Service == "" || req.Message == "" {
		return nil, status.Error(codes.InvalidArgument, "level, service, and message are required")
	}

	validLevels := map[string]bool{"DEBUG": true, "INFO": true, "WARN": true, "ERROR": true, "FATAL": true}
	if !validLevels[req.Level] {
		return nil, status.Error(codes.InvalidArgument, "invalid log level")
	}

	id := uuid.New()
	timestamp := time.Now().UTC()

	metadataJSON, err := json.Marshal(req.Metadata)
	if err != nil {
		metadataJSON = []byte("{}")
	}

	_, err = s.db.Pool().Exec(ctx, `
		INSERT INTO logs (id, timestamp, level, service, component, message, metadata, trace_id, span_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, id, timestamp, req.Level, req.Service, req.Component, req.Message, metadataJSON, req.TraceID, req.SpanID)

	if err != nil {
		if pgxErr, ok := err.(pgx.SqlErr); ok && pgxErr.Code == "42P01" {
			partitionDate := timestamp.Truncate(24 * time.Hour)
			_, createErr := s.db.Pool().Exec(ctx, "SELECT create_monthly_partition('logs', $1)", partitionDate)
			if createErr == nil {
				_, err = s.db.Pool().Exec(ctx, `
					INSERT INTO logs (id, timestamp, level, service, component, message, metadata, trace_id, span_id)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				`, id, timestamp, req.Level, req.Service, req.Component, req.Message, metadataJSON, req.TraceID, req.SpanID)
			}
		}
		if err != nil {
			return nil, status.Errorf(codes.Internal, "insert log: %v", err)
		}
	}

	partitionDate := timestamp.Truncate(24 * time.Hour)
	go s.db.Exec(context.Background(), "SELECT create_monthly_partition('logs', $1)", partitionDate)

	logEntry := &models.Log{
		ID:        id,
		Timestamp: timestamp,
		Level:     req.Level,
		Service:   req.Service,
		Component: req.Component,
		Message:   req.Message,
		Metadata:  req.Metadata,
		TraceID:   req.TraceID,
		SpanID:    req.SpanID,
		CreatedAt: timestamp,
	}

	logJSON, _ := json.Marshal(logEntry)
	s.redis.Publish(ctx, "logs:"+req.Service, string(logJSON))

	s.notifySSE(logEntry)

	return &models.EmitLogResponse{
		ID:        id.String(),
		Timestamp: timestamp.Format(time.RFC3339Nano),
	}, nil
}

func (s *LoggingServiceServer) notifySSE(entry *models.Log) {
	if s.sseMgr == nil {
		return
	}
}

type LogEmitterImpl struct {
	server *LoggingServiceServer
}

func NewLogEmitter(s *LoggingServiceServer) logging.LogEmitter {
	return &LogEmitterImpl{server: s}
}

func (e *LogEmitterImpl) Emit(ctx context.Context, level, service, component, message string, metadata map[string]interface{}, traceID, spanID string) error {
	req := &models.EmitLogRequest{
		Level:     level,
		Service:   service,
		Component: component,
		Message:   message,
		Metadata:  metadata,
		TraceID:   traceID,
		SpanID:    spanID,
	}
	_, err := e.server.EmitLog(ctx, req)
	return err
}

type LogWriterImpl struct {
	emitter logging.LogEmitter
	service string
	level   string
}

func NewLogWriter(emitter logging.LogEmitter, service, level string) *LogWriterImpl {
	return &LogWriterImpl{emitter: emitter, service: service, level: level}
}

func (w *LogWriterImpl) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	ctx := context.Background()
	metadata := map[string]interface{}{}
	if json.Valid(p) {
		var m map[string]interface{}
		if json.Unmarshal(p, &m) == nil {
			metadata = m
		}
	}
	err = w.emitter.Emit(ctx, w.level, w.service, "", string(p), metadata, "", "")
	if err != nil {
		return 0, err
	}
	return len(p), nil
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

	sseMgr := sse.NewManager(redisClient)

	server := NewLoggingServiceServer(db, redisClient, sseMgr)
	emitter := NewLogEmitter(server)
	logger := logging.NewLogger(emitter)
	_ = logger

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	models.RegisterLoggingServiceServer(grpcServer, server)

	log.Printf("Logging service listening on gRPC port %d", cfg.Server.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}