-- Portfolio Sim Database Schema
-- PostgreSQL with monthly partitions

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE data_sources (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    source_priority INT NOT NULL,
    api_key_encrypted TEXT,
    rate_limit_per_min INT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO data_sources (id, name, source_priority, rate_limit_per_min) VALUES
    ('questrade', 'Questrade', 1, 100),
    ('polygon', 'Polygon', 2, 60),
    ('fmp', 'Financial Modeling Prep', 3, 250);

CREATE TABLE tickers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol TEXT NOT NULL UNIQUE,
    company_name TEXT,
    exchange TEXT,
    asset_type TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_tickers_symbol ON tickers(symbol);

CREATE TABLE raw_price_ticks (
    id UUID DEFAULT gen_random_uuid(),
    ticker_id UUID NOT NULL REFERENCES tickers(id),
    source_id TEXT NOT NULL REFERENCES data_sources(id),
    raw_json JSONB NOT NULL,
    price DECIMAL(18,8),
    bid DECIMAL(18,8),
    ask DECIMAL(18,8),
    volume BIGINT,
    received_at TIMESTAMPTZ NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
) PARTITION BY RANGE (received_at);

CREATE INDEX idx_raw_price_ticks_ticker_time ON raw_price_ticks(ticker_id, received_at DESC);

CREATE TABLE normalized_prices (
    ticker_id UUID NOT NULL REFERENCES tickers(id),
    price DECIMAL(18,8) NOT NULL,
    bid DECIMAL(18,8),
    ask DECIMAL(18,8),
    volume BIGINT,
    source_id TEXT NOT NULL REFERENCES data_sources(id),
    timestamp TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (ticker_id, timestamp)
) PARTITION BY RANGE (timestamp);

CREATE INDEX idx_normalized_prices_ticker ON normalized_prices(ticker_id, timestamp DESC);

CREATE TABLE intraday_bars (
    ticker_id UUID NOT NULL REFERENCES tickers(id),
    interval TEXT NOT NULL,
    open DECIMAL(18,8) NOT NULL,
    high DECIMAL(18,8) NOT NULL,
    low DECIMAL(18,8) NOT NULL,
    close DECIMAL(18,8) NOT NULL,
    volume BIGINT,
    timestamp TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (ticker_id, interval, timestamp)
) PARTITION BY RANGE (timestamp);

CREATE TABLE fundamental_data (
    ticker_id UUID NOT NULL REFERENCES tickers(id),
    source_id TEXT NOT NULL REFERENCES data_sources(id),
    data_type TEXT NOT NULL,
    period TEXT,
    json_data JSONB NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (ticker_id, source_id, data_type, period, timestamp)
);

CREATE TABLE raw_option_chains (
    id UUID DEFAULT gen_random_uuid(),
    underlying_ticker_id UUID NOT NULL REFERENCES tickers(id),
    source_id TEXT NOT NULL REFERENCES data_sources(id),
    raw_json JSONB NOT NULL,
    received_at TIMESTAMPTZ NOT NULL,
    fetched_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (id)
);

CREATE TABLE option_chains (
    id UUID DEFAULT gen_random_uuid(),
    underlying_ticker_id UUID NOT NULL REFERENCES tickers(id),
    source_id TEXT NOT NULL REFERENCES data_sources(id),
    expiration TIMESTAMPTZ NOT NULL,
    strike DECIMAL(18,4) NOT NULL,
    option_type TEXT NOT NULL CHECK (option_type IN ('call', 'put')),
    bid DECIMAL(18,4),
    ask DECIMAL(18,4),
    delta DECIMAL(18,8),
    gamma DECIMAL(18,8),
    theta DECIMAL(18,8),
    vega DECIMAL(18,8),
    implied_vol DECIMAL(18,8),
    volume BIGINT,
    open_interest BIGINT,
    timestamp TIMESTAMPTZ NOT NULL,
    fetched_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (underlying_ticker_id, expiration, strike, option_type, timestamp)
) PARTITION BY RANGE (timestamp);

CREATE TABLE option_greeks (
    chain_id UUID NOT NULL REFERENCES raw_option_chains(id),
    strike DECIMAL(18,4) NOT NULL,
    option_type TEXT NOT NULL CHECK (option_type IN ('call', 'put')),
    delta DECIMAL(18,8),
    gamma DECIMAL(18,8),
    theta DECIMAL(18,8),
    vega DECIMAL(18,8),
    implied_vol DECIMAL(18,8),
    computed_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (chain_id, strike, option_type)
);

CREATE TABLE news_articles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticker_ids UUID[] NOT NULL,
    source TEXT NOT NULL,
    title TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE,
    summary TEXT,
    sentiment TEXT,
    published_at TIMESTAMPTZ,
    fetched_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_news_articles_published ON news_articles(published_at DESC);

CREATE TABLE news_videos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    youtube_id TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    channel TEXT,
    transcript_text TEXT,
    summary_text TEXT,
    summary_model TEXT DEFAULT 'gemini-2.5-flash',
    published_at TIMESTAMPTZ,
    fetched_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE rss_feeds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE,
    ticker_ids UUID[],
    scrape_interval_min INT DEFAULT 60,
    last_scrape_at TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT true
);

