package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// LocalTime 本地时间类型（兼容数据库 NULL）
type LocalTime struct {
	Time time.Time
}

// Scan 实现 database/sql 的 Scanner 接口
func (lt *LocalTime) Scan(src interface{}) error {
	if src == nil {
		lt.Time = time.Time{}
		return nil
	}
	switch v := src.(type) {
	case time.Time:
		lt.Time = v
	case []byte:
		lt.Time, _ = time.Parse(time.RFC3339, string(v))
	case string:
		lt.Time, _ = time.Parse(time.RFC3339, v)
	default:
		lt.Time = time.Now()
	}
	return nil
}

// MarshalJSON 自定义 JSON 序列化
func (lt LocalTime) MarshalJSON() ([]byte, error) {
	if lt.Time.IsZero() {
		return []byte("null"), nil
	}
	return lt.Time.MarshalJSON()
}

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusRunning    TaskStatus = "running"
	TaskStatusPassed     TaskStatus = "passed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusTimeout    TaskStatus = "timeout"
	TaskStatusCancelled  TaskStatus = "cancelled"
	TaskStatusSkipped    TaskStatus = "skipped"
	TaskStatusNoResource TaskStatus = "no_resource"
)

// CheckResult 检查结果
type CheckResult struct {
	CheckType string      `json:"check_type"`
	Result    string      `json:"result"` // pass/fail/timeout/cancelled/skipped/running
	Extra     interface{} `json:"extra"`
}

