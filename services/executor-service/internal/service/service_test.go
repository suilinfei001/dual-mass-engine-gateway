package service

import (
	"context"
	"testing"
	"time"

	"github.com/quality-gateway/executor-service/internal/azure"
	"github.com/quality-gateway/shared/pkg/logger"
)

func TestExecutorService_ListExecutions(t *testing.T) {
	log := logger.New(logger.Config{Service: "test"})
	service := NewExecutorService(nil, log)

	executions := service.ListExecutions()
	if len(executions) != 0 {
		t.Errorf("Expected 0 executions for new service, got %d", len(executions))
	}
}

func TestExecutorService_GetExecutionStatus_NotFound(t *testing.T) {
	log := logger.New(logger.Config{Service: "test"})
	service := NewExecutorService(nil, log)

	_, err := service.GetExecutionStatus(context.Background(), "non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent execution")
	}
}

func TestExecutorService_CancelExecution_NotFound(t *testing.T) {
	log := logger.New(logger.Config{Service: "test"})
	service := NewExecutorService(nil, log)

	err := service.CancelExecution(context.Background(), "non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent execution")
	}
}

func TestExecutorService_CleanupOldExecutions(t *testing.T) {
	log := logger.New(logger.Config{Service: "test"})
	service := NewExecutorService(nil, log)

	count := service.CleanupOldExecutions(1 * time.Hour)
	if count != 0 {
		t.Errorf("Expected 0 executions cleaned up for new service, got %d", count)
	}
}

func TestExecutionRecord_Creation(t *testing.T) {
	now := time.Now()
	record := &ExecutionRecord{
		ExecutionID:  "test-exec-123",
		TaskUUID:     "test-task-456",
		RunID:        1001,
		Status:       azure.PipelineStatusRunning,
		Project:      "test-project",
		Organization: "test-org",
		CreatedAt:    now,
	}

	if record.ExecutionID != "test-exec-123" {
		t.Errorf("Expected ExecutionID=test-exec-123, got %s", record.ExecutionID)
	}
	if record.Status != azure.PipelineStatusRunning {
		t.Errorf("Expected Status=running, got %s", record.Status)
	}
}

func TestPipelineStatus_Constants(t *testing.T) {
	tests := []struct {
		status   azure.PipelineStatus
		expected string
	}{
		{azure.PipelineStatusPending, "pending"},
		{azure.PipelineStatusQueued, "queued"},
		{azure.PipelineStatusRunning, "running"},
		{azure.PipelineStatusCompleted, "completed"},
		{azure.PipelineStatusFailed, "failed"},
		{azure.PipelineStatusCanceled, "canceled"},
	}

	for _, tt := range tests {
		t.Run(string(tt.expected), func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.status))
			}
		})
	}
}

func TestPipelineResult_Constants(t *testing.T) {
	tests := []struct {
		result   azure.PipelineResult
		expected string
	}{
		{azure.PipelineResultSucceeded, "succeeded"},
		{azure.PipelineResultFailed, "failed"},
		{azure.PipelineResultCanceled, "canceled"},
	}

	for _, tt := range tests {
		t.Run(string(tt.result), func(t *testing.T) {
			if string(tt.result) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.result))
			}
		})
	}
}

func TestTaskExecutionRequest_Fields(t *testing.T) {
	req := &azure.TaskExecutionRequest{
		TaskUUID:          "test-uuid",
		TaskType:          "basic_ci_all",
		ChartURL:          "http://example.com/chart.tgz",
		TestbedIP:         "192.168.1.100",
		TestbedSSHPort:    22,
		TestbedSSHUser:    "root",
		TestbedSSHPassword: "password",
		Parameters: map[string]string{
			"param1": "value1",
		},
	}

	if req.TaskUUID != "test-uuid" {
		t.Errorf("Expected TaskUUID=test-uuid, got %s", req.TaskUUID)
	}
	if req.TaskType != "basic_ci_all" {
		t.Errorf("Expected TaskType=basic_ci_all, got %s", req.TaskType)
	}
	if len(req.Parameters) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(req.Parameters))
	}
}

func TestTaskExecutionStatus_Fields(t *testing.T) {
	startedAt := time.Now().Add(-1 * time.Hour)
	completedAt := time.Now()

	status := &azure.TaskExecutionStatus{
		ExecutionID: "exec-123",
		RunID:       1001,
		Status:      azure.PipelineStatusCompleted,
		Result:      azure.PipelineResultSucceeded,
		StartedAt:   &startedAt,
		CompletedAt: &completedAt,
		Finished:    true,
	}

	if status.ExecutionID != "exec-123" {
		t.Errorf("Expected ExecutionID=exec-123, got %s", status.ExecutionID)
	}
	if !status.Finished {
		t.Error("Expected Finished=true")
	}
}
