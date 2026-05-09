package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Questrade QuestradeConfig
	Polygon  PolygonConfig
	FMP      FMPConfig
}

type ServerConfig struct {
	HTTPPort int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	MaxConns int
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
}

type QuestradeConfig struct {
	RefreshToken string
	APIURL       string
	AccountID    string
	RateLimitMin int
}

type PolygonConfig struct {
	APIKey       string
	RateLimitMin int
}

type FMPConfig struct {
	APIKey       string
	RateLimitMin int
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			HTTPPort: getEnvInt("SERVER_HTTP_PORT", 8080),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DATABASE_HOST", "localhost"),
			Port:     getEnvInt("DATABASE_PORT", 5432),
			User:     getEnv("DATABASE_USER", "postgres"),
			Password: getEnv("DATABASE_PASSWORD", "postgres"),
			DBName:   getEnv("DATABASE_NAME", "portfolio_sim"),
			MaxConns: getEnvInt("DATABASE_MAX_CONNS", 20),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
		},
		Questrade: QuestradeConfig{
			APIURL:       "https://login.questrade.com/oauth2/token",
			RateLimitMin: 100,
		},
		Polygon: PolygonConfig{
			APIKey:       getEnv("POLYGON_API_KEY", ""),
			RateLimitMin: getEnvInt("POLYGON_RATE_LIMIT", 60),
		},
		FMP: FMPConfig{
			APIKey:       getEnv("FMP_API_KEY", ""),
			RateLimitMin: getEnvInt("FMP_RATE_LIMIT", 250),
		},
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}
