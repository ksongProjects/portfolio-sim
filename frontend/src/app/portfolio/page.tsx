"use client";

import { useState } from "react";
import { PageGrid, PageCell, PageHeader, MetricLabel, MetricValue, MetricSubValue } from "@/components/page-layout";
import { CardTitle } from "@/components/ui/card";
import { StatusIndicator } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Plus, Search, Filter, Download, ChevronUp, ChevronDown } from "lucide-react";

type SortKey = "symbol" | "shares" | "avgCost" | "price" | "value" | "dayChange" | "totalGain";

function SortIcon({ col, sortKey, sortAsc }: { col: SortKey; sortKey: SortKey; sortAsc: boolean }) {
  if (sortKey !== col) return <ChevronDown className="h-3 w-3 opacity-30" />;
  return sortAsc ? <ChevronUp className="h-3 w-3 text-primary" /> : <ChevronDown className="h-3 w-3 text-primary" />;
}

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
  { sector: "Technology", value: "$232,520.00", percent: "72.4%", color: "#3fe56c" },
  { sector: "Financial", value: "$14,887.50", percent: "4.6%", color: "#869583" },
  { sector: "Consumer", value: "$14,748.00", percent: "4.6%", color: "#c6c6c7" },
  { sector: "Cash", value: "$59,324.40", percent: "18.4%", color: "#454747" },
];

