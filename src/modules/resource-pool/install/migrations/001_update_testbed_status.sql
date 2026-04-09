-- Migration: Update testbed status enum
-- Date: 2025-01-XX
-- Description: Replace 'maintenance' status with 'deleted' status
--              Testbed is one-time use, should be marked as deleted after release

-- Modify the status enum
ALTER TABLE testbeds
MODIFY COLUMN status ENUM('available', 'allocated', 'in_use', 'releasing', 'deleted') DEFAULT 'available';

-- Update any existing 'maintenance' status testbeds to 'available'
-- (If any exist - they should be reviewed by admin)
UPDATE testbeds SET status = 'available' WHERE status = 'maintenance';
