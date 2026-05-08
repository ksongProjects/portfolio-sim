package database

import "github.com/portfolio-sim/backend/internal/database"

type Config = database.Config
type Postgres = database.Postgres

func NewPostgres(cfg Config) (*Postgres, error) {
	return database.NewPostgres(database.Config{
		Host:     cfg.Host,
		Port:     cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
		DBName:   cfg.DBName,
		MaxConns: cfg.MaxConns,
	})
}
