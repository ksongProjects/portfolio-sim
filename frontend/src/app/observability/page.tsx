"use client";

import { useEffect } from "react";
import { PageGrid, PageCell, PageHeader, MetricLabel, MetricValue } from "@/components/page-layout";
import { CardTitle } from "@/components/ui/card";
import { Badge, StatusIndicator } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Activity, AlertTriangle, CheckCircle, Clock, RefreshCw, Zap } from "lucide-react";
import { useObservability } from "@/hooks/useObservability";

function formatTimestamp(ts: string): string {
  try {
    const d = new Date(ts);
    return d.toLocaleTimeString("en-US", { hour: "2-digit", minute: "2-digit", second: "2-digit", hour12: true });
  } catch {
    return ts;
  }
}

export default function ObservabilityPage() {
  const { services, logs, loading, error, refresh } = useObservability();

  useEffect(() => {
    refresh();
  }, [refresh]);

  const healthyCount = services.filter((s) => s.status === "healthy").length;
  const warningCount = services.filter((s) => s.status === "warning").length;
  const errorCount = services.filter((s) => s.status === "error").length;

  return (
    <div className="flex flex-col h-full">
      <div className="px-6 pt-6 pb-4">
        <PageHeader title="Observability" description="System health and monitoring">
          <div className="flex gap-3">
            <Button variant="secondary" size="sm" onClick={refresh} disabled={loading}>
              <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
              {loading ? "Refreshing..." : "Refresh"}
            </Button>
            <Button variant="default" size="sm" onClick={() => {}} disabled={true}>
              <Activity className="h-4 w-4" /> Live Dashboard
            </Button>
          </div>
        </PageHeader>
      </div>

      <div className="flex-1 px-6 pb-6 overflow-auto">
        <PageGrid className="mb-4" style={{ gridTemplateColumns: "repeat(4, 1fr)" }}>
          <PageCell>
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 flex items-center justify-center bg-primary/10">
                <CheckCircle className="h-5 w-5 text-primary" />
              </div>
              <div>
                <MetricLabel>Healthy</MetricLabel>
                <MetricValue>{healthyCount}</MetricValue>
              </div>
            </div>
          </PageCell>
          <PageCell>
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 flex items-center justify-center bg-warning/10">
                <AlertTriangle className="h-5 w-5 text-warning" />
              </div>
              <div>
                <MetricLabel>Warnings</MetricLabel>
                <MetricValue highlight={warningCount > 0} style={{ color: warningCount > 0 ? "var(--color-warning)" : undefined }}>{warningCount}</MetricValue>
              </div>
            </div>
          </PageCell>
          <PageCell>
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 flex items-center justify-center bg-error/10">
                <AlertTriangle className="h-5 w-5 text-error" />
              </div>
              <div>
                <MetricLabel>Errors</MetricLabel>
                <MetricValue style={{ color: errorCount > 0 ? "var(--color-error)" : undefined }}>{errorCount}</MetricValue>
              </div>
            </div>
          </PageCell>
          <PageCell>
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 flex items-center justify-center bg-surface-container-high">
                <Zap className="h-5 w-5 text-on-surface-variant" />
              </div>
              <div>
                <MetricLabel>Services</MetricLabel>
                <MetricValue>{services.length}</MetricValue>
              </div>
            </div>
          </PageCell>
        </PageGrid>

        <PageGrid className="mt-px" style={{ gridTemplateColumns: "1fr" }}>
          <PageCell>
            <CardTitle className="mb-4">Services</CardTitle>
            {error && <div className="text-error text-sm mb-2">Error: {error}</div>}
            <div className="grid grid-cols-3 gap-2">
              {services.map((service) => (
                <div key={service.name} className="flex items-center justify-between p-4 border border-outline-variant/30">
                  <div className="flex items-center gap-3">
                    <StatusIndicator active={service.status === "healthy"} className={service.status === "warning" ? "bg-warning" : service.status === "error" ? "bg-error" : ""} />
                    <div>
                      <div className="text-sm font-medium">{service.name}</div>
                      <div className="text-[11px] text-on-surface-variant">Uptime: {service.uptime}</div>
                    </div>
                  </div>
                  <div className="text-right">
                    <Badge variant={service.status === "healthy" ? "success" : service.status === "warning" ? "warning" : "error"}>{service.status}</Badge>
                    <div className="text-[11px] text-on-surface-variant mt-1">{service.lastCheck}</div>
                  </div>
                </div>
              ))}
              {services.length === 0 && !loading && (
                <div className="col-span-3 text-center text-on-surface-variant text-sm py-8">No services data available</div>
              )}
            </div>
          </PageCell>
        </PageGrid>

        <PageGrid className="mt-px" style={{ gridTemplateColumns: "1fr" }}>
          <PageCell>
            <CardTitle className="mb-4">Recent Logs</CardTitle>
            <div className="border border-outline-variant/30">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Level</TableHead>
                    <TableHead>Service</TableHead>
                    <TableHead>Message</TableHead>
                    <TableHead>Time</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {logs.map((log) => (
                    <TableRow key={log.id}>
                      <TableCell><Badge variant={log.level === "INFO" ? "secondary" : log.level === "WARN" ? "warning" : log.level === "ERROR" || log.level === "FATAL" ? "error" : "secondary"}>{log.level}</Badge></TableCell>
                      <TableCell className="font-mono text-xs">{log.service}</TableCell>
                      <TableCell className="text-sm">{log.message}</TableCell>
                      <TableCell className="font-mono text-xs text-on-surface-variant"><div className="flex items-center gap-1"><Clock className="h-3 w-3" />{formatTimestamp(log.timestamp)}</div></TableCell>
                    </TableRow>
                  ))}
                  {logs.length === 0 && !loading && (
                    <TableRow>
                      <TableCell colSpan={4} className="text-center text-on-surface-variant text-sm py-8">No logs available</TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          </PageCell>
        </PageGrid>
      </div>
    </div>
  );
}