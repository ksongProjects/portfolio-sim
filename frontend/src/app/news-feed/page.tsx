"use client";

import { useEffect, useState } from "react";
import { PageCell, PageHeader } from "@/components/page-layout";
import { CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { DataTable } from "@/components/ui/data-table";
import { ColumnDef } from "@tanstack/react-table";
import { Clock, ExternalLink, Search, Bookmark, RefreshCw, Plus, Rss, Check, Video, Play, Loader, Trash2 } from "lucide-react";
import { useNews, type NewsArticle, type YouTubeVideo } from "@/hooks/useNews";
import { useRSSFeeds, type RSSFeed } from "@/hooks/useRSSFeeds";

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
  const { articles, channels, latestVideos, loading, fetchNews, fetchLatestVideos, summarizeVideos, searchChannels, addChannel, deleteChannel } = useNews();
  const [search, setSearch] = useState("");
  const [saved, setSaved] = useState<string[]>([]);
  const [toast, setToast] = useState<string | null>(null);
  const [showAddFeed, setShowAddFeed] = useState(false);
  const [newFeedName, setNewFeedName] = useState("");
  const [newFeedUrl, setNewFeedUrl] = useState("");
const { feeds, loading: feedsLoading, addFeed, scrapeFeeds, scrapeFeed, deleteFeed, refetchFeeds } = useRSSFeeds();
  const [selectedChannel, setSelectedChannel] = useState<string>("");
  const [selectedVideoIds, setSelectedVideoIds] = useState<Set<string>>(new Set());
  const [summarizingIds, setSummarizingIds] = useState<Set<string>>(new Set());
  const [showDetailModal, setShowDetailModal] = useState(false);
  const [detailItem, setDetailItem] = useState<NewsArticle | null>(null);
  const [showAddChannel, setShowAddChannel] = useState(false);
  const [newChannelId, setNewChannelId] = useState("");
  const [newChannelName, setNewChannelName] = useState("");
  const [channelSearch, setChannelSearch] = useState("");
  const [channelResults, setChannelResults] = useState<{id: string; name: string; handle: string}[]>([]);
  const [searchingChannels, setSearchingChannels] = useState(false);
  const [deleteChannelTarget, setDeleteChannelTarget] = useState<{id: string; name: string} | null>(null);
  const [activeMainTab, setActiveMainTab] = useState<"news" | "youtube" | "rss">("news");
  const [feedToDelete, setFeedToDelete] = useState<RSSFeed | null>(null);
  const [channelVideoCount, setChannelVideoCount] = useState<Record<string, number>>({});
  const [refreshingChannels, setRefreshingChannels] = useState<Set<string>>(new Set());
  const [selectedFeed, setSelectedFeed] = useState<string>("");
  const [refreshingFeeds, setRefreshingFeeds] = useState<Set<string>>(new Set());
  const [selectedArticleIds, setSelectedArticleIds] = useState<Set<string>>(new Set());

  const videoColumns: ColumnDef<YouTubeVideo>[] = [
    {
      id: "select",
      size: 40,
      header: "",
      cell: ({ row }) => {
        const video = row.original;
        const isSelected = selectedVideoIds.has(video.id);
        const isAnalyzed = articles.some(a => a.SourceType === "youtube" && a.URL.includes(video.id));
        return (
          <button
            onClick={(e) => { e.stopPropagation(); !isAnalyzed && toggleVideoSelection(video.id); }}
            disabled={isAnalyzed}
            className={`h-5 w-5 rounded border flex items-center justify-center transition-colors ${
              isSelected ? "bg-primary border-primary" : isAnalyzed ? "bg-primary/30 border-primary/30" : "border-outline hover:border-primary"
            }`}
          >
            {isSelected && <Check className="h-3 w-3 text-on-primary" />}
            {isAnalyzed && <Check className="h-3 w-3 text-primary" />}
          </button>
        );
      },
    },
    {
      accessorKey: "title",
      header: "Title",
      size: 300,
      cell: ({ row }) => <div className="font-medium">{row.original.title}</div>,
    },
    {
      accessorKey: "published_at",
      header: "Published",
      size: 120,
      cell: ({ row }) => <span className="text-on-surface-variant">{timeAgo(row.original.published_at)}</span>,
    },
    {
      id: "duration",
      header: "Duration",
      size: 80,
      cell: () => <span className="text-on-surface-variant">-</span>,
    },
    {
      id: "keywords",
      header: "Keywords",
      size: 200,
      cell: ({ row }) => {
        const keywords = row.original.description?.split('\n').find(l => l.toLowerCase().startsWith('keywords:'))?.replace(/^keywords:\s*/i, '') || '';
        return (
          <div className="flex flex-wrap gap-1">
            {keywords.split(',').map((kw, i) => kw.trim() && (
              <Badge key={i} variant="outline" className="text-xs">{kw.trim()}</Badge>
            ))}
          </div>
        );
      },
    },
  ];

  const articleColumns: ColumnDef<NewsArticle>[] = [
    {
      id: "select",
      size: 40,
      header: "",
      cell: ({ row }) => {
        const isSelected = selectedArticleIds.has(row.original.ID);
        return (
          <button
            onClick={(e) => { e.stopPropagation(); toggleArticleSelection(row.original.ID); }}
            className={`h-5 w-5 rounded border flex items-center justify-center transition-colors ${
              isSelected ? "bg-primary border-primary" : "border-outline hover:border-primary"
            }`}
          >
            {isSelected && <Check className="h-3 w-3 text-on-primary" />}
          </button>
        );
      },
    },
    {
      accessorKey: "Title",
      header: "Title",
      cell: ({ row }) => <div className="font-medium">{row.original.Title}</div>,
    },
    {
      accessorKey: "PublishedAt",
      header: "Published",
      size: 120,
      cell: ({ row }) => <span className="text-on-surface-variant">{timeAgo(row.original.PublishedAt)}</span>,
    },
    {
      accessorKey: "TickerSymbols",
      header: "Tickers",
      size: 200,
      cell: ({ row }) => (
        <div className="flex flex-wrap gap-1">
          {row.original.TickerSymbols?.map((sym) => (
            <Badge key={sym} variant="outline" className="text-xs">{sym}</Badge>
          ))}
        </div>
      ),
    },
  ];

  useEffect(() => {
    if (selectedChannel) {
      fetchLatestVideos(selectedChannel);
    }
  }, [selectedChannel, fetchLatestVideos]);

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
    setTimeout(() => fetchNews(), 2000);
  };

  const showToast = (msg: string) => {
    setToast(msg);
    setTimeout(() => setToast(null), 3000);
  };

  const toggleSaved = (id: string) => {
    setSaved((prev) => (prev.includes(id) ? prev.filter((s) => s !== id) : [...prev, id]));
  };

  const toggleVideoSelection = (videoId: string) => {
    setSelectedVideoIds((prev) => {
      const next = new Set(prev);
      if (next.has(videoId)) next.delete(videoId);
      else next.add(videoId);
      return next;
    });
  };

  const toggleArticleSelection = (articleId: string) => {
    setSelectedArticleIds((prev) => {
      const next = new Set(prev);
      if (next.has(articleId)) next.delete(articleId);
      else next.add(articleId);
      return next;
    });
  };

  const handleSummarizeSelected = async () => {
    if (selectedVideoIds.size === 0) return;
    const selectedVideos = latestVideos
      .filter((video) => selectedVideoIds.has(video.id))
      .map((video) => ({ video_id: video.id, title: video.title }));
    if (selectedVideos.length === 0) return;

    showToast("Summarizing videos...");
    setSummarizingIds(new Set(selectedVideoIds));
    try {
      const results = await summarizeVideos(selectedVideos);
      const failed = results.filter((result) => result.status !== "ok");
      const succeeded = results.length - failed.length;
      setSelectedVideoIds(new Set(failed.map((result) => result.video_id)));
      showToast(failed.length > 0 ? `${succeeded} summarized, ${failed.length} failed` : "Videos summarized");
    } catch {
      showToast("Failed to summarize videos");
    } finally {
      setSummarizingIds(new Set());
    }
  };

  const handleAddChannel = async () => {
    if (!newChannelId || !newChannelName) return;

    const success = await addChannel(newChannelId, newChannelName);

    if (success) {
      const addedChannelId = newChannelId;
      setNewChannelId("");
      setNewChannelName("");
      setShowAddChannel(false);
      setChannelSearch("");
      setChannelResults([]);
      setSelectedChannel(addedChannelId);
      setSelectedVideoIds(new Set());
      showToast("Channel added");
    } else {
      showToast("Failed to add channel");
    }
  };

  const handleSearchChannels = async () => {
    if (!channelSearch.trim()) return;
    setSearchingChannels(true);
    try {
      const results = await searchChannels(channelSearch);
      setChannelResults(results);
    } finally {
      setSearchingChannels(false);
    }
  };

  const selectSearchResult = (ch: {id: string; name: string; handle: string}) => {
    setNewChannelId(ch.id);
    setNewChannelName(ch.name);
    setChannelResults([]);
    setChannelSearch("");
  };

