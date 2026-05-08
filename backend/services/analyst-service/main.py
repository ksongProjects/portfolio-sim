import json
import math
import uuid
from datetime import datetime
from typing import Any

import asyncpg
import redis.asyncio as redis
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from scipy.stats import norm

import config

app = FastAPI(title="Analyst Service")

redis_client: redis.Redis = None
pg_pool: asyncpg.Pool = None


class BacktestRequest(BaseModel):
    portfolio_config: dict
    start_date: str
    end_date: str
    initial_cash: float


class BacktestResponse(BaseModel):
    run_id: uuid.UUID
    total_return: float
    sharpe_ratio: float
    max_drawdown: float
    win_rate: float
    trades: list


class ComputeGreeksRequest(BaseModel):
    chain_id: uuid.UUID
    underlying_price: float
    risk_free_rate: float = 0.05
    options: list[dict]


class OptionGreeksResponse(BaseModel):
    chain_id: uuid.UUID
    options: list[dict]


async def process_greeks_queue():
    while True:
        try:
            job_data = await redis_client.blpop("queue:compute-greeks", timeout=5)
            if job_data:
                _, payload = job_data
                job = json.loads(payload)
                req = ComputeGreeksRequest(**job)
                await compute_greeks(req)
        except Exception:
            pass


@app.on_event("startup")
async def start_workers():
    import asyncio
    asyncio.create_task(process_backtest_queue())
    asyncio.create_task(process_greeks_queue())