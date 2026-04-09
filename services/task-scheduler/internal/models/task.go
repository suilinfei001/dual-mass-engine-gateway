package models

import (
	"encoding/json"
	"time"

	"github.com/quality-gateway/shared/pkg/models"
)

// Task 任务定义（扩展共享模型）
type Task struct {
	ID             int                  `json:"id" db:"id"`
	TaskID         string               `json:"task_id" db:"task_id"`
	TaskName       string               `json:"task_name" db:"task_name"`
	EventID        int                  `json:"event_id" db:"event_id"`
	CheckType      string               `json:"check_type,omitempty" db:"check_type"`
	Stage          string               `json:"stage" db:"stage"`
	StageOrder     int                  `json:"stage_order" db:"stage_order"`
	CheckOrder     int                  `json:"check_order,omitempty" db:"check_order"`
	ExecuteOrder   int                  `json:"execute_order" db:"execute_order"`
	ResourceID     int                  `json:"resource_id,omitempty" db:"resource_id"`
	RequestURL     string               `json:"request_url" db:"request_url"`
	BuildID        int                  `json:"build_id,omitempty" db:"build_id"`
	Status         models.TaskStatus    `json:"status" db:"status"`
	StartTime      *time.Time           `json:"start_time,omitempty" db:"start_time"`
	EndTime        *time.Time           `json:"end_time,omitempty" db:"end_time"`
	ErrorMessage   *string              `json:"error_message,omitempty" db:"error_message"`
	Results        []TaskResult  `json:"results,omitempty" db:"results"`
	LogFilePath    string               `json:"log_file_path,omitempty" db:"log_file_path"`
	Analyzing      bool                 `json:"analyzing" db:"analyzing"`
	TestbedUUID    string               `json:"testbed_uuid,omitempty" db:"testbed_uuid"`
	TestbedIP      string               `json:"testbed_ip,omitempty" db:"testbed_ip"`
	SSHUser        string               `json:"ssh_user,omitempty" db:"ssh_user"`
	SSHPassword    string               `json:"ssh_password,omitempty" db:"ssh_password"`
	ChartURL       string               `json:"chart_url,omitempty" db:"chart_url"`
	AllocationUUID string               `json:"allocation_uuid,omitempty" db:"allocation_uuid"`
	CreatedAt      time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at" db:"updated_at"`
}

