"use client";

import { useState, useCallback } from "react";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface NewsArticle {
  ID: string;
  Title: string;
  Source: string;
  URL: string;
  Summary: string;
  Sentiment: string;
  PublishedAt: string;
  TickerSymbols: string[];
}

export function useNews() {
  const [articles, setArticles] = useState<NewsArticle[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchNews = useCallback(async (limit = 20) => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${API_BASE}/api/news?limit=${limit}`);
      if (!res.ok) throw new Error("Failed to fetch news");
      const data = await res.json();
      setArticles(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
    } finally {
      setLoading(false);
    }
  }, []);

  return { articles, loading, error, fetchNews };
}