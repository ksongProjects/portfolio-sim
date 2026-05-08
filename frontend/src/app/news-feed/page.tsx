"use client";

import { useState } from "react";
import { PageCell, PageHeader } from "@/components/page-layout";
import { CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Clock, TrendingUp, ExternalLink, Search, Filter, Bookmark } from "lucide-react";

const featuredNews = {
  id: 1,
  title: "Fed Signals Potential Rate Cuts Amid Economic Uncertainty",
  source: "Reuters",
  time: "2 hours ago",
  summary: "Federal Reserve officials indicated they may consider reducing interest rates in the coming months as economic data shows signs of cooling inflation and stable employment figures.",
  sentiment: "bullish",
  relatedSymbols: ["SPY", "DIA", "TLT"],
};

const newsItems = [
  { id: 2, title: "NVIDIA Reports Record Data Center Revenue, Stock Surges 5%", source: "Bloomberg", time: "4 hours ago", sentiment: "bullish", symbols: ["NVDA"] },
  { id: 3, title: "Apple Unveils New AI Features for iPhone Pro Line", source: "CNBC", time: "5 hours ago", sentiment: "neutral", symbols: ["AAPL"] },
  { id: 4, title: "Oil Prices Drop on Increased Supply Forecasts", source: "WSJ", time: "6 hours ago", sentiment: "bearish", symbols: ["XOM", "CVX"] },
  { id: 5, title: "Tesla Expands Supercharger Network in Europe", source: "Reuters", time: "8 hours ago", sentiment: "neutral", symbols: ["TSLA"] },
  { id: 6, title: "Microsoft Cloud Revenue Exceeds Expectations", source: "Bloomberg", time: "10 hours ago", sentiment: "bullish", symbols: ["MSFT"] },
  { id: 7, title: "JPMorgan Raises S&P 500 Target to 5,500", source: "CNBC", time: "12 hours ago", sentiment: "bullish", symbols: ["SPY"] },
  { id: 8, title: "Chinese Tech Stocks Rally on Policy Support", source: "FT", time: "Yesterday", sentiment: "bullish", symbols: ["BABA", "JD"] },
];

const marketIndices = [
  { symbol: "SPY", name: "S&P 500 ETF", price: "$523.45", change: "+0.87%", positive: true },
  { symbol: "QQQ", name: "Nasdaq ETF", price: "$448.30", change: "+1.12%", positive: true },
  { symbol: "DIA", name: "Dow Jones ETF", price: "$398.20", change: "+0.45%", positive: true },
];

