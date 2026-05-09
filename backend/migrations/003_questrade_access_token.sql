-- Migration: Add access_token to provider_configurations for Questrade

ALTER TABLE provider_configurations
ADD COLUMN IF NOT EXISTS access_token TEXT;