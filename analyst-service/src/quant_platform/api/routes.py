from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from typing import Literal

from ..analysis.factors import FactorDefinition, build_factor_scorecard, build_market_snapshot, summarize_selected_stocks
from ..analysis.portfolio import build_long_short_book, neutralize_by_group
from ..analysis.statarb import latest_pairs_signal, estimate_hedge_ratio, compute_spread
from ..analysis.pit import filter_as_of, latest_known_observations, audit_lookahead_rows
from ..analysis.stress import student_t_tail_report, bootstrap_return_paths, geometric_brownian_motion_paths
from ..analysis.execution import twap_schedule, vwap_schedule, square_root_market_impact, implementation_shortfall, ExecutionFill
from ..analysis.microstructure import latest_toxicity_signal, compute_vpin


router = APIRouter(prefix="/api/v1/quant", tags=["quant"])


class FactorInput(BaseModel):
    name: str
    source_col: str
    weight: float
    direction: Literal["positive", "invert"]
    neutralization_group: str | None = None


class FactorScorecardRequest(BaseModel):
    data: list[dict]
    factors: list[FactorInput]


class FactorScorecardResponse(BaseModel):
    scorecard: list[dict]
    market_snapshot: dict


@router.post("/factor-scorecard", response_model=FactorScorecardResponse)
def factor_scorecard(req: FactorScorecardRequest):
    import pandas as pd
    df = pd.DataFrame(req.data)
    factor_defs = [FactorDefinition(**f.model_dump()) for f in req.factors]
    scorecard = build_factor_scorecard(df, factor_defs)
    snapshot = build_market_snapshot(scorecard)
    return FactorScorecardResponse(
        scorecard=scorecard.to_dict(orient="records"),
        market_snapshot=snapshot,
    )


class PortfolioBookRequest(BaseModel):
    data: list[dict]
    score_col: str = "composite_score"
    n_long: int = 10
    n_short: int = 10
    neutralize: bool = True


class PortfolioBookResponse(BaseModel):
    book: list[dict]


@router.post("/portfolio", response_model=PortfolioBookResponse)
def portfolio(req: PortfolioBookRequest):
    import pandas as pd
    df = pd.DataFrame(req.data)
    book = build_long_short_book(df, req.score_col, n_long=req.n_long, n_short=req.n_short, neutralize=req.neutralize)
    return PortfolioBookResponse(book=book.to_dict(orient="records"))


class StatArbRequest(BaseModel):
    price_a: list[float]
    price_b: list[float]
    z_threshold: float = 2.0
    window: int = 60


class StatArbResponse(BaseModel):
    signal: str
    spread_zscore: float
    hedge_ratio: float
    half_life: float | None


@router.post("/statarb", response_model=StatArbResponse)
def statarb(req: StatArbRequest):
    import pandas as pd
    pa = pd.Series(req.price_a)
    pb = pd.Series(req.price_b)
    result = latest_pairs_signal(pa, pb, req.z_threshold, req.window)
    return StatArbResponse(**result)


class PITRequest(BaseModel):
    data: list[dict]
    as_of: str
    known_at_col: str = "known_at"


class PITResponse(BaseModel):
    filtered: list[dict]
    audit_leaks: list[dict]


@router.post("/pit/filter", response_model=PITResponse)
def pit_filter(req: PITRequest):
    import pandas as pd
    df = pd.DataFrame(req.data)
    filtered = filter_as_of(df, req.as_of, req.known_at_col)
    leaks = audit_lookahead_rows(df, req.as_of, req.known_at_col)
    return PITResponse(filtered=filtered.to_dict(orient="records"), audit_leaks=leaks.to_dict(orient="records"))


class StressRequest(BaseModel):
    returns: list[float]
    confidence: float = 0.95
    n_paths: int = 5000


class StressResponse(BaseModel):
    VaR: float
    CVaR: float
    worst_case: float
    confidence: float
    n_paths: int


@router.post("/stress", response_model=StressResponse)
def stress(req: StressRequest):
    import pandas as pd
    series = pd.Series(req.returns)
    result = student_t_tail_report(series, req.confidence, req.n_paths)
    return StressResponse(**result)


class ExecutionRequest(BaseModel):
    decision_price: float
    order_size: float
    side: Literal["buy", "sell"]
    schedule_type: Literal["twap", "vwap"] = "twap"
    volume_profile: list[float] | None = None


class ExecutionResponse(BaseModel):
    avg_fill: float
    shortfall_bps: float
    total_cost: float
    expected_impact: float


@router.post("/execution", response_model=ExecutionResponse)
def execution(req: ExecutionRequest):
    sigma = 0.02
    adv = req.order_size * 10
    impact = square_root_market_impact(req.order_size, adv, sigma)

    if req.schedule_type == "twap":
        fills = twap_schedule(req.order_size)
    else:
        if req.volume_profile:
            import numpy as np
            fills = vwap_schedule(req.order_size, np.array(req.volume_profile))
        else:
            fills = twap_schedule(req.order_size)

    for f in fills:
        f.price = req.decision_price * (1 + impact if req.side == "buy" else 1 - impact)

    result = implementation_shortfall(req.decision_price, fills, req.side)
    result["expected_impact"] = round(float(impact), 4)
    return ExecutionResponse(**result)


class MicrostructureRequest(BaseModel):
    price_changes: list[float]
    volume: list[float]
    vpin_window: int = 50


class MicrostructureResponse(BaseModel):
    state: Literal["normal", "elevated", "toxic"]
    vpin: float
    vpin_smooth: float


@router.post("/microstructure", response_model=MicrostructureResponse)
def microstructure(req: MicrostructureRequest):
    import pandas as pd
    pc = pd.Series(req.price_changes)
    vol = pd.Series(req.volume)
    result = latest_toxicity_signal(pc, vol, req.vpin_window)
    return MicrostructureResponse(**result)