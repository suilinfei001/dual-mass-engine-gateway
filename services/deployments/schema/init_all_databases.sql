-- 双引擎质量网关微服务 - 统一数据库初始化脚本
-- 版本: 2.0
-- 日期: 2026-03-24
-- 修复: 数据库名称和表结构与服务代码匹配

-- ============================================
-- Event Store 数据库 (服务使用 event_store)
-- ============================================

CREATE DATABASE IF NOT EXISTS event_store
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

USE event_store;

-- 事件表
CREATE TABLE IF NOT EXISTS events (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    uuid VARCHAR(255) UNIQUE NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    source VARCHAR(100),
    repo_id BIGINT,
    repo_name VARCHAR(255),
    repo_owner VARCHAR(100),
    pr_number INT,
    commit_sha VARCHAR(40),
    author VARCHAR(100),
    payload JSON,
    received_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_event_status (status),
    INDEX idx_event_type (event_type),
    INDEX idx_repo_name (repo_name),
    INDEX idx_pr_number (pr_number),
    INDEX idx_received_at (received_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 质量检查表
CREATE TABLE IF NOT EXISTS quality_checks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    event_uuid VARCHAR(255) NOT NULL,
    check_type VARCHAR(50) NOT NULL,
    check_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    result VARCHAR(20),
    score DECIMAL(5,2),
    details TEXT,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_event_uuid (event_uuid),
    INDEX idx_check_status (check_status),
    INDEX idx_check_type (check_type),
    FOREIGN KEY (event_uuid) REFERENCES events(uuid) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Task Scheduler 数据库 (服务使用 task_scheduler)
-- ============================================

CREATE DATABASE IF NOT EXISTS task_scheduler
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

USE task_scheduler;

-- 任务表
CREATE TABLE IF NOT EXISTS tasks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    task_id VARCHAR(255) UNIQUE NOT NULL,
    task_name VARCHAR(100) NOT NULL,
    event_id BIGINT NOT NULL,
    check_type VARCHAR(100),
    stage VARCHAR(50) NOT NULL,
    stage_order INT NOT NULL,
    check_order INT,
    execute_order INT NOT NULL,
    resource_id BIGINT,
    request_url TEXT,
    build_id BIGINT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    start_time TIMESTAMP NULL,
    end_time TIMESTAMP NULL,
    error_message TEXT,
    log_file_path TEXT,
    analyzing BOOLEAN NOT NULL DEFAULT FALSE,
    testbed_uuid VARCHAR(255),
    testbed_ip VARCHAR(50),
    ssh_user VARCHAR(100),
    ssh_password VARCHAR(100),
    chart_url TEXT,
    allocation_uuid VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_event_id (event_id),
    INDEX idx_status (status),
    INDEX idx_execute_order (execute_order),
    INDEX idx_analyzing (analyzing),
    INDEX idx_task_name (task_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 任务结果表
CREATE TABLE IF NOT EXISTS task_results (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    task_id BIGINT NOT NULL,
    check_type VARCHAR(100) NOT NULL,
    result VARCHAR(20) NOT NULL,
    output TEXT,
    extra JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_task_id (task_id),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 任务执行记录表
CREATE TABLE IF NOT EXISTS task_executions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    task_id BIGINT NOT NULL,
    execution_id VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL,
    error_message TEXT,
    INDEX idx_task_id (task_id),
    INDEX idx_execution_id (execution_id),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Resource Manager 数据库 (服务使用 resource_manager)
-- ============================================

CREATE DATABASE IF NOT EXISTS resource_manager
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

USE resource_manager;

-- 资源实例表 (匹配 ResourceInstance 模型)
CREATE TABLE IF NOT EXISTS resource_instances (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    uuid VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    ip_address VARCHAR(50),
    ssh_port INT DEFAULT 22,
    ssh_user VARCHAR(100),
    ssh_password VARCHAR(100),
    category_id BIGINT,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    created_by VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'available',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_category_id (category_id),
    INDEX idx_status (status),
    INDEX idx_uuid (uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 分类表 (匹配 Category 模型)
CREATE TABLE IF NOT EXISTS categories (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    uuid VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_uuid (uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 配额策略表 (匹配 QuotaPolicy 模型)
CREATE TABLE IF NOT EXISTS quota_policies (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    uuid VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL UNIQUE,
    category_id BIGINT,
    max_count INT NOT NULL DEFAULT 10,
    replenish_rate INT NOT NULL DEFAULT 1,
    replenish_unit VARCHAR(20) NOT NULL DEFAULT 'day',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_category_id (category_id),
    INDEX idx_uuid (uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 分配历史表 (匹配 Allocation 模型)
CREATE TABLE IF NOT EXISTS allocations (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    uuid VARCHAR(255) UNIQUE NOT NULL,
    resource_id BIGINT NOT NULL,
    task_id BIGINT,
    policy_id BIGINT,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    allocated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    released_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_resource_id (resource_id),
    INDEX idx_task_id (task_id),
    INDEX idx_status (status),
    INDEX idx_uuid (uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Testbed 表 (匹配 Testbed 模型)
CREATE TABLE IF NOT EXISTS testbeds (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    uuid VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    resource_id BIGINT,
    status VARCHAR(20) NOT NULL DEFAULT 'available',
    ip_address VARCHAR(50),
    ssh_user VARCHAR(100),
    ssh_password VARCHAR(100),
    allocation_uuid VARCHAR(255),
    last_health_check TIMESTAMP NULL,
    attributes JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_resource_id (resource_id),
    INDEX idx_status (status),
    INDEX idx_uuid (uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 部署任务表
CREATE TABLE IF NOT EXISTS deployment_tasks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    uuid VARCHAR(255) UNIQUE NOT NULL,
    testbed_id BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    task_type VARCHAR(50) NOT NULL,
    config JSON,
    result JSON,
    error_message TEXT,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_testbed_id (testbed_id),
    INDEX idx_status (status),
    INDEX idx_uuid (uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_role (role)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- 插入默认数据
-- ============================================

USE resource_manager;

-- 默认分类
INSERT IGNORE INTO categories (id, uuid, name, description) VALUES
(1, UUID(), 'default', '默认资源分类'),
(2, UUID(), 'hardware', '硬件资源'),
(3, UUID(), 'software', '软件资源');

-- 默认配额策略
INSERT IGNORE INTO quota_policies (uuid, name, category_id, max_count, replenish_rate, replenish_unit) VALUES
(UUID(), 'default_policy', 1, 10, 1, 'day'),
(UUID(), 'high_priority', 1, 20, 2, 'day'),
(UUID(), 'low_priority', 1, 5, 1, 'day');

-- 默认用户
INSERT IGNORE INTO users (username, email, password_hash, role) VALUES
('admin', 'admin@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye/krJ9EEBP2NaYlVpPLqfXVj2Wd5qJ3G', 'admin'),
('system', 'system@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye/krJ9EEBP2NaYlVpPLqfXVj2Wd5qJ3G', 'system');

-- ============================================
-- 创建视图（可选）
-- ============================================

USE event_store;

-- 事件统计视图
CREATE OR REPLACE VIEW event_statistics AS
SELECT
    status,
    COUNT(*) as count
FROM events
GROUP BY status;

USE task_scheduler;

-- 任务统计视图
CREATE OR REPLACE VIEW task_statistics AS
SELECT
    status,
    COUNT(*) as count
FROM tasks
GROUP BY status;

USE resource_manager;

-- 资源利用率视图
CREATE OR REPLACE VIEW resource_utilization AS
SELECT
    r.id,
    r.name,
    r.status,
    COUNT(DISTINCT a.id) as active_allocations
FROM resource_instances r
LEFT JOIN allocations a ON r.id = a.resource_id AND a.status = 'active'
GROUP BY r.id;

-- ============================================
-- 完成初始化
-- ============================================

SELECT 'Database initialization completed successfully!' AS message;
