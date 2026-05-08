# Point-In-Time Data

## Goal

Prevent look-ahead bias by enforcing what market actually knew at each historical timestamp.

## Key Timestamps

- `effective_at`: period data refers to
- `known_at`: moment data became visible to market
- optional `ingested_at`: when platform received and stored data

## Core Rule

- only use rows where `known_at <= as_of`

## Responsibilities

- filter visible rows at chosen backtest date
- keep latest visible revision for each `(entity, effective_at)` pair
- build latest as-of snapshot for each ticker
- audit rows that would leak future information

## Current Code

- `filter_as_of(...)`
- `latest_known_observations(...)`
- `audit_lookahead_rows(...)`

Code path: `src/quant_platform/analysis/pit.py`

## Why It Matters

- earnings revisions can arrive days or weeks later
- restatements must not overwrite history in backtest
- selected-stock analysis must reflect what trader could have seen at that date

## Future Upgrades

- corporate action adjustment layer
- permanent ID mapping across ticker changes
- explicit restatement lineage
- PIT joins across fundamentals, prices, and alt-data feeds
