package scheduler

import (
	"testing"
	"time"

	"github-hub/event-processor/internal/api"
	"github-hub/event-processor/internal/models"
)

func TestNewSchedulerWithStorage(t *testing.T) {
	client := api.NewClient()
	store := newMockTaskStorage()
	resourceStore := newMockResourceStorage()
	aiMatcher := newMockAIMatcher()

	sched := NewSchedulerWithStorage(client, store, resourceStore, aiMatcher)
	if sched == nil {
		t.Fatal("NewSchedulerWithStorage should not return nil")
	}
}

func TestSchedulerProcessEvents(t *testing.T) {
	client := api.NewClient()
	store := newMockTaskStorage()
		resourceStore := newMockResourceStorage()
	aiMatcher := newMockAIMatcher()
	sched := NewSchedulerWithStorage(client, store, resourceStore, aiMatcher)

	events := []api.Event{
		{
			ID:          1,
			EventType:   "push",
			EventStatus: "pending",
			QualityChecks: []api.QualityCheck{
				{ID: 1, CheckType: "compilation", Stage: "basic_ci", StageOrder: 1, CheckOrder: 1},
				{ID: 2, CheckType: "code_lint", Stage: "basic_ci", StageOrder: 1, CheckOrder: 2},
			},
		},
		{
			ID:          2,
			EventType:   "push",
			EventStatus: "processing",
		},
		{
			ID:          3,
			EventType:   "push",
			EventStatus: "completed",
		},
	}

	err := sched.ProcessEvents(events)
	if err != nil {
		t.Fatalf("ProcessEvents failed: %v", err)
	}

	tasks, _ := store.GetTasksByEventID(1)
	if len(tasks) != 1 {
		t.Errorf("Event 1 should have 1 task, got %d", len(tasks))
	}

	tasks, _ = store.GetTasksByEventID(2)
	if len(tasks) != 0 {
		t.Errorf("Event 2 (processing) should have 0 tasks, got %d", len(tasks))
	}

	tasks, _ = store.GetTasksByEventID(3)
	if len(tasks) != 0 {
		t.Errorf("Event 3 (completed) should have 0 tasks, got %d", len(tasks))
	}
}

func TestSchedulerGetNextPendingTask(t *testing.T) {
	client := api.NewClient()
	store := newMockTaskStorage()
		resourceStore := newMockResourceStorage()
	aiMatcher := newMockAIMatcher()
	sched := NewSchedulerWithStorage(client, store, resourceStore, aiMatcher)

	event := &api.Event{
		ID:          1,
		EventType:   "push",
		EventStatus: "pending",
	}
	sched.eventCache[1] = event

	task1 := models.NewBasicCITask(1, 1, "http://test")
	task2 := models.NewTask(1, "deployment_deployment", "deployment", "deployment", 2, 1, 2, "http://test")

	store.CreateTask(task1)
	store.CreateTask(task2)

	task2.Status = models.TaskStatusPending

	pending, _, err := sched.GetNextPendingTask()
	if err != nil {
		t.Fatalf("GetNextPendingTask failed: %v", err)
	}

	if pending == nil {
		t.Fatal("GetNextPendingTask should return a task")
	}

	if pending.ExecuteOrder != 1 {
		t.Errorf("Pending task ExecuteOrder = %d, want 1", pending.ExecuteOrder)
	}
}

func TestSchedulerMarkTaskRunning(t *testing.T) {
	client := api.NewClient()
	store := newMockTaskStorage()
		resourceStore := newMockResourceStorage()
	aiMatcher := newMockAIMatcher()
	sched := NewSchedulerWithStorage(client, store, resourceStore, aiMatcher)

	task := models.NewBasicCITask(1, 1, "http://test")
	store.CreateTask(task)

	err := sched.MarkTaskRunning(task)
	if err != nil {
		t.Fatalf("MarkTaskRunning failed: %v", err)
	}

	if task.Status != models.TaskStatusRunning {
		t.Errorf("Task status = %v, want %v", task.Status, models.TaskStatusRunning)
	}

	if task.StartTime == nil {
		t.Error("Task StartTime should not be nil")
	}
}

