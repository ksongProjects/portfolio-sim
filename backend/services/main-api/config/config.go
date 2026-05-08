package config

import (
	"github.com/portfolio-sim/backend/internal/config"
)

type Config struct {
	*config.Config
	GRPCPort int
}

func Load() *Config {
	return &Config{
		Config:   config.Load(),
		GRPCPort: 50051,
	}
}