"use client";

import { useState, useCallback } from "react";
import { ColumnDef } from "@tanstack/react-table";
import { PageGrid, PageCell, PageHeader, MetricLabel, MetricValue } from "@/components/page-layout";
import { CardTitle } from "@/components/ui/card";
import { Badge, StatusIndicator } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/ui/data-table";
import { Skeleton } from "@/components/ui/skeleton";
import { AlertTriangle, CheckCircle, Clock, RefreshCw, Zap, ChevronDown, ChevronRight, Copy, Check } from "lucide-react";
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

function formatJson(obj: unknown): string {
  if (!obj) return "";
  if (typeof obj === "string") return obj;
  return JSON.stringify(obj, null, 2);
}

interface LogRowProps {
  id: string;
  message: string;
  metadata: Record<string, unknown> | null;
  isExpanded: boolean;
  onToggle: (id: string) => void;
}

function LogRow({ id, message, metadata, isExpanded, onToggle }: LogRowProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(formatJson(metadata));
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  }, [metadata]);

  return (
    <div className="flex flex-col gap-1 py-1">
      <span className="text-sm">{message}</span>
      {metadata && (
        <div className="flex items-center gap-2">
          <button
            onClick={() => onToggle(id)}
            className="flex items-center gap-1 text-xs text-primary hover:underline"
          >
            {isExpanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
            {isExpanded ? "Hide" : "Show"} Details
          </button>
          <button
            onClick={handleCopy}
            className="flex items-center gap-1 text-xs text-on-surface-variant hover:text-on-surface"
          >
            {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
            {copied ? "Copied!" : "Copy"}
          </button>
        </div>
      )}
      {isExpanded && metadata && (
        <pre className="mt-2 p-2 bg-surface-container-high rounded text-xs overflow-auto max-h-40 text-left whitespace-pre-wrap">
          {formatJson(metadata)}
        </pre>
      )}
    </div>
  );
}

export default function ObservabilityPage() {
  const { services, logs, loading, error, lastUpdated, refresh, filters, setLogFilters, fetchLogs } = useObservability();
  const [expandedLogIds, setExpandedLogIds] = useState<Set<string>>(new Set());
  const [logsLimit, setLogsLimit] = useState(100);

  const toggleLogExpansion = useCallback((id: string) => {
    setExpandedLogIds(prev => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  }, []);

  const healthyCount = services.filter((s) => s.status === "healthy").length;
  const warningCount = services.filter((s) => s.status === "warning").length;
  const errorCount = services.filter((s) => s.status === "error").length;

  const handleRefresh = () => {
    refresh();
  };

  const expandedColumns: ColumnDef<{
    id: string;
    timestamp: string;
    level: string;
    service: string;
    message: string;
    metadata: Record<string, unknown> | null;
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
      cell: ({ row }) => (
        <LogRow
          id={row.original.id}
          message={row.original.message}
          metadata={row.original.metadata}
          isExpanded={expandedLogIds.has(row.original.id)}
          onToggle={toggleLogExpansion}
        />
      ),
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
            <div className="flex items-center justify-between mb-4">
              <CardTitle>Recent Logs</CardTitle>
            </div>
            <div className="flex flex-wrap gap-3 mb-4 p-3 bg-surface-container-low rounded-md">
              <div className="flex items-center gap-2">
                <label className="text-xs text-on-surface-variant">Level:</label>
                <select
                  value={filters.level || ""}
                  onChange={(e) => setLogFilters({ ...filters, level: e.target.value || undefined })}
                  className="h-8 rounded-md border border-outline bg-surface-container px-2 text-sm text-on-surface focus:outline-none focus:ring-1 focus:ring-primary"
                >
                  <option value="">All</option>
                  <option value="DEBUG">DEBUG</option>
                  <option value="INFO">INFO</option>
                  <option value="WARN">WARN</option>
                  <option value="ERROR">ERROR</option>
                </select>
              </div>
              <div className="flex items-center gap-2">
                <label className="text-xs text-on-surface-variant">Service:</label>
                <select
                  value={filters.service || ""}
                  onChange={(e) => setLogFilters({ ...filters, service: e.target.value || undefined })}
                  className="h-8 rounded-md border border-outline bg-surface-container px-2 text-sm text-on-surface focus:outline-none focus:ring-1 focus:ring-primary"
                >
                  <option value="">All</option>
                  <option value="backend">Backend</option>
                  <option value="market-data-service">Market Data</option>
                  <option value="news-feed-service">News Feed</option>
                  <option value="frontend">Frontend</option>
                </select>
              </div>
              <div className="flex items-center gap-2">
                <label className="text-xs text-on-surface-variant">Action:</label>
                <select
                  value={filters.action || ""}
                  onChange={(e) => setLogFilters({ ...filters, action: e.target.value || undefined })}
                  className="h-8 rounded-md border border-outline bg-surface-container px-2 text-sm text-on-surface focus:outline-none focus:ring-1 focus:ring-primary"
                >
                  <option value="">All</option>
                  <option value="subscribe">subscribe</option>
                  <option value="unsubscribe">unsubscribe</option>
                </select>
              </div>
              <div className="flex items-center gap-2">
                <label className="text-xs text-on-surface-variant">Logs:</label>
                <select
                  value={logsLimit}
                  onChange={(e) => setLogsLimit(parseInt(e.target.value))}
                  className="h-8 rounded-md border border-outline bg-surface-container px-2 text-sm text-on-surface focus:outline-none focus:ring-1 focus:ring-primary"
                >
                  <option value="100">100</option>
                  <option value="500">500</option>
                  <option value="1000">1000</option>
                </select>
              </div>
              <Button
                variant="secondary"
                size="sm"
                onClick={() => fetchLogs(logsLimit, filters)}
                className="h-8 text-xs"
              >
                Get Logs
              </Button>
            </div>
            {error && logs.length === 0 && (
              <div className="text-error text-sm mb-4 p-3 border border-error/30 rounded bg-error/10">
                Failed to fetch logs. Is the logging service running?
              </div>
            )}
            <DataTable
              columns={expandedColumns}
              data={logs}
              loading={loading}
              emptyMessage="No logs available"
              enablePagination={false}
              maxHeight="80vh"
            />
          </PageCell>
        </PageGrid>
      </div>
    </div>
  );
}
