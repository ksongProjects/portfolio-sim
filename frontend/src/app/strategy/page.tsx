"use client";

import { useState } from "react";
import { PageGrid, PageCell, PageHeader, MetricLabel, MetricValue } from "@/components/page-layout";
import { CardTitle } from "@/components/ui/card";
import { Badge, StatusIndicator } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Play, Pause, Plus, Settings, Zap, BarChart3, TrendingUp, Shield } from "lucide-react";

const strategies = [
  { id: 1, name: "Momentum Growth", status: "active" as const, returns: "+18.4%", sharpe: "1.42", maxDD: "-12.3%", trades: 847, winRate: "64.2%" },
  { id: 2, name: "Value Scanner", status: "active" as const, returns: "+12.1%", sharpe: "1.18", maxDD: "-8.7%", trades: 423, winRate: "58.9%" },
  { id: 3, name: "Mean Reversion", status: "paused" as const, returns: "+8.3%", sharpe: "0.95", maxDD: "-15.2%", trades: 612, winRate: "52.1%" },
  { id: 4, name: "Sector Rotation", status: "active" as const, returns: "+15.7%", sharpe: "1.28", maxDD: "-10.1%", trades: 234, winRate: "61.4%" },
];

const recentSignals = [
  { id: 1, strategy: "Momentum Growth", symbol: "NVDA", action: "BUY", price: "$875.28", time: "10:32 AM", confidence: "HIGH" },
  { id: 2, strategy: "Value Scanner", symbol: "JPM", action: "BUY", price: "$198.50", time: "09:15 AM", confidence: "MEDIUM" },
  { id: 3, strategy: "Sector Rotation", symbol: "TLT", action: "SELL", price: "$92.45", time: "Yesterday", confidence: "HIGH" },
  { id: 4, strategy: "Momentum Growth", symbol: "TSLA", action: "BUY", price: "$245.80", time: "Yesterday", confidence: "MEDIUM" },
];

const avgWinRate = (strategies.reduce((sum, s) => sum + parseFloat(s.winRate), 0) / strategies.length).toFixed(1);
const avgSharpe = (strategies.reduce((sum, s) => sum + parseFloat(s.sharpe), 0) / strategies.length).toFixed(2);

export default function StrategyPage() {
  const [strategyList, setStrategyList] = useState([...strategies]);
  const [toast, setToast] = useState<string | null>(null);

  const showToast = (msg: string) => {
    setToast(msg);
    setTimeout(() => setToast(null), 3000);
  };

  const toggleStrategy = (id: number) => {
    setStrategyList((prev) =>
      prev.map((s) => s.id === id ? { ...s, status: s.status === "active" ? ("paused" as const) : ("active" as const) } : s)
    );
    const s = strategyList.find((str) => str.id === id);
    showToast(`${s?.name} ${s?.status === "active" ? "paused" : "started"}`);
  };

  const totalTrades = strategyList.reduce((sum, s) => sum + s.trades, 0);
  const activeCount = strategyList.filter((s) => s.status === "active").length;

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
                  {strategyList.map((strategy) => (
                    <TableRow key={strategy.id}>
                      <TableCell className="font-medium">{strategy.name}</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <StatusIndicator active={strategy.status === "active"} />
                          <span className="text-xs capitalize text-on-surface-variant">{strategy.status}</span>
                        </div>
                      </TableCell>
                      <TableCell className="text-right font-mono text-primary">{strategy.returns}</TableCell>
                      <TableCell className="text-right font-mono">{strategy.sharpe}</TableCell>
                      <TableCell className="text-right font-mono text-error">{strategy.maxDD}</TableCell>
                      <TableCell className="text-right font-mono">{strategy.trades.toLocaleString()}</TableCell>
                      <TableCell className="text-right font-mono">{strategy.winRate}</TableCell>
                      <TableCell className="text-center">
                        <Button variant="ghost" size="icon" onClick={() => toggleStrategy(strategy.id)}>
                          {strategy.status === "active" ? <Pause className="h-4 w-4" /> : <Play className="h-4 w-4" />}
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </PageCell>
        </PageGrid>

        <PageGrid className="mt-px" style={{ gridTemplateColumns: "repeat(2, 1fr)" }}>
          <PageCell>
            <CardTitle className="mb-4">Recent Signals</CardTitle>
            <div className="grid grid-cols-2 gap-2">
              {recentSignals.map((signal) => (
                <div key={signal.id} className="flex items-center justify-between p-4 border border-outline-variant/30">
                  <div className="flex items-center gap-3">
                    <Badge variant={signal.action === "BUY" ? "success" : "error"}>{signal.action}</Badge>
                    <div>
                      <div className="font-mono text-sm font-semibold">{signal.symbol}</div>
                      <div className="text-[11px] text-on-surface-variant">{signal.strategy}</div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="font-mono text-sm">{signal.price}</div>
                    <div className="text-[11px] text-on-surface-variant">{signal.time}</div>
                  </div>
                </div>
              ))}
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
