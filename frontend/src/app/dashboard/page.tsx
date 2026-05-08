"use client";

import { useState } from "react";
import { PageCell, PageHeader, MetricCard } from "@/components/page-layout";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Activity, BarChart3 } from "lucide-react";

const portfolioSummary = {
  totalValue: "$1,247,832.40",
  dayChange: "+$12,458.20",
  dayChangePercent: "+1.01%",
  totalGain: "+$247,832.40",
  totalGainPercent: "+24.78%",
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
    <div className="flex flex-col h-full">
      <PageHeader
        title="Dashboard"
        description="Portfolio overview and performance metrics"
      >
        <Button variant="default" size="sm" onClick={handleRunSimulation} disabled={simRunning}>
          <Activity className="h-4 w-4" />
          {simRunning ? "Running..." : "Run Simulation"}
        </Button>
      </PageHeader>

      <div className="flex-1 p-6">
        <div className="grid gap-[1px] bg-outline-variant p-[1px]" style={{ gridTemplateColumns: "1fr 1fr 1fr", gridTemplateRows: "auto 1fr 1fr" }}>
          <PageCell className="col-span-3">
            <div className="grid grid-cols-4 gap-6">
              <MetricCard
                label="Total Portfolio Value"
                value={portfolioSummary.totalValue}
                change={portfolioSummary.dayChangePercent}
                positive={true}
              />
              <MetricCard
                label="Today's Change"
                value={portfolioSummary.dayChange}
                change={portfolioSummary.dayChangePercent}
                positive={true}
              />
              <MetricCard
                label="Total Gain/Loss"
                value={portfolioSummary.totalGain}
                change={portfolioSummary.totalGainPercent}
                positive={true}
              />
              <MetricCard
                label="Cash Available"
                value="$24,582.40"
              />
            </div>
          </PageCell>

          <PageCell className="col-span-2 row-span-2">
            <Card className="h-full border-0 bg-transparent">
              <CardHeader className="px-0">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base font-semibold">Portfolio Performance</CardTitle>
                  <div className="flex gap-2">
                    <Badge variant="secondary">1D</Badge>
                    <Badge variant="outline">1W</Badge>
                    <Badge variant="outline">1M</Badge>
                    <Badge variant="outline">1Y</Badge>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="px-0 pt-4">
                <div className="flex items-center justify-center h-[280px] border border-outline-variant">
                  <div className="text-center text-on-surface-variant">
                    <BarChart3 className="h-12 w-12 mx-auto mb-2 opacity-50" />
                    <p className="text-sm">Chart visualization</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </PageCell>

          <PageCell className="row-span-2">
            <Card className="h-full border-0 bg-transparent">
              <CardHeader className="px-0">
                <CardTitle className="text-base font-semibold">Top Holdings</CardTitle>
              </CardHeader>
              <CardContent className="px-0 pt-4">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Symbol</TableHead>
                      <TableHead>Price</TableHead>
                      <TableHead>Change</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {positions.slice(0, 4).map((pos) => (
                      <TableRow key={pos.id}>
                        <TableCell className="font-mono font-medium">{pos.symbol}</TableCell>
                        <TableCell className="font-mono">{pos.price}</TableCell>
                        <TableCell className={pos.change.startsWith("+") ? "text-primary" : "text-error"}>
                          {pos.change}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          </PageCell>

          <PageCell className="row-span-2">
            <Card className="h-full border-0 bg-transparent">
              <CardHeader className="px-0">
                <CardTitle className="text-base font-semibold">Recent Activity</CardTitle>
              </CardHeader>
              <CardContent className="px-0 pt-4">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Type</TableHead>
                      <TableHead>Symbol</TableHead>
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
                        <TableCell className="font-mono font-medium">{trade.symbol}</TableCell>
                        <TableCell className="text-on-surface-variant">{trade.time}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          </PageCell>
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
