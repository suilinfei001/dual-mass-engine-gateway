-- Resource Manager Database Schema
-- 资源管理器数据库结构

-- 创建数据库
CREATE DATABASE IF NOT EXISTS resource_manager DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE resource_manager;

-- 类别表
CREATE TABLE IF NOT EXISTS categories (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 资源实例表
CREATE TABLE IF NOT EXISTS resource_instances (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    uuid VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    ip_address VARCHAR(45),
    ssh_port INT DEFAULT 22,
    ssh_user VARCHAR(100),
    ssh_password VARCHAR(255),
    category_id BIGINT NOT NULL,
    is_public BOOLEAN DEFAULT FALSE,
    created_by VARCHAR(100),
    status ENUM('pending', 'active', 'unreachable') DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE,
    INDEX idx_uuid (uuid),
    INDEX idx_category (category_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 配额策略表
CREATE TABLE IF NOT EXISTS quota_policies (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    category_id BIGINT NOT NULL,
    max_count INT NOT NULL DEFAULT 0,
    replenish_rate INT NOT NULL DEFAULT 0,
    replenish_unit ENUM('second', 'minute', 'hour', 'day') DEFAULT 'hour',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE,
    INDEX idx_category (category_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 资源分配记录表
CREATE TABLE IF NOT EXISTS allocations (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    resource_uuid VARCHAR(64) NOT NULL,
    policy_uuid VARCHAR(64),
    task_uuid VARCHAR(64) NOT NULL,
    allocated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    released_at TIMESTAMP NULL,
    status ENUM('active', 'released') DEFAULT 'active',
    INDEX idx_resource (resource_uuid),
    INDEX idx_task (task_uuid),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 测试床表
CREATE TABLE IF NOT EXISTS testbeds (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE,
    ip_address VARCHAR(45) NOT NULL,
    ssh_port INT DEFAULT 22,
    ssh_user VARCHAR(100),
    ssh_password VARCHAR(255),
    capacity INT DEFAULT 10,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 部署任务表
CREATE TABLE IF NOT EXISTS deployment_tasks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    instance_uuid VARCHAR(64) NOT NULL,
    policy_uuid VARCHAR(64),
    category_uuid VARCHAR(64),
    chart_url VARCHAR(500),
    status ENUM('pending', 'running', 'completed', 'failed') DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_instance (instance_uuid),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    full_name VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入初始数据
INSERT INTO categories (name, description) VALUES
('basic_ci_all', 'Basic CI 资源池'),
('deployment_deployment', 'Deployment 资源池'),
('specialized_tests_api_test', 'API 测试资源池'),
('specialized_tests_module_e2e', 'Module E2E 资源池'),
('specialized_tests_agent_e2e', 'Agent E2E 资源池'),
('specialized_tests_ai_e2e', 'AI E2E 资源池')
ON DUPLICATE KEY UPDATE description=VALUES(description);
