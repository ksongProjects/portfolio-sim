"use client";

import { useEffect, useState } from "react";
import { PageGrid, PageCell, PageHeader, MetricLabel, MetricValue, MetricSubValue } from "@/components/page-layout";
import { CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Clock, TrendingUp, ExternalLink, Search, Filter, Bookmark, RefreshCw, Plus, X, Rss } from "lucide-react";
import { useNews } from "@/hooks/useNews";
import { useRSSFeeds } from "@/hooks/useRSSFeeds";

function timeAgo(dateStr: string): string {
  try {
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return date.toLocaleDateString();
  } catch {
    return dateStr;
  }
}

export default function NewsFeedPage() {
  const { articles, loading, fetchNews } = useNews();
  const [search, setSearch] = useState("");
  const [saved, setSaved] = useState<string[]>([]);
  const [toast, setToast] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<"all" | "bullish" | "bearish" | "neutral">("all");
  const [refreshInterval, setRefreshInterval] = useState(5);
  const [showAddFeed, setShowAddFeed] = useState(false);
  const [newFeedName, setNewFeedName] = useState("");
  const [newFeedUrl, setNewFeedUrl] = useState("");
  const { feeds, loading: feedsLoading, fetchFeeds, addFeed, deleteFeed, scrapeFeeds } = useRSSFeeds();

  useEffect(() => { fetchFeeds(); }, [fetchFeeds]);
  useEffect(() => { fetchNews(20); }, [fetchNews]);

  useEffect(() => {
    const intervalId = setInterval(() => fetchNews(20), refreshInterval * 60 * 1000);
    return () => clearInterval(intervalId);
  }, [refreshInterval, fetchNews]);

  const handleAddFeed = async () => {
    if (newFeedName && newFeedUrl) {
      const success = await addFeed(newFeedName, newFeedUrl);
      if (success) {
        setNewFeedName("");
        setNewFeedUrl("");
        setShowAddFeed(false);
        showToast("Feed added successfully");
      }
    }
  };

  const handleRefresh = async () => {
    showToast("Scraping feeds...");
    await scrapeFeeds();
    setTimeout(() => fetchNews(20), 2000);
  };

  const showToast = (msg: string) => {
    setToast(msg);
    setTimeout(() => setToast(null), 3000);
  };

  const toggleSaved = (id: string) => {
    setSaved((prev) => (prev.includes(id) ? prev.filter((s) => s !== id) : [...prev, id]));
  };

  const filteredNews = articles.filter((n) => {
    const matchesSearch = n.Title.toLowerCase().includes(search.toLowerCase()) || n.Source.toLowerCase().includes(search.toLowerCase());
    const matchesTab = activeTab === "all" || n.Sentiment === activeTab;
    return matchesSearch && matchesTab;
  });

  const bullishCount = articles.filter((n) => n.Sentiment === "bullish").length;
  const featured = articles[0] || null;

  return (
    <div className="flex flex-col h-full">
      <div className="px-6 pt-6 pb-4">
        <PageHeader title="News Feed" description="Real-time market news and sentiment analysis">
          <div className="flex gap-3 items-center">
            <Button variant="default" size="sm" onClick={handleRefresh} disabled={feedsLoading}>
              <RefreshCw className="h-4 w-4" /> Refresh
            </Button>
            <Button variant={saved.length > 0 ? "default" : "secondary"} size="sm" onClick={() => showToast(`Saved articles: ${saved.length}`)}>
              <Bookmark className="h-4 w-4" /> Saved ({saved.length})
            </Button>
            <Button variant="default" size="sm" onClick={() => setShowAddFeed(true)}>
              <Plus className="h-4 w-4" /> Add Feed
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
            <MetricValue>{articles.length}</MetricValue>
            <MetricSubValue positive>+{articles.length} articles</MetricSubValue>
          </PageCell>
          <PageCell>
            <MetricLabel>Bullish Sentiment</MetricLabel>
            <MetricValue highlight>{bullishCount}</MetricValue>
            <MetricSubValue>{articles.length > 0 ? Math.round((bullishCount / articles.length) * 100) : 0}% of coverage</MetricSubValue>
          </PageCell>
          <PageCell>
            <MetricLabel>Top Symbol</MetricLabel>
            <MetricValue>--</MetricValue>
            <MetricSubValue>No data</MetricSubValue>
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
                <div key={news.ID} className="flex items-start gap-4 p-4 border border-outline-variant/30 hover:bg-surface-container-low transition-colors cursor-pointer">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <Badge variant={news.Sentiment === "bullish" ? "success" : news.Sentiment === "bearish" ? "error" : "secondary"}>{news.Sentiment}</Badge>
                      <span className="text-xs text-on-surface-variant">{news.Source}</span>
                      <span className="text-xs text-on-surface-variant flex items-center gap-1"><Clock className="h-3 w-3" />{timeAgo(news.PublishedAt)}</span>
                    </div>
                    <h3 className="text-sm font-medium mb-2">{news.Title}</h3>
                    <div className="flex gap-2">
                      {news.TickerSymbols?.map((sym) => <Badge key={sym} variant="outline">{sym}</Badge>)}
                    </div>
                  </div>
                  <div className="flex flex-col items-end gap-2">
                    <a href={news.URL} target="_blank" rel="noopener noreferrer" className="text-on-surface-variant hover:text-primary transition-colors">
                      <ExternalLink className="h-4 w-4" />
                    </a>
                    <button onClick={(e) => { e.stopPropagation(); toggleSaved(news.ID); }} className="text-on-surface-variant hover:text-primary transition-colors">
                      <Bookmark className={`h-4 w-4 ${saved.includes(news.ID) ? "fill-primary text-primary" : ""}`} />
                    </button>
                  </div>
                </div>
              ))}
              {filteredNews.length === 0 && !loading && (
                <div className="text-center text-on-surface-variant text-sm py-8">No news articles available</div>
              )}
            </div>
          </PageCell>

          <PageCell>
            <div className="space-y-[1px]">
              {featured ? (
                <div className="bg-surface-container p-4">
                  <div className="flex items-center justify-between mb-3">
                    <CardTitle>Featured Story</CardTitle>
                    <Badge variant={featured.Sentiment === "bullish" ? "success" : featured.Sentiment === "bearish" ? "error" : "secondary"}>{featured.Sentiment}</Badge>
                  </div>
                  <h3 className="text-sm font-medium leading-snug mb-2">{featured.Title}</h3>
                  <p className="text-xs text-on-surface-variant leading-relaxed mb-3">{featured.Summary}</p>
                  <div className="flex items-center gap-2 text-xs text-on-surface-variant">
                    <span>{featured.Source}</span><span>&bull;</span><span>{timeAgo(featured.PublishedAt)}</span>
                  </div>
                  <div className="flex gap-2 mt-3">
                    {featured.TickerSymbols?.map((sym) => <Badge key={sym} variant="outline">{sym}</Badge>)}
                  </div>
                </div>
              ) : (
                <div className="bg-surface-container p-4">
                  <CardTitle>Featured Story</CardTitle>
                  <div className="text-on-surface-variant text-sm py-4">No featured article available</div>
                </div>
              )}

              <div className="bg-surface-container p-4">
                <CardTitle className="mb-3">RSS Feeds</CardTitle>
                {feedsLoading ? (
                  <div className="text-on-surface-variant text-sm">Loading...</div>
                ) : feeds.length === 0 ? (
                  <div className="text-on-surface-variant text-sm mb-3">No feeds configured</div>
                ) : (
                  <div className="space-y-2 mb-3">
                    {feeds.map((feed) => (
                      <div key={feed.id} className="flex items-center justify-between text-sm">
                        <div className="flex items-center gap-2">
                          <Rss className="h-3 w-3 text-on-surface-variant" />
                          <span className="text-on-surface">{feed.name}</span>
                        </div>
                        <button onClick={() => deleteFeed(feed.id)} className="text-on-surface-variant hover:text-error transition-colors">
                          <X className="h-3 w-3" />
                        </button>
                      </div>
                    ))}
                  </div>
                )}
                <Button variant="secondary" size="sm" onClick={() => setShowAddFeed(true)} className="w-full">
                  <Plus className="h-3 w-3 mr-1" /> Add Feed
                </Button>
              </div>
            </div>
          </PageCell>
        </PageGrid>
      </div>

      {showAddFeed && (
        <div className="fixed inset-0 z-50 bg-black/50 flex items-center justify-center">
          <div className="bg-surface-container-highest border border-outline rounded-lg p-6 w-96 max-w-[90vw]">
            <div className="flex items-center justify-between mb-4">
              <CardTitle>Add RSS Feed</CardTitle>
              <button onClick={() => setShowAddFeed(false)} className="text-on-surface-variant hover:text-on-surface">
                <X className="h-4 w-4" />
              </button>
            </div>
            <div className="space-y-4">
              <div>
                <label className="text-xs text-on-surface-variant mb-1 block">Feed Name</label>
                <Input placeholder="e.g., Reuters Markets" value={newFeedName} onChange={(e) => setNewFeedName(e.target.value)} />
              </div>
              <div>
                <label className="text-xs text-on-surface-variant mb-1 block">Feed URL</label>
                <Input placeholder="https://feeds.example.com/market.xml" value={newFeedUrl} onChange={(e) => setNewFeedUrl(e.target.value)} />
              </div>
              <Button onClick={handleAddFeed} disabled={!newFeedName || !newFeedUrl || feedsLoading} className="w-full">
                Add Feed
              </Button>
            </div>
          </div>
        </div>
      )}

      {toast && (
        <div className="fixed bottom-6 right-6 z-50 bg-surface-container-highest border border-primary text-on-surface px-4 py-3 text-sm font-medium">{toast}</div>
      )}
    </div>
  );
}