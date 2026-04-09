package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/hugoh/go-designs/resource-pool/internal/models"
)

// ResourceInstanceTaskStorage ResourceInstanceTask 存储接口
type ResourceInstanceTaskStorage interface {
	// CreateTask 创建新任务
	CreateTask(task *models.ResourceInstanceTask) error

	// GetTask 根据 ID 获取任务
	GetTask(id int) (*models.ResourceInstanceTask, error)

	// GetTaskByUUID 根据 UUID 获取任务
	GetTaskByUUID(uuid string) (*models.ResourceInstanceTask, error)

	// UpdateTask 更新任务
	UpdateTask(task *models.ResourceInstanceTask) error

	// UpdateTaskStatus 更新任务状态
	UpdateTaskStatus(uuid string, status models.TaskStatus) error

	// ListTasksByResourceInstance 列出资源实例的所有任务
	ListTasksByResourceInstance(resourceInstanceUUID string) ([]*models.ResourceInstanceTask, error)

	// ListTasksByResourceInstanceWithPagination 分页列出资源实例的任务
	ListTasksByResourceInstanceWithPagination(resourceInstanceUUID string, offset, limit int) ([]*models.ResourceInstanceTask, int, error)

	// ListTasksByType 按任务类型列出任务
	ListTasksByType(taskType models.TaskType) ([]*models.ResourceInstanceTask, error)

	// ListTasksByStatus 按状态列出任务
	ListTasksByStatus(status models.TaskStatus) ([]*models.ResourceInstanceTask, error)

	// ListTasksByQuotaPolicy 列出配额策略相关的任务
	ListTasksByQuotaPolicy(quotaPolicyUUID string) ([]*models.ResourceInstanceTask, error)

	// ListRunningTasks 列出所有运行中的任务
	ListRunningTasks() ([]*models.ResourceInstanceTask, error)

	// ListRecentTasks 列出最近的任务
	ListRecentTasks(limit int) ([]*models.ResourceInstanceTask, error)

	// ListFailedTasks 列出失败的任务
	ListFailedTasks(since time.Time) ([]*models.ResourceInstanceTask, error)

	// HasRunningTasksByResourceInstance 检查资源实例是否有运行中的任务
	HasRunningTasksByResourceInstance(resourceInstanceUUID string) (bool, error)

	// CountTasksByStatus 按状态统计任务数量
	CountTasksByStatus(status models.TaskStatus) (int, error)

	// TryStartTask CAS 操作尝试启动任务
	TryStartTask(uuid string) (bool, error)

	// DeleteTask 删除任务
	DeleteTask(id int) error

	// DeleteOldTasks 删除旧任务
	DeleteOldTasks(olderThan time.Time) (int, error)

	// GetTaskStatistics 获取任务统计信息
	GetTaskStatistics() (*TaskStatistics, error)
}

// TaskStatistics 任务统计信息
type TaskStatistics struct {
	Total             int                              `json:"total"`
	Pending           int                              `json:"pending"`
	Running           int                              `json:"running"`
	Completed         int                              `json:"completed"`
	Failed            int                              `json:"failed"`
	Cancelled         int                              `json:"cancelled"`
	ByType            map[models.TaskType]int          `json:"by_type"`
	ByTrigger         map[models.TriggerSource]int     `json:"by_trigger"`
	AverageDurationMs *int                             `json:"average_duration_ms"`
}

// MySQLResourceInstanceTaskStorage MySQL ResourceInstanceTask 存储实现
type MySQLResourceInstanceTaskStorage struct {
	db *sql.DB
}

// NewMySQLResourceInstanceTaskStorage 创建 MySQL ResourceInstanceTask 存储
func NewMySQLResourceInstanceTaskStorage(db *sql.DB) *MySQLResourceInstanceTaskStorage {
	return &MySQLResourceInstanceTaskStorage{db: db}
}

