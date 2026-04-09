-- Add system_configs table for storing configuration
CREATE TABLE IF NOT EXISTS system_configs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    config_key VARCHAR(64) NOT NULL UNIQUE,
    config_value TEXT NOT NULL,
    description VARCHAR(255),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_config_key (config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Insert default CMP configuration (empty values, will be set via API)
INSERT IGNORE INTO system_configs (config_key, config_value, description) VALUES
    ('cmp_api_url', 'http://devops-api.aishu.cn:8081', 'CMP API Base URL'),
    ('cmp_access_key', '', 'CMP API Access Key'),
    ('cmp_secret_key', '', 'CMP API Secret Key');
