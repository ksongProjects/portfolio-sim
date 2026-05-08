# Research Pipeline

## Goal

Connect every quant component into repeatable research workflow instead of isolated notebooks.

## Current System Map

- PIT data and feature materialization:
  - `src/quant_platform/pipeline/data.py`
  - `src/quant_platform/pipeline/features.py`
- testing and risk flows:
  - `src/quant_platform/pipeline/testing.py`
  - `src/quant_platform/pipeline/risk.py`
- reusable analysis modules:
  - `src/quant_platform/analysis/`

## Recommended Dependency Order

1. PIT snapshot
2. factor scorecard
3. selected-stock summary / market snapshot
4. portfolio construction
5. stress and tail-risk overlays
6. execution cost model
7. microstructure filter
8. model training / backtesting / monitoring

## Near-Term Integration Targets

- expose `analysis` modules through API endpoints
- add UI cards for market snapshot, selected-stock scorecard, stat-arb pair view
- log module outputs as testing artifacts
- bind execution and VPIN metrics into monitoring screens

## Longer-Term Targets

- optimizer-backed portfolio construction
- copula scenarios
- RL execution policy layer
- multi-asset hedging layer
- research experiment registry per module family
