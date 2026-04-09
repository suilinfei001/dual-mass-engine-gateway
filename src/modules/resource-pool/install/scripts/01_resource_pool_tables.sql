-- Resource Pool Management System Database Schema
-- 资源池管理系统数据库表结构

-- ============================================================
-- Categories (类别配置)
-- ============================================================
CREATE TABLE IF NOT EXISTS categories (
    id INT PRIMARY KEY AUTO_INCREMENT,
    uuid CHAR(36) UNIQUE NOT NULL,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_categories_enabled (enabled),
    INDEX idx_categories_uuid (uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- Quota Policies (配额策略)
-- ============================================================
CREATE TABLE IF NOT EXISTS quota_policies (
    id INT PRIMARY KEY AUTO_INCREMENT,
    uuid CHAR(36) UNIQUE NOT NULL,
    category_uuid CHAR(36) NOT NULL,
    min_instances INT NOT NULL DEFAULT 0,
    max_instances INT NOT NULL DEFAULT 10,
    priority INT NOT NULL DEFAULT 0,
    service_target ENUM('robot', 'normal') NOT NULL DEFAULT 'normal',
    auto_replenish BOOLEAN DEFAULT TRUE,
    replenish_threshold INT NOT NULL DEFAULT 1,
    max_lifetime_seconds INT NOT NULL DEFAULT 86400,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (category_uuid) REFERENCES categories(uuid) ON DELETE CASCADE,
    INDEX idx_quota_policies_category (category_uuid),
    INDEX idx_quota_policies_priority (priority),
    INDEX idx_quota_policies_service_target (service_target)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- Resource Instances (资源实例)
-- ============================================================
CREATE TABLE IF NOT EXISTS resource_instances (
    id INT PRIMARY KEY AUTO_INCREMENT,
    uuid CHAR(36) UNIQUE NOT NULL,
    instance_type ENUM('VirtualMachine', 'Machine') NOT NULL,
    snapshot_id VARCHAR(255),
    ip_address VARCHAR(50),
    port INT,
    ssh_user VARCHAR(100) DEFAULT 'root',
    passwd VARCHAR(255),
    description TEXT,
    is_public BOOLEAN DEFAULT TRUE,
    created_by VARCHAR(100) NOT NULL,
    status ENUM('pending', 'active', 'unreachable') DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    terminated_at TIMESTAMP NULL,
    INDEX idx_resource_instances_type (instance_type),
    INDEX idx_resource_instances_status (status),
    INDEX idx_resource_instances_public (is_public),
    INDEX idx_resource_instances_created_by (created_by),
    INDEX idx_resource_instances_uuid (uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- Testbeds (测试床)
-- ============================================================
CREATE TABLE IF NOT EXISTS testbeds (
    id INT PRIMARY KEY AUTO_INCREMENT,
    uuid CHAR(36) UNIQUE NOT NULL,
    name VARCHAR(100) UNIQUE NOT NULL,
    category_uuid CHAR(36) NOT NULL,
    service_target ENUM('robot', 'normal') NOT NULL DEFAULT 'normal',
    resource_instance_uuid CHAR(36) NOT NULL,
    current_alloc_uuid CHAR(36),
    mariadb_port INT,
    mariadb_user VARCHAR(100),
    mariadb_passwd VARCHAR(255),
    status ENUM('available', 'allocated', 'in_use', 'releasing', 'deleted') DEFAULT 'available',
    last_health_check TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (category_uuid) REFERENCES categories(uuid) ON DELETE CASCADE,
    FOREIGN KEY (resource_instance_uuid) REFERENCES resource_instances(uuid) ON DELETE CASCADE,
    INDEX idx_testbeds_category (category_uuid),
    INDEX idx_testbeds_category_service (category_uuid, service_target),
    INDEX idx_testbeds_status_service (status, service_target),
    INDEX idx_testbeds_resource (resource_instance_uuid),
    INDEX idx_testbeds_status (status),
    INDEX idx_testbeds_uuid (uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- Allocations (分配记录)
-- ============================================================
CREATE TABLE IF NOT EXISTS allocations (
    id INT PRIMARY KEY AUTO_INCREMENT,
    uuid CHAR(36) UNIQUE NOT NULL,
    testbed_uuid CHAR(36) NOT NULL,
    category_uuid CHAR(36) NOT NULL,
    requester VARCHAR(100) NOT NULL,
    requester_comment TEXT,
    status ENUM('pending', 'active', 'released', 'expired') DEFAULT 'pending',
    expires_at TIMESTAMP NULL,
    released_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (testbed_uuid) REFERENCES testbeds(uuid) ON DELETE CASCADE,
    FOREIGN KEY (category_uuid) REFERENCES categories(uuid) ON DELETE CASCADE,
    INDEX idx_allocations_testbed (testbed_uuid),
    INDEX idx_allocations_category (category_uuid),
    INDEX idx_allocations_requester (requester),
    INDEX idx_allocations_status (status),
    INDEX idx_allocations_expires (expires_at),
    INDEX idx_allocations_uuid (uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 初始化数据
-- ============================================================

-- 创建默认类别 (main 分支)
INSERT INTO categories (uuid, name, description, enabled)
VALUES (
    'cat-main-default',
    'main',
    '主分支回归测试环境',
    TRUE
) ON DUPLICATE KEY UPDATE name=name;

-- 为 main 类别创建默认配额策略 - robot 用户
INSERT INTO quota_policies (uuid, category_uuid, service_target, min_instances, max_instances, priority, auto_replenish, replenish_threshold, max_lifetime_seconds)
VALUES (
    'quota-main-robot',
    'cat-main-default',
    'robot',
    2,
    3,
    10,
    TRUE,
    2,
    86400
) ON DUPLICATE KEY UPDATE min_instances=min_instances;

-- 为 main 类别创建默认配额策略 - 普通用户
INSERT INTO quota_policies (uuid, category_uuid, service_target, min_instances, max_instances, priority, auto_replenish, replenish_threshold, max_lifetime_seconds)
VALUES (
    'quota-main-normal',
    'cat-main-default',
    'normal',
    0,
    5,
    100,
    TRUE,
    1,
    3600
) ON DUPLICATE KEY UPDATE min_instances=min_instances;

-- ============================================================
-- Resource Instance Tasks (资源实例任务)
-- 记录对资源实例的所有操作（部署、回滚等）
-- ============================================================
CREATE TABLE IF NOT EXISTS resource_instance_tasks (
    id INT PRIMARY KEY AUTO_INCREMENT,
    uuid CHAR(36) UNIQUE NOT NULL,
    resource_instance_uuid CHAR(36) NOT NULL COMMENT '关联的资源实例UUID',

    -- 任务类型和状态
    task_type ENUM('deploy', 'rollback', 'health_check') NOT NULL COMMENT '任务类型：部署/回滚/健康检查',
    status ENUM('pending', 'running', 'completed', 'failed', 'cancelled') NOT NULL DEFAULT 'pending' COMMENT '任务状态',

    -- 触发来源信息
    trigger_source ENUM('manual', 'auto_replenish', 'auto_expire', 'allocation_release', 'system_init') NOT NULL COMMENT '触发来源',
    trigger_user VARCHAR(100) COMMENT '触发任务的用户（手动触发时）',
    quota_policy_uuid CHAR(36) COMMENT '关联的配额策略UUID（自动补充时）',
    category_uuid CHAR(36) COMMENT '关联的类别UUID',

    -- 关联资源
    testbed_uuid CHAR(36) COMMENT '关联的Testbed UUID（deploy任务创建的testbed）',
    allocation_uuid CHAR(36) COMMENT '关联的Allocation UUID（release相关的rollback）',

    -- 执行时间信息
    started_at TIMESTAMP NULL COMMENT '开始执行时间',
    completed_at TIMESTAMP NULL COMMENT '完成时间',
    duration_ms INT COMMENT '执行时长（毫秒）',

    -- 执行结果
    success BOOLEAN DEFAULT NULL COMMENT '是否成功',
    error_code VARCHAR(50) COMMENT '错误代码',
    error_message TEXT COMMENT '详细错误信息',
    result_details JSON COMMENT '额外结果详情（如部署后的端口、凭证等）',

    -- 重试信息
    retry_count INT DEFAULT 0 COMMENT '已重试次数',
    max_retries INT DEFAULT 3 COMMENT '最大重试次数',
    parent_task_uuid CHAR(36) COMMENT '父任务UUID（重试任务关联到原始任务）',

    -- 元数据
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    -- 外键约束
    FOREIGN KEY (resource_instance_uuid) REFERENCES resource_instances(uuid) ON DELETE CASCADE,
    FOREIGN KEY (quota_policy_uuid) REFERENCES quota_policies(uuid) ON DELETE SET NULL,
    FOREIGN KEY (category_uuid) REFERENCES categories(uuid) ON DELETE SET NULL,
    FOREIGN KEY (testbed_uuid) REFERENCES testbeds(uuid) ON DELETE SET NULL,
    FOREIGN KEY (allocation_uuid) REFERENCES allocations(uuid) ON DELETE SET NULL,

    -- 索引
    INDEX idx_resource_instance (resource_instance_uuid),
    INDEX idx_task_type (task_type),
    INDEX idx_status (status),
    INDEX idx_trigger_source (trigger_source),
    INDEX idx_quota_policy (quota_policy_uuid),
    INDEX idx_category (category_uuid),
    INDEX idx_created_at (created_at),
    INDEX idx_completed_at (completed_at),
    INDEX idx_trigger_user (trigger_user),

    -- 组合索引用于常见查询
    INDEX idx_resource_status_type (resource_instance_uuid, status, task_type),
    INDEX idx_policy_status (quota_policy_uuid, status),
    INDEX idx_category_status_type (category_uuid, status, task_type),
    INDEX idx_trigger_source_status (trigger_source, status),
    INDEX idx_status_started (status, started_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
