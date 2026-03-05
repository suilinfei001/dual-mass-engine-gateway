package mock

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewMockServer(t *testing.T) {
	server := NewMockServer(8090)
	if server == nil {
		t.Fatal("NewMockServer should not return nil")
	}
	if server.port != 8090 {
		t.Errorf("port = %d, want 8090", server.port)
	}
}

func TestMockServerHandleBasicCI(t *testing.T) {
	server := NewMockServer(8090)

	req := httptest.NewRequest("GET", "/mock/basic-ci?event_id=1", nil)
	w := httptest.NewRecorder()

	server.handleBasicCI(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var task MockTask
	if err := json.Unmarshal(w.Body.Bytes(), &task); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if task.TaskName != "basic_ci_all" {
		t.Errorf("TaskName = %v, want 'basic_ci_all'", task.TaskName)
	}

	if task.Status != "running" {
		t.Errorf("Status = %v, want 'running'", task.Status)
	}

	if task.EventID != 1 {
		t.Errorf("EventID = %d, want 1", task.EventID)
	}
}

func TestMockServerHandleDeployment(t *testing.T) {
	server := NewMockServer(8090)

	req := httptest.NewRequest("GET", "/mock/deployment?event_id=2", nil)
	w := httptest.NewRecorder()

	server.handleDeployment(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var task MockTask
	if err := json.Unmarshal(w.Body.Bytes(), &task); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if task.TaskName != "deployment_deployment" {
		t.Errorf("TaskName = %v, want 'deployment_deployment'", task.TaskName)
	}

	if task.ExecuteOrder != 2 {
		t.Errorf("ExecuteOrder = %d, want 2", task.ExecuteOrder)
	}
}

func TestMockServerHandleAPITest(t *testing.T) {
	server := NewMockServer(8090)

	req := httptest.NewRequest("GET", "/mock/api-test?event_id=3", nil)
	w := httptest.NewRecorder()

	server.handleAPITest(w, req)

	var task MockTask
	json.Unmarshal(w.Body.Bytes(), &task)

	if task.TaskName != "specialized_tests_api_test" {
		t.Errorf("TaskName = %v, want 'specialized_tests_api_test'", task.TaskName)
	}

	if task.ExecuteOrder != 3 {
		t.Errorf("ExecuteOrder = %d, want 3", task.ExecuteOrder)
	}
}

func TestMockServerHandleModuleE2E(t *testing.T) {
	server := NewMockServer(8090)

	req := httptest.NewRequest("GET", "/mock/module-e2e?event_id=4", nil)
	w := httptest.NewRecorder()

	server.handleModuleE2E(w, req)

	var task MockTask
	json.Unmarshal(w.Body.Bytes(), &task)

	if task.TaskName != "specialized_tests_module_e2e" {
		t.Errorf("TaskName = %v, want 'specialized_tests_module_e2e'", task.TaskName)
	}

	if task.ExecuteOrder != 4 {
		t.Errorf("ExecuteOrder = %d, want 4", task.ExecuteOrder)
	}
}

func TestMockServerHandleAgentE2E(t *testing.T) {
	server := NewMockServer(8090)

	req := httptest.NewRequest("GET", "/mock/agent-e2e?event_id=5", nil)
	w := httptest.NewRecorder()

	server.handleAgentE2E(w, req)

	var task MockTask
	json.Unmarshal(w.Body.Bytes(), &task)

	if task.TaskName != "specialized_tests_agent_e2e" {
		t.Errorf("TaskName = %v, want 'specialized_tests_agent_e2e'", task.TaskName)
	}

	if task.ExecuteOrder != 5 {
		t.Errorf("ExecuteOrder = %d, want 5", task.ExecuteOrder)
	}
}

func TestMockServerHandleAIE2E(t *testing.T) {
	server := NewMockServer(8090)

	req := httptest.NewRequest("GET", "/mock/ai-e2e?event_id=6", nil)
	w := httptest.NewRecorder()

	server.handleAIE2E(w, req)

	var task MockTask
	json.Unmarshal(w.Body.Bytes(), &task)

	if task.TaskName != "specialized_tests_ai_e2e" {
		t.Errorf("TaskName = %v, want 'specialized_tests_ai_e2e'", task.TaskName)
	}

	if task.ExecuteOrder != 6 {
		t.Errorf("ExecuteOrder = %d, want 6", task.ExecuteOrder)
	}
}

func TestMockServerHandleStatus(t *testing.T) {
	server := NewMockServer(8090)

	task := &MockTask{
		TaskID:   "test-task-id",
		TaskName: "basic_ci_all",
		Status:   "running",
	}
	server.tasks["test-task-id"] = task

	req := httptest.NewRequest("GET", "/mock/status/test-task-id", nil)
	w := httptest.NewRecorder()

	server.handleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var result MockTask
	json.Unmarshal(w.Body.Bytes(), &result)

	if result.TaskID != "test-task-id" {
		t.Errorf("TaskID = %v, want 'test-task-id'", result.TaskID)
	}

	req = httptest.NewRequest("GET", "/mock/status/non-existent", nil)
	w = httptest.NewRecorder()

	server.handleStatus(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status for non-existent task = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestMockServerHandleCancel(t *testing.T) {
	server := NewMockServer(8090)

	task := &MockTask{
		TaskID:   "test-task-id",
		TaskName: "basic_ci_all",
		Status:   "running",
	}
	server.tasks["test-task-id"] = task

	req := httptest.NewRequest("POST", "/mock/basic-ci/cancel?task_id=test-task-id", nil)
	w := httptest.NewRecorder()

	server.handleCancel(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	if task.Status != "cancelled" {
		t.Errorf("Task status = %v, want 'cancelled'", task.Status)
	}

	req = httptest.NewRequest("POST", "/mock/basic-ci/cancel", nil)
	w = httptest.NewRecorder()

	server.handleCancel(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status without task_id = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestMockServerGenerateResults(t *testing.T) {
	server := NewMockServer(8090)

	tests := []struct {
		taskName      string
		expectedCount int
	}{
		{"basic_ci_all", 4},
		{"deployment_deployment", 1},
		{"specialized_tests_api_test", 1},
		{"specialized_tests_module_e2e", 1},
		{"specialized_tests_agent_e2e", 1},
		{"specialized_tests_ai_e2e", 1},
		{"unknown", 0},
	}

	for _, tt := range tests {
		results := server.generateResults(tt.taskName)
		if len(results) != tt.expectedCount {
			t.Errorf("generateResults(%s) count = %d, want %d", tt.taskName, len(results), tt.expectedCount)
		}
	}
}

func TestMockServerTaskCompletion(t *testing.T) {
	server := NewMockServer(8090)

	req := httptest.NewRequest("GET", "/mock/basic-ci?event_id=1", nil)
	w := httptest.NewRecorder()

	server.handleBasicCI(w, req)

	var task MockTask
	json.Unmarshal(w.Body.Bytes(), &task)

	if task.Status != "running" {
		t.Errorf("Initial status = %v, want 'running'", task.Status)
	}

	time.Sleep(4 * time.Second)

	storedTask, exists := server.tasks[task.TaskID]
	if !exists {
		t.Fatal("Task should exist in server")
	}

	if storedTask.Status != "pass" {
		t.Errorf("After completion status = %v, want 'pass'", storedTask.Status)
	}

	if len(storedTask.Results) != 4 {
		t.Errorf("Results count = %d, want 4", len(storedTask.Results))
	}
}

func TestMockServerGetTaskStatus(t *testing.T) {
	server := NewMockServer(8090)

	task := &MockTask{
		TaskID:   "test-id",
		TaskName: "basic_ci_all",
		Status:   "pass",
	}
	server.tasks["test-id"] = task

	result, exists := server.GetTaskStatus("test-id")
	if !exists {
		t.Error("GetTaskStatus should return true for existing task")
	}
	if result.Status != "pass" {
		t.Errorf("Status = %v, want 'pass'", result.Status)
	}

	_, exists = server.GetTaskStatus("non-existent")
	if exists {
		t.Error("GetTaskStatus should return false for non-existing task")
	}
}

func TestMockServerGetEventID(t *testing.T) {
	server := NewMockServer(8090)

	req := httptest.NewRequest("GET", "/mock/basic-ci?event_id=123", nil)
	eventID := server.getEventID(req)

	if eventID != 123 {
		t.Errorf("EventID = %d, want 123", eventID)
	}

	req = httptest.NewRequest("GET", "/mock/basic-ci", nil)
	eventID = server.getEventID(req)

	if eventID != 0 {
		t.Errorf("EventID without param = %d, want 0", eventID)
	}
}

func TestMockServerRespondJSON(t *testing.T) {
	server := NewMockServer(8090)

	data := map[string]string{"key": "value"}
	w := httptest.NewRecorder()

	server.respondJSON(w, data)

	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Content-Type should be application/json")
	}

	var result map[string]string
	json.Unmarshal(w.Body.Bytes(), &result)

	if result["key"] != "value" {
		t.Errorf("Response key = %v, want 'value'", result["key"])
	}
}
