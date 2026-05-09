"use client";

import { useEffect, useState } from "react";
import { ColumnDef } from "@tanstack/react-table";
import { PageGrid, PageCell, PageHeader, MetricLabel, MetricValue, MetricSubValue } from "@/components/page-layout";
import { CardTitle } from "@/components/ui/card";
import { StatusIndicator } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/ui/data-table";
import { Skeleton, MetricSkeleton } from "@/components/ui/skeleton";
import { Plus, Download } from "lucide-react";
import { usePortfolio } from "@/hooks/usePortfolio";
import { AddPositionModal } from "@/components/add-position-modal";
import { toast } from "sonner";

interface Position {
  ID: string;
  Symbol: string;
  CompanyName: string;
  Quantity: number;
  AvgCost: number;
  CurrentPrice: number;
  CurrentValue: number;
  DayChangePct: number;
  TotalGainPct: number;
}

function fmtCurrency(v: number): string {
  return new Intl.NumberFormat("en-US", { style: "currency", currency: "USD" }).format(v);
}
function fmtPct(v: number): string {
  return (v >= 0 ? "+" : "") + v.toFixed(2) + "%";
}

const columns: ColumnDef<Position>[] = [
  {
    accessorKey: "Symbol",
    header: "Symbol",
    cell: ({ row }) => <span className="font-mono font-semibold">{row.original.Symbol}</span>,
  },
  {
    accessorKey: "CompanyName",
    header: "Name",
    cell: ({ row }) => (
      <div className="text-sm">
        <div>{row.original.CompanyName}</div>
        <div className="text-[11px] text-on-surface-variant">Sector</div>
      </div>
    ),
  },
  {
    accessorKey: "Quantity",
    header: ({ column }) => (
      <button onClick={() => column.getToggleSortingHandler()?.(undefined)} className="flex items-center gap-1 ml-auto hover:text-on-surface">
        Shares
      </button>
    ),
    cell: ({ row }) => <span className="font-mono">{row.original.Quantity}</span>,
  },
  {
    accessorKey: "AvgCost",
    header: ({ column }) => (
      <button onClick={() => column.getToggleSortingHandler()?.(undefined)} className="flex items-center gap-1 ml-auto hover:text-on-surface">
        Avg Cost
      </button>
    ),
    cell: ({ row }) => <span className="font-mono">{fmtCurrency(row.original.AvgCost)}</span>,
  },
  {
    accessorKey: "CurrentPrice",
    header: ({ column }) => (
      <button onClick={() => column.getToggleSortingHandler()?.(undefined)} className="flex items-center gap-1 ml-auto hover:text-on-surface">
        Price
      </button>
    ),
    cell: ({ row }) => <span className="font-mono">{fmtCurrency(row.original.CurrentPrice)}</span>,
  },
  {
    accessorKey: "CurrentValue",
    header: ({ column }) => (
      <button onClick={() => column.getToggleSortingHandler()?.(undefined)} className="flex items-center gap-1 ml-auto hover:text-on-surface">
        Value
      </button>
    ),
    cell: ({ row }) => <span className="font-mono font-medium">{fmtCurrency(row.original.CurrentValue)}</span>,
  },
  {
    accessorKey: "DayChangePct",
    header: ({ column }) => (
      <button onClick={() => column.getToggleSortingHandler()?.(undefined)} className="flex items-center gap-1 ml-auto hover:text-on-surface">
        Day
      </button>
    ),
    cell: ({ row }) => (
      <span className={`font-mono ${row.original.DayChangePct >= 0 ? "text-primary" : "text-error"}`}>
        {fmtPct(row.original.DayChangePct)}
      </span>
    ),
  },
  {
    accessorKey: "TotalGainPct",
    header: ({ column }) => (
      <button onClick={() => column.getToggleSortingHandler()?.(undefined)} className="flex items-center gap-1 ml-auto hover:text-on-surface">
        Total
      </button>
    ),
    cell: ({ row }) => (
      <span className={`font-mono ${row.original.TotalGainPct >= 0 ? "text-primary" : "text-error"}`}>
        {fmtPct(row.original.TotalGainPct)}
      </span>
    ),
  },
  {
    id: "status",
    header: "Status",
    cell: () => (
      <div className="flex items-center justify-center gap-1.5">
        <StatusIndicator active={true} />
        <span className="text-[11px] text-on-surface-variant capitalize">active</span>
      </div>
    ),
  },
];

export default function PortfolioPage() {
  const { positions, summary, loading, refresh } = usePortfolio();
  const [modalOpen, setModalOpen] = useState(false);

  useEffect(() => { refresh(); }, [refresh]);

  const handleAddPosition = (position: { symbol: string; shares: number; price: number }) => {
    toast.success(`Added ${position.shares} shares of ${position.symbol} @ $${position.price.toFixed(2)}`);
    setModalOpen(false);
    refresh();
  };

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
            <Button variant="secondary" size="sm" onClick={() => toast.info("Exporting portfolio as CSV...")}>
              <Download className="h-4 w-4" /> Export
            </Button>
            <Button variant="default" size="sm" onClick={() => setModalOpen(true)}>
              <Plus className="h-4 w-4" /> Add Position
            </Button>
          </div>
        </PageHeader>
      </div>

      <div className="flex-1 px-6 pb-6 overflow-auto">
        <PageGrid className="mb-4" style={{ gridTemplateColumns: "repeat(5, 1fr)" }}>
          {[1, 2, 3, 4, 5].map((i) => (
            <PageCell key={i}>
              {loading ? <MetricSkeleton /> : (
                <>
                  <MetricLabel>{["Total Value", "Day P&L", "Total Invested", "Total Gain/Loss", "Cash Balance"][i - 1]}</MetricLabel>
                  <MetricValue highlight={i === 2}>{fmtCurrency([summary?.TotalValue ?? 0, summary?.DayChange ?? 0, summary?.TotalInvested ?? 0, summary?.TotalGain ?? 0, summary?.CashBalance ?? 0][i - 1] as number)}</MetricValue>
                  <MetricSubValue positive>{i === 1 ? fmtPct(summary?.DayChangePct ?? 0) : i === 2 ? fmtPct(summary?.DayChangePct ?? 0) : i === 3 ? "Cost Basis" : i === 4 ? fmtPct(summary?.TotalGainPct ?? 0) : "Buying Power"}</MetricSubValue>
                </>
              )}
            </PageCell>
          ))}
        </PageGrid>

        <PageGrid style={{ gridTemplateColumns: "2fr 1fr" }}>
          <PageCell>
            <div className="flex items-center justify-between mb-4">
              <CardTitle>All Positions</CardTitle>
            </div>
            <DataTable
              columns={columns}
              data={positions}
              loading={loading}
              emptyMessage="No positions available"
              searchPlaceholder="Search holdings..."
            />
          </PageCell>

          <PageCell>
            <CardTitle className="mb-4">Sector Allocation</CardTitle>
            <div className="space-y-3">
              {loading ? (
                Array.from({ length: 3 }).map((_, i) => (
                  <div key={i} className="space-y-1">
                    <div className="flex items-center justify-between">
                      <Skeleton className="h-4 w-20" />
                      <Skeleton className="h-4 w-16" />
                    </div>
                    <Skeleton className="h-0.5 w-full" />
                  </div>
                ))
              ) : sectors.length > 0 ? sectors.map((sector) => (
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

      <AddPositionModal open={modalOpen} onClose={() => setModalOpen(false)} onAdd={handleAddPosition} />
    </div>
  );
}