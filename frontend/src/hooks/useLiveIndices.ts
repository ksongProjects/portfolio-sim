"use client";

import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { useMarketStream } from "./useMarketStream";
import { useMarketStatus } from "@/hooks/useMarketStatus";
import { fetchJson, getErrorMessage } from "@/lib/api";
import type { MarketIndex } from "./usePortfolio";

export type LiveIndex = MarketIndex;
const EMPTY_INDICES: LiveIndex[] = [];

async function fetchBaseIndices(): Promise<LiveIndex[]> {
	const indices = await fetchJson<LiveIndex[]>("/api/market/indices", undefined, "Failed to fetch base indices");
	return Array.isArray(indices) ? indices : [];
}

export function useLiveIndices() {
	const { isLive } = useMarketStatus();
	const baseIndicesQuery = useQuery({
		queryKey: ["portfolio", "indices"],
		queryFn: fetchBaseIndices,
		enabled: isLive,
	});

	const baseIndices = baseIndicesQuery.data ?? EMPTY_INDICES;
	const symbols = useMemo(() => baseIndices.map(index => index.Symbol), [baseIndices]);

	const { ticks, connected, error: streamError } = useMarketStream(
		symbols,
		isLive && !baseIndicesQuery.isLoading && symbols.length > 0
	);

	const indices: LiveIndex[] = useMemo(() => {
		return baseIndices.map(index => {
			const tick = ticks.get(index.Symbol);

			return tick
				? {
						...index,
						Price: tick.Price,
						Change: tick.Change ?? index.Change,
						ChangePct: tick.ChangePct ?? index.ChangePct,
					}
				: index;
		});
	}, [ticks, baseIndices]);

	return {
		indices,
		connected,
		isLive,
		loading: baseIndicesQuery.isLoading,
		error: baseIndicesQuery.error ? getErrorMessage(baseIndicesQuery.error) : streamError,
	};
}
