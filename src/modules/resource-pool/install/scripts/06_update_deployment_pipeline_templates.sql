-- 更新部署管道模板表结构
-- 简化为只包含 Azure DevOps Pipeline 运行所需的核心参数

-- 删除旧表（如果存在）
DROP TABLE IF EXISTS deployment_pipeline_templates;

-- 创建新的简化表
CREATE TABLE IF NOT EXISTS deployment_pipeline_templates (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(128) NOT NULL COMMENT '模板名称（用于识别）',
    description TEXT COMMENT '描述信息',
    organization VARCHAR(128) NOT NULL COMMENT 'Azure DevOps 组织名',
    project VARCHAR(128) NOT NULL COMMENT 'Azure DevOps 项目名',
    pipeline_id INT NOT NULL COMMENT 'Azure DevOps Pipeline ID',
    pipeline_parameters JSON COMMENT 'Pipeline 参数（JSON 格式）',
    enabled BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by VARCHAR(64) DEFAULT 'system' COMMENT '创建者',

    INDEX idx_name (name),
    INDEX idx_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='模块部署管道模板配置表';
