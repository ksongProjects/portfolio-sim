# Market Data Service

Real-time market data aggregation service that fetches prices from multiple providers, normalizes data, and streams updates via SSE.

## Overview

Written in Go. Fetches prices from Polygon, Questrade, and FMP providers, normalizes them, stores in PostgreSQL, and publishes real-time updates via Redis pub/sub.

## Configuration

Environment variables:

| Variable | Description |
|----------|-------------|
| `DATABASE_HOST` | PostgreSQL host |
| `DATABASE_PORT` | PostgreSQL port (default: 5432) |
| `DATABASE_USER` | PostgreSQL user |
| `DATABASE_PASSWORD` | PostgreSQL password |
| `DATABASE_NAME` | Database name |
| `DATABASE_MAX_CONNS` | Max connections (default: 20) |
| `REDIS_HOST` | Redis host |
| `REDIS_PORT` | Redis port (default: 6379) |
| `REDIS_PASSWORD` | Redis password |
| `POLYGON_API_KEY` | Polygon.io API key |
| `QUESTRADE_API_KEY` | Questrade API key |
| `FMP_API_KEY` | Financial Modeling Prep API key |

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Market Data Service                       │
├─────────────────────────────────────────────────────────────┤
│  Providers          Normalizer          Storage             │
│  ┌─────────┐       ┌──────────┐       ┌────────┐           │
│  │ Polygon │       │          │       │        │           │
│  ├─────────┤  ───► │ Normalizer│ ───► │ Postgres│           │
│  │ Questrade│       │          │       │        │           │
│  ├─────────┤       └──────────┘       └────────┘           │
│  │ FMP     │                           │                    │
│  └─────────┘                           ▼                    │
│                                   ┌────────┐               │
│                                   │ Redis  │               │
│                                   │ PubSub │               │
│                                   └────────┘               │
├─────────────────────────────────────────────────────────────┤
│  SSE Manager ───► Real-time WebSocket Streaming            │
│  Backfill Queue ───► Redis BRPOP                            │
└─────────────────────────────────────────────────────────────┘
```

## Providers

### Polygon.io
- Real-time and historical market data
- Rate limit: 60 req/min (free tier)

### Questrade
- Canadian market data
- Rate limit: 100 req/min

### Financial Modeling Prep
- Financial statements and fundamental data
- Rate limit: 250 req/min

## Data Flow

1. **Price Fetching**: Each provider runs a continuous loop fetching prices every second
2. **Normalization**: Prices are converted to a common format with normalized tickers
3. **Storage**: Normalized prices stored in `normalized_prices` table
4. **Streaming**: Updates published to Redis channel for SSE delivery

## Backfill System

The service processes backfill requests from Redis queue `queue:backfill`:

```json
{
  "ticker": "AAPL",
  "data_type": "intraday_bars",
  "interval": "1m",
  "source": "polygon"
}
```

Supported data types:
- `intraday_bars`: OHLCV bars at specified interval
- `option_chain`: Full option chain with Greeks

## Storage Schema

### normalized_prices
```sql
CREATE TABLE normalized_prices (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  ticker_id UUID REFERENCES tickers(id),
  price DECIMAL(18,8),
  bid DECIMAL(18,8),
  ask DECIMAL(18,8),
  volume BIGINT,
  source_id VARCHAR(50),
  timestamp TIMESTAMPTZ
);
```

### intraday_bars
```sql
CREATE TABLE intraday_bars (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  ticker_id UUID REFERENCES tickers(id),
  interval VARCHAR(10),
  open DECIMAL(18,8),
  high DECIMAL(18,8),
  low DECIMAL(18,8),
  close DECIMAL(18,8),
  volume BIGINT,
  timestamp TIMESTAMPTZ
);
```

### option_chains
```sql
CREATE TABLE option_chains (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  ticker_id UUID REFERENCES tickers(id),
  source VARCHAR(50),
  expiration DATE,
  strike DECIMAL(18,8),
  option_type VARCHAR(10),
  bid DECIMAL(18,8),
  ask DECIMAL(18,8),
  delta DECIMAL(18,8),
  gamma DECIMAL(18,8),
  theta DECIMAL(18,8),
  vega DECIMAL(18,8),
  implied_vol DECIMAL(18,8),
  volume BIGINT,
  open_interest BIGINT,
  timestamp TIMESTAMPTZ
);
```

## SSE Streaming

Publishes to Redis channels:
- `tick:{provider}:{symbol}` - Real-time price updates

Example message:
```json
{
  "ticker_id": "uuid",
  "price": 175.50,
  "bid": 175.48,
  "ask": 175.52,
  "volume": 1000000,
  "source_id": "polygon",
  "timestamp": "2026-05-09T10:30:00Z"
}
```