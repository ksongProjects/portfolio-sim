"use client";

import { useCallback, useState, useEffect, useRef } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { fetchJson, getErrorMessage } from "@/lib/api";

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
  close: number;
  volume: number;
  open?: number;
  high?: number;
  low?: number;
}

export interface IntradayData {
  bars: IntradayBar[];
  change: number;
  changePct: number;
  dataDate?: string;
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

async function fetchTickerCompany(symbol: string) {
  return fetchJson<TickerDetails>(`/api/tickers/${encodeURIComponent(symbol)}/details`, undefined, "Failed to fetch company data");
}

async function fetchTickerBars(symbol: string, range: string) {
  return fetchJson<{
    symbol: string;
    bars: { timestamp: string; open: number; high: number; low: number; close: number; volume: number }[];
    change: number;
    changePct: number;
    dataDate?: string;
  }>(
    `/api/tickers/bars?symbol=${encodeURIComponent(symbol)}&range=${encodeURIComponent(range)}`,
    undefined,
    "Failed to fetch ticker bars"
  );
}

async function fetchIntradayBars(symbol: string, range: string) {
  const response = await fetchTickerBars(symbol, range);
  return {
    bars: response.bars.map(b => ({ timestamp: b.timestamp, close: b.close, volume: b.volume ?? 0 })),
    change: response.change,
    changePct: response.changePct,
    dataDate: response.dataDate,
  };
}

export type ChartRange = "1d" | "1w" | "1m";

export function useTickerLookup(initialSymbol?: string) {
  const queryClient = useQueryClient();
  const [searchResults, setSearchResultsState] = useState<TickerDetails[]>([]);
  const [manualSelectedSymbol, setManualSelectedSymbol] = useState<string | null>(null);
  const [searchLoading, setSearchLoading] = useState(false);
  const [chartRange, setChartRange] = useState<ChartRange>("1d");
  const [livePrice, setLivePrice] = useState<number | null>(null);
  const [liveChangePct, setLiveChangePct] = useState<number | null>(null);
  const selectedSymbol = initialSymbol ?? manualSelectedSymbol;

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const companyQuery = useQuery({
    queryKey: ["tickers", "company", selectedSymbol],
    queryFn: () => fetchTickerCompany(selectedSymbol!),
    enabled: Boolean(selectedSymbol),
  });

  const intradayQuery = useQuery({
    queryKey: ["tickers", "intraday", selectedSymbol, chartRange],
    queryFn: () => fetchIntradayBars(selectedSymbol!, chartRange),
    enabled: Boolean(selectedSymbol),
  });

  useEffect(() => {
    if (!selectedSymbol) {
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
      setLivePrice(null);
      setLiveChangePct(null);
      return;
    }

    const url = `ws://${API_BASE.replace("http://", "")}/api/ws/market?symbols=${encodeURIComponent(selectedSymbol)}`;
    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {};

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.symbol === selectedSymbol && data.close) {
          setLivePrice(data.close);
          if (data.changePct !== undefined) {
            setLiveChangePct(data.changePct);
          }
        }
      } catch {
      }
    };

    ws.onerror = () => {};

    ws.onclose = () => {
      wsRef.current = null;
    };

    return () => {
      ws.close();
      wsRef.current = null;
    };
  }, [selectedSymbol]);

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
      setLivePrice(null);
      setLiveChangePct(null);

      await queryClient.invalidateQueries({ queryKey: ["tickers", "company", symbol] });
      await queryClient.invalidateQueries({ queryKey: ["tickers", "intraday", symbol] });
    },
    [queryClient]
  );

  const clearSelection = useCallback(() => {
    setManualSelectedSymbol(null);
    setLivePrice(null);
    setLiveChangePct(null);
  }, []);

  const setSearchResults = useCallback(
    (results: TickerDetails[]) => {
      setSearchResultsState(results);
    },
    []
  );

  const effectivePrice = livePrice ?? companyQuery.data?.price ?? 0;
  const effectiveChangePct = liveChangePct ?? companyQuery.data?.changePct ?? 0;

  const selectedTickerWithLive: TickerDetails | null = selectedSymbol && companyQuery.data ? {
    ...companyQuery.data,
    price: effectivePrice,
    changePct: effectiveChangePct,
  } : null;

  return {
    searchResults,
    selectedTicker: selectedTickerWithLive,
    intradayData: selectedSymbol ? (intradayQuery.data?.bars ?? []) : [],
    intradayChange: selectedSymbol ? (intradayQuery.data?.change ?? 0) : 0,
    intradayChangePct: selectedSymbol ? (intradayQuery.data?.changePct ?? 0) : 0,
    intradayDataDate: selectedSymbol ? (intradayQuery.data?.dataDate ?? undefined) : undefined,
    loading: companyQuery.isFetching,
    chartLoading: intradayQuery.isFetching,
    chartError: intradayQuery.error ? getErrorMessage(intradayQuery.error) : null,
    searchLoading,
    error: companyQuery.error ? getErrorMessage(companyQuery.error) : null,
    searchTickers,
    lookupTicker,
    clearSelection,
    setSearchResults,
    chartRange,
    setChartRange,
    liveConnected: wsRef.current?.readyState === WebSocket.OPEN,
  };
}