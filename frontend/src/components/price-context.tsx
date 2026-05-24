"use client";

import { createContext, useContext, useEffect, useState, useCallback, useRef, type ReactNode } from "react";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

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

interface PriceContextValue {
	prices: Map<string, MarketTick>;
	getPrice: (symbol: string) => MarketTick | undefined;
	connected: boolean;
}

const PriceContext = createContext<PriceContextValue | null>(null);

export function usePriceContext() {
	const ctx = useContext(PriceContext);
	if (!ctx) throw new Error("usePriceContext must be used within PriceProvider");
	return ctx;
}

export function useLivePrice(symbol: string | undefined): { price: number | null; changePct: number | null; tick: MarketTick | undefined } {
	const ctx = usePriceContext();
	const tick = symbol ? ctx.getPrice(symbol) : undefined;
	return {
		price: tick?.Price ?? null,
		changePct: tick?.ChangePct ?? null,
		tick,
	};
}

export function PriceProvider({ children }: { children: ReactNode }) {
	const [prices, setPrices] = useState<Map<string, MarketTick>>(new Map());
	const [connected, setConnected] = useState(false);
	const esRef = useRef<EventSource | null>(null);
	const updateTickRef = useRef<((tick: MarketTick) => void) | null>(null);

	const updateTick = useCallback((tick: MarketTick) => {
		setPrices((prev) => {
			const next = new Map(prev);
			next.set(tick.Symbol, tick);
			return next;
		});
	}, []);

	useEffect(() => {
		updateTickRef.current = updateTick;
	}, [updateTick]);

	useEffect(() => {
		if (esRef.current) {
			esRef.current.close();
			esRef.current = null;
		}

		const url = `${API_BASE}/api/stream/market`;

		const es = new EventSource(url);
		esRef.current = es;

		es.onopen = () => setConnected(true);
		es.onerror = () => setConnected(false);

		es.onmessage = (event) => {
			try {
				const tick: MarketTick = JSON.parse(event.data);
				updateTickRef.current?.(tick);
			} catch {}
		};

		return () => {
			es.close();
			esRef.current = null;
		};
	}, []);

	const value: PriceContextValue = {
		prices,
		getPrice: useCallback((symbol: string) => prices.get(symbol), [prices]),
		connected,
	};

	return (
		<PriceContext.Provider value={value}>
			{children}
		</PriceContext.Provider>
	);
}