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

interface UsePortfolioOptions {
	includeIndices?: boolean;
}

interface AddPositionVariables {
	targetPortfolioId: string;
	symbol: string;
	shares: number;
	price: number;
}

interface AddPositionContext {
	previousPositions?: Position[];
	previousSummary?: PortfolioSummary | null;
}

export function usePortfolio(portfolioId = "default", options: UsePortfolioOptions = {}) {
	const queryClient = useQueryClient();
	const includeIndices = options.includeIndices ?? true;

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
		enabled: includeIndices,
	});

	const addPositionMutation = useMutation({
		mutationFn: async ({ targetPortfolioId, symbol, shares, price }: AddPositionVariables) => {
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
		onMutate: async (variables): Promise<AddPositionContext> => {
			await Promise.all([
				queryClient.cancelQueries({
					queryKey: ["portfolio", "positions", variables.targetPortfolioId],
				}),
				queryClient.cancelQueries({
					queryKey: ["portfolio", "summary", variables.targetPortfolioId],
				}),
			]);

			const previousPositions = queryClient.getQueryData<Position[]>([
				"portfolio",
				"positions",
				variables.targetPortfolioId,
			]);
			const previousSummary = queryClient.getQueryData<PortfolioSummary | null>([
				"portfolio",
				"summary",
				variables.targetPortfolioId,
			]);

			const symbol = variables.symbol.trim().toUpperCase();
			const now = new Date().toISOString();
			const costBasis = variables.shares * variables.price;
			const optimisticPosition: Position = {
				ID: `optimistic:${symbol}:${now}`,
				PortfolioID: variables.targetPortfolioId,
				TickerID: "",
				Symbol: symbol,
				CompanyName: symbol,
				Quantity: variables.shares,
				AvgCost: variables.price,
				CurrentPrice: variables.price,
				CurrentValue: costBasis,
				DayChange: 0,
				DayChangePct: 0,
				TotalGain: 0,
				TotalGainPct: 0,
				OpenedAt: now,
			};

			queryClient.setQueryData<Position[]>(
				["portfolio", "positions", variables.targetPortfolioId],
				(current = []) => [optimisticPosition, ...current]
			);

			if (previousSummary) {
				const totalValue = previousSummary.TotalValue + costBasis;
				const totalInvested = previousSummary.TotalInvested + costBasis;
				const totalGain = totalValue - totalInvested;

				queryClient.setQueryData<PortfolioSummary>(
					["portfolio", "summary", variables.targetPortfolioId],
					{
						...previousSummary,
						TotalValue: totalValue,
						TotalInvested: totalInvested,
						TotalGain: totalGain,
						TotalGainPct: totalInvested === 0 ? 0 : (totalGain / totalInvested) * 100,
						DayChangePct: totalValue === 0 ? 0 : (previousSummary.DayChange / totalValue) * 100,
					}
				);
			}

			return { previousPositions, previousSummary };
		},
		onError: (_error, variables, context) => {
			if (context?.previousPositions) {
				queryClient.setQueryData(
					["portfolio", "positions", variables.targetPortfolioId],
					context.previousPositions
				);
			}
			if (context?.previousSummary) {
				queryClient.setQueryData(
					["portfolio", "summary", variables.targetPortfolioId],
					context.previousSummary
				);
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
		(includeIndices ? indicesQuery.error : null) ??
		addPositionMutation.error;

	const positionsLoading = positionsQuery.isLoading;
	const summaryLoading = summaryQuery.isLoading;
	const indicesLoading = includeIndices && indicesQuery.isLoading;

	return {
		positions: positionsQuery.data ?? [],
		summary: summaryQuery.data ?? null,
		indices: indicesQuery.data ?? [],
		loading: positionsLoading || summaryLoading,
		positionsLoading,
		summaryLoading,
		indicesLoading,
		isAddingPosition: addPositionMutation.isPending,
		error: combinedError ? getErrorMessage(combinedError) : null,
		fetchPositions,
		fetchSummary,
		fetchIndices,
		refresh,
		addPosition,
	};
}
