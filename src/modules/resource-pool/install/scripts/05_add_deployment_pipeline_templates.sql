-- 模块部署管道模板表
-- 用于存储各模块在不同环境下的 Azure DevOps Pipeline 配置模板

CREATE TABLE IF NOT EXISTS deployment_pipeline_templates (
    id INT AUTO_INCREMENT PRIMARY KEY,
    module_name VARCHAR(64) NOT NULL COMMENT '模块名称',
    environment VARCHAR(32) NOT NULL COMMENT '环境名称: dev, staging, prod',
    build_task VARCHAR(128) NOT NULL COMMENT '构建任务名称（Azure Pipeline）',
    release_task VARCHAR(128) NOT NULL COMMENT '发布任务名称（Azure Pipeline）',
    description TEXT COMMENT '描述信息',
    enabled BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by VARCHAR(64) DEFAULT 'system' COMMENT '创建者',

    UNIQUE KEY uk_module_env (module_name, environment),
    INDEX idx_module_name (module_name),
    INDEX idx_environment (environment),
    INDEX idx_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='模块部署管道模板配置表';
