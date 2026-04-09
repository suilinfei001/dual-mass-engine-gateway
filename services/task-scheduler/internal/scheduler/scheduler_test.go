package scheduler

import (
	"errors"
	"testing"
	"time"

	sharedmodels "github.com/quality-gateway/shared/pkg/models"
	"github.com/quality-gateway/task-scheduler/internal/models"
)

// mockTaskStorage 模拟任务存储
type mockTaskStorage struct {
	tasks    map[int]*models.Task
	results  map[int][]models.TaskResult
	errOnGet bool
}

func newMockTaskStorage() *mockTaskStorage {
	return &mockTaskStorage{
		tasks:   make(map[int]*models.Task),
		results: make(map[int][]models.TaskResult),
	}
}

func (m *mockTaskStorage) CreateTask(task *models.Task) error {
	task.ID = len(m.tasks) + 1
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskStorage) GetTask(id int) (*models.Task, error) {
	if m.errOnGet {
		return nil, errors.New("task not found")
	}
	return m.tasks[id], nil
}

func (m *mockTaskStorage) GetTasksByEventID(eventID int) ([]*models.Task, error) {
	var result []*models.Task
	for _, t := range m.tasks {
		if t.EventID == eventID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockTaskStorage) GetPendingTasks() ([]*models.Task, error) {
	var result []*models.Task
	for _, t := range m.tasks {
		if t.Status == sharedmodels.TaskStatusPending {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockTaskStorage) GetRunningTasks() ([]*models.Task, error) {
	var result []*models.Task
	for _, t := range m.tasks {
		if t.Status == sharedmodels.TaskStatusRunning {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockTaskStorage) UpdateTask(task *models.Task) error {
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskStorage) TryMarkTaskRunning(id int) (bool, error) {
	task, ok := m.tasks[id]
	if !ok {
		return false, nil
	}
	if task.Status != sharedmodels.TaskStatusPending {
		return false, nil
	}
	task.Status = sharedmodels.TaskStatusRunning
	return true, nil
}

func (m *mockTaskStorage) SaveTaskResults(taskID int, results []models.TaskResult) error {
	m.results[taskID] = results
	return nil
}

func (m *mockTaskStorage) GetTaskResults(taskID int) ([]models.TaskResult, error) {
	return m.results[taskID], nil
}

func (m *mockTaskStorage) CancelTasksByEventID(eventID int, reason string) (int, error) {
	count := 0
	for _, t := range m.tasks {
		if t.EventID == eventID && (t.Status == sharedmodels.TaskStatusPending || t.Status == sharedmodels.TaskStatusRunning) {
			t.Status = sharedmodels.TaskStatusCancelled
			t.ErrorMessage = &reason
			count++
		}
	}
	return count, nil
}

func (m *mockTaskStorage) ReleaseTestbed(taskID int) error {
	return nil
}

func (m *mockTaskStorage) ListTasks(limit, offset int) ([]*models.Task, error) {
	var result []*models.Task
	for _, t := range m.tasks {
		result = append(result, t)
	}
	return result, nil
}

func (m *mockTaskStorage) GetTaskCountByStatus(status sharedmodels.TaskStatus) (int, error) {
	count := 0
	for _, t := range m.tasks {
		if t.Status == status {
			count++
		}
	}
	return count, nil
}

func (m *mockTaskStorage) DeleteTask(id int) error {
	delete(m.tasks, id)
	return nil
}

func (m *mockTaskStorage) GetTasksByStatus(status sharedmodels.TaskStatus) ([]*models.Task, error) {
	var result []*models.Task
	for _, t := range m.tasks {
		if t.Status == status {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockTaskStorage) TryMarkTaskPassed(id int) (bool, error) {
	task, ok := m.tasks[id]
	if !ok {
		return false, nil
	}
	task.Status = sharedmodels.TaskStatusPassed
	return true, nil
}

func (m *mockTaskStorage) TryMarkTaskFailed(id int) (bool, error) {
	task, ok := m.tasks[id]
	if !ok {
		return false, nil
	}
	task.Status = sharedmodels.TaskStatusFailed
	return true, nil
}

func (m *mockTaskStorage) TryMarkTaskCancelled(id int) (bool, error) {
	task, ok := m.tasks[id]
	if !ok {
		return false, nil
	}
	task.Status = sharedmodels.TaskStatusCancelled
	return true, nil
}

func (m *mockTaskStorage) UpdateTaskAnalyzing(taskID int, analyzing bool) error {
	return nil
}

func (m *mockTaskStorage) IsTaskAnalyzing(taskID int) (bool, error) {
	return false, nil
}

func (m *mockTaskStorage) TryStartAnalysis(taskID int) (bool, error) {
	return true, nil
}

func (m *mockTaskStorage) ResetStaleAnalysisTasks(timeout time.Duration) (int, error) {
	return 0, nil
}

func (m *mockTaskStorage) UpdateTaskTestbed(taskID int, testbedUUID, testbedIP, sshUser, sshPassword, allocationUUID string) error {
	return nil
}

func (m *mockTaskStorage) GetLatestTaskByEventID(eventID int) (*models.Task, error) {
	var latest *models.Task
	for _, t := range m.tasks {
		if t.EventID == eventID {
			if latest == nil || t.ID > latest.ID {
				latest = t
			}
		}
	}
	return latest, nil
}

func (m *mockTaskStorage) GetTaskCountByEventID(eventID int) (int, error) {
	count := 0
	for _, t := range m.tasks {
		if t.EventID == eventID {
			count++
		}
	}
	return count, nil
}

func (m *mockTaskStorage) DeleteTaskResults(taskID int) error {
	delete(m.results, taskID)
	return nil
}

// mockEventStoreClient 模拟事件存储客户端
type mockEventStoreClient struct {
	events         map[int]*Event
	statusUpdates  []statusUpdate
	qualityUpdates []QualityCheckUpdate
}

type statusUpdate struct {
	eventID int
	status  string
}

func newMockEventStoreClient() *mockEventStoreClient {
	return &mockEventStoreClient{
		events: make(map[int]*Event),
	}
}

func (m *mockEventStoreClient) GetEvent(eventID int) (*Event, error) {
	return m.events[eventID], nil
}

func (m *mockEventStoreClient) UpdateEventStatus(eventID int, status string) error {
	m.statusUpdates = append(m.statusUpdates, statusUpdate{eventID: eventID, status: status})
	return nil
}

func (m *mockEventStoreClient) BatchUpdateQualityChecks(eventID int, updates []QualityCheckUpdate) error {
	m.qualityUpdates = append(m.qualityUpdates, updates...)
	return nil
}

func TestScheduler_CreateTask(t *testing.T) {
	store := newMockTaskStorage()
	eventStore := newMockEventStoreClient()
	scheduler := NewScheduler(store, eventStore)

	task, err := scheduler.CreateTask(1, "basic_ci_all", "code_lint", "basic_ci", 1, 1, 1, "http://example.com")
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	if task.EventID != 1 {
		t.Errorf("Expected EventID=1, got %d", task.EventID)
	}
	if task.TaskName != "basic_ci_all" {
		t.Errorf("Expected TaskName=basic_ci_all, got %s", task.TaskName)
	}
	if task.Status != sharedmodels.TaskStatusPending {
		t.Errorf("Expected status=pending, got %s", task.Status)
	}
}

func TestScheduler_GetTask(t *testing.T) {
	store := newMockTaskStorage()
	eventStore := newMockEventStoreClient()
	scheduler := NewScheduler(store, eventStore)

	// Create a task first
	created, _ := scheduler.CreateTask(1, "test_task", "test_check", "test_stage", 1, 1, 1, "http://example.com")

	// Get the task
	task, err := scheduler.GetTask(created.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}

	if task.ID != created.ID {
		t.Errorf("Expected ID=%d, got %d", created.ID, task.ID)
	}
}

func TestScheduler_StartTask(t *testing.T) {
	store := newMockTaskStorage()
	eventStore := newMockEventStoreClient()

	// Add an event with quality checks
	eventStore.events[1] = &Event{
		ID:          1,
		EventStatus: "pending",
		QualityChecks: []QualityCheck{
			{ID: 1, CheckType: "code_lint", CheckStatus: "pending", Stage: "basic_ci"},
		},
	}

	scheduler := NewScheduler(store, eventStore)

	// Create a task
	created, _ := scheduler.CreateTask(1, "basic_ci_all", "code_lint", "basic_ci", 1, 1, 1, "http://example.com")

	// Start the task
	task, err := scheduler.StartTask(created.ID)
	if err != nil {
		t.Fatalf("StartTask failed: %v", err)
	}

	if task.Status != sharedmodels.TaskStatusRunning {
		t.Errorf("Expected status=running, got %s", task.Status)
	}

	// Try to start again (should fail)
	_, err = scheduler.StartTask(created.ID)
	if err == nil {
		t.Error("Expected error when starting already running task")
	}
}

func TestScheduler_CompleteTask(t *testing.T) {
	store := newMockTaskStorage()
	eventStore := newMockEventStoreClient()

	eventStore.events[1] = &Event{
		ID:          1,
		EventStatus: "processing",
		QualityChecks: []QualityCheck{
			{ID: 1, CheckType: "code_lint", CheckStatus: "running", Stage: "basic_ci"},
		},
	}

	scheduler := NewScheduler(store, eventStore)

	created, _ := scheduler.CreateTask(1, "basic_ci_all", "code_lint", "basic_ci", 1, 1, 1, "http://example.com")
	scheduler.StartTask(created.ID)

	results := []models.TaskResult{
		{CheckType: "code_lint", Result: "pass", Output: "All checks passed"},
	}

	err := scheduler.CompleteTask(created.ID, results)
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}

	task, _ := scheduler.GetTask(created.ID)
	if task.Status != sharedmodels.TaskStatusPassed {
		t.Errorf("Expected status=passed, got %s", task.Status)
	}
}

func TestScheduler_CompleteTask_WithFailures(t *testing.T) {
	store := newMockTaskStorage()
	eventStore := newMockEventStoreClient()

	eventStore.events[1] = &Event{
		ID:          1,
		EventStatus: "processing",
		QualityChecks: []QualityCheck{
			{ID: 1, CheckType: "code_lint", CheckStatus: "running", Stage: "basic_ci"},
		},
	}

	scheduler := NewScheduler(store, eventStore)

	created, _ := scheduler.CreateTask(1, "basic_ci_all", "code_lint", "basic_ci", 1, 1, 1, "http://example.com")
	scheduler.StartTask(created.ID)

	results := []models.TaskResult{
		{CheckType: "code_lint", Result: "fail", Output: "Lint errors found"},
	}

	err := scheduler.CompleteTask(created.ID, results)
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}

	task, _ := scheduler.GetTask(created.ID)
	if task.Status != sharedmodels.TaskStatusFailed {
		t.Errorf("Expected status=failed, got %s", task.Status)
	}
}

func TestScheduler_FailTask(t *testing.T) {
	store := newMockTaskStorage()
	eventStore := newMockEventStoreClient()

	eventStore.events[1] = &Event{
		ID:          1,
		EventStatus: "processing",
		QualityChecks: []QualityCheck{
			{ID: 1, CheckType: "code_lint", CheckStatus: "running", Stage: "basic_ci"},
		},
	}

	scheduler := NewScheduler(store, eventStore)

	created, _ := scheduler.CreateTask(1, "basic_ci_all", "code_lint", "basic_ci", 1, 1, 1, "http://example.com")
	scheduler.StartTask(created.ID)

	err := scheduler.FailTask(created.ID, "Test failure")
	if err != nil {
		t.Fatalf("FailTask failed: %v", err)
	}

	task, _ := scheduler.GetTask(created.ID)
	if task.Status != sharedmodels.TaskStatusFailed {
		t.Errorf("Expected status=failed, got %s", task.Status)
	}
	if task.ErrorMessage == nil || *task.ErrorMessage != "Test failure" {
		t.Error("Expected error message to be set")
	}
}

func TestScheduler_CancelTask(t *testing.T) {
	store := newMockTaskStorage()
	eventStore := newMockEventStoreClient()

	eventStore.events[1] = &Event{
		ID:          1,
		EventStatus: "processing",
		QualityChecks: []QualityCheck{
			{ID: 1, CheckType: "code_lint", CheckStatus: "pending", Stage: "basic_ci"},
		},
	}

	scheduler := NewScheduler(store, eventStore)

	created, _ := scheduler.CreateTask(1, "basic_ci_all", "code_lint", "basic_ci", 1, 1, 1, "http://example.com")

	err := scheduler.CancelTask(created.ID, "PR synchronized")
	if err != nil {
		t.Fatalf("CancelTask failed: %v", err)
	}

	task, _ := scheduler.GetTask(created.ID)
	if task.Status != sharedmodels.TaskStatusCancelled {
		t.Errorf("Expected status=cancelled, got %s", task.Status)
	}
}

func TestScheduler_CancelEventTasks(t *testing.T) {
	store := newMockTaskStorage()
	eventStore := newMockEventStoreClient()

	eventStore.events[1] = &Event{
		ID:          1,
		EventStatus: "processing",
		QualityChecks: []QualityCheck{
			{ID: 1, CheckType: "code_lint", CheckStatus: "pending", Stage: "basic_ci"},
		},
	}

	scheduler := NewScheduler(store, eventStore)

	// Create multiple tasks
	scheduler.CreateTask(1, "task1", "check1", "stage1", 1, 1, 1, "http://example.com")
	scheduler.CreateTask(1, "task2", "check2", "stage2", 2, 2, 2, "http://example.com")
	scheduler.CreateTask(1, "task3", "check3", "stage3", 3, 3, 3, "http://example.com")

	count, err := scheduler.CancelEventTasks(1, "PR synchronized")
	if err != nil {
		t.Fatalf("CancelEventTasks failed: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 tasks cancelled, got %d", count)
	}

	// Verify event status was updated
	if len(eventStore.statusUpdates) == 0 {
		t.Error("Expected event status to be updated")
	}
	if eventStore.statusUpdates[0].status != "cancelled" {
		t.Errorf("Expected status=cancelled, got %s", eventStore.statusUpdates[0].status)
	}
}

func TestScheduler_GetPendingTasks(t *testing.T) {
	store := newMockTaskStorage()
	eventStore := newMockEventStoreClient()
	scheduler := NewScheduler(store, eventStore)

	// Create tasks with different execute orders
	scheduler.CreateTask(1, "task1", "check1", "stage1", 1, 1, 3, "http://example.com")
	scheduler.CreateTask(1, "task2", "check2", "stage2", 2, 2, 1, "http://example.com")
	scheduler.CreateTask(1, "task3", "check3", "stage3", 3, 3, 2, "http://example.com")

	tasks, err := scheduler.GetPendingTasks()
	if err != nil {
		t.Fatalf("GetPendingTasks failed: %v", err)
	}

	if len(tasks) != 3 {
		t.Fatalf("Expected 3 tasks, got %d", len(tasks))
	}

	// Verify sorted by execute order
	if tasks[0].ExecuteOrder != 1 {
		t.Errorf("Expected first task ExecuteOrder=1, got %d", tasks[0].ExecuteOrder)
	}
	if tasks[1].ExecuteOrder != 2 {
		t.Errorf("Expected second task ExecuteOrder=2, got %d", tasks[1].ExecuteOrder)
	}
	if tasks[2].ExecuteOrder != 3 {
		t.Errorf("Expected third task ExecuteOrder=3, got %d", tasks[2].ExecuteOrder)
	}
}
