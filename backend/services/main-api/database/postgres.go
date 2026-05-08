package database

import (
	"context"
)

type Postgres struct {
	pool interface {
		Query(ctx context.Context, sql string, args ...interface{}) (interface{}, error)
		QueryRow(ctx context.Context, sql string, args ...interface{}) interface {
			Scan(dest ...interface{}) error
		}
		Exec(ctx context.Context, sql string, args ...interface{}) (interface{}, error)
		Close()
	}
}

func (p *Postgres) Close() {}

func (p *Postgres) Pool() interface{} {
	return p.pool
}