CREATE TABLE jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_type TEXT NOT NULL CHECK (job_type IN ('backfill', 'scrape-news', 'transcribe', 'compute-greeks', 'backtest')),
    status TEXT NOT NULL CHECK (status IN ('pending', 'running', 'complete', 'failed')),
    payload JSONB NOT NULL,
    result JSONB,
    error TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_jobs_status ON jobs(status);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clerk_id TEXT NOT NULL UNIQUE,
    email TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE portfolios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    name TEXT NOT NULL,
    initial_cash DECIMAL(18,2) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE positions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    portfolio_id UUID NOT NULL REFERENCES portfolios(id),
    ticker_id UUID NOT NULL REFERENCES tickers(id),
    quantity DECIMAL(18,8) NOT NULL,
    avg_cost DECIMAL(18,2) NOT NULL,
    opened_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE watchlists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE watchlist_tickers (
    watchlist_id UUID NOT NULL REFERENCES watchlists(id),
    ticker_id UUID NOT NULL REFERENCES tickers(id),
    added_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (watchlist_id, ticker_id)
);

CREATE TABLE simulation_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    portfolio_id UUID NOT NULL REFERENCES portfolios(id),
    run_type TEXT NOT NULL CHECK (run_type IN ('backtest', 'forward')),
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ,
    initial_cash DECIMAL(18,2),
    final_value DECIMAL(18,2),
    metrics JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE logs (
    id UUID DEFAULT gen_random_uuid(),
    timestamp TIMESTAMPTZ NOT NULL,
    level TEXT NOT NULL CHECK (level IN ('DEBUG', 'INFO', 'WARN', 'ERROR', 'FATAL')),
    service TEXT NOT NULL,
    component TEXT,
    message TEXT NOT NULL,
    metadata JSONB,
    trace_id TEXT,
    span_id TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
) PARTITION BY RANGE (timestamp);

CREATE INDEX idx_logs_timestamp ON logs(timestamp DESC);
CREATE INDEX idx_logs_service ON logs(service);
CREATE INDEX idx_logs_level ON logs(level);

CREATE OR REPLACE FUNCTION create_monthly_partition(
    table_name TEXT,
    partition_date DATE
) RETURNS TEXT AS $$
DECLARE
    partition_name TEXT;
    start_date DATE;
    end_date DATE;
BEGIN
    start_date := date_trunc('month', partition_date);
    end_date := start_date + INTERVAL '1 month';
    partition_name := table_name || '_' || to_char(start_date, 'YYYY_MM');
    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF %I FOR VALUES FROM (%L) TO (%L)',
        partition_name, table_name, start_date, end_date
    );
    RETURN partition_name;
END;
$$ LANGUAGE plpgsql;

SELECT create_monthly_partition('raw_price_ticks', CURRENT_DATE);
SELECT create_monthly_partition('normalized_prices', CURRENT_DATE);
SELECT create_monthly_partition('intraday_bars', CURRENT_DATE);
SELECT create_monthly_partition('option_chains', CURRENT_DATE);
SELECT create_monthly_partition('logs', CURRENT_DATE);