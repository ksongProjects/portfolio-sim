module github.com/portfolio-sim/backend/services/market-data-service

go 1.22

require (
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.5.3
	github.com/redis/go-redis/v9 v9.4.0
	github.com/portfolio-sim/backend v0.0.0
)

replace github.com/portfolio-sim/backend => ../..
