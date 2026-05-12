INSERT INTO users (id, clerk_id, email)
VALUES ('00000000-0000-0000-0000-000000000001', 'demo', 'demo@demo.com')
ON CONFLICT (clerk_id) DO NOTHING;

INSERT INTO portfolios (id, user_id, name, initial_cash)
VALUES ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'Default Portfolio', 100000.00)
ON CONFLICT (id) DO NOTHING;