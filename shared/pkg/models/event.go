// Package models provides shared data models for all microservices.
package models

import (
	"time"
)

// EventStatus represents the status of an event.
type EventStatus string

const (
	EventStatusPending    EventStatus = "pending"
	EventStatusProcessing EventStatus = "processing"
	EventStatusPassed     EventStatus = "passed"
	EventStatusFailed     EventStatus = "failed"
	EventStatusCancelled  EventStatus = "cancelled"
)

// EventType represents the type of GitHub event.
type EventType string

const (
	EventTypePROpened      EventType = "pull_request.opened"
	EventTypePRSynchronize EventType = "pull_request.synchronize"
	EventTypePRClosed      EventType = "pull_request.closed"
	EventTypePRMerged      EventType = "pull_request.merged"
)

// Event represents a GitHub webhook event.
type Event struct {
	ID           int64       `json:"id" db:"id"`
	EventID      string      `json:"event_id" db:"event_id"`
	EventType    EventType   `json:"event_type" db:"event_type"`
	PRNumber     int         `json:"pr_number" db:"pr_number"`
	PRTitle      string      `json:"pr_title" db:"pr_title"`
	SourceBranch string      `json:"source_branch" db:"source_branch"`
	TargetBranch string      `json:"target_branch" db:"target_branch"`
	RepoURL      string      `json:"repo_url" db:"repo_url"`
	Sender       string      `json:"sender" db:"sender"`
	EventStatus  EventStatus `json:"event_status" db:"event_status"`
	CreatedAt    time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at" db:"updated_at"`
}

// IsFinalStatus returns true if the event status is a final status.
func (e *Event) IsFinalStatus() bool {
	return e.EventStatus == EventStatusPassed ||
		e.EventStatus == EventStatusFailed ||
		e.EventStatus == EventStatusCancelled
}

// QualityCheckType represents the type of quality check.
type QualityCheckType string

const (
	QualityCheckTypeBasicCI               QualityCheckType = "basic_ci"
	QualityCheckTypeBasicCIAll            QualityCheckType = "basic_ci_all"
	QualityCheckTypeAPITest               QualityCheckType = "api_test"
	QualityCheckTypeModuleE2E             QualityCheckType = "module_e2e"
	QualityCheckTypeAgentE2E              QualityCheckType = "agent_e2e"
	QualityCheckTypeAIAnalysis            QualityCheckType = "ai_analysis"
	QualityCheckTypeDeployment            QualityCheckType = "deployment"
	QualityCheckTypeDeploymentAll         QualityCheckType = "deployment_all"
	QualityCheckTypeSpecializedAPITest    QualityCheckType = "specialized_tests_api_test"
	QualityCheckTypeSpecializedModuleE2E  QualityCheckType = "specialized_tests_module_e2e"
	QualityCheckTypeSpecializedAgentE2E   QualityCheckType = "specialized_tests_agent_e2e"
	QualityCheckTypeSpecializedAIAnalysis QualityCheckType = "specialized_tests_ai_e2e"
)

// QualityCheckStatus represents the status of a quality check.
type QualityCheckStatus string

const (
	QualityCheckStatusPending   QualityCheckStatus = "pending"
	QualityCheckStatusRunning   QualityCheckStatus = "running"
	QualityCheckStatusPassed    QualityCheckStatus = "passed"
	QualityCheckStatusFailed    QualityCheckStatus = "failed"
	QualityCheckStatusSkipped   QualityCheckStatus = "skipped"
	QualityCheckStatusCancelled QualityCheckStatus = "cancelled"
)

// QualityCheck represents a quality check for an event.
type QualityCheck struct {
	ID          int64              `json:"id" db:"id"`
	EventID     int64              `json:"event_id" db:"event_id"`
	CheckType   QualityCheckType   `json:"check_type" db:"check_type"`
	CheckStatus QualityCheckStatus `json:"check_status" db:"check_status"`
	StageOrder  int                `json:"stage_order" db:"stage_order"`
	Result      CheckResult        `json:"result" db:"result"`
	CreatedAt   time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" db:"updated_at"`
}

// CheckResult represents the result of a quality check.
type CheckResult map[string]interface{}

