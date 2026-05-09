"use client";

import { useState, useEffect, useCallback, useRef } from "react";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface MarketTick {
	TickerID: string;
	Symbol: string;
	Price: number;
	Bid: number;
	Ask: number;
	Volume: number;
	SourceID: string;
	Timestamp: string;
}

export function useMarketStream(symbols: string[] = []) {
	const [ticks, setTicks] = useState<Map<string, MarketTick>>(new Map());
	const [connected, setConnected] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const esRef = useRef<EventSource | null>(null);

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
		}

		const symbolsParam = symbols.length > 0 ? `symbols=${symbols.join(",")}` : "";
		const url = `${API_BASE}/api/stream/market${symbolsParam ? "?" + symbolsParam : ""}`;

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
			es.close();
			setTimeout(() => {
				if (esRef.current === es) {
					const reconnectEs = new EventSource(url);
					esRef.current = reconnectEs;
				}
			}, 5000);
		};

		return () => {
			es.close();
		};
	}, [symbols.join(","), updateTick]);

	const getTick = useCallback(
		(symbol: string): MarketTick | undefined => {
			return ticks.get(symbol);
		},
		[ticks]
	);

	const getAllTicks = useCallback((): MarketTick[] => {
		return Array.from(ticks.values());
	}, [ticks]);

	return {
		ticks,
		getTick,
		getAllTicks,
		connected,
		error,
	};
}