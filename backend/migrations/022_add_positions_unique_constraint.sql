-- Add unique constraint on portfolio_id + ticker_id (idempotent)
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'positions_portfolio_ticker_unique'
  ) THEN
    ALTER TABLE positions ADD CONSTRAINT positions_portfolio_ticker_unique UNIQUE (portfolio_id, ticker_id);
  END IF;
END $$;