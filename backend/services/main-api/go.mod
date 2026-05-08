module github.com/portfolio-sim/backend/services/main-api

go 1.22

require (
	github.com/jackc/pgx/v5 v5.5.3
	github.com/redis/go-redis/v9 v9.4.0
	golang.org/x/net v0.20.0
)

replace github.com/portfolio-sim/backend => ../../