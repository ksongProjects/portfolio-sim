package database

import "github.com/jackc/pgx/v5/pgxpool"

type Config struct {
	ConnString string
}

func (c *Config) Pool() (*pgxpool.Pool, error) {
	return pgxpool.New(context.Background(), c.ConnString)
}
