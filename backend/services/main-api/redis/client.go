package redis

import "github.com/portfolio-sim/backend/internal/redis"

type Client = redis.Client

func NewClient(host string, port int, password string) (*Client, error) {
	return redis.NewClient(host, port, password)
}
