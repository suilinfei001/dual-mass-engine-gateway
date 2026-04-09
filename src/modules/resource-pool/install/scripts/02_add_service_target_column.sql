-- Migration: Add service_target column to quota_policies table
-- Date: 2025-03-10
-- Description: Adds service_target ENUM field to distinguish robot vs normal user quota policies

-- Check if column exists first
SET @column_exists = (
    SELECT COUNT(*)
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
    AND table_name = 'quota_policies'
    AND column_name = 'service_target'
);

-- Add the service_target column if it doesn't exist
SET @sql = IF(@column_exists = 0,
    'ALTER TABLE quota_policies ADD COLUMN service_target ENUM(''robot'', ''normal'') NOT NULL DEFAULT ''normal'' AFTER priority',
    'SELECT ''Column service_target already exists'' AS message'
);

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add index if it doesn't exist
SET @index_exists = (
    SELECT COUNT(*)
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
    AND table_name = 'quota_policies'
    AND index_name = 'idx_quota_policies_service_target'
);

SET @sql = IF(@index_exists = 0,
    'ALTER TABLE quota_policies ADD INDEX idx_quota_policies_service_target (service_target)',
    'SELECT ''Index already exists'' AS message'
);

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
