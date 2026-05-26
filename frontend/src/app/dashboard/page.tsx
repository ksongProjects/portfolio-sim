"use client";

import { useState } from "react";
import { PageGrid, PageCell, PageHeader, MetricLabel, MetricValue, MetricSubValue } from "@/components/page-layout";
import { CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { usePortfolioPerformance } from "@/hooks/usePortfolio";
import { useLivePositions } from "@/hooks/useLivePositions";
import { useLiveIndices } from "@/hooks/useLiveIndices";
import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from "recharts";

function fmtCurrency(v: number): string {
  return new Intl.NumberFormat("en-US", { style: "currency", currency: "USD" }).format(v);
}
function fmtPct(v: number): string {
  return (v >= 0 ? "+" : "") + v.toFixed(2) + "%";
}

type ChartRange = "1d" | "1w" | "1m";

function generate1dTicks(from: Date, to: Date): Date[] {
  const ticks: Date[] = [];
  const start = new Date(from);
  start.setMinutes(0, 0, 0);
  const end = new Date(to);
  end.setMinutes(0, 0, 0);
  const current = new Date(start);
  while (current <= end) {
    const day = current.getDay();
    if (day !== 0 && day !== 6) {
      const hour = current.getHours();
      if (hour >= 4 && hour < 20) {
        ticks.push(new Date(current));
      }
    }
    current.setHours(current.getHours() + 1);
  }
  return ticks;
}

function generate1wTicks(from: Date, to: Date): Date[] {
  const ticks: Date[] = [];
  const current = new Date(from);
  current.setHours(4, 0, 0, 0);
  const end = new Date(to);
  while (current <= end) {
    const day = current.getDay();
    if (day !== 0 && day !== 6) {
      ticks.push(new Date(current));
    }
    current.setDate(current.getDate() + 1);
  }
  return ticks;
}

function generate1mTicks(from: Date, to: Date): Date[] {
  const ticks: Date[] = [];
  const current = new Date(from);
  current.setHours(4, 0, 0, 0);
  const end = new Date(to);
  while (current <= end) {
    const day = current.getDay();
    if (day !== 0 && day !== 6) {
      ticks.push(new Date(current));
    }
    current.setDate(current.getDate() + 1);
  }
  return ticks;
}

function getTicksFromData(data: { timestamp: string }[], range: ChartRange): string[] {
  if (data.length === 0) return [];
  const seen = new Set<string>();
  const ticks: string[] = [];
  for (const d of data) {
    const date = new Date(d.timestamp);
    if (date.getDay() === 0 || date.getDay() === 6) continue;
    const hour = date.getHours();
    if (hour < 4 || hour >= 20) continue;
    if (range === "1d") {
      if (date.getMinutes() === 0) {
        const key = date.getTime().toString();
        if (!seen.has(key)) {
          seen.add(key);
          ticks.push(d.timestamp);
        }
      }
    } else {
      date.setHours(4, 0, 0, 0);
      const key = date.getTime().toString();
      if (!seen.has(key)) {
        seen.add(key);
        ticks.push(date.toISOString());
      }
    }
  }
  return ticks;
}

function formatXAxisTick(timestamp: string, range: ChartRange): string {
  const date = new Date(timestamp);
  if (isNaN(date.getTime())) return "";
  switch (range) {
    case "1d":
      return date.toLocaleTimeString("en-US", { hour: "numeric", minute: "2-digit", timeZone: "America/New_York" });
    case "1w":
      return date.toLocaleDateString("en-US", { weekday: "short", month: "short", day: "numeric", timeZone: "America/New_York" });
    case "1m":
      return date.toLocaleDateString("en-US", { month: "short", day: "numeric", timeZone: "America/New_York" });
    default:
      return date.toLocaleDateString("en-US", { month: "short", day: "numeric", timeZone: "America/New_York" });
  }
}

function fmtSignedCurrency(v: number): string {
  return `${v >= 0 ? "+" : "-"}${fmtCurrency(Math.abs(v))}`;
}

function NoDataMessage() {
  return (
    <div className="h-full flex items-center justify-center text-on-surface-variant text-sm">
      No data available
    </div>
  );
}

interface PortfolioLineChartProps {
  data: { timestamp: string; value: number }[];
  range: string;
  dataDate?: string;
}

function PortfolioLineChart({ data, range, dataDate }: PortfolioLineChartProps) {
  if (data.length === 0) {
    return <NoDataMessage />;
  }

  const chartData = data.map((d) => ({
    timestamp: d.timestamp,
    value: d.value,
  }));

  const minValue = Math.min(...data.map((d) => d.value));
  const maxValue = Math.max(...data.map((d) => d.value));
  const padding = (maxValue - minValue) * 0.1;

  const formatValue = (v: number) => {
    if (v >= 1000) return `$${(v / 1000).toFixed(0)}k`;
    return `$${v.toFixed(0)}`;
  };

  const firstValue = data[0]?.value ?? 0;
  const lastValue = data[data.length - 1]?.value ?? 0;
  const isUp = lastValue >= firstValue;
  const color = isUp ? "#3fe56c" : "#ff4d4d";

  const intervalLabel = range === "1d" ? "Today" : range === "1w" ? "This Week" : range === "1m" ? "This Month" : range;

  const ticks = getTicksFromData(data, range as ChartRange);

  const dateRangeLabel = dataDate ? `${dataDate}` : (range === "1d" ? "Today" : range === "1w" ? "This Week" : range === "1m" ? "This Month" : "");

  return (
    <div className="relative h-full w-full">
      {dateRangeLabel && (
        <div className="absolute top-2 right-12 z-10 px-2 py-1 bg-surface-container-high/90 rounded text-xs text-on-surface font-mono">
          {dateRangeLabel}
        </div>
      )}
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={chartData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="var(--outline-variant)" />
        <XAxis
          dataKey="timestamp"
          tick={{ fontSize: 10, fill: "#9ca3af" }}
          tickLine={true}
          axisLine={{ stroke: "var(--outline)" }}
          tickFormatter={(ts) => formatXAxisTick(ts.toString(), range as ChartRange)}
          ticks={ticks}
        />
        <YAxis
          domain={[minValue - padding, maxValue + padding]}
          tick={{ fontSize: 10, fill: "#9ca3af" }}
          tickLine={true}
          axisLine={false}
          tickFormatter={formatValue}
          width={60}
        />
        <Tooltip
          formatter={(value: number) => [fmtCurrency(value), "Portfolio Value"]}
          labelFormatter={(label) => `${intervalLabel} - ${label}`}
          contentStyle={{
            backgroundColor: "var(--surface-container-high)",
            border: "1px solid var(--outline-variant)",
            borderRadius: "8px",
            fontSize: "12px",
          }}
        />
        <Line
          type="monotone"
          dataKey="value"
          stroke={color}
          strokeWidth={2}
          dot={false}
          activeDot={{ r: 4, fill: color }}
        />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}

export default function DashboardPage() {
  const { positions, summary, indices: restIndices } = useLivePositions();
  const { indices: liveIndices, isLive } = useLiveIndices();
  const [activePeriod, setActivePeriod] = useState("1d");

  const indices = isLive ? liveIndices : restIndices;

  const { data: performance } = usePortfolioPerformance("default", activePeriod);

  const defaultPeriod = performance?.range ?? "1d";

  const posValue = summary?.TotalValue ?? 0;
  const posDayChange = summary?.DayChange ?? 0;
  const posDayPct = summary?.DayChangePct ?? 0;
  const posTotalGain = summary?.TotalGain ?? 0;
  const posTotalGainPct = summary?.TotalGainPct ?? 0;
  const posCash = summary?.CashBalance ?? 0;

  return (
    <div className="flex flex-col h-full">
      <div className="px-6 pt-6 pb-4">
        <PageHeader
          title="Dashboard"
          description="Portfolio overview and performance metrics"
        />
      </div>

      <div className="flex-1 px-6 pb-6 overflow-auto">
        <PageGrid style={{ gridTemplateColumns: "repeat(4, 1fr)" }}>
          {[1, 2, 3, 4].map((i) => (
            <PageCell key={i}>
              <MetricLabel>{i === 1 ? "Total Portfolio Value" : i === 2 ? "Today's P&L" : i === 3 ? "Total Gain/Loss" : "Cash Balance"}</MetricLabel>
              <MetricValue>{i === 1 ? fmtCurrency(posValue) : i === 2 ? fmtCurrency(posDayChange) : i === 3 ? fmtCurrency(posTotalGain) : fmtCurrency(posCash)}</MetricValue>
              <MetricSubValue positive>
                {i === 1 ? `${fmtCurrency(posDayChange)} (${fmtPct(posDayPct)})` : i === 2 ? fmtPct(posDayPct) : i === 3 ? fmtPct(posTotalGainPct) : "Buying Power"}
              </MetricSubValue>
            </PageCell>
          ))}
        </PageGrid>

        <PageGrid style={{ gridTemplateColumns: "7fr 5fr", gridTemplateRows: "1fr 1fr" }}>
          <PageCell className="col-span-1 row-span-1">
            <div className="flex items-center justify-between mb-4">
              <CardTitle>Portfolio Performance</CardTitle>
              <div className="flex items-center gap-2">
                {performance?.dataDate && (
                  <span className="text-xs text-on-surface-variant">{performance.dataDate}</span>
                )}
                {["1d", "1w", "1m", "3m", "1y", "all"].map((period) => (
                  <button
                    key={period}
                    onClick={() => setActivePeriod(period)}
                    className={`px-2 py-1 text-[11px] font-semibold uppercase tracking-[0.06em] transition-colors ${
                      period === activePeriod
                        ? "bg-primary text-on-primary"
                        : "text-on-surface-variant hover:text-on-surface hover:bg-surface-container"
                    }`}
                  >
                    {period.toUpperCase()}
                  </button>
                ))}
              </div>
            </div>
            <div className="h-[200px] border border-outline-variant/30">
              {performance?.data && performance.data.length > 0 ? (
                <PortfolioLineChart data={performance.data} range={activePeriod.toLowerCase()} dataDate={performance.dataDate} />
              ) : (
                <NoDataMessage />
              )}
            </div>
          </PageCell>

          <PageCell className="col-span-1 row-span-1">
            <CardTitle className="mb-4">Top Holdings</CardTitle>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Symbol</TableHead>
                  <TableHead>Price</TableHead>
                  <TableHead>Value</TableHead>
                  <TableHead>Change</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {positions.slice(0, 5).map((pos) => (
                  <TableRow key={pos.ID}>
                    <TableCell>
                      <span className="font-mono font-semibold">{pos.Symbol}</span>
                    </TableCell>
                    <TableCell className="font-mono">{fmtCurrency(pos.CurrentPrice)}</TableCell>
                    <TableCell className="font-mono">{fmtCurrency(pos.CurrentValue)}</TableCell>
                    <TableCell className={`font-mono ${pos.DayChangePct >= 0 ? "text-primary" : "text-error"}`}>
                      {fmtPct(pos.DayChangePct)}
                    </TableCell>
                  </TableRow>
                ))}
                {positions.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={4} className="text-center text-on-surface-variant text-sm py-8">No positions available</TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </PageCell>

          <PageCell className="col-span-1 row-span-1">
            <CardTitle className="mb-4">Recent Activity</CardTitle>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Type</TableHead>
                  <TableHead>Symbol</TableHead>
                  <TableHead>Shares</TableHead>
                  <TableHead>Price</TableHead>
                  <TableHead>Total</TableHead>
                  <TableHead>Time</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {positions.slice(0, 4).map((pos) => (
                  <TableRow key={pos.ID}>
                    <TableCell><Badge variant="success">BUY</Badge></TableCell>
                    <TableCell className="font-mono font-semibold">{pos.Symbol}</TableCell>
                    <TableCell className="font-mono">{pos.Quantity}</TableCell>
                    <TableCell className="font-mono">{fmtCurrency(pos.AvgCost)}</TableCell>
                    <TableCell className="font-mono">{fmtCurrency(pos.Quantity * pos.AvgCost)}</TableCell>
                    <TableCell className="text-on-surface-variant">Today</TableCell>
                  </TableRow>
                ))}
                {positions.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center text-on-surface-variant text-sm py-8">No recent activity</TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </PageCell>

          <PageCell className="col-span-1 row-span-1">
            <CardTitle className="mb-4">Market Indices</CardTitle>
            <div className="grid grid-cols-2 gap-2">
              {indices.length > 0 ? indices.map((index) => (
                <div key={index.Symbol} className="flex items-center justify-between p-3 border border-outline-variant/30">
                  <div>
                    <div className="font-mono text-sm font-semibold">{index.Symbol}</div>
                    <div className="text-[11px] text-on-surface-variant">{index.Name}</div>
                  </div>
                  <div className="text-right">
                    <div className="font-mono text-sm">{fmtCurrency(index.Price)}</div>
                    <div className={`text-[11px] font-mono ${index.ChangePct >= 0 ? "text-primary" : "text-error"}`}>
                      {fmtSignedCurrency(index.Change)} ({fmtPct(index.ChangePct)})
                    </div>
                  </div>
                </div>
              )) : (
                <div className="col-span-2 text-center text-on-surface-variant text-sm py-8">No market data available</div>
              )}
            </div>
          </PageCell>
        </PageGrid>
      </div>
    </div>
  );
}
