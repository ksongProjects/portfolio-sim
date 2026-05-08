from dataclasses import dataclass
from typing import Literal

import numpy as np
import pandas as pd


@dataclass
class FactorDefinition:
    name: str
    source_col: str
    weight: float
    direction: Literal["positive", "invert"]
    neutralization_group: str | None = None


def z_score(series: pd.Series) -> pd.Series:
    mask = series.notna() & (series.std() != 0)
    result = pd.Series(np.nan, index=series.index)
    result.loc[mask] = (series.loc[mask] - series.loc[mask].mean()) / series.loc[mask].std()
    return result


def winsorize(series: pd.Series, limit: float = 3.0) -> pd.Series:
    return series.clip(-limit, limit)


def build_factor_scorecard(
    df: pd.DataFrame,
    factors: list[FactorDefinition],
) -> pd.DataFrame:
    result = df[["ticker", "sector"]].copy()

    for f in factors:
        raw = df[f.source_col]
        z = z_score(raw)
        z_win = winsorize(z)
        if f.direction == "invert":
            z_win = -z_win
        result[f"{f.name}_z"] = z_win

    weight_sum = sum(abs(f.weight) for f in factors)
    z_cols = [f"{f.name}_z" for f in factors]

    result["composite_score"] = sum(
        result[col] * f.weight for col, f in zip(z_cols, factors)
    ) / weight_sum

    result["score_rank"] = result["composite_score"].rank(ascending=False, method="min")
    result["score_percentile"] = result["composite_score"].rank(pct=True) * 100

    return result


def summarize_selected_stocks(df: pd.DataFrame, tickers: list[str]) -> pd.DataFrame:
    latest = df.sort_values("effective_at").groupby("ticker").last().reset_index()
    selected = latest[latest["ticker"].isin(tickers)]
    return selected


def build_market_snapshot(df: pd.DataFrame, n: int = 10) -> dict:
    latest = df.sort_values("effective_at").groupby("ticker").last().reset_index()

    snapshot = build_factor_scorecard(latest, [])
    snapshot = snapshot.sort_values("composite_score", ascending=False)

    return {
        "top_stocks": snapshot.head(n)["ticker"].tolist(),
        "bottom_stocks": snapshot.tail(n)["ticker"].tolist(),
        "top_sectors": (
            snapshot.groupby("sector")["composite_score"]
            .mean()
            .sort_values(ascending=False)
            .head(5)
            .index.tolist()
        ),
        "as_of": latest["effective_at"].max(),
    }