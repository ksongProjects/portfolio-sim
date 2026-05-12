"use client";

import { useState, useCallback, useEffect, useRef } from "react";

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

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export function useObservability(options?: { autoRefresh?: boolean; interval?: number }) {
  const [services, setServices] = useState<Service[]>([]);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const [filters, setFilters] = useState<LogFilters>({});
  const [appliedFilters, setAppliedFilters] = useState<LogFilters>({});
  const intervalRef = useRef<NodeJS.Timeout | null>(null);

  const fetchServices = useCallback(async () => {
    try {
      const res = await fetch(`${API_BASE}/api/observability/services`);
      if (!res.ok) {
        throw new Error("Failed to fetch services");
      }
      const data = await res.json();
      setServices(data);
      setLastUpdated(new Date());
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
    }
  }, []);

const buildLogsUrl = useCallback((limit: number, logFilters: LogFilters) => {
    const params = new URLSearchParams();
    params.set("limit", limit.toString());
    if (logFilters.level) params.set("level", logFilters.level);
    if (logFilters.service) params.set("service", logFilters.service);
    if (logFilters.routes && logFilters.routes.length > 0) {
      logFilters.routes.forEach(r => params.append("route", r));
    }
    return `${API_BASE}/api/observability/logs?${params.toString()}`;
  }, []);

  const fetchLogs = useCallback(async (limit = 100, logFilters?: LogFilters) => {
    const activeFilters = logFilters ?? appliedFilters;
    try {
      const url = buildLogsUrl(limit, activeFilters);
      const res = await fetch(url);
      if (!res.ok) {
        throw new Error("Failed to fetch logs");
      }
      const data = await res.json();
      setLogs(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
    }
  }, [appliedFilters, buildLogsUrl]);

  const setLogFilters = useCallback((newFilters: LogFilters) => {
    setFilters(newFilters);
  }, []);

  const applyFilters = useCallback(() => {
    setAppliedFilters(filters);
    fetchLogs(100, filters);
  }, [fetchLogs, filters]);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    await Promise.all([fetchServices(), fetchLogs(100, appliedFilters)]);
    setLoading(false);
  }, [fetchServices, fetchLogs, appliedFilters]);

  const startAutoRefresh = useCallback(() => {
    if (intervalRef.current) return;
    const interval = options?.interval ?? 5000;
    intervalRef.current = setInterval(() => {
      fetchServices();
      fetchLogs(100, appliedFilters);
    }, interval);
  }, [options?.interval, fetchServices, fetchLogs, appliedFilters]);

  const stopAutoRefresh = useCallback(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
  }, []);

  useEffect(() => {
    const init = () => {
      refresh();
    };
    init();
    if (options?.autoRefresh) {
      startAutoRefresh();
    }
    return () => stopAutoRefresh();
  }, [refresh, options?.autoRefresh, startAutoRefresh, stopAutoRefresh]);

  return {
    services,
    logs,
    loading,
    error,
    lastUpdated,
    refresh,
    fetchServices,
    fetchLogs,
    filters,
    setLogFilters,
    applyFilters,
  };
}
