package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	sharedmodels "github.com/quality-gateway/shared/pkg/models"
	"github.com/quality-gateway/task-scheduler/internal/models"
)

// TaskStorage 任务存储接口
type TaskStorage interface {
	// 任务基础操作
	CreateTask(task *models.Task) error
	GetTask(id int) (*models.Task, error)
	GetTasksByEventID(eventID int) ([]*models.Task, error)
	GetTasksByStatus(status sharedmodels.TaskStatus) ([]*models.Task, error)
	GetPendingTasks() ([]*models.Task, error)
	GetRunningTasks() ([]*models.Task, error)
	ListTasks(limit, offset int) ([]*models.Task, error)
	UpdateTask(task *models.Task) error
	DeleteTask(id int) error

	// 任务状态转换（CAS操作）
	TryMarkTaskRunning(taskID int) (bool, error)
	TryMarkTaskPassed(taskID int) (bool, error)
	TryMarkTaskFailed(taskID int) (bool, error)
	TryMarkTaskCancelled(taskID int) (bool, error)

	// 任务结果操作
	SaveTaskResults(taskID int, results []models.TaskResult) error
	GetTaskResults(taskID int) ([]models.TaskResult, error)
	DeleteTaskResults(taskID int) error

	// 任务分析状态操作
	UpdateTaskAnalyzing(taskID int, analyzing bool) error
	IsTaskAnalyzing(taskID int) (bool, error)
	TryStartAnalysis(taskID int) (bool, error)
	ResetStaleAnalysisTasks(timeout time.Duration) (int, error)

	// Testbed 相关操作
	UpdateTaskTestbed(taskID int, testbedUUID, testbedIP, sshUser, sshPassword, allocationUUID string) error
	ReleaseTestbed(taskID int) error

	// 事件级别操作
	CancelTasksByEventID(eventID int, reason string) (int, error)
	GetLatestTaskByEventID(eventID int) (*models.Task, error)

	// 统计操作
	GetTaskCountByStatus(status sharedmodels.TaskStatus) (int, error)
	GetTaskCountByEventID(eventID int) (int, error)
}

// MySQLStorage MySQL 实现
type MySQLStorage struct {
	db *sql.DB
}

// NewMySQLStorage 创建 MySQL 存储
func NewMySQLStorage(db *sql.DB) *MySQLStorage {
	return &MySQLStorage{db: db}
}

