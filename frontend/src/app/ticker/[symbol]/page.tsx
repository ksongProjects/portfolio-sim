"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, TrendingUp, TrendingDown, Building, DollarSign, BarChart2, Percent, Calendar, Activity, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/page-layout";
import { Spinner } from "@/components/ui/skeleton";
import { useTickerLookup } from "@/hooks/useTickerLookup";
import { useMarketSocket } from "@/hooks/useMarketSocket";
import { usePortfolio } from "@/hooks/usePortfolio";
import { cn } from "@/lib/utils";
import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid, ReferenceLine } from "recharts";
import { toast } from "sonner";
import { useMemo } from "react";
import { type IntradayBar } from "@/hooks/useTickerLookup";

function fmtCurrency(v: number): string {
  return new Intl.NumberFormat("en-US", { style: "currency", currency: "USD" }).format(v);
}

function fmtPct(v: number): string {
  return (v >= 0 ? "+" : "") + v.toFixed(2) + "%";
}

function fmtNumber(v: number | undefined | null): string {
  if (v === undefined || v === null || isNaN(v)) return "N/A";
  if (v >= 1e12) return (v / 1e12).toFixed(2) + "T";
  if (v >= 1e9) return (v / 1e9).toFixed(2) + "B";
  if (v >= 1e6) return (v / 1e6).toFixed(2) + "M";
  if (v >= 1e3) return (v / 1e3).toFixed(2) + "K";
  return v.toFixed(2);
}

function formatTime(timestamp: string): string {
  const date = new Date(timestamp);
  if (isNaN(date.getTime())) {
    const parts = timestamp.split(" ");
    return parts.length > 1 ? parts[1].substring(0, 5) : timestamp;
  }
  return date.toLocaleTimeString("en-US", { hour: "2-digit", minute: "2-digit", timeZone: "America/New_York" });
}

function formatHourLabel(timestamp: string): string {
  const date = new Date(timestamp);
  if (isNaN(date.getTime())) return "";
  return date.toLocaleTimeString("en-US", { hour: "numeric", minute: "2-digit", timeZone: "America/New_York" });
}

function formatDateTimeLabel(timestamp: string): string {
  const date = new Date(timestamp);
  if (isNaN(date.getTime())) return "";
  return date.toLocaleString("en-US", { month: "short", day: "numeric", hour: "numeric", minute: "2-digit", timeZone: "America/New_York" });
}

function formatDayLabel(timestamp: string): string {
  const date = new Date(timestamp);
  if (isNaN(date.getTime())) return "";
  return date.toLocaleDateString([], { month: "short", day: "numeric" });
}

type ChartRange = "1d" | "1w" | "1m";

function generate1dSlots(from: Date, to: Date): Date[] {
  const slots: Date[] = [];
  const current = new Date(from);
  current.setSeconds(0, 0);
  const end = new Date(to);
  end.setSeconds(0, 0);
  while (current <= end) {
    const day = current.getDay();
    if (day !== 0 && day !== 6) {
      const hour = current.getHours();
      if (hour >= 4 && hour < 20) {
        slots.push(new Date(current));
      }
    }
    current.setMinutes(current.getMinutes() + 1);
  }
  return slots;
}

function generate1wSlots(from: Date, to: Date): Date[] {
  const slots: Date[] = [];
  const current = new Date(from);
  current.setHours(4, 0, 0, 0);
  const end = new Date(to);
  while (current <= end) {
    const day = current.getDay();
    if (day !== 0 && day !== 6) {
      const hour = current.getHours();
      if (hour >= 4 && hour < 20) {
        slots.push(new Date(current));
      }
    }
    current.setDate(current.getDate() + 1);
    current.setHours(4, 0, 0, 0);
  }
  return slots;
}

function generate1mSlots(from: Date, to: Date): Date[] {
  const slots: Date[] = [];
  const current = new Date(from);
  current.setHours(4, 0, 0, 0);
  const end = new Date(to);
  while (current <= end) {
    const day = current.getDay();
    if (day !== 0 && day !== 6) {
      const hour = current.getHours();
      if (hour >= 4 && hour < 20) {
        slots.push(new Date(current));
      }
    }
    current.setDate(current.getDate() + 1);
    current.setHours(4, 0, 0, 0);
  }
  return slots;
}

