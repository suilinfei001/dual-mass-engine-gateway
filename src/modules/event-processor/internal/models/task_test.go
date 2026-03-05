package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTaskIsCompleted(t *testing.T) {
	tests := []struct {
		name     string
		status   TaskStatus
		expected bool
	}{
		{"pending", TaskStatusPending, false},
		{"running", TaskStatusRunning, false},
		{"passed", TaskStatusPassed, true},
		{"failed", TaskStatusFailed, true},
		{"timeout", TaskStatusTimeout, true},
		{"cancelled", TaskStatusCancelled, true},
		{"skipped", TaskStatusSkipped, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Status: tt.status}
			if task.IsCompleted() != tt.expected {
				t.Errorf("IsCompleted() = %v, want %v", task.IsCompleted(), tt.expected)
			}
		})
	}
}

func TestTaskIsRunning(t *testing.T) {
	task := &Task{Status: TaskStatusRunning}
	if !task.IsRunning() {
		t.Error("IsRunning() should return true for running status")
	}

	task.Status = TaskStatusPending
	if task.IsRunning() {
		t.Error("IsRunning() should return false for pending status")
	}
}

func TestTaskMarkRunning(t *testing.T) {
	task := &Task{Status: TaskStatusPending}
	task.MarkRunning()

	if task.Status != TaskStatusRunning {
		t.Errorf("Status = %v, want %v", task.Status, TaskStatusRunning)
	}
	if task.StartTime == nil {
		t.Error("StartTime should not be nil after MarkRunning")
	}
}

func TestTaskMarkPassed(t *testing.T) {
	task := &Task{Status: TaskStatusRunning}
	results := []TaskResult{
		{CheckType: "test", Result: "pass"},
	}
	task.MarkPassed(results)

	if task.Status != TaskStatusPassed {
		t.Errorf("Status = %v, want %v", task.Status, TaskStatusPassed)
	}
	if task.EndTime == nil {
		t.Error("EndTime should not be nil after MarkPassed")
	}
	if len(task.Results) != 1 {
		t.Errorf("Results length = %d, want 1", len(task.Results))
	}
}

func TestTaskMarkFailed(t *testing.T) {
	task := &Task{Status: TaskStatusRunning}
	task.MarkFailed("test error")

	if task.Status != TaskStatusFailed {
		t.Errorf("Status = %v, want %v", task.Status, TaskStatusFailed)
	}
	if task.ErrorMessage == nil || *task.ErrorMessage != "test error" {
		t.Errorf("ErrorMessage = %v, want 'test error'", task.ErrorMessage)
	}
}

func TestTaskMarkCancelled(t *testing.T) {
	task := &Task{Status: TaskStatusRunning}
	task.MarkCancelled("cancelled by user")

	if task.Status != TaskStatusCancelled {
		t.Errorf("Status = %v, want %v", task.Status, TaskStatusCancelled)
	}
}

func TestTaskMarkSkipped(t *testing.T) {
	task := &Task{Status: TaskStatusPending}
	task.MarkSkipped("skipped due to previous failure")

	if task.Status != TaskStatusSkipped {
		t.Errorf("Status = %v, want %v", task.Status, TaskStatusSkipped)
	}
}

func TestTaskMarkTimeout(t *testing.T) {
	task := &Task{Status: TaskStatusRunning}
	task.MarkTimeout("execution timeout")

	if task.Status != TaskStatusTimeout {
		t.Errorf("Status = %v, want %v", task.Status, TaskStatusTimeout)
	}
}

func TestTaskGenerateTaskID(t *testing.T) {
	task := &Task{}
	task.GenerateTaskID()

	if task.TaskID == "" {
		t.Error("TaskID should not be empty after GenerateTaskID")
	}
	if len(task.TaskID) != 36 {
		t.Errorf("TaskID length = %d, want 36 (UUID format)", len(task.TaskID))
	}
}

func TestTaskGetResultsJSON(t *testing.T) {
	task := &Task{
		Results: []TaskResult{
			{CheckType: "compilation", Result: "pass"},
		},
	}

	jsonStr := task.GetResultsJSON()
	if jsonStr == "" {
		t.Error("GetResultsJSON should not return empty string")
	}

	task.Results = nil
	jsonStr = task.GetResultsJSON()
	if jsonStr != "" {
		t.Errorf("GetResultsJSON should return empty string for nil results, got %s", jsonStr)
	}
}

func TestTaskSetResultsFromJSON(t *testing.T) {
	task := &Task{}
	jsonStr := `[{"check_type":"test","result":"pass"}]`

	err := task.SetResultsFromJSON(jsonStr)
	if err != nil {
		t.Errorf("SetResultsFromJSON failed: %v", err)
	}
	if len(task.Results) != 1 {
		t.Errorf("Results length = %d, want 1", len(task.Results))
	}

	err = task.SetResultsFromJSON("")
	if err != nil {
		t.Errorf("SetResultsFromJSON with empty string should not error: %v", err)
	}
}

