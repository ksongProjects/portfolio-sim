# Portfolio Simulation Backend

## Architecture

Microservices backend with Go services and Python analyst service.

### Services

| Service | Language | Port | Purpose |
|---------|----------|------|---------|
| main-api | Go | 8080 | REST API, SSE streaming |
| market-data-service | Go | 8081 | Price feeds, normalization |
| news-feed-service | Go | 8082 | RSS scraping, video transcription |
| logging-service | Go | 50051 | gRPC log ingestion |
| analyst-service | Python | 8083 | Backtesting, Greeks computation |

### Data Flow

```
Market Data Providers (Questrade, Polygon, FMP)
           |
           v
market-data-service (normalizes & stores)
           |
           +--> Redis pub/sub --> main-api --> SSE --> frontend
           |
           v
     PostgreSQL (normalized_prices, intraday_bars)

News Sources (RSS, YouTube)
           |
           v
news-feed-service (scrapes, transcribes, summarizes)
           |
           +--> Redis pub/sub --> main-api --> SSE --> frontend
           v
     PostgreSQL (news_articles, news_videos)

Jobs Queue (Redis)
queue:backfill       --> market-data-service
queue:scrape-news    --> news-feed-service
queue:transcribe     --> news-feed-service
queue:compute-greeks --> analyst-service
queue:backtest       --> analyst-service
```

### Database

PostgreSQL 16 with monthly partitions:
- `raw_price_ticks`, `normalized_prices`, `intraday_bars`
- `option_chains`, `option_greeks`
- `news_articles`, `news_videos`
- `logs` (partitioned by month)

### Redis

- Pub/Sub channels: `market:ticks:{ticker}`, `market:chains:{ticker}`, `news:articles`, `news:videos`
- Job queues: `queue:{job_type}`

### Auth

Clerk OAuth for frontend authentication. JWT token passed in Authorization header.

## Development

```bash
# Start all services
docker-compose up -d

# Run migrations
docker-compose exec postgres psql -U postgres -d portfolio_sim -f /docker-entrypoint-initdb.d/001_initial_schema.sql

# Run a single service
go run ./services/main-api/main.go
```

## API Endpoints

### Portfolios
- `GET /api/portfolios` - List user portfolios
- `POST /api/portfolios` - Create portfolio
- `GET /api/portfolios/:id` - Get portfolio with positions
- `PUT /api/portfolios/:id` - Update portfolio
- `DELETE /api/portfolios/:id` - Delete portfolio

### Watchlists
- `GET /api/watchlists` - List user watchlists
- `POST /api/watchlists` - Create watchlist
- `GET /api/watchlists/:id` - Get watchlist with tickers
- `PUT /api/watchlists/:id` - Update watchlist
- `DELETE /api/watchlists/:id` - Delete watchlist
- `POST /api/watchlists/:id/tickers` - Add ticker
- `DELETE /api/watchlists/:id/tickers/:ticker_id` - Remove ticker

### Tickers
- `GET /api/tickers` - List tickers
- `GET /api/tickers/:id` - Get ticker with latest price

### Jobs
- `GET /api/jobs` - List jobs
- `POST /api/jobs` - Create job
- `GET /api/jobs/:id` - Get job status/result

### SSE Stream
- `GET /stream?channels=market:ticks:AAPL,news:articles` - Real-time data stream
