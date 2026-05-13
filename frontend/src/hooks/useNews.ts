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
    `/api/videos/latest?channel_id=${encodeURIComponent(channelId)}`,
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
  });

  const analyzeVideoMutation = useMutation({
    mutationFn: async ({ videoId, title }: { videoId: string; title: string }) => {
      const res = await apiFetch("/api/videos/analyze", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ video_id: videoId, title }),
      });

      if (!res.ok) {
        throw new Error("Failed to analyze video");
      }
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

  const fetchNews = useCallback(async () => {
    await articlesQuery.refetch();
  }, [articlesQuery]);

  const fetchChannels = useCallback(async () => {
    await channelsQuery.refetch();
  }, [channelsQuery]);

  const fetchLatestVideos = useCallback(
    async (channelId: string) => {
      setActionError(null);

      if (!channelId) {
        setLatestChannelId(null);
        return;
      }

      if (channelId !== latestChannelId) {
        setLatestChannelId(channelId);
        return;
      }

      await latestVideosQuery.refetch();
    },
    [latestChannelId, latestVideosQuery]
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
        });

        return Array.isArray(data) ? data : [];
      } catch (err) {
        setActionError(getErrorMessage(err));
        return [];
      }
    },
    [queryClient]
  );

  const analyzeVideo = useCallback(
    async (videoId: string, title: string) => {
      setActionError(null);
      await analyzeVideoMutation.mutateAsync({ videoId, title });
    },
    [analyzeVideoMutation]
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
          : analyzeVideoMutation.error
            ? getErrorMessage(analyzeVideoMutation.error)
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
      analyzeVideoMutation.isPending ||
      addChannelMutation.isPending,
    error: combinedError,
    fetchNews,
    fetchChannels,
    fetchLatestVideos,
    analyzeVideo,
    searchChannels,
    addChannel,
  };
}
