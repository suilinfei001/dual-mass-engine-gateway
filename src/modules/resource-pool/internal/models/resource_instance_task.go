package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TaskType 任务类型
type TaskType string

const (
	TaskTypeDeploy      TaskType = "deploy"
	TaskTypeRollback    TaskType = "rollback"
	TaskTypeHealthCheck TaskType = "health_check"
)

// ValidTaskTypes 所有有效的任务类型
var ValidTaskTypes = map[string]TaskType{
	"deploy":       TaskTypeDeploy,
	"rollback":     TaskTypeRollback,
	"health_check": TaskTypeHealthCheck,
}

// ParseTaskType 解析任务类型字符串
func ParseTaskType(s string) (TaskType, error) {
	t, ok := ValidTaskTypes[s]
	if !ok {
		return "", fmt.Errorf("invalid task type: %s", s)
	}
	return t, nil
}

// DisplayName 返回任务类型的显示名称
func (t TaskType) DisplayName() string {
	switch t {
	case TaskTypeDeploy:
		return "部署"
	case TaskTypeRollback:
		return "回滚"
	case TaskTypeHealthCheck:
		return "健康检查"
	default:
		return string(t)
	}
}

// Value 实现 driver.Valuer 接口
func (t TaskType) Value() (driver.Value, error) {
	return string(t), nil
}

// Scan 实现 sql.Scanner 接口
func (t *TaskType) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*t = TaskType(v)
	case string:
		*t = TaskType(v)
	default:
		return fmt.Errorf("cannot scan %T into TaskType", value)
	}
	return nil
}

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// ValidTaskStatuses 所有有效的任务状态
var ValidTaskStatuses = map[string]TaskStatus{
	"pending":   TaskStatusPending,
	"running":   TaskStatusRunning,
	"completed": TaskStatusCompleted,
	"failed":    TaskStatusFailed,
	"cancelled": TaskStatusCancelled,
}

// ParseTaskStatus 解析任务状态字符串
func ParseTaskStatus(s string) (TaskStatus, error) {
	status, ok := ValidTaskStatuses[s]
	if !ok {
		return "", fmt.Errorf("invalid task status: %s", s)
	}
	return status, nil
}

// DisplayName 返回任务状态的显示名称
func (s TaskStatus) DisplayName() string {
	switch s {
	case TaskStatusPending:
		return "等待中"
	case TaskStatusRunning:
		return "执行中"
	case TaskStatusCompleted:
		return "已完成"
	case TaskStatusFailed:
		return "失败"
	case TaskStatusCancelled:
		return "已取消"
	default:
		return string(s)
	}
}

// Value 实现 driver.Valuer 接口
func (s TaskStatus) Value() (driver.Value, error) {
	return string(s), nil
}

// Scan 实现 sql.Scanner 接口
func (s *TaskStatus) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*s = TaskStatus(v)
	case string:
		*s = TaskStatus(v)
	default:
		return fmt.Errorf("cannot scan %T into TaskStatus", value)
	}
	return nil
}

// IsTerminal 检查状态是否为终态
func (s TaskStatus) IsTerminal() bool {
	return s == TaskStatusCompleted || s == TaskStatusFailed || s == TaskStatusCancelled
}

// TriggerSource 任务触发来源
type TriggerSource string

const (
	TriggerSourceManual            TriggerSource = "manual"
	TriggerSourceAutoReplenish     TriggerSource = "auto_replenish"
	TriggerSourceAutoExpire        TriggerSource = "auto_expire"
	TriggerSourceAllocationRelease TriggerSource = "allocation_release"
	TriggerSourceSystemInit        TriggerSource = "system_init"
)

// ValidTriggerSources 所有有效的触发来源
var ValidTriggerSources = map[string]TriggerSource{
	"manual":             TriggerSourceManual,
	"auto_replenish":     TriggerSourceAutoReplenish,
	"auto_expire":        TriggerSourceAutoExpire,
	"allocation_release": TriggerSourceAllocationRelease,
	"system_init":        TriggerSourceSystemInit,
}

