# Execution And Costs

## Goal

Translate research alpha into realistic trade costs so backtests do not assume free fills.

## Core Modules

- TWAP schedule
- VWAP schedule
- square-root impact model
- implementation shortfall report

## Core Math

- TWAP:
  - split order evenly across time buckets
- VWAP:
  - allocate size proportional to expected volume profile
- square-root impact:
  - `impact ~= k * sigma * sqrt(order_size / ADV)`
- implementation shortfall:
  - buy side: `avg_fill - decision_price`
  - sell side: `decision_price - avg_fill`

## Current Code

- `twap_schedule(...)`
- `vwap_schedule(...)`
- `square_root_market_impact(...)`
- `implementation_shortfall(...)`
- `ExecutionFill`

Code path: `src/quant_platform/analysis/execution.py`

## Outputs

- schedule by bucket
- expected cost estimate
- realized shortfall in price and bps

## Future Upgrades

- order-book-aware slippage
- participation caps
- intraday alpha decay model
- multi-asset hedge scheduling
