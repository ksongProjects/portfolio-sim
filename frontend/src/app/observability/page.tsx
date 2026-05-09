"use client";

import { useState } from "react";
import { ColumnDef } from "@tanstack/react-table";
import { PageGrid, PageCell, PageHeader, MetricLabel, MetricValue } from "@/components/page-layout";
import { CardTitle } from "@/components/ui/card";
import { Badge, StatusIndicator } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/ui/data-table";
import { Skeleton } from "@/components/ui/skeleton";
import { AlertTriangle, CheckCircle, Clock, RefreshCw, Zap, Pause, Play } from "lucide-react";
import { useObservability } from "@/hooks/useObservability";

function formatTimestamp(ts: string): string {
  try {
    const d = new Date(ts);
    return d.toLocaleTimeString("en-US", { hour: "2-digit", minute: "2-digit", second: "2-digit", hour12: true });
  } catch {
    return ts;
  }
}

function formatTimeAgo(date: Date | null): string {
  if (!date) return "Never";
  const seconds = Math.floor((Date.now() - date.getTime()) / 1000);
  if (seconds < 5) return "Just now";
  if (seconds < 60) return `${seconds}s ago`;
  return date.toLocaleTimeString();
}

const logColumns: ColumnDef<{
  id: string;
  timestamp: string;
  level: string;
  service: string;
  message: string;
}>[] = [
  {
    accessorKey: "level",
    header: "Level",
    cell: ({ row }) => {
      const level = row.original.level;
      const variant = level === "INFO" || level === "DEBUG" ? "secondary" : level === "WARN" ? "warning" : "error";
      return <Badge variant={variant}>{level}</Badge>;
    },
  },
  {
    accessorKey: "service",
    header: "Service",
    cell: ({ row }) => <span className="font-mono text-xs">{row.original.service}</span>,
  },
  {
    accessorKey: "message",
    header: "Message",
    cell: ({ row }) => <span className="text-sm truncate max-w-md">{row.original.message}</span>,
  },
  {
    accessorKey: "timestamp",
    header: "Time",
    cell: ({ row }) => (
      <span className="font-mono text-xs text-on-surface-variant flex items-center gap-1">
        <Clock className="h-3 w-3" />
        {formatTimestamp(row.original.timestamp)}
      </span>
    ),
  },
];

