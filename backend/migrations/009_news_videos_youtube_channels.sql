ALTER TABLE news_videos ADD COLUMN sentiment TEXT;
ALTER TABLE news_videos ADD COLUMN tickers JSONB DEFAULT '[]'::jsonb;

CREATE TABLE IF NOT EXISTS youtube_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    youtube_handle TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO youtube_channels (channel_id, name, youtube_handle) VALUES
    ('UC5zaV8kAMRzRJXq6TvieB2w', 'Yahoo Finance', '@yahoofinance'),
    ('UCvjgEDsH9BbR8lL3fm3ed1A', 'CNBC', '@cnbc'),
    ('UC0vWUnOECI8x-LUQUoqJDPg', 'Bloomberg', '@bloomberg');