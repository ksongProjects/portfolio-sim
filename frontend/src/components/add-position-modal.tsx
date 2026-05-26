"use client";

import { useState, useEffect, useRef } from "react";
import { Search, TrendingUp, TrendingDown, BarChart2, DollarSign, Percent, Clock, ChevronRight, ChevronLeft } from "lucide-react";
import { LineChart, Line, XAxis, YAxis, ResponsiveContainer, CartesianGrid } from "recharts";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Modal } from "@/components/ui/modal";
import { useTickerLookup, type TickerDetails } from "@/hooks/useTickerLookup";
import { cn } from "@/lib/utils";

interface AddPositionModalProps {
  open: boolean;
  onClose: () => void;
  onAdd: (position: { symbol: string; shares: number; price: number }) => void;
}

function fmtCurrency(v: number): string {
  return new Intl.NumberFormat("en-US", { style: "currency", currency: "USD" }).format(v);
}

function fmtPct(v: number | undefined): string {
  if (v === undefined || isNaN(v)) return "N/A";
  return (v >= 0 ? "+" : "") + v.toFixed(2) + "%";
}

function fmtNumber(v: number | undefined): string {
  if (v === undefined || isNaN(v)) return "N/A";
  if (v >= 1e12) return (v / 1e12).toFixed(2) + "T";
  if (v >= 1e9) return (v / 1e9).toFixed(2) + "B";
  if (v >= 1e6) return (v / 1e6).toFixed(2) + "M";
  if (v >= 1e3) return (v / 1e3).toFixed(2) + "K";
  return v.toFixed(2);
}

function formatTime(timestamp: string): string {
  const date = new Date(timestamp);
  if (isNaN(date.getTime())) {
    const parts = timestamp.split(" ");
    return parts.length > 1 ? parts[1].substring(0, 5) : timestamp;
  }
  return date.toLocaleTimeString("en-US", { hour: "2-digit", minute: "2-digit", timeZone: "America/New_York" });
}

