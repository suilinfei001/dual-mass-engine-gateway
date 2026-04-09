package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// DeploymentPipelineTemplate 部署管道模板
type DeploymentPipelineTemplate struct {
	ID                  int
	Name                string
	Description         string
	Organization        string
	Project             string
	PipelineID          int
	PipelineParameters  map[string]interface{}
	Enabled             bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
	CreatedBy           string
}

// DeploymentPipelineTemplateStorage 部署管道模板存储接口
type DeploymentPipelineTemplateStorage interface {
	// CreateTemplate 创建新模板
	CreateTemplate(template *DeploymentPipelineTemplate) error

	// GetTemplate 根据 ID 获取模板
	GetTemplate(id int) (*DeploymentPipelineTemplate, error)

	// ListTemplates 列出所有模板
	ListTemplates() ([]*DeploymentPipelineTemplate, error)

	// ListEnabledTemplates 列出启用的模板
	ListEnabledTemplates() ([]*DeploymentPipelineTemplate, error)

	// UpdateTemplate 更新模板
	UpdateTemplate(template *DeploymentPipelineTemplate) error

	// DeleteTemplate 删除模板
	DeleteTemplate(id int) error

	// EnableTemplate 启用模板
	EnableTemplate(id int) error

	// DisableTemplate 禁用模板
	DisableTemplate(id int) error
}

// MySQLDeploymentPipelineTemplateStorage MySQL 部署管道模板存储实现
type MySQLDeploymentPipelineTemplateStorage struct {
	db *sql.DB
}

// NewMySQLDeploymentPipelineTemplateStorage 创建 MySQL 部署管道模板存储
func NewMySQLDeploymentPipelineTemplateStorage(db *sql.DB) *MySQLDeploymentPipelineTemplateStorage {
	return &MySQLDeploymentPipelineTemplateStorage{db: db}
}

// CreateTemplate 创建新模板
func (s *MySQLDeploymentPipelineTemplateStorage) CreateTemplate(template *DeploymentPipelineTemplate) error {
	query := `
		INSERT INTO deployment_pipeline_templates (
			name, description, organization, project, pipeline_id,
			pipeline_parameters, enabled, created_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	paramsJSON, _ := json.Marshal(template.PipelineParameters)

	result, err := s.db.Exec(
		query,
		template.Name, template.Description, template.Organization,
		template.Project, template.PipelineID, paramsJSON,
		template.Enabled, template.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to create deployment pipeline template: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	template.ID = int(id)
	return nil
}

// GetTemplate 根据 ID 获取模板
func (s *MySQLDeploymentPipelineTemplateStorage) GetTemplate(id int) (*DeploymentPipelineTemplate, error) {
	query := `
		SELECT id, name, description, organization, project, pipeline_id,
			pipeline_parameters, enabled, created_at, updated_at, created_by
		FROM deployment_pipeline_templates
		WHERE id = ?
	`
	return s.scanTemplate(s.db.QueryRow(query, id))
}

// ListTemplates 列出所有模板
func (s *MySQLDeploymentPipelineTemplateStorage) ListTemplates() ([]*DeploymentPipelineTemplate, error) {
	query := `
		SELECT id, name, description, organization, project, pipeline_id,
			pipeline_parameters, enabled, created_at, updated_at, created_by
		FROM deployment_pipeline_templates
		ORDER BY name
	`
	return s.listTemplatesByQuery(query)
}

// ListEnabledTemplates 列出启用的模板
func (s *MySQLDeploymentPipelineTemplateStorage) ListEnabledTemplates() ([]*DeploymentPipelineTemplate, error) {
	query := `
		SELECT id, name, description, organization, project, pipeline_id,
			pipeline_parameters, enabled, created_at, updated_at, created_by
		FROM deployment_pipeline_templates
		WHERE enabled = TRUE
		ORDER BY name
	`
	return s.listTemplatesByQuery(query)
}

// UpdateTemplate 更新模板
func (s *MySQLDeploymentPipelineTemplateStorage) UpdateTemplate(template *DeploymentPipelineTemplate) error {
	query := `
		UPDATE deployment_pipeline_templates SET
			name = ?, description = ?, organization = ?, project = ?,
			pipeline_id = ?, pipeline_parameters = ?, enabled = ?
		WHERE id = ?
	`

	paramsJSON, _ := json.Marshal(template.PipelineParameters)

	_, err := s.db.Exec(
		query,
		template.Name, template.Description, template.Organization,
		template.Project, template.PipelineID, paramsJSON,
		template.Enabled, template.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update deployment pipeline template: %w", err)
	}

	return nil
}

// DeleteTemplate 删除模板
func (s *MySQLDeploymentPipelineTemplateStorage) DeleteTemplate(id int) error {
	query := `DELETE FROM deployment_pipeline_templates WHERE id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete deployment pipeline template: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("deployment pipeline template not found")
	}

	return nil
}

// EnableTemplate 启用模板
func (s *MySQLDeploymentPipelineTemplateStorage) EnableTemplate(id int) error {
	query := `UPDATE deployment_pipeline_templates SET enabled = TRUE WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to enable deployment pipeline template: %w", err)
	}
	return nil
}

// DisableTemplate 禁用模板
func (s *MySQLDeploymentPipelineTemplateStorage) DisableTemplate(id int) error {
	query := `UPDATE deployment_pipeline_templates SET enabled = FALSE WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to disable deployment pipeline template: %w", err)
	}
	return nil
}

// scanTemplate 扫描单行数据到 DeploymentPipelineTemplate 对象
func (s *MySQLDeploymentPipelineTemplateStorage) scanTemplate(row *sql.Row) (*DeploymentPipelineTemplate, error) {
	template := &DeploymentPipelineTemplate{
		PipelineParameters: make(map[string]interface{}),
	}

	var description sql.NullString
	var paramsJSON []byte

	err := row.Scan(
		&template.ID, &template.Name, &description,
		&template.Organization, &template.Project, &template.PipelineID,
		&paramsJSON,
		&template.Enabled, &template.CreatedAt, &template.UpdatedAt,
		&template.CreatedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("deployment pipeline template not found")
		}
		return nil, fmt.Errorf("failed to scan deployment pipeline template: %w", err)
	}

	if description.Valid {
		template.Description = description.String
	}
	if len(paramsJSON) > 0 {
		json.Unmarshal(paramsJSON, &template.PipelineParameters)
	}

	return template, nil
}

// listTemplatesByQuery 根据查询列出模板
func (s *MySQLDeploymentPipelineTemplateStorage) listTemplatesByQuery(query string, args ...interface{}) ([]*DeploymentPipelineTemplate, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query deployment pipeline templates: %w", err)
	}
	defer rows.Close()

	var templates []*DeploymentPipelineTemplate
	for rows.Next() {
		template := &DeploymentPipelineTemplate{
			PipelineParameters: make(map[string]interface{}),
		}
		var description sql.NullString
		var paramsJSON []byte

		err := rows.Scan(
			&template.ID, &template.Name, &description,
			&template.Organization, &template.Project, &template.PipelineID,
			&paramsJSON,
			&template.Enabled, &template.CreatedAt, &template.UpdatedAt,
			&template.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deployment pipeline template: %w", err)
		}

		if description.Valid {
			template.Description = description.String
		}
		if len(paramsJSON) > 0 {
			json.Unmarshal(paramsJSON, &template.PipelineParameters)
		}

		templates = append(templates, template)
	}

	return templates, nil
}
