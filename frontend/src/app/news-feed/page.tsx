"use client";

import { useEffect, useState } from "react";
import { PageGrid, PageCell, PageHeader } from "@/components/page-layout";
import { CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Clock, ExternalLink, Search, Bookmark, RefreshCw, Plus, X, Rss, Check, Video, Play, Loader } from "lucide-react";
import { useNews, type NewsArticle } from "@/hooks/useNews";
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
  const { articles, channels, latestVideos, loading, fetchNews, fetchLatestVideos, analyzeVideo, searchChannels, addChannel } = useNews();
  const [search, setSearch] = useState("");
  const [saved, setSaved] = useState<string[]>([]);
  const [toast, setToast] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<"all" | "bullish" | "bearish" | "neutral">("all");
  const [selectedFeeds, setSelectedFeeds] = useState<string[]>([]);
  const [showAddFeed, setShowAddFeed] = useState(false);
  const [newFeedName, setNewFeedName] = useState("");
  const [newFeedUrl, setNewFeedUrl] = useState("");
  const { feeds, loading: feedsLoading, addFeed, scrapeFeeds } = useRSSFeeds();
  const [showVideoSection, setShowVideoSection] = useState(false);
  const [selectedChannel, setSelectedChannel] = useState<string>("");
  const [selectedVideoIds, setSelectedVideoIds] = useState<Set<string>>(new Set());
  const [analyzingIds, setAnalyzingIds] = useState<Set<string>>(new Set());
