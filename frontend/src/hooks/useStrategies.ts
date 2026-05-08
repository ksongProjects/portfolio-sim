"use client";

import { useState, useCallback } from "react";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface Strategy {
  ID: string;
  Name: string;
  Status: string;
  Returns: number;
  Sharpe: number;
  MaxDD: number;
  Trades: number;
  WinRate: number;
}

export interface Signal {
  ID: string;
  Strategy: string;
  Symbol: string;
  Action: string;
  Price: number;
  Confidence: string;
  Timestamp: string;
}

export function useStrategies() {
  const [strategies, setStrategies] = useState<Strategy[]>([]);
  const [signals, setSignals] = useState<Signal[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchStrategies = useCallback(async () => {
    try {
      const res = await fetch(`${API_BASE}/api/strategies`);
      if (!res.ok) throw new Error("Failed to fetch strategies");
      const data = await res.json();
      setStrategies(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
    }
  }, []);

  const fetchSignals = useCallback(async (limit = 10) => {
    try {
      const res = await fetch(`${API_BASE}/api/signals?limit=${limit}`);
      if (!res.ok) throw new Error("Failed to fetch signals");
      const data = await res.json();
      setSignals(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
    }
  }, []);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    await Promise.all([fetchStrategies(), fetchSignals()]);
    setLoading(false);
  }, [fetchStrategies, fetchSignals]);

  return { strategies, signals, loading, error, fetchStrategies, fetchSignals, refresh };
}