// CreateTask 创建新任务
func (s *MySQLResourceInstanceTaskStorage) CreateTask(task *models.ResourceInstanceTask) error {
	query := `
		INSERT INTO resource_instance_tasks (
			uuid, resource_instance_uuid, task_type, status, trigger_source,
			trigger_user, quota_policy_uuid, category_uuid, testbed_uuid, allocation_uuid,
			started_at, completed_at, duration_ms, success, error_code, error_message,
			result_details, retry_count, max_retries, parent_task_uuid, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	resultDetailsJSON, _ := task.ResultDetails.Value()
	var startedAt, completedAt interface{} = nil, nil
	if task.StartedAt != nil {
		startedAt = *task.StartedAt
	}
	if task.CompletedAt != nil {
		completedAt = *task.CompletedAt
	}

	result, err := s.db.Exec(
		query,
		task.UUID, task.ResourceInstanceUUID, task.TaskType, task.Status, task.TriggerSource,
		task.TriggerUser, task.QuotaPolicyUUID, task.CategoryUUID, task.TestbedUUID, task.AllocationUUID,
		startedAt, completedAt, task.DurationMs, task.Success, task.ErrorCode, task.ErrorMessage,
		resultDetailsJSON, task.RetryCount, task.MaxRetries, task.ParentTaskUUID,
		task.CreatedAt, task.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	task.ID = int(id)
	return nil
}

// GetTask 根据 ID 获取任务
func (s *MySQLResourceInstanceTaskStorage) GetTask(id int) (*models.ResourceInstanceTask, error) {
	query := `
		SELECT id, uuid, resource_instance_uuid, task_type, status, trigger_source,
			trigger_user, quota_policy_uuid, category_uuid, testbed_uuid, allocation_uuid,
			started_at, completed_at, duration_ms, success, error_code, error_message,
			result_details, retry_count, max_retries, parent_task_uuid, created_at, updated_at
		FROM resource_instance_tasks
		WHERE id = ?
	`

	return s.scanTask(s.db.QueryRow(query, id))
}

// GetTaskByUUID 根据 UUID 获取任务
func (s *MySQLResourceInstanceTaskStorage) GetTaskByUUID(uuid string) (*models.ResourceInstanceTask, error) {
	query := `
		SELECT id, uuid, resource_instance_uuid, task_type, status, trigger_source,
			trigger_user, quota_policy_uuid, category_uuid, testbed_uuid, allocation_uuid,
			started_at, completed_at, duration_ms, success, error_code, error_message,
			result_details, retry_count, max_retries, parent_task_uuid, created_at, updated_at
		FROM resource_instance_tasks
		WHERE uuid = ?
	`

	return s.scanTask(s.db.QueryRow(query, uuid))
}

// UpdateTask 更新任务
func (s *MySQLResourceInstanceTaskStorage) UpdateTask(task *models.ResourceInstanceTask) error {
	query := `
		UPDATE resource_instance_tasks SET
			status = ?, started_at = ?, completed_at = ?, duration_ms = ?,
			success = ?, error_code = ?, error_message = ?, result_details = ?,
			retry_count = ?, updated_at = ?
		WHERE id = ?
	`

	resultDetailsJSON, _ := task.ResultDetails.Value()
	var startedAt, completedAt interface{} = nil, nil
	if task.StartedAt != nil {
		startedAt = *task.StartedAt
	}
	if task.CompletedAt != nil {
		completedAt = *task.CompletedAt
	}

	_, err := s.db.Exec(
		query,
		task.Status, startedAt, completedAt, task.DurationMs,
		task.Success, task.ErrorCode, task.ErrorMessage, resultDetailsJSON,
		task.RetryCount, task.UpdatedAt, task.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

// UpdateTaskStatus 更新任务状态
func (s *MySQLResourceInstanceTaskStorage) UpdateTaskStatus(uuid string, status models.TaskStatus) error {
	query := `UPDATE resource_instance_tasks SET status = ?, updated_at = NOW() WHERE uuid = ?`
	_, err := s.db.Exec(query, status, uuid)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}
	return nil
}

// ListTasksByResourceInstance 列出资源实例的所有任务
func (s *MySQLResourceInstanceTaskStorage) ListTasksByResourceInstance(resourceInstanceUUID string) ([]*models.ResourceInstanceTask, error) {
	query := `
		SELECT id, uuid, resource_instance_uuid, task_type, status, trigger_source,
			trigger_user, quota_policy_uuid, category_uuid, testbed_uuid, allocation_uuid,
			started_at, completed_at, duration_ms, success, error_code, error_message,
			result_details, retry_count, max_retries, parent_task_uuid, created_at, updated_at
		FROM resource_instance_tasks
		WHERE resource_instance_uuid = ?
		ORDER BY created_at DESC
	`

	return s.listTasksByQuery(query, resourceInstanceUUID)
}

// ListTasksByResourceInstanceWithPagination 分页列出资源实例的任务
func (s *MySQLResourceInstanceTaskStorage) ListTasksByResourceInstanceWithPagination(resourceInstanceUUID string, offset, limit int) ([]*models.ResourceInstanceTask, int, error) {
	// 先获取总数
	countQuery := `SELECT COUNT(*) FROM resource_instance_tasks WHERE resource_instance_uuid = ?`
	var total int
	err := s.db.QueryRow(countQuery, resourceInstanceUUID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	query := `
		SELECT id, uuid, resource_instance_uuid, task_type, status, trigger_source,
			trigger_user, quota_policy_uuid, category_uuid, testbed_uuid, allocation_uuid,
			started_at, completed_at, duration_ms, success, error_code, error_message,
			result_details, retry_count, max_retries, parent_task_uuid, created_at, updated_at
		FROM resource_instance_tasks
		WHERE resource_instance_uuid = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	tasks, err := s.listTasksByQuery(query, resourceInstanceUUID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// ListTasksByType 按任务类型列出任务
func (s *MySQLResourceInstanceTaskStorage) ListTasksByType(taskType models.TaskType) ([]*models.ResourceInstanceTask, error) {
	query := `
		SELECT id, uuid, resource_instance_uuid, task_type, status, trigger_source,
			trigger_user, quota_policy_uuid, category_uuid, testbed_uuid, allocation_uuid,
			started_at, completed_at, duration_ms, success, error_code, error_message,
			result_details, retry_count, max_retries, parent_task_uuid, created_at, updated_at
		FROM resource_instance_tasks
		WHERE task_type = ?
		ORDER BY created_at DESC
	`

	return s.listTasksByQuery(query, taskType)
}

// ListTasksByStatus 按状态列出任务
func (s *MySQLResourceInstanceTaskStorage) ListTasksByStatus(status models.TaskStatus) ([]*models.ResourceInstanceTask, error) {
	query := `
		SELECT id, uuid, resource_instance_uuid, task_type, status, trigger_source,
			trigger_user, quota_policy_uuid, category_uuid, testbed_uuid, allocation_uuid,
			started_at, completed_at, duration_ms, success, error_code, error_message,
			result_details, retry_count, max_retries, parent_task_uuid, created_at, updated_at
		FROM resource_instance_tasks
		WHERE status = ?
		ORDER BY created_at DESC
	`

	return s.listTasksByQuery(query, status)
}

// ListTasksByQuotaPolicy 列出配额策略相关的任务
func (s *MySQLResourceInstanceTaskStorage) ListTasksByQuotaPolicy(quotaPolicyUUID string) ([]*models.ResourceInstanceTask, error) {
	query := `
		SELECT id, uuid, resource_instance_uuid, task_type, status, trigger_source,
			trigger_user, quota_policy_uuid, category_uuid, testbed_uuid, allocation_uuid,
			started_at, completed_at, duration_ms, success, error_code, error_message,
			result_details, retry_count, max_retries, parent_task_uuid, created_at, updated_at
		FROM resource_instance_tasks
		WHERE quota_policy_uuid = ?
		ORDER BY created_at DESC
	`

	return s.listTasksByQuery(query, quotaPolicyUUID)
}

// ListRunningTasks 列出所有运行中的任务
func (s *MySQLResourceInstanceTaskStorage) ListRunningTasks() ([]*models.ResourceInstanceTask, error) {
	query := `
		SELECT id, uuid, resource_instance_uuid, task_type, status, trigger_source,
			trigger_user, quota_policy_uuid, category_uuid, testbed_uuid, allocation_uuid,
			started_at, completed_at, duration_ms, success, error_code, error_message,
			result_details, retry_count, max_retries, parent_task_uuid, created_at, updated_at
		FROM resource_instance_tasks
		WHERE status = 'running'
		ORDER BY started_at ASC
	`

	return s.listTasksByQuery(query)
}

// HasRunningTasksByResourceInstance 检查资源实例是否有运行中的任务
func (s *MySQLResourceInstanceTaskStorage) HasRunningTasksByResourceInstance(resourceInstanceUUID string) (bool, error) {
	query := `
		SELECT COUNT(*) > 0
		FROM resource_instance_tasks
		WHERE resource_instance_uuid = ? AND status = 'running'
	`

	var hasRunning bool
	err := s.db.QueryRow(query, resourceInstanceUUID).Scan(&hasRunning)
	if err != nil {
		return false, fmt.Errorf("failed to check running tasks: %w", err)
	}

	return hasRunning, nil
}

// ListRecentTasks 列出最近的任务
func (s *MySQLResourceInstanceTaskStorage) ListRecentTasks(limit int) ([]*models.ResourceInstanceTask, error) {
	query := `
		SELECT id, uuid, resource_instance_uuid, task_type, status, trigger_source,
			trigger_user, quota_policy_uuid, category_uuid, testbed_uuid, allocation_uuid,
			started_at, completed_at, duration_ms, success, error_code, error_message,
			result_details, retry_count, max_retries, parent_task_uuid, created_at, updated_at
		FROM resource_instance_tasks
		ORDER BY created_at DESC
		LIMIT ?
	`

	return s.listTasksByQuery(query, limit)
}

// ListFailedTasks 列出失败的任务
func (s *MySQLResourceInstanceTaskStorage) ListFailedTasks(since time.Time) ([]*models.ResourceInstanceTask, error) {
	query := `
		SELECT id, uuid, resource_instance_uuid, task_type, status, trigger_source,
			trigger_user, quota_policy_uuid, category_uuid, testbed_uuid, allocation_uuid,
			started_at, completed_at, duration_ms, success, error_code, error_message,
			result_details, retry_count, max_retries, parent_task_uuid, created_at, updated_at
		FROM resource_instance_tasks
		WHERE status = 'failed' AND completed_at >= ?
		ORDER BY completed_at DESC
	`

	return s.listTasksByQuery(query, since)
}

// CountTasksByStatus 按状态统计任务数量
func (s *MySQLResourceInstanceTaskStorage) CountTasksByStatus(status models.TaskStatus) (int, error) {
	query := `SELECT COUNT(*) FROM resource_instance_tasks WHERE status = ?`
	var count int
	err := s.db.QueryRow(query, status).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count tasks by status: %w", err)
	}
	return count, nil
}

// TryStartTask CAS 操作尝试启动任务
func (s *MySQLResourceInstanceTaskStorage) TryStartTask(uuid string) (bool, error) {
	query := `
		UPDATE resource_instance_tasks
		SET status = 'running', started_at = NOW(), updated_at = NOW()
		WHERE uuid = ? AND status = 'pending'
	`
	result, err := s.db.Exec(query, uuid)
	if err != nil {
		return false, fmt.Errorf("failed to start task: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0, nil
}

// DeleteTask 删除任务
func (s *MySQLResourceInstanceTaskStorage) DeleteTask(id int) error {
	query := `DELETE FROM resource_instance_tasks WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

// DeleteOldTasks 删除旧任务
func (s *MySQLResourceInstanceTaskStorage) DeleteOldTasks(olderThan time.Time) (int, error) {
	query := `DELETE FROM resource_instance_tasks WHERE created_at < ? AND status IN ('completed', 'failed', 'cancelled')`
	result, err := s.db.Exec(query, olderThan)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old tasks: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}

// GetTaskStatistics 获取任务统计信息
func (s *MySQLResourceInstanceTaskStorage) GetTaskStatistics() (*TaskStatistics, error) {
	stats := &TaskStatistics{
		ByType:    make(map[models.TaskType]int),
		ByTrigger: make(map[models.TriggerSource]int),
	}

	// 总数和各状态数量
	statusQuery := `
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending,
			SUM(CASE WHEN status = 'running' THEN 1 ELSE 0 END) as running,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed,
			SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END) as cancelled
		FROM resource_instance_tasks
	`
	err := s.db.QueryRow(statusQuery).Scan(
		&stats.Total, &stats.Pending, &stats.Running,
		&stats.Completed, &stats.Failed, &stats.Cancelled,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get status statistics: %w", err)
	}

	// 按类型统计
	typeQuery := `
		SELECT task_type, COUNT(*) as count
		FROM resource_instance_tasks
		GROUP BY task_type
	`
	rows, err := s.db.Query(typeQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get type statistics: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var taskType models.TaskType
		var count int
		if err := rows.Scan(&taskType, &count); err != nil {
			continue
		}
		stats.ByType[taskType] = count
	}

	// 按触发来源统计
	triggerQuery := `
		SELECT trigger_source, COUNT(*) as count
		FROM resource_instance_tasks
		GROUP BY trigger_source
	`
	rows, err = s.db.Query(triggerQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get trigger statistics: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var triggerSource models.TriggerSource
		var count int
		if err := rows.Scan(&triggerSource, &count); err != nil {
			continue
		}
		stats.ByTrigger[triggerSource] = count
	}

	// 平均执行时长（只统计已完成的任务）
	avgQuery := `
		SELECT AVG(duration_ms)
		FROM resource_instance_tasks
		WHERE status = 'completed' AND duration_ms IS NOT NULL
	`
	err = s.db.QueryRow(avgQuery).Scan(&stats.AverageDurationMs)
	if err != nil {
		stats.AverageDurationMs = nil
	}

	return stats, nil
}

// scanTask 扫描单行数据到 ResourceInstanceTask 对象
func (s *MySQLResourceInstanceTaskStorage) scanTask(row *sql.Row) (*models.ResourceInstanceTask, error) {
	task := &models.ResourceInstanceTask{
		ResultDetails: make(models.ResultDetails),
	}

	var startedAt, completedAt sql.NullTime
	var success sql.NullBool
	var triggerUser, quotaPolicyUUID, categoryUUID, testbedUUID, allocationUUID sql.NullString
	var errorCode, errorMessage, parentTaskUUID sql.NullString
	var durationMs sql.NullInt64

	err := row.Scan(
		&task.ID, &task.UUID, &task.ResourceInstanceUUID, &task.TaskType, &task.Status, &task.TriggerSource,
		&triggerUser, &quotaPolicyUUID, &categoryUUID, &testbedUUID, &allocationUUID,
		&startedAt, &completedAt, &durationMs, &success, &errorCode, &errorMessage,
		&task.ResultDetails, &task.RetryCount, &task.MaxRetries, &parentTaskUUID,
		&task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to scan task: %w", err)
	}

	// 处理可空字段
	if startedAt.Valid {
		task.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}
	if durationMs.Valid {
		ms := int(durationMs.Int64)
		task.DurationMs = &ms
	}
	if success.Valid {
		b := success.Bool
		task.Success = &b
	}
	if triggerUser.Valid {
		s := triggerUser.String
		task.TriggerUser = &s
	}
	if quotaPolicyUUID.Valid {
		s := quotaPolicyUUID.String
		task.QuotaPolicyUUID = &s
	}
	if categoryUUID.Valid {
		s := categoryUUID.String
		task.CategoryUUID = &s
	}
	if testbedUUID.Valid {
		s := testbedUUID.String
		task.TestbedUUID = &s
	}
	if allocationUUID.Valid {
		s := allocationUUID.String
		task.AllocationUUID = &s
	}
	if errorCode.Valid {
		s := errorCode.String
		task.ErrorCode = &s
	}
	if errorMessage.Valid {
		s := errorMessage.String
		task.ErrorMessage = &s
	}
	if parentTaskUUID.Valid {
		s := parentTaskUUID.String
		task.ParentTaskUUID = &s
	}

	return task, nil
}

// listTasksByQuery 根据查询列出任务
func (s *MySQLResourceInstanceTaskStorage) listTasksByQuery(query string, args ...interface{}) ([]*models.ResourceInstanceTask, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.ResourceInstanceTask
	for rows.Next() {
		task := &models.ResourceInstanceTask{
			ResultDetails: make(models.ResultDetails),
		}

		var startedAt, completedAt sql.NullTime
		var success sql.NullBool
		var triggerUser, quotaPolicyUUID, categoryUUID, testbedUUID, allocationUUID sql.NullString
		var errorCode, errorMessage, parentTaskUUID sql.NullString
		var durationMs sql.NullInt64

		err := rows.Scan(
			&task.ID, &task.UUID, &task.ResourceInstanceUUID, &task.TaskType, &task.Status, &task.TriggerSource,
			&triggerUser, &quotaPolicyUUID, &categoryUUID, &testbedUUID, &allocationUUID,
			&startedAt, &completedAt, &durationMs, &success, &errorCode, &errorMessage,
			&task.ResultDetails, &task.RetryCount, &task.MaxRetries, &parentTaskUUID,
			&task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		// 处理可空字段
		if startedAt.Valid {
			task.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}
		if durationMs.Valid {
			ms := int(durationMs.Int64)
			task.DurationMs = &ms
		}
		if success.Valid {
			b := success.Bool
			task.Success = &b
		}
		if triggerUser.Valid {
			s := triggerUser.String
			task.TriggerUser = &s
		}
		if quotaPolicyUUID.Valid {
			s := quotaPolicyUUID.String
			task.QuotaPolicyUUID = &s
		}
		if categoryUUID.Valid {
			s := categoryUUID.String
			task.CategoryUUID = &s
		}
		if testbedUUID.Valid {
			s := testbedUUID.String
			task.TestbedUUID = &s
		}
		if allocationUUID.Valid {
			s := allocationUUID.String
			task.AllocationUUID = &s
		}
		if errorCode.Valid {
			s := errorCode.String
			task.ErrorCode = &s
		}
		if errorMessage.Valid {
			s := errorMessage.String
			task.ErrorMessage = &s
		}
		if parentTaskUUID.Valid {
			s := parentTaskUUID.String
			task.ParentTaskUUID = &s
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}
