from typing import Literal

import numpy as np
import pandas as pd


def estimate_hedge_ratio(price_a: pd.Series, price_b: pd.Series) -> float:
    cov = np.cov(price_a, price_b)[0, 1]
    var = np.var(price_b)
    return cov / var if var != 0 else 0.0


def compute_spread(price_a: pd.Series, price_b: pd.Series, beta: float) -> pd.Series:
    return price_a - beta * price_b


def rolling_spread_zscore(spread: pd.Series, window: int = 60) -> pd.Series:
    roll_mean = spread.rolling(window).mean()
    roll_std = spread.rolling(window).std()
    return (spread - roll_mean) / roll_std


def mean_reversion_half_life(spread: pd.Series) -> float | None:
    returns = spread.diff().dropna()
    if len(returns) < 2:
        return None
    autocorr = returns.autocorr()
    if autocorr <= 0 or autocorr >= 1:
        return None
    return -np.log(2) / np.log(autocorr)


def latest_pairs_signal(
    price_a: pd.Series,
    price_b: pd.Series,
    z_threshold: float = 2.0,
    window: int = 60,
) -> dict:
    beta = estimate_hedge_ratio(price_a, price_b)
    spread = compute_spread(price_a, price_b, beta)
    z = rolling_spread_zscore(spread, window)

    latest_z = z.iloc[-1] if not z.empty else 0.0

    half_life = mean_reversion_half_life(spread)

    if latest_z > z_threshold:
        signal = "long_a_short_b"
    elif latest_z < -z_threshold:
        signal = "short_a_long_b"
    elif abs(latest_z) < 0.5:
        signal = "close"
    else:
        signal = "hold"

    return {
        "signal": signal,
        "spread_zscore": round(float(latest_z), 4),
        "hedge_ratio": round(float(beta), 4),
        "half_life": round(float(half_life), 4) if half_life else None,
    }