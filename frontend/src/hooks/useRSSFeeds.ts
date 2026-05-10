"use client";

import { useState, useCallback } from "react";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface RSSFeed {
  id: string;
  name: string;
  url: string;
  is_active: boolean;
  last_scrape_at: string | null;
}

export function useRSSFeeds() {
  const [feeds, setFeeds] = useState<RSSFeed[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchFeeds = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${API_BASE}/api/rss-feeds`);
      if (!res.ok) throw new Error("Failed to fetch feeds");
      const data = await res.json();
      setFeeds(Array.isArray(data) ? data : []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
    } finally {
      setLoading(false);
    }
  }, []);

  const addFeed = useCallback(async (name: string, url: string) => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${API_BASE}/api/rss-feeds`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, url }),
      });
      if (!res.ok) throw new Error("Failed to add feed");
      const data = await res.json();
      setFeeds(Array.isArray(data) ? data : []);
      return true;
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
      return false;
    } finally {
      setLoading(false);
    }
  }, []);

  const deleteFeed = useCallback(async (id: string) => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${API_BASE}/api/rss-feeds?id=${id}`, { method: "DELETE" });
      if (!res.ok) throw new Error("Failed to delete feed");
      setFeeds((prev) => prev.filter((f) => f.id !== id));
      return true;
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
      return false;
    } finally {
      setLoading(false);
    }
  }, []);

  const scrapeFeeds = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${API_BASE}/api/rss-feeds/scrape`, { method: "POST" });
      if (!res.ok) throw new Error("Failed to trigger scrape");
      return true;
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
      return false;
    } finally {
      setLoading(false);
    }
  }, []);

  return { feeds, loading, error, fetchFeeds, addFeed, deleteFeed, scrapeFeeds };
}