export default function PortfolioPage() {
  const [search, setSearch] = useState("");
  const [sortKey, setSortKey] = useState<SortKey>("value");
  const [sortAsc, setSortAsc] = useState(false);
  const [toast, setToast] = useState<string | null>(null);

  const showToast = (msg: string) => {
    setToast(msg);
    setTimeout(() => setToast(null), 3000);
  };

  const handleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortAsc(!sortAsc);
    } else {
      setSortKey(key);
      setSortAsc(false);
    }
  };

  const filteredPositions = positions
    .filter((p) => p.symbol.toLowerCase().includes(search.toLowerCase()) || p.name.toLowerCase().includes(search.toLowerCase()))
    .sort((a, b) => {
      const aVal = a[sortKey] as string;
      const bVal = b[sortKey] as string;
      const aNum = parseFloat(aVal.replace(/[^0-9.-]/g, ""));
      const bNum = parseFloat(bVal.replace(/[^0-9.-]/g, ""));
      return sortAsc ? aNum - bNum : bNum - aNum;
    });

  return (
    <div className="flex flex-col h-full">
      <div className="px-6 pt-6 pb-4">
        <PageHeader title="Portfolio" description="Holdings and position management">
          <div className="flex gap-3">
            <Button variant="secondary" size="sm" onClick={() => showToast("Exporting portfolio as CSV...")}>
              <Download className="h-4 w-4" /> Export
            </Button>
            <Button variant="default" size="sm" onClick={() => showToast("Add position modal coming soon...")}>
              <Plus className="h-4 w-4" /> Add Position
            </Button>
          </div>
        </PageHeader>
      </div>

      <div className="flex-1 px-6 pb-6 overflow-auto">
        <PageGrid className="mb-4" style={{ gridTemplateColumns: "repeat(5, 1fr)" }}>
          <PageCell>
            <MetricLabel>Total Value</MetricLabel>
            <MetricValue>{portfolioSummary.totalValue}</MetricValue>
            <MetricSubValue positive>{portfolioSummary.dayChangePercent}</MetricSubValue>
          </PageCell>
          <PageCell>
            <MetricLabel>Day P&amp;L</MetricLabel>
            <MetricValue highlight>{portfolioSummary.dayPNL}</MetricValue>
            <MetricSubValue positive>{portfolioSummary.dayChangePercent}</MetricSubValue>
          </PageCell>
          <PageCell>
            <MetricLabel>Total Invested</MetricLabel>
            <MetricValue>{portfolioSummary.totalInvested}</MetricValue>
            <MetricSubValue>Cost Basis</MetricSubValue>
          </PageCell>
          <PageCell>
            <MetricLabel>Total Gain/Loss</MetricLabel>
            <MetricValue>{portfolioSummary.totalGain}</MetricValue>
            <MetricSubValue positive>{portfolioSummary.totalGainPercent}</MetricSubValue>
          </PageCell>
          <PageCell>
            <MetricLabel>Cash Balance</MetricLabel>
            <MetricValue>$59,324.40</MetricValue>
            <MetricSubValue>Buying Power</MetricSubValue>
          </PageCell>
        </PageGrid>

        <PageGrid style={{ gridTemplateColumns: "2fr 1fr" }}>
          <PageCell>
            <div className="flex items-center justify-between mb-4">
              <CardTitle>All Positions</CardTitle>
              <div className="flex gap-3">
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-on-surface-variant" />
                  <Input placeholder="Search holdings..." className="pl-9 w-[200px]" value={search} onChange={(e) => setSearch(e.target.value)} />
                </div>
                <Button variant="ghost" size="icon" onClick={() => showToast("Filter options coming soon...")}>
                  <Filter className="h-4 w-4" />
                </Button>
              </div>
            </div>
            <div className="border border-outline-variant/30">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[90px]">
                      <button onClick={() => handleSort("symbol")} className="flex items-center gap-1 hover:text-on-surface">
                        Symbol <SortIcon col="symbol" sortKey={sortKey} sortAsc={sortAsc} />
                      </button>
                    </TableHead>
                    <TableHead>Name</TableHead>
                    <TableHead className="text-right">
                      <button onClick={() => handleSort("shares")} className="flex items-center gap-1 ml-auto hover:text-on-surface">
                        Shares <SortIcon col="shares" sortKey={sortKey} sortAsc={sortAsc} />
                      </button>
                    </TableHead>
                    <TableHead className="text-right">
                      <button onClick={() => handleSort("avgCost")} className="flex items-center gap-1 ml-auto hover:text-on-surface">
                        Avg Cost <SortIcon col="avgCost" sortKey={sortKey} sortAsc={sortAsc} />
                      </button>
                    </TableHead>
                    <TableHead className="text-right">
                      <button onClick={() => handleSort("price")} className="flex items-center gap-1 ml-auto hover:text-on-surface">
                        Price <SortIcon col="price" sortKey={sortKey} sortAsc={sortAsc} />
                      </button>
                    </TableHead>
                    <TableHead className="text-right">
                      <button onClick={() => handleSort("value")} className="flex items-center gap-1 ml-auto hover:text-on-surface">
                        Value <SortIcon col="value" sortKey={sortKey} sortAsc={sortAsc} />
                      </button>
                    </TableHead>
                    <TableHead className="text-right">
                      <button onClick={() => handleSort("dayChange")} className="flex items-center gap-1 ml-auto hover:text-on-surface">
                        Day <SortIcon col="dayChange" sortKey={sortKey} sortAsc={sortAsc} />
                      </button>
                    </TableHead>
                    <TableHead className="text-right">
                      <button onClick={() => handleSort("totalGain")} className="flex items-center gap-1 ml-auto hover:text-on-surface">
                        Total <SortIcon col="totalGain" sortKey={sortKey} sortAsc={sortAsc} />
                      </button>
                    </TableHead>
                    <TableHead className="text-center">Status</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredPositions.map((pos) => (
                    <TableRow key={pos.id}>
                      <TableCell><span className="font-mono font-semibold">{pos.symbol}</span></TableCell>
                      <TableCell>
                        <div className="text-sm">{pos.name}</div>
                        <div className="text-[11px] text-on-surface-variant">{pos.sector}</div>
                      </TableCell>
                      <TableCell className="text-right font-mono">{pos.shares}</TableCell>
                      <TableCell className="text-right font-mono">{pos.avgCost}</TableCell>
                      <TableCell className="text-right font-mono">{pos.price}</TableCell>
                      <TableCell className="text-right font-mono font-medium">{pos.value}</TableCell>
                      <TableCell className={`text-right font-mono ${pos.dayChange.startsWith("+") ? "text-primary" : "text-error"}`}>{pos.dayChange}</TableCell>
                      <TableCell className={`text-right font-mono ${pos.totalGain.startsWith("+") ? "text-primary" : "text-error"}`}>{pos.totalGain}</TableCell>
                      <TableCell className="text-center">
                        <div className="flex items-center justify-center gap-1.5">
                          <StatusIndicator active={pos.status === "active"} />
                          <span className="text-[11px] text-on-surface-variant capitalize">{pos.status}</span>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </PageCell>

          <PageCell>
            <CardTitle className="mb-4">Sector Allocation</CardTitle>
            <div className="space-y-3">
              {sectorAllocation.map((sector) => (
                <div key={sector.sector}>
                  <div className="flex items-center justify-between mb-1">
                    <div className="flex items-center gap-2">
                      <div className="h-2 w-2" style={{ backgroundColor: sector.color }} />
                      <span className="text-sm">{sector.sector}</span>
                    </div>
                    <div className="flex items-center gap-3">
                      <span className="text-sm font-mono">{sector.value}</span>
                      <span className="text-sm font-mono text-on-surface-variant w-10 text-right">{sector.percent}</span>
                    </div>
                  </div>
                  <div className="h-0.5 bg-surface-container-high w-full">
                    <div className="h-full transition-all" style={{ width: sector.percent, backgroundColor: sector.color }} />
                  </div>
                </div>
              ))}
            </div>

            <div className="mt-6 pt-5 border-t border-outline-variant/30">
              <CardTitle className="mb-4">Quick Stats</CardTitle>
              <div className="grid grid-cols-2 gap-2">
                {[
                  { label: "Positions", value: String(positions.length) },
                  { label: "Avg Cost Basis", value: "$243.21" },
                  { label: "Day Change", value: "+$7,234.56", highlight: true },
                  { label: "Wtd. Avg Gain", value: "+18.24%", highlight: true },
                ].map((stat) => (
                  <div key={stat.label} className="p-3 border border-outline-variant/30">
                    <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant">{stat.label}</div>
                    <div className={`text-[16px] font-mono font-medium mt-1 ${stat.highlight ? "text-primary" : "text-on-surface"}`}>{stat.value}</div>
                  </div>
                ))}
              </div>
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
