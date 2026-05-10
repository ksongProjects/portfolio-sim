"use client";

import { useState, useCallback } from "react";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface Position {
	ID: string;
	PortfolioID: string;
	TickerID: string;
	Symbol: string;
	CompanyName: string;
	Quantity: number;
	AvgCost: number;
	CurrentPrice: number;
	CurrentValue: number;
	DayChange: number;
	DayChangePct: number;
	TotalGain: number;
	TotalGainPct: number;
	OpenedAt: string;
}

export interface PortfolioSummary {
	TotalValue: number;
	DayChange: number;
	DayChangePct: number;
	TotalInvested: number;
	TotalGain: number;
	TotalGainPct: number;
	CashBalance: number;
}

export interface MarketIndex {
	Symbol: string;
	Name: string;
	Price: number;
	Change: number;
	ChangePct: number;
}

export interface Trade {
	ID: string;
	Type: string;
	Symbol: string;
	Shares: number;
	Price: number;
	Total: number;
	Timestamp: string;
}

export function usePortfolio() {
	const [positions, setPositions] = useState<Position[]>([]);
	const [summary, setSummary] = useState<PortfolioSummary | null>(null);
	const [indices, setIndices] = useState<MarketIndex[]>([]);
	const [loading, setLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);

	const fetchPositions = useCallback(async (portfolioId = "default") => {
		try {
			const res = await fetch(`${API_BASE}/api/portfolio/positions?portfolio_id=${portfolioId}`);
			if (!res.ok) throw new Error("Failed to fetch positions");
			const data = await res.json();
			setPositions(data);
		} catch (err) {
			setError(err instanceof Error ? err.message : "Unknown error");
		}
	}, []);

	const fetchSummary = useCallback(async (portfolioId = "default") => {
		try {
			const res = await fetch(`${API_BASE}/api/portfolio/summary?portfolio_id=${portfolioId}`);
			if (!res.ok) throw new Error("Failed to fetch summary");
			const data = await res.json();
			setSummary(data);
		} catch (err) {
			setError(err instanceof Error ? err.message : "Unknown error");
		}
	}, []);

	const fetchIndices = useCallback(async () => {
		try {
			const res = await fetch(`${API_BASE}/api/market/indices`);
			if (!res.ok) throw new Error("Failed to fetch indices");
			const data = await res.json();
			setIndices(data);
		} catch (err) {
			setError(err instanceof Error ? err.message : "Unknown error");
		}
	}, []);

	const refresh = useCallback(async (portfolioId = "default") => {
		setLoading(true);
		setError(null);
		await Promise.all([fetchPositions(portfolioId), fetchSummary(portfolioId), fetchIndices()]);
		setLoading(false);
	}, [fetchPositions, fetchSummary, fetchIndices]);

	const addPosition = useCallback(async (portfolioId = "default", symbol: string, shares: number, price: number) => {
		const res = await fetch(`${API_BASE}/api/portfolio/positions`, {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ portfolio_id: portfolioId, symbol, shares, price }),
		});
		if (!res.ok) throw new Error("Failed to add position");
		await refresh(portfolioId);
	}, [refresh]);

	return { positions, summary, indices, loading, error, fetchPositions, fetchSummary, fetchIndices, refresh, addPosition };
}