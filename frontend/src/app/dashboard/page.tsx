"use client";

import { useState } from "react";
import { PageGrid, PageCell, PageHeader, MetricLabel, MetricValue, MetricSubValue } from "@/components/page-layout";
import { CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Activity } from "lucide-react";

const portfolioSummary = {
  totalValue: "$1,247,832.40",
  dayChange: "+$12,458.20",
  dayChangePercent: "+1.01%",
  totalGain: "+$247,832.40",
  totalGainPercent: "+24.78%",
  totalBalance: "$24,582.40",
};

const positions = [
  { id: 1, symbol: "NVDA", name: "NVIDIA Corporation", shares: 150, price: "$875.28", value: "$131,292.00", change: "+2.34%" },
  { id: 2, symbol: "AAPL", name: "Apple Inc.", shares: 200, price: "$178.42", value: "$35,684.00", change: "+0.87%" },
  { id: 3, symbol: "MSFT", name: "Microsoft Corporation", shares: 100, price: "$415.20", value: "$41,520.00", change: "+1.12%" },
  { id: 4, symbol: "GOOGL", name: "Alphabet Inc.", shares: 80, price: "$175.30", value: "$14,024.00", change: "-0.45%" },
  { id: 5, symbol: "TSLA", name: "Tesla Inc.", shares: 60, price: "$245.80", value: "$14,748.00", change: "+3.21%" },
];

const recentTrades = [
  { id: 1, type: "BUY", symbol: "NVDA", shares: 50, price: "$860.00", total: "$43,000.00", time: "10:32 AM" },
  { id: 2, type: "SELL", symbol: "AAPL", shares: 25, price: "$177.50", total: "$4,437.50", time: "09:15 AM" },
  { id: 3, type: "BUY", symbol: "TSLA", shares: 30, price: "$240.00", total: "$7,200.00", time: "Yesterday" },
  { id: 4, type: "BUY", symbol: "MSFT", shares: 20, price: "$410.00", total: "$8,200.00", time: "Yesterday" },
];

const marketIndices = [
  { symbol: "SPY", name: "S&P 500 ETF", price: "$523.45", change: "+0.87%", positive: true },
  { symbol: "DIA", name: "Dow Jones ETF", price: "$398.20", change: "+0.45%", positive: true },
  { symbol: "QQQ", name: "Nasdaq ETF", price: "$448.30", change: "+1.12%", positive: true },
  { symbol: "IWM", name: "Russell 2000 ETF", price: "$198.30", change: "-0.23%", positive: false },
  { symbol: "VIX", name: "Volatility Index", price: "$14.82", change: "-4.12%", positive: false },
  { symbol: "DXY", name: "US Dollar Index", price: "$104.20", change: "+0.15%", positive: true },
];

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
    showToast("Simulation completed. +1.01% return.");
  };

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
        <PageGrid className="mb-4" style={{ gridTemplateColumns: "repeat(4, 1fr)" }}>
          <PageCell>
            <MetricLabel>Total Portfolio Value</MetricLabel>
            <MetricValue>{portfolioSummary.totalValue}</MetricValue>
            <MetricSubValue positive>{portfolioSummary.dayChange} ({portfolioSummary.dayChangePercent})</MetricSubValue>
          </PageCell>
          <PageCell>
            <MetricLabel>Today&apos;s P&amp;L</MetricLabel>
            <MetricValue highlight>{portfolioSummary.dayChange}</MetricValue>
            <MetricSubValue positive>{portfolioSummary.dayChangePercent}</MetricSubValue>
          </PageCell>
          <PageCell>
            <MetricLabel>Total Gain/Loss</MetricLabel>
            <MetricValue>{portfolioSummary.totalGain}</MetricValue>
            <MetricSubValue positive>{portfolioSummary.totalGainPercent}</MetricSubValue>
          </PageCell>
          <PageCell>
            <MetricLabel>Cash Balance</MetricLabel>
            <MetricValue>{portfolioSummary.totalBalance}</MetricValue>
            <MetricSubValue>Buying Power</MetricSubValue>
          </PageCell>
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
              <MiniChart />
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
                {positions.map((pos) => (
                  <TableRow key={pos.id}>
                    <TableCell>
                      <span className="font-mono font-semibold">{pos.symbol}</span>
                    </TableCell>
                    <TableCell className="font-mono">{pos.price}</TableCell>
                    <TableCell className="font-mono">{pos.value}</TableCell>
                    <TableCell className={`font-mono ${pos.change.startsWith("+") ? "text-primary" : "text-error"}`}>
                      {pos.change}
                    </TableCell>
                  </TableRow>
                ))}
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
                {recentTrades.map((trade) => (
                  <TableRow key={trade.id}>
                    <TableCell>
                      <Badge variant={trade.type === "BUY" ? "success" : "error"}>
                        {trade.type}
                      </Badge>
                    </TableCell>
                    <TableCell className="font-mono font-semibold">{trade.symbol}</TableCell>
                    <TableCell className="font-mono">{trade.shares}</TableCell>
                    <TableCell className="font-mono">{trade.price}</TableCell>
                    <TableCell className="font-mono">{trade.total}</TableCell>
                    <TableCell className="text-on-surface-variant">{trade.time}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </PageCell>

          <PageCell className="col-span-1 row-span-1">
            <CardTitle className="mb-4">Market Indices</CardTitle>
            <div className="grid grid-cols-2 gap-2">
              {marketIndices.map((index) => (
                <div key={index.symbol} className="flex items-center justify-between p-3 border border-outline-variant/30">
                  <div>
                    <div className="font-mono text-sm font-semibold">{index.symbol}</div>
                    <div className="text-[11px] text-on-surface-variant">{index.name}</div>
                  </div>
                  <div className="text-right">
                    <div className="font-mono text-sm">{index.price}</div>
                    <div className={`text-[11px] font-mono ${index.positive ? "text-primary" : "text-error"}`}>
                      {index.change}
                    </div>
                  </div>
                </div>
              ))}
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
