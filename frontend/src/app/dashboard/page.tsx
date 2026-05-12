"use client";

import { useState } from "react";
import { PageGrid, PageCell, PageHeader, MetricLabel, MetricValue, MetricSubValue } from "@/components/page-layout";
import { CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Skeleton, MetricSkeleton } from "@/components/ui/skeleton";
import { Activity } from "lucide-react";
import { usePortfolio } from "@/hooks/usePortfolio";

function fmtCurrency(v: number): string {
  return new Intl.NumberFormat("en-US", { style: "currency", currency: "USD" }).format(v);
}
function fmtPct(v: number): string {
  return (v >= 0 ? "+" : "") + v.toFixed(2) + "%";
}

function NoDataMessage() {
  return (
    <div className="h-full flex items-center justify-center text-on-surface-variant text-sm">
      No data available
    </div>
  );
}

function MiniChart() {
  return (
    <svg viewBox="0 0 400 120" className="w-full h-full" preserveAspectRatio="none">
      <defs>
        <linearGradient id="chartFill" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="#3fe56c" stopOpacity="0.15" />
          <stop offset="100%" stopColor="#3fe56c" stopOpacity="0" />
        </linearGradient>
      </defs>
      <polyline
        points="0,100 40,88 80,75 120,82 160,65 200,55 240,48 280,52 320,38 360,30 400,25"
        fill="none"
        stroke="#3fe56c"
        strokeWidth="1.5"
        vectorEffect="non-scaling-stroke"
      />
      <polygon
        points="0,100 40,88 80,75 120,82 160,65 200,55 240,48 280,52 320,38 360,30 400,25 400,120 0,120"
        fill="url(#chartFill)"
        vectorEffect="none"
      />
    </svg>
  );
}

