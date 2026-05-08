package config

import "os"

type Config struct {
	GeminiAPIKey  string
	YouTubeAPIKey string
	DatabaseURL   string
	RedisAddr     string
	Port          string
}

func Load() (*Config, error) {
	return &Config{
		GeminiAPIKey:  os.Getenv("GEMINI_API_KEY"),
		YouTubeAPIKey: os.Getenv("YOUTUBE_API_KEY"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		Port:          os.Getenv("PORT"),
	}, nil
}
