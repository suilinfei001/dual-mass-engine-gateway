package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// DeploymentTaskStatus 部署任务状态
type DeploymentTaskStatus string

const (
	DeploymentTaskStatusPending   DeploymentTaskStatus = "pending"
	DeploymentTaskStatusRunning   DeploymentTaskStatus = "running"
	DeploymentTaskStatusCompleted DeploymentTaskStatus = "completed"
	DeploymentTaskStatusFailed    DeploymentTaskStatus = "failed"
	DeploymentTaskStatusCancelled DeploymentTaskStatus = "cancelled"
)

// DeploymentTask 部署任务
type DeploymentTask struct {
	ID            int
	TaskUUID      string
	AllocationID  int
	PipelineID    int
	BuildID       int
	Status        DeploymentTaskStatus
	Analyzing     bool
	LogDirectory  string
	ResultDetails map[string]interface{}
	ErrorMessage  string
	WebURL        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// DeploymentTaskStorage 部署任务存储接口
type DeploymentTaskStorage interface {
	// CreateTask 创建新部署任务
	CreateTask(task *DeploymentTask) error

	// GetTask 根据 ID 获取任务
	GetTask(id int) (*DeploymentTask, error)

	// GetTaskByUUID 根据 UUID 获取任务
	GetTaskByUUID(uuid string) (*DeploymentTask, error)

	// GetTaskByAllocationID 根据分配ID获取任务
	GetTaskByAllocationID(allocationID int) (*DeploymentTask, error)

	// UpdateTask 更新任务
	UpdateTask(task *DeploymentTask) error

	// UpdateTaskStatus 更新任务状态
	UpdateTaskStatus(id int, status DeploymentTaskStatus) error

	// UpdateBuildInfo 更新 Azure Build 信息
	UpdateBuildInfo(id int, buildID int, webURL string) error

	// GetPendingTasks 获取待处理的任务
	GetPendingTasks() ([]*DeploymentTask, error)

	// GetRunningTasks 获取运行中的任务
	GetRunningTasks() ([]*DeploymentTask, error)

	// TryStartAnalysis CAS 操作尝试启动分析
	TryStartAnalysis(id int) (bool, error)

	// ResetAnalyzingFlag 重置分析标志
	ResetAnalyzingFlag(id int) error

	// ListRecentTasks 列出最近的任务
	ListRecentTasks(limit int) ([]*DeploymentTask, error)

	// DeleteOldTasks 删除旧任务
	DeleteOldTasks(olderThan time.Time) (int, error)
}

// MySQLDeploymentTaskStorage MySQL 部署任务存储实现
type MySQLDeploymentTaskStorage struct {
	db *sql.DB
}

// NewMySQLDeploymentTaskStorage 创建 MySQL 部署任务存储
func NewMySQLDeploymentTaskStorage(db *sql.DB) *MySQLDeploymentTaskStorage {
	return &MySQLDeploymentTaskStorage{db: db}
}

// CreateTask 创建新部署任务
func (s *MySQLDeploymentTaskStorage) CreateTask(task *DeploymentTask) error {
	query := `
		INSERT INTO deployment_tasks (
			task_uuid, allocation_id, pipeline_id, build_id, status, analyzing,
			log_directory, result_details, error_message, web_url, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	resultDetailsJSON, _ := json.Marshal(task.ResultDetails)

	result, err := s.db.Exec(
		query,
		task.TaskUUID, task.AllocationID, task.PipelineID, task.BuildID, task.Status,
		task.Analyzing, task.LogDirectory, resultDetailsJSON, task.ErrorMessage,
		task.WebURL, task.CreatedAt, task.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create deployment task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	task.ID = int(id)
	return nil
}

// GetTask 根据 ID 获取任务
func (s *MySQLDeploymentTaskStorage) GetTask(id int) (*DeploymentTask, error) {
	query := `
		SELECT id, task_uuid, allocation_id, pipeline_id, build_id, status, analyzing,
			log_directory, result_details, error_message, web_url, created_at, updated_at
		FROM deployment_tasks
		WHERE id = ?
	`
	return s.scanTask(s.db.QueryRow(query, id))
}

// GetTaskByUUID 根据 UUID 获取任务
func (s *MySQLDeploymentTaskStorage) GetTaskByUUID(uuid string) (*DeploymentTask, error) {
	query := `
		SELECT id, task_uuid, allocation_id, pipeline_id, build_id, status, analyzing,
			log_directory, result_details, error_message, web_url, created_at, updated_at
		FROM deployment_tasks
		WHERE task_uuid = ?
	`
	return s.scanTask(s.db.QueryRow(query, uuid))
}

// GetTaskByAllocationID 根据分配ID获取任务
func (s *MySQLDeploymentTaskStorage) GetTaskByAllocationID(allocationID int) (*DeploymentTask, error) {
	query := `
		SELECT id, task_uuid, allocation_id, pipeline_id, build_id, status, analyzing,
			log_directory, result_details, error_message, web_url, created_at, updated_at
		FROM deployment_tasks
		WHERE allocation_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`
	return s.scanTask(s.db.QueryRow(query, allocationID))
}

// UpdateTask 更新任务
func (s *MySQLDeploymentTaskStorage) UpdateTask(task *DeploymentTask) error {
	query := `
		UPDATE deployment_tasks SET
			status = ?, analyzing = ?, log_directory = ?, result_details = ?,
			error_message = ?, web_url = ?, updated_at = ?
		WHERE id = ?
	`

	resultDetailsJSON, _ := json.Marshal(task.ResultDetails)

	_, err := s.db.Exec(
		query,
		task.Status, task.Analyzing, task.LogDirectory, resultDetailsJSON,
		task.ErrorMessage, task.WebURL, task.UpdatedAt, task.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update deployment task: %w", err)
	}

	return nil
}

// UpdateTaskStatus 更新任务状态
func (s *MySQLDeploymentTaskStorage) UpdateTaskStatus(id int, status DeploymentTaskStatus) error {
	query := `UPDATE deployment_tasks SET status = ?, updated_at = NOW() WHERE id = ?`
	_, err := s.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}
	return nil
}

// UpdateBuildInfo 更新 Azure Build 信息
func (s *MySQLDeploymentTaskStorage) UpdateBuildInfo(id int, buildID int, webURL string) error {
	query := `UPDATE deployment_tasks SET build_id = ?, web_url = ?, updated_at = NOW() WHERE id = ?`
	_, err := s.db.Exec(query, buildID, webURL, id)
	if err != nil {
		return fmt.Errorf("failed to update build info: %w", err)
	}
	return nil
}

// GetPendingTasks 获取待处理的任务
func (s *MySQLDeploymentTaskStorage) GetPendingTasks() ([]*DeploymentTask, error) {
	query := `
		SELECT id, task_uuid, allocation_id, pipeline_id, build_id, status, analyzing,
			log_directory, result_details, error_message, web_url, created_at, updated_at
		FROM deployment_tasks
		WHERE status = 'pending'
		ORDER BY created_at ASC
	`
	return s.listTasksByQuery(query)
}

// GetRunningTasks 获取运行中的任务
func (s *MySQLDeploymentTaskStorage) GetRunningTasks() ([]*DeploymentTask, error) {
	query := `
		SELECT id, task_uuid, allocation_id, pipeline_id, build_id, status, analyzing,
			log_directory, result_details, error_message, web_url, created_at, updated_at
		FROM deployment_tasks
		WHERE status = 'running'
		ORDER BY created_at ASC
	`
	return s.listTasksByQuery(query)
}

// TryStartAnalysis CAS 操作尝试启动分析
func (s *MySQLDeploymentTaskStorage) TryStartAnalysis(id int) (bool, error) {
	query := `
		UPDATE deployment_tasks
		SET analyzing = true, updated_at = NOW()
		WHERE id = ? AND analyzing = false
	`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return false, fmt.Errorf("failed to start analysis: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0, nil
}

// ResetAnalyzingFlag 重置分析标志
func (s *MySQLDeploymentTaskStorage) ResetAnalyzingFlag(id int) error {
	query := `UPDATE deployment_tasks SET analyzing = false, updated_at = NOW() WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to reset analyzing flag: %w", err)
	}
	return nil
}