func TestSchedulerCompleteTask(t *testing.T) {
	client := api.NewClient()
	store := newMockTaskStorage()
		resourceStore := newMockResourceStorage()
	aiMatcher := newMockAIMatcher()
	sched := NewSchedulerWithStorage(client, store, resourceStore, aiMatcher)

	event := &api.Event{
		ID:          1,
		EventType:   "push",
		EventStatus: "pending",
		QualityChecks: []api.QualityCheck{
			{ID: 1, CheckType: "compilation", Stage: "basic_ci", StageOrder: 1, CheckOrder: 1, CheckStatus: "pending"},
			{ID: 2, CheckType: "code_lint", Stage: "basic_ci", StageOrder: 1, CheckOrder: 2, CheckStatus: "pending"},
			{ID: 3, CheckType: "security_scan", Stage: "basic_ci", StageOrder: 1, CheckOrder: 3, CheckStatus: "pending"},
			{ID: 4, CheckType: "unit_test", Stage: "basic_ci", StageOrder: 1, CheckOrder: 4, CheckStatus: "pending"},
		},
	}
	sched.eventCache[1] = event

	task := models.NewBasicCITask(1, 1, "http://test")
	task.Status = models.TaskStatusRunning
	store.CreateTask(task)

	results := []models.TaskResult{
		{CheckType: "compilation", Result: "pass"},
		{CheckType: "code_lint", Result: "pass"},
		{CheckType: "security_scan", Result: "pass"},
		{CheckType: "unit_test", Result: "pass", Extra: map[string]interface{}{"score": 95}},
	}

	err := sched.CompleteTask(task, results)
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}

	if task.Status != models.TaskStatusPassed {
		t.Errorf("Task status = %v, want %v", task.Status, models.TaskStatusPassed)
	}

	tasks, _ := store.GetTasksByEventID(1)
	hasNextTask := false
	for _, t := range tasks {
		if t.ExecuteOrder == 2 {
			hasNextTask = true
		}
	}
	if !hasNextTask {
		t.Error("CompleteTask should create next task")
	}
}

func TestSchedulerFailTask(t *testing.T) {
	client := api.NewClient()
	store := newMockTaskStorage()
		resourceStore := newMockResourceStorage()
	aiMatcher := newMockAIMatcher()
	sched := NewSchedulerWithStorage(client, store, resourceStore, aiMatcher)

	event := &api.Event{
		ID:          1,
		EventType:   "push",
		EventStatus: "processing",
		QualityChecks: []api.QualityCheck{
			{ID: 1, CheckType: "compilation", Stage: "basic_ci", CheckStatus: "pending"},
		},
	}
	sched.eventCache[1] = event

	task := models.NewBasicCITask(1, 1, "http://test")
	task.Status = models.TaskStatusRunning
	store.CreateTask(task)

	err := sched.FailTask(task, "Build failed")
	if err != nil {
		t.Fatalf("FailTask failed: %v", err)
	}

	if task.Status != models.TaskStatusFailed {
		t.Errorf("Task status = %v, want %v", task.Status, models.TaskStatusFailed)
	}

	if task.ErrorMessage == nil || *task.ErrorMessage != "Build failed" {
		t.Errorf("ErrorMessage = %v, want 'Build failed'", task.ErrorMessage)
	}
}

func TestSchedulerCancelTask(t *testing.T) {
	client := api.NewClient()
	store := newMockTaskStorage()
		resourceStore := newMockResourceStorage()
	aiMatcher := newMockAIMatcher()
	sched := NewSchedulerWithStorage(client, store, resourceStore, aiMatcher)

	event := &api.Event{
		ID:          1,
		EventType:   "push",
		EventStatus: "processing",
		QualityChecks: []api.QualityCheck{
			{ID: 1, CheckType: "compilation", Stage: "basic_ci", CheckStatus: "running"},
		},
	}
	sched.eventCache[1] = event

	task := models.NewBasicCITask(1, 1, "http://test")
	task.Status = models.TaskStatusRunning
	store.CreateTask(task)

	err := sched.CancelTask(task, "Cancelled by user")
	if err != nil {
		t.Fatalf("CancelTask failed: %v", err)
	}

	if task.Status != models.TaskStatusCancelled {
		t.Errorf("Task status = %v, want %v", task.Status, models.TaskStatusCancelled)
	}
}

