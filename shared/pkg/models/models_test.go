package models_test

import (
	"testing"

	"github.com/quality-gateway/shared/pkg/models"
)

func TestEventStatus(t *testing.T) {
	t.Run("IsFinalStatus", func(t *testing.T) {
		tests := []struct {
			status   models.EventStatus
			expected bool
		}{
			{models.EventStatusPassed, true},
			{models.EventStatusFailed, true},
			{models.EventStatusCancelled, true},
			{models.EventStatusPending, false},
			{models.EventStatusProcessing, false},
		}

		for _, tt := range tests {
			e := &models.Event{EventStatus: tt.status}
			if e.IsFinalStatus() != tt.expected {
				t.Errorf("expected %v for status %s, got %v", tt.expected, tt.status, e.IsFinalStatus())
			}
		}
	})
}

func TestParseEventStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected models.EventStatus
		hasError bool
	}{
		{"pending", models.EventStatusPending, false},
		{"processing", models.EventStatusProcessing, false},
		{"passed", models.EventStatusPassed, false},
		{"failed", models.EventStatusFailed, false},
		{"cancelled", models.EventStatusCancelled, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		result, err := models.ParseEventStatus(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("expected error for input %s", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error for input %s: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		}
	}
}

func TestParseQualityCheckType(t *testing.T) {
	tests := []struct {
		input    string
		expected models.QualityCheckType
		hasError bool
	}{
		{"basic_ci", models.QualityCheckTypeBasicCI, false},
		{"basic_ci_all", models.QualityCheckTypeBasicCIAll, false},
		{"api_test", models.QualityCheckTypeAPITest, false},
		{"module_e2e", models.QualityCheckTypeModuleE2E, false},
		{"agent_e2e", models.QualityCheckTypeAgentE2E, false},
		{"ai_analysis", models.QualityCheckTypeAIAnalysis, false},
		{"deployment", models.QualityCheckTypeDeployment, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		result, err := models.ParseQualityCheckType(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("expected error for input %s", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error for input %s: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		}
	}
}

func TestQualityCheckDisplayName(t *testing.T) {
	tests := []struct {
		checkType models.QualityCheckType
		expected  string
	}{
		{models.QualityCheckTypeBasicCI, "Basic CI"},
		{models.QualityCheckTypeBasicCIAll, "Basic CI All"},
		{models.QualityCheckTypeAPITest, "API Test"},
		{models.QualityCheckTypeAIAnalysis, "AI Analysis"},
	}

	for _, tt := range tests {
		if tt.checkType.DisplayName() != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.checkType.DisplayName())
		}
	}
}

func TestTaskStatus(t *testing.T) {
	t.Run("IsFinalStatus", func(t *testing.T) {
		tests := []struct {
			status   models.TaskStatus
			expected bool
		}{
			{models.TaskStatusPassed, true},
			{models.TaskStatusFailed, true},
			{models.TaskStatusSkipped, true},
			{models.TaskStatusCancelled, true},
			{models.TaskStatusTimeout, true},
			{models.TaskStatusNoResource, true},
			{models.TaskStatusPending, false},
			{models.TaskStatusRunning, false},
		}

		for _, tt := range tests {
			task := &models.Task{Status: tt.status}
			if task.IsFinalStatus() != tt.expected {
				t.Errorf("expected %v for status %s, got %v", tt.expected, tt.status, task.IsFinalStatus())
			}
		}
	})

	t.Run("CanStart", func(t *testing.T) {
		task := &models.Task{Status: models.TaskStatusPending}
		if !task.CanStart() {
			t.Error("expected pending task to be startable")
		}

		task.Status = models.TaskStatusRunning
		if task.CanStart() {
			t.Error("expected running task to not be startable")
		}
	})

	t.Run("CanRetry", func(t *testing.T) {
		task := &models.Task{Status: models.TaskStatusFailed}
		if !task.CanRetry() {
			t.Error("expected failed task to be retryable")
		}

		task.Status = models.TaskStatusTimeout
		if !task.CanRetry() {
			t.Error("expected timeout task to be retryable")
		}

		task.Status = models.TaskStatusPending
		if task.CanRetry() {
			t.Error("expected pending task to not be retryable")
		}
	})
}

func TestParseTaskStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected models.TaskStatus
		hasError bool
	}{
		{"pending", models.TaskStatusPending, false},
		{"running", models.TaskStatusRunning, false},
		{"passed", models.TaskStatusPassed, false},
		{"failed", models.TaskStatusFailed, false},
		{"skipped", models.TaskStatusSkipped, false},
		{"cancelled", models.TaskStatusCancelled, false},
		{"timeout", models.TaskStatusTimeout, false},
		{"no_resource", models.TaskStatusNoResource, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		result, err := models.ParseTaskStatus(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("expected error for input %s", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error for input %s: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		}
	}
}

func TestTaskTypeDisplayName(t *testing.T) {
	tests := []struct {
		taskType models.TaskType
		expected string
	}{
		{models.TaskTypeBasicCI, "Basic CI"},
		{models.TaskTypeBasicCIAll, "Basic CI All"},
		{models.TaskTypeAPITest, "API Test"},
		{models.TaskTypeAIAnalysis, "AI Analysis"},
		{models.TaskTypeSpecializedAIAnalysis, "Specialized AI E2E"},
	}

	for _, tt := range tests {
		if tt.taskType.DisplayName() != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.taskType.DisplayName())
		}
	}
}

func TestTaskTypeToQualityCheckType(t *testing.T) {
	tests := []struct {
		taskType      models.TaskType
		expectedCheck models.QualityCheckType
	}{
		{models.TaskTypeBasicCI, models.QualityCheckTypeBasicCI},
		{models.TaskTypeBasicCIAll, models.QualityCheckTypeBasicCIAll},
		{models.TaskTypeAIAnalysis, models.QualityCheckTypeAIAnalysis},
		{models.TaskTypeSpecializedAIAnalysis, models.QualityCheckTypeSpecializedAIAnalysis},
	}

	for _, tt := range tests {
		if tt.taskType.ToQualityCheckType() != tt.expectedCheck {
			t.Errorf("expected %s, got %s", tt.expectedCheck, tt.taskType.ToQualityCheckType())
		}
	}
}

func TestTaskIsAnalysisTask(t *testing.T) {
	tests := []struct {
		taskType models.TaskType
		expected bool
	}{
		{models.TaskTypeBasicCIAll, true},
		{models.TaskTypeSpecializedAIAnalysis, true},
		{models.TaskTypeBasicCI, false},
		{models.TaskTypeAPITest, false},
	}

	for _, tt := range tests {
		task := &models.Task{TaskType: tt.taskType}
		if task.IsAnalysisTask() != tt.expected {
			t.Errorf("expected %v for task type %s", tt.expected, tt.taskType)
		}
	}
}
