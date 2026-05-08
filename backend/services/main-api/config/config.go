package config

import "github.com/portfolio-sim/backend/internal/config"

type Config struct {
	*config.Config
	GRPCPort int
}
