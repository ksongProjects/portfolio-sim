"use client";

import { useState } from "react";
import { PageGrid, PageCell, PageHeader, MetricLabel, MetricValue } from "@/components/page-layout";
import { CardTitle } from "@/components/ui/card";
import { Badge, StatusIndicator } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Activity, AlertTriangle, CheckCircle, Clock, RefreshCw, Zap } from "lucide-react";

type ServiceStatus = "healthy" | "warning" | "error";

interface SystemMetric {
  id: number;
  name: string;
  value: string;
  status: ServiceStatus;
  threshold: string;
}

interface LogEntry {
  id: number;
  level: "INFO" | "WARN" | "ERROR";
  service: string;
  message: string;
  time: string;
}

interface Service {
  name: string;
  status: ServiceStatus;
  uptime: string;
  lastCheck: string;
}

const initialMetrics: SystemMetric[] = [
  { id: 1, name: "API Response Time", value: "45ms", status: "healthy", threshold: "<100ms" },
  { id: 2, name: "Database Latency", value: "12ms", status: "healthy", threshold: "<50ms" },
  { id: 3, name: "Cache Hit Rate", value: "94.2%", status: "healthy", threshold: ">85%" },
  { id: 4, name: "Message Queue", value: "23", status: "warning", threshold: "<100" },
  { id: 5, name: "Active Connections", value: "1,847", status: "healthy", threshold: "<5000" },
  { id: 6, name: "Error Rate", value: "0.12%", status: "healthy", threshold: "<1%" },
];

const initialServices: Service[] = [
  { name: "Trading Engine", status: "healthy", uptime: "99.98%", lastCheck: "Just now" },
  { name: "Signal Generator", status: "healthy", uptime: "99.95%", lastCheck: "Just now" },
  { name: "Data Feed", status: "warning", uptime: "99.12%", lastCheck: "2 min ago" },
  { name: "Portfolio Manager", status: "healthy", uptime: "99.99%", lastCheck: "Just now" },
  { name: "Notification Service", status: "error", uptime: "98.45%", lastCheck: "5 min ago" },
];

const initialLogs: LogEntry[] = [
  { id: 1, level: "INFO", service: "trading-engine", message: "Order executed: NVDA 50 shares @ $860.00", time: "10:32:15 AM" },
  { id: 2, level: "INFO", service: "signal-generator", message: "Strategy signal generated: BUY NVDA", time: "10:32:10 AM" },
  { id: 3, level: "WARN", service: "data-feed", message: "Delayed data for source: BLOOMBERG", time: "10:31:58 AM" },
  { id: 4, level: "INFO", service: "portfolio-manager", message: "Portfolio rebalanced successfully", time: "10:30:00 AM" },
  { id: 5, level: "ERROR", service: "notification-service", message: "Failed to send webhook notification", time: "10:28:45 AM" },
  { id: 6, level: "INFO", service: "trading-engine", message: "Position closed: AAPL 25 shares @ $177.50", time: "09:15:30 AM" },
];

export default function ObservabilityPage() {
  const [metrics, setMetrics] = useState<SystemMetric[]>(initialMetrics);
  const [services] = useState<Service[]>(initialServices);
  const [logs, setLogs] = useState<LogEntry[]>(initialLogs);
  const [refreshing, setRefreshing] = useState(false);
  const [toast, setToast] = useState<string | null>(null);

  const showToast = (msg: string) => {
    setToast(msg);
    setTimeout(() => setToast(null), 3000);
  };

  const handleRefresh = async () => {
    setRefreshing(true);
    showToast("Refreshing metrics...");
    await new Promise((r) => setTimeout(r, 1500));
    setMetrics((prev) =>
      prev.map((m) => ({
        ...m,
        value: m.name === "API Response Time" ? `${Math.floor(Math.random() * 30 + 40)}ms` : m.name === "Cache Hit Rate" ? `${(Math.random() * 5 + 92).toFixed(1)}%` : m.value,
      }))
    );
    setLogs((prev) => [
      {
        id: Date.now(),
        level: "INFO",
        service: "system",
        message: "Metrics refreshed successfully",
        time: new Date().toLocaleTimeString("en-US", { hour: "2-digit", minute: "2-digit", second: "2-digit" }) + " " + new Date().toLocaleTimeString("en-US", { hour: "numeric", hour12: true }).split(" ")[1],
      },
      ...prev,
    ]);
    setRefreshing(false);
    showToast("Metrics updated");
  };

  const healthyCount = services.filter((s) => s.status === "healthy").length;
  const warningCount = services.filter((s) => s.status === "warning").length;
  const errorCount = services.filter((s) => s.status === "error").length;

  return (
    <div className="flex flex-col h-full">
      <div className="px-6 pt-6 pb-4">
        <PageHeader title="Observability" description="System health and monitoring">
          <div className="flex gap-3">
            <Button variant="secondary" size="sm" onClick={handleRefresh} disabled={refreshing}>
              <RefreshCw className={`h-4 w-4 ${refreshing ? "animate-spin" : ""}`} />
              {refreshing ? "Refreshing..." : "Refresh"}
            </Button>
            <Button variant="default" size="sm" onClick={() => showToast("Live dashboard coming soon...")}>
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
                <MetricLabel>Avg Response</MetricLabel>
                <MetricValue>{metrics.find((m) => m.name === "API Response Time")?.value}</MetricValue>
              </div>
            </div>
          </PageCell>
        </PageGrid>

        <PageGrid className="mb-4" style={{ gridTemplateColumns: "1fr" }}>
          <PageCell>
            <CardTitle className="mb-4">System Metrics</CardTitle>
            <div className="border border-outline-variant/30">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Metric</TableHead>
                    <TableHead className="text-right">Current Value</TableHead>
                    <TableHead className="text-right">Threshold</TableHead>
                    <TableHead className="text-center">Status</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {metrics.map((metric) => (
                    <TableRow key={metric.id}>
                      <TableCell className="font-medium">{metric.name}</TableCell>
                      <TableCell className="text-right font-mono">{metric.value}</TableCell>
                      <TableCell className="text-right font-mono text-on-surface-variant">{metric.threshold}</TableCell>
                      <TableCell className="text-center">
                        <Badge variant={metric.status === "healthy" ? "success" : metric.status === "warning" ? "warning" : "error"}>{metric.status}</Badge>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </PageCell>
        </PageGrid>

        <PageGrid className="mt-px" style={{ gridTemplateColumns: "1fr" }}>
          <PageCell>
            <CardTitle className="mb-4">Services</CardTitle>
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
                      <TableCell><Badge variant={log.level === "INFO" ? "secondary" : log.level === "WARN" ? "warning" : "error"}>{log.level}</Badge></TableCell>
                      <TableCell className="font-mono text-xs">{log.service}</TableCell>
                      <TableCell className="text-sm">{log.message}</TableCell>
                      <TableCell className="font-mono text-xs text-on-surface-variant"><div className="flex items-center gap-1"><Clock className="h-3 w-3" />{log.time}</div></TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </PageCell>
        </PageGrid>
      </div>

      {toast && (
        <div className="fixed bottom-6 right-6 z-50 bg-surface-container-highest border border-primary text-on-surface px-4 py-3 text-sm font-medium">{toast}</div>
      )}
    </div>
  );
}
