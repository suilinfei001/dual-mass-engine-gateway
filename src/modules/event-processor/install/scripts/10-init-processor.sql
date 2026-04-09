-- Event Processor Database Initialization Script
-- This script initializes database for event-processor module

CREATE DATABASE IF NOT EXISTS event_processor DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE event_processor;

-- Tasks table
CREATE TABLE IF NOT EXISTS tasks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    task_id VARCHAR(64),
    task_name VARCHAR(128) NOT NULL,
    event_id INT NOT NULL,
    check_type VARCHAR(64),
    stage VARCHAR(64),
    stage_order INT,
    check_order INT,
    execute_order INT NOT NULL,
    resource_id INT COMMENT 'AI matched resource ID from executable_resources',
    request_url TEXT,
    build_id INT COMMENT 'Azure DevOps build ID',
    log_file_path VARCHAR(512) COMMENT 'Path to the full logs file for this task',
    analyzing BOOLEAN DEFAULT FALSE COMMENT 'AI analysis is in progress',
    testbed_uuid VARCHAR(64) COMMENT 'Testbed UUID (for release)',
    testbed_ip VARCHAR(64) COMMENT 'Testbed IP address',
    ssh_user VARCHAR(128) COMMENT 'SSH username',
    ssh_password VARCHAR(256) COMMENT 'SSH password',
    chart_url VARCHAR(512) COMMENT 'Chart URL (from basic_ci_all result)',
    allocation_uuid VARCHAR(64) COMMENT 'Allocation UUID (for release)',
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    start_time DATETIME,
    end_time DATETIME,
    error_message TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_event_id (event_id),
    INDEX idx_status (status),
    INDEX idx_execute_order (execute_order),
    INDEX idx_task_name (task_name),
    INDEX idx_build_id (build_id),
    INDEX idx_resource_id (resource_id),
    INDEX idx_testbed_uuid (testbed_uuid),
    INDEX idx_allocation_uuid (allocation_uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Task results table (for tasks with multiple check results, e.g., basic_ci_all)
CREATE TABLE IF NOT EXISTS task_results (
    id INT AUTO_INCREMENT PRIMARY KEY,
    task_id INT NOT NULL,
    check_type VARCHAR(64) NOT NULL,
    result VARCHAR(32) NOT NULL,
    output TEXT,
    extra JSON,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_task_id (task_id),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Cleanup: Delete tasks older than 7 days
DELETE FROM tasks WHERE created_at < DATE_SUB(NOW(), INTERVAL 7 DAY);

-- System configs table
CREATE TABLE IF NOT EXISTS system_configs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    config_key VARCHAR(128) NOT NULL UNIQUE,
    config_value TEXT,
    description VARCHAR(256),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_config_key (config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(64) NOT NULL UNIQUE,
    password VARCHAR(256) NOT NULL,
    role VARCHAR(32) NOT NULL DEFAULT 'user',
    email VARCHAR(128),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_role (role)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    session_id VARCHAR(128) NOT NULL UNIQUE,
    user_id INT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_session_id (session_id),
    INDEX idx_user_id (user_id),
    INDEX idx_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Executable resources table
CREATE TABLE IF NOT EXISTS executable_resources (
    id INT AUTO_INCREMENT PRIMARY KEY,
    resource_name VARCHAR(128) NOT NULL,
    resource_type VARCHAR(64) NOT NULL,
    organization VARCHAR(128) COMMENT 'Azure DevOps organization name',
    project VARCHAR(128) COMMENT 'Azure DevOps project name',
    pipeline_id INT COMMENT 'Azure DevOps pipeline ID',
    pipeline_params JSON COMMENT 'Pipeline template parameters',
    microservice_name VARCHAR(128) COMMENT 'Microservice name',
    pod_name VARCHAR(128) COMMENT 'Pod name',
    repo_path TEXT COMMENT 'Repository path',
    description TEXT,
    creator_id INT COMMENT 'Creator user ID',
    creator_name VARCHAR(128) COMMENT 'Creator username',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_resource_name (resource_name),
    INDEX idx_resource_type (resource_type),
    INDEX idx_pipeline_id (pipeline_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default admin user (password: admin123)
INSERT INTO users (username, password, role, email) VALUES ('admin', '$2a$10$kIanJKluOroOgtTdBFtEfe3CYSY/7cufGQTFNcO8bBnLL4y0lbYBe', 'admin', 'admin@example.com') ON DUPLICATE KEY UPDATE username=username;

-- Migration: Add resource_id column if it doesn't exist
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS resource_id INT COMMENT 'AI matched resource ID from executable_resources' AFTER execute_order;
ALTER TABLE tasks ADD INDEX IF NOT EXISTS idx_resource_id (resource_id);

-- Migration: Add analyzing column if it doesn't exist
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS analyzing BOOLEAN DEFAULT FALSE COMMENT 'AI analysis is in progress' AFTER log_file_path;

-- Migration: Add allow_skip column to executable_resources
ALTER TABLE executable_resources ADD COLUMN IF NOT EXISTS allow_skip BOOLEAN DEFAULT FALSE COMMENT 'Allow this check to be skipped' AFTER resource_type;
ALTER TABLE executable_resources ADD INDEX IF NOT EXISTS idx_allow_skip (allow_skip);
