# Logging Service

Centralized logging service with structured storage, querying, and real-time streaming.

## Overview

Written in Go. Provides centralized log aggregation with a PostgreSQL backend and Redis pub/sub for real-time log streaming.

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
| `LOGGING_HTTP_PORT` | HTTP listen port (default: 9090) |

## Endpoints

### Health

#### `GET /health`

```json
{"status": "ok"}
```

---

### Emit Log

#### `POST /api/logs`

Submit a log entry to be stored.

**Request:**
```json
{
  "level": "INFO",
  "service": "market-data-service",
  "component": "priceFetcher",
  "message": "Price updated for AAPL",
  "metadata": {
    "ticker": "AAPL",
    "price": 175.50
  },
  "trace_id": "abc123",
  "span_id": "def456"
}
```

**Response:**
```json
{"id": "generated-uuid"}
```

**Valid levels:** `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`

---

### Query Logs

#### `GET /api/logs`

Query stored log entries.

**Query Parameters:**
- `limit` (optional): Max results (default: 50, max: 500)
- `level` (optional): Filter by log level (e.g., `ERROR`)
- `service` (optional): Filter by service name

**Response:**
```json
[
  {
    "id": "uuid",
    "timestamp": "2026-05-09T10:30:00Z",
    "level": "INFO",
    "service": "market-data-service",
    "component": "priceFetcher",
    "message": "Price updated for AAPL",
    "metadata": {"ticker": "AAPL", "price": 175.50},
    "trace_id": "abc123",
    "span_id": "def456"
  }
]
```

---

## Redis Pub/Sub

When a log is emitted, it's published to a Redis channel named `logs:{service}`.

For example, a log from `market-data-service` is published to `logs:market-data-service`.

**Message format:**
```json
{
  "id": "uuid",
  "timestamp": "2026-05-09T10:30:00Z",
  "level": "INFO",
  "service": "market-data-service",
  "component": "priceFetcher",
  "message": "Price updated for AAPL",
  "metadata": {},
  "trace_id": null,
  "span_id": null
}
```

## Database Schema

```sql
CREATE TABLE logs (
  id UUID PRIMARY KEY,
  timestamp TIMESTAMPTZ NOT NULL,
  level VARCHAR(10) NOT NULL,
  service VARCHAR(100) NOT NULL,
  component VARCHAR(100),
  message TEXT NOT NULL,
  metadata JSONB,
  trace_id VARCHAR(100),
  span_id VARCHAR(100)
);

CREATE INDEX idx_logs_timestamp ON logs(timestamp DESC);
CREATE INDEX idx_logs_service ON logs(service);
CREATE INDEX idx_logs_level ON logs(level);
```

## Log Entry Structure

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Unique log entry identifier |
| `timestamp` | TIMESTAMPTZ | When the log was created |
| `level` | VARCHAR | DEBUG, INFO, WARN, ERROR, FATAL |
| `service` | VARCHAR | Originating service name |
| `component` | VARCHAR | Specific component within service |
| `message` | TEXT | Log message content |
| `metadata` | JSONB | Additional structured data |
| `trace_id` | VARCHAR | Distributed trace ID |
| `span_id` | VARCHAR | Span ID for trace context |

## Usage Example

From any service, emit logs via HTTP:

```bash
curl -X POST http://logging-service:9090/api/logs \
  -H "Content-Type: application/json" \
  -d '{
    "level": "INFO",
    "service": "my-service",
    "component": "handler",
    "message": "Request processed",
    "metadata": {"method": "GET", "path": "/api/data", "status": 200}
  }'
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Logging Service                           │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│   POST /api/logs ──────────────────────────────────────┐   │
│                                                         │   │
│   ┌────────────┐    ┌────────────┐    ┌────────────┐  │   │
│   │  Validate  │───►│   Insert   │───►│    Redis    │  │   │
│   │  Request   │    │   Postgres │    │   Publish   │  │   │
│   └────────────┘    └────────────┘    └────────────┘  │   │
│                                                 │        │   │
│                                                 ▼        │   │
│                                           ┌────────────┐ │   │
│   GET /api/logs ─────────────────────►    │   Query    │ │   │
│                                           │   Postgres │ │   │
│                                           └────────────┘ │   │
│                                           ┌────────────┐ │   │
│                                           │   Redis    │ │   │
│                                           │   Subscribe│ │   │
│                                           └────────────┘ │   │
└─────────────────────────────────────────────────────────────┘
```

## Running

```bash
# With Docker
docker-compose up logging-service

# Standalone
go build -o logging-service && ./logging-service
```