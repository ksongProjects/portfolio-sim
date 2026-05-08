# Factor Engine

## Goal

Convert raw market and company fields into comparable stock scores for market-wide ranking and selected-stock review.

## Inputs

- `effective_at`
- `ticker`
- `sector`
- valuation fields such as `ev_ebitda`
- quality fields such as `roic`
- momentum fields such as `12m-1m return`
- optional sentiment, macro, and alt-data features

## Core Math

- Z-score standardization:
  - `z = (x - mean) / std`
- Winsorization:
  - clamp extreme Z-scores to a fixed limit such as `[-3, 3]`
- Directionality:
  - cheap valuation metrics invert sign so lower raw values become higher scores
- Composite score:
  - `score = sum(weight_i * factor_z_i) / sum(abs(weight_i))`

## Responsibilities

- normalize unlike units into one scale
- neutralize factors inside sector groups when needed
- rank full universe
- create selected-stock scorecards
- generate market snapshot of top and bottom stocks and sectors

## Current Code

- `FactorDefinition`: declare factor name, source column, weight, direction, neutralization group
- `build_factor_scorecard(...)`: build `*_z`, `composite_score`, rank, percentile
- `summarize_selected_stocks(...)`: latest row for watchlist names
- `build_market_snapshot(...)`: latest market-wide leaderboard

Code path: `src/quant_platform/analysis/factors.py`

## Outputs

- per-factor Z-scores
- `composite_score`
- `score_rank`
- `score_percentile`
- selected-stock breakdowns
- top/bottom market snapshot

## Next Extensions

- factor covariance diagnostics
- regime-specific weights
- analyst estimate revisions
- alternative data factor families
