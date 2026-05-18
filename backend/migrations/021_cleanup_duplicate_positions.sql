-- Cleanup duplicate positions: keep the one with highest quantity, delete others
-- First, find duplicates and keep the one with largest quantity
DELETE FROM positions
WHERE id IN (
    SELECT id FROM (
        SELECT id, ROW_NUMBER() OVER (PARTITION BY portfolio_id, ticker_id ORDER BY quantity DESC) as rn
        FROM positions
    ) sub
    WHERE rn > 1
);