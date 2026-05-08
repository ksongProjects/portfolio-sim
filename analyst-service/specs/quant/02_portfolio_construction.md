# Portfolio Construction

## Goal

Turn raw alpha scores into tradeable long/short books without letting one sector or theme dominate portfolio by accident.

## Core Ideas

- sector de-meaning:
  - subtract sector average score before selecting names
- ranking:
  - highest adjusted scores become longs
  - lowest adjusted scores become shorts
- exposure split:
  - default gross exposure split `+50% / -50%`
- weighting:
  - score-weighted by absolute conviction

## Responsibilities

- pick latest cross-section
- optionally neutralize score inside sector
- select top `N` longs and bottom `N` shorts
- normalize weights so book is dollar balanced

## Current Code

- `neutralize_by_group(...)`
- `build_long_short_book(...)`

Code path: `src/quant_platform/analysis/portfolio.py`

## Outputs

- `portfolio_score`
- `side`
- `weight`
- latest trade date book

## Future Upgrades

- optimizer constraints for sector, beta, turnover
- benchmark-relative active risk limits
- transaction-cost-aware sizing
- multi-period rebalance logic
