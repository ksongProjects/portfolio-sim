"use client";

import { Bell } from "lucide-react";
import { Badge } from "@/components/ui/badge";

interface NotificationBellProps {
  unreadCount: number;
  onClick: () => void;
}

export function NotificationBell({ unreadCount, onClick }: NotificationBellProps) {
  return (
    <button
      onClick={onClick}
      className="relative flex items-center justify-center h-9 w-9 rounded-md text-on-surface-variant hover:text-on-surface hover:bg-surface-container transition-colors"
      aria-label="Notifications"
    >
      <Bell className="h-5 w-5" />
      {unreadCount > 0 && (
        <span className="absolute -top-0.5 -right-0.5 flex h-4 w-4 items-center justify-center">
          <Badge variant="error" className="h-4 w-4 p-0 flex items-center justify-center">
            {unreadCount > 99 ? "99+" : unreadCount}
          </Badge>
        </span>
      )}
    </button>
  );
}