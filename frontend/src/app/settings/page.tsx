"use client";

import { useEffect, useState } from "react";
import { PageGrid, PageCell, PageHeader, MetricLabel, MetricValue } from "@/components/page-layout";
import { CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Key,
  Plug,
  Bell,
  Shield,
  Save,
  TestTube,
  Eye,
  EyeOff,
  CheckCircle,
  XCircle,
  ExternalLink,
  Plus,
  Trash2,
  Rss,
} from "lucide-react";
import { useProviders, ProviderConfig, ConnectionStatus, RSSFeed } from "@/hooks/useProviders";

type ProviderStatus = "connected" | "disconnected" | "error";

function ProviderCard({
  provider,
  onSave,
  onValidate,
}: {
  provider: ProviderConfig;
  onSave: (providerId: string, apiKey: string) => Promise<boolean>;
  onValidate: (providerId: string, apiKey: string) => Promise<{ valid: boolean; error?: string }>;
}) {
  const [apiKey, setApiKey] = useState("");
  const [showSecret, setShowSecret] = useState(false);
  const [saved, setSaved] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<ProviderStatus | null>(null);

  const handleSave = async () => {
    if (!apiKey) return;
    const success = await onSave(provider.provider_id, apiKey);
    if (success) {
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    }
  };

  const handleTest = async () => {
    if (!apiKey) return;
    setTesting(true);
    setTestResult(null);
    const result = await onValidate(provider.provider_id, apiKey);
    setTestResult(result.valid ? "connected" : "error");
    setTesting(false);
  };

  const status: ProviderStatus = provider.is_connected ? "connected" : provider.api_key_set ? "connected" : "disconnected";
  const statusVariant: Record<ProviderStatus, "success" | "error" | "warning"> = {
    connected: "success",
    error: "error",
    disconnected: "warning",
  };

  return (
    <div className="p-4 border border-outline-variant/30">
      <div className="flex items-start justify-between mb-3">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <CardTitle className="text-sm font-semibold">{provider.name}</CardTitle>
            <Badge variant={statusVariant[status]}>{status}</Badge>
          </div>
          <div className="text-[11px] text-on-surface-variant">{provider.description}</div>
        </div>
        {provider.docs_url && (
          <a href={provider.docs_url} target="_blank" rel="noopener noreferrer" className="text-on-surface-variant hover:text-primary transition-colors">
            <ExternalLink className="h-4 w-4" />
          </a>
        )}
      </div>
      <div className="space-y-3">
        <div className="relative">
          <Input
            label="API Key"
            type={showSecret ? "text" : "password"}
            placeholder="sk_xxxxxxxxxxxx"
            value={apiKey}
            onChange={(e) => setApiKey(e.target.value)}
          />
          <button
            type="button"
            onClick={() => setShowSecret(!showSecret)}
            className="absolute right-3 top-[34px] text-on-surface-variant hover:text-on-surface transition-colors"
          >
            {showSecret ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
          </button>
        </div>
        {testResult && (
          <div className="flex items-center gap-2 text-sm">
            {testResult === "connected" ? <CheckCircle className="h-4 w-4 text-primary" /> : <XCircle className="h-4 w-4 text-error" />}
            <span className={testResult === "connected" ? "text-primary" : "text-error"}>
              {testResult === "connected" ? "Connection successful" : "Connection failed"}
            </span>
          </div>
        )}
        <div className="flex gap-2">
          <Button variant="secondary" size="sm" onClick={handleTest} disabled={testing || !apiKey}>
            <TestTube className="h-4 w-4" />
            {testing ? "Testing..." : "Test"}
          </Button>
          <Button variant="default" size="sm" onClick={handleSave} disabled={!apiKey}>
            <Save className="h-4 w-4" />
            {saved ? "Saved!" : "Save"}
          </Button>
        </div>
      </div>
    </div>
  );
}

function ConnectionCard({ conn }: { conn: ConnectionStatus }) {
  return (
    <div className="flex items-center justify-between p-3 border border-outline-variant/30">
      <div className="flex items-center gap-3">
        <div className={`h-2 w-2 rounded-full ${conn.is_up ? "bg-primary" : "bg-error"}`} />
        <div>
          <div className="text-sm font-medium">{conn.name}</div>
          <div className="text-[11px] text-on-surface-variant">{conn.type}</div>
        </div>
      </div>
      <Badge variant={conn.is_up ? "success" : "error"}>{conn.is_up ? "connected" : "error"}</Badge>
    </div>
  );
}

function RSSFeedCard({
  feed,
  onDelete,
}: {
  feed: RSSFeed;
  onDelete: (feedId: string) => Promise<boolean>;
}) {
  const [deleting, setDeleting] = useState(false);

  const handleDelete = async () => {
    setDeleting(true);
    await onDelete(feed.id);
    setDeleting(false);
  };

  return (
    <div className="flex items-center justify-between p-3 border border-outline-variant/30">
      <div className="flex items-center gap-3">
        <Rss className="h-4 w-4 text-on-surface-variant" />
        <div>
          <div className="text-sm font-medium">{feed.name}</div>
          <div className="text-[11px] text-on-surface-variant truncate max-w-[200px]">{feed.url}</div>
        </div>
      </div>
      <Button variant="ghost" size="sm" onClick={handleDelete} disabled={deleting}>
        <Trash2 className="h-4 w-4 text-error" />
      </Button>
    </div>
  );
}

function AddRSSFeedForm({ onAdd }: { onAdd: (name: string, url: string, scrapeInterval: number) => Promise<boolean> }) {
  const [name, setName] = useState("");
  const [url, setUrl] = useState("");
  const [scrapeInterval, setScrapeInterval] = useState(60);
  const [adding, setAdding] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name || !url) return;
    setAdding(true);
    await onAdd(name, url, scrapeInterval);
    setName("");
    setUrl("");
    setScrapeInterval(60);
    setAdding(false);
  };

  return (
    <form onSubmit={handleSubmit} className="flex gap-2 items-end">
      <div className="flex-1">
        <Input
          label="Name"
          placeholder="e.g., Yahoo Finance"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
      </div>
      <div className="flex-[2]">
        <Input
          label="Feed URL"
          placeholder="https://example.com/feed.xml"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
        />
      </div>
      <div className="w-24">
        <Input
          label="Interval (min)"
          type="number"
          min={5}
          value={scrapeInterval}
          onChange={(e) => setScrapeInterval(parseInt(e.target.value) || 60)}
        />
      </div>
      <Button type="submit" variant="default" size="sm" disabled={adding || !name || !url}>
        <Plus className="h-4 w-4" />
        {adding ? "Adding..." : "Add"}
      </Button>
    </form>
  );
}

