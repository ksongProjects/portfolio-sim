# Quant Module Map

`specs/quant.md` now acts as module index instead of raw idea dump.

## Objective

Turn point-in-time market data into:

1. ranked stock ideas
2. selected-stock scorecards
3. sector-aware long/short books
4. stress and tail-risk reports
5. execution and microstructure diagnostics

## Core Workflow

1. `Point-In-Time Data` filters market and fundamentals by what was known on trade date.
2. `Factor Engine` standardizes value, quality, momentum, sentiment, and alt-data inputs.
3. `Portfolio Construction` converts scores into neutralized long/short weights.
4. `Statistical Arbitrage` builds spread and mean-reversion signals for pair trades.
5. `Stress Testing` simulates joint shocks, tail events, and scenario losses.
6. `Execution` estimates schedules, impact, and implementation shortfall.
7. `Microstructure` measures order-flow toxicity and tradeable conditions.
8. `Research Pipeline` ties data, features, training, testing, and monitoring together.

## Module Inventory

| Module | Job | Main formulas / ideas | Code |
| :--- | :--- | :--- | :--- |
| PIT Data | As-of filtering, revision control, look-ahead prevention | `known_at <= as_of`, latest visible revision | `src/quant_platform/analysis/pit.py` |
| Factor Engine | Cross-sectional stock scoring | Z-score, winsorization, sector-neutral factors, composite score | `src/quant_platform/analysis/factors.py` |
| Portfolio Construction | Turn scores into books | sector de-meaning, long/short ranking, gross exposure split | `src/quant_platform/analysis/portfolio.py` |
| Statistical Arbitrage | Relative-value pair signals | hedge ratio `beta`, spread, rolling spread Z-score, half-life | `src/quant_platform/analysis/statarb.py` |
| Stress Testing | Tail and correlation scenarios | bootstrap, GBM, Cholesky, Student-t | `src/quant_platform/analysis/stress.py` |
| Execution | Trade schedule and cost model | TWAP, VWAP, square-root impact, implementation shortfall | `src/quant_platform/analysis/execution.py` |
| Microstructure | Toxic flow detection | tick rule, volume buckets, VPIN | `src/quant_platform/analysis/microstructure.py` |

## Detailed Specs

- [Factor Engine](quant/01_factor_engine.md)
- [Portfolio Construction](quant/02_portfolio_construction.md)
- [Statistical Arbitrage](quant/03_statistical_arbitrage.md)
- [Point-In-Time Data](quant/04_point_in_time_data.md)
- [Stress Testing](quant/05_stress_testing.md)
- [Execution And Costs](quant/06_execution_and_costs.md)
- [Microstructure And Toxicity](quant/07_microstructure_and_toxicity.md)
- [Research Pipeline](quant/08_research_pipeline.md)

## Practical Build Order

1. Start with PIT snapshot + factor scorecard.
2. Add selected-stock summaries and market snapshot views.
3. Add portfolio neutralization and stress testing.
4. Add execution costs before trusting backtests.
5. Add microstructure filters before intraday or high-turnover deployment.
6. Add advanced dependence models such as copulas after core workflow is stable.

## Advanced Topics Kept As Future Modules

- Copula calibration for asymmetric tail dependence
- Optimizer-backed portfolio construction with explicit constraints
- RL-driven execution policies
- FPGA / hardware execution
- Multi-asset option hedging during execution
