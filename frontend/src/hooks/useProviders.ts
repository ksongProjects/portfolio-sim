"use client";

import { useState, useCallback } from "react";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface ProviderConfig {
  id: string;
  provider_id: string;
  name: string;
  description: string;
  api_key_set: boolean;
  is_connected: boolean;
  rate_limit: number;
  docs_url: string;
}

export interface ConnectionStatus {
  id: string;
  name: string;
  type: string;
  is_up: boolean;
  latency_ms: number;
}

export function useProviders() {
  const [providers, setProviders] = useState<ProviderConfig[]>([]);
  const [connections, setConnections] = useState<ConnectionStatus[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchProviders = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${API_BASE}/api/providers`);
      if (!res.ok) throw new Error("Failed to fetch providers");
      const data = await res.json();
      setProviders(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchConnections = useCallback(async () => {
    try {
      const res = await fetch(`${API_BASE}/api/connections`);
      if (!res.ok) throw new Error("Failed to fetch connections");
      const data = await res.json();
      setConnections(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
    }
  }, []);

  const saveProviderKey = useCallback(async (providerId: string, apiKey: string): Promise<boolean> => {
    try {
      const res = await fetch(`${API_BASE}/api/providers`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ provider_id: providerId, api_key: apiKey }),
      });
      if (!res.ok) throw new Error("Failed to save");
      await fetchProviders();
      return true;
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
      return false;
    }
  }, [fetchProviders]);

  const refresh = useCallback(async () => {
    setLoading(true);
    await Promise.all([fetchProviders(), fetchConnections()]);
    setLoading(false);
  }, [fetchProviders, fetchConnections]);

  return { providers, connections, loading, error, fetchProviders, fetchConnections, saveProviderKey, refresh };
}