"use client";

import { useCallback, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { apiFetch, fetchJson } from "@/lib/api";

export interface Notification {
  id: string;
  title: string;
  message: string;
  type: "info" | "success" | "warning" | "error";
  timestamp: Date;
  read: boolean;
}

type NotificationPayload = {
  id: string;
  title: string;
  message: string;
  type: string;
  timestamp: string;
  read?: boolean;
};

async function fetchNotificationsData() {
  const data = await fetchJson<NotificationPayload[]>(
    "/api/notifications",
    undefined,
    "Failed to fetch"
  );

  return Array.isArray(data)
    ? data.map((notification) => ({
        ...notification,
        type: notification.type as Notification["type"],
        timestamp: new Date(notification.timestamp),
        read: notification.read ?? false,
      }))
    : [];
}

export function useNotifications() {
  const queryClient = useQueryClient();
  const [drawerOpen, setDrawerOpen] = useState(false);

  const notificationsQuery = useQuery({
    queryKey: ["notifications"],
    queryFn: fetchNotificationsData,
    retry: false,
  });

  const notifications = notificationsQuery.data ?? [];
  const unreadCount = notifications.filter((notification) => !notification.read).length;

  const updateNotifications = useCallback(
    (updater: (current: Notification[]) => Notification[]) => {
      queryClient.setQueryData<Notification[]>(["notifications"], (current = []) =>
        updater(current)
      );
    },
    [queryClient]
  );

  const fetchNotifications = useCallback(async () => {
    try {
      await notificationsQuery.refetch();
    } catch {
      // silent fail
    }
  }, [notificationsQuery]);

  const markAsRead = useCallback(
    async (id: string) => {
      try {
        await apiFetch("/api/notifications/dismiss", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ id }),
        });
      } catch {
        // silent fail
      }

      updateNotifications((current) =>
        current.map((notification) =>
          notification.id === id ? { ...notification, read: true } : notification
        )
      );
    },
    [updateNotifications]
  );

  const markAllAsRead = useCallback(() => {
    updateNotifications((current) =>
      current.map((notification) => ({ ...notification, read: true }))
    );
  }, [updateNotifications]);

  const addNotification = useCallback(
    (notification: Omit<Notification, "id" | "timestamp" | "read">) => {
      const newNotification: Notification = {
        ...notification,
        id: crypto.randomUUID(),
        timestamp: new Date(),
        read: false,
      };

      updateNotifications((current) => [newNotification, ...current]);

      if (notification.type === "error") {
        toast.error(notification.title, { description: notification.message });
      } else if (notification.type === "warning") {
        toast.warning(notification.title, { description: notification.message });
      } else if (notification.type === "success") {
        toast.success(notification.title, { description: notification.message });
      } else {
        toast(notification.title, { description: notification.message });
      }
    },
    [updateNotifications]
  );

  const removeNotification = useCallback(
    async (id: string) => {
      try {
        await apiFetch("/api/notifications/dismiss", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ id }),
        });
      } catch {
        // silent fail
      }

      updateNotifications((current) =>
        current.filter((notification) => notification.id !== id)
      );
    },
    [updateNotifications]
  );

  const clearAll = useCallback(async () => {
    const ids = notifications.map((n) => n.id);
    await Promise.allSettled(
      ids.map((id) =>
        apiFetch("/api/notifications/dismiss", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ id }),
        })
      )
    );
    updateNotifications(() => []);
  }, [notifications, updateNotifications]);

  return {
    notifications,
    unreadCount,
    drawerOpen,
    setDrawerOpen,
    markAsRead,
    markAllAsRead,
    addNotification,
    removeNotification,
    clearAll,
    refresh: fetchNotifications,
  };
}
