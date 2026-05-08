package config

import (
	"fmt"
	"os"
)

type Config struct {
	GeminiAPIKey      string
	YouTubeAPIKey     string
	DatabaseURL       string
	RedisAddr         string
	Port              string
	ScrapeIntervalMin int
}

func Load() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	scrapeInterval := 15
	if v := os.Getenv("SCRAPE_INTERVAL_MIN"); v != "" {
		fmt.Sscanf(v, "%d", &scrapeInterval)
	}
	return &Config{
		GeminiAPIKey:      os.Getenv("GEMINI_API_KEY"),
		YouTubeAPIKey:     os.Getenv("YOUTUBE_API_KEY"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		RedisAddr:         fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Port:              port,
		ScrapeIntervalMin: scrapeInterval,
	}, nil
}
