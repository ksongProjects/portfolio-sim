# Analyst Service

Quantitative analysis and algorithmic trading research platform.

## Overview

Written in Python using FastAPI. Provides quantitative finance functionality including factor analysis, portfolio construction, statistical arbitrage, stress testing, and execution optimization.

## Configuration

Environment variables:

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | PostgreSQL connection string |
| `REDIS_ADDR` | Redis address (default: redis://redis:6379) |

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Analyst Service (FastAPI)                  │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │                    API Routes                         │   │
│  │   /api/v1/quant/{factor,portfolio,statarb,pit,      │   │
│  │                    stress,execution,microstructure}   │   │
│  └──────────────────────────────────────────────────────┘   │
│                              │                               │
│                              ▼                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │               Analysis Modules                       │   │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────────┐ │   │
│  │  │ Factors │ │Portfolio│ │StatArb  │ │ Stress Test │ │   │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────────┘ │   │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────────────────────┐ │   │
│  │  │   PIT   │ │ Execute │ │   Microstructure        │ │   │
│  │  └─────────┘ └─────────┘ └─────────────────────────┘ │   │
│  └──────────────────────────────────────────────────────┘   │
│                              │                               │
│                              ▼                               │
│                      ┌─────────────┐                         │
│                      │  Redis/DB   │                         │
│                      └─────────────┘                         │
└─────────────────────────────────────────────────────────────┘
```

## API Endpoints

### Health

#### `GET /health`
```json
{"status": "ok"}
```

---

### Factor Analysis

#### `POST /api/v1/quant/factor-scorecard`

Build a multi-factor scorecard for stocks.

**Request:**
```json
{
  "data": [
    {"symbol": "AAPL", "pe_ratio": 25.5, "earnings_growth": 0.15, "momentum_1m": 0.05},
    {"symbol": "MSFT", "pe_ratio": 30.2, "earnings_growth": 0.12, "momentum_1m": 0.03}
  ],
  "factors": [
    {"name": "value", "source_col": "pe_ratio", "weight": 0.3, "direction": "invert"},
    {"name": "growth", "source_col": "earnings_growth", "weight": 0.4, "direction": "positive"},
    {"name": "momentum", "source_col": "momentum_1m", "weight": 0.3, "direction": "positive"}
  ]
}
```

**Response:**
```json
{
  "scorecard": [
    {"symbol": "AAPL", "composite_score": 0.65, "value": 0.3, "growth": 0.4, "momentum": 0.3},
    {"symbol": "MSFT", "composite_score": 0.55, "value": 0.2, "growth": 0.35, "momentum": 0.2}
  ],
  "market_snapshot": {
    "avg_score": 0.45,
    "high_score": 0.75,
    "low_score": 0.25
  }
}
```

---

### Portfolio Construction

#### `POST /api/v1/quant/portfolio`

Build a long/short portfolio from scored securities.

**Request:**
```json
{
  "data": [
    {"symbol": "AAPL", "composite_score": 0.8, "sector": "tech"},
    {"symbol": "MSFT", "composite_score": 0.75, "sector": "tech"},
    {"symbol": "XOM", "composite_score": 0.2, "sector": "energy"}
  ],
  "score_col": "composite_score",
  "n_long": 5,
  "n_short": 5,
  "neutralize": true
}
```

**Response:**
```json
{
  "book": [
    {"symbol": "AAPL", "side": "long", "score": 0.8, "weight": 0.1},
    {"symbol": "MSFT", "side": "long", "score": 0.75, "weight": 0.08},
    {"symbol": "XOM", "side": "short", "score": 0.2, "weight": -0.05}
  ]
}
```

---

### Statistical Arbitrage

#### `POST /api/v1/quant/statarb`

Compute pairs trading signals.

**Request:**
```json
{
  "price_a": [100, 101, 102, 101.5, 103],
  "price_b": [200, 201, 202, 201.5, 203],
  "z_threshold": 2.0,
  "window": 60
}
```

**Response:**
```json
{
  "signal": "long_a_short_b",
  "spread_zscore": 2.3,
  "hedge_ratio": 0.502,
  "half_life": 15.5
}
```

---

### Point-in-Time Data

#### `POST /api/v1/quant/pit/filter`

Filter data to simulate historical knowledge.

**Request:**
```json
{
  "data": [
    {"symbol": "AAPL", "earnings": 5.0, "known_at": "2024-01-15"},
    {"symbol": "AAPL", "earnings": 6.0, "known_at": "2024-04-15"}
  ],
  "as_of": "2024-02-01",
  "known_at_col": "known_at"
}
```

**Response:**
```json
{
  "filtered": [
    {"symbol": "AAPL", "earnings": 5.0, "known_at": "2024-01-15"}
  ],
  "audit_leaks": [
    {"row": 1, "leaked_at": "2024-04-15", "value": 6.0}
  ]
}
```

---

### Stress Testing

#### `POST /api/v1/quant/stress`

Monte Carlo VaR/CVaR calculation using Student's t-distribution.

**Request:**
```json
{
  "returns": [0.01, -0.02, 0.015, -0.005, 0.02],
  "confidence": 0.95,
  "n_paths": 5000
}
```

**Response:**
```json
{
  "VaR": -0.035,
  "CVaR": -0.048,
  "worst_case": -0.065,
  "confidence": 0.95,
  "n_paths": 5000
}
```

---

### Execution Optimization

#### `POST /api/v1/quant/execution`

Calculate execution schedule and expected market impact.

**Request:**
```json
{
  "decision_price": 175.50,
  "order_size": 10000,
  "side": "buy",
  "schedule_type": "twap",
  "volume_profile": [0.1, 0.15, 0.2, 0.25, 0.2, 0.1]
}
```

**Response:**
```json
{
  "avg_fill": 176.10,
  "shortfall_bps": 3.42,
  "total_cost": 3420.00,
  "expected_impact": 0.0042
}
```

**Market Impact Model:**
Uses square root market impact: `impact = sigma * sqrt(order_size / adv)`

---

### Market Microstructure

#### `POST /api/v1/quant/microstructure`

Calculate VPIN (Volume-synchronized Probability of Informed Trading) for toxicity detection.

**Request:**
```json
{
  "price_changes": [0.01, -0.02, 0.015, -0.005, 0.02],
  "volume": [1000000, 1200000, 800000, 1500000, 900000],
  "vpin_window": 50
}
```

**Response:**
```json
{
  "state": "normal",
  "vpin": 0.35,
  "vpin_smooth": 0.32
}
```

**States:**
- `normal`: VPIN < 0.4
- `elevated`: 0.4 <= VPIN < 0.7
- `toxic`: VPIN >= 0.7

---

## Background Workers

### Backtest Queue Processor
Listens on `queue:backtest` for backtest jobs.

**Job format:**
```json
{
  "portfolio_config": {"id": "strategy-1", "positions": [...]},
  "start_date": "2024-01-01",
  "end_date": "2024-12-31",
  "initial_cash": 100000
}
```

**Result stored at:** `backtest:{portfolio_id}:result`

### Greeks Calculator Queue
Listens on `queue:compute-greeks` for options Greek calculations.

## Dependencies

- **fastapi**: Web framework
- **pydantic**: Data validation
- **scipy**: Statistical functions (norm, t-distribution)
- **pandas**: Data manipulation
- **numpy**: Numerical operations
- **asyncpg**: PostgreSQL async driver
- **redis**: Redis client

## Running

```bash
pip install -r requirements.txt
python main.py
```

The service will start on port 8080 (internal) with health endpoint and all quant routes.