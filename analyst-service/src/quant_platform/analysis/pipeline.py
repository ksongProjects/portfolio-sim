import pandas as pd

from .factors import build_factor_scorecard, build_market_snapshot, summarize_selected_stocks, FactorDefinition
from .portfolio import build_long_short_book, neutralize_by_group
from .statarb import latest_pairs_signal, estimate_hedge_ratio, compute_spread, rolling_spread_zscore
from .pit import filter_as_of, latest_known_observations, audit_lookahead_rows
from .stress import (
    bootstrap_return_paths,
    geometric_brownian_motion_paths,
    cholesky_correlated_shocks,
    student_t_shocks,
    student_t_tail_report,
)
from .execution import (
    twap_schedule,
    vwap_schedule,
    square_root_market_impact,
    implementation_shortfall,
    ExecutionFill,
)
from .microstructure import latest_toxicity_signal, compute_vpin, tick_rule_trade_signs


def run_research_pipeline(
    raw_data: pd.DataFrame,
    factors: list[FactorDefinition],
    as_of: str,
    n_long: int = 10,
    n_short: int = 10,
) -> dict:
    pit_data = filter_as_of(raw_data, as_of)
    pit_data = latest_known_observations(pit_data)

    scorecard = build_factor_scorecard(pit_data, factors)
    selected_summary = summarize_selected_stocks(scorecard, raw_data["ticker"].unique().tolist())
    market_snap = build_market_snapshot(scorecard)

    book = build_long_short_book(scorecard, n_long=n_long, n_short=n_short)

    return {
        "as_of": as_of,
        "scorecard": scorecard,
        "market_snapshot": market_snap,
        "portfolio_book": book,
    }