export default function SettingsPage() {
  const { providers, connections, rssFeeds, loading, refresh, saveProviderKey, validateProviderKey, addRSSFeed, deleteRSSFeed } = useProviders();

  useEffect(() => { refresh(); }, [refresh]);

  const connectedCount = connections.filter((c) => c.is_up).length;

  const marketDataProviders = providers.filter((p) => p.type === "market_data");
  const youtubeProvider = providers.find((p) => p.type === "youtube");
  const geminiProvider = providers.find((p) => p.type === "gemini");

  return (
    <div className="flex flex-col h-full">
      <div className="px-6 pt-6 pb-4">
        <PageHeader title="Settings" description="Manage API keys, connections, and platform configuration">
          <Button variant="default" size="sm" onClick={refresh}>
            <Plug className="h-4 w-4" /> Test All Connections
          </Button>
        </PageHeader>
      </div>

      <div className="flex-1 px-6 pb-6 overflow-auto">
        <PageGrid className="mb-4">
          <PageCell>
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 flex items-center justify-center bg-primary/10">
                <Key className="h-5 w-5 text-primary" />
              </div>
              <div>
                <MetricLabel>Providers</MetricLabel>
                <MetricValue>{providers.length}</MetricValue>
              </div>
            </div>
          </PageCell>
          <PageCell>
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 flex items-center justify-center bg-primary/10">
                <Plug className="h-5 w-5 text-primary" />
              </div>
              <div>
                <MetricLabel>Connections</MetricLabel>
                <MetricValue>{connectedCount}/{connections.length}</MetricValue>
              </div>
            </div>
          </PageCell>
          <PageCell>
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 flex items-center justify-center bg-surface-container-high">
                <Shield className="h-5 w-5 text-on-surface-variant" />
              </div>
              <div>
                <MetricLabel>Security</MetricLabel>
                <MetricValue>Keys encrypted</MetricValue>
              </div>
            </div>
          </PageCell>
          <PageCell>
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 flex items-center justify-center bg-surface-container-high">
                <Bell className="h-5 w-5 text-on-surface-variant" />
              </div>
              <div>
                <MetricLabel>RSS Feeds</MetricLabel>
                <MetricValue>{rssFeeds.length}</MetricValue>
              </div>
            </div>
          </PageCell>
        </PageGrid>

        <PageGrid style={{ gridTemplateColumns: "2fr 1fr" }}>
          <PageCell>
            <div className="space-y-6">
              <div>
                <CardTitle className="mb-4">Market Data Providers</CardTitle>
                <div className="grid grid-cols-2 gap-3">
                  {marketDataProviders.map((p) => (
                    <ProviderCard key={p.provider_id} provider={p} onSave={saveProviderKey} onValidate={validateProviderKey} />
                  ))}
                </div>
              </div>

              <div>
                <CardTitle className="mb-4">AI / Video Providers</CardTitle>
                <div className="grid grid-cols-2 gap-3">
                  {youtubeProvider && (
                    <ProviderCard provider={youtubeProvider} onSave={saveProviderKey} onValidate={validateProviderKey} />
                  )}
                  {geminiProvider && (
                    <ProviderCard provider={geminiProvider} onSave={saveProviderKey} onValidate={validateProviderKey} />
                  )}
                </div>
              </div>

              <div>
                <CardTitle className="mb-4">RSS Feeds</CardTitle>
                <div className="space-y-3">
                  {rssFeeds.map((feed) => (
                    <RSSFeedCard key={feed.id} feed={feed} onDelete={deleteRSSFeed} />
                  ))}
                  {rssFeeds.length === 0 && !loading && (
                    <div className="text-center text-on-surface-variant text-sm py-4">No RSS feeds configured</div>
                  )}
                </div>
                <div className="mt-4 pt-4 border-t border-outline-variant/30">
                  <AddRSSFeedForm onAdd={addRSSFeed} />
                </div>
              </div>
            </div>
          </PageCell>

          <PageCell>
            <CardTitle className="mb-4">Connections</CardTitle>
            <div className="space-y-3">
              {connections.map((conn) => (
                <ConnectionCard key={conn.id} conn={conn} />
              ))}
              {connections.length === 0 && !loading && (
                <div className="text-center text-on-surface-variant text-sm py-4">No connection data available</div>
              )}
            </div>

            <div className="mt-6 pt-5 border-t border-outline-variant/30">
              <CardTitle className="mb-4">Notifications</CardTitle>
              <div className="space-y-3">
                {[
                  { label: "Trade executed", enabled: true },
                  { label: "Strategy signal", enabled: true },
                  { label: "Price alert", enabled: false },
                  { label: "System error", enabled: true },
                ].map((item) => (
                  <div key={item.label} className="flex items-center justify-between p-3 border border-outline-variant/30">
                    <span className="text-sm">{item.label}</span>
                    <button
                      type="button"
                      className={`relative h-5 w-9 rounded-full transition-colors ${item.enabled ? "bg-primary" : "bg-surface-container-high"}`}
                    >
                      <span className={`absolute top-0.5 left-0.5 h-4 w-4 rounded-full bg-on-primary transition-transform ${item.enabled ? "translate-x-4" : ""}`} />
                    </button>
                  </div>
                ))}
              </div>
            </div>
          </PageCell>
        </PageGrid>
      </div>
    </div>
  );
}