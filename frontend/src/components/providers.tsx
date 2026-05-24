"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useState } from "react";
import { ApiError } from "@/lib/api";
import { PriceProvider } from "./price-context";

function shouldRetryQuery(failureCount: number, error: unknown) {
  if (failureCount >= 2) {
    return false;
  }

  if (error instanceof ApiError) {
    if (!error.retryable) {
      return false;
    }
  }

  const message = error instanceof Error ? error.message : String(error);
  if (message.includes("ECONNREFUSED") || message.includes("connection refused")) {
    return false;
  }

  return true;
}

export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(() => new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 30 * 1000,
        gcTime: 5 * 60 * 1000,
        refetchOnWindowFocus: false,
        retry: shouldRetryQuery,
      },
    },
  }));

  return (
    <QueryClientProvider client={queryClient}>
      <PriceProvider>
        {children}
      </PriceProvider>
    </QueryClientProvider>
  );
}
