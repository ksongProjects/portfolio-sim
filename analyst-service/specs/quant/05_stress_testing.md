# Stress Testing

## Goal

Probe fragility before live trading by simulating path risk, correlation spikes, and fat tails.

## Modules Inside This Area

- bootstrap resampling
- geometric Brownian motion paths
- Cholesky-based correlated shocks
- Student-t tail simulations

## Core Math

- bootstrap:
  - sample historical returns with replacement
- GBM:
  - `S_{t+1} = S_t * exp((mu - 0.5*sigma^2)dt + sigma*sqrt(dt)*Z)`
- Cholesky:
  - if `C = L L^T`, then `epsilon_corr = epsilon_independent * L^T`
- Student-t:
  - replace Gaussian shocks with heavier-tailed draws

## Responsibilities

- generate path ensembles
- simulate joint moves across assets
- build heavier-tail shock scenarios
- summarize VaR / CVaR style tail outcomes

## Current Code

- `bootstrap_return_paths(...)`
- `geometric_brownian_motion_paths(...)`
- `cholesky_correlated_shocks(...)`
- `correlated_return_paths(...)`
- `student_t_shocks(...)`
- `student_t_tail_report(...)`

Code path: `src/quant_platform/analysis/stress.py`

## Why Cholesky Is Its Own Component

Cholesky takes correlation assumptions and converts them into coherent joint shocks. It is reusable across:

- stress testing
- portfolio scenario analysis
- copula and tail-dependence extensions
- multi-asset execution hedging simulations

## Future Upgrades

- copula layer for asymmetric dependence
- regime-specific correlation matrices
- pathwise drawdown and ruin probability reports
- historical scenario replay
