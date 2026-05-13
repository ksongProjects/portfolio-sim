"use client";

import { Badge } from "@/components/ui/badge";
import { useMarketStatus, type MarketStatus } from "@/hooks/useMarketStatus";

function getBadgeVariant(status: MarketStatus) {
  switch (status) {
    case "Pre-Market":
      return "warning";
    case "Regular Hours":
      return "success";
    case "After Hours":
      return "warning";
    case "Closed":
      return "secondary";
  }
}

export function LiveMarketIndicator() {
  const { status, time } = useMarketStatus();

  return (
    <div className="flex items-center gap-2">
      <span className="text-on-surface-variant text-xs font-medium">
        ET {time}
      </span>
      <Badge variant={getBadgeVariant(status)}>
        {status}
      </Badge>
    </div>
  );
}
