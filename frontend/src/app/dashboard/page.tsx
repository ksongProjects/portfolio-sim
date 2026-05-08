"use client";

import { useState } from "react";
import { PageCell, PageHeader } from "@/components/page-layout";
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
    <div className="flex flex-col h-full bg-surface">
      <div className="px-6 pt-6">
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
        <div className="grid gap-[1px] bg-outline-variant">
          <div className="grid grid-cols-4 gap-[1px] bg-outline-variant">
            <PageCell>
              <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant mb-2">
                Total Portfolio Value
              </div>
              <div className="text-[28px] font-medium tracking-[-0.02em] text-on-surface" style={{ fontFamily: "var(--font-work-sans)" }}>
                {portfolioSummary.totalValue}
              </div>
              <div className="text-[13px] font-mono text-primary mt-1">
                {portfolioSummary.dayChange} ({portfolioSummary.dayChangePercent})
              </div>
            </PageCell>

            <PageCell>
              <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant mb-2">
                Today&apos;s P&amp;L
              </div>
              <div className="text-[28px] font-medium tracking-[-0.02em] text-primary" style={{ fontFamily: "var(--font-work-sans)" }}>
                {portfolioSummary.dayChange}
              </div>
              <div className="text-[13px] font-mono text-primary mt-1">
                {portfolioSummary.dayChangePercent}
              </div>
            </PageCell>

            <PageCell>
              <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant mb-2">
                Total Gain/Loss
              </div>
              <div className="text-[28px] font-medium tracking-[-0.02em] text-on-surface" style={{ fontFamily: "var(--font-work-sans)" }}>
                {portfolioSummary.totalGain}
              </div>
              <div className="text-[13px] font-mono text-primary mt-1">
                {portfolioSummary.totalGainPercent}
              </div>
            </PageCell>

            <PageCell>
              <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant mb-2">
                Cash Balance
              </div>
              <div className="text-[28px] font-medium tracking-[-0.02em] text-on-surface" style={{ fontFamily: "var(--font-work-sans)" }}>
                {portfolioSummary.totalBalance}
              </div>
              <div className="text-[13px] font-mono text-on-surface-variant mt-1">
                Buying Power
              </div>
            </PageCell>
          </div>

          <div className="grid grid-cols-12 gap-[1px] bg-outline-variant">
            <PageCell className="col-span-7">
              <div className="flex items-center justify-between mb-4">
                <CardTitle className="text-base font-semibold text-on-surface">Portfolio Performance</CardTitle>
                <div className="flex gap-1">
                  {["1D", "1W", "1M", "3M", "1Y", "ALL"].map((period) => (
                    <button
                      key={period}
                      className={`px-2 py-1 text-[11px] font-semibold uppercase tracking-[0.06em] transition-colors ${
                        period === "1M"
                          ? "bg-primary text-on-primary"
                          : "text-on-surface-variant hover:text-on-surface hover:bg-surface-container-high"
                      }`}
                    >
                      {period}
                    </button>
                  ))}
                </div>
              </div>
              <div className="h-[220px] border border-outline-variant">
                <MiniChart />
              </div>
            </PageCell>

            <PageCell className="col-span-5">
              <CardTitle className="text-base font-semibold text-on-surface mb-4">Top Holdings</CardTitle>
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
                        <div className="font-mono font-semibold text-sm text-on-surface">{pos.symbol}</div>
                        <div className="text-[11px] text-on-surface-variant truncate max-w-[120px]">{pos.name}</div>
                      </TableCell>
                      <TableCell className="font-mono text-sm">{pos.price}</TableCell>
                      <TableCell className="font-mono text-sm">{pos.value}</TableCell>
                      <TableCell className={`font-mono text-sm ${pos.change.startsWith("+") ? "text-primary" : "text-error"}`}>
                        {pos.change}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </PageCell>
          </div>

          <div className="grid grid-cols-12 gap-[1px] bg-outline-variant">
            <PageCell className="col-span-5">
              <CardTitle className="text-base font-semibold text-on-surface mb-4">Recent Activity</CardTitle>
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
                        <Badge variant={trade.type === "BUY" ? "success" : "error"} className="text-[10px]">
                          {trade.type}
                        </Badge>
                      </TableCell>
                      <TableCell className="font-mono font-semibold text-sm">{trade.symbol}</TableCell>
                      <TableCell className="font-mono text-sm">{trade.shares}</TableCell>
                      <TableCell className="font-mono text-sm">{trade.price}</TableCell>
                      <TableCell className="font-mono text-sm">{trade.total}</TableCell>
                      <TableCell className="text-sm text-on-surface-variant">{trade.time}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </PageCell>

            <PageCell className="col-span-7">
              <CardTitle className="text-base font-semibold text-on-surface mb-4">Market Indices</CardTitle>
              <div className="grid grid-cols-3 gap-4">
                {[
                  { symbol: "SPY", name: "S&P 500 ETF", price: "$523.45", change: "+0.87%", positive: true },
                  { symbol: "DIA", name: "Dow Jones ETF", price: "$398.20", change: "+0.45%", positive: true },
                  { symbol: "QQQ", name: "Nasdaq ETF", price: "$448.30", change: "+1.12%", positive: true },
                  { symbol: "IWM", name: "Russell 2000 ETF", price: "$198.30", change: "-0.23%", positive: false },
                  { symbol: "VIX", name: "Volatility Index", price: "$14.82", change: "-4.12%", positive: false },
                  { symbol: "DXY", name: "US Dollar Index", price: "$104.20", change: "+0.15%", positive: true },
                ].map((index) => (
                  <div key={index.symbol} className="flex items-center justify-between p-3 border border-outline-variant">
                    <div>
                      <div className="font-mono text-sm font-semibold text-on-surface">{index.symbol}</div>
                      <div className="text-[11px] text-on-surface-variant">{index.name}</div>
                    </div>
                    <div className="text-right">
                      <div className="font-mono text-sm text-on-surface">{index.price}</div>
                      <div className={`text-[11px] font-mono ${index.positive ? "text-primary" : "text-error"}`}>
                        {index.change}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </PageCell>
          </div>
        </div>
      </div>

      {toast && (
        <div className="fixed bottom-6 right-6 z-50 bg-surface-container-highest border border-primary text-on-surface px-4 py-3 text-sm font-medium">
          {toast}
        </div>
      )}
    </div>
  );
}
