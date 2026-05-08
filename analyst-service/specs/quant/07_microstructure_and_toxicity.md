# Microstructure And Toxicity

## Goal

Measure whether market conditions are safe enough to trade or whether flow has become toxic.

## Core Ideas

- tick rule:
  - infer buy or sell pressure from price changes
- volume buckets:
  - analyze flow in equal-volume slices instead of equal time slices
- VPIN:
  - `VPIN = imbalance_volume / bucket_volume`, often smoothed with rolling average

## Responsibilities

- sign trades using tape direction
- aggregate buy and sell volume by bucket
- compute bucket imbalance and rolling VPIN
- emit simple toxicity state:
  - `normal`
  - `elevated`
  - `toxic`

## Current Code

- `tick_rule_trade_signs(...)`
- `compute_vpin(...)`
- `latest_toxicity_signal(...)`

Code path: `src/quant_platform/analysis/microstructure.py`

## Why It Matters

- prevents buying into one-sided sell pressure
- improves execution timing
- serves as feature for future RL execution layer

## Future Upgrades

- Lee-Ready or quote-based trade signing
- order-book imbalance
- toxicity-aware execution throttles
- FPGA-friendly signal serialization
