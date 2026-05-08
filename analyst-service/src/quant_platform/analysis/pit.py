import pandas as pd


def filter_as_of(df: pd.DataFrame, as_of: str, known_at_col: str = "known_at") -> pd.DataFrame:
    return df[df[known_at_col] <= as_of]


def latest_known_observations(
    df: pd.DataFrame,
    entity_col: str = "ticker",
    effective_col: str = "effective_at",
    known_at_col: str = "known_at",
) -> pd.DataFrame:
    return (
        df.sort_values(known_at_col)
        .groupby([entity_col, effective_col], as_index=False)
        .last()
    )


def audit_lookahead_rows(
    df: pd.DataFrame,
    as_of: str,
    known_at_col: str = "known_at",
    effective_col: str = "effective_at",
) -> pd.DataFrame:
    return df[df[known_at_col] > as_of]