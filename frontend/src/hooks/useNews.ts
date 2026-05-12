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

export interface YouTubeChannel {
  id: string;
  channel_id: string;
  name: string;
  handle: string;
}

export interface YouTubeVideo {
  id: string;
  title: string;
  description: string;
  channel_id: string;
  channel_name: string;
  published_at: string;
  thumb_url: string;
}

export interface NewsVideo {
  id: string;
  youtube_id: string;
  title: string;
  channel: string;
  summary: string;
  sentiment: string;
  tickers: string;
  published_at: string;
}

export function useNews() {
  const [articles, setArticles] = useState<NewsArticle[]>([]);
  const [videos, setVideos] = useState<NewsVideo[]>([]);
  const [channels, setChannels] = useState<YouTubeChannel[]>([]);
  const [latestVideos, setLatestVideos] = useState<YouTubeVideo[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchNews = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${API_BASE}/api/news`);
      if (!res.ok) throw new Error("Failed to fetch news");
      const data = await res.json();
      setArticles(Array.isArray(data) ? data : []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchChannels = useCallback(async () => {
    const res = await fetch(`${API_BASE}/api/channels`);
    if (!res.ok) return;
    const data = await res.json();
    setChannels(Array.isArray(data) ? data : []);
  }, []);

  const fetchLatestVideos = useCallback(async (channelId: string) => {
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/api/videos/latest?channel_id=${encodeURIComponent(channelId)}`);
      if (!res.ok) throw new Error("Failed to fetch videos");
      const data = await res.json();
      setLatestVideos(Array.isArray(data) ? data : []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchStoredVideos = useCallback(async () => {
    const res = await fetch(`${API_BASE}/api/videos`);
    if (!res.ok) return;
    const data = await res.json();
    setVideos(Array.isArray(data) ? data : []);
  }, []);

  const searchChannels = useCallback(async (query: string) => {
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/api/channels/search?q=${encodeURIComponent(query)}`);
      if (!res.ok) throw new Error("Search failed");
      return await res.json();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
      return [];
    } finally {
      setLoading(false);
    }
  }, []);

  const analyzeVideo = useCallback(async (videoId: string, title: string) => {
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/api/videos/analyze`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ video_id: videoId, title }),
      });
      if (!res.ok) throw new Error("Failed to analyze video");
      await fetchStoredVideos();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
    } finally {
      setLoading(false);
    }
  }, [fetchStoredVideos]);

  return { articles, videos, channels, latestVideos, loading, error, fetchNews, fetchChannels, fetchLatestVideos, fetchStoredVideos, analyzeVideo, searchChannels };
}