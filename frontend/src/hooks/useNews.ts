"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";

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

async function fetchJSON(url: string) {
  const res = await fetch(url);
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json();
}

async function postJSON(url: string, body: unknown) {
  const res = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.status === 204 ? null : res.json();
}

export function useNews() {
  const queryClient = useQueryClient();

  const articles = useQuery<NewsArticle[]>({
    queryKey: ["news", "articles"],
    queryFn: () => fetchJSON(`${API_BASE}/api/news`),
  });

  const channels = useQuery<YouTubeChannel[]>({
    queryKey: ["news", "channels"],
    queryFn: () => fetchJSON(`${API_BASE}/api/channels`),
  });

  const videos = useQuery<NewsVideo[]>({
    queryKey: ["news", "videos"],
    queryFn: () => fetchJSON(`${API_BASE}/api/videos`),
  });

  const latestVideos = useQuery<YouTubeVideo[]>({
    queryKey: ["news", "latestVideos"],
    queryFn: () => fetchJSON(`${API_BASE}/api/videos/latest`),
    enabled: false,
  });

  const searchChannelsMutation = useMutation({
    mutationFn: (query: string) => fetchJSON(`${API_BASE}/api/channels/search?q=${encodeURIComponent(query)}`) as Promise<YouTubeChannel[]>,
  });

  const addChannelMutation = useMutation({
    mutationFn: (data: { channel_id: string; name: string }) => postJSON(`${API_BASE}/api/channels`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["news", "channels"] });
    },
  });

  const analyzeVideoMutation = useMutation({
    mutationFn: (data: { video_id: string; title: string }) => postJSON(`${API_BASE}/api/videos/analyze`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["news", "videos"] });
    },
  });

  return {
    articles: articles.data ?? [],
    videos: videos.data ?? [],
    channels: channels.data ?? [],
    latestVideos: latestVideos.data ?? [],
    loading: articles.isLoading || videos.isLoading || channels.isLoading,
    error: articles.error?.message ?? null,
    fetchNews: () => queryClient.invalidateQueries({ queryKey: ["news", "articles"] }),
    fetchChannels: () => queryClient.invalidateQueries({ queryKey: ["news", "channels"] }),
    fetchLatestVideos: (channelId: string) =>
      queryClient.fetchQuery({ queryKey: ["news", "latestVideos", channelId], queryFn: () => fetchJSON(`${API_BASE}/api/videos/latest?channel_id=${encodeURIComponent(channelId)}`) }),
    fetchStoredVideos: () => queryClient.invalidateQueries({ queryKey: ["news", "videos"] }),
    searchChannels: searchChannelsMutation.mutateAsync,
    addChannel: addChannelMutation.mutateAsync,
    analyzeVideo: analyzeVideoMutation.mutateAsync,
  };
}