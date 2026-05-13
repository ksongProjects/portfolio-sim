"use client";

import { useState, useEffect, useCallback, useRef } from "react";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
const EMPTY_TICKS = new Map<string, MarketTick>();

export interface MarketTick {
	TickerID: string;
	Symbol: string;
	Price: number;
	Change?: number;
	ChangePct?: number;
	Bid: number;
	Ask: number;
	Volume: number;
	SourceID: string;
	Timestamp: string;
}

export function useMarketStream(symbols: string[] = [], enabled = true) {
	const [ticks, setTicks] = useState<Map<string, MarketTick>>(new Map());
	const [connected, setConnected] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const esRef = useRef<EventSource | null>(null);
	const symbolsKey = symbols.join(",");

	const updateTick = useCallback((tick: MarketTick) => {
		setTicks((prev) => {
			const next = new Map(prev);
			next.set(tick.Symbol, tick);
			return next;
		});
	}, []);

	useEffect(() => {
		if (esRef.current) {
			esRef.current.close();
			esRef.current = null;
		}

		if (!enabled) {
			return;
		}

		const params = new URLSearchParams();
		for (const symbol of symbolsKey.split(",").filter(Boolean)) {
			params.append("symbols", symbol);
		}
		const query = params.toString();
		const url = `${API_BASE}/api/stream/market${query ? "?" + query : ""}`;

		const es = new EventSource(url);
		esRef.current = es;

		es.onopen = () => {
			setConnected(true);
			setError(null);
		};

		es.onmessage = (event) => {
			try {
				const tick: MarketTick = JSON.parse(event.data);
				updateTick(tick);
			} catch {
			}
		};

		es.onerror = () => {
			setConnected(false);
			setError("Stream connection lost");
		};

		return () => {
			es.close();
			if (esRef.current === es) {
				esRef.current = null;
			}
		};
	}, [symbolsKey, updateTick, enabled]);

	const getTick = useCallback(
		(symbol: string): MarketTick | undefined => {
			return enabled ? ticks.get(symbol) : undefined;
		},
		[ticks, enabled]
	);

	const getAllTicks = useCallback((): MarketTick[] => {
		return enabled ? Array.from(ticks.values()) : [];
	}, [ticks, enabled]);

	return {
		ticks: enabled ? ticks : EMPTY_TICKS,
		getTick,
		getAllTicks,
		connected: enabled && connected,
		error: enabled ? error : null,
	};
}
