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

	dbHost := os.Getenv("DATABASE_HOST")
	dbPort := os.Getenv("DATABASE_PORT")
	dbUser := os.Getenv("DATABASE_USER")
	dbPassword := os.Getenv("DATABASE_PASSWORD")
	dbName := os.Getenv("DATABASE_NAME")
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)
	}

	return &Config{
		GeminiAPIKey:      os.Getenv("GEMINI_API_KEY"),
		YouTubeAPIKey:     os.Getenv("YOUTUBE_API_KEY"),
		DatabaseURL:       dbURL,
		RedisAddr:         fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Port:              port,
		ScrapeIntervalMin: scrapeInterval,
	}, nil
}
