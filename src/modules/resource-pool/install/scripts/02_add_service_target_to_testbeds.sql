-- Add service_target column to testbeds table
-- This migration adds the service_target field to distinguish testbeds created for different service targets

-- Add the column (run without IF NOT EXISTS since MySQL doesn't support it in ALTER TABLE)
-- Note: If the column already exists, this will fail - that's expected
ALTER TABLE testbeds ADD COLUMN service_target ENUM('robot', 'normal') NOT NULL DEFAULT 'normal' AFTER category_uuid;

-- Add index for better query performance
-- Note: If indexes already exist, these will fail - that's expected
ALTER TABLE testbeds ADD INDEX idx_testbeds_category_service (category_uuid, service_target);
ALTER TABLE testbeds ADD INDEX idx_testbeds_status_service (status, service_target);
