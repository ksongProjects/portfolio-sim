"use client";

import { useMemo } from "react";
import { usePortfolio, type Position, type PortfolioSummary, type MarketIndex } from "./usePortfolio";
import { usePriceContext } from "@/components/price-context";

export function useLivePositions(portfolioId = "default", options: { includeIndices?: boolean } = {}) {
	const includeIndices = options.includeIndices ?? true;
	const { positions, summary, indices, isLoading, error, addPosition, removePosition } = usePortfolio(portfolioId, {
		includeIndices,
	});
	const { prices } = usePriceContext();

	const livePositions: Position[] = useMemo(() => {
		if (!positions) return [];
		return positions.map((pos) => {
			const tick = prices.get(pos.Symbol);
			if (!tick) return pos;
			const currentPrice = tick.Price;
			const currentValue = pos.Quantity * currentPrice;
			const dayChangePct = tick.ChangePct ?? pos.DayChangePct;
			const dayChange = tick.Change ?? (currentPrice * dayChangePct / 100);
			const costBasis = pos.Quantity * pos.AvgCost;
			const totalGain = currentValue - costBasis;
			const totalGainPct = costBasis > 0 ? (totalGain / costBasis) * 100 : 0;
			return {
				...pos,
				CurrentPrice: currentPrice,
				CurrentValue: currentValue,
				DayChange: dayChange,
				DayChangePct: dayChangePct,
				TotalGain: totalGain,
				TotalGainPct: totalGainPct,
			};
		});
	}, [positions, prices]);

	const liveIndices: MarketIndex[] = useMemo(() => {
		if (!indices) return [];
		return indices.map((index) => {
			const tick = prices.get(index.Symbol);
			if (!tick) return index;
			return {
				...index,
				Price: tick.Price,
				Change: tick.Change ?? index.Change,
				ChangePct: tick.ChangePct ?? index.ChangePct,
			};
		});
	}, [indices, prices]);

	const liveSummary: PortfolioSummary | undefined = useMemo(() => {
		if (!summary) return undefined;
		const updatedPositions = livePositions;
		const totalValue = updatedPositions.reduce((sum, p) => sum + p.CurrentValue, 0);
		const totalInvested = updatedPositions.reduce((sum, p) => sum + p.Quantity * p.AvgCost, 0);
		const totalGain = totalValue - totalInvested;
		const totalGainPct = totalInvested > 0 ? (totalGain / totalInvested) * 100 : 0;
		return {
			...summary,
			TotalValue: totalValue,
			TotalInvested: totalInvested,
			TotalGain: totalGain,
			TotalGainPct: totalGainPct,
		};
	}, [summary, livePositions]);

	return {
		positions: livePositions,
		summary: liveSummary,
		indices: liveIndices,
		isLoading,
		error,
		addPosition,
		removePosition,
	};
}