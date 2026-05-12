"use client";

import { useCallback } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { buildApiUrl, fetchJson, getErrorMessage } from "@/lib/api";

export interface ProviderConfig {
  id: string;
  provider_id: string;
  name: string;
  description: string;
  type: string;
  api_key_set: boolean;
  is_connected: boolean;
  token_expired: boolean;
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
  last_scrape_at: string;
  is_active: boolean;
}

async function fetchProvidersData() {
  return fetchJson<ProviderConfig[]>("/api/providers", undefined, "Failed to fetch providers");
}

async function fetchConnectionsData() {
  return fetchJson<ConnectionStatus[]>("/api/connections", undefined, "Failed to fetch connections");
}

async function fetchRSSFeedsData() {
  const data = await fetchJson<RSSFeed[]>("/api/rss-feeds", undefined, "Failed to fetch RSS feeds");
  return Array.isArray(data) ? data : [];
}

export function useProviders() {
  const queryClient = useQueryClient();

  const providersQuery = useQuery({
    queryKey: ["providers"],
    queryFn: fetchProvidersData,
  });

  const connectionsQuery = useQuery({
    queryKey: ["connections"],
    queryFn: fetchConnectionsData,
  });

  const rssFeedsQuery = useQuery({
    queryKey: ["rss-feeds"],
    queryFn: fetchRSSFeedsData,
  });

  const saveProviderKeyMutation = useMutation({
    mutationFn: async ({
      providerId,
      apiKey,
    }: {
      providerId: string;
      apiKey: string;
    }) => {
      const res = await fetch(buildApiUrl("/api/providers"), {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ provider_id: providerId, api_key: apiKey }),
      });

      if (!res.ok) {
        throw new Error("Failed to save");
      }
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["providers"] });
    },
  });

  const addRSSFeedMutation = useMutation({
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

  const deleteRSSFeedMutation = useMutation({
    mutationFn: async (feedId: string) => {
      const res = await fetch(buildApiUrl(`/api/rss-feeds?id=${encodeURIComponent(feedId)}`), {
        method: "DELETE",
      });

      if (!res.ok) {
        throw new Error("Failed to delete feed");
      }

      return feedId;
    },
    onSuccess: (feedId) => {
      queryClient.setQueryData<RSSFeed[]>(["rss-feeds"], (prev = []) =>
        prev.filter((feed) => feed.id !== feedId)
      );
    },
  });

  const refreshQuestradeTokenMutation = useMutation({
    mutationFn: async () => {
      const res = await fetch(buildApiUrl("/api/providers/questrade/refresh"), {
        method: "POST",
      });

      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error || "Failed to refresh");
      }
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["providers"] });
    },
  });

  const fetchProviders = useCallback(async () => {
    await providersQuery.refetch();
  }, [providersQuery]);

  const fetchConnections = useCallback(async () => {
    await connectionsQuery.refetch();
  }, [connectionsQuery]);

  const fetchRSSFeeds = useCallback(async () => {
    await rssFeedsQuery.refetch();
  }, [rssFeedsQuery]);

  const saveProviderKey = useCallback(
    async (providerId: string, apiKey: string): Promise<boolean> => {
      try {
        await saveProviderKeyMutation.mutateAsync({ providerId, apiKey });
        return true;
      } catch {
        return false;
      }
    },
    [saveProviderKeyMutation]
  );

  const validateProviderKey = useCallback(
    async (
      providerId: string,
      apiKey: string
    ): Promise<{ valid: boolean; error?: string }> => {
      try {
        const res = await fetch(buildApiUrl("/api/providers/validate"), {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ provider_id: providerId, api_key: apiKey }),
        });
        const data = await res.json();

        if (!res.ok || !data.valid) {
          throw new Error(data.error || "Failed to validate");
        }

        return data;
      } catch (err) {
        return { valid: false, error: getErrorMessage(err) };
      }
    },
    []
  );

  const addRSSFeed = useCallback(
    async (name: string, url: string): Promise<boolean> => {
      try {
        await addRSSFeedMutation.mutateAsync({ name, url });
        return true;
      } catch {
        return false;
      }
    },
    [addRSSFeedMutation]
  );

  const deleteRSSFeed = useCallback(
    async (feedId: string): Promise<boolean> => {
      try {
        await deleteRSSFeedMutation.mutateAsync(feedId);
        return true;
      } catch {
        return false;
      }
    },
    [deleteRSSFeedMutation]
  );

  const refresh = useCallback(async () => {
    await Promise.all([
      providersQuery.refetch(),
      connectionsQuery.refetch(),
      rssFeedsQuery.refetch(),
    ]);
  }, [connectionsQuery, providersQuery, rssFeedsQuery]);

  const refreshQuestradeToken = useCallback(async (): Promise<{ success: boolean; error?: string }> => {
    try {
      await refreshQuestradeTokenMutation.mutateAsync();
      return { success: true };
    } catch (err) {
      return { success: false, error: getErrorMessage(err) };
    }
  }, [refreshQuestradeTokenMutation]);

  const combinedError =
    providersQuery.error ??
    connectionsQuery.error ??
    rssFeedsQuery.error ??
    saveProviderKeyMutation.error ??
    addRSSFeedMutation.error ??
    deleteRSSFeedMutation.error ??
    refreshQuestradeTokenMutation.error;

  return {
    providers: providersQuery.data ?? [],
    connections: connectionsQuery.data ?? [],
    rssFeeds: rssFeedsQuery.data ?? [],
    loading:
      providersQuery.isFetching ||
      connectionsQuery.isFetching ||
      rssFeedsQuery.isFetching,
    error: combinedError ? getErrorMessage(combinedError) : null,
    fetchProviders,
    fetchConnections,
    fetchRSSFeeds,
    saveProviderKey,
    validateProviderKey,
    addRSSFeed,
    deleteRSSFeed,
    refresh,
    refreshQuestradeToken,
  };
}
