"use client";

import { useCallback } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiFetch, fetchJson, getErrorMessage } from "@/lib/api";

export interface RSSFeed {
  id: string;
  name: string;
  url: string;
  is_active: boolean;
  last_scrape_at: string | null;
}

async function fetchFeedsData() {
  const data = await fetchJson<RSSFeed[]>("/api/rss-feeds", undefined, "Failed to fetch feeds");
  return Array.isArray(data) ? data : [];
}

export function useRSSFeeds() {
  const queryClient = useQueryClient();

  const feedsQuery = useQuery({
    queryKey: ["rss-feeds"],
    queryFn: fetchFeedsData,
  });

  const addFeedMutation = useMutation({
    mutationFn: async ({ name, url }: { name: string; url: string }) => {
      const data = await fetchJson<RSSFeed[]>(
        "/api/rss-feeds",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ name, url }),
        },
        "Failed to add feed"
      );

      return Array.isArray(data) ? data : [];
    },
    onSuccess: (feeds) => {
      queryClient.setQueryData(["rss-feeds"], feeds);
    },
  });

  const deleteFeedMutation = useMutation({
    mutationFn: async (id: string) => {
      const res = await apiFetch(`/api/rss-feeds?id=${encodeURIComponent(id)}`, {
        method: "DELETE",
      });

      if (!res.ok) {
        throw new Error("Failed to delete feed");
      }

      return id;
    },
    onSuccess: (id) => {
      queryClient.setQueryData<RSSFeed[]>(["rss-feeds"], (prev = []) =>
        prev.filter((feed) => feed.id !== id)
      );
    },
  });

  const scrapeFeedsMutation = useMutation({
    mutationFn: async () => {
      const res = await apiFetch("/api/rss-feeds/scrape", { method: "POST" });

      if (!res.ok) {
        throw new Error("Failed to trigger scrape");
      }
    },
  });

  const fetchFeeds = useCallback(async () => {
    await feedsQuery.refetch();
  }, [feedsQuery]);

  const addFeed = useCallback(
    async (name: string, url: string) => {
      try {
        await addFeedMutation.mutateAsync({ name, url });
        return true;
      } catch {
        return false;
      }
    },
    [addFeedMutation]
  );

  const deleteFeed = useCallback(
    async (id: string) => {
      try {
        await deleteFeedMutation.mutateAsync(id);
        return true;
      } catch {
        return false;
      }
    },
    [deleteFeedMutation]
  );

  const scrapeFeeds = useCallback(async () => {
    try {
      await scrapeFeedsMutation.mutateAsync();
      return true;
    } catch {
      return false;
    }
  }, [scrapeFeedsMutation]);

  const scrapeFeed = useCallback(
    async (feedId: string) => {
      try {
        await apiFetch(`/api/rss-feeds/${feedId}/scrape`, { method: "POST" });
        return true;
      } catch {
        return false;
      }
    },
    []
  );

  const combinedError =
    feedsQuery.error ?? addFeedMutation.error ?? deleteFeedMutation.error ?? scrapeFeedsMutation.error;

  return {
    feeds: feedsQuery.data ?? [],
    loading: feedsQuery.isFetching || addFeedMutation.isPending || deleteFeedMutation.isPending || scrapeFeedsMutation.isPending,
    error: combinedError ? getErrorMessage(combinedError) : null,
    fetchFeeds,
    addFeed,
    deleteFeed,
    scrapeFeeds,
    scrapeFeed,
    refetchFeeds: () => feedsQuery.refetch(),
  };
}
