-- Redis initialization for portfolio-sim
-- Queues: backfill, scrape-news, transcribe, compute-greeks, backtest
-- Channels: market:ticks, market:chains, news:articles, news:videos

HSET queue:meta:backfill max_retries 3 retry_delay 5
HSET queue:meta:scrape-news max_retries 3 retry_delay 10
HSET queue:meta:transcribe max_retries 2 retry_delay 30
HSET queue:meta:compute-greeks max_retries 2 retry_delay 5
HSET queue:meta:backtest max_retries 1 retry_delay 60

SET config:option_chain_interval 60
HSET config:news_intervals rss 300 youtube 600 gemini 900
HSET cron:partitions table_names "raw_price_ticks,normalized_prices,intraday_bars,raw_option_chains,option_chains,logs" months_ahead 3

puts "Redis setup complete"