# Portfolio Simulation Platform

A real-time portfolio simulation and analysis platform with market data aggregation, automated trading strategies, and comprehensive observability.

## Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                           Frontend (Next.js)                        │
│         Dashboard | Portfolio | Strategy | News | Observability      │
└──────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌──────────────────────────────────────────────────────────────────────┐
│                         Main API (Go) - :8080                        │
│    Portfolio │ Market Data │ Strategies │ Providers │ Observability   │
└──────────────────────────────────────────────────────────────────────┘
         │            │            │            │
         ▼            ▼            ▼            ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│    Market   │ │    News     │ │   Analyst  │ │   Logging  │
│ Data Service│ │ Feed Service│ │  Service   │ │   Service  │
│   :8081     │ │   :8082     │ │   :8083    │ │   :9090    │
└─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘
       │               │              │              │
       └───────────────┴──────────────┴──────────────┘
                          │
                          ▼
              ┌───────────────────────┐
              │    PostgreSQL :5432   │
              │    Redis :6379        │
              └───────────────────────┘
```

## Services

| Service | Port | Description |
|---------|------|-------------|
| [Main API](./backend/README.md) | 8080 | Central API gateway, portfolio management, provider configuration |
| [Market Data Service](./market-data-service/README.md) | 8081 | Real-time price aggregation from multiple providers |
| [News Feed Service](./news-feed-service/README.md) | 8082 | RSS scraping, YouTube transcription, AI summarization |
| [Analyst Service](./analyst-service/README.md) | 8083 | Quantitative analysis, backtesting, execution optimization |
| [Logging Service](./logging-service/README.md) | 9090 | Centralized logging with structured storage and querying |

## Features

### Portfolio Management
- Real-time position tracking with P&L calculations
- Multi-provider market data aggregation (Polygon, Questrade, FMP)
- Cash balance and buying power monitoring
- Historical performance metrics

### Trading Strategies
- Strategy configuration and monitoring
- Signal generation and tracking
- Backtest result storage and visualization
- Sharpe ratio, win rate, and drawdown metrics

### Market Data
- Normalized price data from multiple providers
- Real-time SSE streaming for live ticker updates
- Intraday bar storage and backfill capabilities
- Option chain data aggregation

### News & Sentiment
- RSS feed aggregation from multiple sources
- YouTube video transcription
- AI-powered summarization via Google Gemini
- Sentiment analysis and ticker symbol extraction

### Quantitative Analysis
- Multi-factor alpha engine
- Statistical arbitrage (pairs trading)
- Portfolio construction with long/short books
- Stress testing and VaR/CVaR calculations
- Execution optimization (TWAP/VWAP)

### Observability
- Service health monitoring
- Structured JSON logging
- Real-time log streaming via Redis pub/sub
- Centralized log aggregation and querying

## Quick Start

```bash
# Start all services
docker-compose up -d

# Check service health
curl http://localhost:8080/api/observability/services

# View logs
curl http://localhost:8080/api/observability/logs?limit=50
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_HOST` | PostgreSQL host | localhost |
| `DATABASE_PORT` | PostgreSQL port | 5432 |
| `DATABASE_USER` | PostgreSQL user | postgres |
| `DATABASE_PASSWORD` | PostgreSQL password | postgres |
| `DATABASE_NAME` | Database name | portfolio_sim |
| `REDIS_HOST` | Redis host | localhost |
| `REDIS_PORT` | Redis port | 6379 |

### Provider API Keys

| Provider | Environment Variable |
|----------|---------------------|
| Polygon.io | `POLYGON_API_KEY` |
| Questrade | `QUESTRADE_API_KEY` |
| Financial Modeling Prep | `FMP_API_KEY` |
| YouTube Data API | `YOUTUBE_API_KEY` |
| Google Gemini | `GEMINI_API_KEY` |

## Tech Stack

- **Frontend**: Next.js 16, React 19, TanStack Table, shadcn/ui
- **Backend**: Go, net/http, pgx (PostgreSQL driver)
- **Services**: Go (market-data, logging), Python/FastAPI (analyst), Python (news-feed)
- **Databases**: PostgreSQL 16, Redis 7
- **Containerization**: Docker, Docker Compose

## Project Structure

```
portfolio-sim/
├── backend/              # Main API service (Go)
├── market-data-service/  # Market data aggregation (Go)
├── news-feed-service/    # News scraping and processing (Python)
├── analyst-service/      # Quantitative analysis (Python/FastAPI)
├── logging-service/      # Centralized logging (Go)
├── frontend/             # Next.js web application
├── docker-compose.yml    # Service orchestration
└── README.md            # This file
```