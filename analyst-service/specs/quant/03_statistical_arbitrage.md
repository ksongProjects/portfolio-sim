# Statistical Arbitrage

## Goal

Model relative-value opportunities between related assets instead of ranking each stock alone.

## Core Math

- hedge ratio:
  - `beta = cov(A, B) / var(B)`
- spread:
  - `spread = price_A - beta * price_B`
- rolling spread Z-score:
  - compare current spread to rolling mean and rolling std
- mean-reversion half-life:
  - estimate how quickly spread decays back toward mean

## Responsibilities

- align pair histories
- estimate hedge ratio
- compute spread
- score spread extremeness with rolling Z-score
- emit latest trading state:
  - `long_a_short_b`
  - `short_a_long_b`
  - `close`
  - `hold`

## Current Code

- `estimate_hedge_ratio(...)`
- `compute_spread(...)`
- `rolling_spread_zscore(...)`
- `mean_reversion_half_life(...)`
- `latest_pairs_signal(...)`

Code path: `src/quant_platform/analysis/statarb.py`

## Limits

- current module uses lightweight regression math
- no full Engle-Granger or Johansen test yet
- use as signal engine, not final institutional cointegration validator

## Future Upgrades

- residual stationarity tests
- rolling beta stability checks
- spread volatility targeting
- basket and sector-pair stat-arb
