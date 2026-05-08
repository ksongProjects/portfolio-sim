import { PageCell, PageHeader } from "@/components/page-layout";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge, StatusIndicator } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Activity, AlertTriangle, CheckCircle, Clock, RefreshCw, Zap } from "lucide-react";

const systemMetrics = [
  { id: 1, name: "API Response Time", value: "45ms", status: "healthy", threshold: "<100ms" },
  { id: 2, name: "Database Latency", value: "12ms", status: "healthy", threshold: "<50ms" },
  { id: 3, name: "Cache Hit Rate", value: "94.2%", status: "healthy", threshold: ">85%" },
  { id: 4, name: "Message Queue", value: "23", status: "warning", threshold: "<100" },
  { id: 5, name: "Active Connections", value: "1,847", status: "healthy", threshold: "<5000" },
  { id: 6, name: "Error Rate", value: "0.12%", status: "healthy", threshold: "<1%" },
];

const recentLogs = [
  { id: 1, level: "INFO", service: "trading-engine", message: "Order executed: NVDA 50 shares @ $860.00", time: "10:32:15 AM" },
  { id: 2, level: "INFO", service: "signal-generator", message: "Strategy signal generated: BUY NVDA", time: "10:32:10 AM" },
  { id: 3, level: "WARN", service: "data-feed", message: "Delayed data for source: BLOOMBERG", time: "10:31:58 AM" },
  { id: 4, level: "INFO", service: "portfolio-manager", message: "Portfolio rebalanced successfully", time: "10:30:00 AM" },
  { id: 5, level: "ERROR", service: "notification-service", message: "Failed to send webhook notification", time: "10:28:45 AM" },
  { id: 6, level: "INFO", service: "trading-engine", message: "Position closed: AAPL 25 shares @ $177.50", time: "09:15:30 AM" },
];

const services = [
  { name: "Trading Engine", status: "healthy", uptime: "99.98%", lastCheck: "Just now" },
  { name: "Signal Generator", status: "healthy", uptime: "99.95%", lastCheck: "Just now" },
  { name: "Data Feed", status: "warning", uptime: "99.12%", lastCheck: "2 min ago" },
  { name: "Portfolio Manager", status: "healthy", uptime: "99.99%", lastCheck: "Just now" },
  { name: "Notification Service", status: "error", uptime: "98.45%", lastCheck: "5 min ago" },
];

export default function ObservabilityPage() {
  return (
    <div className="flex flex-col h-full">
      <PageHeader
        title="Observability"
        description="System health and monitoring"
      >
        <div className="flex gap-3">
          <Button variant="secondary" size="sm">
            <RefreshCw className="h-4 w-4" />
            Refresh
          </Button>
          <Button variant="default" size="sm">
            <Activity className="h-4 w-4" />
            Live Dashboard
          </Button>
        </div>
      </PageHeader>

      <div className="flex-1 p-6">
        <div className="flex flex-col gap-[1px] bg-outline-variant p-[1px]">
          <PageCell>
            <div className="grid grid-cols-4 gap-6">
              <div className="flex items-center gap-3">
                <div className="h-10 w-10 flex items-center justify-center bg-primary/10">
                  <CheckCircle className="h-5 w-5 text-primary" />
                </div>
                <div>
                  <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant">Healthy Services</div>
                  <div className="text-[13px] font-mono font-medium text-on-surface">3</div>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="h-10 w-10 flex items-center justify-center bg-warning/10">
                  <AlertTriangle className="h-5 w-5 text-warning" />
                </div>
                <div>
                  <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant">Warnings</div>
                  <div className="text-[13px] font-mono font-medium text-on-surface">1</div>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="h-10 w-10 flex items-center justify-center bg-error/10">
                  <AlertTriangle className="h-5 w-5 text-error" />
                </div>
                <div>
                  <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant">Errors</div>
                  <div className="text-[13px] font-mono font-medium text-on-surface">1</div>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="h-10 w-10 flex items-center justify-center bg-surface-container-high">
                  <Zap className="h-5 w-5 text-on-surface-variant" />
                </div>
                <div>
                  <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant">Avg Response</div>
                  <div className="text-[13px] font-mono font-medium text-on-surface">45ms</div>
                </div>
              </div>
            </div>
          </PageCell>
        </div>

        <div className="mt-[1px] flex gap-[1px] bg-outline-variant p-[1px]">
          <PageCell className="flex-[2]">
            <Card className="border-0 bg-transparent">
              <CardHeader className="px-0 pb-4">
                <CardTitle className="text-base font-semibold">System Metrics</CardTitle>
              </CardHeader>
              <CardContent className="px-0">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Metric</TableHead>
                      <TableHead>Current Value</TableHead>
                      <TableHead>Threshold</TableHead>
                      <TableHead>Status</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {systemMetrics.map((metric) => (
                      <TableRow key={metric.id}>
                        <TableCell className="font-medium">{metric.name}</TableCell>
                        <TableCell className="font-mono">{metric.value}</TableCell>
                        <TableCell className="font-mono text-on-surface-variant">{metric.threshold}</TableCell>
                        <TableCell>
                          <Badge
                            variant={metric.status === "healthy" ? "success" : metric.status === "warning" ? "warning" : "error"}
                          >
                            {metric.status}
                          </Badge>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          </PageCell>

          <PageCell className="flex-1">
            <Card className="border-0 bg-transparent">
              <CardHeader className="px-0 pb-4">
                <CardTitle className="text-base font-semibold">Services</CardTitle>
              </CardHeader>
              <CardContent className="px-0">
                <div className="space-y-3">
                  {services.map((service) => (
                    <div key={service.name} className="flex items-center justify-between p-3 border border-outline-variant">
                      <div className="flex items-center gap-3">
                        <StatusIndicator
                          active={service.status === "healthy"}
                          className={service.status === "warning" ? "bg-warning" : service.status === "error" ? "bg-error" : ""}
                        />
                        <div>
                          <div className="text-sm font-medium text-on-surface">{service.name}</div>
                          <div className="text-xs text-on-surface-variant">Uptime: {service.uptime}</div>
                        </div>
                      </div>
                      <div className="text-right">
                        <Badge
                          variant={service.status === "healthy" ? "success" : service.status === "warning" ? "warning" : "error"}
                          className="text-[10px]"
                        >
                          {service.status}
                        </Badge>
                        <div className="text-xs text-on-surface-variant mt-1">{service.lastCheck}</div>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </PageCell>
        </div>

        <div className="mt-[1px] bg-outline-variant p-[1px]">
          <PageCell>
            <Card className="border-0 bg-transparent">
              <CardHeader className="px-0 pb-4">
                <CardTitle className="text-base font-semibold">Recent Logs</CardTitle>
              </CardHeader>
              <CardContent className="px-0">
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
                    {recentLogs.map((log) => (
                      <TableRow key={log.id}>
                        <TableCell>
                          <Badge
                            variant={
                              log.level === "INFO" ? "secondary" : log.level === "WARN" ? "warning" : "error"
                            }
                            className="text-[10px]"
                          >
                            {log.level}
                          </Badge>
                        </TableCell>
                        <TableCell className="font-mono text-xs">{log.service}</TableCell>
                        <TableCell className="text-sm">{log.message}</TableCell>
                        <TableCell className="font-mono text-xs text-on-surface-variant">
                          <div className="flex items-center gap-1">
                            <Clock className="h-3 w-3" />
                            {log.time}
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          </PageCell>
        </div>
      </div>
    </div>
  );
}
