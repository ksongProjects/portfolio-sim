"use client";

import { useState, useEffect, useCallback, useRef } from "react";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface IntradayBarData {
  symbol: string;
  interval: string;
  timestamp: string;
  open?: number;
  high?: number;
  low?: number;
  close: number;
  volume?: number;
}

export function useMarketSocket(symbols: string[] = [], enabled = true) {
  const [bars, setBars] = useState<Map<string, IntradayBarData>>(new Map());
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const symbolsKey = symbols.join(",");

  const updateBar = useCallback((bar: IntradayBarData) => {
    setBars((prev) => {
      const next = new Map(prev);
      next.set(bar.symbol, bar);
      return next;
    });
  }, []);

  useEffect(() => {
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    if (!enabled) {
      return;
    }

    const params = new URLSearchParams();
    for (const sym of symbolsKey.split(",").filter(Boolean)) {
      params.append("symbols", sym);
    }
    const query = params.toString();
    const url = `ws://${API_BASE.replace("http://", "")}/api/ws/market${query ? "?" + query : ""}`;

    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      setConnected(true);
      setError(null);
    };

    ws.onmessage = (event) => {
      try {
        const bar: IntradayBarData = JSON.parse(event.data);
        updateBar(bar);
      } catch {
      }
    };

    ws.onerror = () => {
      setConnected(false);
      setError("WebSocket connection error");
    };

    ws.onclose = () => {
      setConnected(false);
    };

    return () => {
      ws.close();
      if (wsRef.current === ws) {
        wsRef.current = null;
      }
    };
  }, [symbolsKey, updateBar, enabled]);

  const getBar = useCallback(
    (symbol: string): IntradayBarData | undefined => {
      return enabled ? bars.get(symbol) : undefined;
    },
    [bars, enabled]
  );

  const getAllBars = useCallback((): IntradayBarData[] => {
    return enabled ? Array.from(bars.values()) : [];
  }, [bars, enabled]);

  return {
    bars: enabled ? bars : new Map(),
    getBar,
    getAllBars,
    connected: enabled && connected,
    error: enabled ? error : null,
  };
}