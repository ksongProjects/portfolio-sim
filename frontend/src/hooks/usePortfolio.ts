"use client";

import { useCallback } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiFetch, fetchJson, getErrorMessage } from "@/lib/api";

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

async function fetchPositionsData(portfolioId: string) {
	return fetchJson<Position[]>(
		`/api/portfolio/positions?portfolio_id=${encodeURIComponent(portfolioId)}`,
		undefined,
		"Failed to fetch positions"
	);
}

async function fetchSummaryData(portfolioId: string) {
	return fetchJson<PortfolioSummary>(
		`/api/portfolio/summary?portfolio_id=${encodeURIComponent(portfolioId)}`,
		undefined,
		"Failed to fetch summary"
	);
}

async function fetchIndicesData() {
	return fetchJson<MarketIndex[]>("/api/market/indices", undefined, "Failed to fetch indices");
}

export function usePortfolio(portfolioId = "default") {
	const queryClient = useQueryClient();

	const positionsQuery = useQuery({
		queryKey: ["portfolio", "positions", portfolioId],
		queryFn: () => fetchPositionsData(portfolioId),
	});

	const summaryQuery = useQuery({
		queryKey: ["portfolio", "summary", portfolioId],
		queryFn: () => fetchSummaryData(portfolioId),
	});

	const indicesQuery = useQuery({
		queryKey: ["portfolio", "indices"],
		queryFn: fetchIndicesData,
	});

	const addPositionMutation = useMutation({
		mutationFn: async ({
			targetPortfolioId,
			symbol,
			shares,
			price,
		}: {
			targetPortfolioId: string;
			symbol: string;
			shares: number;
			price: number;
		}) => {
			const res = await apiFetch("/api/portfolio/positions", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					portfolio_id: targetPortfolioId,
					symbol,
					shares,
					price,
				}),
			});

			if (!res.ok) {
				throw new Error("Failed to add position");
			}
		},
		onSuccess: async (_, variables) => {
			await Promise.all([
				queryClient.invalidateQueries({
					queryKey: ["portfolio", "positions", variables.targetPortfolioId],
				}),
				queryClient.invalidateQueries({
					queryKey: ["portfolio", "summary", variables.targetPortfolioId],
				}),
				queryClient.invalidateQueries({ queryKey: ["portfolio", "indices"] }),
			]);
		},
	});

	const fetchPositions = useCallback(
		async (targetPortfolioId = portfolioId) => {
			if (targetPortfolioId !== portfolioId) {
				await queryClient.fetchQuery({
					queryKey: ["portfolio", "positions", targetPortfolioId],
					queryFn: () => fetchPositionsData(targetPortfolioId),
				});
				return;
			}

			await positionsQuery.refetch();
		},
		[portfolioId, positionsQuery, queryClient]
	);

	const fetchSummary = useCallback(
		async (targetPortfolioId = portfolioId) => {
			if (targetPortfolioId !== portfolioId) {
				await queryClient.fetchQuery({
					queryKey: ["portfolio", "summary", targetPortfolioId],
					queryFn: () => fetchSummaryData(targetPortfolioId),
				});
				return;
			}

			await summaryQuery.refetch();
		},
		[portfolioId, queryClient, summaryQuery]
	);

	const fetchIndices = useCallback(async () => {
		await indicesQuery.refetch();
	}, [indicesQuery]);

	const refresh = useCallback(
		async (targetPortfolioId = portfolioId) => {
			if (targetPortfolioId !== portfolioId) {
				await Promise.all([
					queryClient.fetchQuery({
						queryKey: ["portfolio", "positions", targetPortfolioId],
						queryFn: () => fetchPositionsData(targetPortfolioId),
					}),
					queryClient.fetchQuery({
						queryKey: ["portfolio", "summary", targetPortfolioId],
						queryFn: () => fetchSummaryData(targetPortfolioId),
					}),
					queryClient.fetchQuery({
						queryKey: ["portfolio", "indices"],
						queryFn: fetchIndicesData,
					}),
				]);
				return;
			}

			await Promise.all([
				positionsQuery.refetch(),
				summaryQuery.refetch(),
				indicesQuery.refetch(),
			]);
		},
		[indicesQuery, portfolioId, positionsQuery, queryClient, summaryQuery]
	);

	const addPosition = useCallback(
		async (
			targetPortfolioId = portfolioId,
			symbol: string,
			shares: number,
			price: number
		) => {
			await addPositionMutation.mutateAsync({
				targetPortfolioId,
				symbol,
				shares,
				price,
			});
		},
		[addPositionMutation, portfolioId]
	);

	const combinedError =
		positionsQuery.error ??
		summaryQuery.error ??
		indicesQuery.error ??
		addPositionMutation.error;

	return {
		positions: positionsQuery.data ?? [],
		summary: summaryQuery.data ?? null,
		indices: indicesQuery.data ?? [],
		loading:
			positionsQuery.isLoading ||
			summaryQuery.isLoading ||
			indicesQuery.isLoading ||
			addPositionMutation.isPending,
		error: combinedError ? getErrorMessage(combinedError) : null,
		fetchPositions,
		fetchSummary,
		fetchIndices,
		refresh,
		addPosition,
	};
}
