"use client";

import { useState } from "react";
import { PageCell, PageHeader, MetricLabel, MetricValue } from "@/components/page-layout";
import { CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import {
  Key,
  Plug,
  Bell,
  Shield,
  Save,
  Eye,
  EyeOff,
  CheckCircle,
  XCircle,
  ExternalLink,
  Plus,
  Trash2,
  Rss,
  RefreshCw,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useProviders, ProviderConfig, ConnectionStatus, RSSFeed, MarketIndexConfig, ProviderSavePayload, ProviderValidationResult } from "@/hooks/useProviders";

type ProviderStatus = "connected" | "disconnected" | "error" | "expired" | "configured";

function ProviderCard({
  provider,
  onSave,
  onValidate,
  onRefresh,
  onDelete,
}: {
  provider: ProviderConfig;
  onSave: (providerId: string, apiKey: string, payload?: ProviderSavePayload) => Promise<boolean>;
  onValidate: (providerId: string, apiKey: string) => Promise<ProviderValidationResult>;
  onRefresh?: () => Promise<{ success: boolean; error?: string }>;
  onDelete?: (providerId: string) => Promise<boolean>;
}) {
  const [apiKey, setApiKey] = useState("");
  const [showSecret, setShowSecret] = useState(false);
  const [saved, setSaved] = useState(false);
  const [saving, setSaving] = useState(false);
  const [testResult, setTestResult] = useState<ProviderStatus | null>(null);
  const [refreshing, setRefreshing] = useState(false);
  const [refreshError, setRefreshError] = useState<string | null>(null);

  const handleSave = async () => {
    if (!apiKey) return;
    setSaving(true);
    setTestResult(null);
    setRefreshError(null);
    const validation = await onValidate(provider.provider_id, apiKey);
    setTestResult(validation.valid ? "connected" : "error");
    if (!validation.valid) {
      setSaving(false);
      return;
    }

    const saveKey = validation.save_key || apiKey;
    const payload: ProviderSavePayload = {
      validated: true,
    };
    if (provider.provider_id === "questrade") {
      payload.access_token = validation.access_token;
      payload.refresh_token = validation.refresh_token;
      payload.api_server = validation.api_server;
      payload.expires_in = validation.expires_in;
    }

    const savedOk = await onSave(provider.provider_id, saveKey, payload);
    if (savedOk) {
      setSaved(true);
      if (provider.provider_id === "questrade") {
        setApiKey("");
      }
      setTimeout(() => setSaved(false), 2000);
    } else {
      setTestResult("error");
    }
    setSaving(false);
  };

const handleRefresh = async () => {
    if (!onRefresh) return;
    setRefreshing(true);
    setRefreshError(null);
    const result = await onRefresh();
    if (!result.success) {
      setRefreshError(result.error || "Failed to refresh");
    }
    setRefreshing(false);
  };

  const handleDelete = async () => {
    if (!onDelete) return;
    if (!confirm(`Remove ${provider.name} configuration?`)) return;
    await onDelete(provider.provider_id);
  };

  const status: ProviderStatus = provider.token_expired
    ? "expired"
    : provider.is_connected
      ? "connected"
      : provider.validation_error
        ? "error"
        : provider.api_key_set
          ? "configured"
          : "disconnected";
  const statusVariant: Record<ProviderStatus, "success" | "error" | "warning" | "secondary"> = {
    connected: "success",
    error: "error",
    disconnected: "warning",
    expired: "error",
    configured: "secondary",
  };
  const hasStoredKey = provider.api_key_set;

  return (
    <div className="p-4 border border-outline-variant/30">
      <div className="flex items-start justify-between mb-3">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <CardTitle className="text-sm font-semibold">{provider.name}</CardTitle>
            <Badge variant={statusVariant[status]}>{status}</Badge>
            {hasStoredKey && !provider.token_expired && (
              <Badge variant="secondary" className="gap-1">
                <Key className="h-3 w-3" />
                Key saved
              </Badge>
            )}
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
        <div className="space-y-1">
          <label
            htmlFor={`provider-api-key-${provider.provider_id}`}
            className="text-xs text-on-surface-variant"
          >
            API Key
          </label>
          <div className="relative">
            <Input
              id={`provider-api-key-${provider.provider_id}`}
              type={showSecret ? "text" : "password"}
              placeholder="sk_xxxxxxxxxxxx"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              className="pr-10"
            />
            <button
              type="button"
              onClick={() => setShowSecret(!showSecret)}
              className="absolute inset-y-0 right-0 flex items-center pr-3 text-on-surface-variant hover:text-on-surface transition-colors"
            >
              {showSecret ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
            </button>
          </div>
        </div>
        {testResult && (
          <div className="flex items-center gap-2 text-sm">
            {testResult === "connected" ? <CheckCircle className="h-4 w-4 text-primary" /> : <XCircle className="h-4 w-4 text-error" />}
            <span className={testResult === "connected" ? "text-primary" : "text-error"}>
              {testResult === "connected" ? "Connection successful" : "Connection failed"}
            </span>
          </div>
        )}
        {provider.validation_error && !testResult && (
          <div className="flex items-center gap-2 text-sm text-error">
            <XCircle className="h-4 w-4" />
            <span>{provider.validation_error}</span>
          </div>
        )}
{refreshError && (
          <div className="flex items-center gap-2 text-sm text-error">
            <XCircle className="h-4 w-4" />
            <span>{refreshError}</span>
          </div>
        )}
        <div className="flex gap-2">
          <Button variant="default" size="sm" onClick={handleSave} disabled={!apiKey || saving}>
            <Save className="h-4 w-4" />
            {saving ? "Saving..." : saved ? "Saved!" : "Save"}
          </Button>
          {provider.provider_id === "questrade" && onRefresh && (
            <Button variant="warning" size="sm" onClick={handleRefresh} disabled={refreshing}>
              <RefreshCw className={cn("h-4 w-4", refreshing && "animate-spin")} />
              {refreshing ? "Refreshing..." : "Refresh Token"}
            </Button>
          )}
          {hasStoredKey && onDelete && (
            <Button variant="warning" size="sm" onClick={handleDelete}>
              <Trash2 className="h-4 w-4" />
              Remove
            </Button>
          )}
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

function AddRSSFeedForm({ onAdd }: { onAdd: (name: string, url: string) => Promise<boolean> }) {
  const [name, setName] = useState("");
  const [url, setUrl] = useState("");
  const [adding, setAdding] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name || !url) return;
    setAdding(true);
    await onAdd(name, url);
    setName("");
    setUrl("");
    setAdding(false);
  };

  return (
    <form onSubmit={handleSubmit} className="flex flex-col sm:flex-row gap-2 items-end">
      <div className="flex-1 w-full">
        <Input
          label="Name"
          placeholder="e.g., Yahoo Finance"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
      </div>
      <div className="flex-[2] w-full">
        <Input
          label="Feed URL"
          placeholder="https://example.com/feed.xml"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
        />
      </div>
      <Button type="submit" variant="default" size="sm" disabled={adding || !name || !url}>
        <Plus className="h-4 w-4" />
        {adding ? "Adding..." : "Add"}
      </Button>
    </form>
  );
}

function MarketIndicesCard({
  indices,
  loading,
  onSave,
}: {
  indices: MarketIndexConfig[];
  loading: boolean;
  onSave: (indices: MarketIndexConfig[]) => Promise<boolean>;
}) {
  const [draft, setDraft] = useState<MarketIndexConfig[]>(() => indices);
  const [saving, setSaving] = useState(false);

  const hasBlankRows = draft.some((index) => !index.symbol.trim());
  const hasChanges =
    JSON.stringify(draft) !== JSON.stringify(indices);

  const updateIndex = (target: number, field: keyof MarketIndexConfig, value: string) => {
    setDraft((current) =>
      current.map((index, indexPosition) =>
        indexPosition === target ? { ...index, [field]: value } : index
      )
    );
  };

  const handleAdd = () => {
    setDraft((current) => [...current, { symbol: "", name: "" }]);
  };

  const handleRemove = (target: number) => {
    setDraft((current) => current.filter((_, index) => index !== target));
  };

  const handleSave = async () => {
    setSaving(true);
    await onSave(draft);
    setSaving(false);
  };

  return (
    <div>
      <div className="mb-4 flex items-start justify-between gap-4">
        <div>
          <CardTitle>Market Indices</CardTitle>
          <div className="mt-1 text-[11px] text-on-surface-variant">
            Controls which major market indices appear on the dashboard.
          </div>
        </div>
      </div>

      <div className="space-y-3">
{draft.map((index, indexPosition) => (
          <div
            key={`index-${indexPosition}`}
            className="grid grid-cols-1 gap-2 border border-outline-variant/30 p-3 md:grid-cols-[140px_minmax(0,1fr)_auto]"
          >
            <Input
              label="Symbol"
              placeholder="SPY"
              value={index.symbol}
              onChange={(e) => updateIndex(indexPosition, "symbol", e.target.value.toUpperCase())}
              className="font-mono uppercase"
            />
            <Input
              label="Name"
              placeholder="S&P 500"
              value={index.name}
              onChange={(e) => updateIndex(indexPosition, "name", e.target.value)}
            />
            <div className="flex items-end">
              <Button type="button" variant="ghost" size="icon" onClick={() => handleRemove(indexPosition)}>
                <Trash2 className="h-4 w-4 text-error" />
              </Button>
            </div>
          </div>
        ))}

        {draft.length === 0 && !loading && (
          <div className="border border-outline-variant/30 p-4 text-sm text-on-surface-variant">
            No indices configured. Add at least one symbol to show market context on the dashboard.
          </div>
        )}
      </div>

      <div className="mt-4 flex flex-wrap gap-2">
        <Button type="button" variant="secondary" size="sm" onClick={handleAdd}>
          <Plus className="h-4 w-4" />
          Add Index
        </Button>
        <Button
          type="button"
          variant="default"
          size="sm"
          onClick={handleSave}
          disabled={saving || hasBlankRows || !hasChanges}
        >
          <Save className="h-4 w-4" />
          {saving ? "Saving..." : "Save List"}
        </Button>
      </div>
    </div>
  );
}

export default function SettingsPage() {
const {
    providers,
    connections,
    rssFeeds,
    marketIndexSettings,
    loading,
    refresh,
    saveProviderKey,
    validateProviderKey,
    addRSSFeed,
    deleteRSSFeed,
    refreshQuestradeToken,
    saveMarketIndexSettings,
    deleteProviderKey,
  } = useProviders();

  const [notificationSettings, setNotificationSettings] = useState({
    tradeExecuted: true,
    strategySignal: true,
    priceAlert: false,
    systemError: true,
  });

  const handleNotificationChange = (key: keyof typeof notificationSettings, value: boolean) => {
    setNotificationSettings((prev) => ({ ...prev, [key]: value }));
  };

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
        <div className="grid grid-cols-2 md:grid-cols-4 gap-px bg-outline-variant mb-4">
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
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-px bg-outline-variant">
          <PageCell className="lg:col-span-2">
            <div className="space-y-6">
              <div>
                <CardTitle className="mb-4">Market Data Providers</CardTitle>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                  {marketDataProviders.map((p) => (
                    <ProviderCard key={p.provider_id} provider={p} onSave={saveProviderKey} onValidate={validateProviderKey} onRefresh={p.provider_id === "questrade" ? refreshQuestradeToken : undefined} onDelete={deleteProviderKey} />
                  ))}
                </div>
              </div>

              <div>
                <CardTitle className="mb-4">AI / Video Providers</CardTitle>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  {youtubeProvider && (
                    <ProviderCard provider={youtubeProvider} onSave={saveProviderKey} onValidate={validateProviderKey} onDelete={deleteProviderKey} />
                  )}
                  {geminiProvider && (
                    <ProviderCard provider={geminiProvider} onSave={saveProviderKey} onValidate={validateProviderKey} onDelete={deleteProviderKey} />
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

              <div>
                <MarketIndicesCard
                  key={marketIndexSettings.map((index) => `${index.symbol}:${index.name}`).join("|")}
                  indices={marketIndexSettings}
                  loading={loading}
                  onSave={saveMarketIndexSettings}
                />
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
                <div className="flex items-center justify-between p-3 border border-outline-variant/30">
                  <span className="text-sm">Trade executed</span>
                  <Switch
                    checked={notificationSettings.tradeExecuted}
                    onCheckedChange={(v) => handleNotificationChange("tradeExecuted", v)}
                  />
                </div>
                <div className="flex items-center justify-between p-3 border border-outline-variant/30">
                  <span className="text-sm">Strategy signal</span>
                  <Switch
                    checked={notificationSettings.strategySignal}
                    onCheckedChange={(v) => handleNotificationChange("strategySignal", v)}
                  />
                </div>
                <div className="flex items-center justify-between p-3 border border-outline-variant/30">
                  <span className="text-sm">Price alert</span>
                  <Switch
                    checked={notificationSettings.priceAlert}
                    onCheckedChange={(v) => handleNotificationChange("priceAlert", v)}
                  />
                </div>
                <div className="flex items-center justify-between p-3 border border-outline-variant/30">
                  <span className="text-sm">System error</span>
                  <Switch
                    checked={notificationSettings.systemError}
                    onCheckedChange={(v) => handleNotificationChange("systemError", v)}
                  />
                </div>
              </div>
            </div>
          </PageCell>
        </div>
      </div>
    </div>
  );
}
