-- Rollback admin password update.
-- Migration: 000002_update_admin_password.down.sql

UPDATE users
SET password_hash = '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    updated_at    = CURRENT_TIMESTAMP
WHERE email = 'admin@heimdall.local';
