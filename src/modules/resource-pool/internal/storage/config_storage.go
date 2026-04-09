package storage

import (
	"database/sql"
	"fmt"
)

// AIConfig AI 配置
type AIConfig struct {
	IP    string
	Model string
	Token string
}

// AzureConfig Azure DevOps 配置
type AzureConfig struct {
	Organization string
	Project      string
	PAT          string
	BaseURL      string
}

// ConfigStorage 系统配置存储接口
type ConfigStorage interface {
	// GetConfig 获取配置值
	GetConfig(key string) (string, error)
	// SetConfig 设置配置值
	SetConfig(key, value string) error
	// GetAllConfigs 获取所有配置
	GetAllConfigs() (map[string]string, error)
	// GetAIConfig 获取 AI 配置
	GetAIConfig() (*AIConfig, error)
	// GetAzureConfig 获取 Azure DevOps 配置
	GetAzureConfig() (*AzureConfig, error)
}

// MySQLConfigStorage MySQL 配置存储实现
type MySQLConfigStorage struct {
	db *sql.DB
}

// NewMySQLConfigStorage 创建 MySQL 配置存储
func NewMySQLConfigStorage(db *sql.DB) *MySQLConfigStorage {
	return &MySQLConfigStorage{db: db}
}

// GetConfig 获取配置值
func (s *MySQLConfigStorage) GetConfig(key string) (string, error) {
	query := `SELECT config_value FROM system_configs WHERE config_key = ?`
	var value string
	err := s.db.QueryRow(query, key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("config key not found: %s", key)
		}
		return "", fmt.Errorf("failed to get config: %w", err)
	}
	return value, nil
}

// SetConfig 设置配置值
func (s *MySQLConfigStorage) SetConfig(key, value string) error {
	query := `
		INSERT INTO system_configs (config_key, config_value, created_at, updated_at)
		VALUES (?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE config_value = ?, updated_at = NOW()
	`
	_, err := s.db.Exec(query, key, value, value)
	if err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}
	return nil
}

// GetAllConfigs 获取所有配置
func (s *MySQLConfigStorage) GetAllConfigs() (map[string]string, error) {
	query := `SELECT config_key, config_value FROM system_configs`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all configs: %w", err)
	}
	defer rows.Close()

	configs := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("failed to scan config: %w", err)
		}
		configs[key] = value
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating configs: %w", err)
	}

	return configs, nil
}

// GetAIConfig 获取 AI 配置
func (s *MySQLConfigStorage) GetAIConfig() (*AIConfig, error) {
	ip, _ := s.GetConfig("ai_ip")
	model, _ := s.GetConfig("ai_model")
	token, _ := s.GetConfig("ai_token")

	if ip == "" || model == "" || token == "" {
		return nil, fmt.Errorf("AI config not complete: missing ip, model, or token")
	}

	return &AIConfig{
		IP:    ip,
		Model: model,
		Token: token,
	}, nil
}

// GetAzureConfig 获取 Azure DevOps 配置
func (s *MySQLConfigStorage) GetAzureConfig() (*AzureConfig, error) {
	org, _ := s.GetConfig("azure_organization")
	project, _ := s.GetConfig("azure_project")
	pat, _ := s.GetConfig("azure_pat")
	baseURL, _ := s.GetConfig("azure_base_url")

	if org == "" || project == "" || pat == "" {
		return nil, fmt.Errorf("Azure config not complete: missing organization, project, or pat")
	}

	if baseURL == "" {
		baseURL = "https://devops.aishu.cn"
	}

	return &AzureConfig{
		Organization: org,
		Project:      project,
		PAT:          pat,
		BaseURL:      baseURL,
	}, nil
}
