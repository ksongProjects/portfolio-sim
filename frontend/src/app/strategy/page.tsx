import { PageCell, PageHeader } from "@/components/page-layout";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge, StatusIndicator } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Play, Pause, Plus, Settings, TrendingUp, BarChart3, Zap, Shield } from "lucide-react";

const strategies = [
  { id: 1, name: "Momentum Growth", status: "active", returns: "+18.4%", sharpe: "1.42", maxDD: "-12.3%", trades: 847, winRate: "64.2%" },
  { id: 2, name: "Value Scanner", status: "active", returns: "+12.1%", sharpe: "1.18", maxDD: "-8.7%", trades: 423, winRate: "58.9%" },
  { id: 3, name: "Mean Reversion", status: "paused", returns: "+8.3%", sharpe: "0.95", maxDD: "-15.2%", trades: 612, winRate: "52.1%" },
  { id: 4, name: "Sector Rotation", status: "active", returns: "+15.7%", sharpe: "1.28", maxDD: "-10.1%", trades: 234, winRate: "61.4%" },
];

const recentSignals = [
  { id: 1, strategy: "Momentum Growth", symbol: "NVDA", action: "BUY", price: "$875.28", time: "10:32 AM", confidence: "HIGH" },
  { id: 2, strategy: "Value Scanner", symbol: "JPM", action: "BUY", price: "$198.50", time: "09:15 AM", confidence: "MEDIUM" },
  { id: 3, strategy: "Sector Rotation", symbol: "TLT", action: "SELL", price: "$92.45", time: "Yesterday", confidence: "HIGH" },
  { id: 4, strategy: "Momentum Growth", symbol: "TSLA", action: "BUY", price: "$245.80", time: "Yesterday", confidence: "MEDIUM" },
];

export default function StrategyPage() {
  return (
    <div className="flex flex-col h-full">
      <PageHeader
        title="Strategy"
        description="Trading strategies and signal generation"
      >
        <div className="flex gap-3">
          <Button variant="secondary" size="sm">
            <Settings className="h-4 w-4" />
            Configure
          </Button>
          <Button variant="default" size="sm">
            <Plus className="h-4 w-4" />
            New Strategy
          </Button>
        </div>
      </PageHeader>

      <div className="flex-1 p-6">
        <div className="flex flex-col gap-[1px] bg-outline-variant p-[1px]">
          <PageCell>
            <div className="grid grid-cols-4 gap-6">
              <div className="flex items-center gap-3">
                <div className="h-10 w-10 flex items-center justify-center bg-primary/10">
                  <Zap className="h-5 w-5 text-primary" />
                </div>
                <div>
                  <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant">Active Strategies</div>
                  <div className="text-[13px] font-mono font-medium text-on-surface">3</div>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="h-10 w-10 flex items-center justify-center bg-surface-container-high">
                  <BarChart3 className="h-5 w-5 text-on-surface-variant" />
                </div>
                <div>
                  <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant">Total Trades</div>
                  <div className="text-[13px] font-mono font-medium text-on-surface">2,116</div>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="h-10 w-10 flex items-center justify-center bg-surface-container-high">
                  <TrendingUp className="h-5 w-5 text-on-surface-variant" />
                </div>
                <div>
                  <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant">Avg Win Rate</div>
                  <div className="text-[13px] font-mono font-medium text-on-surface">59.1%</div>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="h-10 w-10 flex items-center justify-center bg-surface-container-high">
                  <Shield className="h-5 w-5 text-on-surface-variant" />
                </div>
                <div>
                  <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant">Avg Sharpe</div>
                  <div className="text-[13px] font-mono font-medium text-on-surface">1.21</div>
                </div>
              </div>
            </div>
          </PageCell>
        </div>

        <div className="mt-[1px] flex gap-[1px] bg-outline-variant p-[1px]">
          <PageCell className="flex-[2]">
            <Card className="border-0 bg-transparent">
              <CardHeader className="px-0 pb-4">
                <CardTitle className="text-base font-semibold">Strategies</CardTitle>
              </CardHeader>
              <CardContent className="px-0">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Name</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Returns</TableHead>
                      <TableHead>Sharpe</TableHead>
                      <TableHead>Max DD</TableHead>
                      <TableHead>Trades</TableHead>
                      <TableHead>Win Rate</TableHead>
                      <TableHead>Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {strategies.map((strategy) => (
                      <TableRow key={strategy.id}>
                        <TableCell className="font-medium">{strategy.name}</TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <StatusIndicator active={strategy.status === "active"} />
                            <span className="text-xs capitalize">{strategy.status}</span>
                          </div>
                        </TableCell>
                        <TableCell className="font-mono text-primary">{strategy.returns}</TableCell>
                        <TableCell className="font-mono">{strategy.sharpe}</TableCell>
                        <TableCell className="font-mono text-error">{strategy.maxDD}</TableCell>
                        <TableCell className="font-mono">{strategy.trades}</TableCell>
                        <TableCell className="font-mono">{strategy.winRate}</TableCell>
                        <TableCell>
                          <Button variant="ghost" size="icon">
                            {strategy.status === "active" ? (
                              <Pause className="h-4 w-4" />
                            ) : (
                              <Play className="h-4 w-4" />
                            )}
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          </PageCell>

          <PageCell className="flex-1">
            <Card className="border-0 bg-transparent">
              <CardHeader className="px-0 pb-4">
                <CardTitle className="text-base font-semibold">Recent Signals</CardTitle>
              </CardHeader>
              <CardContent className="px-0">
                <div className="space-y-3">
                  {recentSignals.map((signal) => (
                    <div key={signal.id} className="flex items-center justify-between p-3 border border-outline-variant">
                      <div className="flex items-center gap-3">
                        <Badge variant={signal.action === "BUY" ? "success" : "error"}>
                          {signal.action}
                        </Badge>
                        <div>
                          <div className="font-mono text-sm font-medium text-on-surface">{signal.symbol}</div>
                          <div className="text-xs text-on-surface-variant">{signal.strategy}</div>
                        </div>
                      </div>
                      <div className="text-right">
                        <div className="text-sm font-mono text-on-surface">{signal.price}</div>
                        <div className="text-xs text-on-surface-variant">{signal.time}</div>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </PageCell>
        </div>
      </div>
    </div>
  );
}
