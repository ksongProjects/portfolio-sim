from typing import Literal

import numpy as np
import pandas as pd


def bootstrap_return_paths(returns: pd.Series, n_paths: int = 1000, n_steps: int = 60) -> np.ndarray:
    data = returns.dropna().values
    if len(data) == 0:
        return np.zeros((n_paths, n_steps))
    indices = np.random.randint(0, len(data), size=(n_paths, n_steps))
    return data[indices]


def geometric_brownian_motion_paths(
    s0: float,
    mu: float,
    sigma: float,
    n_paths: int = 1000,
    n_steps: int = 60,
    dt: float = 1.0,
) -> np.ndarray:
    Z = np.random.standard_normal((n_paths, n_steps))
    drift = (mu - 0.5 * sigma**2) * dt
    diffusion = sigma * np.sqrt(dt) * Z
    log_returns = drift + diffusion
    return s0 * np.exp(np.cumsum(log_returns, axis=1))


def cholesky_correlated_shocks(corr_matrix: np.ndarray, n_shocks: int) -> np.ndarray:
    L = np.linalg.cholesky(corr_matrix)
    independent = np.random.standard_normal((n_shocks, corr_matrix.shape[0]))
    return independent @ L.T


def correlated_return_paths(
    returns: pd.DataFrame,
    n_paths: int = 1000,
    n_steps: int = 60,
) -> np.ndarray:
    corr = returns.corr().values
    L = np.linalg.cholesky(corr)
    standardized = (returns - returns.mean()).dropna().values
    if standardized.shape[0] < n_steps:
        return np.zeros((n_paths, n_steps, corr.shape[0]))
    indices = np.random.randint(0, len(standardized), size=(n_paths, n_steps))
    shocks = standardized[indices]
    return shocks @ L.T


def student_t_shocks(df: float, size: tuple[int, int]) -> np.ndarray:
    return np.random.standard_t(df, size=size)


def student_t_tail_report(
    returns: pd.Series,
    confidence: float = 0.95,
    n_paths: int = 5000,
) -> dict:
    alpha = 1 - confidence
    sorted_returns = np.sort(returns.dropna().values)
    VaR = sorted_returns[int(len(sorted_returns) * alpha)]
    CVaR = sorted_returns[: int(len(sorted_returns) * alpha)].mean()

    paths = bootstrap_return_paths(returns, n_paths=n_paths)
    path_final = paths[:, -1] if paths.shape[1] > 0 else paths
    worst = np.percentile(path_final, alpha * 100)

    return {
        "VaR": round(float(VaR), 6),
        "CVaR": round(float(CVaR), 6),
        "worst_case": round(float(worst), 6),
        "confidence": confidence,
        "n_paths": n_paths,
    }