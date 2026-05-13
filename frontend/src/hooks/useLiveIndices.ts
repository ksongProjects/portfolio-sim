"use client";

import { useMemo, useEffect, useState } from "react";
import { useMarketStream } from "./useMarketStream";
import { useMarketStatus } from "@/components/live-market-indicator";
import { fetchJson } from "@/lib/api";

export interface LiveIndex {
	Symbol: string;
	Name: string;
	Price: number;
	Change: number;
	ChangePct: number;
}

interface IndexConfig {
	symbol: string;
	name: string;
}

interface MarketIndex {
	Symbol: string;
	Name: string;
	Price: number;
	Change: number;
	ChangePct: number;
}

const DEFAULT_INDICES: IndexConfig[] = [
	{ symbol: "SPY", name: "S&P 500" },
	{ symbol: "QQQ", name: "Nasdaq-100" },
	{ symbol: "DIA", name: "Dow Jones" },
	{ symbol: "IWM", name: "Russell 2000" },
];

function getDefaultIndicesConfig(): IndexConfig[] {
	return DEFAULT_INDICES;
}

async function fetchBaseIndices(): Promise<Map<string, { price: number; change: number; changePct: number }>> {
	try {
		const indices = await fetchJson<MarketIndex[]>("/api/market/indices", undefined, "Failed to fetch base indices");
		const map = new Map<string, { price: number; change: number; changePct: number }>();
		for (const idx of indices) {
			map.set(idx.Symbol, { price: idx.Price, change: idx.Change, changePct: idx.ChangePct });
		}
		return map;
	} catch {
		return new Map();
	}
}

export function useLiveIndices() {
	const { isLive } = useMarketStatus();
	const [indicesConfig] = useState<IndexConfig[]>(getDefaultIndicesConfig);
	const [baseData, setBaseData] = useState<Map<string, { price: number; change: number; changePct: number }>>(new Map());

	const symbols = useMemo(() => indicesConfig.map(i => i.symbol), [indicesConfig]);

	const { ticks, connected } = useMarketStream(symbols);

	useEffect(() => {
		if (isLive && baseData.size === 0) {
			fetchBaseIndices().then(setBaseData);
		}
	}, [isLive, baseData.size]);

	const streamReady = connected && ticks.size > 0;

	const indices: LiveIndex[] = useMemo(() => {
		return indicesConfig.map(config => {
			const tick = ticks.get(config.symbol);
			const base = baseData.get(config.symbol);

			if (tick && isLive && streamReady) {
				return {
					Symbol: config.symbol,
					Name: config.name,
					Price: tick.Price,
					Change: 0,
					ChangePct: 0,
				};
			}

			if (base) {
				return {
					Symbol: config.symbol,
					Name: config.name,
					Price: tick?.Price ?? base.price,
					Change: base.change,
					ChangePct: base.changePct,
				};
			}

			return {
				Symbol: config.symbol,
				Name: config.name,
				Price: tick?.Price ?? 0,
				Change: 0,
				ChangePct: 0,
			};
		});
	}, [ticks, baseData, streamReady, indicesConfig, isLive]);

	return {
		indices,
		connected,
		isLive,
	};
}