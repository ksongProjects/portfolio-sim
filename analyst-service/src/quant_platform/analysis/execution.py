from dataclasses import dataclass
from typing import Literal

import numpy as np
import pandas as pd


@dataclass
class ExecutionFill:
    bucket: int
    price: float
    size: float
    fill_cost: float


def twap_schedule(order_size: float, n_buckets: int = 20) -> list[ExecutionFill]:
    size_per_bucket = order_size / n_buckets
    return [
        ExecutionFill(bucket=i, price=0.0, size=size_per_bucket, fill_cost=0.0)
        for i in range(n_buckets)
    ]


def vwap_schedule(order_size: float, volume_profile: np.ndarray) -> list[ExecutionFill]:
    weights = volume_profile / volume_profile.sum()
    sizes = order_size * weights
    return [
        ExecutionFill(bucket=i, price=0.0, size=size, fill_cost=0.0)
        for i, size in enumerate(sizes)
    ]


def square_root_market_impact(
    order_size: float,
    adv: float,
    sigma: float,
    k: float = 1.0,
) -> float:
    return k * sigma * np.sqrt(order_size / adv)


def implementation_shortfall(
    decision_price: float,
    fills: list[ExecutionFill],
    side: Literal["buy", "sell"] = "buy",
) -> dict:
    if not fills:
        return {
            "avg_fill": decision_price,
            "shortfall_bps": 0.0,
            "total_cost": 0.0,
        }

    avg_fill = sum(f.price * f.size for f in fills) / sum(f.size for f in fills)

    if side == "buy":
        shortfall = avg_fill - decision_price
    else:
        shortfall = decision_price - avg_fill

    shortfall_bps = (shortfall / decision_price) * 10000

    return {
        "avg_fill": round(float(avg_fill), 4),
        "shortfall_bps": round(float(shortfall_bps), 2),
        "total_cost": round(float(shortfall * sum(f.size for f in fills)), 4),
    }