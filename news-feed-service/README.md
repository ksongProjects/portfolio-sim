# News Feed Service

RSS feed aggregation, YouTube transcription, and AI-powered content summarization service.

## Overview

Written in Python. Scrapes RSS feeds, transcribes YouTube videos, generates summaries via Google Gemini, and publishes updates via Redis pub/sub.

## Configuration

Environment variables:

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | PostgreSQL connection string |
| `REDIS_ADDR` | Redis address (default: localhost:6379) |
| `GEMINI_API_KEY` | Google Gemini API key for summarization |
| `YOUTUBE_API_KEY` | YouTube Data API key for transcripts |
| `SCRAPE_INTERVAL_MIN` | RSS scrape interval in minutes (default: 15) |

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    News Feed Service                         │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────┐    ┌────────────┐    ┌────────────┐           │
│  │  RSS     │───►│   Feed     │───►│  News      │           │
│  │  Feeds   │    │  Manager   │    │  Articles  │           │
│  └──────────┘    └────────────┘    └────────────┘           │
│                      │                                        │
│                      ▼                                        │
│  ┌──────────┐    ┌────────────┐    ┌────────────┐           │
│  │ YouTube  │───►│ Transcribe │───►│  Gemini    │           │
│  │  Videos  │    │   Queue    │    │ Summarize  │           │
│  └──────────┘    └────────────┘    └────────────┘           │
│                                          │                    │
│                                          ▼                    │
│                                   ┌────────────┐             │
│                                   │    SSE     │             │
│                                   │  Manager   │             │
│                                   └────────────┘             │
└─────────────────────────────────────────────────────────────┘
```

## Processing Pipelines

### RSS Scraping Pipeline

1. Scheduler triggers scrape every `SCRAPE_INTERVAL_MIN` minutes
2. Each feed is fetched and parsed
3. New articles are stored in the database
4. Sentiment analysis applied (bullish/bearish/neutral)
5. Ticker symbols extracted from article content
6. Updates published to Redis `news:articles` channel

### YouTube Transcription Pipeline

1. Video IDs pushed to `queue:transcribe` Redis queue
2. Worker pulls job from queue
3. Fetches transcript via YouTube Data API
4. Sends transcript to Gemini for summarization
5. Publishes summary to Redis `news:videos` channel

## Redis Queues

### `queue:scrape-news`
Triggers RSS feed scraping.

### `queue:transcribe`
YouTube video transcription jobs.

```json
{
  "video_id": "dQw4w9WgXcQ"
}
```

## SSE Channels

### `news:articles`
New article notifications.

### `news:videos`
Video summary notifications.

```json
{
  "youtube_id": "dQw4w9WgXcQ",
  "summary": "This video discusses..."
}
```

## Database Schema

### news_articles
```sql
CREATE TABLE news_articles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title VARCHAR(500),
  source VARCHAR(100),
  url TEXT,
  summary TEXT,
  sentiment VARCHAR(20),
  published_at TIMESTAMPTZ,
  ticker_symbols TEXT[],
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### rss_feeds (managed by main-api)
```sql
CREATE TABLE rss_feeds (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(200),
  url TEXT,
  scrape_interval_min INT DEFAULT 60,
  last_scrape_at TIMESTAMPTZ,
  is_active BOOLEAN DEFAULT TRUE
);
```

## Dependencies

- **feedparser**: RSS feed parsing
- **google-generativeai**: Gemini API client
- **redis-py**: Redis client
- **asyncpg**: PostgreSQL async driver
- **FastAPI**: HTTP server (minimal, just for health endpoint)

## Health Check

```
GET /health
Response: "ok"
```

## Environment Setup

```bash
pip install -r requirements.txt

export DATABASE_URL="postgres://postgres:postgres@localhost:5432/portfolio_sim"
export REDIS_ADDR="localhost:6379"
export GEMINI_API_KEY="your-gemini-key"
export YOUTUBE_API_KEY="your-youtube-key"
export SCRAPE_INTERVAL_MIN=15

python main.py
```