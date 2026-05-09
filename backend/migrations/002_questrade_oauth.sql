-- Migration: Add Questrade OAuth fields to provider_configurations

ALTER TABLE provider_configurations
ADD COLUMN IF NOT EXISTS access_token TEXT,
ADD COLUMN IF NOT EXISTS refresh_token TEXT,
ADD COLUMN IF NOT EXISTS api_server TEXT,
ADD COLUMN IF NOT EXISTS token_expires_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_provider_configurations_provider_id ON provider_configurations(provider_id);