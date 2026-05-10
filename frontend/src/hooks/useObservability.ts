"use client";

import { useState, useCallback, useEffect, useRef } from "react";
import { useFrontendLogging, logPageView, logUserAction } from "./useFrontendLogging";

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
};

export interface LogFilters {
  level?: string;
  service?: string;
  startDate?: string;
  endDate?: string;
}

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export function useObservability(options?: { autoRefresh?: boolean; interval?: number }) {
  const [services, setServices] = useState<Service[]>([]);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const [filters, setFilters] = useState<LogFilters>({});
  const intervalRef = useRef<NodeJS.Timeout | null>(null);
  const { logAPICall } = useFrontendLogging();

  const fetchServices = useCallback(async () => {
    const startTime = Date.now();
    try {
      const res = await fetch(`${API_BASE}/api/observability/services`);
      const duration = Date.now() - startTime;
      if (!res.ok) {
        logAPICall("GET", "/api/observability/services", res.status, duration, "Failed to fetch services");
        throw new Error("Failed to fetch services");
      }
      logAPICall("GET", "/api/observability/services", res.status, duration);
      const data = await res.json();
      setServices(data);
      setLastUpdated(new Date());
    } catch (err) {
      const duration = Date.now() - startTime;
      logAPICall("GET", "/api/observability/services", 0, duration, err instanceof Error ? err.message : "Unknown error");
      setError(err instanceof Error ? err.message : "Unknown error");
    }
  }, [logAPICall]);

  const buildLogsUrl = useCallback((limit: number, logFilters: LogFilters) => {
    const params = new URLSearchParams();
    params.set("limit", limit.toString());
    if (logFilters.level) params.set("level", logFilters.level);
    if (logFilters.service) params.set("service", logFilters.service);
    if (logFilters.startDate) params.set("start", logFilters.startDate);
    if (logFilters.endDate) params.set("end", logFilters.endDate);
    return `${API_BASE}/api/observability/logs?${params.toString()}`;
  }, []);

  const fetchLogs = useCallback(async (limit = 100, logFilters?: LogFilters) => {
    const activeFilters = logFilters ?? filters;
    const startTime = Date.now();
    try {
      const url = buildLogsUrl(limit, activeFilters);
      const res = await fetch(url);
      const duration = Date.now() - startTime;
      if (!res.ok) {
        logAPICall("GET", `/api/observability/logs`, res.status, duration, "Failed to fetch logs");
        throw new Error("Failed to fetch logs");
      }
      logAPICall("GET", `/api/observability/logs`, res.status, duration);
      const data = await res.json();
      setLogs(data);
    } catch (err) {
      const duration = Date.now() - startTime;
      logAPICall("GET", `/api/observability/logs`, 0, duration, err instanceof Error ? err.message : "Unknown error");
      setError(err instanceof Error ? err.message : "Unknown error");
    }
  }, [filters, buildLogsUrl, logAPICall]);

  const setLogFilters = useCallback((newFilters: LogFilters) => {
    setFilters(newFilters);
    fetchLogs(100, newFilters);
  }, [fetchLogs]);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    await Promise.all([fetchServices(), fetchLogs(100, filters)]);
    setLoading(false);
  }, [fetchServices, fetchLogs, filters]);

  const startAutoRefresh = useCallback(() => {
    if (intervalRef.current) return;
    const interval = options?.interval ?? 5000;
    intervalRef.current = setInterval(() => {
      fetchServices();
      fetchLogs(100, filters);
    }, interval);
  }, [options?.interval, fetchServices, fetchLogs, filters]);

  const stopAutoRefresh = useCallback(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
  }, []);

  useEffect(() => {
    const init = () => {
      logPageView("observability");
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
    startAutoRefresh,
    stopAutoRefresh,
    fetchServices,
    fetchLogs,
    filters,
    setLogFilters,
  };
}