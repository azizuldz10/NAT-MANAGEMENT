-- Migration: 005_create_ont_wifi_info
-- Description: Create table for storing ONT WiFi extraction information
-- Date: 2025-10-16
-- Author: NAT Management Team

-- Create ont_wifi_info table
CREATE TABLE IF NOT EXISTS ont_wifi_info (
    id SERIAL PRIMARY KEY,
    pppoe_username VARCHAR(255),
    router VARCHAR(255),
    ssid VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    security VARCHAR(100),
    encryption VARCHAR(100),
    authentication VARCHAR(100),
    ont_url VARCHAR(500) NOT NULL,
    ont_model VARCHAR(100),
    extracted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    extracted_by VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Indexes for faster queries
    CONSTRAINT ont_wifi_info_ssid_password_check CHECK (ssid <> '' AND password <> '')
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_ont_wifi_info_pppoe_username ON ont_wifi_info(pppoe_username);
CREATE INDEX IF NOT EXISTS idx_ont_wifi_info_router ON ont_wifi_info(router);
CREATE INDEX IF NOT EXISTS idx_ont_wifi_info_ssid ON ont_wifi_info(ssid);
CREATE INDEX IF NOT EXISTS idx_ont_wifi_info_extracted_at ON ont_wifi_info(extracted_at DESC);
CREATE INDEX IF NOT EXISTS idx_ont_wifi_info_pppoe_router ON ont_wifi_info(pppoe_username, router);

-- Create composite index for history queries
CREATE INDEX IF NOT EXISTS idx_ont_wifi_info_history ON ont_wifi_info(pppoe_username, router, extracted_at DESC);

-- Add comment to table
COMMENT ON TABLE ont_wifi_info IS 'Stores WiFi information extracted from ONT devices using webautomation tools';

-- Add comments to columns
COMMENT ON COLUMN ont_wifi_info.pppoe_username IS 'PPPoE username associated with this ONT (optional)';
COMMENT ON COLUMN ont_wifi_info.router IS 'Router name where this ONT is connected (optional)';
COMMENT ON COLUMN ont_wifi_info.ssid IS 'WiFi SSID extracted from ONT';
COMMENT ON COLUMN ont_wifi_info.password IS 'WiFi password extracted from ONT';
COMMENT ON COLUMN ont_wifi_info.security IS 'WiFi security type (e.g., WPA2, WPA/WPA2-PSK)';
COMMENT ON COLUMN ont_wifi_info.encryption IS 'WiFi encryption method (e.g., AES, TKIP)';
COMMENT ON COLUMN ont_wifi_info.authentication IS 'WiFi authentication method';
COMMENT ON COLUMN ont_wifi_info.ont_url IS 'Public URL of the ONT device';
COMMENT ON COLUMN ont_wifi_info.ont_model IS 'Detected ONT model (e.g., ZTE F477V2, GM220-S)';
COMMENT ON COLUMN ont_wifi_info.extracted_at IS 'Timestamp when WiFi info was extracted';
COMMENT ON COLUMN ont_wifi_info.extracted_by IS 'Username who triggered the extraction';
COMMENT ON COLUMN ont_wifi_info.created_at IS 'Timestamp when record was created';

-- Insert migration record
INSERT INTO schema_migrations (version, description, applied_at)
VALUES ('005', 'Create ont_wifi_info table for WiFi extraction', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO NOTHING;

-- Success message
DO $$
BEGIN
    RAISE NOTICE '✅ Migration 005: ont_wifi_info table created successfully';
    RAISE NOTICE '📋 Table: ont_wifi_info';
    RAISE NOTICE '📊 Indexes: 6 indexes created for optimized queries';
    RAISE NOTICE '🔍 Ready for ONT WiFi extraction integration';
END $$;
