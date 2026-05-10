"use client";

import { useState, useCallback, useEffect } from "react";
import { toast } from "sonner";

export interface Notification {
  id: string;
  title: string;
  message: string;
  type: "info" | "success" | "warning" | "error";
  timestamp: Date;
  read: boolean;
}

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export function useNotifications() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [drawerOpen, setDrawerOpen] = useState(false);

  const fetchNotifications = useCallback(async () => {
    try {
      const res = await fetch(`${API_BASE}/api/notifications`);
      if (!res.ok) throw new Error("Failed to fetch");
      const data = await res.json();
      setNotifications(
        Array.isArray(data)
          ? data.map((n: { id: string; title: string; message: string; type: string; timestamp: string; read?: boolean }) => ({
              ...n,
              type: n.type as Notification["type"],
              timestamp: new Date(n.timestamp),
              read: n.read ?? false,
            }))
          : []
      );
    } catch {
      // silent fail
    }
  }, []);

  useEffect(() => {
    let cancelled = false;

    void (async () => {
      try {
        const res = await fetch(`${API_BASE}/api/notifications`);
        if (!res.ok) throw new Error("Failed to fetch");
        const data = await res.json();
        if (cancelled) return;
        setNotifications(
          Array.isArray(data)
            ? data.map((n: { id: string; title: string; message: string; type: string; timestamp: string; read?: boolean }) => ({
                ...n,
                type: n.type as Notification["type"],
                timestamp: new Date(n.timestamp),
                read: n.read ?? false,
              }))
            : []
        );
      } catch {
        // silent fail
      }
    })();

    return () => {
      cancelled = true;
    };
  }, []);

  const unreadCount = notifications.filter((n) => !n.read).length;

  const markAsRead = useCallback((id: string) => {
    setNotifications((prev) =>
      prev.map((n) => (n.id === id ? { ...n, read: true } : n))
    );
  }, []);

  const markAllAsRead = useCallback(() => {
    setNotifications((prev) => prev.map((n) => ({ ...n, read: true })));
  }, []);

  const addNotification = useCallback(
    (notification: Omit<Notification, "id" | "timestamp" | "read">) => {
      const newNotification: Notification = {
        ...notification,
        id: crypto.randomUUID(),
        timestamp: new Date(),
        read: false,
      };
      setNotifications((prev) => [newNotification, ...prev]);
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
    []
  );

  const removeNotification = useCallback((id: string) => {
    setNotifications((prev) => prev.filter((n) => n.id !== id));
  }, []);

  return {
    notifications,
    unreadCount,
    drawerOpen,
    setDrawerOpen,
    markAsRead,
    markAllAsRead,
    addNotification,
    removeNotification,
    refresh: fetchNotifications,
  };
}
