"use client";

import { useEffect, useState } from "react";
import { Badge } from "@/components/ui/badge";

type MarketStatus = "Pre-Market" | "Regular Hours" | "After Hours" | "Closed";

interface MarketState {
  status: MarketStatus;
  time: string;
}

function formatEstTime(date: Date): string {
  return date.toLocaleTimeString("en-US", {
    timeZone: "America/New_York",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: true,
  });
}

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
  const [marketState, setMarketState] = useState<MarketState>({
    status: "Closed",
    time: "--:--:--",
  });

  useEffect(() => {
    const update = () => {
      const now = new Date();
      setMarketState({
        status: getMarketStatus(now),
        time: formatEstTime(now),
      });
    };

    update();
    const interval = setInterval(update, 1000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="flex items-center gap-2">
      <span className="text-on-surface-variant text-xs font-medium">
        EST {marketState.time}
      </span>
      <Badge variant={getBadgeVariant(marketState.status)}>
        {marketState.status}
      </Badge>
    </div>
  );
}

function getMarketStatus(date: Date): MarketStatus {
  const hours = date.getHours();
  const minutes = date.getMinutes();
  const totalMinutes = hours * 60 + minutes;

  const preMarketStart = 4 * 60;
  const marketOpen = 9 * 60 + 30;
  const marketClose = 16 * 60;
  const afterHoursEnd = 20 * 60;

  if (totalMinutes >= preMarketStart && totalMinutes < marketOpen) {
    return "Pre-Market";
  }
  if (totalMinutes >= marketOpen && totalMinutes < marketClose) {
    return "Regular Hours";
  }
  if (totalMinutes >= marketClose && totalMinutes < afterHoursEnd) {
    return "After Hours";
  }
  return "Closed";
}

export function useMarketStatus() {
  const [status, setStatus] = useState<MarketStatus>("Closed");

  useEffect(() => {
    const update = () => {
      const now = new Date();
      setStatus(getMarketStatus(now));
    };

    update();
    const interval = setInterval(update, 1000);
    return () => clearInterval(interval);
  }, []);

  const isLive = status === "Pre-Market" || status === "Regular Hours" || status === "After Hours";
  return { status, isLive };
}
