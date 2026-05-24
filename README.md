# Portfolio Simulation Platform

A real-time portfolio simulation and market research platform. The app combines a Next.js frontend, a Go API gateway, market data and news microservices, a FastAPI quant service, PostgreSQL, and Redis.

## Architecture

```text
Browser
  |
  v
Frontend, Next.js on :3000
  |
  v
Main API, Go on :8080
  |-- portfolio, providers, settings, observability, notifications
  |-- log ingestion at POST /api/logs
  |-- SSE market stream at GET /api/stream/market
  |
  |--> Market Data Service, Go on :8081
  |      ticker search, quotes, details, intraday bars, ratios, backfill queues
  |
  |--> News Feed Service, Go on :8082
  |      RSS scraping, YouTube channel/video lookup, Gemini analysis
  |
  |--> Analyst Service, Python/FastAPI on :8083
         factor analysis, portfolio construction, stat arb, PIT filtering,
         stress testing, execution analysis, market microstructure

PostgreSQL 16 on :5433
Redis 7 on :6379
```

## Services

| Service | Host Port | Runtime | Description |
| --- | ---: | --- | --- |
| [Main API](./backend/README.md) | 8080 | Go | API gateway, portfolio management, provider configuration, settings, notifications, observability, and log storage |
| [Market Data Service](./market-data-service/README.md) | 8081 | Go | Provider-backed market data, ticker lookup, quotes, intraday bars, financial ratios, Redis streams |
| [News Feed Service](./news-feed-service/README.md) | 8082 | Go | RSS scraping, YouTube channel/video workflows, Gemini-powered article and video analysis |
| [Analyst Service](./analyst-service/README.md) | 8083 | Python/FastAPI | Quant endpoints and background workers for backtests and option Greek jobs |
| Frontend | 3000 | Next.js | Dashboard, portfolio, news, strategy, observability, settings, and ticker detail views |
| PostgreSQL | 5433 | Postgres 16 | Application data, provider credentials, logs, market data, and migrations |
| Redis | 6379 | Redis 7 | Streams, pub/sub, and background queues |

## Features

### Frontend

- **Dashboard** — portfolio value, P&L, top holdings, recent activity, and live market indices via SSE.
- **Portfolio** — add/remove positions with ticker search; current prices update live from the market stream.
- **Ticker Detail** — company profile, intraday bars, financial ratios, market stats, and WebSocket-powered live price.
- **News Feed** — RSS articles, YouTube channel search, latest videos, and manual video analysis with Gemini.
- **Strategy / Signals** — views backed by the Main API.
- **Observability** — service health, structured logs, and route-level filtering.
- **Settings** — provider credentials, Questrade OAuth, RSS feeds, market index configuration.

### Backend and Services

- Main API endpoints for portfolio positions, summary data, market indices, news, strategies, signals, providers, notifications, RSS feeds, ticker proxying, videos, logs, and SSE market streams.
- Provider credential validation and encrypted storage for Massive, Questrade, FMP, YouTube, and Gemini.
- Market data provider orchestration with provider priority by operation.
- Redis queues for market-data backfills, ticker subscriptions, RSS scraping, transcription jobs, backtests, and option Greek jobs.
- News articles and YouTube analysis unified in `news_articles` with source type, ticker extraction, sentiment, content, and channel metadata.
- Quant API routes under `/api/v1/quant` for factor scorecards, portfolio books, stat arb, point-in-time filters, stress tests, execution costs, and microstructure state.

## Quick Start

Start the backend services, database, and Redis:

```bash
docker compose up -d --build
```

Check the API and service status:

```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/observability/services
curl "http://localhost:8080/api/observability/logs?limit=50"
```

Run the frontend separately:

```bash
cd frontend
pnpm install
pnpm dev
```