// TaskResult 任务结果（单个检查的结果）
type TaskResult struct {
	CheckType string                 `json:"check_type"`
	Result    string                 `json:"result"`
	Output    string                 `json:"output,omitempty"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

type TaskResultResponse struct {
	CheckType string                 `json:"check_type"`
	Result    string                 `json:"result"`
	Output    string                 `json:"output,omitempty"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

type TaskResponse struct {
	ID           int                  `json:"id"`
	TaskID       string               `json:"task_id"`
	TaskName     string               `json:"task_name"`
	EventID      int                  `json:"event_id"`
	Stage        string               `json:"stage"`
	StageOrder   int                  `json:"stage_order"`
	ExecuteOrder int                  `json:"execute_order"`
	ResourceID   int                  `json:"resource_id,omitempty"`
	RequestURL   string               `json:"request_url"`
	BuildID      int                  `json:"build_id,omitempty"`
	LogFilePath  string               `json:"log_file_path,omitempty"`
	Status       string               `json:"status"`
	StartTime    *string              `json:"start_time,omitempty"`
	EndTime      *string              `json:"end_time,omitempty"`
	ErrorMessage string               `json:"error_message,omitempty"`
	Results      []TaskResultResponse `json:"results,omitempty"`
	Analyzing    bool                 `json:"analyzing"` // AI分析是否正在进行
	CreatedAt    string               `json:"created_at"`
	UpdatedAt    string               `json:"updated_at"`
}

// Task 任务定义
type Task struct {
	ID           int          `json:"id"`
	TaskID       string       `json:"task_id"`
	TaskName     string       `json:"task_name"`
	EventID      int          `json:"event_id"`
	CheckType    string       `json:"check_type,omitempty"` // 仅用于单一检查的任务
	Stage        string       `json:"stage"`
	StageOrder   int          `json:"stage_order"`
	CheckOrder   int          `json:"check_order,omitempty"`
	ExecuteOrder int          `json:"execute_order"`
	ResourceID   int          `json:"resource_id,omitempty"` // AI匹配的资源ID
	RequestURL   string       `json:"request_url"`
	BuildID      int          `json:"build_id,omitempty"` // Azure DevOps build ID
	Status       TaskStatus   `json:"status"`
	StartTime    *LocalTime   `json:"start_time"`
	EndTime      *LocalTime   `json:"end_time"`
	ErrorMessage *string      `json:"error_message,omitempty"`
	Results      []TaskResult `json:"results,omitempty"`
	LogFilePath  string       `json:"log_file_path,omitempty"`
	Analyzing    bool         `json:"analyzing"` // AI分析是否正在进行
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// IsCompleted 检查任务是否已完成（成功或失败）
func (t *Task) IsCompleted() bool {
	return t.Status == TaskStatusPassed || t.Status == TaskStatusFailed ||
		t.Status == TaskStatusTimeout || t.Status == TaskStatusCancelled ||
		t.Status == TaskStatusSkipped || t.Status == TaskStatusNoResource
}

// IsRunning 检查任务是否正在运行
func (t *Task) IsRunning() bool {
	return t.Status == TaskStatusRunning
}

// MarkRunning 标记任务为运行状态
func (t *Task) MarkRunning() {
	now := time.Now()
	t.Status = TaskStatusRunning
	t.StartTime = &LocalTime{Time: now}
	t.UpdatedAt = now
}

// MarkPassed 标记任务为通过
func (t *Task) MarkPassed(results []TaskResult) {
	now := time.Now()
	t.Status = TaskStatusPassed
	t.EndTime = &LocalTime{Time: now}
	t.Results = results
	t.UpdatedAt = now
}

// MarkFailed 标记任务为失败
func (t *Task) MarkFailed(reason string) {
	now := time.Now()
	t.Status = TaskStatusFailed
	t.EndTime = &LocalTime{Time: now}
	t.ErrorMessage = &reason
	t.UpdatedAt = now
}

// MarkCancelled 标记任务为取消
func (t *Task) MarkCancelled(reason string) {
	now := time.Now()
	t.Status = TaskStatusCancelled
	t.EndTime = &LocalTime{Time: now}
	t.ErrorMessage = &reason
	t.UpdatedAt = now
}

// MarkSkipped 标记任务为跳过
func (t *Task) MarkSkipped(reason string) {
	now := time.Now()
	t.Status = TaskStatusSkipped
	t.EndTime = &LocalTime{Time: now}
	t.ErrorMessage = &reason
	t.UpdatedAt = now
}

// MarkNoResource 标记任务为无法执行（无匹配资源）
func (t *Task) MarkNoResource(reason string) {
	now := time.Now()
	t.Status = TaskStatusNoResource
	t.EndTime = &LocalTime{Time: now}
	t.ErrorMessage = &reason
	t.UpdatedAt = now
}

// MarkTimeout 标记任务为超时
func (t *Task) MarkTimeout(reason string) {
	now := time.Now()
	t.Status = TaskStatusTimeout
	t.EndTime = &LocalTime{Time: now}
	t.ErrorMessage = &reason
	t.UpdatedAt = now
}

// GenerateTaskID 生成任务唯一ID
func (t *Task) GenerateTaskID() string {
	t.TaskID = uuid.New().String()
	return t.TaskID
}

// GetResultsJSON 获取结果的JSON字符串
func (t *Task) GetResultsJSON() string {
	if len(t.Results) == 0 {
		return ""
	}
	data, err := json.Marshal(t.Results)
	if err != nil {
		return ""
	}
	return string(data)
}

// SetResultsFromJSON 从JSON字符串设置结果
func (t *Task) SetResultsFromJSON(jsonStr string) error {
	if jsonStr == "" {
		return nil
	}
	return json.Unmarshal([]byte(jsonStr), &t.Results)
}

// IsSuccessful 检查任务是否成功完成
func (t *Task) IsSuccessful() bool {
	return t.Status == TaskStatusPassed || t.Status == TaskStatusSkipped
}

// IsBlocking 检查任务是否阻塞后续流程（失败、超时、无资源等）
func (t *Task) IsBlocking() bool {
	return t.Status == TaskStatusFailed || t.Status == TaskStatusTimeout ||
		t.Status == TaskStatusNoResource || t.Status == TaskStatusCancelled
}

// NewTask 创建新任务
func NewTask(eventID int, taskName, checkType, stage string, stageOrder, checkOrder, executeOrder int, requestURL string) *Task {
	now := time.Now()
	task := &Task{
		EventID:      eventID,
		TaskName:     taskName,
		CheckType:    checkType,
		Stage:        stage,
		StageOrder:   stageOrder,
		CheckOrder:   checkOrder,
		ExecuteOrder: executeOrder,
		RequestURL:   requestURL,
		Status:       TaskStatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	task.GenerateTaskID()
	return task
}

// NewBasicCITask 创建基础 CI 任务（包含多个检查）
func NewBasicCITask(eventID int, executeOrder int, requestURL string) *Task {
	now := time.Now()
	task := &Task{
		EventID:      eventID,
		TaskName:     "basic_ci_all",
		Stage:        "basic_ci",
		StageOrder:   1,
		ExecuteOrder: executeOrder,
		RequestURL:   requestURL,
		Status:       TaskStatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	task.GenerateTaskID()
	return task
}

// TaskDefinition 任务定义（用于创建任务模板）
type TaskDefinition struct {
	TaskName     string
	CheckType    string
	Stage        string
	StageOrder   int
	CheckOrder   int
	ExecuteOrder int
	RequestURL   string
}

// TaskDefinitions 按执行顺序定义的任务模板
var TaskDefinitions = []TaskDefinition{
	{
		TaskName:     "basic_ci_all",
		Stage:        "basic_ci",
		StageOrder:   1,
		ExecuteOrder: 1,
	},
	{
		TaskName:     "deployment_deployment",
		CheckType:    "deployment",
		Stage:        "deployment",
		StageOrder:   2,
		CheckOrder:   1,
		ExecuteOrder: 2,
	},
	{
		TaskName:     "specialized_tests_api_test",
		CheckType:    "specialized_tests_api_test",
		Stage:        "specialized_tests",
		StageOrder:   3,
		CheckOrder:   1,
		ExecuteOrder: 3,
	},
	{
		TaskName:     "specialized_tests_module_e2e",
		CheckType:    "specialized_tests_module_e2e",
		Stage:        "specialized_tests",
		StageOrder:   3,
		CheckOrder:   2,
		ExecuteOrder: 4,
	},
	{
		TaskName:     "specialized_tests_agent_e2e",
		CheckType:    "specialized_tests_agent_e2e",
		Stage:        "specialized_tests",
		StageOrder:   3,
		CheckOrder:   3,
		ExecuteOrder: 5,
	},
	{
		TaskName:     "specialized_tests_ai_e2e",
		CheckType:    "specialized_tests_ai_e2e",
		Stage:        "specialized_tests",
		StageOrder:   3,
		CheckOrder:   4,
		ExecuteOrder: 6,
	},
}

// GetNextExecuteOrder 获取下一个执行顺序
func GetNextExecuteOrder(currentOrder int) int {
	for _, def := range TaskDefinitions {
		if def.ExecuteOrder > currentOrder {
			return def.ExecuteOrder
		}
	}
	return -1
}

// GetTaskDefinitionByOrder 根据执行顺序获取任务定义
func GetTaskDefinitionByOrder(order int) *TaskDefinition {
	for i := range TaskDefinitions {
		if TaskDefinitions[i].ExecuteOrder == order {
			return &TaskDefinitions[i]
		}
	}
	return nil
}

// MaxExecuteOrder 最大执行顺序
const MaxExecuteOrder = 6
