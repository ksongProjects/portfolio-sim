ALTER TABLE news_articles ADD COLUMN IF NOT EXISTS source_type TEXT NOT NULL DEFAULT 'rss';
ALTER TABLE news_articles ADD COLUMN IF NOT EXISTS sentiment_value TEXT;
ALTER TABLE news_articles ADD COLUMN IF NOT EXISTS content TEXT;
ALTER TABLE news_articles ADD COLUMN IF NOT EXISTS channel TEXT;

DROP INDEX IF EXISTS idx_news_articles_published;
CREATE INDEX idx_news_articles_source_type_published ON news_articles(source_type, published_at DESC);

UPDATE news_articles SET source_type = 'rss' WHERE source_type IS NULL OR source_type = '';