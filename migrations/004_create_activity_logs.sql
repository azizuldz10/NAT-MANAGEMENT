-- Migration: Create activity_logs table
-- Description: Store all user activities for audit trail

-- Create activity_logs table
CREATE TABLE IF NOT EXISTS activity_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER,
    username VARCHAR(100) NOT NULL,
    user_role VARCHAR(50),
    action_type VARCHAR(50) NOT NULL,  -- LOGIN, LOGOUT, CREATE, UPDATE, DELETE, NAT_UPDATE, PPPOE_CHECK
    resource_type VARCHAR(50),          -- USER, ROUTER, NAT_RULE, PPPOE
    resource_id VARCHAR(255),           -- ID or name of the resource affected
    description TEXT NOT NULL,          -- Human-readable description
    ip_address VARCHAR(45),             -- IPv4 or IPv6
    user_agent TEXT,                    -- Browser/client info
    status VARCHAR(20) DEFAULT 'SUCCESS', -- SUCCESS, FAILED, ERROR
    error_message TEXT,                 -- If status is FAILED/ERROR
    metadata JSONB,                     -- Additional structured data
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- Foreign key (optional, can be NULL for deleted users)
    CONSTRAINT fk_activity_logs_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE SET NULL
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_activity_logs_user_id ON activity_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_activity_logs_username ON activity_logs(username);
CREATE INDEX IF NOT EXISTS idx_activity_logs_action_type ON activity_logs(action_type);
CREATE INDEX IF NOT EXISTS idx_activity_logs_resource_type ON activity_logs(resource_type);
CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_activity_logs_status ON activity_logs(status);

-- Create index for common query patterns
CREATE INDEX IF NOT EXISTS idx_activity_logs_user_action ON activity_logs(user_id, action_type, created_at DESC);

-- Add comment to table
COMMENT ON TABLE activity_logs IS 'Activity audit log for tracking all user actions in the system';
COMMENT ON COLUMN activity_logs.action_type IS 'Type of action: LOGIN, LOGOUT, CREATE, UPDATE, DELETE, NAT_UPDATE, PPPOE_CHECK';
COMMENT ON COLUMN activity_logs.resource_type IS 'Type of resource affected: USER, ROUTER, NAT_RULE, PPPOE';
COMMENT ON COLUMN activity_logs.metadata IS 'Additional structured data in JSON format';
