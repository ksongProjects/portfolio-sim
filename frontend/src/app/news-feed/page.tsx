import { PageCell, PageHeader } from "@/components/page-layout";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
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
  { symbol: "SPY", name: "S&P 500", price: "$523.45", change: "+0.87%" },
  { symbol: "DIA", name: "Dow Jones", price: "$398.20", change: "+0.45%" },
  { symbol: "QQQ", name: "Nasdaq", price: "$448.30", change: "+1.12%" },
];

export default function NewsFeedPage() {
  return (
    <div className="flex flex-col h-full">
      <PageHeader
        title="News Feed"
        description="Real-time market news and sentiment analysis"
      >
        <div className="flex gap-3">
          <Button variant="secondary" size="sm">
            <Bookmark className="h-4 w-4" />
            Saved
          </Button>
          <Button variant="default" size="sm">
            <TrendingUp className="h-4 w-4" />
            Market Summary
          </Button>
        </div>
      </PageHeader>

      <div className="flex-1 p-6">
        <div className="flex gap-[1px] bg-outline-variant p-[1px]">
          <PageCell className="flex-[2]">
            <Card className="border-0 bg-transparent">
              <CardHeader className="px-0 pb-4">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base font-semibold">Latest News</CardTitle>
                  <div className="flex gap-3">
                    <div className="relative">
                      <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-on-surface-variant" />
                      <Input placeholder="Search news..." className="pl-9 w-[200px]" />
                    </div>
                    <Button variant="ghost" size="icon">
                      <Filter className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="px-0">
                <div className="space-y-4">
                  {newsItems.map((news) => (
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
                      <ExternalLink className="h-4 w-4 text-on-surface-variant shrink-0 mt-1" />
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </PageCell>

          <PageCell className="flex-1">
            <div className="space-y-[1px] bg-outline-variant">
              <div className="bg-surface-container p-4">
                <CardTitle className="text-base font-semibold mb-4">Featured Story</CardTitle>
                <div className="space-y-3">
                  <Badge
                    variant={featuredNews.sentiment === "bullish" ? "success" : featuredNews.sentiment === "bearish" ? "error" : "secondary"}
                  >
                    {featuredNews.sentiment}
                  </Badge>
                  <h3 className="text-base font-medium text-on-surface leading-snug">{featuredNews.title}</h3>
                  <p className="text-sm text-on-surface-variant leading-relaxed">{featuredNews.summary}</p>
                  <div className="flex items-center justify-between pt-2">
                    <div className="flex items-center gap-2 text-xs text-on-surface-variant">
                      <span>{featuredNews.source}</span>
                      <span>•</span>
                      <span>{featuredNews.time}</span>
                    </div>
                  </div>
                  <div className="flex gap-2 pt-2">
                    {featuredNews.relatedSymbols.map((sym) => (
                      <Badge key={sym} variant="outline" className="text-[10px] font-mono">
                        {sym}
                      </Badge>
                    ))}
                  </div>
                </div>
              </div>

              <div className="bg-surface-container p-4">
                <CardTitle className="text-base font-semibold mb-4">Market Indices</CardTitle>
                <div className="space-y-3">
                  {marketIndices.map((index) => (
                    <div key={index.symbol} className="flex items-center justify-between">
                      <div>
                        <div className="font-mono text-sm font-medium text-on-surface">{index.symbol}</div>
                        <div className="text-xs text-on-surface-variant">{index.name}</div>
                      </div>
                      <div className="text-right">
                        <div className="font-mono text-sm text-on-surface">{index.price}</div>
                        <div className="text-xs text-primary">{index.change}</div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </PageCell>
        </div>
      </div>
    </div>
  );
}
