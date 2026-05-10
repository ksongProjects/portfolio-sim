"use client";

import { useState, useEffect, useRef } from "react";
import { X, Search, TrendingUp, TrendingDown, BarChart2, DollarSign, Percent, Clock, ChevronRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { useTickerLookup, type TickerDetails, type IntradayBar, type FinancialRatio } from "@/hooks/useTickerLookup";
import { cn } from "@/lib/utils";

interface AddPositionModalProps {
  open: boolean;
  onClose: () => void;
  onAdd: (position: { symbol: string; shares: number; price: number }) => void;
}

function fmtCurrency(v: number): string {
  return new Intl.NumberFormat("en-US", { style: "currency", currency: "USD" }).format(v);
}

function fmtPct(v: number): string {
  return (v >= 0 ? "+" : "") + v.toFixed(2) + "%";
}

function fmtNumber(v: number): string {
  if (v >= 1e12) return (v / 1e12).toFixed(2) + "T";
  if (v >= 1e9) return (v / 1e9).toFixed(2) + "B";
  if (v >= 1e6) return (v / 1e6).toFixed(2) + "M";
  if (v >= 1e3) return (v / 1e3).toFixed(2) + "K";
  return v.toFixed(2);
}

function IntradayChart({ data }: { data: IntradayBar[] }) {
  if (data.length === 0) return <div className="h-32 flex items-center justify-center text-on-surface-variant text-sm">No chart data</div>;

  const prices = data.map((d) => d.close);
  const min = Math.min(...prices);
  const max = Math.max(...prices);
  const range = max - min || 1;
  const height = 120;

  const firstClose = data[0]?.close || 0;
  const lastClose = data[data.length - 1]?.close || 0;
  const isUp = lastClose >= firstClose;

  const points = data.map((bar, i) => {
    const x = (i / (data.length - 1)) * 100;
    const y = height - ((bar.close - min) / range) * height;
    return `${x},${y}`;
  }).join(" ");

  return (
    <div className="relative h-32">
      <svg viewBox="0 0 100 100" preserveAspectRatio="none" className="w-full h-full" style={{ width: "100%", height: height }}>
        <polyline
          points={points}
          fill="none"
          stroke={isUp ? "#3fe56c" : "#ff4d4d"}
          strokeWidth="1.5"
          vectorEffect="non-scaling-stroke"
        />
      </svg>
      <div className="absolute bottom-0 left-0 right-0 flex justify-between text-[10px] text-on-surface-variant/50 pt-1">
        <span>{data[0]?.timestamp?.split(" ")[1] || ""}</span>
        <span>{data[data.length - 1]?.timestamp?.split(" ")[1] || ""}</span>
      </div>
    </div>
  );
}

function RatioCard({ ratio }: { ratio: FinancialRatio }) {
  return (
    <div className="p-3 border border-outline-variant/30">
      <div className="text-[10px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant">{ratio.label}</div>
      <div className="text-base font-mono font-medium text-on-surface mt-0.5">{ratio.value}</div>
      <div className="text-[10px] text-on-surface-variant/60 mt-0.5">{ratio.description}</div>
    </div>
  );
}

export function AddPositionModal({ open, onClose, onAdd }: AddPositionModalProps) {
  const [step, setStep] = useState<"search" | "details" | "confirm">("search");
  const [query, setQuery] = useState("");
  const [shares, setShares] = useState("");
  const [price, setPrice] = useState("");
  const searchRef = useRef<HTMLInputElement>(null);

  const { searchResults, selectedTicker, intradayData, ratios, loading: searchLoading, searchTickers, lookupTicker, clearSelection } = useTickerLookup();

  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (open && searchRef.current) {
      searchRef.current.focus();
    }
  }, [open]);

  const handleSearch = async () => {
    if (!query.trim()) return;
    setLoading(true);
    await searchTickers(query);
    setLoading(false);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      handleSearch();
    }
  };

  const handleTickerSelect = (ticker: TickerDetails) => {
    setPrice(ticker.price.toFixed(2));
    lookupTicker(ticker.symbol);
    setStep("details");
  };

  const handleConfirm = () => {
    if (!selectedTicker || !shares || !price) return;
    onAdd({
      symbol: selectedTicker.symbol,
      shares: parseFloat(shares),
      price: parseFloat(price),
    });
    handleClose();
  };

  const handleClose = () => {
    setStep("search");
    setQuery("");
    setShares("");
    setPrice("");
    clearSelection();
    onClose();
  };

  if (!open) return null;

  return (
    <>
      <div className="fixed inset-0 z-50 bg-black/50" onClick={handleClose} />
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div className="bg-surface border border-outline-variant w-full max-w-2xl max-h-[90vh] overflow-hidden flex flex-col rounded-lg shadow-xl">
          <div className="flex items-center justify-between px-5 py-4 border-b border-outline-variant">
            <div>
              <h2 className="text-base font-semibold text-on-surface">Add Position</h2>
              <div className="flex items-center gap-2 mt-1">
                {step !== "search" && (
                  <button onClick={() => step === "confirm" ? setStep("details") : setStep("search")} className="text-xs text-primary hover:underline">
                    ← Back
                  </button>
                )}
                <span className="text-xs text-on-surface-variant capitalize">
                  {step === "search" ? "Search for a ticker" : step === "details" ? "Review ticker details" : "Confirm position"}
                </span>
              </div>
            </div>
            <button onClick={handleClose} className="text-on-surface-variant hover:text-on-surface transition-colors">
              <X className="h-5 w-5" />
            </button>
          </div>

          <div className="flex-1 overflow-y-auto p-5">
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
                  <Button variant="default" size="default" onClick={handleSearch} disabled={loading || !query.trim()}>
                    <Search className="h-4 w-4" />
                    {loading ? "..." : "Search"}
                  </Button>
                </div>

                {loading && (
                  <div className="text-sm text-on-surface-variant py-4 text-center">Searching...</div>
                )}

                {searchResults.length > 0 && (
                  <ul className="border border-outline-variant/30 divide-y divide-outline-variant/30">
                    {searchResults.map((ticker) => (
                      <li key={ticker.symbol}>
                        <button
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
                            <div className="text-right">
                              <div className="text-sm font-mono">{fmtCurrency(ticker.price)}</div>
                              <div className={cn("text-xs font-mono", ticker.changePct >= 0 ? "text-primary" : "text-error")}>
                                {fmtPct(ticker.changePct)}
                              </div>
                            </div>
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
                  <IntradayChart data={intradayData} />
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

                {ratios.length > 0 && (
                  <div>
                    <div className="text-xs font-semibold uppercase tracking-[0.08em] text-on-surface-variant mb-3">Financial Ratios</div>
                    <div className="grid grid-cols-2 gap-2">
                      {ratios.map((ratio) => (
                        <RatioCard key={ratio.label} ratio={ratio} />
                      ))}
                    </div>
                  </div>
                )}

                <Button variant="default" size="default" className="w-full" onClick={() => setStep("confirm")}>
                  Continue to Confirm
                </Button>
              </div>
            )}

            {step === "confirm" && selectedTicker && (
              <div className="space-y-5">
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
                  />
                  <Input
                    label="Price per Share"
                    type="number"
                    value={price}
                    onChange={(e) => setPrice(e.target.value)}
                    placeholder="0.00"
                    min="0"
                    step="0.01"
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
                  <Button variant="secondary" size="default" className="flex-1" onClick={handleClose}>
                    Cancel
                  </Button>
                  <Button
                    variant="default"
                    size="default"
                    className="flex-1"
                    onClick={handleConfirm}
                    disabled={!shares || !price || parseFloat(shares) <= 0 || parseFloat(price) <= 0}
                  >
                    Add Position
                  </Button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </>
  );
}
