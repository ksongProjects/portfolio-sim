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

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export function useObservability(options?: { autoRefresh?: boolean; interval?: number }) {
  const [services, setServices] = useState<Service[]>([]);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
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

  const fetchLogs = useCallback(async (limit = 100) => {
    const startTime = Date.now();
    try {
      const res = await fetch(`${API_BASE}/api/observability/logs?limit=${limit}`);
      const duration = Date.now() - startTime;
      if (!res.ok) {
        logAPICall("GET", `/api/observability/logs?limit=${limit}`, res.status, duration, "Failed to fetch logs");
        throw new Error("Failed to fetch logs");
      }
      logAPICall("GET", `/api/observability/logs?limit=${limit}`, res.status, duration);
      const data = await res.json();
      setLogs(data);
    } catch (err) {
      const duration = Date.now() - startTime;
      logAPICall("GET", `/api/observability/logs?limit=${limit}`, 0, duration, err instanceof Error ? err.message : "Unknown error");
      setError(err instanceof Error ? err.message : "Unknown error");
    }
  }, [logAPICall]);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    await Promise.all([fetchServices(), fetchLogs()]);
    setLoading(false);
  }, [fetchServices, fetchLogs]);

  const startAutoRefresh = useCallback(() => {
    if (intervalRef.current) return;
    const interval = options?.interval ?? 5000;
    intervalRef.current = setInterval(() => {
      fetchServices();
      fetchLogs();
    }, interval);
  }, [options?.interval, fetchServices, fetchLogs]);

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
  };
}