// ListRecentTasks 列出最近的任务
func (s *MySQLDeploymentTaskStorage) ListRecentTasks(limit int) ([]*DeploymentTask, error) {
	query := `
		SELECT id, task_uuid, allocation_id, pipeline_id, build_id, status, analyzing,
			log_directory, result_details, error_message, web_url, created_at, updated_at
		FROM deployment_tasks
		ORDER BY created_at DESC
		LIMIT ?
	`
	return s.listTasksByQuery(query, limit)
}

// DeleteOldTasks 删除旧任务
func (s *MySQLDeploymentTaskStorage) DeleteOldTasks(olderThan time.Time) (int, error) {
	query := `DELETE FROM deployment_tasks WHERE created_at < ? AND status IN ('completed', 'failed', 'cancelled')`
	result, err := s.db.Exec(query, olderThan)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old tasks: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}

// scanTask 扫描单行数据到 DeploymentTask 对象
func (s *MySQLDeploymentTaskStorage) scanTask(row *sql.Row) (*DeploymentTask, error) {
	task := &DeploymentTask{
		ResultDetails: make(map[string]interface{}),
	}

	var resultDetailsJSON []byte
	var logDirectory, errorMessage, webURL sql.NullString

	err := row.Scan(
		&task.ID, &task.TaskUUID, &task.AllocationID, &task.PipelineID, &task.BuildID,
		&task.Status, &task.Analyzing, &logDirectory, &resultDetailsJSON,
		&errorMessage, &webURL, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("deployment task not found")
		}
		return nil, fmt.Errorf("failed to scan deployment task: %w", err)
	}

	// 处理可空字段
	if logDirectory.Valid {
		task.LogDirectory = logDirectory.String
	}
	if errorMessage.Valid {
		task.ErrorMessage = errorMessage.String
	}
	if webURL.Valid {
		task.WebURL = webURL.String
	}
	if len(resultDetailsJSON) > 0 {
		json.Unmarshal(resultDetailsJSON, &task.ResultDetails)
	}

	return task, nil
}

// listTasksByQuery 根据查询列出任务
func (s *MySQLDeploymentTaskStorage) listTasksByQuery(query string, args ...interface{}) ([]*DeploymentTask, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query deployment tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*DeploymentTask
	for rows.Next() {
		task := &DeploymentTask{
			ResultDetails: make(map[string]interface{}),
		}

		var resultDetailsJSON []byte
		var logDirectory, errorMessage, webURL sql.NullString

		err := rows.Scan(
			&task.ID, &task.TaskUUID, &task.AllocationID, &task.PipelineID, &task.BuildID,
			&task.Status, &task.Analyzing, &logDirectory, &resultDetailsJSON,
			&errorMessage, &webURL, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deployment task: %w", err)
		}

		// 处理可空字段
		if logDirectory.Valid {
			task.LogDirectory = logDirectory.String
		}
		if errorMessage.Valid {
			task.ErrorMessage = errorMessage.String
		}
		if webURL.Valid {
			task.WebURL = webURL.String
		}
		if len(resultDetailsJSON) > 0 {
			json.Unmarshal(resultDetailsJSON, &task.ResultDetails)
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}