func TestSchedulerTimeoutTask(t *testing.T) {
	client := api.NewClient()
	store := newMockTaskStorage()
		resourceStore := newMockResourceStorage()
	aiMatcher := newMockAIMatcher()
	sched := NewSchedulerWithStorage(client, store, resourceStore, aiMatcher)

	event := &api.Event{
		ID:          1,
		EventType:   "push",
		EventStatus: "processing",
		QualityChecks: []api.QualityCheck{
			{ID: 1, CheckType: "compilation", Stage: "basic_ci", CheckStatus: "running"},
		},
	}
	sched.eventCache[1] = event

	task := models.NewBasicCITask(1, 1, "http://test")
	task.Status = models.TaskStatusRunning
	store.CreateTask(task)

	err := sched.TimeoutTask(task, "Execution timeout")
	if err != nil {
		t.Fatalf("TimeoutTask failed: %v", err)
	}

	if task.Status != models.TaskStatusTimeout {
		t.Errorf("Task status = %v, want %v", task.Status, models.TaskStatusTimeout)
	}
}

func TestSchedulerSkipTask(t *testing.T) {
	client := api.NewClient()
	store := newMockTaskStorage()
		resourceStore := newMockResourceStorage()
	aiMatcher := newMockAIMatcher()
	sched := NewSchedulerWithStorage(client, store, resourceStore, aiMatcher)

	event := &api.Event{
		ID:          1,
		EventType:   "push",
		EventStatus: "processing",
	}
	sched.eventCache[1] = event

	task := models.NewBasicCITask(1, 1, "http://test")
	task.Status = models.TaskStatusPending
	store.CreateTask(task)

	err := sched.SkipTask(task, "Skipped due to configuration")
	if err != nil {
		t.Fatalf("SkipTask failed: %v", err)
	}

	if task.Status != models.TaskStatusSkipped {
		t.Errorf("Task status = %v, want %v", task.Status, models.TaskStatusSkipped)
	}

	tasks, _ := store.GetTasksByEventID(1)
	hasNextTask := false
	for _, t := range tasks {
		if t.ExecuteOrder == 2 {
			hasNextTask = true
		}
	}
	if !hasNextTask {
		t.Error("SkipTask should create next task")
	}
}

