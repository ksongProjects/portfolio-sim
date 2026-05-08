package config

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
	APIKey         string
	APISecret      string
	RefreshToken   string
	AccountID      string
	RateLimitMin   int
}

type PolygonConfig struct {
	APIKey       string
	RateLimitMin int
}

type FMPConfig struct {
	APIKey       string
	RateLimitMin int
}
