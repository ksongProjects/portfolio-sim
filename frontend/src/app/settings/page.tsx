"use client";

import { useState } from "react";
import { PageCell, PageHeader } from "@/components/page-layout";
import { Card, CardHeader, CardTitle, CardContent, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge, StatusIndicator } from "@/components/ui/badge";
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
} from "lucide-react";

type ProviderStatus = "connected" | "disconnected" | "error";

interface Provider {
  id: string;
  name: string;
  description: string;
  fields: { key: string; label: string; placeholder: string; masked?: boolean }[];
  status: ProviderStatus;
  docsUrl?: string;
}

const providers: Provider[] = [
  {
    id: "polygon",
    name: "Polygon.io",
    description: "Real-time and historical market data",
    fields: [
      { key: "api_key", label: "API Key", placeholder: "sk_xxxxxxxxxxxx" },
      { key: "api_secret", label: "API Secret", placeholder: "xxxxxxxxxxxx", masked: true },
    ],
    status: "connected",
    docsUrl: "https://polygon.io/docs",
  },
  {
    id: "alpaca",
    name: "Alpaca",
    description: "Stock trading and market data API",
    fields: [
      { key: "api_key", label: "API Key", placeholder: "PKXXXXXXXXXX" },
      { key: "api_secret", label: "API Secret", placeholder: "xxxxxxxxxxxx", masked: true },
    ],
    status: "disconnected",
    docsUrl: "https://alpaca.markets/docs",
  },
  {
    id: "fmp",
    name: "Financial Modeling Prep",
    description: "Financial statements and fundamental data",
    fields: [
      { key: "api_key", label: "API Key", placeholder: "xxxxxxxxxxxx" },
    ],
    status: "disconnected",
    docsUrl: "https://site.financialmodelingprep.com/developers/docs",
  },
  {
    id: "newsapi",
    name: "News API",
    description: "Real-time news and sentiment analysis",
    fields: [
      { key: "api_key", label: "API Key", placeholder: "xxxxxxxxxxxx" },
    ],
    status: "error",
    docsUrl: "https://newsapi.org/docs",
  },
  {
    id: "fred",
    name: "FRED",
    description: "Federal Reserve economic data",
    fields: [
      { key: "api_key", label: "API Key", placeholder: "xxxxxxxxxxxx" },
    ],
    status: "disconnected",
    docsUrl: "https://fred.stlouisfed.org/docs/api/fred",
  },
];

const connections = [
  { id: "redis", name: "Redis", description: "Cache and real-time data", status: "connected" },
  { id: "postgres", name: "PostgreSQL", description: "Persistent storage", status: "connected" },
  { id: "websocket", name: "WebSocket", description: "Real-time streaming", status: "connected" },
];