func TestSchedulerGetRunningTasks(t *testing.T) {
	client := api.NewClient()
	store := newMockTaskStorage()
		resourceStore := newMockResourceStorage()
	aiMatcher := newMockAIMatcher()
	sched := NewSchedulerWithStorage(client, store, resourceStore, aiMatcher)

	task1 := models.NewBasicCITask(1, 1, "http://test")
	task1.Status = models.TaskStatusRunning
	task2 := models.NewBasicCITask(2, 1, "http://test")
	task2.Status = models.TaskStatusPending

	store.CreateTask(task1)
	store.CreateTask(task2)

	tasks, err := sched.GetRunningTasks()
	if err != nil {
		t.Fatalf("GetRunningTasks failed: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("GetRunningTasks should return 1 task, got %d", len(tasks))
	}
}

func TestSchedulerSequentialTaskCreation(t *testing.T) {
	client := api.NewClient()
	store := newMockTaskStorage()
		resourceStore := newMockResourceStorage()
	aiMatcher := newMockAIMatcher()
	sched := NewSchedulerWithStorage(client, store, resourceStore, aiMatcher)

	event := &api.Event{
		ID:          1,
		EventType:   "push",
		EventStatus: "pending",
		QualityChecks: []api.QualityCheck{
			{ID: 1, CheckType: "compilation", Stage: "basic_ci"},
			{ID: 2, CheckType: "deployment", Stage: "deployment"},
			{ID: 3, CheckType: "api_test", Stage: "specialized_tests"},
		},
	}
	sched.eventCache[1] = event

	task1, _ := sched.creator.CreateFirstTask(event)
	store.CreateTask(task1)

	tasks, _ := store.GetTasksByEventID(1)
	if len(tasks) != 1 {
		t.Errorf("Should have 1 task initially, got %d", len(tasks))
	}

	task1.Status = models.TaskStatusRunning
	task1.StartTime = &models.LocalTime{Time: time.Now()}

	results := []models.TaskResult{
		{CheckType: "compilation", Result: "pass"},
	}
	sched.CompleteTask(task1, results)

	tasks, _ = store.GetTasksByEventID(1)
	if len(tasks) != 2 {
		t.Errorf("Should have 2 tasks after first completion, got %d", len(tasks))
	}
}

func TestSchedulerEventCache(t *testing.T) {
	client := api.NewClient()
	store := newMockTaskStorage()
		resourceStore := newMockResourceStorage()
	aiMatcher := newMockAIMatcher()
	sched := NewSchedulerWithStorage(client, store, resourceStore, aiMatcher)

	event := &api.Event{
		ID:          1,
		EventType:   "push",
		EventStatus: "pending",
	}
	sched.eventCache[1] = event

	cached := sched.GetEventFromCache(1)
	if cached == nil {
		t.Error("GetEventFromCache should return event")
	}

	cached = sched.GetEventFromCache(999)
	if cached != nil {
		t.Error("GetEventFromCache should return nil for non-existent event")
	}
}

func TestSchedulerCompleteTaskWithFailedCheck(t *testing.T) {
	client := api.NewClient()
	store := newMockTaskStorage()
	resourceStore := newMockResourceStorage()
	aiMatcher := newMockAIMatcher()
	sched := NewSchedulerWithStorage(client, store, resourceStore, aiMatcher)

	task := &models.Task{
		ID:           1,
		TaskID:       "test-task-1",
		TaskName:     "basic_ci_all",
		EventID:      1,
		Stage:        "basic_ci",
		ExecuteOrder: 1,
		Status:       models.TaskStatusRunning,
		StartTime:    &models.LocalTime{Time: time.Now()},
	}

	// Results with one failed check
	results := []models.TaskResult{
		{CheckType: "compilation", Result: "pass"},
		{CheckType: "code_lint", Result: "fail"},
		{CheckType: "security_scan", Result: "pass"},
		{CheckType: "unit_test", Result: "pass"},
	}

	err := sched.CompleteTask(task, results)
	if err != nil {
		t.Fatalf("CompleteTask should not return error: %v", err)
	}

	// Task should be marked as failed
	if task.Status != models.TaskStatusFailed {
		t.Errorf("Task status should be failed, got %s", task.Status)
	}

	// Task should not be successful
	if task.IsSuccessful() {
		t.Error("Failed task should not be successful")
	}

	// Results should be saved
	if len(task.Results) != 4 {
		t.Errorf("Task should have 4 results, got %d", len(task.Results))
	}
}

func TestSchedulerCompleteTaskWithAllPassedChecks(t *testing.T) {
	client := api.NewClient()
	store := newMockTaskStorage()
	resourceStore := newMockResourceStorage()
	aiMatcher := newMockAIMatcher()
	sched := NewSchedulerWithStorage(client, store, resourceStore, aiMatcher)

	task := &models.Task{
		ID:           1,
		TaskID:       "test-task-1",
		TaskName:     "basic_ci_all",
		EventID:      1,
		Stage:        "basic_ci",
		ExecuteOrder: 1,
		Status:       models.TaskStatusRunning,
		StartTime:    &models.LocalTime{Time: time.Now()},
	}

	// All checks passed
	results := []models.TaskResult{
		{CheckType: "compilation", Result: "pass"},
		{CheckType: "code_lint", Result: "pass"},
		{CheckType: "security_scan", Result: "pass"},
		{CheckType: "unit_test", Result: "pass"},
	}

	err := sched.CompleteTask(task, results)
	if err != nil {
		t.Fatalf("CompleteTask should not return error: %v", err)
	}

	// Task should be marked as passed
	if task.Status != models.TaskStatusPassed {
		t.Errorf("Task status should be passed, got %s", task.Status)
	}

	// Task should be successful
	if !task.IsSuccessful() {
		t.Error("Task with all passed checks should be successful")
	}
}
