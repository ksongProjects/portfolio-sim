"use client";

import { useState, useCallback } from "react";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface TickerDetails {
  symbol: string;
  name: string;
  exchange: string;
  isTradeable: boolean;
  sector: string;
  industry: string;
  price: number;
  change: number;
  changePct: number;
  volume: number;
  avgVolume: number;
  marketCap: number;
  peRatio: number;
  eps: number;
  dividendYield: number;
  week52High: number;
  week52Low: number;
}

export interface IntradayBar {
  timestamp: string;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

export interface FinancialRatio {
  label: string;
  value: string;
  description: string;
}

export function useTickerLookup() {
  const [searchResults, setSearchResults] = useState<TickerDetails[]>([]);
  const [selectedTicker, setSelectedTicker] = useState<TickerDetails | null>(null);
  const [intradayData, setIntradayData] = useState<IntradayBar[]>([]);
  const [ratios, setRatios] = useState<FinancialRatio[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchLoading, setSearchLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const searchTickers = useCallback(async (query: string) => {
    if (query.length < 1) {
      setSearchResults([]);
      return;
    }
    setSearchLoading(true);
    try {
      const res = await fetch(`${API_BASE}/api/tickers/search?q=${encodeURIComponent(query)}`);
      if (!res.ok) throw new Error("Search failed");
      const data = await res.json();
      const results = Array.isArray(data) ? data : [];
      const seen = new Set<string>();
      setSearchResults(results.filter((t) => seen.has(t.symbol) ? false : seen.add(t.symbol)));
    } catch {
      setSearchResults([]);
    } finally {
      setSearchLoading(false);
    }
  }, []);

  const lookupTicker = useCallback(async (symbol: string) => {
    setLoading(true);
    setError(null);
    try {
      const [detailsRes, intradayRes, ratiosRes] = await Promise.all([
        fetch(`${API_BASE}/api/tickers/${symbol}/details`),
        fetch(`${API_BASE}/api/tickers/${symbol}/intraday`),
        fetch(`${API_BASE}/api/tickers/${symbol}/ratios`),
      ]);

      if (!detailsRes.ok) throw new Error("Failed to fetch ticker details");

      const detailsData = await detailsRes.json();
      setSelectedTicker(detailsData);

      if (intradayRes.ok) {
        const intradayData = await intradayRes.json();
        setIntradayData(Array.isArray(intradayData) ? intradayData : []);
      }

      if (ratiosRes.ok) {
        const ratiosData = await ratiosRes.json();
        setRatios(Array.isArray(ratiosData) ? ratiosData : []);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
      setSelectedTicker(null);
    } finally {
      setLoading(false);
    }
  }, []);

  const clearSelection = useCallback(() => {
    setSelectedTicker(null);
    setIntradayData([]);
    setRatios([]);
  }, []);

  return {
    searchResults,
    selectedTicker,
    intradayData,
    ratios,
    loading,
    searchLoading,
    error,
    searchTickers,
    lookupTicker,
    clearSelection,
    setSearchResults,
  };
}