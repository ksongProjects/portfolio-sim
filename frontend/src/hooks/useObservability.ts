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
};

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export function useObservability(options?: { autoRefresh?: boolean; interval?: number }) {
  const [services, setServices] = useState<Service[]>([]);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const intervalRef = useRef<NodeJS.Timeout | null>(null);

  const fetchServices = useCallback(async () => {
    try {
      const res = await fetch(`${API_BASE}/api/observability/services`);
      if (!res.ok) throw new Error("Failed to fetch services");
      const data = await res.json();
      setServices(data);
      setLastUpdated(new Date());
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
    }
  }, []);

  const fetchLogs = useCallback(async (limit = 100) => {
    try {
      const res = await fetch(`${API_BASE}/api/observability/logs?limit=${limit}`);
      if (!res.ok) throw new Error("Failed to fetch logs");
      const data = await res.json();
      setLogs(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
    }
  }, []);

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