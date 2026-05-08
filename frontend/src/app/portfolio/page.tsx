"use client";

import { useEffect, useState } from "react";
import { PageGrid, PageCell, PageHeader, MetricLabel, MetricValue, MetricSubValue } from "@/components/page-layout";
import { CardTitle } from "@/components/ui/card";
import { StatusIndicator } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Plus, Search, Filter, Download, ChevronUp, ChevronDown } from "lucide-react";
import { usePortfolio } from "@/hooks/usePortfolio";

type SortKey = "Symbol" | "Quantity" | "AvgCost" | "CurrentPrice" | "CurrentValue" | "DayChangePct" | "TotalGainPct";

function SortIcon({ col, sortKey, sortAsc }: { col: SortKey; sortKey: SortKey; sortAsc: boolean }) {
  if (sortKey !== col) return <ChevronDown className="h-3 w-3 opacity-30" />;
  return sortAsc ? <ChevronUp className="h-3 w-3 text-primary" /> : <ChevronDown className="h-3 w-3 text-primary" />;
}

function fmtCurrency(v: number): string {
  return new Intl.NumberFormat("en-US", { style: "currency", currency: "USD" }).format(v);
}
function fmtPct(v: number): string {
  return (v >= 0 ? "+" : "") + v.toFixed(2) + "%";
}