// TaskResult 任务执行结果
type TaskResult struct {
	ID        int64     `json:"id" db:"id"`
	TaskID    int64     `json:"task_id" db:"task_id"`
	CheckType string    `json:"check_type" db:"check_type"`
	Result    string    `json:"result" db:"result"`
	Output    string    `json:"output,omitempty" db:"output"`
	Extra     string    `json:"extra,omitempty" db:"extra"` // JSON string
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ToTaskResponse 转换为响应格式
func (t *Task) ToTaskResponse() map[string]interface{} {
	results := make([]map[string]interface{}, len(t.Results))
	for i, r := range t.Results {
		var extra map[string]interface{}
		if r.Extra != "" {
			json.Unmarshal([]byte(r.Extra), &extra)
		}
		results[i] = map[string]interface{}{
			"check_type": r.CheckType,
			"result":     r.Result,
			"output":     r.Output,
			"extra":      extra,
		}
	}

	return map[string]interface{}{
		"id":              t.ID,
		"task_id":         t.TaskID,
		"task_name":       t.TaskName,
		"event_id":        t.EventID,
		"check_type":      t.CheckType,
		"stage":           t.Stage,
		"stage_order":     t.StageOrder,
		"execute_order":   t.ExecuteOrder,
		"resource_id":     t.ResourceID,
		"request_url":     t.RequestURL,
		"build_id":        t.BuildID,
		"status":          string(t.Status),
		"start_time":      formatTime(t.StartTime),
		"end_time":        formatTime(t.EndTime),
		"error_message":   formatString(t.ErrorMessage),
		"results":         results,
		"log_file_path":   t.LogFilePath,
		"analyzing":       t.Analyzing,
		"testbed_uuid":    t.TestbedUUID,
		"testbed_ip":      t.TestbedIP,
		"chart_url":       t.ChartURL,
		"allocation_uuid": t.AllocationUUID,
		"created_at":      t.CreatedAt.Format(time.RFC3339),
		"updated_at":      t.UpdatedAt.Format(time.RFC3339),
	}
}

func formatTime(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}

func formatString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// IsCompleted 检查任务是否已完成
func (t *Task) IsCompleted() bool {
	return t.Status == models.TaskStatusPassed ||
		t.Status == models.TaskStatusFailed ||
		t.Status == models.TaskStatusTimeout ||
		t.Status == models.TaskStatusCancelled ||
		t.Status == models.TaskStatusSkipped ||
		t.Status == models.TaskStatusNoResource
}

// IsRunning 检查任务是否正在运行
func (t *Task) IsRunning() bool {
	return t.Status == models.TaskStatusRunning
}

// IsPending 检查任务是否待执行
func (t *Task) IsPending() bool {
	return t.Status == models.TaskStatusPending
}

// MarkRunning 标记任务为运行状态
func (t *Task) MarkRunning() {
	now := time.Now()
	t.Status = models.TaskStatusRunning
	t.StartTime = &now
	t.UpdatedAt = now
}

// MarkPassed 标记任务为通过
func (t *Task) MarkPassed() {
	now := time.Now()
	t.Status = models.TaskStatusPassed
	t.EndTime = &now
	t.UpdatedAt = now
}

// MarkFailed 标记任务为失败
func (t *Task) MarkFailed(reason string) {
	now := time.Now()
	t.Status = models.TaskStatusFailed
	t.EndTime = &now
	t.ErrorMessage = &reason
	t.UpdatedAt = now
}

// MarkCancelled 标记任务为取消
func (t *Task) MarkCancelled(reason string) {
	now := time.Now()
	t.Status = models.TaskStatusCancelled
	t.EndTime = &now
	t.ErrorMessage = &reason
	t.UpdatedAt = now
}

// MarkSkipped 标记任务为跳过
func (t *Task) MarkSkipped(reason string) {
	now := time.Now()
	t.Status = models.TaskStatusSkipped
	t.EndTime = &now
	t.ErrorMessage = &reason
	t.UpdatedAt = now
}

// MarkTimeout 标记任务为超时
func (t *Task) MarkTimeout(reason string) {
	now := time.Now()
	t.Status = models.TaskStatusTimeout
	t.EndTime = &now
	t.ErrorMessage = &reason
	t.UpdatedAt = now
}

// NewTask 创建新任务
func NewTask(eventID int, taskName, checkType, stage string, stageOrder, checkOrder, executeOrder int, requestURL string) *Task {
	now := time.Now()
	return &Task{
		EventID:      eventID,
		TaskName:     taskName,
		CheckType:    checkType,
		Stage:        stage,
		StageOrder:   stageOrder,
		CheckOrder:   checkOrder,
		ExecuteOrder: executeOrder,
		RequestURL:   requestURL,
		Status:       models.TaskStatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// StartTaskRequest 启动任务请求
type StartTaskRequest struct {
	TaskID int `json:"task_id"`
}

// CompleteTaskRequest 完成任务请求
type CompleteTaskRequest struct {
	TaskID  int                    `json:"task_id"`
	Results []TaskResult           `json:"results,omitempty"`
}

// FailTaskRequest 标记任务失败请求
type FailTaskRequest struct {
	TaskID int    `json:"task_id"`
	Reason string `json:"reason"`
}

// CancelTaskRequest 取消任务请求
type CancelTaskRequest struct {
	TaskID int    `json:"task_id"`
	Reason string `json:"reason"`
}

// CancelEventTasksRequest 取消事件的所有任务请求
type CancelEventTasksRequest struct {
	EventID int    `json:"event_id"`
	Reason  string `json:"reason"`
}
