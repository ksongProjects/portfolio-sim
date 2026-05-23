"use client";

import { useCallback, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiFetch, fetchJson, getErrorMessage } from "@/lib/api";

export interface NewsArticle {
  ID: string;
  Title: string;
  Source: string;
  SourceType: string;
  URL: string;
  Summary: string;
  Content: string;
  Sentiment: string;
  SentimentValue: string;
  PublishedAt: string;
  TickerSymbols: string[];
  Channel: string;
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

export interface YouTubeVideoSummaryRequest {
  video_id: string;
  title?: string;
}

export interface YouTubeVideoSummaryResult {
  video_id: string;
  title?: string;
  summary?: string;
  status: "ok" | "error";
  error?: string;
  transcript_source?: "captions" | "description";
}

async function fetchNewsData() {
  const data = await fetchJson<NewsArticle[]>("/api/news", undefined, "Failed to fetch news");
  return Array.isArray(data) ? data : [];
}

async function fetchChannelsData() {
  const data = await fetchJson<YouTubeChannel[]>("/api/channels", undefined, "Failed to fetch channels");
  return Array.isArray(data) ? data : [];
}

async function fetchLatestVideosData(channelId: string) {
  const data = await fetchJson<YouTubeVideo[]>(
    `/api/videos/latest?channel_id=${encodeURIComponent(channelId)}&limit=10`,
    undefined,
    "Failed to fetch videos"
  );
  return Array.isArray(data) ? data : [];
}

export function useNews() {
  const queryClient = useQueryClient();
  const [latestChannelId, setLatestChannelId] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  const articlesQuery = useQuery({
    queryKey: ["news", "articles"],
    queryFn: fetchNewsData,
  });

  const channelsQuery = useQuery({
    queryKey: ["news", "channels"],
    queryFn: fetchChannelsData,
  });

  const latestVideosQuery = useQuery({
    queryKey: ["news", "latest-videos", latestChannelId],
    queryFn: () => fetchLatestVideosData(latestChannelId!),
    enabled: Boolean(latestChannelId),
    placeholderData: (previousData) => previousData,
    retry: false,
  });

  const summarizeVideosMutation = useMutation({
    mutationFn: async (videos: YouTubeVideoSummaryRequest[]) => {
      const data = await fetchJson<{ results: YouTubeVideoSummaryResult[] }>("/api/videos/summarize", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ videos }),
      }, "Failed to summarize videos");
      return Array.isArray(data.results) ? data.results : [];
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["news", "articles"] });
    },
  });

  const addChannelMutation = useMutation({
    mutationFn: async ({
      channelId,
      name,
    }: {
      channelId: string;
      name: string;
    }) => {
      const res = await apiFetch("/api/channels", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ channel_id: channelId, name }),
      });

      if (!res.ok) {
        throw new Error("Failed to add channel");
      }
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["news", "channels"] });
    },
  });

  const deleteChannelMutation = useMutation({
    mutationFn: async (channelId: string) => {
      const res = await apiFetch(`/api/channels?id=${encodeURIComponent(channelId)}`, {
        method: "DELETE",
      });
      if (!res.ok) {
        throw new Error("Failed to delete channel");
      }
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["news", "channels"] });
    },
  });

  const fetchNews = useCallback(async () => {
    await articlesQuery.refetch();
  }, [articlesQuery]);

  const fetchChannels = useCallback(async () => {
    await channelsQuery.refetch();
  }, [channelsQuery]);

  const fetchLatestVideos = useCallback(
    async (channelId: string) => {
      setActionError(null);
      setLatestChannelId(channelId || null);
    },
    []
  );

  const searchChannels = useCallback(
    async (query: string) => {
      if (!query.trim()) {
        return [];
      }

      setActionError(null);

      try {
        const data = await queryClient.fetchQuery({
          queryKey: ["news", "channel-search", query],
          queryFn: () =>
            fetchJson<{ id: string; name: string; handle: string }[]>(
              `/api/channels/search?q=${encodeURIComponent(query)}`,
              undefined,
              "Search failed"
            ),
          staleTime: 5 * 60 * 1000,
          retry: false,
        });

        return Array.isArray(data) ? data : [];
      } catch (err) {
        setActionError(getErrorMessage(err));
        return [];
      }
    },
    [queryClient]
  );

  const summarizeVideos = useCallback(
    async (videos: YouTubeVideoSummaryRequest[]) => {
      setActionError(null);
      return summarizeVideosMutation.mutateAsync(videos);
    },
    [summarizeVideosMutation]
  );

  const addChannel = useCallback(
    async (channelId: string, name: string) => {
      setActionError(null);

      try {
        await addChannelMutation.mutateAsync({ channelId, name });
        return true;
      } catch (err) {
        setActionError(getErrorMessage(err));
        return false;
      }
    },
    [addChannelMutation]
  );

  const combinedError =
    actionError ??
    (articlesQuery.error
      ? getErrorMessage(articlesQuery.error)
      : channelsQuery.error
        ? getErrorMessage(channelsQuery.error)
        : latestVideosQuery.error
          ? getErrorMessage(latestVideosQuery.error)
          : summarizeVideosMutation.error
            ? getErrorMessage(summarizeVideosMutation.error)
            : addChannelMutation.error
              ? getErrorMessage(addChannelMutation.error)
              : null);

  return {
    articles: articlesQuery.data ?? [],
    channels: channelsQuery.data ?? [],
    latestVideos: latestVideosQuery.data ?? [],
    loading:
      articlesQuery.isFetching ||
      latestVideosQuery.isFetching ||
      summarizeVideosMutation.isPending ||
      addChannelMutation.isPending,
    error: combinedError,
    fetchNews,
    fetchChannels,
    fetchLatestVideos,
    summarizeVideos,
    searchChannels,
    addChannel,
    deleteChannel: deleteChannelMutation.mutateAsync,
  };
}
