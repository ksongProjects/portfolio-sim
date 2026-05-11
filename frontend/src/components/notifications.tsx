"use client";

import { useNotifications } from "@/hooks/useNotifications";
import { NotificationBell } from "@/components/notification-bell";
import { NotificationsDrawer } from "@/components/notifications-drawer";

export function Notifications() {
  const {
    notifications,
    unreadCount,
    drawerOpen,
    setDrawerOpen,
    markAsRead,
    markAllAsRead,
    removeNotification,
    clearAll,
  } = useNotifications();

  return (
    <>
      <NotificationBell unreadCount={unreadCount} onClick={() => setDrawerOpen(true)} />
      <NotificationsDrawer
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        notifications={notifications}
        onMarkAsRead={markAsRead}
        onMarkAllAsRead={markAllAsRead}
        onRemove={removeNotification}
        onClearAll={clearAll}
      />
    </>
  );
}