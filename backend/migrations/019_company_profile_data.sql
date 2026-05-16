-- Add company_profile to fundamental_data data_types
INSERT INTO data_sources (id, provider_id, name, description, type, doc_url, is_active)
VALUES ('company_profile', 'company_profile', 'Company Profile Data', 'Static company information including sector, industry, market cap', 'company_data', '', true)
ON CONFLICT (id) DO NOTHING;