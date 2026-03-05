-- Event Processor Database Initialization Script
-- This script initializes the database for the event-processor module

CREATE DATABASE IF NOT EXISTS event_processor DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE event_processor;

-- Users table for authentication
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(64) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(32) NOT NULL DEFAULT 'admin',
    email VARCHAR(128) DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default admin user (password: admin123)
INSERT INTO users (username, password, role) 
VALUES ('admin', '$2a$10$cP8F0liOdyBdvz1Ky1OvyuUVU3JIvbHW9HL86Iqb30WK7OyPf/pG6', 'admin')
ON DUPLICATE KEY UPDATE username=username;

-- Sessions table for session management
CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(64) PRIMARY KEY,
    user_id INT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_expires_at (expires_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

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
    request_url TEXT,
    cancelled_request_url TEXT,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    start_time DATETIME,
    end_time DATETIME,
    error_message TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_event_id (event_id),
    INDEX idx_status (status),
    INDEX idx_execute_order (execute_order),
    INDEX idx_task_name (task_name)
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

-- Executable resources table
CREATE TABLE IF NOT EXISTS executable_resources (
    id INT AUTO_INCREMENT PRIMARY KEY,
    resource_name VARCHAR(128) NOT NULL,
    resource_type VARCHAR(64) NOT NULL,
    pipeline_id VARCHAR(64) NOT NULL,
    pipeline_params TEXT NOT NULL,
    microservice_name VARCHAR(128) DEFAULT NULL,
    pod_name VARCHAR(128) DEFAULT NULL,
    repo_path VARCHAR(255) NOT NULL,
    description TEXT DEFAULT NULL,
    creator_id INT NOT NULL,
    creator_name VARCHAR(64) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_creator_id (creator_id),
    INDEX idx_resource_name (resource_name),
    INDEX idx_resource_type (resource_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- System Configuration table
CREATE TABLE IF NOT EXISTS system_configs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    config_key VARCHAR(64) NOT NULL UNIQUE,
    config_value TEXT NOT NULL,
    description VARCHAR(255),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_config_key (config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default system configurations
INSERT INTO system_configs (config_key, config_value, description) VALUES
('ai_ip', '', 'AI model server IP address'),
('ai_model', '', 'AI model name for resource matching'),
('ai_token', '', 'AI API token'),
('azure_pat', '', 'Azure DevOps Personal Access Token'),
('event_receiver_ip', '', 'Event receiver server IP address')
ON DUPLICATE KEY UPDATE config_key=config_key;

-- Cleanup: Delete tasks older than 7 days
-- This can be run periodically to clean up old data
DELETE FROM tasks WHERE created_at < DATE_SUB(NOW(), INTERVAL 7 DAY);
