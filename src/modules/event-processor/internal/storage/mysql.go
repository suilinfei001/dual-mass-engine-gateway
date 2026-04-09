package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github-hub/event-processor/internal/models"

	_ "github.com/go-sql-driver/mysql"
)

// TaskStorage 任务存储接口
type TaskStorage interface {
	// CreateTask 创建新任务
	CreateTask(task *models.Task) error

	// GetTask 根据 ID 获取任务
	GetTask(id int) (*models.Task, error)

	// GetTaskByTaskID 根据 task_id 获取任务
	GetTaskByTaskID(taskID string) (*models.Task, error)

	// ListTasks 列出所有任务
	ListTasks() ([]*models.Task, error)

	// UpdateTask 更新任务
	UpdateTask(task *models.Task) error

	// DeleteTask 删除任务
	DeleteTask(id int) error

	// DeleteAllTasks 删除所有任务（包括任务结果）
	DeleteAllTasks() error

	// GetPendingTasks 获取待执行的任务（按 execute_order 排序）
	GetPendingTasks() ([]*models.Task, error)

	// GetRunningTasks 获取正在运行的任务
	GetRunningTasks() ([]*models.Task, error)

	// GetTasksByStatus 根据状态获取任务
	GetTasksByStatus(status models.TaskStatus) ([]*models.Task, error)

	// GetTasksByEventID 根据事件 ID 获取任务
	GetTasksByEventID(eventID int) ([]*models.Task, error)

	// GetLatestTaskByEventID 获取事件的最新任务（按execute_order降序）
	GetLatestTaskByEventID(eventID int) (*models.Task, error)

	// GetCompletedTasksByEventID 获取事件的已完成任务
	GetCompletedTasksByEventID(eventID int) ([]*models.Task, error)

	// SaveTaskResults 保存任务结果
	SaveTaskResults(taskID int, results []models.TaskResult) error

	// GetTaskResults 获取任务结果
	GetTaskResults(taskID int) ([]models.TaskResult, error)

	// MarkTaskCancelled 标记任务为取消
	MarkTaskCancelled(task *models.Task, reason string) error

	// MarkTaskTimeout 标记任务为超时
	MarkTaskTimeout(task *models.Task, reason string) error

	// MarkTaskSkipped 标记任务为跳过
	MarkTaskSkipped(task *models.Task, reason string) error

	// UpdateTaskAnalyzing 更新任务的分析状态
	UpdateTaskAnalyzing(taskID int, analyzing bool) error

	// IsTaskAnalyzing 检查任务是否正在进行AI分析
	IsTaskAnalyzing(taskID int) (bool, error)

	// TryStartAnalysis 尝试启动任务的 AI 分析（原子操作）
	// 只有当 analyzing = false 时才会设置为 true，并返回 true 表示成功
	TryStartAnalysis(taskID int) (bool, error)

	// TryStartAnalysisOrResetStale 尝试启动分析，如果分析已超时(>10分钟)则重置并重新开始
	TryStartAnalysisOrResetStale(taskID int) (bool, error)

	// TryMarkTaskRunning 尝试将任务标记为运行中（原子操作）
	// 只有当 status = 'pending' 时才会设置为 'running'，并返回 true 表示成功
	TryMarkTaskRunning(taskID int) (bool, error)

	// Close 关闭数据库连接
	Close() error
}

// MySQLTaskStorage MySQL 任务存储实现
type MySQLTaskStorage struct {
	db *sql.DB
}

// NewMySQLTaskStorage 创建 MySQL 任务存储
func NewMySQLTaskStorage(dsn string) (*MySQLTaskStorage, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 设置连接池参数以避免空闲连接被服务器关闭
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	// 连接最大存活时间（小于 MySQL wait_timeout，默认 8 小时）
	db.SetConnMaxLifetime(4 * time.Hour)
	// 空闲连接最大存活时间，10 分钟后回收空闲连接
	db.SetConnMaxIdleTime(10 * time.Minute)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &MySQLTaskStorage{db: db}, nil
}

