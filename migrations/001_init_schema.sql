-- ============================================================================
-- NAT Management System - PostgreSQL Schema
-- Database: Neon PostgreSQL (Serverless)
-- Version: 1.0
-- ============================================================================

-- Enable UUID extension for generating unique IDs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- USERS TABLE
-- Stores user accounts with authentication credentials
-- ============================================================================
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL, -- bcrypt hashed password
    full_name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    role VARCHAR(50) NOT NULL, -- Administrator, Head Branch 1, Head Branch 2, Head Branch 3
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE
);

-- Index for fast username lookup
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);

-- ============================================================================
-- ROUTERS TABLE
-- Stores router configurations for NAT management
-- ============================================================================
CREATE TABLE IF NOT EXISTS routers (
    id VARCHAR(100) PRIMARY KEY, -- Format: "routername-uuid"
    name VARCHAR(100) UNIQUE NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL CHECK (port > 0 AND port <= 65535),
    username VARCHAR(100) NOT NULL,
    password VARCHAR(255) NOT NULL, -- Encrypted in production
    tunnel_endpoint VARCHAR(255) NOT NULL,
    public_ont_url VARCHAR(255) NOT NULL,
    enabled BOOLEAN DEFAULT true,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for fast router lookup
CREATE INDEX idx_routers_name ON routers(name);
CREATE INDEX idx_routers_enabled ON routers(enabled);
CREATE INDEX idx_routers_created_at ON routers(created_at DESC);

-- ============================================================================
-- ROUTER ACCESS CONTROL TABLE
-- Stores role-based access control for routers
-- ============================================================================
CREATE TABLE IF NOT EXISTS router_access_control (
    id SERIAL PRIMARY KEY,
    role VARCHAR(50) NOT NULL,
    router_name VARCHAR(100) NOT NULL, -- Can be "*" for all routers
    permissions VARCHAR(50)[] DEFAULT ARRAY['read'], -- Array: read, write, delete, manage
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(role, router_name)
);

-- Index for fast role-based access lookup
CREATE INDEX idx_router_access_role ON router_access_control(role);
CREATE INDEX idx_router_access_router_name ON router_access_control(router_name);

-- ============================================================================
-- PPPOE SEARCH HISTORY TABLE (Optional - for future analytics)
-- Stores history of PPPoE username searches
-- ============================================================================
CREATE TABLE IF NOT EXISTS pppoe_search_history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    username VARCHAR(100) NOT NULL, -- PPPoE username being searched
    router_name VARCHAR(100), -- Which router was searched
    is_online BOOLEAN,
    ip_address VARCHAR(45), -- IPv4 or IPv6
    search_timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for search history queries
CREATE INDEX idx_pppoe_history_user_id ON pppoe_search_history(user_id);
CREATE INDEX idx_pppoe_history_username ON pppoe_search_history(username);
CREATE INDEX idx_pppoe_history_timestamp ON pppoe_search_history(search_timestamp DESC);

-- ============================================================================
-- AUDIT LOG TABLE (Optional - for security and compliance)
-- Tracks all important actions in the system
-- ============================================================================
CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    username VARCHAR(50),
    action VARCHAR(100) NOT NULL, -- login, logout, add_router, delete_router, update_nat, etc.
    resource_type VARCHAR(50), -- user, router, nat_rule
    resource_id VARCHAR(100),
    ip_address VARCHAR(45),
    user_agent TEXT,
    details JSONB, -- Store additional details as JSON
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for audit log queries
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);

-- ============================================================================
-- FUNCTIONS: Auto-update updated_at timestamp
-- ============================================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply auto-update triggers
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_routers_updated_at BEFORE UPDATE ON routers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_router_access_updated_at BEFORE UPDATE ON router_access_control
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- COMMENTS for documentation
-- ============================================================================
COMMENT ON TABLE users IS 'User accounts with authentication credentials';
COMMENT ON TABLE routers IS 'Router configurations for NAT management';
COMMENT ON TABLE router_access_control IS 'Role-based access control for routers';
COMMENT ON TABLE pppoe_search_history IS 'History of PPPoE username searches';
COMMENT ON TABLE audit_logs IS 'Audit trail for security and compliance';

-- ============================================================================
-- SUCCESS MESSAGE
-- ============================================================================
DO $$
BEGIN
    RAISE NOTICE '✅ NAT Management System schema created successfully!';
    RAISE NOTICE '📊 Tables created: users, routers, router_access_control, pppoe_search_history, audit_logs';
    RAISE NOTICE '🔧 Next step: Run 002_seed_data.sql to populate initial data';
END $$;
