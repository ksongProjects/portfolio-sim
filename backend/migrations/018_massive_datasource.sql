INSERT INTO data_sources (id, name, source_priority, rate_limit_per_min, is_active, created_at)
SELECT 'massive', 'Massive', source_priority, 5, is_active, created_at
FROM data_sources
WHERE id = 'polygon'
ON CONFLICT (id) DO UPDATE SET
	name = EXCLUDED.name,
	rate_limit_per_min = EXCLUDED.rate_limit_per_min;

UPDATE provider_configurations SET provider_id = 'massive' WHERE provider_id = 'polygon';
UPDATE normalized_prices SET source_id = 'massive' WHERE source_id = 'polygon';
UPDATE raw_price_ticks SET source_id = 'massive' WHERE source_id = 'polygon';
UPDATE fundamental_data SET source_id = 'massive' WHERE source_id = 'polygon';
UPDATE raw_option_chains SET source_id = 'massive' WHERE source_id = 'polygon';
UPDATE option_chains SET source_id = 'massive' WHERE source_id = 'polygon';

UPDATE data_sources
SET name = 'Massive',
	rate_limit_per_min = 5
WHERE id = 'massive';

DELETE FROM data_sources WHERE id = 'polygon';
