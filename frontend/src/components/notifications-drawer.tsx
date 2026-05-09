"use client";

import { X, Bell, Check, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import type { Notification } from "@/hooks/useNotifications";

interface NotificationsDrawerProps {
  open: boolean;
  onClose: () => void;
  notifications: Notification[];
  onMarkAsRead: (id: string) => void;
  onMarkAllAsRead: () => void;
  onRemove: (id: string) => void;
}

function getTypeStyles(type: Notification["type"]) {
  switch (type) {
    case "success":
      return "border-l-success";
    case "warning":
      return "border-l-warning";
    case "error":
      return "border-l-error";
    default:
      return "border-l-primary";
  }
}

export function NotificationsDrawer({
  open,
  onClose,
  notifications,
  onMarkAsRead,
  onMarkAllAsRead,
  onRemove,
}: NotificationsDrawerProps) {
  const unreadCount = notifications.filter((n) => !n.read).length;

  if (!open) return null;

  return (
    <>
      <div
        className="fixed inset-0 z-50 bg-black/40"
        onClick={onClose}
        aria-hidden="true"
      />
      <div className="fixed right-0 top-0 z-50 h-full w-96 max-w-full bg-surface border-l border-outline-variant shadow-xl flex flex-col">
        <div className="flex items-center justify-between px-4 py-3 border-b border-outline-variant">
          <div className="flex items-center gap-2">
            <Bell className="h-5 w-5 text-on-surface" />
            <span className="text-sm font-semibold text-on-surface">Notifications</span>
            {unreadCount > 0 && (
              <Badge variant="error">{unreadCount} new</Badge>
            )}
          </div>
          <div className="flex items-center gap-1">
            {unreadCount > 0 && (
              <Button variant="ghost" size="icon" onClick={onMarkAllAsRead} title="Mark all as read">
                <Check className="h-4 w-4" />
              </Button>
            )}
            <Button variant="ghost" size="icon" onClick={onClose} title="Close">
              <X className="h-4 w-4" />
            </Button>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto">
          {notifications.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-full text-on-surface-variant">
              <Bell className="h-10 w-10 mb-2 opacity-30" />
              <p className="text-sm">No notifications</p>
            </div>
          ) : (
            <ul className="divide-y divide-outline-variant/30">
              {notifications.map((notification) => (
                <li
                  key={notification.id}
                  className={cn(
                    "flex gap-3 p-4 transition-colors hover:bg-surface-container-low",
                    !notification.read && "bg-primary/5",
                    getTypeStyles(notification.type)
                  )}
                  onClick={() => !notification.read && onMarkAsRead(notification.id)}
                >
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className={cn("text-sm font-medium", !notification.read ? "text-on-surface" : "text-on-surface-variant")}>
                        {notification.title}
                      </span>
                      {!notification.read && (
                        <span className="h-2 w-2 rounded-full bg-primary shrink-0" />
                      )}
                    </div>
                    <p className="text-xs text-on-surface-variant mt-0.5 line-clamp-2">{notification.message}</p>
                    <span className="text-[10px] text-on-surface-variant/60 mt-1 block">
                      {notification.timestamp.toLocaleString()}
                    </span>
                  </div>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      onRemove(notification.id);
                    }}
                    className="shrink-0 text-on-surface-variant hover:text-error transition-colors"
                    aria-label="Remove notification"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </>
  );
}