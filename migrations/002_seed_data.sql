-- ============================================================================
-- NAT Management System - Initial Data Seeding
-- Purpose: Populate database with default users and role access control
-- ============================================================================

-- ============================================================================
-- INSERT DEFAULT USERS
-- Password hash generated using bcrypt (cost=10)
-- Default passwords:
--   - admin: admin123
--   - head1: head123
--   - head2: head123
--   - head3: head123
-- ============================================================================

-- Admin User (full access)
INSERT INTO users (username, password, full_name, email, role, is_active)
VALUES (
    'admin',
    '$2a$10$OWZ4HdWnllt6YW5p8/PV..OPK4ws6c8vZEKiksuhLIPyiinExg3.y', -- admin123 (valid bcrypt hash)
    'NAT Administrator',
    'admin@nat-management.local',
    'Administrator',
    true
) ON CONFLICT (username) DO NOTHING;

-- Head Branch 1 User
INSERT INTO users (username, password, full_name, email, role, is_active)
VALUES (
    'head1',
    '$2a$10$NuahW2wrTIqi1P1FdEomQu3R/5VTc46/HNZxSOsHBwXUKY7sW6CQC', -- head123 (valid bcrypt hash)
    'Head Branch 1 - NAT Manager',
    'head1@nat-management.local',
    'Head Branch 1',
    true
) ON CONFLICT (username) DO NOTHING;

-- Head Branch 2 User
INSERT INTO users (username, password, full_name, email, role, is_active)
VALUES (
    'head2',
    '$2a$10$NuahW2wrTIqi1P1FdEomQu3R/5VTc46/HNZxSOsHBwXUKY7sW6CQC', -- head123 (valid bcrypt hash)
    'Head Branch 2 - NAT Manager',
    'head2@nat-management.local',
    'Head Branch 2',
    true
) ON CONFLICT (username) DO NOTHING;

-- Head Branch 3 User
INSERT INTO users (username, password, full_name, email, role, is_active)
VALUES (
    'head3',
    '$2a$10$NuahW2wrTIqi1P1FdEomQu3R/5VTc46/HNZxSOsHBwXUKY7sW6CQC', -- head123 (valid bcrypt hash)
    'Head Branch 3 - NAT Manager',
    'head3@nat-management.local',
    'Head Branch 3',
    true
) ON CONFLICT (username) DO NOTHING;

-- ============================================================================
-- INSERT DEFAULT ROUTER ACCESS CONTROL
-- Defines which roles can access which routers
-- ============================================================================

-- Administrator: Full access to all routers (wildcard)
INSERT INTO router_access_control (role, router_name, permissions, description)
VALUES (
    'Administrator',
    '*',
    ARRAY['read', 'write', 'delete', 'manage'],
    'Full access to all routers and management functions'
) ON CONFLICT (role, router_name) DO NOTHING;

-- Head Branch 1: Access to SAMSAT and LANE1
INSERT INTO router_access_control (role, router_name, permissions, description)
VALUES
    ('Head Branch 1', 'SAMSAT', ARRAY['read', 'write'], 'Access to SAMSAT router'),
    ('Head Branch 1', 'LANE1', ARRAY['read', 'write'], 'Access to LANE1 router')
ON CONFLICT (role, router_name) DO NOTHING;

-- Head Branch 2: Access to LANE2 and LANE4
INSERT INTO router_access_control (role, router_name, permissions, description)
VALUES
    ('Head Branch 2', 'LANE2', ARRAY['read', 'write'], 'Access to LANE2 router'),
    ('Head Branch 2', 'LANE4', ARRAY['read', 'write'], 'Access to LANE4 router')
ON CONFLICT (role, router_name) DO NOTHING;

-- Head Branch 3: Access to BT JAYA/PK JAYA and SUKAWANGI
INSERT INTO router_access_control (role, router_name, permissions, description)
VALUES
    ('Head Branch 3', 'BT JAYA/PK JAYA', ARRAY['read', 'write'], 'Access to BT JAYA/PK JAYA router'),
    ('Head Branch 3', 'SUKAWANGI', ARRAY['read', 'write'], 'Access to SUKAWANGI router')
ON CONFLICT (role, router_name) DO NOTHING;

-- ============================================================================
-- VERIFICATION QUERIES (for debugging)
-- ============================================================================
DO $$
DECLARE
    user_count INTEGER;
    access_count INTEGER;
BEGIN
    -- Count users
    SELECT COUNT(*) INTO user_count FROM users;

    -- Count access control rules
    SELECT COUNT(*) INTO access_count FROM router_access_control;

    -- Display results
    RAISE NOTICE '✅ Data seeding completed!';
    RAISE NOTICE '👥 Users created: %', user_count;
    RAISE NOTICE '🔐 Access control rules created: %', access_count;
    RAISE NOTICE '';
    RAISE NOTICE '📝 Default Login Credentials:';
    RAISE NOTICE '   Admin: username=admin, password=admin123';
    RAISE NOTICE '   Head1: username=head1, password=head123';
    RAISE NOTICE '   Head2: username=head2, password=head123';
    RAISE NOTICE '   Head3: username=head3, password=head123';
    RAISE NOTICE '';
    RAISE NOTICE '⚠️  IMPORTANT: Change default passwords in production!';
END $$;