function IntradayChart({ data, dataDate }: { data: { timestamp: string; close: number; volume?: number }[]; dataDate?: string }) {
  if (data.length === 0) return <div className="h-32 flex items-center justify-center text-on-surface-variant text-sm">No chart data</div>;

  const firstClose = data[0]?.close || 0;
  const lastClose = data[data.length - 1]?.close || 0;
  const isUp = lastClose >= firstClose;
  const color = isUp ? "#3fe56c" : "#ff4d4d";

  const chartData = data.map((d) => ({
    ...d,
    time: formatTime(d.timestamp),
  }));

  return (
    <div className="relative">
      {dataDate && (
        <div className="absolute top-2 left-2 z-10 px-2 py-1 bg-surface-container-high/90 rounded text-xs text-on-surface font-mono">
          {dataDate}
        </div>
      )}
      <ResponsiveContainer width="100%" height={120}>
      <LineChart data={chartData} margin={{ top: 5, right: 5, left: 0, bottom: 0 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="var(--outline-variant)" />
        <XAxis
          dataKey="time"
          tick={{ fontSize: 9, fill: "#9ca3af" }}
          tickLine={true}
          axisLine={{ stroke: "var(--outline)" }}
        />
        <YAxis
          domain={["auto", "auto"]}
          tick={{ fontSize: 9, fill: "#9ca3af" }}
          tickLine={false}
          axisLine={false}
          tickFormatter={(v) => v.toFixed(0)}
          width={45}
        />
        <Line
          type="monotone"
          dataKey="close"
          stroke={color}
          strokeWidth={1.5}
          dot={false}
          activeDot={{ r: 3, fill: color }}
        />
      </LineChart>
    </ResponsiveContainer>
    </div>
  );
}

export function AddPositionModal({ open, onClose, onAdd }: AddPositionModalProps) {
  const [step, setStep] = useState<"search" | "details" | "confirm">("search");
  const [query, setQuery] = useState("");
  const [shares, setShares] = useState("");
  const [price, setPrice] = useState("");
  const searchRef = useRef<HTMLInputElement>(null);
  const formRef = useRef<HTMLFormElement>(null);

  const {
    searchResults,
    selectedTicker,
    intradayData,
    intradayDataDate,
    loading: tickerLoading,
    searchLoading,
    searchTickers,
    lookupTicker,
    clearSelection,
  } = useTickerLookup();

  useEffect(() => {
    if (!open) {
      setStep("search");
      setQuery("");
      setShares("");
      setPrice("");
      clearSelection();
    }
  }, [open, clearSelection]);

  const handleSearch = async () => {
    if (!query.trim()) return;
    await searchTickers(query);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      if (step === "search") {
        e.preventDefault();
        handleSearch();
      } else if (step === "confirm") {
        e.preventDefault();
        formRef.current?.requestSubmit();
      }
    }
  };

  const handleTickerSelect = (ticker: TickerDetails) => {
    void lookupTicker(ticker.symbol);
    setStep("details");
  };

  const handleConfirm = () => {
    if (!selectedTicker || !shares || !price) return;
    onAdd({
      symbol: selectedTicker.symbol,
      shares: parseFloat(shares),
      price: parseFloat(price),
    });
    onClose();
  };

  const handleBack = () => {
    if (step === "details") {
      setStep("search");
    } else if (step === "confirm") {
      setStep("details");
    }
  };

  const handleClose = () => {
    setStep("search");
    setQuery("");
    setShares("");
    setPrice("");
    onClose();
  };

  return (
    <Modal
      open={open}
      onClose={handleClose}
      title="Add Position"
      description={step === "search" ? "Search for a ticker to add to your portfolio" : step === "details" ? "Review ticker details before adding" : "Confirm position details"}
    >
      <div className="space-y-4">
        {step === "search" && (
          <div className="space-y-4">
            <div className="flex gap-2">
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-on-surface-variant" />
                <input
                  ref={searchRef}
                  type="text"
                  value={query}
                  onChange={(e) => setQuery(e.target.value.toUpperCase())}
                  onKeyDown={handleKeyDown}
                  placeholder="Search by ticker or company name..."
                  className="w-full h-10 pl-10 pr-4 bg-surface-container-low border border-outline/30 text-sm text-on-surface placeholder:text-on-surface-variant/50 focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary/30"
                />
              </div>
              <Button variant="default" size="default" onClick={handleSearch} disabled={searchLoading || !query.trim()}>
                <Search className="h-4 w-4" />
                {searchLoading ? "..." : "Search"}
              </Button>
            </div>

            {searchLoading && (
              <div className="text-sm text-on-surface-variant py-4 text-center">Searching...</div>
            )}

            {searchResults.length > 0 && (
              <ul className="border border-outline-variant/30 divide-y divide-outline-variant/30 max-h-64 overflow-y-auto">
                {searchResults
                  .filter((ticker) => {
                    const isEquity = !ticker.symbol.includes("/") && !ticker.symbol.includes("^") && !ticker.exchange?.includes("Index");
                    return isEquity;
                  })
                  .map((ticker) => (
                  <li key={ticker.symbol}>
                    <button
                      type="button"
                      onClick={() => handleTickerSelect(ticker)}
                      className="w-full flex items-center justify-between p-3 hover:bg-surface-container-low transition-colors text-left"
                    >
                      <div className="flex items-center gap-3">
                        <div className="flex flex-col">
                          <span className="text-sm font-mono font-semibold text-on-surface">{ticker.symbol}</span>
                          <span className="text-xs text-on-surface-variant">{ticker.name}</span>
                        </div>
                      </div>
                      <div className="flex items-center gap-4">
                        <ChevronRight className="h-4 w-4 text-on-surface-variant" />
                      </div>
                    </button>
                  </li>
                ))}
              </ul>
            )}

            {query.length > 0 && !searchLoading && searchResults.length === 0 && (
              <div className="text-sm text-on-surface-variant py-8 text-center">No tickers found for &quot;{query}&quot;</div>
            )}
          </div>
        )}

        {step === "details" && !selectedTicker && (
          <div className="h-48 flex items-center justify-center text-sm text-on-surface-variant">
            {tickerLoading ? "Loading ticker details..." : "Ticker details unavailable"}
          </div>
        )}

        {step === "details" && selectedTicker && (
          <div className="space-y-5">
            <div className="flex items-start justify-between">
              <div>
                <div className="flex items-center gap-2">
                  <span className="text-lg font-mono font-bold text-on-surface">{selectedTicker.symbol}</span>
                  <Badge variant="secondary">{selectedTicker.exchange}</Badge>
                </div>
                <div className="text-sm text-on-surface-variant mt-0.5">{selectedTicker.name}</div>
              </div>
              <div className="text-right">
                <div className="text-xl font-mono font-semibold text-on-surface">{fmtCurrency(selectedTicker.price)}</div>
                <div className={cn("text-sm font-mono flex items-center gap-1", selectedTicker.changePct >= 0 ? "text-primary" : "text-error")}>
                  {selectedTicker.changePct >= 0 ? <TrendingUp className="h-3 w-3" /> : <TrendingDown className="h-3 w-3" />}
                  {fmtPct(selectedTicker.changePct)}
                </div>
              </div>
            </div>

            <div className="border border-outline-variant/30 p-4">
              <div className="flex items-center gap-1 mb-3">
                <Clock className="h-4 w-4 text-on-surface-variant" />
                <span className="text-xs font-semibold uppercase tracking-[0.08em] text-on-surface-variant">Intraday Chart</span>
              </div>
              <IntradayChart data={intradayData} dataDate={intradayDataDate} />
            </div>

            <div className="grid grid-cols-2 gap-3">
              {[
                { icon: DollarSign, label: "Market Cap", value: fmtNumber(selectedTicker.marketCap) },
                { icon: BarChart2, label: "P/E Ratio", value: selectedTicker.peRatio?.toFixed(2) || "N/A" },
                { icon: Percent, label: "Div Yield", value: selectedTicker.dividendYield ? (selectedTicker.dividendYield * 100).toFixed(2) + "%" : "N/A" },
                { icon: TrendingUp, label: "52W High", value: fmtCurrency(selectedTicker.week52High) },
              ].map((item) => (
                <div key={item.label} className="flex items-center gap-2 p-3 border border-outline-variant/30">
                  <item.icon className="h-4 w-4 text-on-surface-variant shrink-0" />
                  <div>
                    <div className="text-[10px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant">{item.label}</div>
                    <div className="text-sm font-mono text-on-surface">{item.value}</div>
                  </div>
                </div>
              ))}
            </div>

            <div className="flex gap-3">
              <Button variant="secondary" size="default" className="flex-1" type="button" onClick={handleBack}>
                <ChevronLeft className="h-4 w-4" />
                Back
              </Button>
              <Button variant="default" size="default" className="flex-1" onClick={() => { setPrice(selectedTicker.price.toString()); setStep("confirm"); }}>
                Continue
                <ChevronRight className="h-4 w-4" />
              </Button>
            </div>
          </div>
        )}

        {step === "confirm" && selectedTicker && (
          <form
            ref={formRef}
            onSubmit={(e) => {
              e.preventDefault();
              handleConfirm();
            }}
            className="space-y-5"
          >
            <div className="flex items-center justify-between p-4 border border-outline-variant/30">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-primary/10 rounded-full flex items-center justify-center">
                  <span className="text-sm font-mono font-bold text-primary">{selectedTicker.symbol}</span>
                </div>
                <div>
                  <div className="text-sm font-semibold text-on-surface">{selectedTicker.symbol}</div>
                  <div className="text-xs text-on-surface-variant">{selectedTicker.name}</div>
                </div>
              </div>
              <div className="text-right">
                <div className="text-xs text-on-surface-variant">Current Price</div>
                <div className="text-sm font-mono text-on-surface">{fmtCurrency(selectedTicker.price)}</div>
              </div>
            </div>

            <div className="space-y-4">
              <Input
                label="Number of Shares"
                type="number"
                value={shares}
                onChange={(e) => setShares(e.target.value)}
                placeholder="0"
                min="0"
                step="any"
                className="[&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none [&]:appearance-none"
              />
              <Input
                label="Price per Share"
                type="number"
                value={price}
                onChange={(e) => setPrice(e.target.value)}
                placeholder="0.00"
                min="0"
                step="0.01"
                className="[&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none [&]:appearance-none"
              />
            </div>

            <div className="p-4 bg-surface-container-low border border-outline-variant/30">
              <div className="flex items-center justify-between">
                <span className="text-sm text-on-surface-variant">Estimated Total</span>
                <span className="text-xl font-mono font-semibold text-on-surface">
                  {fmtCurrency(parseFloat(shares || "0") * parseFloat(price || "0"))}
                </span>
              </div>
            </div>

            <div className="flex gap-3">
              <Button variant="secondary" size="default" className="flex-1" type="button" onClick={handleClose}>
                Cancel
              </Button>
              <Button
                variant="default"
                size="default"
                className="flex-1"
                type="submit"
                disabled={!shares || !price || parseFloat(shares) <= 0 || parseFloat(price) <= 0}
              >
                Add Position
              </Button>
            </div>
            <Button variant="ghost" size="sm" type="button" onClick={handleBack} className="w-full">
              <ChevronLeft className="h-4 w-4" />
              Back to Details
            </Button>
          </form>
        )}
      </div>
    </Modal>
  );
}
