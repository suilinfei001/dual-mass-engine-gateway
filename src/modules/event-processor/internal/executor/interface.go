package executor

import (
	"context"
)

type TaskStatus string

const (
	TaskStatusInProgress TaskStatus = "inProgress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusCanceled   TaskStatus = "canceled"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusNotStarted TaskStatus = "notStarted"
)

type TaskResult string

const (
	TaskResultSucceeded TaskResult = "succeeded"
	TaskResultFailed    TaskResult = "failed"
	TaskResultCanceled  TaskResult = "canceled"
	TaskResultTimeout   TaskResult = "timeout"
)

type RunResult struct {
	BuildID     int        `json:"build_id"`
	BuildNumber string     `json:"build_number"`
	Status      TaskStatus `json:"status"`
	WebURL      string     `json:"web_url"`
}

type StatusResult struct {
	BuildID     int        `json:"build_id"`
	BuildNumber string     `json:"build_number"`
	Status      TaskStatus `json:"status"`
	Result      TaskResult `json:"result,omitempty"`
	FinishTime  string     `json:"finish_time,omitempty"`
}

type LogResult struct {
	LogID     int    `json:"log_id"`
	Content   string `json:"content"`
	LineCount int    `json:"line_count"`
}

type TimelineRecord struct {
	Name   string     `json:"name"`
	Type   string     `json:"type"`
	State  string     `json:"state"`
	Result TaskResult `json:"result,omitempty"`
	LogID  *int       `json:"logId,omitempty"` // Pointer to distinguish 0 from nil
}

type TimelineResult struct {
	Records []TimelineRecord `json:"records"`
}

type TaskExecutor interface {
	Run(ctx context.Context, pipelineID int, params map[string]interface{}) (*RunResult, error)
	Cancel(ctx context.Context, buildID int) error
	GetStatus(ctx context.Context, buildID int) (*StatusResult, error)
	GetLogs(ctx context.Context, buildID int, logID int) (*LogResult, error)
	GetTimeline(ctx context.Context, buildID int) (*TimelineResult, error)
	GetLogList(ctx context.Context, buildID int) ([]LogResult, error)
}

type ExecutorConfig struct {
	Organization string
	Project      string
	PAT          string
	Branch       string
	VerifySSL    bool
}
