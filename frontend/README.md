# Frontend

Next.js web app for the Portfolio Simulation Platform. Displays portfolio positions with live market prices, market indices, news feeds, and ticker details.

## Stack

- **Next.js** (App Router)
- **React** with hooks
- **TanStack Query** for data fetching and caching
- **TanStack Table** for data tables
- **Recharts** for portfolio performance charts
- **Sonner** for toast notifications
- **Lucide React** for icons
- **CSS** with custom properties (no Tailwind)

## Pages

| Route | Description |
| --- | --- |
| `/` | Dashboard — portfolio summary, top holdings, recent activity, market indices |
| `/portfolio` | Positions table with add/remove workflow |
| `/ticker/[symbol]` | Ticker detail — profile, intraday bars, ratios, live price via WebSocket |
| `/news-feed` | RSS articles, YouTube channels, latest videos, Gemini analysis |
| `/strategy` | Strategy and signal views |
| `/observability` | Service health, structured logs |
| `/settings` | Provider credentials, RSS feeds, market indices configuration |

## Live Prices

Prices flow from a single SSE connection (`/api/stream/market`) through a global `PriceProvider` context. The `useLivePositions` hook merges REST position data with live tick data so portfolio and dashboard always show current prices without polling.

```
EventSource /api/stream/market
  --> PriceProvider (global context)
       --> useLivePositions (overlays live prices on REST positions)
       --> useLiveIndices (overlays live prices on market indices)
       --> useTickerLookup (per-ticker WebSocket)
```

## Scripts

```bash
pnpm install
pnpm dev       # Start dev server on :3000
pnpm build     # Production build
pnpm lint      # ESLint
```

## Architecture

- `src/app/` — Next.js App Router pages and layouts
- `src/components/` — Shared UI components (data-table, card, badge, button, etc.)
- `src/hooks/` — Data hooks (`usePortfolio`, `useTickerLookup`, `useNews`, `useLivePositions`, etc.)
- `src/lib/` — API client utilities (`fetchJson`, `apiFetch`)
- `src/components/price-context.tsx` — Global live price context (SSE consumer)
- `src/components/providers.tsx` — `QueryClientProvider` and `PriceProvider` wrapper

## API Configuration

The frontend connects to the Main API at `http://localhost:8080` by default. Override with the `NEXT_PUBLIC_API_URL` environment variable.

## Notes

- `package.json` enforces `pnpm` via a `preinstall` script.
- Lint errors and warnings are expected in some files (pre-existing issues not addressed).
- The `useLivePositions` hook intentionally preserves REST-derived `DayChange` and `TotalGain` fields so gain/loss and day-change calculations remain correct with live current prices.