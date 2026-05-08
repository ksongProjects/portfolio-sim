"use client";

import { useState } from "react";
import { PageCell, PageHeader } from "@/components/page-layout";
import { CardTitle } from "@/components/ui/card";
import { StatusIndicator } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Plus, Search, Filter, Download, ChevronUp, ChevronDown } from "lucide-react";

type SortKey = "symbol" | "shares" | "avgCost" | "price" | "value" | "dayChange" | "totalGain";

function SortIcon({ col, sortKey, sortAsc }: { col: SortKey; sortKey: SortKey; sortAsc: boolean }) {
  if (sortKey !== col) return <ChevronDown className="h-3 w-3 opacity-30" />;
  return sortAsc ? (
    <ChevronUp className="h-3 w-3 text-primary" />
  ) : (
    <ChevronDown className="h-3 w-3 text-primary" />
  );
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
    .filter(
      (p) =>
        p.symbol.toLowerCase().includes(search.toLowerCase()) ||
        p.name.toLowerCase().includes(search.toLowerCase())
    )
    .sort((a, b) => {
      const aVal = a[sortKey];
      const bVal = b[sortKey];
      if (typeof aVal === "string" && typeof bVal === "string") {
        const aNum = parseFloat(aVal.replace(/[^0-9.-]/g, ""));
        const bNum = parseFloat(bVal.replace(/[^0-9.-]/g, ""));
        return sortAsc ? aNum - bNum : bNum - aNum;
      }
      return 0;
    });

  return (
    <div className="flex flex-col h-full bg-surface">
      <div className="px-6 pt-6">
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
      </div>

      <div className="flex-1 px-6 pb-6 overflow-auto">
        <div className="grid gap-[1px] bg-outline-variant mb-[1px]">
          <div className="grid grid-cols-5 gap-[1px] bg-outline-variant">
            <PageCell>
              <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant mb-2">
                Total Value
              </div>
              <div className="text-[24px] font-medium tracking-[-0.01em] text-on-surface" style={{ fontFamily: "var(--font-work-sans)" }}>
                {portfolioSummary.totalValue}
              </div>
              <div className="text-[13px] font-mono text-primary mt-1">
                {portfolioSummary.dayChangePercent}
              </div>
            </PageCell>
            <PageCell>
              <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant mb-2">
                Day P&amp;L
              </div>
              <div className="text-[24px] font-medium tracking-[-0.01em] text-primary" style={{ fontFamily: "var(--font-work-sans)" }}>
                {portfolioSummary.dayPNL}
              </div>
              <div className="text-[13px] font-mono text-primary mt-1">
                {portfolioSummary.dayChangePercent}
              </div>
            </PageCell>
            <PageCell>
              <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant mb-2">
                Total Invested
              </div>
              <div className="text-[24px] font-medium tracking-[-0.01em] text-on-surface" style={{ fontFamily: "var(--font-work-sans)" }}>
                {portfolioSummary.totalInvested}
              </div>
              <div className="text-[13px] font-mono text-on-surface-variant mt-1">
                Cost Basis
              </div>
            </PageCell>
            <PageCell>
              <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant mb-2">
                Total Gain/Loss
              </div>
              <div className="text-[24px] font-medium tracking-[-0.01em] text-on-surface" style={{ fontFamily: "var(--font-work-sans)" }}>
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
              <div className="text-[24px] font-medium tracking-[-0.01em] text-on-surface" style={{ fontFamily: "var(--font-work-sans)" }}>
                $59,324.40
              </div>
              <div className="text-[13px] font-mono text-on-surface-variant mt-1">
                Buying Power
              </div>
            </PageCell>
          </div>
        </div>

        <div className="grid grid-cols-12 gap-[1px] bg-outline-variant">
          <PageCell className="col-span-8">
            <div className="flex items-center justify-between mb-4">
              <CardTitle className="text-base font-semibold text-on-surface">All Positions</CardTitle>
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
            <div className="border border-outline-variant">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[100px]">
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
                      <TableCell>
                        <div className="font-mono font-semibold text-sm text-on-surface">{pos.symbol}</div>
                      </TableCell>
                      <TableCell>
                        <div className="text-sm text-on-surface">{pos.name}</div>
                        <div className="text-[11px] text-on-surface-variant">{pos.sector}</div>
                      </TableCell>
                      <TableCell className="text-right font-mono text-sm">{pos.shares}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{pos.avgCost}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{pos.price}</TableCell>
                      <TableCell className="text-right font-mono text-sm font-medium">{pos.value}</TableCell>
                      <TableCell className={`text-right font-mono text-sm ${pos.dayChange.startsWith("+") ? "text-primary" : "text-error"}`}>
                        {pos.dayChange}
                      </TableCell>
                      <TableCell className={`text-right font-mono text-sm ${pos.totalGain.startsWith("+") ? "text-primary" : "text-error"}`}>
                        {pos.totalGain}
                      </TableCell>
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

          <PageCell className="col-span-4">
            <CardTitle className="text-base font-semibold text-on-surface mb-4">Sector Allocation</CardTitle>
            <div className="space-y-3">
              {sectorAllocation.map((sector) => (
                <div key={sector.sector}>
                  <div className="flex items-center justify-between mb-1">
                    <div className="flex items-center gap-2">
                      <div className="h-2 w-2" style={{ backgroundColor: sector.color }} />
                      <span className="text-sm text-on-surface">{sector.sector}</span>
                    </div>
                    <div className="flex items-center gap-3">
                      <span className="text-sm font-mono text-on-surface">{sector.value}</span>
                      <span className="text-sm font-mono text-on-surface-variant w-10 text-right">{sector.percent}</span>
                    </div>
                  </div>
                  <div className="h-1 bg-surface-container-high w-full">
                    <div
                      className="h-full transition-all"
                      style={{ width: sector.percent, backgroundColor: sector.color }}
                    />
                  </div>
                </div>
              ))}
            </div>

            <div className="mt-6 pt-6 border-t border-outline-variant">
              <CardTitle className="text-base font-semibold text-on-surface mb-4">Quick Stats</CardTitle>
              <div className="grid grid-cols-2 gap-3">
                <div className="p-3 border border-outline-variant">
                  <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant">
                    Positions
                  </div>
                  <div className="text-[18px] font-mono font-medium text-on-surface mt-1">
                    {positions.length}
                  </div>
                </div>
                <div className="p-3 border border-outline-variant">
                  <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant">
                    Avg Cost Basis
                  </div>
                  <div className="text-[18px] font-mono font-medium text-on-surface mt-1">
                    $243.21
                  </div>
                </div>
                <div className="p-3 border border-outline-variant">
                  <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant">
                    Day Change
                  </div>
                  <div className="text-[18px] font-mono font-medium text-primary mt-1">
                    +$7,234.56
                  </div>
                </div>
                <div className="p-3 border border-outline-variant">
                  <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant">
                    Wtd. Avg Gain
                  </div>
                  <div className="text-[18px] font-mono font-medium text-primary mt-1">
                    +18.24%
                  </div>
                </div>
              </div>
            </div>
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