// ParseTriggerSource 解析触发来源字符串
func ParseTriggerSource(s string) (TriggerSource, error) {
	source, ok := ValidTriggerSources[s]
	if !ok {
		return "", fmt.Errorf("invalid trigger source: %s", s)
	}
	return source, nil
}

// DisplayName 返回触发来源的显示名称
func (t TriggerSource) DisplayName() string {
	switch t {
	case TriggerSourceManual:
		return "手动触发"
	case TriggerSourceAutoReplenish:
		return "自动补充"
	case TriggerSourceAutoExpire:
		return "自动过期"
	case TriggerSourceAllocationRelease:
		return "分配释放"
	case TriggerSourceSystemInit:
		return "系统初始化"
	default:
		return string(t)
	}
}

// Value 实现 driver.Valuer 接口
func (t TriggerSource) Value() (driver.Value, error) {
	return string(t), nil
}

// Scan 实现 sql.Scanner 接口
func (t *TriggerSource) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*t = TriggerSource(v)
	case string:
		*t = TriggerSource(v)
	default:
		return fmt.Errorf("cannot scan %T into TriggerSource", value)
	}
	return nil
}

// ResultDetails 任务执行结果详情（存储在 JSON 字段中）
type ResultDetails map[string]interface{}

// Scan 实现 sql.Scanner 接口
func (r *ResultDetails) Scan(value interface{}) error {
	if value == nil {
		*r = make(ResultDetails)
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into ResultDetails", value)
	}

	return json.Unmarshal(data, r)
}

// Value 实现 driver.Valuer 接口
func (r ResultDetails) Value() (driver.Value, error) {
	if len(r) == 0 {
		return nil, nil
	}
	return json.Marshal(r)
}

