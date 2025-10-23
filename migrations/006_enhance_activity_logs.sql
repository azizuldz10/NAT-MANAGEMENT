-- Migration: Enhance activity_logs table with duration and device context
-- Description: Add operation duration tracking and device information for better activity insights

-- Add new columns
ALTER TABLE activity_logs ADD COLUMN IF NOT EXISTS duration_ms INTEGER;
ALTER TABLE activity_logs ADD COLUMN IF NOT EXISTS device_info JSONB;

-- Add comments
COMMENT ON COLUMN activity_logs.duration_ms IS 'Operation duration in milliseconds';
COMMENT ON COLUMN activity_logs.device_info IS 'Device context: browser, OS, device type extracted from user agent';

-- Create index for performance queries on duration
CREATE INDEX IF NOT EXISTS idx_activity_logs_duration ON activity_logs(duration_ms) WHERE duration_ms IS NOT NULL;

-- Create index for device_info queries
CREATE INDEX IF NOT EXISTS idx_activity_logs_device_info ON activity_logs USING GIN (device_info);
