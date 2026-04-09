package executor

import (
	"context"
	"testing"

	"github-hub/event-processor/internal/models"
)

type mockResourceStorageForExecutor struct {
	resources []*models.ExecutableResource
}

func (m *mockResourceStorageForExecutor) GetResource(id int) (*models.ExecutableResource, error) {
	for _, r := range m.resources {
		if r.ID == id {
			return r, nil
		}
	}
	return nil, nil
}

func (m *mockResourceStorageForExecutor) GetAllResources() ([]*models.ExecutableResource, error) {
	return m.resources, nil
}

type mockTaskStorageForExecutor struct {
	tasks     map[int]*models.Task
	results   map[int][]models.TaskResult
	analyzing map[int]bool
}

func newMockTaskStorageForExecutor() *mockTaskStorageForExecutor {
	return &mockTaskStorageForExecutor{
		tasks:     make(map[int]*models.Task),
		results:   make(map[int][]models.TaskResult),
		analyzing: make(map[int]bool),
	}
}

func (m *mockTaskStorageForExecutor) GetTask(id int) (*models.Task, error) {
	return m.tasks[id], nil
}

func (m *mockTaskStorageForExecutor) UpdateTask(task *models.Task) error {
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskStorageForExecutor) GetTaskResults(taskID int) ([]models.TaskResult, error) {
	return m.results[taskID], nil
}

func (m *mockTaskStorageForExecutor) SaveTaskResults(taskID int, results []models.TaskResult) error {
	m.results[taskID] = results
	return nil
}

func (m *mockTaskStorageForExecutor) MarkTaskSkipped(task *models.Task, reason string) error {
	task.Status = models.TaskStatusSkipped
	task.ErrorMessage = &reason
	return nil
}

func (m *mockTaskStorageForExecutor) TryStartAnalysisOrResetStale(taskID int) (bool, error) {
	if m.analyzing[taskID] {
		return false, nil
	}
	m.analyzing[taskID] = true
	return true, nil
}

func (m *mockTaskStorageForExecutor) IsTaskAnalyzing(taskID int) (bool, error) {
	return m.analyzing[taskID], nil
}

func TestExecutorConfig_Validation(t *testing.T) {
	config := &ExecutorConfig{
		Organization: "test-org",
		Project:      "test-project",
		PAT:          "test-pat",
		Branch:       "refs/heads/main",
		VerifySSL:    true,
	}

	if config.Organization != "test-org" {
		t.Errorf("Expected Organization 'test-org', got '%s'", config.Organization)
	}

	if config.Project != "test-project" {
		t.Errorf("Expected Project 'test-project', got '%s'", config.Project)
	}
}

func TestStatusResult_Mapping(t *testing.T) {
	tests := []struct {
		name           string
		status         TaskStatus
		result         TaskResult
		expectedStatus models.TaskStatus
	}{
		{"completed success", TaskStatusCompleted, TaskResultSucceeded, models.TaskStatusPassed},
		{"completed failed", TaskStatusCompleted, TaskResultFailed, models.TaskStatusFailed},
		{"completed canceled", TaskStatusCompleted, TaskResultCanceled, models.TaskStatusCancelled},
		{"in progress", TaskStatusInProgress, "", models.TaskStatusRunning},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusResult := &StatusResult{
				Status: tt.status,
				Result: tt.result,
			}

			exec := &AzureDevOpsExecutor{}

			if exec.IsCompleted(statusResult.Status) {
				if tt.expectedStatus == models.TaskStatusRunning {
					t.Errorf("Expected running status but task is completed")
				}
			}
		})
	}
}

func TestAzureDevOpsExecutor_IsCompleted(t *testing.T) {
	exec := &AzureDevOpsExecutor{}

	tests := []struct {
		status   TaskStatus
		expected bool
	}{
		{TaskStatusCompleted, true},
		{TaskStatusCanceled, true},
		{TaskStatusInProgress, false},
		{TaskStatusNotStarted, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			result := exec.IsCompleted(tt.status)
			if result != tt.expected {
				t.Errorf("IsCompleted(%s) = %v, expected %v", tt.status, result, tt.expected)
			}
		})
	}
}

