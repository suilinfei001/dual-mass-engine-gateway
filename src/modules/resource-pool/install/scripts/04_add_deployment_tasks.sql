-- 部署任务表
-- 用于跟踪 Azure DevOps Pipeline 的异步部署任务

CREATE TABLE IF NOT EXISTS deployment_tasks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    task_uuid VARCHAR(64) NOT NULL UNIQUE COMMENT '任务UUID，用于外部引用',
    allocation_id INT NOT NULL COMMENT '关联的资源分配ID',
    pipeline_id INT NOT NULL COMMENT 'Azure DevOps Pipeline ID',
    build_id INT DEFAULT 0 COMMENT 'Azure DevOps Build ID',
    status VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '任务状态: pending, running, completed, failed, cancelled',
    analyzing BOOLEAN DEFAULT FALSE COMMENT '是否正在分析日志（CAS标志）',
    log_directory VARCHAR(255) DEFAULT '' COMMENT '日志存储目录路径',
    result_details JSON COMMENT 'AI分析结果（JSON格式）',
    error_message TEXT COMMENT '错误信息',
    web_url VARCHAR(512) DEFAULT '' COMMENT 'Azure DevOps Web URL',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_task_uuid (task_uuid),
    INDEX idx_allocation_id (allocation_id),
    INDEX idx_status (status),
    INDEX idx_build_id (build_id),
    INDEX idx_analyzing (analyzing)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='部署任务表';
