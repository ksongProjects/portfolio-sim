import numpy as np
import pandas as pd


def neutralize_by_group(df: pd.DataFrame, score_col: str, group_col: str) -> pd.Series:
    group_means = df.groupby(group_col)[score_col].transform("mean")
    return df[score_col] - group_means


def build_long_short_book(
    df: pd.DataFrame,
    score_col: str = "composite_score",
    group_col: str = "sector",
    n_long: int = 10,
    n_short: int = 10,
    neutralize: bool = True,
) -> pd.DataFrame:
    data = df.copy()

    if neutralize and group_col in data.columns:
        data["portfolio_score"] = neutralize_by_group(data, score_col, group_col)
    else:
        data["portfolio_score"] = data[score_col]

    ranked = data.sort_values("portfolio_score", ascending=False)

    longs = ranked.head(n_long).copy()
    longs["side"] = "long"

    shorts = ranked.tail(n_short).copy()
    shorts["side"] = "short"

    book = pd.concat([longs, shorts]).reset_index(drop=True)

    total_abs = book["portfolio_score"].abs().sum()
    book["weight"] = book["portfolio_score"] / total_abs if total_abs != 0 else 0.0

    return book[["ticker", "sector", "portfolio_score", "side", "weight"]]