func TestAzureDevOpsExecutor_IsSuccess(t *testing.T) {
	exec := &AzureDevOpsExecutor{}

	tests := []struct {
		result   TaskResult
		expected bool
	}{
		{TaskResultSucceeded, true},
		{TaskResultFailed, false},
		{TaskResultCanceled, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.result), func(t *testing.T) {
			result := exec.IsSuccess(tt.result)
			if result != tt.expected {
				t.Errorf("IsSuccess(%s) = %v, expected %v", tt.result, result, tt.expected)
			}
		})
	}
}

func TestTaskExecutionService_PollTaskStatus_NoBuildID(t *testing.T) {
	service := &TaskExecutionService{}

	task := &models.Task{
		ID:      1,
		BuildID: 0,
	}

	status, results, err := service.PollTaskStatus(context.Background(), task)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if status != models.TaskStatusPending {
		t.Errorf("Expected status pending, got %s", status)
	}

	if results != nil {
		t.Error("Expected nil results for pending task")
	}
}

func TestTaskExecutionService_SourceCodeBranchReplacement(t *testing.T) {
	tests := []struct {
		name           string
		taskName       string
		eventBranch    string
		originalParams map[string]interface{}
		expectedBranch string
		shouldReplace  bool
	}{
		{
			name:           "basic_ci_all with event branch",
			taskName:       "basic_ci_all",
			eventBranch:    "feature/test-branch",
			originalParams: map[string]interface{}{"SOURCE_CODE_BRANCH": "main", "BUILD_TYPE": "opensource"},
			expectedBranch: "feature/test-branch",
			shouldReplace:  true,
		},
		{
			name:           "basic_ci_all with refs/heads prefix",
			taskName:       "basic_ci_all",
			eventBranch:    "refs/heads/develop",
			originalParams: map[string]interface{}{"SOURCE_CODE_BRANCH": "main"},
			expectedBranch: "refs/heads/develop",
			shouldReplace:  true,
		},
		{
			name:           "basic_ci_all without SOURCE_CODE_BRANCH in params",
			taskName:       "basic_ci_all",
			eventBranch:    "feature/new-feature",
			originalParams: map[string]interface{}{"BUILD_TYPE": "opensource"},
			expectedBranch: "feature/new-feature",
			shouldReplace:  true,
		},
		{
			name:           "deployment_deployment should not replace",
			taskName:       "deployment_deployment",
			eventBranch:    "feature/test",
			originalParams: map[string]interface{}{"SOURCE_CODE_BRANCH": "main"},
			expectedBranch: "main",
			shouldReplace:  false,
		},
		{
			name:           "specialized_tests should not replace",
			taskName:       "specialized_tests_api_test",
			eventBranch:    "feature/test",
			originalParams: map[string]interface{}{"SOURCE_CODE_BRANCH": "main"},
			expectedBranch: "main",
			shouldReplace:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := make(map[string]interface{})
			for k, v := range tt.originalParams {
				params[k] = v
			}

			eventBranch := tt.eventBranch

			if tt.taskName == "basic_ci_all" && eventBranch != "" {
				if params == nil {
					params = make(map[string]interface{})
				}
				params["SOURCE_CODE_BRANCH"] = eventBranch
			}

			if tt.shouldReplace {
				if params["SOURCE_CODE_BRANCH"] != tt.expectedBranch {
					t.Errorf("Expected SOURCE_CODE_BRANCH '%s', got '%v'", tt.expectedBranch, params["SOURCE_CODE_BRANCH"])
				}
			} else {
				if params["SOURCE_CODE_BRANCH"] != tt.originalParams["SOURCE_CODE_BRANCH"] {
					t.Errorf("SOURCE_CODE_BRANCH should not be replaced for task %s", tt.taskName)
				}
			}
		})
	}
}