function generateSlotsForRange(range: ChartRange, from: Date, to: Date): Date[] {
  switch (range) {
    case "1d":
      return generate1dSlots(from, to);
    case "1w":
      return generate1wSlots(from, to);
    case "1m":
      return generate1mSlots(from, to);
    default:
      return [];
  }
}

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

function getTicksForRange(range: ChartRange, from: Date, to: Date): Date[] {
  switch (range) {
    case "1d":
      return generate1dTicks(from, to);
    case "1w":
      return generate1wTicks(from, to);
    case "1m":
      return generate1mTicks(from, to);
    default:
      return [];
  }
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

function prepareChartData(
  rawData: { timestamp: string; open?: number; high?: number; low?: number; close: number; volume?: number }[],
  range: ChartRange
): { timestamp: string; open: number; high: number; low: number; close: number; volume: number }[] {
  if (rawData.length === 0) return [];

  const dataMap = new Map<string, { open: number; high: number; low: number; close: number; volume: number }>();
  for (const bar of rawData) {
    const key = new Date(bar.timestamp).getTime().toString();
    const existing = dataMap.get(key);
    if (existing) {
      existing.high = Math.max(existing.high, bar.high ?? bar.close);
      existing.low = Math.min(existing.low, bar.low ?? bar.close);
      existing.close = bar.close;
      existing.volume += bar.volume ?? 0;
    } else {
      dataMap.set(key, {
        open: bar.open ?? bar.close,
        high: bar.high ?? bar.close,
        low: bar.low ?? bar.close,
        close: bar.close,
        volume: bar.volume ?? 0,
      });
    }
  }

  const from = new Date(rawData[0].timestamp);
  const to = new Date(rawData[rawData.length - 1].timestamp);
  const expectedSlots = generateSlotsForRange(range, from, to);

  const result: { timestamp: string; open: number; high: number; low: number; close: number; volume: number }[] = [];

  for (const slot of expectedSlots) {
    const key = slot.getTime().toString();
    const bar = dataMap.get(key);
    if (bar && bar.close !== 0) {
      result.push({ timestamp: slot.toISOString(), open: bar.open, high: bar.high, low: bar.low, close: bar.close, volume: bar.volume });
    }
  }

  return result;
}

function CustomTooltip({ active, payload }: { active?: boolean; payload?: Array<{ value: number; payload: { timestamp: string; open: number; high: number; low: number; close: number; volume: number } }> }) {
  if (!active || !payload || payload.length === 0) return null;
  const data = payload[0].payload;
  return (
    <div className="bg-surface-container-high border border-outline-variant p-3 text-xs">
      <div className="font-mono font-semibold text-on-surface mb-1">{formatTime(data.timestamp)}</div>
      <div className="grid grid-cols-2 gap-x-4 gap-y-1">
        <span className="text-on-surface-variant">Open</span><span className="font-mono text-on-surface">{(data.open ?? 0).toFixed(2)}</span>
        <span className="text-on-surface-variant">High</span><span className="font-mono text-on-surface">{(data.high ?? 0).toFixed(2)}</span>
        <span className="text-on-surface-variant">Low</span><span className="font-mono text-on-surface">{(data.low ?? 0).toFixed(2)}</span>
        <span className="text-on-surface-variant">Close</span><span className="font-mono text-on-surface">{(data.close ?? 0).toFixed(2)}</span>
        <span className="text-on-surface-variant">Volume</span><span className="font-mono text-on-surface">{(data.volume ?? 0).toLocaleString()}</span>
      </div>
    </div>
  );
}

function IntradayChart({ data, range = "1d" }: { data: { timestamp: string; open?: number; high?: number; low?: number; close: number; volume?: number }[]; range?: ChartRange }) {
  if (data.length === 0) {
    return <div className="h-64 flex items-center justify-center text-on-surface-variant">No chart data available</div>;
  }

  const chartData = prepareChartData(data, range);

  const firstClose = chartData[0]?.close || 0;
  const lastClose = chartData[chartData.length - 1]?.close || 0;
  const isUp = lastClose >= firstClose;
  const color = isUp ? "#3fe56c" : "#ff4d4d";

  const from = new Date(data[0].timestamp);
  const to = new Date(data[data.length - 1].timestamp);
  const ticks = getTicksForRange(range, from, to);

  return (
    <ResponsiveContainer width="100%" height={280}>
      <LineChart data={chartData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="var(--outline-variant)" />
        <XAxis
          dataKey="timestamp"
          tick={{ fontSize: 10, fill: "#9ca3af" }}
          tickLine={true}
          axisLine={{ stroke: "var(--outline)" }}
          tickFormatter={(ts) => formatXAxisTick(ts.toString(), range)}
          ticks={ticks.map(t => t.toISOString())}
        />
        <YAxis
          domain={["auto", "auto"]}
          tick={{ fontSize: 10, fill: "#9ca3af" }}
          tickLine={true}
          axisLine={false}
          tickFormatter={(v) => v.toFixed(0)}
          width={60}
          tickCount={5}
        />
        <Tooltip content={<CustomTooltip />} />
        <Line
          type="monotone"
          dataKey="close"
          stroke={color}
          strokeWidth={2}
          dot={false}
          activeDot={{ r: 4, fill: color }}
        />
      </LineChart>
    </ResponsiveContainer>
  );
}

export default function TickerPage() {
  const params = useParams();
  const symbol = params.symbol as string;
  const { selectedTicker, intradayData, intradayChange, intradayChangePct, loading: tickerLoading, chartLoading, chartRange, setChartRange } = useTickerLookup(symbol);
  const { positions, removePosition, isRemovingPosition } = usePortfolio("default", { includeIndices: false });
  const position = positions.find(p => p.Symbol === symbol);
  const { getBar } = useMarketSocket([symbol], true);

const chartData = useMemo(() => {
    const baseData: IntradayBar[] = [...intradayData];
    const liveBar = getBar(symbol);
    if (liveBar && liveBar.close > 0) {
      const lastIdx = baseData.findIndex(b => {
        const d = new Date(b.timestamp);
        const lb = new Date(liveBar.timestamp);
        return d.getTime() === lb.getTime();
      });
      const barEntry: IntradayBar = { timestamp: liveBar.timestamp, close: liveBar.close, volume: liveBar.volume ?? baseData[0]?.volume, open: liveBar.open, high: liveBar.high, low: liveBar.low };
      if (lastIdx >= 0) {
        baseData[lastIdx] = barEntry;
      } else {
        baseData.push(barEntry);
      }
    }
    return baseData;
  }, [intradayData, symbol, getBar]);

  const dateRangeLabel = useMemo(() => {
    if (chartData.length === 0) return "";
    const first = new Date(chartData[0].timestamp);
    const last = new Date(chartData[chartData.length - 1].timestamp);
    const fmt = (d: Date) => d.toLocaleDateString("en-US", { month: "short", day: "numeric" });
    if (first.toDateString() === last.toDateString()) {
      return fmt(first);
    }
    if (chartRange === "1w" || chartRange === "1m") {
      return `${fmt(first)} - ${fmt(last)}`;
    }
    return fmt(first);
  }, [chartData, chartRange]);

  const handleRemovePosition = async () => {
    if (!position) return;
    try {
      await removePosition("default", position.ID);
      toast.success(`Removed ${symbol} from portfolio`);
    } catch {
      toast.error(`Failed to remove ${symbol}`);
    }
  };

  if (tickerLoading && !selectedTicker) {
    return (
      <div className="flex flex-col h-full">
        <div className="px-6 pt-6 pb-4">
          <div className="h-8 w-32 bg-surface-container-high animate-pulse rounded" />
        </div>
        <div className="flex-1 px-6 pb-6 flex items-center justify-center">
          <div className="text-on-surface-variant">Loading {symbol}...</div>
        </div>
      </div>
    );
  }

  if (!selectedTicker && !tickerLoading) {
    return (
      <div className="flex flex-col h-full">
        <div className="px-6 pt-6 pb-4">
          <div className="h-8 w-32 bg-surface-container-high animate-pulse rounded" />
        </div>
        <div className="flex-1 px-6 pb-6 flex items-center justify-center">
          <div className="text-on-surface-variant">Ticker not found</div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      <div className="px-6 pt-6 pb-4">
        <PageHeader 
          title={selectedTicker?.symbol ?? symbol} 
          description={selectedTicker?.name}
        >
          <div className="flex gap-2">
            {position && (
              <Button variant="destructive" size="sm" onClick={handleRemovePosition} disabled={isRemovingPosition}>
                <Trash2 className="h-4 w-4" /> Remove Position
              </Button>
            )}
            <Link href="/portfolio">
              <Button variant="secondary" size="sm">
                <ArrowLeft className="h-4 w-4" /> Back to Portfolio
              </Button>
            </Link>
          </div>
        </PageHeader>
      </div>

      <div className="flex-1 px-6 pb-6 overflow-auto space-y-6">
        <div className="flex items-start justify-between">
          <div>
            <div className="flex items-center gap-2">
              <span className="text-2xl font-mono font-bold text-on-surface">{selectedTicker?.symbol}</span>
              <Badge variant="secondary">{selectedTicker?.exchange}</Badge>
            </div>
          </div>
          <div className="text-right">
            <div className="text-2xl font-mono font-semibold text-on-surface">{fmtCurrency(selectedTicker?.price ?? 0)}</div>
            <div className={cn("text-sm font-mono flex items-center gap-1", intradayChangePct >= 0 ? "text-primary" : "text-error")}>
              {intradayChangePct >= 0 ? <TrendingUp className="h-4 w-4" /> : <TrendingDown className="h-4 w-4" />}
              {fmtCurrency(Math.abs(intradayChange))} ({fmtPct(intradayChangePct)})
            </div>
          </div>
        </div>

        <div className="bg-surface-container-low border border-outline-variant p-4">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-1">
              <Activity className="h-4 w-4 text-on-surface-variant" />
              <span className="text-xs font-semibold uppercase tracking-[0.08em] text-on-surface-variant">Intraday Chart</span>
              {dateRangeLabel && (
                <span className="text-xs text-on-surface-variant/70 font-mono ml-1">({dateRangeLabel})</span>
              )}
            </div>
            <div className="flex gap-1">
              {(["1d", "1w", "1m"] as const).map((r) => (
                <button
                  key={r}
                  onClick={() => setChartRange(r)}
                  className={cn(
                    "px-2 py-1 text-xs rounded",
                    chartRange === r ? "bg-primary text-on-primary" : "text-on-surface-variant hover:bg-surface-container-high"
                  )}
                >
                  {r.toUpperCase()}
                </button>
              ))}
            </div>
          </div>
          {chartLoading ? (
            <div className="h-64 flex flex-col items-center justify-center gap-3">
              <Spinner className="h-8 w-8" />
              <span className="text-sm text-on-surface-variant">Loading chart data...</span>
            </div>
          ) : (
            <IntradayChart data={chartData} range={chartRange} />
          )}
        </div>

        <div className="grid grid-cols-4 gap-3">
          {[
            { icon: DollarSign, label: "Market Cap", value: fmtNumber(selectedTicker?.marketCap) },
            { icon: BarChart2, label: "P/E Ratio", value: selectedTicker?.peRatio?.toFixed(2) || "N/A" },
            { icon: Percent, label: "Div Yield", value: selectedTicker?.dividendYield ? (selectedTicker.dividendYield * 100).toFixed(2) + "%" : "N/A" },
            { icon: Calendar, label: "52W Range", value: `${fmtCurrency(selectedTicker?.week52Low ?? 0)} - ${fmtCurrency(selectedTicker?.week52High ?? 0)}` },
          ].map((item) => (
            <div key={item.label} className="flex items-center gap-2 p-3 border border-outline-variant/30">
              <item.icon className="h-4 w-4 text-on-surface-variant shrink-0" />
              <div>
                <div className="text-[10px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant">{item.label}</div>
                <div className="text-sm font-mono text-on-surface">{item.value}</div>
              </div>
            </div>
          ))}
        </div>

        <div className="grid grid-cols-2 gap-6">
          <div className="bg-surface-container-low border border-outline-variant p-4">
            <div className="flex items-center gap-1 mb-3">
              <Building className="h-4 w-4 text-on-surface-variant" />
              <CardTitle>Company Details</CardTitle>
            </div>
            <div className="space-y-3">
              {[
                { label: "Sector", value: selectedTicker?.sector || "N/A" },
                { label: "Industry", value: selectedTicker?.industry || "N/A" },
                { label: "Volume", value: selectedTicker?.volume?.toLocaleString() || "N/A" },
                { label: "Avg Volume", value: selectedTicker?.avgVolume?.toLocaleString() || "N/A" },
                { label: "EPS", value: selectedTicker?.eps?.toFixed(2) || "N/A" },
              ].map((item) => (
                <div key={item.label} className="flex items-center justify-between">
                  <span className="text-sm text-on-surface-variant">{item.label}</span>
                  <span className="text-sm font-mono text-on-surface">{item.value}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