export default function DashboardPage() {
  const { positions, summary, indices, loading } = usePortfolio();
  const [simRunning, setSimRunning] = useState(false);
  const [toast, setToast] = useState<string | null>(null);
  const [activePeriod, setActivePeriod] = useState("1M");

  const showToast = (msg: string) => {
    setToast(msg);
    setTimeout(() => setToast(null), 3000);
  };

  const handleRunSimulation = async () => {
    setSimRunning(true);
    showToast("Simulation started...");
    await new Promise((r) => setTimeout(r, 2000));
    setSimRunning(false);
    showToast("Simulation completed.");
  };

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
        >
          <Button variant="default" size="sm" onClick={handleRunSimulation} disabled={simRunning}>
            <Activity className="h-4 w-4" />
            {simRunning ? "Running..." : "Run Simulation"}
          </Button>
        </PageHeader>
      </div>

      <div className="flex-1 px-6 pb-6 overflow-auto">
        <PageGrid style={{ gridTemplateColumns: "repeat(4, 1fr)" }}>
          {[1, 2, 3, 4].map((i) => (
            <PageCell key={i}>
              {loading ? <MetricSkeleton /> : (
                <>
                  <MetricLabel>{i === 1 ? "Total Portfolio Value" : i === 2 ? "Today's P&L" : i === 3 ? "Total Gain/Loss" : "Cash Balance"}</MetricLabel>
                  <MetricValue>{i === 1 ? fmtCurrency(posValue) : i === 2 ? fmtCurrency(posDayChange) : i === 3 ? fmtCurrency(posTotalGain) : fmtCurrency(posCash)}</MetricValue>
                  <MetricSubValue positive>
                    {i === 1 ? `${fmtCurrency(posDayChange)} (${fmtPct(posDayPct)})` : i === 2 ? fmtPct(posDayPct) : i === 3 ? fmtPct(posTotalGainPct) : "Buying Power"}
                  </MetricSubValue>
                </>
              )}
            </PageCell>
          ))}
        </PageGrid>

        <PageGrid style={{ gridTemplateColumns: "7fr 5fr", gridTemplateRows: "1fr 1fr" }}>
          <PageCell className="col-span-1 row-span-1">
            <div className="flex items-center justify-between mb-4">
              <CardTitle>Portfolio Performance</CardTitle>
              <div className="flex gap-1">
                {["1D", "1W", "1M", "3M", "1Y", "ALL"].map((period) => (
                  <button
                    key={period}
                    onClick={() => setActivePeriod(period)}
                    className={`px-2 py-1 text-[11px] font-semibold uppercase tracking-[0.06em] transition-colors ${
                      period === activePeriod
                        ? "bg-primary text-on-primary"
                        : "text-on-surface-variant hover:text-on-surface hover:bg-surface-container"
                    }`}
                  >
                    {period}
                  </button>
                ))}
              </div>
            </div>
            <div className="h-[200px] border border-outline-variant/30">
              {loading ? <div className="h-full flex items-center justify-center"><Skeleton className="h-full w-full" /></div> : positions.length > 0 ? <MiniChart /> : <NoDataMessage />}
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
              {loading ? (
                <TableBody>
                  {[1, 2, 3, 4, 5].map((i) => (
                    <TableRow key={i}>
                      <TableCell><Skeleton className="h-4 w-12" /></TableCell>
                      <TableCell><Skeleton className="h-4 w-16" /></TableCell>
                      <TableCell><Skeleton className="h-4 w-20" /></TableCell>
                      <TableCell><Skeleton className="h-4 w-14" /></TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              ) : (
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
              )}
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
              {loading ? (
                <TableBody>
                  {[1, 2, 3, 4].map((i) => (
                    <TableRow key={i}>
                      <TableCell><Skeleton className="h-5 w-12" /></TableCell>
                      <TableCell><Skeleton className="h-4 w-14" /></TableCell>
                      <TableCell><Skeleton className="h-4 w-12" /></TableCell>
                      <TableCell><Skeleton className="h-4 w-16" /></TableCell>
                      <TableCell><Skeleton className="h-4 w-20" /></TableCell>
                      <TableCell><Skeleton className="h-4 w-12" /></TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              ) : (
                <TableBody>
                  {positions.slice(0, 4).map((pos) => (
                    <TableRow key={pos.ID}>
                      <TableCell><Badge variant="success">BUY</Badge></TableCell>
                      <TableCell className="font-mono font-semibold">{pos.Symbol}</TableCell>
                      <TableCell className="font-mono">{pos.Quantity}</TableCell>
                      <TableCell className="font-mono">{fmtCurrency(pos.AvgCost)}</TableCell>
                      <TableCell className="font-mono">{fmtCurrency(pos.CurrentValue)}</TableCell>
                      <TableCell className="text-on-surface-variant">Today</TableCell>
                    </TableRow>
                  ))}
                  {positions.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={6} className="text-center text-on-surface-variant text-sm py-8">No recent activity</TableCell>
                    </TableRow>
                  )}
                </TableBody>
              )}
            </Table>
          </PageCell>

          <PageCell className="col-span-1 row-span-1">
            <CardTitle className="mb-4">Market Indices</CardTitle>
            <div className="grid grid-cols-2 gap-2">
              {loading ? (
                Array.from({ length: 4 }).map((_, i) => (
                  <div key={i} className="flex items-center justify-between p-3 border border-outline-variant/30">
                    <div className="space-y-1">
                      <Skeleton className="h-4 w-16" />
                      <Skeleton className="h-3 w-20" />
                    </div>
                    <div className="space-y-1 text-right">
                      <Skeleton className="h-4 w-16 ml-auto" />
                      <Skeleton className="h-3 w-12 ml-auto" />
                    </div>
                  </div>
                ))
              ) : indices.length > 0 ? indices.map((index) => (
                <div key={index.Symbol} className="flex items-center justify-between p-3 border border-outline-variant/30">
                  <div>
                    <div className="font-mono text-sm font-semibold">{index.Symbol}</div>
                    <div className="text-[11px] text-on-surface-variant">{index.Name}</div>
                  </div>
                  <div className="text-right">
                    <div className="font-mono text-sm">{fmtCurrency(index.Price)}</div>
                    <div className={`text-[11px] font-mono ${index.ChangePct >= 0 ? "text-primary" : "text-error"}`}>
                      {fmtPct(index.ChangePct)}
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

      {toast && (
        <div className="fixed bottom-6 right-6 z-50 bg-surface-container-highest border border-primary text-on-surface px-4 py-3 text-sm font-medium">
          {toast}
        </div>
      )}
    </div>
  );
}
