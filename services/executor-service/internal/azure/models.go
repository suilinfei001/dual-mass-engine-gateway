// Package azure provides Azure DevOps related models for executor service.
package azure

import "time"

// PipelineStatus Pipeline 状态
type PipelineStatus string

const (
	PipelineStatusPending   PipelineStatus = "pending"
	PipelineStatusQueued    PipelineStatus = "queued"
	PipelineStatusRunning   PipelineStatus = "running"
	PipelineStatusCompleted PipelineStatus = "completed"
	PipelineStatusFailed    PipelineStatus = "failed"
	PipelineStatusCanceled  PipelineStatus = "canceled"
)

// PipelineResult Pipeline 结果
type PipelineResult string

const (
	PipelineResultSucceeded PipelineResult = "succeeded"
	PipelineResultFailed    PipelineResult = "failed"
	PipelineResultCanceled  PipelineResult = "canceled"
)

// PipelineRunRequest Pipeline 执行请求
type PipelineRunRequest struct {
	Organization   string            `json:"organization"`
	Project        string            `json:"project"`
	PipelineID     int               `json:"pipeline_id"`
	SourceBranch   string            `json:"source_branch"`
	TargetBranch   string            `json:"target_branch,omitempty"`
	CommitSHA      string            `json:"commit_sha,omitempty"`
	Parameters     map[string]string `json:"parameters,omitempty"`
}

// PipelineRunResponse Pipeline 执行响应
type PipelineRunResponse struct {
	RunID       int64       `json:"run_id"`
	RunURL      string      `json:"run_url"`
	Status      PipelineStatus `json:"status"`
	QueuedAt    time.Time   `json:"queued_at"`
	StartedAt   *time.Time  `json:"started_at,omitempty"`
	CompletedAt *time.Time  `json:"completed_at,omitempty"`
	Result      PipelineResult `json:"result,omitempty"`
}

// PipelineStatusResponse Pipeline 状态响应
type PipelineStatusResponse struct {
	RunID       int64           `json:"run_id"`
	Status      PipelineStatus  `json:"status"`
	Result      PipelineResult  `json:"result,omitempty"`
	QueuedAt    time.Time       `json:"queued_at"`
	StartedAt   *time.Time      `json:"started_at,omitempty"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
	Finished    bool            `json:"finished"`
}

// PipelineLogResponse Pipeline 日志响应
type PipelineLogResponse struct {
	RunID      int64  `json:"run_id"`
	PhaseID    int64  `json:"phase_id"`
	LogID      int64  `json:"log_id"`
	LogContent string `json:"log_content,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// AzureAuthConfig Azure 认证配置
type AzureAuthConfig struct {
	OrganizationURL string `json:"organization_url"`
	PAT             string `json:"pat"` // Personal Access Token
}

// TaskExecutionRequest 任务执行请求
type TaskExecutionRequest struct {
	TaskUUID     string            `json:"task_uuid"`
	TaskType     string            `json:"task_type"`
	TestbedIP    string            `json:"testbed_ip"`
	TestbedSSHPort int             `json:"testbed_ssh_port"`
	TestbedSSHUser string          `json:"testbed_ssh_user"`
	TestbedSSHPassword string      `json:"testbed_ssh_password,omitempty"`
	ChartURL     string            `json:"chart_url"`
	Parameters   map[string]string `json:"parameters,omitempty"`
}

// TaskExecutionResponse 任务执行响应
type TaskExecutionResponse struct {
	ExecutionID string            `json:"execution_id"`
	RunID       int64             `json:"run_id"`
	RunURL      string            `json:"run_url"`
	Status      PipelineStatus    `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
}

// TaskExecutionStatus 任务执行状态
type TaskExecutionStatus struct {
	ExecutionID string           `json:"execution_id"`
	RunID       int64            `json:"run_id"`
	Status      PipelineStatus   `json:"status"`
	Result      PipelineResult   `json:"result,omitempty"`
	StartedAt   *time.Time       `json:"started_at,omitempty"`
	CompletedAt *time.Time       `json:"completed_at,omitempty"`
	Finished    bool             `json:"finished"`
	ErrorMessage string          `json:"error_message,omitempty"`
}

// TaskExecutionLog 任务执行日志
type TaskExecutionLog struct {
	ExecutionID string `json:"execution_id"`
	RunID       int64  `json:"run_id"`
	PhaseName   string `json:"phase_name"`
	LogContent  string `json:"log_content"`
	Timestamp   time.Time `json:"timestamp"`
}
