from typing import Literal

import numpy as np
import pandas as pd


def tick_rule_trade_signs(price_changes: pd.Series, volume: pd.Series) -> pd.Series:
    signs = np.where(price_changes > 0, 1, np.where(price_changes < 0, -1, 0))
    return pd.Series(signs, index=price_changes.index)


def compute_vpin(
    price_changes: pd.Series,
    volume: pd.Series,
    bucket_size: int = 50,
) -> pd.Series:
    signs = tick_rule_trade_signs(price_changes, volume)
    buys = np.where(signs == 1, volume, 0)
    sells = np.where(signs == -1, volume, 0)

    n = len(volume)
    n_buckets = max(1, n // bucket_size)

    imbalances = []
    for i in range(n_buckets):
        start = i * bucket_size
        end = min((i + 1) * bucket_size, n)
        vbuy = buys[start:end].sum()
        vsell = sells[start:end].sum()
        bucket_vol = volume[start:end].sum()
        if bucket_vol > 0:
            imbalances.append((vbuy - vsell) / bucket_vol)
        else:
            imbalances.append(0)

    return pd.Series(imbalances, index=range(len(imbalances)))


def latest_toxicity_signal(
    price_changes: pd.Series,
    volume: pd.Series,
    vpin_window: int = 50,
    threshold_elevated: float = 0.6,
    threshold_toxic: float = 0.8,
) -> dict:
    vpin_series = compute_vpin(price_changes, volume, vpin_window)

    if len(vpin_series) == 0:
        return {"state": "normal", "vpin": 0.0, "vpin_smooth": 0.0}

    vpin_smooth = vpin_series.rolling(min(vpin_window, len(vpin_series))).mean().iloc[-1]
    vpin_current = vpin_series.iloc[-1]

    if vpin_current >= threshold_toxic:
        state: Literal["normal", "elevated", "toxic"] = "toxic"
    elif vpin_current >= threshold_elevated:
        state = "elevated"
    else:
        state = "normal"

    return {
        "state": state,
        "vpin": round(float(vpin_current), 4),
        "vpin_smooth": round(float(vpin_smooth), 4) if not np.isnan(vpin_smooth) else 0.0,
    }