const [showDetailModal, setShowDetailModal] = useState(false);
  const [detailItem, setDetailItem] = useState<NewsArticle | null>(null);
  const [showAddChannel, setShowAddChannel] = useState(false);
  const [newChannelId, setNewChannelId] = useState("");
  const [newChannelName, setNewChannelName] = useState("");
  const [channelSearch, setChannelSearch] = useState("");
  const [channelResults, setChannelResults] = useState<{id: string; name: string; handle: string}[]>([]);
  const [searchingChannels, setSearchingChannels] = useState(false);

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

  const toggleFeedFilter = (feedName: string) => {
    setSelectedFeeds((prev) =>
      prev.includes(feedName) ? prev.filter((f) => f !== feedName) : [...prev, feedName]
    );
  };

  const toggleVideoSelection = (videoId: string) => {
    setSelectedVideoIds((prev) => {
      const next = new Set(prev);
      if (next.has(videoId)) next.delete(videoId);
      else next.add(videoId);
      return next;
    });
  };

  const handleAnalyzeSelected = async () => {
    if (selectedVideoIds.size === 0) return;
    showToast("Analyzing videos...");
    setAnalyzingIds(new Set(selectedVideoIds));
    for (const vid of selectedVideoIds) {
      const video = latestVideos.find((v) => v.id === vid);
      if (video) {
        await analyzeVideo(vid, video.title);
      }
    }
    setAnalyzingIds(new Set());
    setSelectedVideoIds(new Set());
    showToast("Videos analyzed successfully");
  };

  const handleAddChannel = async () => {
    if (!newChannelId || !newChannelName) return;

    const success = await addChannel(newChannelId, newChannelName);

    if (success) {
      setNewChannelId("");
      setNewChannelName("");
      setShowAddChannel(false);
      setChannelSearch("");
      setChannelResults([]);
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
    const matchesSearch = n.Title.toLowerCase().includes(search.toLowerCase()) || n.Source.toLowerCase().includes(search.toLowerCase());
    const matchesTab = activeTab === "all" || n.Sentiment === activeTab;
    const matchesFeeds = selectedFeeds.length === 0 || selectedFeeds.includes(n.Source);
    return matchesSearch && matchesTab && matchesFeeds;
  });

  return (
    <div className="flex flex-col h-full">
      <div className="px-6 pt-6 pb-4">
        <PageHeader title="News Feed" description="Real-time market news and sentiment analysis">
          <div className="flex gap-3 items-center">
            <Button variant="default" size="sm" onClick={handleRefresh} disabled={feedsLoading}>
              <RefreshCw className="h-4 w-4" /> Refresh
            </Button>
            <Button variant="secondary" size="sm" onClick={() => setShowVideoSection(!showVideoSection)}>
              <Video className="h-4 w-4" /> {showVideoSection ? "Hide Videos" : "YouTube Videos"}
            </Button>
            <Button variant="default" size="sm" onClick={() => setShowAddFeed(true)}>
              <Plus className="h-4 w-4" /> Add Feed
            </Button>
          </div>
        </PageHeader>
      </div>

      <div className="flex-1 px-6 pb-6 overflow-auto">
        {showVideoSection && (
          <PageGrid className="mb-6" style={{ gridTemplateColumns: "1fr 2fr" }}>
            <PageCell>
              <CardTitle className="mb-3">YouTube Channels</CardTitle>
              <div className="space-y-2">
                {channels.map((ch) => (
                  <button
                    key={ch.id}
                    onClick={() => setSelectedChannel(ch.channel_id)}
                    className={`w-full flex items-center gap-2 p-2 rounded border transition-colors text-left ${
                      selectedChannel === ch.channel_id ? "bg-primary/10 border-primary" : "border-outline-variant hover:bg-surface-container"
                    }`}
                  >
                    <Video className="h-4 w-4 text-error" />
                    <span className="text-sm">{ch.name}</span>
                  </button>
                ))}
              </div>
              <Button variant="secondary" size="sm" onClick={() => setShowAddChannel(true)} className="w-full mt-3">
                <Plus className="h-3 w-3 mr-1" /> Add Channel
              </Button>
            </PageCell>
            <PageCell>
              <CardTitle>Latest Videos</CardTitle>
              <div className="flex items-center justify-between mb-3">
                <span className="text-xs text-on-surface-variant">{latestVideos.length} videos</span>
                <div className="flex gap-2">
                  {selectedVideoIds.size > 0 && (
                    <Button size="sm" onClick={handleAnalyzeSelected} disabled={analyzingIds.size > 0}>
                      <Play className="h-4 w-4" /> Analyze {selectedVideoIds.size}
                    </Button>
                  )}
                  <Button variant="ghost" size="sm" onClick={() => setSelectedVideoIds(new Set(latestVideos.filter(v => !videos.some(x => x.youtube_id === v.id)).map(v => v.id)))}>
                    Select All
                  </Button>
                </div>
              </div>
              <div className="border border-outline-variant/30">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-outline-variant/30 bg-surface-container">
                      <th className="w-10 p-2"></th>
                      <th className="text-left p-2 text-xs font-semibold uppercase tracking-wider text-on-surface-variant">Title</th>
                      <th className="text-left p-2 text-xs font-semibold uppercase tracking-wider text-on-surface-variant w-32">Channel</th>
                      <th className="text-left p-2 text-xs font-semibold uppercase tracking-wider text-on-surface-variant w-24">Published</th>
                      <th className="text-left p-2 text-xs font-semibold uppercase tracking-wider text-on-surface-variant w-20">Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    {loading && latestVideos.length === 0 ? (
                      <tr><td colSpan={5} className="p-4 text-center text-on-surface-variant"><Loader className="h-4 w-4 animate-spin inline mr-2" />Loading...</td></tr>
                    ) : latestVideos.map((video) => {
                      const isSelected = selectedVideoIds.has(video.id);
                      const isAnalyzing = analyzingIds.has(video.id);
                      const isStored = videos.some((v) => v.youtube_id === video.id);
                      return (
                        <tr key={video.id} className="border-b border-outline-variant/20 hover:bg-surface-container-low">
                          <td className="p-2 text-center">
                            <button
                              onClick={() => !isAnalyzing && !isStored && toggleVideoSelection(video.id)}
                              disabled={isAnalyzing || isStored}
                              className={`h-5 w-5 rounded border flex items-center justify-center transition-colors ${
                                isSelected ? "bg-primary border-primary" : isStored ? "bg-primary/30 border-primary/30" : "border-outline hover:border-primary"
                              }`}
                            >
                              {isSelected && <Check className="h-3 w-3 text-on-primary" />}
                              {isStored && <Check className="h-3 w-3 text-primary" />}
                            </button>
                          </td>
                          <td className="p-2">
                            <div className="font-medium truncate max-w-[300px]">{video.title}</div>
                          </td>
                          <td className="p-2 text-on-surface-variant text-xs">{video.channel_name}</td>
                          <td className="p-2 text-on-surface-variant text-xs">{timeAgo(video.published_at)}</td>
                          <td className="p-2">
                            {isAnalyzing && <Badge variant="warning" className="text-xs"><Loader className="h-3 w-3 animate-spin inline mr-1"/>Analyzing</Badge>}
                            {isStored && <Badge variant="success" className="text-xs">Analyzed</Badge>}
                          </td>
                        </tr>
                      );
                    })}
                    {selectedChannel && latestVideos.length === 0 && !loading && (
                      <tr><td colSpan={5} className="p-4 text-center text-on-surface-variant">No videos found</td></tr>
                    )}
                    {!selectedChannel && (
                      <tr><td colSpan={5} className="p-4 text-center text-on-surface-variant">Select a channel to view videos</td></tr>
                    )}
                  </tbody>
                </table>
              </div>
            </PageCell>
          </PageGrid>
        )}
        <PageGrid style={{ gridTemplateColumns: "2fr 1fr" }}>
          <PageCell>
            <div className="flex items-center justify-between mb-4">
              <CardTitle>Latest News</CardTitle>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-on-surface-variant" />
                <Input placeholder="Search news..." className="pl-9 w-[200px]" value={search} onChange={(e) => setSearch(e.target.value)} />
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
                <div key={news.ID} className="flex items-start gap-4 p-4 border border-outline-variant/30 hover:bg-surface-container-low transition-colors cursor-pointer" onClick={() => openDetail(news)}>
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
                    <a href={news.URL} target="_blank" rel="noopener noreferrer" className="text-on-surface-variant hover:text-primary transition-colors" onClick={(e) => e.stopPropagation()}>
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
            <CardTitle className="mb-3">Filter by Feed</CardTitle>
            {feedsLoading ? (
              <div className="text-on-surface-variant text-sm">Loading...</div>
            ) : feeds.length === 0 ? (
              <div className="text-on-surface-variant text-sm">No feeds configured</div>
            ) : (
              <div className="flex flex-wrap gap-2">
                {feeds.map((feed) => {
                  const selected = selectedFeeds.includes(feed.name);
                  return (
                    <button
                      key={feed.id}
                      onClick={() => toggleFeedFilter(feed.name)}
                      className={`flex items-center gap-1.5 px-2 py-1 text-xs rounded border transition-colors ${
                        selected ? "bg-primary text-on-primary border-primary" : "bg-surface-container border-outline-variant hover:border-primary/50"
                      }`}
                    >
                      <Rss className="h-3 w-3" />
                      {feed.name}
                    </button>
                  );
                })}
              </div>
            )}
            {selectedFeeds.length > 0 && (
              <Button variant="ghost" size="sm" onClick={() => setSelectedFeeds([])} className="mt-2">
                Clear filters
              </Button>
            )}
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

      {showAddChannel && (
        <div className="fixed inset-0 z-50 bg-black/50 flex items-center justify-center">
          <div className="bg-surface-container-highest border border-outline rounded-lg p-6 w-[450px] max-w-[90vw]">
            <div className="flex items-center justify-between mb-4">
              <CardTitle>Add YouTube Channel</CardTitle>
              <button onClick={() => { setShowAddChannel(false); setChannelSearch(""); setChannelResults([]); setNewChannelId(""); setNewChannelName(""); }} className="text-on-surface-variant hover:text-on-surface">
                <X className="h-4 w-4" />
              </button>
            </div>
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
          </div>
        </div>
      )}

      {showDetailModal && detailItem && (
        <div className="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4" onClick={() => setShowDetailModal(false)}>
          <div className="bg-surface-container-highest border border-outline rounded-lg p-6 w-full max-w-2xl max-h-[80vh] overflow-y-auto" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center gap-2 mb-3">
              {detailItem.SourceType === "youtube" && <Video className="h-5 w-5 text-error" />}
              <Badge variant={detailItem.Sentiment === "bullish" ? "success" : detailItem.Sentiment === "bearish" ? "error" : "secondary"}>{detailItem.Sentiment}</Badge>
              <span className="text-xs text-on-surface-variant">{detailItem.Source}</span>
              <span className="text-xs text-on-surface-variant">&bull;</span>
              <span className="text-xs text-on-surface-variant">{timeAgo(detailItem.PublishedAt)}</span>
            </div>
            <h2 className="text-lg font-semibold mb-3">{detailItem.Title}</h2>
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
            <button onClick={() => setShowDetailModal(false)} className="absolute top-4 right-4 text-on-surface-variant hover:text-on-surface">
              <X className="h-5 w-5" />
            </button>
          </div>
        </div>
      )}

      {toast && (
        <div className="fixed bottom-6 right-6 z-50 bg-surface-container-highest border border-primary text-on-surface px-4 py-3 text-sm font-medium">{toast}</div>
      )}
    </div>
  );
}