export default function NewsFeedPage() {
  const [search, setSearch] = useState("");
  const [saved, setSaved] = useState<number[]>([]);
  const [toast, setToast] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<"all" | "bullish" | "bearish" | "neutral">("all");

  const showToast = (msg: string) => {
    setToast(msg);
    setTimeout(() => setToast(null), 3000);
  };

  const toggleSaved = (id: number) => {
    setSaved((prev) =>
      prev.includes(id) ? prev.filter((s) => s !== id) : [...prev, id]
    );
    showToast(saved.includes(id) ? "Removed from saved" : "Added to saved");
  };

  const filteredNews = newsItems.filter((n) => {
    const matchesSearch =
      n.title.toLowerCase().includes(search.toLowerCase()) ||
      n.source.toLowerCase().includes(search.toLowerCase());
    const matchesTab = activeTab === "all" || n.sentiment === activeTab;
    return matchesSearch && matchesTab;
  });

  return (
    <div className="flex flex-col h-full bg-surface">
      <div className="px-6 pt-6">
        <PageHeader
          title="News Feed"
          description="Real-time market news and sentiment analysis"
        >
          <div className="flex gap-3">
            <Button
              variant={saved.length > 0 ? "default" : "secondary"}
              size="sm"
              onClick={() => showToast(`Saved articles: ${saved.length}`)}
            >
              <Bookmark className="h-4 w-4" />
              Saved ({saved.length})
            </Button>
            <Button variant="default" size="sm" onClick={() => showToast("Market summary coming soon...")}>
              <TrendingUp className="h-4 w-4" />
              Market Summary
            </Button>
          </div>
        </PageHeader>
      </div>

      <div className="flex-1 px-6 pb-6 overflow-auto">
        <div className="grid gap-[1px] bg-outline-variant mb-[1px]">
          <div className="grid grid-cols-3 gap-[1px] bg-outline-variant">
            <PageCell>
              <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant mb-2">
                Articles Today
              </div>
              <div className="text-[24px] font-medium tracking-[-0.01em] text-on-surface" style={{ fontFamily: "var(--font-work-sans)" }}>
                {newsItems.length}
              </div>
              <div className="text-[13px] font-mono text-primary mt-1">+2 since yesterday</div>
            </PageCell>
            <PageCell>
              <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant mb-2">
                Bullish Sentiment
              </div>
              <div className="text-[24px] font-medium tracking-[-0.01em] text-primary" style={{ fontFamily: "var(--font-work-sans)" }}>
                {newsItems.filter((n) => n.sentiment === "bullish").length}
              </div>
              <div className="text-[13px] font-mono text-on-surface-variant mt-1">
                {Math.round((newsItems.filter((n) => n.sentiment === "bullish").length / newsItems.length) * 100)}% of coverage
              </div>
            </PageCell>
            <PageCell>
              <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant mb-2">
                Top Symbol
              </div>
              <div className="text-[24px] font-medium tracking-[-0.01em] text-on-surface" style={{ fontFamily: "var(--font-work-sans)" }}>
                NVDA
              </div>
              <div className="text-[13px] font-mono text-primary mt-1">3 mentions</div>
            </PageCell>
          </div>
        </div>

        <div className="flex gap-[1px] bg-outline-variant">
          <PageCell className="flex-[2]">
            <div className="flex items-center justify-between mb-4">
              <CardTitle className="text-base font-semibold text-on-surface">Latest News</CardTitle>
              <div className="flex gap-3">
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-on-surface-variant" />
                  <Input
                    placeholder="Search news..."
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

            <div className="flex gap-1 mb-4">
              {(["all", "bullish", "bearish", "neutral"] as const).map((tab) => (
                <button
                  key={tab}
                  onClick={() => setActiveTab(tab)}
                  className={`px-3 py-1.5 text-[11px] font-semibold uppercase tracking-[0.06em] transition-colors capitalize ${
                    activeTab === tab
                      ? "bg-primary text-on-primary"
                      : "text-on-surface-variant hover:text-on-surface hover:bg-surface-container-high"
                  }`}
                >
                  {tab}
                </button>
              ))}
            </div>

            <div className="space-y-2">
              {filteredNews.map((news) => (
                <div
                  key={news.id}
                  className="flex items-start gap-4 p-4 border border-outline-variant hover:bg-surface-container-low transition-colors cursor-pointer"
                >
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <Badge
                        variant={news.sentiment === "bullish" ? "success" : news.sentiment === "bearish" ? "error" : "secondary"}
                        className="text-[10px]"
                      >
                        {news.sentiment}
                      </Badge>
                      <span className="text-xs text-on-surface-variant">{news.source}</span>
                      <span className="text-xs text-on-surface-variant flex items-center gap-1">
                        <Clock className="h-3 w-3" />
                        {news.time}
                      </span>
                    </div>
                    <h3 className="text-sm font-medium text-on-surface mb-2">{news.title}</h3>
                    <div className="flex gap-2">
                      {news.symbols.map((sym) => (
                        <Badge key={sym} variant="outline" className="text-[10px] font-mono">
                          {sym}
                        </Badge>
                      ))}
                    </div>
                  </div>
                  <div className="flex flex-col items-end gap-2">
                    <a
                      href="#"
                      className="text-on-surface-variant hover:text-primary transition-colors"
                      onClick={(e) => e.preventDefault()}
                    >
                      <ExternalLink className="h-4 w-4" />
                    </a>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        toggleSaved(news.id);
                      }}
                      className="text-on-surface-variant hover:text-primary transition-colors"
                    >
                      <Bookmark
                        className={`h-4 w-4 ${saved.includes(news.id) ? "fill-primary text-primary" : ""}`}
                      />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </PageCell>

          <PageCell className="flex-1">
            <div className="space-y-[1px] bg-outline-variant">
              <div className="bg-surface-container p-4">
                <div className="flex items-center justify-between mb-3">
                  <CardTitle className="text-base font-semibold text-on-surface">Featured Story</CardTitle>
                  <Badge
                    variant={featuredNews.sentiment === "bullish" ? "success" : featuredNews.sentiment === "bearish" ? "error" : "secondary"}
                    className="text-[10px]"
                  >
                    {featuredNews.sentiment}
                  </Badge>
                </div>
                <h3 className="text-base font-medium text-on-surface leading-snug mb-2">{featuredNews.title}</h3>
                <p className="text-sm text-on-surface-variant leading-relaxed mb-3">{featuredNews.summary}</p>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2 text-xs text-on-surface-variant">
                    <span>{featuredNews.source}</span>
                    <span>&bull;</span>
                    <span>{featuredNews.time}</span>
                  </div>
                </div>
                <div className="flex gap-2 mt-3">
                  {featuredNews.relatedSymbols.map((sym) => (
                    <Badge key={sym} variant="outline" className="text-[10px] font-mono">
                      {sym}
                    </Badge>
                  ))}
                </div>
              </div>

              <div className="bg-surface-container p-4">
                <CardTitle className="text-base font-semibold text-on-surface mb-3">Market Indices</CardTitle>
                <div className="space-y-2">
                  {marketIndices.map((index) => (
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