func TestTaskIsSuccessful(t *testing.T) {
	tests := []struct {
		status   TaskStatus
		expected bool
	}{
		{TaskStatusPassed, true},
		{TaskStatusSkipped, true},
		{TaskStatusFailed, false},
		{TaskStatusRunning, false},
		{TaskStatusPending, false},
	}

	for _, tt := range tests {
		task := &Task{Status: tt.status}
		if task.IsSuccessful() != tt.expected {
			t.Errorf("IsSuccessful() for status %v = %v, want %v", tt.status, task.IsSuccessful(), tt.expected)
		}
	}
}

func TestNewTask(t *testing.T) {
	task := NewTask(1, "test_task", "compilation", "basic_ci", 1, 1, 1, "http://test")

	if task.EventID != 1 {
		t.Errorf("EventID = %d, want 1", task.EventID)
	}
	if task.TaskName != "test_task" {
		t.Errorf("TaskName = %v, want 'test_task'", task.TaskName)
	}
	if task.Status != TaskStatusPending {
		t.Errorf("Status = %v, want %v", task.Status, TaskStatusPending)
	}
	if task.TaskID == "" {
		t.Error("TaskID should be auto-generated")
	}
}

func TestNewBasicCITask(t *testing.T) {
	task := NewBasicCITask(1, 1, "http://test")

	if task.TaskName != "basic_ci_all" {
		t.Errorf("TaskName = %v, want 'basic_ci_all'", task.TaskName)
	}
	if task.Stage != "basic_ci" {
		t.Errorf("Stage = %v, want 'basic_ci'", task.Stage)
	}
	if task.ExecuteOrder != 1 {
		t.Errorf("ExecuteOrder = %d, want 1", task.ExecuteOrder)
	}
}

func TestGetNextExecuteOrder(t *testing.T) {
	tests := []struct {
		current int
		next    int
	}{
		{1, 2},
		{2, 3},
		{3, 4},
		{4, 5},
		{5, 6},
		{6, -1},
		{7, -1},
	}

	for _, tt := range tests {
		result := GetNextExecuteOrder(tt.current)
		if result != tt.next {
			t.Errorf("GetNextExecuteOrder(%d) = %d, want %d", tt.current, result, tt.next)
		}
	}
}

func TestGetTaskDefinitionByOrder(t *testing.T) {
	def := GetTaskDefinitionByOrder(1)
	if def == nil {
		t.Fatal("GetTaskDefinitionByOrder(1) should not return nil")
	}
	if def.TaskName != "basic_ci_all" {
		t.Errorf("TaskName = %v, want 'basic_ci_all'", def.TaskName)
	}

	def = GetTaskDefinitionByOrder(6)
	if def == nil {
		t.Fatal("GetTaskDefinitionByOrder(6) should not return nil")
	}
	if def.TaskName != "specialized_tests_ai_e2e" {
		t.Errorf("TaskName = %v, want 'specialized_tests_ai_e2e'", def.TaskName)
	}

	def = GetTaskDefinitionByOrder(99)
	if def != nil {
		t.Error("GetTaskDefinitionByOrder(99) should return nil")
	}
}

func TestTaskDefinitionsCount(t *testing.T) {
	if len(TaskDefinitions) != 6 {
		t.Errorf("TaskDefinitions count = %d, want 6", len(TaskDefinitions))
	}
}

func TestTaskDefinitionsOrder(t *testing.T) {
	for i, def := range TaskDefinitions {
		if def.ExecuteOrder != i+1 {
			t.Errorf("TaskDefinitions[%d].ExecuteOrder = %d, want %d", i, def.ExecuteOrder, i+1)
		}
	}
}

func TestLocalTimeScan(t *testing.T) {
	lt := &LocalTime{}

	err := lt.Scan(nil)
	if err != nil {
		t.Errorf("Scan(nil) should not error: %v", err)
	}
	if !lt.Time.IsZero() {
		t.Error("Time should be zero after scanning nil")
	}

	testTime := time.Now()
	err = lt.Scan(testTime)
	if err != nil {
		t.Errorf("Scan(time) should not error: %v", err)
	}
	if lt.Time != testTime {
		t.Error("Time should match scanned time")
	}
}

func TestLocalTimeMarshalJSON(t *testing.T) {
	lt := LocalTime{Time: time.Time{}}
	data, err := lt.MarshalJSON()
	if err != nil {
		t.Errorf("MarshalJSON should not error: %v", err)
	}
	if string(data) != "null" {
		t.Errorf("MarshalJSON for zero time = %s, want 'null'", string(data))
	}

	lt.Time = time.Now()
	data, err = lt.MarshalJSON()
	if err != nil {
		t.Errorf("MarshalJSON should not error: %v", err)
	}
	if string(data) == "null" {
		t.Error("MarshalJSON for non-zero time should not be 'null'")
	}
}

func TestTaskResultJSON(t *testing.T) {
	result := TaskResult{
		CheckType: "compilation",
		Result:    "pass",
		Output:    "Build successful",
		Extra: map[string]interface{}{
			"duration": 10.5,
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Errorf("Failed to marshal TaskResult: %v", err)
	}

	var unmarshaled TaskResult
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal TaskResult: %v", err)
	}

	if unmarshaled.CheckType != result.CheckType {
		t.Errorf("CheckType = %v, want %v", unmarshaled.CheckType, result.CheckType)
	}
}
