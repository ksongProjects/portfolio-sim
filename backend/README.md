# Main API Service

The central API gateway for the portfolio simulation platform, handling portfolio management, provider configuration, and observability.

## Overview

Written in Go using the standard `net/http` library. Acts as the single entry point for the frontend, aggregating data from various services and the database.

## Configuration

Environment variables (with defaults):

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_HTTP_PORT` | 8080 | HTTP listen port |
| `SERVER_READ_TIMEOUT` | 15s | HTTP read timeout |
| `SERVER_WRITE_TIMEOUT` | 15s | HTTP write timeout |
| `SERVER_SHUTDOWN_TIMEOUT` | 30s | Graceful shutdown timeout |
| `DATABASE_HOST` | localhost | PostgreSQL host |
| `DATABASE_PORT` | 5432 | PostgreSQL port |
| `DATABASE_USER` | postgres | PostgreSQL user |
| `DATABASE_PASSWORD` | postgres | PostgreSQL password |
| `DATABASE_NAME` | portfolio_sim | Database name |
| `DATABASE_MAX_CONNS` | 20 | Max database connections |
| `REDIS_HOST` | localhost | Redis host |
| `REDIS_PORT` | 6379 | Redis port |
| `REDIS_PASSWORD` | (empty) | Redis password |

## Endpoints

### Health & Observability

#### `GET /health`
Health check endpoint.

**Response:**
```json
{"status": "ok"}
```

#### `GET /api/observability/services`
Returns health status of all services.

**Response:**
```json
[
  {
    "name": "main-api",
    "status": "healthy",
    "uptime": "100%",
    "last_check": "Just now"
  },
  {
    "name": "logging-service",
    "status": "healthy",
    "uptime": "100%",
    "last_check": "Just now"
  }
]
```

#### `GET /api/observability/logs`
Fetches logs from the logging service.

**Query Parameters:**
- `limit` (optional): Max logs to return (default: 50, max: 500)

**Response:**
```json
[
  {
    "id": "uuid",
    "timestamp": "2026-05-09T01:30:00Z",
    "level": "INFO",
    "service": "market-data-service",
    "component": "priceFetcher",
    "message": "Price updated for AAPL",
    "metadata": {},
    "trace_id": null,
    "span_id": null
  }
]
```

---

### Portfolio Management

#### `GET /api/portfolio/positions`
Returns positions for a portfolio.

**Query Parameters:**
- `portfolio_id` (optional): Portfolio ID (default: "default")

**Response:**
```json
[
  {
    "ID": "pos-uuid",
    "PortfolioID": "default",
    "TickerID": "ticker-uuid",
    "Symbol": "AAPL",
    "CompanyName": "Apple Inc.",
    "Quantity": 100,
    "AvgCost": 150.00,
    "CurrentPrice": 175.00,
    "CurrentValue": 17500.00,
    "DayChange": 1.75,
    "DayChangePct": 1.0,
    "TotalGain": 2500.00,
    "TotalGainPct": 16.67,
    "OpenedAt": "2024-01-15T10:30:00Z"
  }
]
```

#### `GET /api/portfolio/summary`
Returns aggregated portfolio summary.

**Query Parameters:**
- `portfolio_id` (optional): Portfolio ID (default: "default")

**Response:**
```json
{
  "TotalValue": 125000.00,
  "DayChange": 1250.00,
  "DayChangePct": 1.01,
  "TotalInvested": 100000.00,
  "TotalGain": 25000.00,
  "TotalGainPct": 25.00,
  "CashBalance": 50000.00
}
```

---

### Market Data

#### `GET /api/market/indices`
Returns market index data (SPY, QQQ, DIA, IWM, VIX, DXY).

**Response:**
```json
[
  {
    "Symbol": "SPY",
    "Name": "S&P 500 ETF",
    "Price": 520.50,
    "Change": 2.30,
    "ChangePct": 0.44
  }
]
```

---

### News

#### `GET /api/news`
Returns news articles with sentiment.

**Query Parameters:**
- `limit` (optional): Number of articles (default: 20)

**Response:**
```json
[
  {
    "ID": "article-uuid",
    "Title": "Apple reports record earnings",
    "Source": "Reuters",
    "URL": "https://example.com/article",
    "Summary": "Apple Inc. reported record Q2 earnings...",
    "Sentiment": "bullish",
    "PublishedAt": "2026-05-09T10:00:00Z",
    "TickerSymbols": ["AAPL"]
  }
]
```

---

### Strategies

#### `GET /api/strategies`
Returns configured trading strategies.

**Response:**
```json
[
  {
    "ID": "strategy-uuid",
    "Name": "Momentum Growth",
    "Status": "active",
    "Returns": 18.4,
    "Sharpe": 1.42,
    "MaxDD": -12.3,
    "Trades": 847,
    "WinRate": 64.2
  }
]
```

#### `GET /api/signals`
Returns recent trading signals.

**Query Parameters:**
- `limit` (optional): Number of signals (default: 10)

**Response:**
```json
[
  {
    "ID": "signal-uuid",
    "Strategy": "Momentum Growth",
    "Symbol": "NVDA",
    "Action": "BUY",
    "Price": 950.00,
    "Confidence": "HIGH",
    "Timestamp": "2026-05-09T10:30:00Z"
  }
]
```

---

### Providers

#### `GET /api/providers`
Returns configured data providers.

**Response:**
```json
[
  {
    "id": "polygon",
    "provider_id": "polygon",
    "name": "Polygon.io",
    "description": "Real-time and historical market data",
    "type": "market_data",
    "api_key_set": true,
    "is_connected": true,
    "rate_limit": 60,
    "docs_url": "https://polygon.io/docs"
  }
]
```

#### `PUT /api/providers`
Saves or updates a provider API key.

**Request:**
```json
{
  "provider_id": "polygon",
  "api_key": "your-api-key"
}
```

**Response:**
```json
{"status": "saved"}
```

#### `POST /api/providers/validate`
Validates a provider API key.

**Request:**
```json
{
  "provider_id": "polygon",
  "api_key": "your-api-key"
}
```

**Response:**
```json
{"valid": true}
```

---

### Connections

#### `GET /api/connections`
Returns connection status for infrastructure services.

**Response:**
```json
[
  {
    "id": "postgres",
    "name": "PostgreSQL",
    "type": "database",
    "is_up": true,
    "latency_ms": 2
  },
  {
    "id": "redis",
    "name": "Redis",
    "type": "cache",
    "is_up": true,
    "latency_ms": 1
  }
]
```

---

### RSS Feeds

#### `GET /api/rss-feeds`
Returns configured RSS feeds.

**Response:**
```json
[
  {
    "id": "feed-uuid",
    "name": "Reuters Markets",
    "url": "https://feeds.reuters.com/reuters/markets",
    "scrape_interval_min": 15,
    "last_scrape_at": "2026-05-09T10:00:00Z",
    "is_active": true
  }
]
```

#### `POST /api/rss-feeds`
Adds a new RSS feed.

**Request:**
```json
{
  "name": "Reuters Markets",
  "url": "https://feeds.reuters.com/reuters/markets",
  "scrape_interval_min": 15
}
```

**Response:** Returns updated list of feeds.

#### `DELETE /api/rss-feeds?id={feed_id}`
Deletes an RSS feed.

**Response:**
```json
{"status": "deleted"}
```

---

## Architecture

The main API handles routing via `http.DefaultServeMux`:

```go
http.HandleFunc("GET /health", s.handleHealth)
http.HandleFunc("GET /api/observability/services", s.handleGetServices)
http.HandleFunc("GET /api/observability/logs", s.handleGetLogs)
http.HandleFunc("GET /api/portfolio/positions", s.handleGetPositions)
http.HandleFunc("GET /api/portfolio/summary", s.handleGetPortfolioSummary)
http.HandleFunc("GET /api/market/indices", s.handleGetMarketIndices)
http.HandleFunc("GET /api/news", s.handleGetNews)
http.HandleFunc("GET /api/strategies", s.handleGetStrategies)
http.HandleFunc("GET /api/signals", s.handleGetSignals)
http.HandleFunc("GET /api/providers", s.handleGetProviders)
http.HandleFunc("PUT /api/providers", s.handleUpdateProvider)
http.HandleFunc("POST /api/providers/validate", s.handleValidateProvider)
http.HandleFunc("GET /api/connections", s.handleGetConnections)
http.HandleFunc("GET /api/rss-feeds", s.handleGetRSSFeeds)
http.HandleFunc("POST /api/rss-feeds", s.handleAddRSSFeed)
http.HandleFunc("DELETE /api/rss-feeds", s.handleDeleteRSSFeed)
```

## Services

| Service | Description |
|---------|-------------|
| `PortfolioService` | Position and portfolio management, market indices |
| `ProviderService` | Provider configuration, API key validation, RSS feeds |
| `ObservabilityService` | Health checks for all microservices |

## Database Schema

Key tables: `positions`, `portfolios`, `tickers`, `normalized_prices`, `data_sources`, `provider_configurations`, `rss_feeds`, `news_articles`, `simulation_runs`, `logs`

## Authentication

Currently no authentication (open CORS). API keys for providers are stored encrypted in `provider_configurations` table.