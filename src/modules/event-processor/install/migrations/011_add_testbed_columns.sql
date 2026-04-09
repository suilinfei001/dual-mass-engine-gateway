-- Add Testbed related columns to tasks table
-- This is a migration script for deployment_deployment quality node enhancement

ALTER TABLE tasks
    ADD COLUMN testbed_uuid VARCHAR(64) COMMENT 'Testbed UUID (for release)' AFTER analyzing,
    ADD COLUMN testbed_ip VARCHAR(64) COMMENT 'Testbed IP address' AFTER testbed_uuid,
    ADD COLUMN ssh_user VARCHAR(128) COMMENT 'SSH username' AFTER testbed_ip,
    ADD COLUMN ssh_password VARCHAR(256) COMMENT 'SSH password' AFTER ssh_user,
    ADD COLUMN chart_url VARCHAR(512) COMMENT 'Chart URL (from basic_ci_all result)' AFTER ssh_password,
    ADD COLUMN allocation_uuid VARCHAR(64) COMMENT 'Allocation UUID (for release)' AFTER chart_url;

-- Add index for faster lookups
ALTER TABLE tasks ADD INDEX idx_testbed_uuid (testbed_uuid);
ALTER TABLE tasks ADD INDEX idx_allocation_uuid (allocation_uuid);
