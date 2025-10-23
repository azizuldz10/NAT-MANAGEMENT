-- ============================================================================
-- NAT Management System - User Router Access
-- Migration: 003
-- Purpose: Simple user-to-router mapping table
-- ============================================================================

-- ============================================================================
-- USER ROUTERS TABLE
-- Simple mapping: which routers can each user access?
-- ============================================================================
CREATE TABLE IF NOT EXISTS user_routers (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    router_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, router_name)
);

-- Indexes for fast lookup
CREATE INDEX idx_user_routers_user_id ON user_routers(user_id);
CREATE INDEX idx_user_routers_router_name ON user_routers(router_name);

-- ============================================================================
-- COMMENTS
-- ============================================================================
COMMENT ON TABLE user_routers IS 'Maps which routers each user can access';
COMMENT ON COLUMN user_routers.user_id IS 'Reference to users table';
COMMENT ON COLUMN user_routers.router_name IS 'Router name (must match routers.name)';

-- ============================================================================
-- SEED DATA FOR EXISTING USERS (Optional)
-- Give admin access to all routers by default
-- ============================================================================
DO $$
DECLARE
    admin_user_id INTEGER;
    router_record RECORD;
BEGIN
    -- Get admin user ID
    SELECT id INTO admin_user_id FROM users WHERE username = 'admin' LIMIT 1;

    IF admin_user_id IS NOT NULL THEN
        -- Give admin access to all routers
        FOR router_record IN SELECT name FROM routers WHERE enabled = true
        LOOP
            INSERT INTO user_routers (user_id, router_name)
            VALUES (admin_user_id, router_record.name)
            ON CONFLICT (user_id, router_name) DO NOTHING;
        END LOOP;

        RAISE NOTICE '✅ Admin user granted access to all routers';
    END IF;
END $$;

-- ============================================================================
-- SUCCESS MESSAGE
-- ============================================================================
DO $$
DECLARE
    mapping_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO mapping_count FROM user_routers;

    RAISE NOTICE '✅ User router access table created successfully!';
    RAISE NOTICE '📊 Total user-router mappings: %', mapping_count;
    RAISE NOTICE '🔧 Next: Implement user management CRUD in backend';
END $$;
