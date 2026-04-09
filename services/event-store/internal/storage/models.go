// Package storage provides data models for event store service.
package storage

import "time"

// EventType 事件类型
type EventType string

const (
	EventTypePullRequestOpened    EventType = "pull_request.opened"
	EventTypePullRequestSynchronized EventType = "pull_request.synchronize"
	EventTypePullRequestClosed      EventType = "pull_request.closed"
	EventTypePush                   EventType = "push"
	EventTypeRelease                EventType = "release"
)

// EventStatus 事件状态
type EventStatus string

const (
	EventStatusPending    EventStatus = "pending"
	EventStatusProcessing EventStatus = "processing"
	EventStatusCompleted  EventStatus = "completed"
	EventStatusFailed     EventStatus = "failed"
	EventStatusCancelled  EventStatus = "cancelled"
)

// Event 事件
type Event struct {
	ID          int64       `json:"id" db:"id"`
	UUID        string      `json:"uuid" db:"uuid"`
	EventType   EventType   `json:"event_type" db:"event_type"`
	Status      EventStatus `json:"status" db:"status"`
	Source      string      `json:"source" db:"source"`           // github, gitlab, etc.
	RepoID      int64       `json:"repo_id" db:"repo_id"`
	RepoName    string      `json:"repo_name" db:"repo_name"`
	RepoOwner   string      `json:"repo_owner" db:"repo_owner"`
	PRNumber    int         `json:"pr_number" db:"pr_number"`
	CommitSHA   string      `json:"commit_sha" db:"commit_sha"`
	Author      string      `json:"author" db:"author"`
	Payload     string      `json:"payload" db:"payload"` // JSON payload
	ReceivedAt  time.Time   `json:"received_at" db:"received_at"`
	ProcessedAt *time.Time  `json:"processed_at" db:"processed_at"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

// EventFilter 事件筛选条件
type EventFilter struct {
	EventType   EventType   `json:"event_type,omitempty"`
	Status      EventStatus `json:"status,omitempty"`
	Source      string      `json:"source,omitempty"`
	RepoName    string      `json:"repo_name,omitempty"`
	PRNumber    int         `json:"pr_number,omitempty"`
	Author      string      `json:"author,omitempty"`
	StartDate   *time.Time  `json:"start_date,omitempty"`
	EndDate     *time.Time  `json:"end_date,omitempty"`
	Limit       int         `json:"limit,omitempty"`
	Offset      int         `json:"offset,omitempty"`
}

// QualityCheck 质量检查
type QualityCheck struct {
	ID            int64          `json:"id" db:"id"`
	EventUUID     string         `json:"event_uuid" db:"event_uuid"`
	CheckType     string         `json:"check_type" db:"check_type"`
	CheckStatus   string         `json:"check_status" db:"check_status"`
	Result        string         `json:"result" db:"result"`         // pass, fail, skipped
	Score         float64        `json:"score" db:"score"`
	Details       string         `json:"details" db:"details"`       // JSON details
	StartedAt     *time.Time     `json:"started_at" db:"started_at"`
	CompletedAt   *time.Time     `json:"completed_at" db:"completed_at"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at" db:"updated_at"`
}

// QualityCheckStatus 质量检查状态
type QualityCheckStatus string

const (
	QualityCheckStatusPending   QualityCheckStatus = "pending"
	QualityCheckStatusRunning   QualityCheckStatus = "running"
	QualityCheckStatusPassed    QualityCheckStatus = "passed"
	QualityCheckStatusFailed    QualityCheckStatus = "failed"
	QualityCheckStatusSkipped   QualityCheckStatus = "skipped"
	QualityCheckStatusCancelled QualityCheckStatus = "cancelled"
)

// QualityCheckResult 质量检查结果
type QualityCheckResult string

const (
	QualityCheckResultPass    QualityCheckResult = "pass"
	QualityCheckResultFail    QualityCheckResult = "fail"
	QualityCheckResultSkip    QualityCheckResult = "skip"
)

// EventStatistics 事件统计
type EventStatistics struct {
	TotalEvents      int64 `json:"total_events"`
	PendingEvents    int64 `json:"pending_events"`
	ProcessingEvents int64 `json:"processing_events"`
	CompletedEvents  int64 `json:"completed_events"`
	FailedEvents     int64 `json:"failed_events"`
	CancelledEvents  int64 `json:"cancelled_events"`
}

// QualityCheckStatistics 质量检查统计
type QualityCheckStatistics struct {
	TotalChecks    int64 `json:"total_checks"`
	PendingChecks  int64 `json:"pending_checks"`
	RunningChecks  int64 `json:"running_checks"`
	PassedChecks   int64 `json:"passed_checks"`
	FailedChecks   int64 `json:"failed_checks"`
	SkippedChecks  int64 `json:"skipped_checks"`
	CancelledChecks int64 `json:"cancelled_checks"`
}
