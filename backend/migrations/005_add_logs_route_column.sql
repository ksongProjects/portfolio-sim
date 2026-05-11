ALTER TABLE logs ADD COLUMN route TEXT;
CREATE INDEX idx_logs_route ON logs(route);