// CreateTask 创建任务
func (s *MySQLStorage) CreateTask(task *models.Task) error {
	query := `
		INSERT INTO tasks (
			task_id, task_name, event_id, check_type, stage, stage_order,
			check_order, execute_order, resource_id, request_url, build_id,
			status, log_file_path, analyzing, testbed_uuid, testbed_ip,
			ssh_user, ssh_password, chart_url, allocation_uuid
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.Exec(query,
		task.TaskID, task.TaskName, task.EventID, task.CheckType,
		task.Stage, task.StageOrder, task.CheckOrder, task.ExecuteOrder,
		task.ResourceID, task.RequestURL, task.BuildID, task.Status,
		task.LogFilePath, task.Analyzing, task.TestbedUUID, task.TestbedIP,
		task.SSHUser, task.SSHPassword, task.ChartURL, task.AllocationUUID,
	)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	task.ID = int(id)
	return nil
}

// GetTask 获取任务
func (s *MySQLStorage) GetTask(id int) (*models.Task, error) {
	query := `
		SELECT id, task_id, task_name, event_id, check_type, stage, stage_order,
		       check_order, execute_order, resource_id, request_url, build_id,
		       status, start_time, end_time, error_message, log_file_path,
		       analyzing, testbed_uuid, testbed_ip, ssh_user, ssh_password,
		       chart_url, allocation_uuid, created_at, updated_at
		FROM tasks WHERE id = ?
	`

	return s.scanTask(s.db.QueryRow(query, id))
}

// GetTasksByEventID 获取事件的所有任务
func (s *MySQLStorage) GetTasksByEventID(eventID int) ([]*models.Task, error) {
	query := `
		SELECT id, task_id, task_name, event_id, check_type, stage, stage_order,
		       check_order, execute_order, resource_id, request_url, build_id,
		       status, start_time, end_time, error_message, log_file_path,
		       analyzing, testbed_uuid, testbed_ip, ssh_user, ssh_password,
		       chart_url, allocation_uuid, created_at, updated_at
		FROM tasks WHERE event_id = ? ORDER BY execute_order
	`

	rows, err := s.db.Query(query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task, err := s.scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

// GetTasksByStatus 按状态获取任务
func (s *MySQLStorage) GetTasksByStatus(status sharedmodels.TaskStatus) ([]*models.Task, error) {
	query := `
		SELECT id, task_id, task_name, event_id, check_type, stage, stage_order,
		       check_order, execute_order, resource_id, request_url, build_id,
		       status, start_time, end_time, error_message, log_file_path,
		       analyzing, testbed_uuid, testbed_ip, ssh_user, ssh_password,
		       chart_url, allocation_uuid, created_at, updated_at
		FROM tasks WHERE status = ? ORDER BY execute_order
	`

	rows, err := s.db.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task, err := s.scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

// GetPendingTasks 获取待执行任务
func (s *MySQLStorage) GetPendingTasks() ([]*models.Task, error) {
	return s.GetTasksByStatus(sharedmodels.TaskStatusPending)
}

// GetRunningTasks 获取运行中的任务
func (s *MySQLStorage) GetRunningTasks() ([]*models.Task, error) {
	return s.GetTasksByStatus(sharedmodels.TaskStatusRunning)
}

// ListTasks 分页获取任务列表
func (s *MySQLStorage) ListTasks(limit, offset int) ([]*models.Task, error) {
	query := `
		SELECT id, task_id, task_name, event_id, check_type, stage, stage_order,
		       check_order, execute_order, resource_id, request_url, build_id,
		       status, start_time, end_time, error_message, log_file_path,
		       analyzing, testbed_uuid, testbed_ip, ssh_user, ssh_password,
		       chart_url, allocation_uuid, created_at, updated_at
		FROM tasks ORDER BY created_at DESC LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task, err := s.scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

// UpdateTask 更新任务
func (s *MySQLStorage) UpdateTask(task *models.Task) error {
	query := `
		UPDATE tasks SET
			status = ?,
			start_time = ?,
			end_time = ?,
			error_message = ?,
			log_file_path = ?,
			analyzing = ?,
			testbed_uuid = ?,
			testbed_ip = ?,
			ssh_user = ?,
			ssh_password = ?,
			chart_url = ?,
			allocation_uuid = ?,
			updated_at = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(query,
		task.Status, task.StartTime, task.EndTime, task.ErrorMessage,
		task.LogFilePath, task.Analyzing, task.TestbedUUID, task.TestbedIP,
		task.SSHUser, task.SSHPassword, task.ChartURL, task.AllocationUUID,
		time.Now(), task.ID,
	)

	return err
}

// DeleteTask 删除任务
func (s *MySQLStorage) DeleteTask(id int) error {
	_, err := s.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}

// TryMarkTaskRunning 尝试将任务标记为运行中（CAS操作）
func (s *MySQLStorage) TryMarkTaskRunning(taskID int) (bool, error) {
	query := `
		UPDATE tasks
		SET status = 'running', start_time = ?, updated_at = ?
		WHERE id = ? AND status = 'pending'
	`
	now := time.Now()
	result, err := s.db.Exec(query, now, now, taskID)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0, nil
}

// TryMarkTaskPassed 尝试将任务标记为通过（CAS操作）
func (s *MySQLStorage) TryMarkTaskPassed(taskID int) (bool, error) {
	query := `
		UPDATE tasks
		SET status = 'passed', end_time = ?, updated_at = ?
		WHERE id = ? AND status = 'running'
	`
	now := time.Now()
	result, err := s.db.Exec(query, now, now, taskID)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0, nil
}

// TryMarkTaskFailed 尝试将任务标记为失败（CAS操作）
func (s *MySQLStorage) TryMarkTaskFailed(taskID int) (bool, error) {
	query := `
		UPDATE tasks
		SET status = 'failed', end_time = ?, updated_at = ?
		WHERE id = ? AND status = 'running'
	`
	now := time.Now()
	result, err := s.db.Exec(query, now, now, taskID)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0, nil
}

// TryMarkTaskCancelled 尝试将任务标记为取消（CAS操作）
func (s *MySQLStorage) TryMarkTaskCancelled(taskID int) (bool, error) {
	query := `
		UPDATE tasks
		SET status = 'cancelled', end_time = ?, updated_at = ?
		WHERE id = ? AND status IN ('pending', 'running')
	`
	now := time.Now()
	result, err := s.db.Exec(query, now, now, taskID)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0, nil
}

// SaveTaskResults 保存任务结果
func (s *MySQLStorage) SaveTaskResults(taskID int, results []models.TaskResult) error {
	// 先删除旧结果
	if err := s.DeleteTaskResults(taskID); err != nil {
		return err
	}

	// 插入新结果
	query := `
		INSERT INTO task_results (task_id, check_type, result, output, extra)
		VALUES (?, ?, ?, ?, ?)
	`

	for _, result := range results {
		// Extra is already a string (JSON)

		_, err := s.db.Exec(query, taskID, result.CheckType, result.Result,
			result.Output, result.Extra)
		if err != nil {
			return fmt.Errorf("failed to save task result: %w", err)
		}
	}

	return nil
}

// GetTaskResults 获取任务结果
func (s *MySQLStorage) GetTaskResults(taskID int) ([]models.TaskResult, error) {
	query := `
		SELECT id, task_id, check_type, result, output, extra, created_at
		FROM task_results WHERE task_id = ? ORDER BY id
	`

	rows, err := s.db.Query(query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.TaskResult
	for rows.Next() {
		var r models.TaskResult
		var extraJSON sql.NullString
		err := rows.Scan(&r.ID, &r.TaskID, &r.CheckType, &r.Result,
			&r.Output, &extraJSON, &r.CreatedAt)
		if err != nil {
			return nil, err
		}

		if extraJSON.Valid && extraJSON.String != "" {
			json.Unmarshal([]byte(extraJSON.String), &r.Extra)
		}

		results = append(results, r)
	}

	return results, rows.Err()
}

// DeleteTaskResults 删除任务结果
func (s *MySQLStorage) DeleteTaskResults(taskID int) error {
	_, err := s.db.Exec("DELETE FROM task_results WHERE task_id = ?", taskID)
	return err
}

// UpdateTaskAnalyzing 更新任务分析状态
func (s *MySQLStorage) UpdateTaskAnalyzing(taskID int, analyzing bool) error {
	_, err := s.db.Exec(
		"UPDATE tasks SET analyzing = ?, updated_at = ? WHERE id = ?",
		analyzing, time.Now(), taskID,
	)
	return err
}

// IsTaskAnalyzing 检查任务是否正在分析
func (s *MySQLStorage) IsTaskAnalyzing(taskID int) (bool, error) {
	var analyzing bool
	err := s.db.QueryRow("SELECT analyzing FROM tasks WHERE id = ?", taskID).Scan(&analyzing)
	if err != nil {
		return false, err
	}
	return analyzing, nil
}

// TryStartAnalysis 尝试开始分析（CAS操作）
func (s *MySQLStorage) TryStartAnalysis(taskID int) (bool, error) {
	query := `
		UPDATE tasks
		SET analyzing = true, updated_at = ?
		WHERE id = ? AND analyzing = false
	`
	result, err := s.db.Exec(query, time.Now(), taskID)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0, nil
}

// ResetStaleAnalysisTasks 重置过期的分析任务
func (s *MySQLStorage) ResetStaleAnalysisTasks(timeout time.Duration) (int, error) {
	query := `
		UPDATE tasks
		SET analyzing = false, updated_at = ?
		WHERE analyzing = true AND updated_at < ?
	`
	cutoff := time.Now().Add(-timeout)
	result, err := s.db.Exec(query, time.Now(), cutoff)
	if err != nil {
		return 0, err
	}
	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}

// UpdateTaskTestbed 更新任务的 testbed 信息
func (s *MySQLStorage) UpdateTaskTestbed(taskID int, testbedUUID, testbedIP, sshUser, sshPassword, allocationUUID string) error {
	query := `
		UPDATE tasks SET
			testbed_uuid = ?,
			testbed_ip = ?,
			ssh_user = ?,
			ssh_password = ?,
			allocation_uuid = ?,
			updated_at = ?
		WHERE id = ?
	`
	_, err := s.db.Exec(query, testbedUUID, testbedIP, sshUser, sshPassword,
		allocationUUID, time.Now(), taskID)
	return err
}

// ReleaseTestbed 释放任务的 testbed
func (s *MySQLStorage) ReleaseTestbed(taskID int) error {
	query := `
		UPDATE tasks SET
			testbed_uuid = '',
			testbed_ip = '',
			allocation_uuid = '',
			updated_at = ?
		WHERE id = ?
	`
	_, err := s.db.Exec(query, time.Now(), taskID)
	return err
}

// CancelTasksByEventID 取消事件的所有任务
func (s *MySQLStorage) CancelTasksByEventID(eventID int, reason string) (int, error) {
	query := `
		UPDATE tasks
		SET status = 'cancelled',
		    error_message = ?,
		    end_time = ?,
		    updated_at = ?
		WHERE event_id = ? AND status IN ('pending', 'running')
	`
	now := time.Now()
	result, err := s.db.Exec(query, reason, now, now, eventID)
	if err != nil {
		return 0, err
	}
	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}

// GetLatestTaskByEventID 获取事件的最新任务
func (s *MySQLStorage) GetLatestTaskByEventID(eventID int) (*models.Task, error) {
	query := `
		SELECT id, task_id, task_name, event_id, check_type, stage, stage_order,
		       check_order, execute_order, resource_id, request_url, build_id,
		       status, start_time, end_time, error_message, log_file_path,
		       analyzing, testbed_uuid, testbed_ip, ssh_user, ssh_password,
		       chart_url, allocation_uuid, created_at, updated_at
		FROM tasks WHERE event_id = ? ORDER BY execute_order DESC LIMIT 1
	`

	return s.scanTask(s.db.QueryRow(query, eventID))
}

// GetTaskCountByStatus 获取指定状态的任务数量
func (s *MySQLStorage) GetTaskCountByStatus(status sharedmodels.TaskStatus) (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM tasks WHERE status = ?", status).Scan(&count)
	return count, err
}

// GetTaskCountByEventID 获取事件的任务数量
func (s *MySQLStorage) GetTaskCountByEventID(eventID int) (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM tasks WHERE event_id = ?", eventID).Scan(&count)
	return count, err
}

// scanTask 扫描任务行
func (s *MySQLStorage) scanTask(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.Task, error) {
	var task models.Task
	var errorMessage sql.NullString

	err := scanner.Scan(
		&task.ID, &task.TaskID, &task.TaskName, &task.EventID, &task.CheckType,
		&task.Stage, &task.StageOrder, &task.CheckOrder, &task.ExecuteOrder,
		&task.ResourceID, &task.RequestURL, &task.BuildID, &task.Status,
		&task.StartTime, &task.EndTime, &errorMessage, &task.LogFilePath,
		&task.Analyzing, &task.TestbedUUID, &task.TestbedIP, &task.SSHUser,
		&task.SSHPassword, &task.ChartURL, &task.AllocationUUID,
		&task.CreatedAt, &task.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if errorMessage.Valid {
		task.ErrorMessage = &errorMessage.String
	}

	return &task, nil
}