export default function PortfolioPage() {
  const { positions, summary, loading, refresh } = usePortfolio();
  const [search, setSearch] = useState("");
  const [sortKey, setSortKey] = useState<SortKey>("CurrentValue");
  const [sortAsc, setSortAsc] = useState(false);
  const [toast, setToast] = useState<string | null>(null);

  useEffect(() => { refresh(); }, [refresh]);

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
    .filter((p) => p.Symbol.toLowerCase().includes(search.toLowerCase()) || p.CompanyName.toLowerCase().includes(search.toLowerCase()))
    .sort((a, b) => {
      const aVal = (a[sortKey] as number) || 0;
      const bVal = (b[sortKey] as number) || 0;
      return sortAsc ? aVal - bVal : bVal - aVal;
    });

  const sectors = Array.from(
    positions.reduce((acc, pos) => {
      const sector = "Sector";
      const val = pos.CurrentValue;
      acc.set(sector, (acc.get(sector) || 0) + val);
      return acc;
    }, new Map<string, number>())
  ).map(([sector, value]) => ({
    sector,
    value,
    percent: positions.reduce((s, p) => s + p.CurrentValue, 0) > 0 ? ((value / positions.reduce((s, p) => s + p.CurrentValue, 0)) * 100).toFixed(1) + "%" : "0%",
    color: "#3fe56c",
  }));

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
            <MetricValue>{fmtCurrency(summary?.TotalValue ?? 0)}</MetricValue>
            <MetricSubValue positive>{fmtPct(summary?.DayChangePct ?? 0)}</MetricSubValue>
          </PageCell>
          <PageCell>
            <MetricLabel>Day P&amp;L</MetricLabel>
            <MetricValue highlight>{fmtCurrency(summary?.DayChange ?? 0)}</MetricValue>
            <MetricSubValue positive>{fmtPct(summary?.DayChangePct ?? 0)}</MetricSubValue>
          </PageCell>
          <PageCell>
            <MetricLabel>Total Invested</MetricLabel>
            <MetricValue>{fmtCurrency(summary?.TotalInvested ?? 0)}</MetricValue>
            <MetricSubValue>Cost Basis</MetricSubValue>
          </PageCell>
          <PageCell>
            <MetricLabel>Total Gain/Loss</MetricLabel>
            <MetricValue>{fmtCurrency(summary?.TotalGain ?? 0)}</MetricValue>
            <MetricSubValue positive>{fmtPct(summary?.TotalGainPct ?? 0)}</MetricSubValue>
          </PageCell>
          <PageCell>
            <MetricLabel>Cash Balance</MetricLabel>
            <MetricValue>{fmtCurrency(summary?.CashBalance ?? 0)}</MetricValue>
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
                      <button onClick={() => handleSort("Symbol")} className="flex items-center gap-1 hover:text-on-surface">
                        Symbol <SortIcon col="Symbol" sortKey={sortKey} sortAsc={sortAsc} />
                      </button>
                    </TableHead>
                    <TableHead>Name</TableHead>
                    <TableHead className="text-right">
                      <button onClick={() => handleSort("Quantity")} className="flex items-center gap-1 ml-auto hover:text-on-surface">
                        Shares <SortIcon col="Quantity" sortKey={sortKey} sortAsc={sortAsc} />
                      </button>
                    </TableHead>
                    <TableHead className="text-right">
                      <button onClick={() => handleSort("AvgCost")} className="flex items-center gap-1 ml-auto hover:text-on-surface">
                        Avg Cost <SortIcon col="AvgCost" sortKey={sortKey} sortAsc={sortAsc} />
                      </button>
                    </TableHead>
                    <TableHead className="text-right">
                      <button onClick={() => handleSort("CurrentPrice")} className="flex items-center gap-1 ml-auto hover:text-on-surface">
                        Price <SortIcon col="CurrentPrice" sortKey={sortKey} sortAsc={sortAsc} />
                      </button>
                    </TableHead>
                    <TableHead className="text-right">
                      <button onClick={() => handleSort("CurrentValue")} className="flex items-center gap-1 ml-auto hover:text-on-surface">
                        Value <SortIcon col="CurrentValue" sortKey={sortKey} sortAsc={sortAsc} />
                      </button>
                    </TableHead>
                    <TableHead className="text-right">
                      <button onClick={() => handleSort("DayChangePct")} className="flex items-center gap-1 ml-auto hover:text-on-surface">
                        Day <SortIcon col="DayChangePct" sortKey={sortKey} sortAsc={sortAsc} />
                      </button>
                    </TableHead>
                    <TableHead className="text-right">
                      <button onClick={() => handleSort("TotalGainPct")} className="flex items-center gap-1 ml-auto hover:text-on-surface">
                        Total <SortIcon col="TotalGainPct" sortKey={sortKey} sortAsc={sortAsc} />
                      </button>
                    </TableHead>
                    <TableHead className="text-center">Status</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredPositions.map((pos) => (
                    <TableRow key={pos.ID}>
                      <TableCell><span className="font-mono font-semibold">{pos.Symbol}</span></TableCell>
                      <TableCell>
                        <div className="text-sm">{pos.CompanyName}</div>
                        <div className="text-[11px] text-on-surface-variant">Sector</div>
                      </TableCell>
                      <TableCell className="text-right font-mono">{pos.Quantity}</TableCell>
                      <TableCell className="text-right font-mono">{fmtCurrency(pos.AvgCost)}</TableCell>
                      <TableCell className="text-right font-mono">{fmtCurrency(pos.CurrentPrice)}</TableCell>
                      <TableCell className="text-right font-mono font-medium">{fmtCurrency(pos.CurrentValue)}</TableCell>
                      <TableCell className={`text-right font-mono ${pos.DayChangePct >= 0 ? "text-primary" : "text-error"}`}>{fmtPct(pos.DayChangePct)}</TableCell>
                      <TableCell className={`text-right font-mono ${pos.TotalGainPct >= 0 ? "text-primary" : "text-error"}`}>{fmtPct(pos.TotalGainPct)}</TableCell>
                      <TableCell className="text-center">
                        <div className="flex items-center justify-center gap-1.5">
                          <StatusIndicator active={true} />
                          <span className="text-[11px] text-on-surface-variant capitalize">active</span>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                  {filteredPositions.length === 0 && !loading && (
                    <TableRow>
                      <TableCell colSpan={9} className="text-center text-on-surface-variant text-sm py-8">No positions available</TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          </PageCell>

          <PageCell>
            <CardTitle className="mb-4">Sector Allocation</CardTitle>
            <div className="space-y-3">
              {sectors.length > 0 ? sectors.map((sector) => (
                <div key={sector.sector}>
                  <div className="flex items-center justify-between mb-1">
                    <div className="flex items-center gap-2">
                      <div className="h-2 w-2" style={{ backgroundColor: sector.color }} />
                      <span className="text-sm">{sector.sector}</span>
                    </div>
                    <div className="flex items-center gap-3">
                      <span className="text-sm font-mono">{fmtCurrency(sector.value)}</span>
                      <span className="text-sm font-mono text-on-surface-variant w-10 text-right">{sector.percent}</span>
                    </div>
                  </div>
                  <div className="h-0.5 bg-surface-container-high w-full">
                    <div className="h-full transition-all" style={{ width: sector.percent, backgroundColor: sector.color }} />
                  </div>
                </div>
              )) : (
                <div className="text-on-surface-variant text-sm">No sector data available</div>
              )}
            </div>

            <div className="mt-6 pt-5 border-t border-outline-variant/30">
              <CardTitle className="mb-4">Quick Stats</CardTitle>
              <div className="grid grid-cols-2 gap-2">
                {[
                  { label: "Positions", value: String(positions.length) },
                  { label: "Avg Cost Basis", value: positions.length > 0 ? fmtCurrency(positions.reduce((s, p) => s + p.AvgCost * p.Quantity, 0) / positions.reduce((s, p) => s + p.Quantity, 0)) : "$0.00" },
                  { label: "Day Change", value: fmtCurrency(summary?.DayChange ?? 0), highlight: true },
                  { label: "Wtd. Avg Gain", value: fmtPct(summary?.TotalGainPct ?? 0), highlight: true },
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