// CreateTask 创建新任务
func (s *MySQLTaskStorage) CreateTask(task *models.Task) error {
	query := `
		INSERT INTO tasks (
			task_id, task_name, event_id, check_type, stage, stage_order,
			check_order, execute_order, resource_id, request_url, build_id, log_file_path, analyzing,
			testbed_uuid, testbed_ip, ssh_user, ssh_password, chart_url, allocation_uuid,
			status, start_time, end_time, error_message, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var startTime, endTime sql.NullTime
	if task.StartTime != nil && !task.StartTime.Time.IsZero() {
		startTime = sql.NullTime{Time: task.StartTime.Time, Valid: true}
	}
	if task.EndTime != nil && !task.EndTime.Time.IsZero() {
		endTime = sql.NullTime{Time: task.EndTime.Time, Valid: true}
	}

	var buildID sql.NullInt64
	if task.BuildID > 0 {
		buildID = sql.NullInt64{Int64: int64(task.BuildID), Valid: true}
	}

	var resourceID sql.NullInt64
	if task.ResourceID > 0 {
		resourceID = sql.NullInt64{Int64: int64(task.ResourceID), Valid: true}
	}

	result, err := s.db.Exec(
		query,
		task.TaskID, task.TaskName, task.EventID, task.CheckType, task.Stage,
		task.StageOrder, task.CheckOrder, task.ExecuteOrder, resourceID,
		task.RequestURL, buildID, task.LogFilePath, task.Analyzing,
		task.TestbedUUID, task.TestbedIP, task.SSHUser, task.SSHPassword, task.ChartURL, task.AllocationUUID,
		task.Status, startTime, endTime, task.ErrorMessage,
		task.CreatedAt, task.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	log.Printf("[CreateTask] Debug: taskName=%s, ChartURL='%s', TestbedIP='%s'", task.TaskName, task.ChartURL, task.TestbedIP)

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	task.ID = int(id)
	return nil
}

// GetTask 根据 ID 获取任务
func (s *MySQLTaskStorage) GetTask(id int) (*models.Task, error) {
	query := `
		SELECT id, task_id, task_name, event_id, check_type, stage, stage_order,
			check_order, execute_order, resource_id, request_url, build_id, log_file_path, analyzing,
			testbed_uuid, testbed_ip, ssh_user, ssh_password, chart_url, allocation_uuid,
			status, start_time, end_time, error_message, created_at, updated_at
		FROM tasks
		WHERE id = ?
	`

	task := &models.Task{}
	var startTime, endTime sql.NullTime
	var buildID sql.NullInt64
	var resourceID sql.NullInt64

	err := s.db.QueryRow(query, id).Scan(
		&task.ID, &task.TaskID, &task.TaskName, &task.EventID, &task.CheckType,
		&task.Stage, &task.StageOrder, &task.CheckOrder, &task.ExecuteOrder,
		&resourceID, &task.RequestURL, &buildID, &task.LogFilePath, &task.Analyzing,
		&task.TestbedUUID, &task.TestbedIP, &task.SSHUser, &task.SSHPassword, &task.ChartURL, &task.AllocationUUID,
		&task.Status, &startTime, &endTime, &task.ErrorMessage, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if startTime.Valid {
		task.StartTime = &models.LocalTime{Time: startTime.Time}
	}
	if endTime.Valid {
		task.EndTime = &models.LocalTime{Time: endTime.Time}
	}
	if buildID.Valid {
		task.BuildID = int(buildID.Int64)
	}
	if resourceID.Valid {
		task.ResourceID = int(resourceID.Int64)
	}

	return task, nil
}

// GetTaskByTaskID 根据 task_id 获取任务
func (s *MySQLTaskStorage) GetTaskByTaskID(taskID string) (*models.Task, error) {
	query := `
		SELECT id, task_id, task_name, event_id, check_type, stage, stage_order,
			check_order, execute_order, resource_id, request_url, build_id, log_file_path, analyzing,
			testbed_uuid, testbed_ip, ssh_user, ssh_password, chart_url, allocation_uuid,
			status, start_time, end_time, error_message, created_at, updated_at
		FROM tasks
		WHERE task_id = ?
	`

	task := &models.Task{}
	var startTime, endTime sql.NullTime
	var buildID sql.NullInt64
	var resourceID sql.NullInt64

	err := s.db.QueryRow(query, taskID).Scan(
		&task.ID, &task.TaskID, &task.TaskName, &task.EventID, &task.CheckType,
		&task.Stage, &task.StageOrder, &task.CheckOrder, &task.ExecuteOrder,
		&resourceID, &task.RequestURL, &buildID, &task.LogFilePath, &task.Analyzing,
		&task.TestbedUUID, &task.TestbedIP, &task.SSHUser, &task.SSHPassword, &task.ChartURL, &task.AllocationUUID,
		&task.Status, &startTime, &endTime, &task.ErrorMessage, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if startTime.Valid {
		task.StartTime = &models.LocalTime{Time: startTime.Time}
	}
	if endTime.Valid {
		task.EndTime = &models.LocalTime{Time: endTime.Time}
	}
	if buildID.Valid {
		task.BuildID = int(buildID.Int64)
	}
	if resourceID.Valid {
		task.ResourceID = int(resourceID.Int64)
	}

	return task, nil
}

// ListTasks 列出所有任务
func (s *MySQLTaskStorage) ListTasks() ([]*models.Task, error) {
	query := `
		SELECT id, task_id, task_name, event_id, check_type, stage, stage_order,
			check_order, execute_order, resource_id, request_url, build_id, log_file_path, analyzing,
			testbed_uuid, testbed_ip, ssh_user, ssh_password, chart_url, allocation_uuid,
			status, start_time, end_time, error_message, created_at, updated_at
		FROM tasks
		ORDER BY execute_order ASC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		var startTime, endTime sql.NullTime
		var buildID sql.NullInt64
		var resourceID sql.NullInt64

		err := rows.Scan(
			&task.ID, &task.TaskID, &task.TaskName, &task.EventID, &task.CheckType,
			&task.Stage, &task.StageOrder, &task.CheckOrder, &task.ExecuteOrder,
			&resourceID, &task.RequestURL, &buildID, &task.LogFilePath, &task.Analyzing,
			&task.TestbedUUID, &task.TestbedIP, &task.SSHUser, &task.SSHPassword, &task.ChartURL, &task.AllocationUUID,
			&task.Status, &startTime, &endTime, &task.ErrorMessage, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		if startTime.Valid {
			task.StartTime = &models.LocalTime{Time: startTime.Time}
		}
		if endTime.Valid {
			task.EndTime = &models.LocalTime{Time: endTime.Time}
		}
		if buildID.Valid {
			task.BuildID = int(buildID.Int64)
		}
		if resourceID.Valid {
			task.ResourceID = int(resourceID.Int64)
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// UpdateTask 更新任务
func (s *MySQLTaskStorage) UpdateTask(task *models.Task) error {
	query := `
		UPDATE tasks SET
			task_name = ?, event_id = ?, check_type = ?, stage = ?, stage_order = ?,
			check_order = ?, execute_order = ?, resource_id = ?, request_url = ?, build_id = ?, log_file_path = ?, analyzing = ?,
			testbed_uuid = ?, testbed_ip = ?, ssh_user = ?, ssh_password = ?, chart_url = ?, allocation_uuid = ?,
			status = ?, start_time = ?, end_time = ?, error_message = ?, updated_at = ?
		WHERE id = ?
	`

	var startTime, endTime sql.NullTime
	if task.StartTime != nil && !task.StartTime.Time.IsZero() {
		startTime = sql.NullTime{Time: task.StartTime.Time, Valid: true}
	}
	if task.EndTime != nil && !task.EndTime.Time.IsZero() {
		endTime = sql.NullTime{Time: task.EndTime.Time, Valid: true}
	}

	var buildID sql.NullInt64
	if task.BuildID > 0 {
		buildID = sql.NullInt64{Int64: int64(task.BuildID), Valid: true}
	}

	var resourceID sql.NullInt64
	if task.ResourceID > 0 {
		resourceID = sql.NullInt64{Int64: int64(task.ResourceID), Valid: true}
	}

	_, err := s.db.Exec(
		query,
		task.TaskName, task.EventID, task.CheckType, task.Stage, task.StageOrder,
		task.CheckOrder, task.ExecuteOrder, resourceID,
		task.RequestURL, buildID, task.LogFilePath, task.Analyzing,
		task.TestbedUUID, task.TestbedIP, task.SSHUser, task.SSHPassword, task.ChartURL, task.AllocationUUID,
		task.Status, startTime, endTime, task.ErrorMessage, task.UpdatedAt, task.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

// DeleteTask 删除任务
func (s *MySQLTaskStorage) DeleteTask(id int) error {
	query := `DELETE FROM tasks WHERE id = ?`

	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}

// DeleteAllTasks 删除所有任务（包括任务结果）
func (s *MySQLTaskStorage) DeleteAllTasks() error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.Exec("DELETE FROM task_results")
	if err != nil {
		return fmt.Errorf("failed to delete task results: %w", err)
	}

	_, err = tx.Exec("DELETE FROM tasks")
	if err != nil {
		return fmt.Errorf("failed to delete tasks: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetPendingTasks 获取待执行的任务（按 execute_order 排序）
func (s *MySQLTaskStorage) GetPendingTasks() ([]*models.Task, error) {
	query := `
		SELECT id, task_id, task_name, event_id, check_type, stage, stage_order,
			check_order, execute_order, resource_id, request_url, build_id, log_file_path, analyzing,
			testbed_uuid, testbed_ip, ssh_user, ssh_password, chart_url, allocation_uuid,
			status, start_time, end_time, error_message, created_at, updated_at
		FROM tasks
		WHERE status = 'pending'
		ORDER BY execute_order ASC
	`

	return s.listTasksByQuery(query)
}

// GetRunningTasks 获取正在运行的任务
func (s *MySQLTaskStorage) GetRunningTasks() ([]*models.Task, error) {
	query := `
		SELECT id, task_id, task_name, event_id, check_type, stage, stage_order,
			check_order, execute_order, resource_id, request_url, build_id, log_file_path, analyzing,
			testbed_uuid, testbed_ip, ssh_user, ssh_password, chart_url, allocation_uuid,
			status, start_time, end_time, error_message, created_at, updated_at
		FROM tasks
		WHERE status = 'running' AND analyzing = false
		ORDER BY start_time ASC
	`

	return s.listTasksByQuery(query)
}

// GetTasksByStatus 根据状态获取任务
func (s *MySQLTaskStorage) GetTasksByStatus(status models.TaskStatus) ([]*models.Task, error) {
	query := `
		SELECT id, task_id, task_name, event_id, check_type, stage, stage_order,
			check_order, execute_order, resource_id, request_url, build_id, log_file_path, analyzing,
			testbed_uuid, testbed_ip, ssh_user, ssh_password, chart_url, allocation_uuid,
			status, start_time, end_time, error_message, created_at, updated_at
		FROM tasks
		WHERE status = ?
		ORDER BY execute_order ASC
	`

	rows, err := s.db.Query(query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by status: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		var startTime, endTime sql.NullTime
		var buildID sql.NullInt64
		var resourceID sql.NullInt64

		err := rows.Scan(
			&task.ID, &task.TaskID, &task.TaskName, &task.EventID, &task.CheckType,
			&task.Stage, &task.StageOrder, &task.CheckOrder, &task.ExecuteOrder,
			&resourceID, &task.RequestURL, &buildID, &task.LogFilePath, &task.Analyzing,
			&task.TestbedUUID, &task.TestbedIP, &task.SSHUser, &task.SSHPassword, &task.ChartURL, &task.AllocationUUID,
			&task.Status, &startTime, &endTime, &task.ErrorMessage, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		if startTime.Valid {
			task.StartTime = &models.LocalTime{Time: startTime.Time}
		}
		if endTime.Valid {
			task.EndTime = &models.LocalTime{Time: endTime.Time}
		}
		if buildID.Valid {
			task.BuildID = int(buildID.Int64)
		}
		if resourceID.Valid {
			task.ResourceID = int(resourceID.Int64)
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetTasksByEventID 根据事件 ID 获取任务
func (s *MySQLTaskStorage) GetTasksByEventID(eventID int) ([]*models.Task, error) {
	query := `
		SELECT id, task_id, task_name, event_id, check_type, stage, stage_order,
			check_order, execute_order, resource_id, request_url, build_id, log_file_path, analyzing,
			testbed_uuid, testbed_ip, ssh_user, ssh_password, chart_url, allocation_uuid,
			status, start_time, end_time, error_message, created_at, updated_at
		FROM tasks
		WHERE event_id = ?
		ORDER BY execute_order ASC
	`

	rows, err := s.db.Query(query, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by event id: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		var startTime, endTime sql.NullTime
		var buildID sql.NullInt64
		var resourceID sql.NullInt64

		err := rows.Scan(
			&task.ID, &task.TaskID, &task.TaskName, &task.EventID, &task.CheckType,
			&task.Stage, &task.StageOrder, &task.CheckOrder, &task.ExecuteOrder,
			&resourceID, &task.RequestURL, &buildID, &task.LogFilePath, &task.Analyzing,
			&task.TestbedUUID, &task.TestbedIP, &task.SSHUser, &task.SSHPassword, &task.ChartURL, &task.AllocationUUID,
			&task.Status, &startTime, &endTime, &task.ErrorMessage, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		if startTime.Valid {
			task.StartTime = &models.LocalTime{Time: startTime.Time}
		}
		if endTime.Valid {
			task.EndTime = &models.LocalTime{Time: endTime.Time}
		}
		if buildID.Valid {
			task.BuildID = int(buildID.Int64)
		}
		if resourceID.Valid {
			task.ResourceID = int(resourceID.Int64)
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// Close 关闭数据库连接
func (s *MySQLTaskStorage) Close() error {
	return s.db.Close()
}

func (s *MySQLTaskStorage) DB() *sql.DB {
	return s.db
}

func (s *MySQLTaskStorage) listTasksByQuery(query string) ([]*models.Task, error) {
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		var startTime, endTime sql.NullTime
		var buildID sql.NullInt64
		var resourceID sql.NullInt64

		err := rows.Scan(
			&task.ID, &task.TaskID, &task.TaskName, &task.EventID, &task.CheckType,
			&task.Stage, &task.StageOrder, &task.CheckOrder, &task.ExecuteOrder,
			&resourceID, &task.RequestURL, &buildID, &task.LogFilePath, &task.Analyzing,
			&task.TestbedUUID, &task.TestbedIP, &task.SSHUser, &task.SSHPassword, &task.ChartURL, &task.AllocationUUID,
			&task.Status, &startTime, &endTime, &task.ErrorMessage, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		if startTime.Valid {
			task.StartTime = &models.LocalTime{Time: startTime.Time}
		}
		if endTime.Valid {
			task.EndTime = &models.LocalTime{Time: endTime.Time}
		}
		if buildID.Valid {
			task.BuildID = int(buildID.Int64)
		}
		if resourceID.Valid {
			task.ResourceID = int(resourceID.Int64)
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// MarkTaskRunning 标记任务为运行中
func (s *MySQLTaskStorage) MarkTaskRunning(task *models.Task) error {
	now := time.Now()
	query := `
		UPDATE tasks SET
			status = 'running',
			start_time = ?,
			updated_at = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(query, now, now, task.ID)
	if err != nil {
		return fmt.Errorf("failed to mark task running: %w", err)
	}

	task.Status = models.TaskStatusRunning
	task.StartTime = &models.LocalTime{Time: now}
	task.UpdatedAt = now
	return nil
}

// MarkTaskCompleted 标记任务为完成
func (s *MySQLTaskStorage) MarkTaskCompleted(task *models.Task, results []models.TaskResult) error {
	now := time.Now()
	query := `
		UPDATE tasks SET
			status = 'passed',
			end_time = ?,
			updated_at = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(query, now, now, task.ID)
	if err != nil {
		return fmt.Errorf("failed to mark task completed: %w", err)
	}

	task.Status = models.TaskStatusPassed
	task.EndTime = &models.LocalTime{Time: now}
	task.Results = results
	task.UpdatedAt = now
	return nil
}

// MarkTaskFailed 标记任务为失败
func (s *MySQLTaskStorage) MarkTaskFailed(task *models.Task, reason string) error {
	now := time.Now()
	query := `
		UPDATE tasks SET
			status = 'failed',
			end_time = ?,
			error_message = ?,
			updated_at = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(query, now, reason, now, task.ID)
	if err != nil {
		return fmt.Errorf("failed to mark task failed: %w", err)
	}

	task.Status = models.TaskStatusFailed
	task.EndTime = &models.LocalTime{Time: now}
	task.ErrorMessage = &reason
	task.UpdatedAt = now
	return nil
}

// MarkTaskCancelled 标记任务为取消
func (s *MySQLTaskStorage) MarkTaskCancelled(task *models.Task, reason string) error {
	now := time.Now()
	query := `
		UPDATE tasks SET
			status = 'cancelled',
			end_time = ?,
			error_message = ?,
			updated_at = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(query, now, reason, now, task.ID)
	if err != nil {
		return fmt.Errorf("failed to mark task cancelled: %w", err)
	}

	task.Status = models.TaskStatusCancelled
	task.EndTime = &models.LocalTime{Time: now}
	task.ErrorMessage = &reason
	task.UpdatedAt = now
	return nil
}

// MarkTaskTimeout 标记任务为超时
func (s *MySQLTaskStorage) MarkTaskTimeout(task *models.Task, reason string) error {
	now := time.Now()
	query := `
		UPDATE tasks SET
			status = 'timeout',
			end_time = ?,
			error_message = ?,
			updated_at = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(query, now, reason, now, task.ID)
	if err != nil {
		return fmt.Errorf("failed to mark task timeout: %w", err)
	}

	task.Status = models.TaskStatusTimeout
	task.EndTime = &models.LocalTime{Time: now}
	task.ErrorMessage = &reason
	task.UpdatedAt = now
	return nil
}

// MarkTaskSkipped 标记任务为跳过
func (s *MySQLTaskStorage) MarkTaskSkipped(task *models.Task, reason string) error {
	now := time.Now()
	query := `
		UPDATE tasks SET
			status = 'skipped',
			end_time = ?,
			error_message = ?,
			updated_at = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(query, now, reason, now, task.ID)
	if err != nil {
		return fmt.Errorf("failed to mark task skipped: %w", err)
	}

	task.Status = models.TaskStatusSkipped
	task.EndTime = &models.LocalTime{Time: now}
	task.ErrorMessage = &reason
	task.UpdatedAt = now
	return nil
}

// UpdateTaskAnalyzing 更新任务的分析状态
func (s *MySQLTaskStorage) UpdateTaskAnalyzing(taskID int, analyzing bool) error {
	query := `UPDATE tasks SET analyzing = ?, updated_at = NOW() WHERE id = ?`
	_, err := s.db.Exec(query, analyzing, taskID)
	if err != nil {
		return fmt.Errorf("failed to update task analyzing status: %w", err)
	}
	return nil
}

// TryStartAnalysis 尝试启动任务的 AI 分析（原子操作）
// 只有当 analyzing = false 时才会设置为 true，并返回 true 表示成功
// 如果 analyzing 已经是 true，则返回 false 表示已被其他进程占用
func (s *MySQLTaskStorage) TryStartAnalysis(taskID int) (bool, error) {
	query := `UPDATE tasks SET analyzing = true, updated_at = NOW() WHERE id = ? AND analyzing = false`
	result, err := s.db.Exec(query, taskID)
	if err != nil {
		return false, fmt.Errorf("failed to try start analysis: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get rows affected: %w", err)
	}
	log.Printf("[TryStartAnalysis] taskID=%d, rowsAffected=%d", taskID, rowsAffected)
	return rowsAffected > 0, nil
}

// TryStartAnalysisOrResetStale 尝试启动分析，如果分析已超时(>10分钟)则重置并重新开始
// 这用于处理executor崩溃导致analyzing标志一直为true的情况
func (s *MySQLTaskStorage) TryStartAnalysisOrResetStale(taskID int) (bool, error) {
	// First try normal CAS
	started, err := s.TryStartAnalysis(taskID)
	if err != nil {
		return false, err
	}
	if started {
		return true, nil
	}

	// Check if the analysis is stale (updated_at > 10 minutes ago)
	query := `SELECT updated_at FROM tasks WHERE id = ?`
	var updatedAt time.Time
	err = s.db.QueryRow(query, taskID).Scan(&updatedAt)
	if err != nil {
		return false, fmt.Errorf("failed to check analysis staleness: %w", err)
	}

	// If analysis has been running for more than 10 minutes, consider it stale
	if time.Since(updatedAt) > 10*time.Minute {
		log.Printf("[TryStartAnalysisOrResetStale] taskID=%d has stale analysis (updated_at=%s), resetting", taskID, updatedAt.Format(time.RFC3339))
		// Reset analyzing flag and try again
		resetQuery := `UPDATE tasks SET analyzing = false, updated_at = NOW() WHERE id = ? AND analyzing = true`
		_, err = s.db.Exec(resetQuery, taskID)
		if err != nil {
			return false, fmt.Errorf("failed to reset stale analysis: %w", err)
		}
		// Now try to start analysis again
		return s.TryStartAnalysis(taskID)
	}

	return false, nil
}

// TryMarkTaskRunning 尝试将任务标记为运行中（原子操作）
// 只有当 status = 'pending' 时才会设置为 'running'，并返回 true 表示成功
// 如果 status 已经不是 'pending'，则返回 false 表示已被其他进程占用
func (s *MySQLTaskStorage) TryMarkTaskRunning(taskID int) (bool, error) {
	query := `UPDATE tasks SET status = 'running', start_time = NOW(), updated_at = NOW() WHERE id = ? AND status = 'pending'`
	result, err := s.db.Exec(query, taskID)
	if err != nil {
		return false, fmt.Errorf("failed to try mark task running: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get rows affected: %w", err)
	}
	log.Printf("[TryMarkTaskRunning] taskID=%d, rowsAffected=%d", taskID, rowsAffected)
	return rowsAffected > 0, nil
}

// IsTaskAnalyzing 检查任务是否正在进行AI分析
func (s *MySQLTaskStorage) IsTaskAnalyzing(taskID int) (bool, error) {
	query := `SELECT analyzing FROM tasks WHERE id = ?`
	var analyzing bool
	err := s.db.QueryRow(query, taskID).Scan(&analyzing)
	if err != nil {
		return false, fmt.Errorf("failed to check analyzing status: %w", err)
	}
	return analyzing, nil
}

// GetLatestTaskByEventID 获取事件的最新任务（按execute_order降序）
func (s *MySQLTaskStorage) GetLatestTaskByEventID(eventID int) (*models.Task, error) {
	query := `
		SELECT id, task_id, task_name, event_id, check_type, stage, stage_order,
			check_order, execute_order, resource_id, request_url, build_id, log_file_path, analyzing,
			testbed_uuid, testbed_ip, ssh_user, ssh_password, chart_url, allocation_uuid,
			status, start_time, end_time, error_message, created_at, updated_at
		FROM tasks
		WHERE event_id = ?
		ORDER BY execute_order DESC
		LIMIT 1
	`

	task := &models.Task{}
	var startTime, endTime sql.NullTime
	var buildID sql.NullInt64
	var resourceID sql.NullInt64

	err := s.db.QueryRow(query, eventID).Scan(
		&task.ID, &task.TaskID, &task.TaskName, &task.EventID, &task.CheckType,
		&task.Stage, &task.StageOrder, &task.CheckOrder, &task.ExecuteOrder,
		&resourceID, &task.RequestURL, &buildID, &task.LogFilePath, &task.Analyzing,
		&task.TestbedUUID, &task.TestbedIP, &task.SSHUser, &task.SSHPassword, &task.ChartURL, &task.AllocationUUID,
		&task.Status, &startTime, &endTime, &task.ErrorMessage, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest task: %w", err)
	}

	if startTime.Valid {
		task.StartTime = &models.LocalTime{Time: startTime.Time}
	}
	if endTime.Valid {
		task.EndTime = &models.LocalTime{Time: endTime.Time}
	}
	if buildID.Valid {
		task.BuildID = int(buildID.Int64)
	}
	if resourceID.Valid {
		task.ResourceID = int(resourceID.Int64)
	}

	return task, nil
}

// GetCompletedTasksByEventID 获取事件的已完成任务
func (s *MySQLTaskStorage) GetCompletedTasksByEventID(eventID int) ([]*models.Task, error) {
	query := `
		SELECT id, task_id, task_name, event_id, check_type, stage, stage_order,
			check_order, execute_order, resource_id, request_url, build_id, log_file_path, analyzing,
			testbed_uuid, testbed_ip, ssh_user, ssh_password, chart_url, allocation_uuid,
			status, start_time, end_time, error_message, created_at, updated_at
		FROM tasks
		WHERE event_id = ? AND status IN ('passed', 'failed', 'timeout', 'cancelled', 'skipped')
		ORDER BY execute_order ASC
	`

	rows, err := s.db.Query(query, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get completed tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		var startTime, endTime sql.NullTime
		var buildID sql.NullInt64
		var resourceID sql.NullInt64

		err := rows.Scan(
			&task.ID, &task.TaskID, &task.TaskName, &task.EventID, &task.CheckType,
			&task.Stage, &task.StageOrder, &task.CheckOrder, &task.ExecuteOrder,
			&resourceID, &task.RequestURL, &buildID, &task.LogFilePath, &task.Analyzing,
			&task.TestbedUUID, &task.TestbedIP, &task.SSHUser, &task.SSHPassword, &task.ChartURL, &task.AllocationUUID,
			&task.Status, &startTime, &endTime, &task.ErrorMessage, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		if startTime.Valid {
			task.StartTime = &models.LocalTime{Time: startTime.Time}
		}
		if endTime.Valid {
			task.EndTime = &models.LocalTime{Time: endTime.Time}
		}
		if buildID.Valid {
			task.BuildID = int(buildID.Int64)
		}
		if resourceID.Valid {
			task.ResourceID = int(resourceID.Int64)
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// SaveTaskResults 保存任务结果
// 先删除该任务的所有旧结果，再保存新结果（避免重复）
func (s *MySQLTaskStorage) SaveTaskResults(taskID int, results []models.TaskResult) error {
	// 先删除旧结果
	_, err := s.db.Exec("DELETE FROM task_results WHERE task_id = ?", taskID)
	if err != nil {
		return fmt.Errorf("failed to delete old task results: %w", err)
	}

	// 保存新结果
	for _, result := range results {
		query := `
			INSERT INTO task_results (task_id, check_type, result, output, extra, created_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`

		var extraJSON []byte
		if result.Extra != nil {
			var err error
			extraJSON, err = json.Marshal(result.Extra)
			if err != nil {
				return fmt.Errorf("failed to marshal extra: %w", err)
			}
		}

		_, err := s.db.Exec(query, taskID, result.CheckType, result.Result, result.Output, extraJSON, time.Now())
		if err != nil {
			return fmt.Errorf("failed to save task result: %w", err)
		}
	}
	return nil
}

// GetTaskResults 获取任务结果
func (s *MySQLTaskStorage) GetTaskResults(taskID int) ([]models.TaskResult, error) {
	query := `
		SELECT check_type, result, output, extra
		FROM task_results
		WHERE task_id = ?
		ORDER BY id ASC
	`

	rows, err := s.db.Query(query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task results: %w", err)
	}
	defer rows.Close()

	var results []models.TaskResult
	for rows.Next() {
		var result models.TaskResult
		var extraJSON []byte

		err := rows.Scan(&result.CheckType, &result.Result, &result.Output, &extraJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task result: %w", err)
		}

		if len(extraJSON) > 0 {
			result.Extra = make(map[string]interface{})
			if err := json.Unmarshal(extraJSON, &result.Extra); err != nil {
				return nil, fmt.Errorf("failed to unmarshal extra: %w", err)
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// UpdateTaskURLsAndStatus 更新任务的请求 URL 和状态（用于重试 AI 匹配）
func (s *MySQLTaskStorage) UpdateTaskURLsAndStatus(task *models.Task) error {
	now := time.Now()
	query := `
		UPDATE tasks SET
			request_url = ?,
			build_id = ?,
			status = ?,
			error_message = NULL,
			updated_at = ?
		WHERE id = ?
	`

	var buildID sql.NullInt64
	if task.BuildID > 0 {
		buildID = sql.NullInt64{Int64: int64(task.BuildID), Valid: true}
	}

	_, err := s.db.Exec(query, task.RequestURL, buildID, task.Status, now, task.ID)
	if err != nil {
		return fmt.Errorf("failed to update task URLs and status: %w", err)
	}

	task.UpdatedAt = now
	return nil
}
