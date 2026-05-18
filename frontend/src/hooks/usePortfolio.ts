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
	Sector?: string;
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

export interface PortfolioPerformancePoint {
	timestamp: string;
	value: number;
}

export interface PortfolioPerformanceData {
	data: PortfolioPerformancePoint[];
	interval: string;
	range: string;
}

async function fetchPerformanceData(portfolioId: string, range: string) {
	return fetchJson<PortfolioPerformanceData>(
		`/api/portfolio/performance?portfolio_id=${encodeURIComponent(portfolioId)}&range=${range}`,
		undefined,
		"Failed to fetch performance"
	);
}

export function usePortfolioPerformance(portfolioId = "default", range = "1d") {
	return useQuery({
		queryKey: ["portfolio", "performance", portfolioId, range],
		queryFn: () => fetchPerformanceData(portfolioId, range),
		enabled: Boolean(portfolioId),
		refetchInterval: range === "1d" ? 60000 : 300000,
	});
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

interface RemovePositionContext {
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
			const portfolioId = targetPortfolioId === "default" ? "00000000-0000-0000-0000-000000000001" : targetPortfolioId;
			const res = await apiFetch("/api/portfolio/positions", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					portfolio_id: portfolioId,
					symbol,
					shares,
					price,
				}),
			});

			if (!res.ok) {
				const errText = await res.text();
				throw new Error(errText || "Failed to add position");
			}
		},
		onMutate: async (variables): Promise<AddPositionContext> => {
			const targetPortfolioId = variables.targetPortfolioId === "default" ? "00000000-0000-0000-0000-000000000001" : variables.targetPortfolioId;
			await Promise.all([
				queryClient.cancelQueries({
					queryKey: ["portfolio", "positions", targetPortfolioId],
				}),
				queryClient.cancelQueries({
					queryKey: ["portfolio", "summary", targetPortfolioId],
				}),
			]);

			const previousPositions = queryClient.getQueryData<Position[]>(["portfolio", "positions", targetPortfolioId]);
			const previousSummary = queryClient.getQueryData<PortfolioSummary | null>(["portfolio", "summary", targetPortfolioId]);

			const symbol = variables.symbol.trim().toUpperCase();
			const costBasis = variables.shares * variables.price;

			queryClient.setQueryData<Position[]>(["portfolio", "positions", targetPortfolioId], (current = []) => {
				const existingIdx = current.findIndex(p => p.Symbol === symbol);
				if (existingIdx >= 0) {
					const existing = current[existingIdx];
					const newTotalQty = existing.Quantity + variables.shares;
					const newAvgCost = (existing.Quantity * existing.AvgCost + variables.shares * variables.price) / newTotalQty;
					const updated = { ...existing, Quantity: newTotalQty, AvgCost: newAvgCost, CurrentValue: newTotalQty * existing.CurrentPrice };
					return [...current.slice(0, existingIdx), updated, ...current.slice(existingIdx + 1)];
				}
				const now = new Date().toISOString();
				const optimisticPosition: Position = {
					ID: `optimistic:${symbol}:${now}`,
					PortfolioID: targetPortfolioId,
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
				return [optimisticPosition, ...current];
			});

			if (previousSummary) {
				const existingPos = previousPositions?.find(p => p.Symbol === symbol);
				const addQty = existingPos ? 0 : variables.shares;
				const addCost = existingPos ? 0 : costBasis;
				queryClient.setQueryData<PortfolioSummary>(["portfolio", "summary", targetPortfolioId], {
					...previousSummary,
					TotalValue: previousSummary.TotalValue + costBasis,
					TotalInvested: previousSummary.TotalInvested + costBasis,
					TotalGain: previousSummary.TotalValue + costBasis - (previousSummary.TotalInvested + costBasis),
					TotalGainPct: previousSummary.TotalInvested + costBasis === 0 ? 0 : ((previousSummary.TotalValue + costBasis - (previousSummary.TotalInvested + costBasis)) / (previousSummary.TotalInvested + costBasis)) * 100,
					DayChangePct: previousSummary.DayChangePct,
				});
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

	const removePositionMutation = useMutation({
		mutationFn: async ({ targetPortfolioId, positionId }: { targetPortfolioId: string; positionId: string }) => {
			const portfolioId = targetPortfolioId === "default" ? "00000000-0000-0000-0000-000000000001" : targetPortfolioId;
			const res = await apiFetch(`/api/portfolio/positions?portfolio_id=${encodeURIComponent(portfolioId)}&position_id=${encodeURIComponent(positionId)}`, {
				method: "DELETE",
			});
			if (!res.ok) {
				const errText = await res.text();
				throw new Error(errText || "Failed to remove position");
			}
		},
		onMutate: async (variables): Promise<RemovePositionContext> => {
			const targetPortfolioId = variables.targetPortfolioId === "default" ? "00000000-0000-0000-0000-000000000001" : variables.targetPortfolioId;
			await Promise.all([
				queryClient.cancelQueries({
					queryKey: ["portfolio", "positions", targetPortfolioId],
				}),
				queryClient.cancelQueries({
					queryKey: ["portfolio", "summary", targetPortfolioId],
				}),
			]);
			const previousPositions = queryClient.getQueryData<Position[]>(["portfolio", "positions", targetPortfolioId]);
			const previousSummary = queryClient.getQueryData<PortfolioSummary | null>(["portfolio", "summary", targetPortfolioId]);
			queryClient.setQueryData<Position[]>(
				["portfolio", "positions", targetPortfolioId],
				(current = []) => current.filter(p => p.ID !== variables.positionId)
			);
			return { previousPositions, previousSummary };
		},
		onError: (_error, variables, context) => {
			const targetPortfolioId = variables.targetPortfolioId === "default" ? "00000000-0000-0000-0000-000000000001" : variables.targetPortfolioId;
			if (context?.previousPositions) {
				queryClient.setQueryData(["portfolio", "positions", targetPortfolioId], context.previousPositions);
			}
			if (context?.previousSummary) {
				queryClient.setQueryData(["portfolio", "summary", targetPortfolioId], context.previousSummary);
			}
		},
		onSuccess: async (_, variables) => {
			const targetPortfolioId = variables.targetPortfolioId === "default" ? "00000000-0000-0000-0000-000000000001" : variables.targetPortfolioId;
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: ["portfolio", "positions", targetPortfolioId] }),
				queryClient.invalidateQueries({ queryKey: ["portfolio", "summary", targetPortfolioId] }),
			]);
		},
	});

	const removePosition = useCallback(
		async (targetPortfolioId = portfolioId, positionId: string) => {
			await removePositionMutation.mutateAsync({ targetPortfolioId, positionId });
		},
		[removePositionMutation, portfolioId]
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
		isRemovingPosition: removePositionMutation.isPending,
		error: combinedError ? getErrorMessage(combinedError) : null,
		fetchPositions,
		fetchSummary,
		fetchIndices,
		refresh,
		addPosition,
		removePosition,
	};
}
