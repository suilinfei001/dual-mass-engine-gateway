-- Executable Resources and System Configuration Tables
-- This script initializes tables for executable resources management

USE event_processor;

-- Extend users table to support email and normal user registration
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS email VARCHAR(128) UNIQUE,
ADD COLUMN IF NOT EXISTS full_name VARCHAR(128);

-- Executable Resources table
CREATE TABLE IF NOT EXISTS executable_resources (
    id INT AUTO_INCREMENT PRIMARY KEY,
    resource_name VARCHAR(128) NOT NULL,
    resource_type VARCHAR(64) NOT NULL,
    pipeline_id INT NOT NULL,
    pipeline_params JSON,
    microservice_name VARCHAR(128),
    pod_name VARCHAR(128),
    repo_path VARCHAR(512) NOT NULL,
    description TEXT,
    creator_id INT NOT NULL,
    creator_name VARCHAR(64) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_resource_type (resource_type),
    INDEX idx_creator_id (creator_id),
    INDEX idx_resource_name (resource_name),
    FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE utf8mb4_unicode_ci;

-- System Configuration table
CREATE TABLE IF NOT EXISTS system_configs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    config_key VARCHAR(64) NOT NULL UNIQUE,
    config_value TEXT NOT NULL,
    description VARCHAR(255),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_config_key (config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Insert default system configurations
INSERT INTO system_configs (config_key, config_value, description) VALUES
('ai_ip', '', 'AI model server IP address'),
('ai_model', '', 'AI model name for resource matching'),
('ai_token', '', 'AI API token'),
('azure_pat', '', 'Azure DevOps Personal Access Token'),
('event_receiver_ip', '10.4.111.141', 'Event receiver server IP address')
ON DUPLICATE KEY UPDATE config_key=config_key;

-- Modify tasks table: remove cancelled_request_url field
-- Note: In MySQL, we can't easily remove a column in a safe way, 
-- so we'll just stop using it in the code
-- ALTER TABLE tasks DROP COLUMN IF EXISTS cancelled_request_url;