function ProviderCard({ provider }: { provider: Provider }) {
  const [values, setValues] = useState<Record<string, string>>({});
  const [showSecret, setShowSecret] = useState<Record<string, boolean>>({});
  const [saved, setSaved] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<ProviderStatus | null>(null);

  const handleSave = () => {
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  const handleTest = async () => {
    setTesting(true);
    setTestResult(null);
    await new Promise((r) => setTimeout(r, 1200));
    const result: ProviderStatus = values[provider.fields[0]?.key] ? "connected" : "disconnected";
    setTestResult(result);
    setTesting(false);
  };

  const statusVariant: Record<ProviderStatus, "success" | "error" | "warning"> = {
    connected: "success",
    error: "error",
    disconnected: "warning",
  };

  return (
    <Card>
      <CardHeader>
        <div className="flex items-start justify-between">
          <div>
            <div className="flex items-center gap-2 mb-1">
              <CardTitle className="text-base font-semibold">{provider.name}</CardTitle>
              <Badge variant={statusVariant[provider.status]}>
                {provider.status}
              </Badge>
            </div>
            <CardDescription>{provider.description}</CardDescription>
          </div>
          {provider.docsUrl && (
            <a
              href={provider.docsUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="text-on-surface-variant hover:text-primary transition-colors"
            >
              <ExternalLink className="h-4 w-4" />
            </a>
          )}
        </div>
      </CardHeader>
      <CardContent className="pt-0">
        <div className="space-y-4">
          {provider.fields.map((field) => (
            <div key={field.key} className="relative">
              <Input
                label={field.label}
                type={showSecret[field.key] ? "text" : "password"}
                placeholder={field.placeholder}
                value={values[field.key] || ""}
                onChange={(e) =>
                  setValues((prev) => ({ ...prev, [field.key]: e.target.value }))
                }
              />
              <button
                type="button"
                onClick={() =>
                  setShowSecret((prev) => ({ ...prev, [field.key]: !prev[field.key] }))
                }
                className="absolute right-3 top-[34px] text-on-surface-variant hover:text-on-surface transition-colors"
              >
                {showSecret[field.key] ? (
                  <EyeOff className="h-4 w-4" />
                ) : (
                  <Eye className="h-4 w-4" />
                )}
              </button>
            </div>
          ))}
          {testResult && (
            <div className="flex items-center gap-2 text-sm">
              {testResult === "connected" ? (
                <CheckCircle className="h-4 w-4 text-primary" />
              ) : (
                <XCircle className="h-4 w-4 text-error" />
              )}
              <span className={testResult === "connected" ? "text-primary" : "text-error"}>
                {testResult === "connected" ? "Connection successful" : "Connection failed"}
              </span>
            </div>
          )}
          <div className="flex gap-2">
            <Button variant="secondary" size="sm" onClick={handleTest} disabled={testing}>
              <TestTube className="h-4 w-4" />
              {testing ? "Testing..." : "Test"}
            </Button>
            <Button variant="default" size="sm" onClick={handleSave}>
              <Save className="h-4 w-4" />
              {saved ? "Saved!" : "Save"}
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export default function SettingsPage() {
  return (
    <div className="flex flex-col h-full">
      <PageHeader
        title="Settings"
        description="Manage API keys, connections, and platform configuration"
      >
        <Button variant="default" size="sm" onClick={() => {}}>
            <Plug className="h-4 w-4" />
            Test All Connections
          </Button>
      </PageHeader>

      <div className="flex-1 p-6">
        <div className="flex flex-col gap-[1px] bg-outline-variant p-[1px]">
          <PageCell>
            <div className="grid grid-cols-4 gap-6">
              <div className="flex items-center gap-3">
                <div className="h-10 w-10 flex items-center justify-center bg-primary/10">
                  <Key className="h-5 w-5 text-primary" />
                </div>
                <div>
                  <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant">
                    Providers
                  </div>
                  <div className="text-[13px] font-mono font-medium text-on-surface">
                    {providers.length}
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="h-10 w-10 flex items-center justify-center bg-primary/10">
                  <Plug className="h-5 w-5 text-primary" />
                </div>
                <div>
                  <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant">
                    Connections
                  </div>
                  <div className="text-[13px] font-mono font-medium text-on-surface">
                    {connections.filter((c) => c.status === "connected").length}/{connections.length}
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="h-10 w-10 flex items-center justify-center bg-surface-container-high">
                  <Shield className="h-5 w-5 text-on-surface-variant" />
                </div>
                <div>
                  <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant">
                    Security
                  </div>
                  <div className="text-[13px] font-medium text-on-surface">
                    Keys encrypted
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="h-10 w-10 flex items-center justify-center bg-surface-container-high">
                  <Bell className="h-5 w-5 text-on-surface-variant" />
                </div>
                <div>
                  <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant">
                    Notifications
                  </div>
                  <div className="text-[13px] font-medium text-on-surface">
                    Enabled
                  </div>
                </div>
              </div>
            </div>
          </PageCell>
        </div>

        <div className="mt-[1px] grid grid-cols-3 gap-[1px] bg-outline-variant p-[1px]">
          <PageCell className="col-span-2">
            <div className="flex items-center justify-between mb-4">
              <div>
                <h2 className="text-base font-semibold text-on-surface">API Providers</h2>
                <p className="text-sm text-on-surface-variant mt-0.5">
                  Configure your API keys for market data and news providers
                </p>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              {providers.map((p) => (
                <ProviderCard key={p.id} provider={p} />
              ))}
            </div>
          </PageCell>

          <PageCell>
            <div className="mb-4">
              <h2 className="text-base font-semibold text-on-surface">Connections</h2>
              <p className="text-sm text-on-surface-variant mt-0.5">
                Infrastructure connections status
              </p>
            </div>
            <div className="space-y-3">
              {connections.map((conn) => (
                <div
                  key={conn.id}
                  className="flex items-center justify-between p-3 border border-outline-variant"
                >
                  <div className="flex items-center gap-3">
                    <StatusIndicator active={conn.status === "connected"} />
                    <div>
                      <div className="text-sm font-medium text-on-surface">{conn.name}</div>
                      <div className="text-xs text-on-surface-variant">{conn.description}</div>
                    </div>
                  </div>
                  <Badge
                    variant={conn.status === "connected" ? "success" : "error"}
                  >
                    {conn.status}
                  </Badge>
                </div>
              ))}
            </div>

            <div className="mt-6 pt-6 border-t border-outline-variant">
              <h2 className="text-base font-semibold text-on-surface mb-4">Notifications</h2>
              <div className="space-y-3">
                {[
                  { label: "Trade executed", enabled: true },
                  { label: "Strategy signal", enabled: true },
                  { label: "Price alert", enabled: false },
                  { label: "System error", enabled: true },
                ].map((item) => (
                  <div
                    key={item.label}
                    className="flex items-center justify-between p-3 border border-outline-variant"
                  >
                    <span className="text-sm text-on-surface">{item.label}</span>
                    <button
                      type="button"
                      className={`relative h-5 w-9 rounded-full transition-colors ${
                        item.enabled ? "bg-primary" : "bg-surface-container-high"
                      }`}
                    >
                      <span
                        className={`absolute top-0.5 left-0.5 h-4 w-4 rounded-full bg-on-primary transition-transform ${
                          item.enabled ? "translate-x-4" : ""
                        }`}
                      />
                    </button>
                  </div>
                ))}
              </div>
            </div>
          </PageCell>
        </div>
      </div>
    </div>
  );
}