export default function ObservabilityPage() {
  const { services, logs, loading, error, lastUpdated, refresh, startAutoRefresh, stopAutoRefresh } = useObservability();
  const [autoRefresh, setAutoRefresh] = useState(false);

  const healthyCount = services.filter((s) => s.status === "healthy").length;
  const warningCount = services.filter((s) => s.status === "warning").length;
  const errorCount = services.filter((s) => s.status === "error").length;

  const handleToggleAutoRefresh = () => {
    if (autoRefresh) {
      stopAutoRefresh();
    } else {
      startAutoRefresh();
    }
    setAutoRefresh(!autoRefresh);
  };

  return (
    <div className="flex flex-col h-full">
      <div className="px-6 pt-6 pb-4">
        <PageHeader title="Observability" description="System health and monitoring">
          <div className="flex items-center gap-2">
            {lastUpdated && (
              <span className="text-xs text-on-surface-variant mr-2">
                Updated: {formatTimeAgo(lastUpdated)}
              </span>
            )}
            <Button
              variant={autoRefresh ? "default" : "secondary"}
              size="sm"
              onClick={handleToggleAutoRefresh}
            >
              {autoRefresh ? <Pause className="h-4 w-4" /> : <Play className="h-4 w-4" />}
              {autoRefresh ? "Live" : "Auto-refresh"}
            </Button>
            <Button variant="secondary" size="sm" onClick={refresh} disabled={loading}>
              <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
              {loading ? "Refreshing..." : "Refresh"}
            </Button>
          </div>
        </PageHeader>
      </div>

      <div className="flex-1 px-6 pb-6 overflow-auto">
        <PageGrid className="mb-4" style={{ gridTemplateColumns: "repeat(4, 1fr)" }}>
          {[
            { icon: CheckCircle, label: "Healthy", count: healthyCount, color: "text-primary", bg: "bg-primary/10" },
            { icon: AlertTriangle, label: "Warnings", count: warningCount, color: warningCount > 0 ? "text-warning" : undefined, bg: "bg-warning/10" },
            { icon: AlertTriangle, label: "Errors", count: errorCount, color: errorCount > 0 ? "text-error" : undefined, bg: "bg-error/10" },
            { icon: Zap, label: "Services", count: services.length, color: undefined, bg: "bg-surface-container-high" },
          ].map(({ icon: Icon, label, count, color, bg }) => (
            <PageCell key={label}>
              {loading && services.length === 0 ? (
                <div className="space-y-2">
                  <Skeleton className="h-3 w-20" />
                  <Skeleton className="h-8 w-12" />
                </div>
              ) : (
                <div className="flex items-center gap-3">
                  <div className={`h-10 w-10 flex items-center justify-center ${bg}`}>
                    <Icon className={`h-5 w-5 ${color || "text-on-surface-variant"}`} />
                  </div>
                  <div>
                    <MetricLabel>{label}</MetricLabel>
                    <MetricValue style={color ? { color } : undefined}>{count}</MetricValue>
                  </div>
                </div>
              )}
            </PageCell>
          ))}
        </PageGrid>

        <PageGrid className="mt-px" style={{ gridTemplateColumns: "1fr" }}>
          <PageCell>
            <div className="flex items-center justify-between mb-4">
              <CardTitle>Services</CardTitle>
              <div className="flex items-center gap-2">
                {autoRefresh && (
                  <div className="flex items-center gap-1 text-xs text-primary">
                    <div className="h-2 w-2 rounded-full bg-primary animate-pulse" />
                    Live
                  </div>
                )}
              </div>
            </div>
            {error && services.length === 0 && (
              <div className="text-error text-sm mb-4 p-3 border border-error/30 rounded bg-error/10">
                Error: {error}
              </div>
            )}
            <div className="grid grid-cols-3 gap-2">
              {loading && services.length === 0 ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <div key={i} className="flex items-center justify-between p-4 border border-outline-variant/30">
                    <div className="flex items-center gap-3">
                      <Skeleton className="h-3 w-3 rounded-full" />
                      <div className="space-y-1">
                        <Skeleton className="h-4 w-24" />
                        <Skeleton className="h-3 w-16" />
                      </div>
                    </div>
                    <Skeleton className="h-5 w-14 rounded" />
                  </div>
                ))
              ) : (
                services.map((service) => (
                  <div key={service.name} className="flex items-center justify-between p-4 border border-outline-variant/30">
                    <div className="flex items-center gap-3">
                      <StatusIndicator
                        active={service.status === "healthy"}
                        className={
                          service.status === "warning" ? "bg-warning" :
                          service.status === "error" ? "bg-error" : ""
                        }
                      />
                      <div>
                        <div className="text-sm font-medium">{service.name}</div>
                        <div className="text-[11px] text-on-surface-variant">Uptime: {service.uptime}</div>
                      </div>
                    </div>
                    <div className="text-right">
                      <Badge variant={service.status === "healthy" ? "success" : service.status === "warning" ? "warning" : "error"}>
                        {service.status}
                      </Badge>
                      <div className="text-[11px] text-on-surface-variant mt-1">{service.lastCheck}</div>
                    </div>
                  </div>
                ))
              )}
              {services.length === 0 && !loading && (
                <div className="col-span-3 text-center text-on-surface-variant text-sm py-8">No services data available</div>
              )}
            </div>
          </PageCell>
        </PageGrid>

        <PageGrid className="mt-px" style={{ gridTemplateColumns: "1fr" }}>
          <PageCell>
            <CardTitle className="mb-4">Recent Logs</CardTitle>
            {error && logs.length === 0 && (
              <div className="text-error text-sm mb-4 p-3 border border-error/30 rounded bg-error/10">
                Failed to fetch logs. Is the logging service running?
              </div>
            )}
            <DataTable
              columns={logColumns}
              data={logs}
              loading={loading}
              emptyMessage="No logs available"
              enablePagination={true}
              pageSizes={[25, 50, 100]}
            />
          </PageCell>
        </PageGrid>
      </div>
    </div>
  );
}