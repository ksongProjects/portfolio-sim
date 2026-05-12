"use client";

import { useCallback, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { fetchJson, getErrorMessage } from "@/lib/api";

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
  Message: string;
  Service: string;
  Level: string;
  Timestamp: string;
}

async function fetchStrategiesData() {
  return fetchJson<Strategy[]>("/api/strategies", undefined, "Failed to fetch strategies");
}

async function fetchSignalsData(limit: number) {
  return fetchJson<Signal[]>(`/api/signals?limit=${limit}`, undefined, "Failed to fetch signals");
}

export function useStrategies() {
  const [signalLimit, setSignalLimit] = useState(10);

  const strategiesQuery = useQuery({
    queryKey: ["strategies"],
    queryFn: fetchStrategiesData,
  });

  const signalsQuery = useQuery({
    queryKey: ["signals", signalLimit],
    queryFn: () => fetchSignalsData(signalLimit),
  });

  const fetchStrategies = useCallback(async () => {
    await strategiesQuery.refetch();
  }, [strategiesQuery]);

  const fetchSignals = useCallback(
    async (limit = 10) => {
      if (limit !== signalLimit) {
        setSignalLimit(limit);
        return;
      }

      await signalsQuery.refetch();
    },
    [signalLimit, signalsQuery]
  );

  const refresh = useCallback(async () => {
    await Promise.all([strategiesQuery.refetch(), signalsQuery.refetch()]);
  }, [signalsQuery, strategiesQuery]);

  const combinedError = strategiesQuery.error ?? signalsQuery.error;

  return {
    strategies: strategiesQuery.data ?? [],
    signals: signalsQuery.data ?? [],
    loading: strategiesQuery.isFetching || signalsQuery.isFetching,
    error: combinedError ? getErrorMessage(combinedError) : null,
    fetchStrategies,
    fetchSignals,
    refresh,
  };
}
