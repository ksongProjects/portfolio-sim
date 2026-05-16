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

interface TickerLookupResponse extends TickerDetails {
  intraday?: IntradayBar[];
  ratios?: FinancialRatio[];
}

function dedupeTickers(results: TickerDetails[]) {
  const seen = new Set<string>();
  return results.filter((ticker) => (seen.has(ticker.symbol) ? false : seen.add(ticker.symbol)));
}

async function fetchTickerDetails(symbol: string) {
  return fetchJson<TickerLookupResponse>(
    `/api/tickers/${encodeURIComponent(symbol)}/details`,
    undefined,
    "Failed to fetch ticker details"
  );
}

async function fetchIntradayBars(symbol: string, range: string) {
  const response = await fetchJson<TickerLookupResponse>(
    `/api/tickers/${encodeURIComponent(symbol)}/details?range=${range}`,
    undefined,
    "Failed to fetch intraday data"
  );
  return response.intraday ?? [];
}

export type ChartRange = "1d" | "1w" | "1m";

export function useTickerLookup(initialSymbol?: string) {
  const queryClient = useQueryClient();
  const [searchResults, setSearchResultsState] = useState<TickerDetails[]>([]);
  const [manualSelectedSymbol, setManualSelectedSymbol] = useState<string | null>(null);
  const [searchLoading, setSearchLoading] = useState(false);
  const [chartRange, setChartRange] = useState<ChartRange>("1d");
  const selectedSymbol = initialSymbol ?? manualSelectedSymbol;

  const detailsQuery = useQuery({
    queryKey: ["tickers", "details", selectedSymbol],
    queryFn: () => fetchTickerDetails(selectedSymbol!),
    enabled: Boolean(selectedSymbol),
  });

  const intradayQuery = useQuery({
    queryKey: ["tickers", "intraday", selectedSymbol, chartRange],
    queryFn: () => fetchIntradayBars(selectedSymbol!, chartRange),
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

      await queryClient.invalidateQueries({ queryKey: ["tickers", "details", symbol] });
      await queryClient.invalidateQueries({ queryKey: ["tickers", "intraday", symbol] });
    },
    [queryClient]
  );

  const clearSelection = useCallback(() => {
    setManualSelectedSymbol(null);
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
    ratios: selectedSymbol ? detailsQuery.data?.ratios ?? [] : [],
    loading: detailsQuery.isFetching,
    searchLoading,
    error: detailsQuery.error ? getErrorMessage(detailsQuery.error) : null,
    searchTickers,
    lookupTicker,
    clearSelection,
    setSearchResults,
    chartRange,
    setChartRange,
  };
}
