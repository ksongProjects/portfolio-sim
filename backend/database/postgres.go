package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	MaxConns int
}

type Postgres struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewPostgres(cfg Config) (*Postgres, error) {
	return NewPostgresWithLogger(cfg, nil)
}

func NewPostgresWithLogger(cfg Config, logger *slog.Logger) (*Postgres, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxConns)

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &Postgres{pool: pool, logger: logger}, nil
}

func (p *Postgres) Close() {
	p.pool.Close()
}

func (p *Postgres) Pool() *pgxpool.Pool {
	return p.pool
}

func (p *Postgres) log(ctx context.Context, level string, msg string, meta map[string]interface{}) {
	if p.logger == nil {
		return
	}
	args := []any{}
	for k, v := range meta {
		args = append(args, k, v)
	}
	switch level {
	case "ERROR":
		p.logger.Log(ctx, slog.LevelError, msg, args...)
	case "WARN":
		p.logger.Log(ctx, slog.LevelWarn, msg, args...)
	case "DEBUG":
		p.logger.Log(ctx, slog.LevelDebug, msg, args...)
	default:
		p.logger.Log(ctx, slog.LevelInfo, msg, args...)
	}
}

func (p *Postgres) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	start := time.Now()
	p.log(ctx, "DEBUG", "DB Query started", map[string]interface{}{
		"sql":       sql,
		"arg_count": len(args),
	})
	rows, err := p.pool.Query(ctx, sql, args...)
	duration := time.Since(start)
	if err != nil {
		p.log(ctx, "ERROR", "DB Query failed", map[string]interface{}{
			"sql":         sql,
			"arg_count":   len(args),
			"error":       err.Error(),
			"duration_ms": duration.Milliseconds(),
		})
		return nil, err
	}
	p.log(ctx, "DEBUG", "DB Query completed", map[string]interface{}{
		"sql":         sql,
		"duration_ms": duration.Milliseconds(),
	})
	return rows, nil
}

func (p *Postgres) QueryRow(ctx context.Context, sql string, args ...interface{}) (pgx.Row, error) {
	start := time.Now()
	p.log(ctx, "DEBUG", "DB QueryRow started", map[string]interface{}{
		"sql":       sql,
		"arg_count": len(args),
	})
	row := p.pool.QueryRow(ctx, sql, args...)
	duration := time.Since(start)
	p.log(ctx, "DEBUG", "DB QueryRow completed", map[string]interface{}{
		"sql":         sql,
		"duration_ms": duration.Milliseconds(),
	})
	return row, nil
}

func (p *Postgres) Exec(ctx context.Context, sql string, args ...interface{}) (int64, error) {
	start := time.Now()
	p.log(ctx, "DEBUG", "DB Exec started", map[string]interface{}{
		"sql":       sql,
		"arg_count": len(args),
	})
	tag, err := p.pool.Exec(ctx, sql, args...)
	duration := time.Since(start)
	if err != nil {
		p.log(ctx, "ERROR", "DB Exec failed", map[string]interface{}{
			"sql":         sql,
			"arg_count":   len(args),
			"error":       err.Error(),
			"duration_ms": duration.Milliseconds(),
		})
		return 0, err
	}
	p.log(ctx, "DEBUG", "DB Exec completed", map[string]interface{}{
		"sql":           sql,
		"rows_affected": tag.RowsAffected(),
		"duration_ms":   duration.Milliseconds(),
	})
	return tag.RowsAffected(), nil
}

func (p *Postgres) GetPool() *pgxpool.Pool {
	return p.pool
}
