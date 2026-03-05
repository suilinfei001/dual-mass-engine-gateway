-- Azure DevOps Integration Migration Script
-- This script adds Azure-specific fields to executable_resources and removes cancelled_request_url from tasks

USE event_processor;

-- Add Azure-specific columns to executable_resources
ALTER TABLE executable_resources ADD COLUMN IF NOT EXISTS organization VARCHAR(128) COMMENT 'Azure DevOps organization name';
ALTER TABLE executable_resources ADD COLUMN IF NOT EXISTS project VARCHAR(128) COMMENT 'Azure DevOps project name';
ALTER TABLE executable_resources ADD COLUMN IF NOT EXISTS pipeline_id INT COMMENT 'Azure DevOps pipeline ID';
ALTER TABLE executable_resources ADD COLUMN IF NOT EXISTS pipeline_params JSON COMMENT 'Pipeline template parameters';
ALTER TABLE executable_resources ADD COLUMN IF NOT EXISTS microservice_name VARCHAR(128) COMMENT 'Microservice name';
ALTER TABLE executable_resources ADD COLUMN IF NOT EXISTS pod_name VARCHAR(128) COMMENT 'Pod name';
ALTER TABLE executable_resources ADD COLUMN IF NOT EXISTS repo_path TEXT COMMENT 'Repository path';
ALTER TABLE executable_resources ADD COLUMN IF NOT EXISTS creator_id INT COMMENT 'Creator user ID';
ALTER TABLE executable_resources ADD COLUMN IF NOT EXISTS creator_name VARCHAR(128) COMMENT 'Creator username';

-- Drop the deprecated resource_path column if it exists (no longer needed)
ALTER TABLE executable_resources DROP COLUMN IF EXISTS resource_path;

-- Add BuildID column to tasks table for tracking Azure pipeline builds
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS build_id INT COMMENT 'Azure DevOps build ID' AFTER request_url;
ALTER TABLE tasks ADD INDEX idx_build_id (build_id);

-- Drop the cancelled_request_url column (no longer needed with executor interface)
ALTER TABLE tasks DROP COLUMN IF EXISTS cancelled_request_url;
