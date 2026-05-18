-- Add unique constraint on portfolio_id + ticker_id
ALTER TABLE positions ADD CONSTRAINT positions_portfolio_ticker_unique UNIQUE (portfolio_id, ticker_id);