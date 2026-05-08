"use client";

import { useState } from "react";
import { PageGrid, PageCell, PageHeader, MetricLabel, MetricValue, MetricSubValue } from "@/components/page-layout";
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
    setSaved((prev) => (prev.includes(id) ? prev.filter((s) => s !== id) : [...prev, id]));
    showToast(saved.includes(id) ? "Removed from saved" : "Added to saved");
  };

  const filteredNews = newsItems.filter((n) => {
    const matchesSearch = n.title.toLowerCase().includes(search.toLowerCase()) || n.source.toLowerCase().includes(search.toLowerCase());
    const matchesTab = activeTab === "all" || n.sentiment === activeTab;
    return matchesSearch && matchesTab;
  });

  const bullishCount = newsItems.filter((n) => n.sentiment === "bullish").length;

  return (
    <div className="flex flex-col h-full">
      <div className="px-6 pt-6 pb-4">
        <PageHeader title="News Feed" description="Real-time market news and sentiment analysis">
          <div className="flex gap-3">
            <Button variant={saved.length > 0 ? "default" : "secondary"} size="sm" onClick={() => showToast(`Saved articles: ${saved.length}`)}>
              <Bookmark className="h-4 w-4" /> Saved ({saved.length})
            </Button>
            <Button variant="default" size="sm" onClick={() => showToast("Market summary coming soon...")}>
              <TrendingUp className="h-4 w-4" /> Market Summary
            </Button>
          </div>
        </PageHeader>
      </div>

      <div className="flex-1 px-6 pb-6 overflow-auto">
        <PageGrid className="mb-4" style={{ gridTemplateColumns: "repeat(3, 1fr)" }}>
          <PageCell>
            <MetricLabel>Articles Today</MetricLabel>
            <MetricValue>{newsItems.length}</MetricValue>
            <MetricSubValue positive>+2 since yesterday</MetricSubValue>
          </PageCell>
          <PageCell>
            <MetricLabel>Bullish Sentiment</MetricLabel>
            <MetricValue highlight>{bullishCount}</MetricValue>
            <MetricSubValue>{Math.round((bullishCount / newsItems.length) * 100)}% of coverage</MetricSubValue>
          </PageCell>
          <PageCell>
            <MetricLabel>Top Symbol</MetricLabel>
            <MetricValue>NVDA</MetricValue>
            <MetricSubValue positive>3 mentions</MetricSubValue>
          </PageCell>
        </PageGrid>

        <PageGrid style={{ gridTemplateColumns: "2fr 1fr" }}>
          <PageCell>
            <div className="flex items-center justify-between mb-4">
              <CardTitle>Latest News</CardTitle>
              <div className="flex gap-3">
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-on-surface-variant" />
                  <Input placeholder="Search news..." className="pl-9 w-[200px]" value={search} onChange={(e) => setSearch(e.target.value)} />
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
                    activeTab === tab ? "bg-primary text-on-primary" : "text-on-surface-variant hover:text-on-surface hover:bg-surface-container"
                  }`}
                >
                  {tab}
                </button>
              ))}
            </div>

            <div className="space-y-2">
              {filteredNews.map((news) => (
                <div key={news.id} className="flex items-start gap-4 p-4 border border-outline-variant/30 hover:bg-surface-container-low transition-colors cursor-pointer">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <Badge variant={news.sentiment === "bullish" ? "success" : news.sentiment === "bearish" ? "error" : "secondary"}>{news.sentiment}</Badge>
                      <span className="text-xs text-on-surface-variant">{news.source}</span>
                      <span className="text-xs text-on-surface-variant flex items-center gap-1"><Clock className="h-3 w-3" />{news.time}</span>
                    </div>
                    <h3 className="text-sm font-medium mb-2">{news.title}</h3>
                    <div className="flex gap-2">
                      {news.symbols.map((sym) => <Badge key={sym} variant="outline">{sym}</Badge>)}
                    </div>
                  </div>
                  <div className="flex flex-col items-end gap-2">
                    <a href="#" className="text-on-surface-variant hover:text-primary transition-colors" onClick={(e) => e.preventDefault()}>
                      <ExternalLink className="h-4 w-4" />
                    </a>
                    <button onClick={(e) => { e.stopPropagation(); toggleSaved(news.id); }} className="text-on-surface-variant hover:text-primary transition-colors">
                      <Bookmark className={`h-4 w-4 ${saved.includes(news.id) ? "fill-primary text-primary" : ""}`} />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </PageCell>

          <PageCell>
            <div className="space-y-[1px]">
              <div className="bg-surface-container p-4">
                <div className="flex items-center justify-between mb-3">
                  <CardTitle>Featured Story</CardTitle>
                  <Badge variant={featuredNews.sentiment === "bullish" ? "success" : featuredNews.sentiment === "bearish" ? "error" : "secondary"}>{featuredNews.sentiment}</Badge>
                </div>
                <h3 className="text-sm font-medium leading-snug mb-2">{featuredNews.title}</h3>
                <p className="text-xs text-on-surface-variant leading-relaxed mb-3">{featuredNews.summary}</p>
                <div className="flex items-center gap-2 text-xs text-on-surface-variant">
                  <span>{featuredNews.source}</span><span>&bull;</span><span>{featuredNews.time}</span>
                </div>
                <div className="flex gap-2 mt-3">
                  {featuredNews.relatedSymbols.map((sym) => <Badge key={sym} variant="outline">{sym}</Badge>)}
                </div>
              </div>

              <div className="bg-surface-container p-4">
                <CardTitle className="mb-3">Market Indices</CardTitle>
                <div className="space-y-2">
                  {marketIndices.map((index) => (
                    <div key={index.symbol} className="flex items-center justify-between p-3 border border-outline-variant/30">
                      <div>
                        <div className="font-mono text-sm font-semibold">{index.symbol}</div>
                        <div className="text-[11px] text-on-surface-variant">{index.name}</div>
                      </div>
                      <div className="text-right">
                        <div className="font-mono text-sm">{index.price}</div>
                        <div className={`text-[11px] font-mono ${index.positive ? "text-primary" : "text-error"}`}>{index.change}</div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </PageCell>
        </PageGrid>
      </div>

      {toast && (
        <div className="fixed bottom-6 right-6 z-50 bg-surface-container-highest border border-primary text-on-surface px-4 py-3 text-sm font-medium">{toast}</div>
      )}
    </div>
  );
}
