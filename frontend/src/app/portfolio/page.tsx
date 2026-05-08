"use client";

import { useState } from "react";
import { PageCell, PageHeader, MetricCard } from "@/components/page-layout";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { StatusIndicator } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Plus, Search, Filter, Download } from "lucide-react";

const portfolioSummary = {
  totalValue: "$1,247,832.40",
  dayChange: "+$12,458.20",
  dayChangePercent: "+1.01%",
  totalInvested: "$1,000,000.00",
  totalGain: "+$247,832.40",
  totalGainPercent: "+24.78%",
  dayPNL: "+$12,458.20",
};

const positions = [
  { id: 1, symbol: "NVDA", name: "NVIDIA Corporation", sector: "Technology", shares: 150, avgCost: "$720.00", price: "$875.28", value: "$131,292.00", dayChange: "+2.34%", totalGain: "+21.56%", status: "active" },
  { id: 2, symbol: "AAPL", name: "Apple Inc.", sector: "Technology", shares: 200, avgCost: "$165.00", price: "$178.42", value: "$35,684.00", dayChange: "+0.87%", totalGain: "+8.13%", status: "active" },
  { id: 3, symbol: "MSFT", name: "Microsoft Corporation", sector: "Technology", shares: 100, avgCost: "$380.00", price: "$415.20", value: "$41,520.00", dayChange: "+1.12%", totalGain: "+9.26%", status: "active" },
  { id: 4, symbol: "GOOGL", name: "Alphabet Inc.", sector: "Technology", shares: 80, avgCost: "$140.00", price: "$175.30", value: "$14,024.00", dayChange: "-0.45%", totalGain: "+25.21%", status: "active" },
  { id: 5, symbol: "TSLA", name: "Tesla Inc.", sector: "Consumer", shares: 60, avgCost: "$195.00", price: "$245.80", value: "$14,748.00", dayChange: "+3.21%", totalGain: "+26.05%", status: "active" },
  { id: 6, symbol: "JPM", name: "JPMorgan Chase", sector: "Financial", shares: 75, avgCost: "$155.00", price: "$198.50", value: "$14,887.50", dayChange: "+0.32%", totalGain: "+28.06%", status: "active" },
];

const sectorAllocation = [
  { sector: "Technology", value: "$232,520.00", percent: "72.4%" },
  { sector: "Financial", value: "$14,887.50", percent: "4.6%" },
  { sector: "Consumer", value: "$14,748.00", percent: "4.6%" },
  { sector: "Cash", value: "$59,324.40", percent: "18.4%" },
];

export default function PortfolioPage() {
  const [search, setSearch] = useState("");
  const [toast, setToast] = useState<string | null>(null);

  const showToast = (msg: string) => {
    setToast(msg);
    setTimeout(() => setToast(null), 3000);
  };

  const filteredPositions = positions.filter(
    (p) =>
      p.symbol.toLowerCase().includes(search.toLowerCase()) ||
      p.name.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="flex flex-col h-full">
      <PageHeader
        title="Portfolio"
        description="Holdings and position management"
      >
        <div className="flex gap-3">
          <Button
            variant="secondary"
            size="sm"
            onClick={() => showToast("Exporting portfolio as CSV...")}
          >
            <Download className="h-4 w-4" />
            Export
          </Button>
          <Button
            variant="default"
            size="sm"
            onClick={() => showToast("Add position modal coming soon...")}
          >
            <Plus className="h-4 w-4" />
            Add Position
          </Button>
        </div>
      </PageHeader>

      <div className="flex-1 p-6">
        <div className="flex flex-col gap-[1px] bg-outline-variant p-[1px]">
          <PageCell>
            <div className="grid grid-cols-5 gap-6">
              <MetricCard label="Total Value" value={portfolioSummary.totalValue} change={portfolioSummary.dayChangePercent} positive={true} />
              <MetricCard label="Day P&L" value={portfolioSummary.dayPNL} change={portfolioSummary.dayChangePercent} positive={true} />
              <MetricCard label="Total Invested" value={portfolioSummary.totalInvested} />
              <MetricCard label="Total Gain/Loss" value={portfolioSummary.totalGain} change={portfolioSummary.totalGainPercent} positive={true} />
              <MetricCard label="Cash Balance" value="$59,324.40" />
            </div>
          </PageCell>
        </div>

        <div className="mt-[1px] flex gap-[1px] bg-outline-variant p-[1px]">
          <PageCell className="flex-[2]">
            <Card className="h-full border-0 bg-transparent">
              <CardHeader className="px-0 pb-4">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base font-semibold">All Positions</CardTitle>
                  <div className="flex gap-3">
                    <div className="relative">
                      <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-on-surface-variant" />
                      <Input
                        placeholder="Search holdings..."
                        className="pl-9 w-[200px]"
                        value={search}
                        onChange={(e) => setSearch(e.target.value)}
                      />
                    </div>
                    <Button variant="ghost" size="icon" onClick={() => showToast("Filter options coming soon...")}>
                      <Filter className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="px-0">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Symbol</TableHead>
                      <TableHead>Name</TableHead>
                      <TableHead>Shares</TableHead>
                      <TableHead>Avg Cost</TableHead>
                      <TableHead>Price</TableHead>
                      <TableHead>Value</TableHead>
                      <TableHead>Day Change</TableHead>
                      <TableHead>Total Gain</TableHead>
                      <TableHead>Status</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredPositions.map((pos) => (
                      <TableRow key={pos.id}>
                        <TableCell className="font-mono font-medium">{pos.symbol}</TableCell>
                        <TableCell className="text-on-surface-variant">{pos.name}</TableCell>
                        <TableCell className="font-mono">{pos.shares}</TableCell>
                        <TableCell className="font-mono">{pos.avgCost}</TableCell>
                        <TableCell className="font-mono">{pos.price}</TableCell>
                        <TableCell className="font-mono">{pos.value}</TableCell>
                        <TableCell className={pos.dayChange.startsWith("+") ? "text-primary" : "text-error"}>
                          {pos.dayChange}
                        </TableCell>
                        <TableCell className={pos.totalGain.startsWith("+") ? "text-primary" : "text-error"}>
                          {pos.totalGain}
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <StatusIndicator active={pos.status === "active"} />
                            <span className="text-xs capitalize">{pos.status}</span>
                          </div>
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
                <CardTitle className="text-base font-semibold">Sector Allocation</CardTitle>
              </CardHeader>
              <CardContent className="px-0">
                <div className="space-y-4">
                  {sectorAllocation.map((sector) => (
                    <div key={sector.sector} className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <div className="h-2 w-2 bg-primary" />
                        <span className="text-sm text-on-surface">{sector.sector}</span>
                      </div>
                      <div className="text-right">
                        <div className="text-sm font-mono text-on-surface">{sector.value}</div>
                        <div className="text-xs text-on-surface-variant">{sector.percent}</div>
                      </div>
                    </div>
                  ))}
                </div>
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
