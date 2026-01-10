-- Rollback initial schema
-- Migration: 000001_init.down.sql

DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS plugins;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;
