"use client";

import { useCallback, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { fetchJson, getErrorMessage } from "@/lib/api";

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

function dedupeTickers(results: TickerDetails[]) {
  const seen = new Set<string>();
  return results.filter((ticker) => (seen.has(ticker.symbol) ? false : seen.add(ticker.symbol)));
}

async function fetchTickerDetails(symbol: string) {
  return fetchJson<TickerDetails>(
    `/api/tickers/${encodeURIComponent(symbol)}/details`,
    undefined,
    "Failed to fetch ticker details"
  );
}

async function fetchIntradayData(symbol: string, interval: string) {
  try {
    const data = await fetchJson<IntradayBar[]>(
      `/api/tickers/${encodeURIComponent(symbol)}/intraday${
        interval !== "1min" ? `?interval=${encodeURIComponent(interval)}` : ""
      }`
    );
    return Array.isArray(data) ? data : [];
  } catch {
    return [];
  }
}

async function fetchRatiosData(symbol: string) {
  try {
    const data = await fetchJson<FinancialRatio[]>(
      `/api/tickers/${encodeURIComponent(symbol)}/ratios`
    );
    return Array.isArray(data) ? data : [];
  } catch {
    return [];
  }
}

export function useTickerLookup(initialSymbol?: string) {
  const queryClient = useQueryClient();
  const [searchResults, setSearchResultsState] = useState<TickerDetails[]>([]);
  const [manualSelectedSymbol, setManualSelectedSymbol] = useState<string | null>(null);
  const [selectedInterval, setSelectedInterval] = useState("1min");
  const [searchLoading, setSearchLoading] = useState(false);
  const selectedSymbol = initialSymbol ?? manualSelectedSymbol;
  const activeInterval = initialSymbol ? "1min" : selectedInterval;

  const detailsQuery = useQuery({
    queryKey: ["tickers", "details", selectedSymbol],
    queryFn: () => fetchTickerDetails(selectedSymbol!),
    enabled: Boolean(selectedSymbol),
  });

  const intradayQuery = useQuery({
    queryKey: ["tickers", "intraday", selectedSymbol, activeInterval],
    queryFn: () => fetchIntradayData(selectedSymbol!, activeInterval),
    enabled: Boolean(selectedSymbol),
  });

  const ratiosQuery = useQuery({
    queryKey: ["tickers", "ratios", selectedSymbol],
    queryFn: () => fetchRatiosData(selectedSymbol!),
    enabled: Boolean(selectedSymbol),
  });

  const searchTickers = useCallback(
    async (query: string) => {
      if (query.length < 1) {
        setSearchResultsState([]);
        return;
      }

      setSearchLoading(true);

      try {
        const results = await queryClient.fetchQuery({
          queryKey: ["tickers", "search", query],
          queryFn: async () => {
            const data = await fetchJson<TickerDetails[]>(
              `/api/tickers/search?q=${encodeURIComponent(query)}`,
              undefined,
              "Search failed"
            );

            return dedupeTickers(Array.isArray(data) ? data : []);
          },
          staleTime: 5 * 60 * 1000,
        });

        setSearchResultsState(results);
      } catch {
        setSearchResultsState([]);
      } finally {
        setSearchLoading(false);
      }
    },
    [queryClient]
  );

  const lookupTicker = useCallback(
    async (symbol: string, interval = "1min") => {
      setManualSelectedSymbol(symbol);
      setSelectedInterval(interval);

      try {
        if (symbol === selectedSymbol && interval === activeInterval) {
          await Promise.all([
            detailsQuery.refetch(),
            intradayQuery.refetch(),
            ratiosQuery.refetch(),
          ]);
          return;
        }

        await Promise.all([
          queryClient.fetchQuery({
            queryKey: ["tickers", "details", symbol],
            queryFn: () => fetchTickerDetails(symbol),
          }),
          queryClient.fetchQuery({
            queryKey: ["tickers", "intraday", symbol, interval],
            queryFn: () => fetchIntradayData(symbol, interval),
          }),
          queryClient.fetchQuery({
            queryKey: ["tickers", "ratios", symbol],
            queryFn: () => fetchRatiosData(symbol),
          }),
        ]);
      } catch {
        // detailsQuery exposes failure state for callers that render errors
      }
    },
    [activeInterval, detailsQuery, intradayQuery, queryClient, ratiosQuery, selectedSymbol]
  );

  const clearSelection = useCallback(() => {
    setManualSelectedSymbol(null);
    setSelectedInterval("1min");
  }, []);

  const setSearchResults = useCallback(
    (results: TickerDetails[]) => {
      setSearchResultsState(results);
    },
    []
  );

  return {
    searchResults,
    selectedTicker: selectedSymbol ? detailsQuery.data ?? null : null,
    intradayData: selectedSymbol ? intradayQuery.data ?? [] : [],
    ratios: selectedSymbol ? ratiosQuery.data ?? [] : [],
    loading:
      detailsQuery.isFetching ||
      intradayQuery.isFetching ||
      ratiosQuery.isFetching,
    searchLoading,
    error: detailsQuery.error ? getErrorMessage(detailsQuery.error) : null,
    searchTickers,
    lookupTicker,
    clearSelection,
    setSearchResults,
  };
}
