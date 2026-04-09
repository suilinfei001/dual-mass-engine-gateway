-- Migration: Update ResourceInstance schema to match lifecycle documentation
-- Date: 2025-01-11
-- Description:
--   1. Update status enum from ('active', 'terminating', 'terminated') to ('pending', 'active', 'unreachable')
--   2. Add missing ssh_user field (defaults to 'root')
--   3. Add missing updated_at field
--
-- Background:
--   - ResourceInstance is a persistent entity that can be reused
--   - New instances start in 'pending' state and become 'active' after health check
--   - Health check failures result in 'unreachable' state
--   - Testbed is one-time use (marked as 'deleted' after release)
--   - See docs/resource-instance-testbed-lifecycle.md for full lifecycle

-- Add ssh_user column if not exists
ALTER TABLE resource_instances
ADD COLUMN IF NOT EXISTS ssh_user VARCHAR(100) DEFAULT 'root' AFTER port;

-- Add updated_at column if not exists
ALTER TABLE resource_instances
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP AFTER created_at;

-- Update the status enum
-- Note: MySQL doesn't support ENUM modification in a single statement with IF NOT EXISTS
-- We'll need to drop and recreate the column in some cases

-- First, set all 'terminating' to 'pending' (transition state)
UPDATE resource_instances SET status = 'pending' WHERE status = 'terminating';

-- Set all 'terminated' to 'unreachable' (non-functional state)
UPDATE resource_instances SET status = 'unreachable' WHERE status = 'terminated';

-- Modify the status enum
ALTER TABLE resource_instances
MODIFY COLUMN status ENUM('pending', 'active', 'unreachable') DEFAULT 'pending';

-- Set default ssh_user for existing records
UPDATE resource_instances SET ssh_user = 'root' WHERE ssh_user IS NULL OR ssh_user = '';

-- Verification query (run after migration to check)
-- SELECT uuid, instance_type, status, ssh_user, created_at, updated_at FROM resource_instances;
