"use client";

import { useCallback, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { fetchJson, getErrorMessage } from "@/lib/api";

export type ServiceStatus = "healthy" | "warning" | "error";

export interface Service {
  name: string;
  status: ServiceStatus;
  uptime: string;
  lastCheck: string;
}

export type LogEntry = {
  id: string;
  timestamp: string;
  level: "DEBUG" | "INFO" | "WARN" | "ERROR" | "FATAL";
  service: string;
  component: string | null;
  message: string;
  metadata: Record<string, unknown> | null;
  trace_id: string | null;
  span_id: string | null;
  route: string;
};

export interface LogFilters {
  level?: string;
  service?: string;
  routes?: string[];
}

function buildLogsPath(limit: number, logFilters: LogFilters) {
  const params = new URLSearchParams();
  params.set("limit", limit.toString());
  if (logFilters.level) params.set("level", logFilters.level);
  if (logFilters.service) params.set("service", logFilters.service);
  if (logFilters.routes && logFilters.routes.length > 0) {
    logFilters.routes.forEach((route) => params.append("route", route));
  }
  return `/api/observability/logs?${params.toString()}`;
}

function sameFilters(a: LogFilters, b: LogFilters) {
  const aRoutes = a.routes ?? [];
  const bRoutes = b.routes ?? [];

  return (
    a.level === b.level &&
    a.service === b.service &&
    aRoutes.length === bRoutes.length &&
    aRoutes.every((route, index) => route === bRoutes[index])
  );
}

async function fetchServicesData() {
  return fetchJson<Service[]>(
    "/api/observability/services",
    undefined,
    "Failed to fetch services"
  );
}

async function fetchLogsData(limit: number, filters: LogFilters) {
  return fetchJson<LogEntry[]>(buildLogsPath(limit, filters), undefined, "Failed to fetch logs");
}

export function useObservability(options?: { autoRefresh?: boolean; interval?: number }) {
  const [filters, setFilters] = useState<LogFilters>({});
  const [appliedFilters, setAppliedFilters] = useState<LogFilters>({});
  const [logsLimit, setLogsLimit] = useState(100);

  const refetchInterval = options?.autoRefresh ? (options.interval ?? 5000) : false;

  const servicesQuery = useQuery({
    queryKey: ["observability", "services"],
    queryFn: fetchServicesData,
    refetchInterval,
  });

  const logsQuery = useQuery({
    queryKey: ["observability", "logs", logsLimit, appliedFilters],
    queryFn: () => fetchLogsData(logsLimit, appliedFilters),
    refetchInterval,
    placeholderData: (previousData) => previousData,
  });

  const fetchServices = useCallback(async () => {
    await servicesQuery.refetch();
  }, [servicesQuery]);

  const fetchLogs = useCallback(
    async (limit = 100, logFilters?: LogFilters) => {
      const nextFilters = logFilters ?? appliedFilters;
      const filtersChanged = !sameFilters(nextFilters, appliedFilters);
      const limitChanged = limit !== logsLimit;

      if (filtersChanged) {
        setAppliedFilters(nextFilters);
      }

      if (limitChanged) {
        setLogsLimit(limit);
      }

      if (!filtersChanged && !limitChanged) {
        await logsQuery.refetch();
      }
    },
    [appliedFilters, logsLimit, logsQuery]
  );

  const setLogFilters = useCallback((newFilters: LogFilters) => {
    setFilters(newFilters);
  }, []);

  const applyFilters = useCallback(
    async (limit = logsLimit) => {
      const filtersChanged = !sameFilters(filters, appliedFilters);
      const limitChanged = limit !== logsLimit;

      if (filtersChanged) {
        setAppliedFilters(filters);
      }

      if (limitChanged) {
        setLogsLimit(limit);
      }

      if (!filtersChanged && !limitChanged) {
        await logsQuery.refetch();
      }
    },
    [appliedFilters, filters, logsLimit, logsQuery]
  );

  const refresh = useCallback(async () => {
    await Promise.all([servicesQuery.refetch(), logsQuery.refetch()]);
  }, [logsQuery, servicesQuery]);

  const latestUpdate = Math.max(servicesQuery.dataUpdatedAt, logsQuery.dataUpdatedAt);
  const combinedError = servicesQuery.error ?? logsQuery.error;

  return {
    services: servicesQuery.data ?? [],
    logs: logsQuery.data ?? [],
    loading: servicesQuery.isFetching || logsQuery.isFetching,
    error: combinedError ? getErrorMessage(combinedError) : null,
    lastUpdated: latestUpdate > 0 ? new Date(latestUpdate) : null,
    refresh,
    fetchServices,
    fetchLogs,
    filters,
    setLogFilters,
    applyFilters,
  };
}