// IsPassed returns true if the check passed.
func (c *QualityCheck) IsPassed() bool {
	return c.CheckStatus == QualityCheckStatusPassed
}

// IsFailed returns true if the check failed.
func (c *QualityCheck) IsFailed() bool {
	return c.CheckStatus == QualityCheckStatusFailed
}

// IsPending returns true if the check is pending.
func (c *QualityCheck) IsPending() bool {
	return c.CheckStatus == QualityCheckStatusPending
}

// IsRunning returns true if the check is running.
func (c *QualityCheck) IsRunning() bool {
	return c.CheckStatus == QualityCheckStatusRunning
}

// DisplayName returns a human-readable display name for the check type.
func (t QualityCheckType) DisplayName() string {
	switch t {
	case QualityCheckTypeBasicCI:
		return "Basic CI"
	case QualityCheckTypeBasicCIAll:
		return "Basic CI All"
	case QualityCheckTypeAPITest:
		return "API Test"
	case QualityCheckTypeModuleE2E:
		return "Module E2E"
	case QualityCheckTypeAgentE2E:
		return "Agent E2E"
	case QualityCheckTypeAIAnalysis:
		return "AI Analysis"
	case QualityCheckTypeDeployment:
		return "Deployment"
	case QualityCheckTypeDeploymentAll:
		return "Deployment All"
	case QualityCheckTypeSpecializedAPITest:
		return "Specialized API Test"
	case QualityCheckTypeSpecializedModuleE2E:
		return "Specialized Module E2E"
	case QualityCheckTypeSpecializedAgentE2E:
		return "Specialized Agent E2E"
	case QualityCheckTypeSpecializedAIAnalysis:
		return "Specialized AI E2E"
	default:
		return string(t)
	}
}

// ParseQualityCheckType parses a string into a QualityCheckType.
func ParseQualityCheckType(s string) (QualityCheckType, error) {
	switch s {
	case "basic_ci":
		return QualityCheckTypeBasicCI, nil
	case "basic_ci_all":
		return QualityCheckTypeBasicCIAll, nil
	case "api_test":
		return QualityCheckTypeAPITest, nil
	case "module_e2e":
		return QualityCheckTypeModuleE2E, nil
	case "agent_e2e":
		return QualityCheckTypeAgentE2E, nil
	case "ai_analysis":
		return QualityCheckTypeAIAnalysis, nil
	case "deployment":
		return QualityCheckTypeDeployment, nil
	case "deployment_all":
		return QualityCheckTypeDeploymentAll, nil
	case "specialized_tests_api_test":
		return QualityCheckTypeSpecializedAPITest, nil
	case "specialized_tests_module_e2e":
		return QualityCheckTypeSpecializedModuleE2E, nil
	case "specialized_tests_agent_e2e":
		return QualityCheckTypeSpecializedAgentE2E, nil
	case "specialized_tests_ai_e2e":
		return QualityCheckTypeSpecializedAIAnalysis, nil
	default:
		return "", ErrInvalidCheckType
	}
}

// ParseQualityCheckStatus parses a string into a QualityCheckStatus.
func ParseQualityCheckStatus(s string) (QualityCheckStatus, error) {
	switch s {
	case "pending":
		return QualityCheckStatusPending, nil
	case "running":
		return QualityCheckStatusRunning, nil
	case "passed":
		return QualityCheckStatusPassed, nil
	case "failed":
		return QualityCheckStatusFailed, nil
	case "skipped":
		return QualityCheckStatusSkipped, nil
	case "cancelled":
		return QualityCheckStatusCancelled, nil
	default:
		return "", ErrInvalidCheckStatus
	}
}

// ParseEventStatus parses a string into an EventStatus.
func ParseEventStatus(s string) (EventStatus, error) {
	switch s {
	case "pending":
		return EventStatusPending, nil
	case "processing":
		return EventStatusProcessing, nil
	case "passed":
		return EventStatusPassed, nil
	case "failed":
		return EventStatusFailed, nil
	case "cancelled":
		return EventStatusCancelled, nil
	default:
		return "", ErrInvalidEventStatus
	}
}
