"use client";

import { useEffect, useState } from "react";
import { PageGrid, PageCell, PageHeader, MetricLabel, MetricValue } from "@/components/page-layout";
import { CardTitle } from "@/components/ui/card";
import { Badge, StatusIndicator } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Play, Pause, Plus, Settings, Zap, BarChart3, TrendingUp, Shield } from "lucide-react";
import { useStrategies, Strategy } from "@/hooks/useStrategies";

function fmtPct(v: number): string {
  return (v >= 0 ? "+" : "") + v.toFixed(2) + "%";
}

export default function StrategyPage() {
  const { strategies, signals, loading, refresh } = useStrategies();
  const [toast, setToast] = useState<string | null>(null);

  useEffect(() => { refresh(); }, [refresh]);

  const showToast = (msg: string) => {
    setToast(msg);
    setTimeout(() => setToast(null), 3000);
  };

  const toggleStrategy = (strategy: Strategy) => {
    showToast(`${strategy.Name} toggled`);
  };

  const totalTrades = strategies.reduce((sum, s) => sum + s.Trades, 0);
  const activeCount = strategies.filter((s) => s.Status === "active").length;
  const avgWinRate = strategies.length > 0 ? (strategies.reduce((sum, s) => sum + s.WinRate, 0) / strategies.length).toFixed(1) : "0.0";
  const avgSharpe = strategies.length > 0 ? (strategies.reduce((sum, s) => sum + s.Sharpe, 0) / strategies.length).toFixed(2) : "0.00";

  return (
    <div className="flex flex-col h-full">
      <div className="px-6 pt-6 pb-4">
        <PageHeader title="Strategy" description="Trading strategies and signal generation">
          <div className="flex gap-3">
            <Button variant="secondary" size="sm" onClick={() => showToast("Strategy configuration coming soon...")}>
              <Settings className="h-4 w-4" /> Configure
            </Button>
            <Button variant="default" size="sm" onClick={() => showToast("New strategy editor coming soon...")}>
              <Plus className="h-4 w-4" /> New Strategy
            </Button>
          </div>
        </PageHeader>
      </div>

      <div className="flex-1 px-6 pb-6 overflow-auto">
        <PageGrid className="mb-4" style={{ gridTemplateColumns: "repeat(4, 1fr)" }}>
          <PageCell>
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 flex items-center justify-center bg-primary/10">
                <Zap className="h-5 w-5 text-primary" />
              </div>
              <div>
                <MetricLabel>Active Strategies</MetricLabel>
                <MetricValue>{activeCount}</MetricValue>
              </div>
            </div>
          </PageCell>
          <PageCell>
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 flex items-center justify-center bg-surface-container-high">
                <BarChart3 className="h-5 w-5 text-on-surface-variant" />
              </div>
              <div>
                <MetricLabel>Total Trades</MetricLabel>
                <MetricValue>{totalTrades.toLocaleString()}</MetricValue>
              </div>
            </div>
          </PageCell>
          <PageCell>
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 flex items-center justify-center bg-surface-container-high">
                <TrendingUp className="h-5 w-5 text-on-surface-variant" />
              </div>
              <div>
                <MetricLabel>Avg Win Rate</MetricLabel>
                <MetricValue>{avgWinRate}%</MetricValue>
              </div>
            </div>
          </PageCell>
          <PageCell>
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 flex items-center justify-center bg-surface-container-high">
                <Shield className="h-5 w-5 text-on-surface-variant" />
              </div>
              <div>
                <MetricLabel>Avg Sharpe</MetricLabel>
                <MetricValue>{avgSharpe}</MetricValue>
              </div>
            </div>
          </PageCell>
        </PageGrid>

        <PageGrid style={{ gridTemplateColumns: "1fr" }}>
          <PageCell>
            <CardTitle className="mb-4">Strategies</CardTitle>
            <div className="border border-outline-variant/30">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="text-right">Returns</TableHead>
                    <TableHead className="text-right">Sharpe</TableHead>
                    <TableHead className="text-right">Max DD</TableHead>
                    <TableHead className="text-right">Trades</TableHead>
                    <TableHead className="text-right">Win Rate</TableHead>
                    <TableHead className="text-center">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {strategies.map((strategy) => (
                    <TableRow key={strategy.ID}>
                      <TableCell className="font-medium">{strategy.Name}</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <StatusIndicator active={strategy.Status === "active"} />
                          <span className="text-xs capitalize text-on-surface-variant">{strategy.Status}</span>
                        </div>
                      </TableCell>
                      <TableCell className="text-right font-mono text-primary">{fmtPct(strategy.Returns)}</TableCell>
                      <TableCell className="text-right font-mono">{strategy.Sharpe.toFixed(2)}</TableCell>
                      <TableCell className="text-right font-mono text-error">{strategy.MaxDD.toFixed(1)}%</TableCell>
                      <TableCell className="text-right font-mono">{strategy.Trades.toLocaleString()}</TableCell>
                      <TableCell className="text-right font-mono">{strategy.WinRate.toFixed(1)}%</TableCell>
                      <TableCell className="text-center">
                        <Button variant="ghost" size="icon" onClick={() => toggleStrategy(strategy)}>
                          {strategy.Status === "active" ? <Pause className="h-4 w-4" /> : <Play className="h-4 w-4" />}
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                  {strategies.length === 0 && !loading && (
                    <TableRow>
                      <TableCell colSpan={8} className="text-center text-on-surface-variant text-sm py-8">No strategies available</TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          </PageCell>
        </PageGrid>

        <PageGrid className="mt-px" style={{ gridTemplateColumns: "repeat(2, 1fr)" }}>
          <PageCell>
            <CardTitle className="mb-4">Recent Signals</CardTitle>
            <div className="grid grid-cols-2 gap-2">
              {signals.map((signal) => (
                <div key={signal.ID} className="flex items-center justify-between p-4 border border-outline-variant/30">
                  <div className="flex items-center gap-3">
                    <Badge variant={signal.Action === "BUY" ? "success" : "error"}>{signal.Action}</Badge>
                    <div>
                      <div className="font-mono text-sm font-semibold">{signal.Symbol}</div>
                      <div className="text-[11px] text-on-surface-variant">{signal.Strategy}</div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="font-mono text-sm">${signal.Price.toFixed(2)}</div>
                    <div className="text-[11px] text-on-surface-variant">{signal.Confidence}</div>
                  </div>
                </div>
              ))}
              {signals.length === 0 && !loading && (
                <div className="col-span-2 text-center text-on-surface-variant text-sm py-8">No signals available</div>
              )}
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