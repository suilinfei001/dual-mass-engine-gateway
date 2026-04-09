package models

import "time"

// TaskStatus represents the status of a task.
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusRunning    TaskStatus = "running"
	TaskStatusPassed     TaskStatus = "passed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusSkipped    TaskStatus = "skipped"
	TaskStatusCancelled  TaskStatus = "cancelled"
	TaskStatusTimeout    TaskStatus = "timeout"
	TaskStatusNoResource TaskStatus = "no_resource"
)

// TaskType represents the type of task.
type TaskType string

const (
	TaskTypeBasicCI               TaskType = "basic_ci"
	TaskTypeBasicCIAll            TaskType = "basic_ci_all"
	TaskTypeAPITest               TaskType = "api_test"
	TaskTypeModuleE2E             TaskType = "module_e2e"
	TaskTypeAgentE2E              TaskType = "agent_e2e"
	TaskTypeAIAnalysis            TaskType = "ai_analysis"
	TaskTypeDeployment            TaskType = "deployment"
	TaskTypeDeploymentAll         TaskType = "deployment_all"
	TaskTypeSpecializedAPITest    TaskType = "specialized_tests_api_test"
	TaskTypeSpecializedModuleE2E  TaskType = "specialized_tests_module_e2e"
	TaskTypeSpecializedAgentE2E   TaskType = "specialized_tests_agent_e2e"
	TaskTypeSpecializedAIAnalysis TaskType = "specialized_tests_ai_e2e"
)

// Task represents a quality check task.
type Task struct {
	ID             int64      `json:"id" db:"id"`
	EventID        int64      `json:"event_id" db:"event_id"`
	EventType      EventType  `json:"event_type" db:"event_type"`
	TaskType       TaskType   `json:"task_type" db:"task_type"`
	Status         TaskStatus `json:"status" db:"status"`
	PRNumber       int        `json:"pr_number" db:"pr_number"`
	SourceBranch   string     `json:"source_branch" db:"source_branch"`
	TargetBranch   string     `json:"target_branch" db:"target_branch"`
	RepoURL        string     `json:"repo_url" db:"repo_url"`
	PipelineID     int64      `json:"pipeline_id" db:"pipeline_id"`
	PipelineURL    string     `json:"pipeline_url" db:"pipeline_url"`
	Analyzing      bool       `json:"analyzing" db:"analyzing"`
	TestbedUUID    string     `json:"testbed_uuid" db:"testbed_uuid"`
	TestbedIP      string     `json:"testbed_ip" db:"testbed_ip"`
	SSHUser        string     `json:"ssh_user" db:"ssh_user"`
	SSHPassword    string     `json:"ssh_password" db:"ssh_password"`
	ChartURL       string     `json:"chart_url" db:"chart_url"`
	AllocationUUID string     `json:"allocation_uuid" db:"allocation_uuid"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// TaskResult represents the result of a task execution.
type TaskResult struct {
	TaskID      int64              `json:"task_id" db:"task_id"`
	CheckType   QualityCheckType   `json:"check_type" db:"check_type"`
	CheckStatus QualityCheckStatus `json:"check_status" db:"check_status"`
	Result      CheckResult        `json:"result" db:"result"`
	LogPath     string             `json:"log_path" db:"log_path"`
	CompletedAt time.Time          `json:"completed_at" db:"completed_at"`
}

// IsFinalStatus returns true if the task status is a final status.
func (t *Task) IsFinalStatus() bool {
	return t.Status == TaskStatusPassed ||
		t.Status == TaskStatusFailed ||
		t.Status == TaskStatusSkipped ||
		t.Status == TaskStatusCancelled ||
		t.Status == TaskStatusTimeout ||
		t.Status == TaskStatusNoResource
}

// IsRunning returns true if the task is running.
func (t *Task) IsRunning() bool {
	return t.Status == TaskStatusRunning
}

// IsPending returns true if the task is pending.
func (t *Task) IsPending() bool {
	return t.Status == TaskStatusPending
}

// CanStart returns true if the task can be started.
func (t *Task) CanStart() bool {
	return t.Status == TaskStatusPending
}

// CanRetry returns true if the task can be retried.
func (t *Task) CanRetry() bool {
	return t.Status == TaskStatusFailed || t.Status == TaskStatusTimeout
}

// ParseTaskStatus parses a string into a TaskStatus.
func ParseTaskStatus(s string) (TaskStatus, error) {
	switch s {
	case "pending":
		return TaskStatusPending, nil
	case "running":
		return TaskStatusRunning, nil
	case "passed":
		return TaskStatusPassed, nil
	case "failed":
		return TaskStatusFailed, nil
	case "skipped":
		return TaskStatusSkipped, nil
	case "cancelled":
		return TaskStatusCancelled, nil
	case "timeout":
		return TaskStatusTimeout, nil
	case "no_resource":
		return TaskStatusNoResource, nil
	default:
		return "", ErrInvalidTaskStatus
	}
}

// ParseTaskType parses a string into a TaskType.
func ParseTaskType(s string) (TaskType, error) {
	switch s {
	case "basic_ci":
		return TaskTypeBasicCI, nil
	case "basic_ci_all":
		return TaskTypeBasicCIAll, nil
	case "api_test":
		return TaskTypeAPITest, nil
	case "module_e2e":
		return TaskTypeModuleE2E, nil
	case "agent_e2e":
		return TaskTypeAgentE2E, nil
	case "ai_analysis":
		return TaskTypeAIAnalysis, nil
	case "deployment":
		return TaskTypeDeployment, nil
	case "deployment_all":
		return TaskTypeDeploymentAll, nil
	case "specialized_tests_api_test":
		return TaskTypeSpecializedAPITest, nil
	case "specialized_tests_module_e2e":
		return TaskTypeSpecializedModuleE2E, nil
	case "specialized_tests_agent_e2e":
		return TaskTypeSpecializedAgentE2E, nil
	case "specialized_tests_ai_e2e":
		return TaskTypeSpecializedAIAnalysis, nil
	default:
		return "", ErrInvalidTaskStatus
	}
}

// DisplayName returns a human-readable display name for the task type.
func (t TaskType) DisplayName() string {
	switch t {
	case TaskTypeBasicCI:
		return "Basic CI"
	case TaskTypeBasicCIAll:
		return "Basic CI All"
	case TaskTypeAPITest:
		return "API Test"
	case TaskTypeModuleE2E:
		return "Module E2E"
	case TaskTypeAgentE2E:
		return "Agent E2E"
	case TaskTypeAIAnalysis:
		return "AI Analysis"
	case TaskTypeDeployment:
		return "Deployment"
	case TaskTypeDeploymentAll:
		return "Deployment All"
	case TaskTypeSpecializedAPITest:
		return "Specialized API Test"
	case TaskTypeSpecializedModuleE2E:
		return "Specialized Module E2E"
	case TaskTypeSpecializedAgentE2E:
		return "Specialized Agent E2E"
	case TaskTypeSpecializedAIAnalysis:
		return "Specialized AI E2E"
	default:
		return string(t)
	}
}

// ToQualityCheckType converts TaskType to QualityCheckType.
func (t TaskType) ToQualityCheckType() QualityCheckType {
	return QualityCheckType(t)
}

// IsAnalysisTask returns true if this is an AI analysis task.
func (t *Task) IsAnalysisTask() bool {
	return t.TaskType == TaskTypeBasicCIAll || t.TaskType == TaskTypeSpecializedAIAnalysis
}