const openDetail = (article: NewsArticle) => {
    setDetailItem(article);
    setShowDetailModal(true);
  };

  const filteredNews = articles.filter((n) => {
    const matchesSearch = search === "" || n.Title.toLowerCase().includes(search.toLowerCase()) || n.Source.toLowerCase().includes(search.toLowerCase()) || n.TickerSymbols.some(s => s.toLowerCase().includes(search.toLowerCase()));
    return matchesSearch;
  });

  return (
    <div className="flex flex-col h-full">
      <div className="px-6 pt-6 pb-4">
        <PageHeader title="News Feed" description="Real-time market news and sentiment analysis">
          <div className="flex gap-3 items-center">
            <Button variant="default" size="sm" onClick={handleRefresh} disabled={feedsLoading}>
              <RefreshCw className="h-4 w-4" /> Refresh
            </Button>
            <Button variant="default" size="sm" onClick={() => setShowAddFeed(true)}>
              <Plus className="h-4 w-4" /> Add Feed
            </Button>
          </div>
        </PageHeader>
      </div>

<div className="flex-1 px-6 pb-6 overflow-auto">
        <div className="flex gap-1 mb-4 border-b border-outline-variant/30">
          {(["news", "youtube", "rss"] as const).map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveMainTab(tab)}
              className={`px-4 py-2 text-sm font-medium capitalize transition-colors border-b-2 ${
                activeMainTab === tab ? "border-primary text-primary" : "border-transparent text-on-surface-variant hover:text-on-surface"
              }`}
            >
              {tab === "youtube" ? "YouTube Channels" : tab === "rss" ? "RSS Feeds" : "News"}
            </button>
          ))}
        </div>

        {activeMainTab === "news" && (
            <PageCell>
              <div className="flex items-center justify-between mb-3">
                <CardTitle>Latest News</CardTitle>
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-on-surface-variant" />
                  <Input placeholder="Search..." className="pl-9 w-[180px]" value={search} onChange={(e) => setSearch(e.target.value)} />
                </div>
              </div>

              <div className="space-y-2">
                {filteredNews.map((news) => (
                  <div key={news.ID} className="flex items-start gap-4 p-4 border border-outline-variant/30 hover:bg-surface-container-low transition-colors">
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-2">
                        <Badge variant={news.Sentiment === "bullish" ? "success" : news.Sentiment === "bearish" ? "error" : "secondary"}>{news.Sentiment}</Badge>
                        {news.SourceType === "youtube" ? <Video className="h-3 w-3 text-error" /> : <Rss className="h-3 w-3 text-on-surface-variant" />}
                        <span className="text-xs text-on-surface-variant">{news.Source}</span>
                        <span className="text-xs text-on-surface-variant flex items-center gap-1"><Clock className="h-3 w-3" />{timeAgo(news.PublishedAt)}</span>
                      </div>
                      <h3 className="text-sm font-medium mb-2">{news.Title}</h3>
                      <div className="flex items-center gap-2">
                        {news.TickerSymbols?.map((sym) => <Badge key={sym} variant="outline">{sym}</Badge>)}
                        <Button variant="ghost" size="sm" onClick={() => openDetail(news)}>
                          View Details
                        </Button>
                      </div>
                    </div>
                    <div className="flex flex-col items-end gap-2">
                      <a href={news.URL} target="_blank" rel="noopener noreferrer" className="text-on-surface-variant hover:text-primary transition-colors" onClick={(e) => e.stopPropagation()}>
                        <ExternalLink className="h-4 w-4" />
                      </a>
                      <button onClick={() => toggleSaved(news.ID)} className="text-on-surface-variant hover:text-primary transition-colors">
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
          )}

        {activeMainTab === "youtube" && (
          <PageCell>
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-3">
                <label className="text-sm font-medium">Channel:</label>
                <select
                  value={selectedChannel}
                  onChange={(e) => {
                    setSelectedChannel(e.target.value);
                    setSelectedVideoIds(new Set());
                  }}
                  className="h-9 px-3 pr-10 min-w-[200px] rounded border border-outline-variant bg-surface-container text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                >
                  <option value="">Select a channel...</option>
                  {channels.map((ch) => (
                    <option key={ch.id} value={ch.channel_id}>{ch.name}</option>
                  ))}
                </select>
                {selectedChannel && (
                  <>
                    <Input
                      type="number"
                      min="1"
                      max="50"
                      value={channelVideoCount[selectedChannel] || 10}
                      onChange={(e) => setChannelVideoCount(prev => ({ ...prev, [selectedChannel]: parseInt(e.target.value) || 10 }))}
                      className="w-20 h-9"
                    />
                    <Button
                      variant="secondary"
                      size="sm"
                      onClick={() => {
                        if (selectedChannel) {
                          setRefreshingChannels(prev => new Set(prev).add(selectedChannel));
                          fetchLatestVideos(selectedChannel).finally(() => {
                            setRefreshingChannels(prev => {
                              const next = new Set(prev);
                              next.delete(selectedChannel);
                              return next;
                            });
                          });
                        }
                      }}
                      disabled={refreshingChannels.has(selectedChannel)}
                    >
                      <RefreshCw className={`h-4 w-4 ${refreshingChannels.has(selectedChannel) ? "animate-spin" : ""}`} />
                    </Button>
                  </>
                )}
              </div>
              <Button variant="secondary" size="sm" onClick={() => setShowAddChannel(true)}>
                <Plus className="h-3 w-3 mr-1" /> Add Channel
              </Button>
            </div>

            {selectedChannel && (
              <div className="mt-4">
                <div className="flex items-center justify-between mb-3">
                  <span className="text-sm font-medium">{latestVideos.length} Videos</span>
                  <div className="flex gap-2">
                    {selectedVideoIds.size > 0 && (
                      <Button size="sm" onClick={handleSummarizeSelected} disabled={summarizingIds.size > 0}>
                        <Play className="h-4 w-4" /> Summarize {selectedVideoIds.size}
                      </Button>
                    )}
                  </div>
                </div>
                <DataTable
                  columns={videoColumns}
                  data={latestVideos}
                  loading={loading && latestVideos.length === 0}
                  emptyMessage="No videos found"
                  enablePagination={true}
                  enableColumnResizing={true}
                  maxHeight="400px"
                  pageSizes={[10, 20, 50]}
                />
              </div>
            )}

            {!selectedChannel && (
              <div className="text-center text-on-surface-variant text-sm py-12">
                Select a channel to view videos
              </div>
            )}
          </PageCell>
        )}

        {activeMainTab === "rss" && (
          <PageCell>
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-3">
                <label className="text-sm font-medium">Feed:</label>
                <select
                  value={selectedFeed}
                  onChange={(e) => {
                    setSelectedFeed(e.target.value);
                    setSelectedArticleIds(new Set());
                  }}
                  className="h-9 px-3 pr-10 min-w-[200px] rounded border border-outline-variant bg-surface-container text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                >
                  <option value="">Select a feed...</option>
                  {feeds.map((feed) => (
                    <option key={feed.id} value={feed.name}>{feed.name}</option>
                  ))}
                </select>
                {selectedFeed && (
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => {
                      setRefreshingFeeds(prev => new Set(prev).add(selectedFeed));
                      scrapeFeed(selectedFeed).finally(() => {
                        setTimeout(() => {
                          refetchFeeds();
                          setRefreshingFeeds(prev => {
                            const next = new Set(prev);
                            next.delete(selectedFeed);
                            return next;
                          });
                        }, 1500);
                      });
                    }}
                    disabled={refreshingFeeds.has(selectedFeed)}
                  >
                    <RefreshCw className={`h-4 w-4 ${refreshingFeeds.has(selectedFeed) ? "animate-spin" : ""}`} />
                  </Button>
                )}
              </div>
              <Button variant="secondary" size="sm" onClick={() => setShowAddFeed(true)}>
                <Plus className="h-3 w-3 mr-1" /> Add Feed
              </Button>
            </div>

            {selectedFeed && (
              <div className="mt-4">
                <div className="flex items-center justify-between mb-3">
                  <span className="text-sm font-medium">Articles</span>
                  <Button variant="secondary" size="sm" onClick={() => setActiveMainTab("news")}>
                    <Search className="h-4 w-4 mr-1" /> Search Articles
                  </Button>
                </div>
                <DataTable
                  columns={articleColumns}
                  data={articles.filter(a => a.Source === selectedFeed)}
                  loading={feedsLoading}
                  emptyMessage="No articles from this feed. Click &ldquo;Search Articles&rdquo; to view all articles, or refresh the feed."
                  enablePagination={true}
                  enableColumnResizing={true}
                  maxHeight="400px"
                  pageSizes={[10, 20, 50]}
                />
              </div>
            )}

            {!selectedFeed && (
              <div className="text-center text-on-surface-variant text-sm py-12">
                Select a feed to view articles
              </div>
            )}
          </PageCell>
        )}
      </div>

      {showAddFeed && (
        <Dialog open={showAddFeed} onOpenChange={(open) => !open && setShowAddFeed(false)}>
          <DialogContent className="w-96 max-w-[90vw]">
            <DialogHeader>
              <DialogTitle>Add RSS Feed</DialogTitle>
              <DialogDescription>Add a new RSS feed to monitor market news and updates.</DialogDescription>
            </DialogHeader>
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
          </DialogContent>
        </Dialog>
      )}

      {showAddChannel && (
        <Dialog open={showAddChannel} onOpenChange={(open) => !open && setShowAddChannel(false)}>
          <DialogContent className="w-[450px] max-w-[90vw]">
            <DialogHeader>
              <DialogTitle>Add YouTube Channel</DialogTitle>
              <DialogDescription>Search for and add a YouTube channel to monitor video content.</DialogDescription>
            </DialogHeader>
            <div className="space-y-4">
              <div>
                <label className="text-xs text-on-surface-variant mb-1 block">Search Channels</label>
                <div className="flex gap-2">
                  <Input placeholder="Search by name..." value={channelSearch} onChange={(e) => setChannelSearch(e.target.value)} onKeyDown={(e) => e.key === "Enter" && handleSearchChannels()} />
                  <Button variant="secondary" onClick={handleSearchChannels} disabled={searchingChannels || !channelSearch.trim()}>
                    {searchingChannels ? <Loader className="h-4 w-4 animate-spin" /> : <Search className="h-4 w-4" />}
                  </Button>
                </div>
              </div>
              {channelResults.length > 0 && (
                <div className="border border-outline-variant/30 max-h-48 overflow-y-auto">
                  {channelResults.map((ch) => (
                    <button key={ch.id} onClick={() => selectSearchResult(ch)} className="w-full flex items-center gap-2 p-2 text-left hover:bg-surface-container transition-colors border-b border-outline-variant/20 last:border-b-0">
                      <Video className="h-4 w-4 text-error" />
                      <div className="flex-1 min-w-0">
                        <div className="text-sm font-medium truncate">{ch.name}</div>
                        <div className="text-xs text-on-surface-variant">{ch.handle}</div>
                      </div>
                    </button>
                  ))}
                </div>
              )}
              <div className="border-t border-outline-variant/30 pt-4">
                <div className="text-xs text-on-surface-variant mb-2">Selected Channel</div>
                <div className="flex items-center gap-2 p-2 bg-surface-container rounded border border-outline-variant/30">
                  <Video className="h-4 w-4 text-error" />
                  <div className="flex-1 min-w-0">
                    <div className="text-sm font-medium truncate">{newChannelName || "None selected"}</div>
                    {newChannelId && <div className="text-xs text-on-surface-variant">{newChannelId}</div>}
                  </div>
                </div>
              </div>
              <Button onClick={handleAddChannel} disabled={!newChannelId || !newChannelName} className="w-full">
                Add Channel
              </Button>
            </div>
          </DialogContent>
        </Dialog>
      )}

      {showDetailModal && detailItem && (
        <Dialog open={showDetailModal} onOpenChange={(open) => !open && setShowDetailModal(false)}>
          <DialogContent className="w-full max-w-2xl max-h-[80vh] overflow-y-auto">
            <DialogHeader>
              <div className="flex items-center gap-2 mb-3">
                {detailItem.SourceType === "youtube" && <Video className="h-5 w-5 text-error" />}
                <Badge variant={detailItem.Sentiment === "bullish" ? "success" : detailItem.Sentiment === "bearish" ? "error" : "secondary"}>{detailItem.Sentiment}</Badge>
                <span className="text-xs text-on-surface-variant">{detailItem.Source}</span>
                <span className="text-xs text-on-surface-variant">&bull;</span>
                <span className="text-xs text-on-surface-variant">{timeAgo(detailItem.PublishedAt)}</span>
              </div>
              <DialogTitle className="text-lg font-semibold">{detailItem.Title}</DialogTitle>
              <DialogDescription>Article details and summary</DialogDescription>
            </DialogHeader>
            {detailItem.Summary && (
              <p className="text-sm text-on-surface-variant leading-relaxed mb-4">{detailItem.Summary}</p>
            )}
            {detailItem.Content && (
              <p className="text-sm text-on-surface-variant leading-relaxed mb-4 text-xs italic">{detailItem.Content.substring(0, 500)}...</p>
            )}
            <div className="flex gap-2 mb-4">
              {detailItem.TickerSymbols?.map((sym: string) => <Badge key={sym} variant="outline">{sym}</Badge>)}
            </div>
            <a href={detailItem.URL} target="_blank" rel="noopener noreferrer" className="flex items-center gap-2 text-primary hover:underline">
              {detailItem.SourceType === "youtube" ? "Watch on YouTube" : "Read full article"} <ExternalLink className="h-4 w-4" />
            </a>
          </DialogContent>
        </Dialog>
)}

      <ConfirmDialog
        open={deleteChannelTarget !== null}
        onOpenChange={(open) => !open && setDeleteChannelTarget(null)}
        title="Remove Channel"
        description={`Are you sure you want to remove ${deleteChannelTarget?.name}? This cannot be undone.`}
        confirmLabel="Remove"
        variant="destructive"
        onConfirm={async () => {
          if (deleteChannelTarget) {
            await deleteChannel(deleteChannelTarget.id);
            if (selectedChannel === deleteChannelTarget.id) {
              setSelectedChannel("");
            }
            setDeleteChannelTarget(null);
          }
        }}
      />

      <ConfirmDialog
        open={feedToDelete !== null}
        onOpenChange={(open) => !open && setFeedToDelete(null)}
        title="Remove RSS Feed"
        description={`Are you sure you want to remove ${feedToDelete?.name}? This cannot be undone.`}
        confirmLabel="Remove"
        variant="destructive"
        onConfirm={async () => {
          if (feedToDelete) {
            await deleteFeed(feedToDelete.id);
            setFeedToDelete(null);
          }
        }}
      />

      {toast && (
        <div className="fixed bottom-6 right-6 z-50 bg-surface-container-highest border border-primary text-on-surface px-4 py-3 text-sm font-medium">{toast}</div>
      )}
    </div>
  );
}
