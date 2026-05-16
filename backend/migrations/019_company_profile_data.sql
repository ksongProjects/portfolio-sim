-- Add company_profile to fundamental_data data_types
INSERT INTO data_sources (id, name, source_priority, is_active)
VALUES ('company_profile', 'Company Profile Data', 4, true)
ON CONFLICT (id) DO NOTHING;