// ResourceInstanceTask 资源实例任务
type ResourceInstanceTask struct {
	ID                   int            `json:"id"`
	UUID                 string         `json:"uuid"`
	ResourceInstanceUUID string         `json:"resource_instance_uuid"`
	TaskType             TaskType       `json:"task_type"`
	Status               TaskStatus     `json:"status"`
	TriggerSource        TriggerSource  `json:"trigger_source"`
	TriggerUser          *string        `json:"trigger_user,omitempty"`
	QuotaPolicyUUID      *string        `json:"quota_policy_uuid,omitempty"`
	CategoryUUID         *string        `json:"category_uuid,omitempty"`
	TestbedUUID          *string        `json:"testbed_uuid,omitempty"`
	AllocationUUID       *string        `json:"allocation_uuid,omitempty"`
	StartedAt            *time.Time     `json:"started_at,omitempty"`
	CompletedAt          *time.Time     `json:"completed_at,omitempty"`
	DurationMs           *int           `json:"duration_ms,omitempty"`
	Success              *bool          `json:"success,omitempty"`
	ErrorCode            *string        `json:"error_code,omitempty"`
	ErrorMessage         *string        `json:"error_message,omitempty"`
	ResultDetails        ResultDetails  `json:"result_details,omitempty"`
	RetryCount           int            `json:"retry_count"`
	MaxRetries           int            `json:"max_retries"`
	ParentTaskUUID       *string        `json:"parent_task_uuid,omitempty"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
}

// ResourceInstanceTaskResponse API 响应格式
type ResourceInstanceTaskResponse struct {
	ID                   int            `json:"id"`
	UUID                 string         `json:"uuid"`
	ResourceInstanceUUID string         `json:"resource_instance_uuid"`
	TaskType             TaskType       `json:"task_type"`
	TaskTypeName         string         `json:"task_type_name"`
	Status               TaskStatus     `json:"status"`
	StatusName           string         `json:"status_name"`
	TriggerSource        TriggerSource  `json:"trigger_source"`
	TriggerSourceName    string         `json:"trigger_source_name"`
	TriggerUser          *string        `json:"trigger_user,omitempty"`
	QuotaPolicyUUID      *string        `json:"quota_policy_uuid,omitempty"`
	CategoryUUID         *string        `json:"category_uuid,omitempty"`
	TestbedUUID          *string        `json:"testbed_uuid,omitempty"`
	AllocationUUID       *string        `json:"allocation_uuid,omitempty"`
	StartedAt            *string        `json:"started_at,omitempty"`
	CompletedAt          *string        `json:"completed_at,omitempty"`
	DurationMs           *int           `json:"duration_ms,omitempty"`
	DurationDisplay      *string        `json:"duration_display,omitempty"`
	Success              *bool          `json:"success,omitempty"`
	ErrorCode            *string        `json:"error_code,omitempty"`
	ErrorMessage         *string        `json:"error_message,omitempty"`
	ResultDetails        ResultDetails  `json:"result_details,omitempty"`
	RetryCount           int            `json:"retry_count"`
	MaxRetries           int            `json:"max_retries"`
	ParentTaskUUID       *string        `json:"parent_task_uuid,omitempty"`
	CreatedAt            string         `json:"created_at"`
	UpdatedAt            string         `json:"updated_at"`
}

// MarkRunning 标记任务为运行中
func (t *ResourceInstanceTask) MarkRunning() {
	now := time.Now()
	t.Status = TaskStatusRunning
	t.StartedAt = &now
	t.UpdatedAt = now
}

// MarkCompleted 标记任务为已完成
func (t *ResourceInstanceTask) MarkCompleted(success bool) {
	now := time.Now()
	t.Status = TaskStatusCompleted
	t.CompletedAt = &now
	t.Success = &success

	// 计算执行时长
	if t.StartedAt != nil {
		duration := now.Sub(*t.StartedAt)
		ms := int(duration.Milliseconds())
		t.DurationMs = &ms
	}
	t.UpdatedAt = now
}

// MarkFailed 标记任务为失败
func (t *ResourceInstanceTask) MarkFailed(errorCode, errorMessage string) {
	now := time.Now()
	t.Status = TaskStatusFailed
	t.CompletedAt = &now

	falseVal := false
	t.Success = &falseVal
	t.ErrorCode = &errorCode
	t.ErrorMessage = &errorMessage

	// 计算执行时长
	if t.StartedAt != nil {
		duration := now.Sub(*t.StartedAt)
		ms := int(duration.Milliseconds())
		t.DurationMs = &ms
	}
	t.UpdatedAt = now
}

// MarkCancelled 标记任务为已取消
func (t *ResourceInstanceTask) MarkCancelled(reason string) {
	now := time.Now()
	t.Status = TaskStatusCancelled
	t.CompletedAt = &now
	t.ErrorMessage = &reason
	t.UpdatedAt = now
}

// CanRetry 检查任务是否可以重试
func (t *ResourceInstanceTask) CanRetry() bool {
	return t.Status == TaskStatusFailed && t.RetryCount < t.MaxRetries
}

// IncrementRetry 增加重试计数
func (t *ResourceInstanceTask) IncrementRetry() {
	t.RetryCount++
	t.Status = TaskStatusPending
	t.StartedAt = nil
	t.CompletedAt = nil
	t.UpdatedAt = time.Now()
}

// SetResultDetails 设置结果详情
func (t *ResourceInstanceTask) SetResultDetails(details map[string]interface{}) {
	if t.ResultDetails == nil {
		t.ResultDetails = make(ResultDetails)
	}
	for k, v := range details {
		t.ResultDetails[k] = v
	}
}

// ToResponse 转换为 API 响应格式
func (t *ResourceInstanceTask) ToResponse() ResourceInstanceTaskResponse {
	resp := ResourceInstanceTaskResponse{
		ID:                t.ID,
		UUID:              t.UUID,
		ResourceInstanceUUID: t.ResourceInstanceUUID,
		TaskType:          t.TaskType,
		TaskTypeName:      t.TaskType.DisplayName(),
		Status:            t.Status,
		StatusName:        t.Status.DisplayName(),
		TriggerSource:     t.TriggerSource,
		TriggerSourceName: t.TriggerSource.DisplayName(),
		TriggerUser:       t.TriggerUser,
		QuotaPolicyUUID:   t.QuotaPolicyUUID,
		CategoryUUID:      t.CategoryUUID,
		TestbedUUID:       t.TestbedUUID,
		AllocationUUID:    t.AllocationUUID,
		DurationMs:        t.DurationMs,
		Success:           t.Success,
		ErrorCode:         t.ErrorCode,
		ErrorMessage:      t.ErrorMessage,
		ResultDetails:     t.ResultDetails,
		RetryCount:        t.RetryCount,
		MaxRetries:        t.MaxRetries,
		ParentTaskUUID:    t.ParentTaskUUID,
		CreatedAt:         t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         t.UpdatedAt.Format(time.RFC3339),
	}

	if t.StartedAt != nil {
		startedAt := t.StartedAt.Format(time.RFC3339)
		resp.StartedAt = &startedAt
	}

	if t.CompletedAt != nil {
		completedAt := t.CompletedAt.Format(time.RFC3339)
		resp.CompletedAt = &completedAt
	}

	if t.DurationMs != nil {
		display := formatDuration(*t.DurationMs)
		resp.DurationDisplay = &display
	}

	return resp
}

// formatDuration 格式化时长
func formatDuration(ms int) string {
	sec := ms / 1000
	min := sec / 60
	sec = sec % 60

	if min > 0 {
		return fmt.Sprintf("%d分%d秒", min, sec)
	}
	return fmt.Sprintf("%d秒", sec)
}

// NewResourceInstanceTask 创建新的资源实例任务
func NewResourceInstanceTask(resourceInstanceUUID string, taskType TaskType, triggerSource TriggerSource) *ResourceInstanceTask {
	now := time.Now()
	return &ResourceInstanceTask{
		UUID:                 uuid.New().String(),
		ResourceInstanceUUID: resourceInstanceUUID,
		TaskType:             taskType,
		Status:               TaskStatusPending,
		TriggerSource:        triggerSource,
		RetryCount:           0,
		MaxRetries:           3,
		ResultDetails:        make(ResultDetails),
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

// NewDeployTask 创建部署任务
func NewDeployTask(resourceInstanceUUID, categoryUUID string, triggerSource TriggerSource, triggerUser string) *ResourceInstanceTask {
	task := NewResourceInstanceTask(resourceInstanceUUID, TaskTypeDeploy, triggerSource)
	if categoryUUID != "" {
		task.CategoryUUID = &categoryUUID
	}
	if triggerUser != "" {
		task.TriggerUser = &triggerUser
	}
	return task
}

// NewRollbackTask 创建回滚任务
func NewRollbackTask(resourceInstanceUUID, testbedUUID, allocationUUID string, triggerSource TriggerSource) *ResourceInstanceTask {
	task := NewResourceInstanceTask(resourceInstanceUUID, TaskTypeRollback, triggerSource)
	if testbedUUID != "" {
		task.TestbedUUID = &testbedUUID
	}
	if allocationUUID != "" {
		task.AllocationUUID = &allocationUUID
	}
	return task
}

// NewHealthCheckTask 创建健康检查任务
func NewHealthCheckTask(resourceInstanceUUID string) *ResourceInstanceTask {
	return NewResourceInstanceTask(resourceInstanceUUID, TaskTypeHealthCheck, TriggerSourceManual)
}

// NewAutoReplenishDeployTask 创建自动补充部署任务
func NewAutoReplenishDeployTask(resourceInstanceUUID, quotaPolicyUUID, categoryUUID string) *ResourceInstanceTask {
	task := NewResourceInstanceTask(resourceInstanceUUID, TaskTypeDeploy, TriggerSourceAutoReplenish)
	task.QuotaPolicyUUID = &quotaPolicyUUID
	if categoryUUID != "" {
		task.CategoryUUID = &categoryUUID
	}
	return task
}
