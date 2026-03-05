package monitor

import (
	"testing"
	"time"

	"github-hub/event-processor/internal/models"
)

type mockScheduler struct {
	tasks     []*models.Task
	completed []*models.Task
	failed    []*models.Task
	timedOut  []*models.Task
	cancelled []*models.Task
}

func newMockScheduler() *mockScheduler {
	return &mockScheduler{
		tasks:     make([]*models.Task, 0),
		completed: make([]*models.Task, 0),
		failed:    make([]*models.Task, 0),
	}
}

func (m *mockScheduler) GetRunningTasks() ([]*models.Task, error) {
	var running []*models.Task
	for _, t := range m.tasks {
		if t.Status == models.TaskStatusRunning {
			running = append(running, t)
		}
	}
	return running, nil
}

func (m *mockScheduler) CompleteTask(task *models.Task, results []models.TaskResult) error {
	m.completed = append(m.completed, task)
	return nil
}

func (m *mockScheduler) FailTask(task *models.Task, reason string) error {
	m.failed = append(m.failed, task)
	return nil
}

func (m *mockScheduler) TimeoutTask(task *models.Task, reason string) error {
	m.timedOut = append(m.timedOut, task)
	return nil
}

func (m *mockScheduler) CancelTask(task *models.Task, reason string) error {
	m.cancelled = append(m.cancelled, task)
	return nil
}

func (m *mockScheduler) SaveTaskResults(taskID int, results []models.TaskResult) error {
	// Mock implementation - does nothing
	return nil
}

func (m *mockScheduler) UpdateTask(task *models.Task) error {
	// Mock implementation - does nothing
	return nil
}

func (m *mockScheduler) TryStartAnalysis(taskID int) (bool, error) {
	// Mock implementation - always returns true (success)
	return true, nil
}

func TestNewMonitor(t *testing.T) {
	sched := newMockScheduler()
	mon := NewMonitor(sched)

	if mon == nil {
		t.Fatal("NewMonitor should not return nil")
	}
}

func TestMonitorCheckTaskStatusTimeout(t *testing.T) {
	sched := newMockScheduler()
	mon := NewMonitor(sched)

	// Task has been running for more than TaskTimeout (60 minutes)
	task := &models.Task{
		ID:           1,
		TaskName:     "basic_ci_all",
		Status:       models.TaskStatusRunning,
		ExecuteOrder: 1,
		StartTime:    &models.LocalTime{Time: time.Now().Add(-65 * time.Minute)},
	}

	mon.checkTaskStatus(task)

	if len(sched.timedOut) != 1 {
		t.Errorf("TimeoutTask should be called, got %d timeouts", len(sched.timedOut))
	}
}

func TestMonitorCheckTaskStatusRunning(t *testing.T) {
	sched := newMockScheduler()
	mon := NewMonitor(sched)

	task := &models.Task{
		ID:           1,
		TaskName:     "basic_ci_all",
		Status:       models.TaskStatusRunning,
		ExecuteOrder: 1,
		StartTime:    &models.LocalTime{Time: time.Now()},
		RequestURL:   "",
	}

	mon.checkTaskStatus(task)

	if len(sched.completed) != 0 && len(sched.failed) != 0 {
		t.Error("No action should be taken for running task without timeout")
	}
}

func TestMonitorCheckRunningTasks(t *testing.T) {
	sched := newMockScheduler()

	task1 := &models.Task{
		ID:        1,
		TaskName:  "basic_ci_all",
		Status:    models.TaskStatusRunning,
		StartTime: &models.LocalTime{Time: time.Now()},
	}
	// Task has been running for more than TaskTimeout (60 minutes)
	task2 := &models.Task{
		ID:        2,
		TaskName:  "deployment_deployment",
		Status:    models.TaskStatusRunning,
		StartTime: &models.LocalTime{Time: time.Now().Add(-65 * time.Minute)},
	}

	sched.tasks = append(sched.tasks, task1, task2)

	mon := NewMonitor(sched)
	mon.checkRunningTasks()

	if len(sched.timedOut) != 1 {
		t.Errorf("Should have 1 timed out task, got %d", len(sched.timedOut))
	}
}

func TestMockExecuteTask(t *testing.T) {
	tests := []struct {
		taskName    string
		resultCount int
		shouldError bool
	}{
		{"basic_ci_all", 4, false},
		{"deployment_deployment", 1, false},
		{"specialized_tests_api_test", 1, false},
		{"specialized_tests_module_e2e", 1, false},
		{"specialized_tests_agent_e2e", 1, false},
		{"specialized_tests_ai_e2e", 1, false},
		{"unknown_task", 0, true},
	}

	for _, tt := range tests {
		task := &models.Task{
			TaskName: tt.taskName,
			EventID:  1,
		}

		results, err := MockExecuteTask(task)

		if tt.shouldError {
			if err == nil {
				t.Errorf("MockExecuteTask(%s) should return error", tt.taskName)
			}
			continue
		}

		if err != nil {
			t.Errorf("MockExecuteTask(%s) error: %v", tt.taskName, err)
			continue
		}

		if len(results) != tt.resultCount {
			t.Errorf("MockExecuteTask(%s) result count = %d, want %d", tt.taskName, len(results), tt.resultCount)
		}
	}
}

func TestMockExecuteTaskBasicCIResults(t *testing.T) {
	task := &models.Task{
		TaskName: "basic_ci_all",
		EventID:  1,
	}

	results, err := MockExecuteTask(task)
	if err != nil {
		t.Fatalf("MockExecuteTask failed: %v", err)
	}

	expectedChecks := []string{"compilation", "code_lint", "security_scan", "unit_test"}
	for i, expected := range expectedChecks {
		if results[i].CheckType != expected {
			t.Errorf("Result[%d].CheckType = %v, want %v", i, results[i].CheckType, expected)
		}
		if results[i].Result != "pass" {
			t.Errorf("Result[%d].Result = %v, want 'pass'", i, results[i].Result)
		}
	}

	if results[3].Extra == nil {
		t.Error("unit_test result should have extra with score")
	}
}

func TestMockExecuteTaskDeploymentResults(t *testing.T) {
	task := &models.Task{
		TaskName: "deployment_deployment",
		EventID:  1,
	}

	results, err := MockExecuteTask(task)
	if err != nil {
		t.Fatalf("MockExecuteTask failed: %v", err)
	}

	if results[0].CheckType != "deployment" {
		t.Errorf("CheckType = %v, want 'deployment'", results[0].CheckType)
	}

	if results[0].Extra == nil {
		t.Error("deployment result should have extra with node info")
	}
}

func TestTaskTimeout(t *testing.T) {
	if TaskTimeout != 60*time.Minute {
		t.Errorf("TaskTimeout = %v, want 60m", TaskTimeout)
	}
}

func TestMonitorStartStop(t *testing.T) {
	sched := newMockScheduler()
	mon := NewMonitor(sched)

	mon.Start()

	time.Sleep(100 * time.Millisecond)

	mon.Stop()

	time.Sleep(100 * time.Millisecond)
}