Open [http://localhost:3000](http://localhost:3000). The frontend defaults to `http://localhost:8080` for the API. Override it with `NEXT_PUBLIC_API_URL` if the Main API runs somewhere else.

## Common Commands

```bash
# Start or rebuild all Docker services
docker compose up -d --build

# Follow one service's logs
docker compose logs -f main-api

# Stop containers without deleting database data
docker compose down

# Frontend checks
cd frontend
pnpm lint
pnpm build

# Backend tests
cd backend
go test ./...
```

## Configuration

Docker Compose starts the backend services, PostgreSQL, and Redis for local development. PostgreSQL is exposed on host port `5433` to avoid conflicts with a local Postgres on `5432`; containers still use `DATABASE_PORT=5432`.

| Variable | Used By | Default / Local Value | Description |
| --- | --- | --- | --- |
| `SERVER_HTTP_PORT` | Go services | `8080` | Internal HTTP listen port |
| `PORT` | News Feed Service | `8080` | News service HTTP listen port |
| `DATABASE_HOST` | API/services | `localhost`, `postgres` in Compose | PostgreSQL host |
| `DATABASE_PORT` | API/services | `5432` | PostgreSQL port inside Compose |
| `DATABASE_USER` | API/services | `postgres` | PostgreSQL user |
| `DATABASE_PASSWORD` | API/services | `postgres` | PostgreSQL password |
| `DATABASE_NAME` | API/services | `portfolio_sim` | PostgreSQL database |
| `DATABASE_MAX_CONNS` | Go services | `20` | PostgreSQL connection pool size |
| `DATABASE_URL` | News Feed Service | built from database fields | Optional Postgres URL override |
| `REDIS_HOST` | API/services | `localhost`, `redis` in Compose | Redis host |
| `REDIS_PORT` | API/services | `6379` | Redis port |
| `REDIS_PASSWORD` | API/services | empty | Redis password |
| `LOGGING_SERVICE_URL` | services | `http://main-api:8080/api/logs` | Structured log ingestion endpoint |
| `MARKET_DATA_SERVICE_URL` | Main API | `http://market-data-service:8080` | Market data service base URL |
| `NEWS_FEED_SERVICE_URL` | Main API | `http://localhost:8082` | News feed service base URL; use `http://news-feed-service:8080` for container-to-container access |
| `PROVIDER_SECRET_KEY` | API/services | falls back to DB password | Encryption key for stored provider credentials |
| `SCRAPE_INTERVAL_MIN` | News Feed Service | `15` | RSS scrape scheduler interval |
| `NEXT_PUBLIC_API_URL` | Frontend | `http://localhost:8080` | Browser API base URL |

## Provider Keys

The Settings page is the normal way to validate and save provider credentials. Saved keys and OAuth tokens are encrypted in `provider_configurations`.

| Provider | Provider ID | Optional env fallback |
| --- | --- | --- |
| Massive | `massive` | `MASSIVE_API_KEY` |
| Questrade | `questrade` | OAuth refresh token saved through Settings |
| Financial Modeling Prep | `fmp` | `FMP_API_KEY` |
| YouTube Data API | `youtube` | `YOUTUBE_API_KEY` |
| Google Gemini | `gemini` | `GEMINI_API_KEY` |

## API Surface

Main API highlights:

```text
GET  /health
POST /api/logs
GET  /api/observability/services
GET  /api/observability/logs
GET  /api/portfolio/positions
POST /api/portfolio/positions
DELETE /api/portfolio/positions?portfolio_id=&position_id=
GET  /api/portfolio/summary
GET  /api/portfolio/performance
GET  /api/market/indices
GET  /api/settings/market-indices
PUT  /api/settings/market-indices
GET  /api/news
GET  /api/strategies
GET  /api/signals
GET  /api/notifications
POST /api/notifications/dismiss
GET  /api/providers
PUT  /api/providers
POST /api/providers/validate
GET  /api/providers/questrade/oauth
POST /api/providers/questrade/oauth
POST /api/providers/questrade/refresh
GET  /api/rss-feeds
POST /api/rss-feeds
DELETE /api/rss-feeds?id={feed_id}
POST /api/rss-feeds/scrape
GET  /api/tickers/search?q={query}
GET  /api/tickers/{symbol}/details
GET  /api/tickers/{symbol}/intraday?range=1d
GET  /api/tickers/bars?symbol=&range=
GET  /api/channels
GET  /api/videos/latest?channel_id={channel_id}
GET  /api/videos
POST /api/videos/analyze
GET  /api/stream/market
```

Service-specific endpoints:

```text
Market Data Service:
GET  /health
POST /api/questrade/oauth/save
GET  /api/tickers/search?q={query}
GET  /api/tickers/{symbol}/details
GET  /api/tickers/{symbol}/intraday?interval=1min
GET  /api/tickers/{symbol}/ratios

News Feed Service:
GET  /health
GET  /api/health
POST /api/scrape
GET  /api/channels/search?q={query}
GET  /api/channels
POST /api/channels
GET  /api/videos/latest?channel_id={channel_id}
POST /api/videos/analyze

Analyst Service:
GET  /health
POST /api/v1/quant/factor-scorecard
POST /api/v1/quant/portfolio
POST /api/v1/quant/statarb
POST /api/v1/quant/pit/filter
POST /api/v1/quant/stress
POST /api/v1/quant/execution
POST /api/v1/quant/microstructure
```

## Project Structure

```
portfolio-sim/
|-- backend/              Main Go API and database migrations
|-- market-data-service/  Go market data service
|-- news-feed-service/    Go news and YouTube service
|-- analyst-service/      Python FastAPI quant service
|-- frontend/             Next.js web app
|-- shared/               Shared Go packages
|-- docker-compose.yml    Local service orchestration
`-- README.md             This file
```

## Notes

- Database migrations live in `backend/migrations` and are mounted into the PostgreSQL container's init directory. They run automatically when the `postgres_data` volume is first created.
- API keys can be supplied through environment variables, but the application expects validated provider credentials to be stored in the database for normal operation.
- The frontend uses `pnpm`; `frontend/package.json` enforces it with a `preinstall` script.