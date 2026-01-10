-- Update seeded admin password to a known value.
-- Migration: 000002_update_admin_password.up.sql

-- Password is: admin123
-- Hash generated with bcrypt cost 10.

INSERT INTO users (email, password_hash, first_name, last_name, role, is_active)
VALUES (
    'admin@heimdall.local',
    '$2a$10$IzlT8XUZJwucAVlFAAA2e.sGidmH9ii1hHBNSyqm4dApAHPHg1IZ6',
    'Admin',
    'User',
    'admin',
    true
)
ON CONFLICT (email) DO UPDATE
SET password_hash = EXCLUDED.password_hash,
    role          = EXCLUDED.role,
    is_active     = EXCLUDED.is_active,
    updated_at    = CURRENT_TIMESTAMP;
