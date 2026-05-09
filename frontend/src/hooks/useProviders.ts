"use client";

import { useState, useCallback } from "react";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface ProviderConfig {
  id: string;
  provider_id: string;
  name: string;
  description: string;
  type: string;
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

export interface RSSFeed {
  id: string;
  name: string;
  url: string;
  scrape_interval_min: number;
  last_scrape_at: string;
  is_active: boolean;
}

export function useProviders() {
  const [providers, setProviders] = useState<ProviderConfig[]>([]);
  const [connections, setConnections] = useState<ConnectionStatus[]>([]);
  const [rssFeeds, setRssFeeds] = useState<RSSFeed[]>([]);
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

  const fetchRSSFeeds = useCallback(async () => {
    try {
      const res = await fetch(`${API_BASE}/api/rss-feeds`);
      if (!res.ok) throw new Error("Failed to fetch RSS feeds");
      const data = await res.json();
      setRssFeeds(Array.isArray(data) ? data : []);
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

  const validateProviderKey = useCallback(async (providerId: string, apiKey: string): Promise<{ valid: boolean; error?: string }> => {
    try {
      const res = await fetch(`${API_BASE}/api/providers/validate`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ provider_id: providerId, api_key: apiKey }),
      });
      if (!res.ok) throw new Error("Failed to validate");
      return await res.json();
    } catch (err) {
      return { valid: false, error: err instanceof Error ? err.message : "Unknown error" };
    }
  }, []);

  const addRSSFeed = useCallback(async (name: string, url: string, scrapeInterval: number): Promise<boolean> => {
    try {
      const res = await fetch(`${API_BASE}/api/rss-feeds`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, url, scrape_interval_min: scrapeInterval }),
      });
      if (!res.ok) throw new Error("Failed to add feed");
      const data = await res.json();
      setRssFeeds(data);
      return true;
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
      return false;
    }
  }, []);

  const deleteRSSFeed = useCallback(async (feedId: string): Promise<boolean> => {
    try {
      const res = await fetch(`${API_BASE}/api/rss-feeds?id=${feedId}`, {
        method: "DELETE",
      });
      if (!res.ok) throw new Error("Failed to delete feed");
      await fetchRSSFeeds();
      return true;
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
      return false;
    }
  }, [fetchRSSFeeds]);

  const refresh = useCallback(async () => {
    setLoading(true);
    await Promise.all([fetchProviders(), fetchConnections(), fetchRSSFeeds()]);
    setLoading(false);
  }, [fetchProviders, fetchConnections, fetchRSSFeeds]);

  return {
    providers,
    connections,
    rssFeeds,
    loading,
    error,
    fetchProviders,
    fetchConnections,
    fetchRSSFeeds,
    saveProviderKey,
    validateProviderKey,
    addRSSFeed,
    deleteRSSFeed,
    refresh,
  };
}