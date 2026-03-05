package scheduler

import (
	"testing"

	"github-hub/event-processor/internal/ai"
	"github-hub/event-processor/internal/api"
	"github-hub/event-processor/internal/models"
)

type mockTaskStorage struct {
	tasks   []*models.Task
	taskMap map[int][]*models.Task
}

type mockResourceStorage struct{}

func (m *mockResourceStorage) CreateResource(resource *models.ExecutableResource) error {
	return nil
}

func (m *mockResourceStorage) GetResource(id int) (*models.ExecutableResource, error) {
	return nil, nil
}

func (m *mockResourceStorage) GetAllResources() ([]*models.ExecutableResource, error) {
	// 返回空资源列表，让测试使用 mock URL
	return []*models.ExecutableResource{}, nil
}

func (m *mockResourceStorage) GetResourcesByCreator(creatorID int) ([]*models.ExecutableResource, error) {
	return nil, nil
}

func (m *mockResourceStorage) UpdateResource(resource *models.ExecutableResource) error {
	return nil
}

func (m *mockResourceStorage) DeleteResource(id int) error {
	return nil
}

func (m *mockResourceStorage) DeleteAllResources() error {
	return nil
}

type mockAIMatcher struct{}

func (m *mockAIMatcher) MatchResource(req *ai.MatchRequest) (*ai.MatchResult, error) {
	return nil, nil
}

func (m *mockAIMatcher) MatchResourceStream(req *ai.MatchRequest, callback func(string)) (*ai.MatchResult, error) {
	return nil, nil
}

func newMockResourceStorage() *mockResourceStorage {
	return &mockResourceStorage{}
}

func newMockAIMatcher() *mockAIMatcher {
	return &mockAIMatcher{}
}

func newMockTaskCreator() *TaskCreator {
	client := api.NewClient()
	store := newMockTaskStorage()
	// 传入 nil 作为 resourceStore 和 aiMatcher，这样会使用 mock URL
	return NewTaskCreator(client, store, nil, nil)
}

func newMockTaskStorage() *mockTaskStorage {
	return &mockTaskStorage{
		tasks:   make([]*models.Task, 0),
		taskMap: make(map[int][]*models.Task),
	}
}

func (m *mockTaskStorage) CreateTask(task *models.Task) error {
	m.tasks = append(m.tasks, task)
	m.taskMap[task.EventID] = append(m.taskMap[task.EventID], task)
	task.ID = len(m.tasks)
	return nil
}

func (m *mockTaskStorage) GetTask(id int) (*models.Task, error) {
	for _, t := range m.tasks {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, nil
}

func (m *mockTaskStorage) GetTaskByTaskID(taskID string) (*models.Task, error) {
	for _, t := range m.tasks {
		if t.TaskID == taskID {
			return t, nil
		}
	}
	return nil, nil
}

func (m *mockTaskStorage) ListTasks() ([]*models.Task, error) {
	return m.tasks, nil
}

func (m *mockTaskStorage) UpdateTask(task *models.Task) error {
	return nil
}

func (m *mockTaskStorage) DeleteTask(id int) error {
	return nil
}

func (m *mockTaskStorage) DeleteAllTasks() error {
	m.tasks = make([]*models.Task, 0)
	m.taskMap = make(map[int][]*models.Task)
	return nil
}

func (m *mockTaskStorage) GetPendingTasks() ([]*models.Task, error) {
	var pending []*models.Task
	for _, t := range m.tasks {
		if t.Status == models.TaskStatusPending {
			pending = append(pending, t)
		}
	}
	return pending, nil
}

func (m *mockTaskStorage) GetRunningTasks() ([]*models.Task, error) {
	var running []*models.Task
	for _, t := range m.tasks {
		if t.Status == models.TaskStatusRunning {
			running = append(running, t)
		}
	}
	return running, nil
}

func (m *mockTaskStorage) GetTasksByStatus(status models.TaskStatus) ([]*models.Task, error) {
	var result []*models.Task
	for _, t := range m.tasks {
		if t.Status == status {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockTaskStorage) GetTasksByEventID(eventID int) ([]*models.Task, error) {
	return m.taskMap[eventID], nil
}

func (m *mockTaskStorage) GetLatestTaskByEventID(eventID int) (*models.Task, error) {
	tasks := m.taskMap[eventID]
	if len(tasks) == 0 {
		return nil, nil
	}
	return tasks[len(tasks)-1], nil
}

func (m *mockTaskStorage) GetCompletedTasksByEventID(eventID int) ([]*models.Task, error) {
	var completed []*models.Task
	for _, t := range m.taskMap[eventID] {
		if t.IsCompleted() {
			completed = append(completed, t)
		}
	}
	return completed, nil
}

func (m *mockTaskStorage) SaveTaskResults(taskID int, results []models.TaskResult) error {
	return nil
}

func (m *mockTaskStorage) GetTaskResults(taskID int) ([]models.TaskResult, error) {
	return nil, nil
}

func (m *mockTaskStorage) MarkTaskCancelled(task *models.Task, reason string) error {
	task.Status = models.TaskStatusCancelled
	return nil
}

func (m *mockTaskStorage) MarkTaskTimeout(task *models.Task, reason string) error {
	task.Status = models.TaskStatusTimeout
	return nil
}

func (m *mockTaskStorage) MarkTaskSkipped(task *models.Task, reason string) error {
	task.Status = models.TaskStatusSkipped
	return nil
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

func (m *mockTaskStorage) TryStartAnalysisOrResetStale(taskID int) (bool, error) {
	return true, nil
}

func (m *mockTaskStorage) TryMarkTaskRunning(taskID int) (bool, error) {
	return true, nil
}

func (m *mockTaskStorage) Close() error {
	return nil
}

func TestNewTaskCreator(t *testing.T) {
	creator := newMockTaskCreator()
	if creator == nil {
		t.Fatal("NewTaskCreator should not return nil")
	}
}

func TestCreateFirstTask(t *testing.T) {
	creator := newMockTaskCreator()

	event := &api.Event{
		ID:          1,
		EventType:   "push",
		EventStatus: "pending",
	}

	task, err := creator.CreateFirstTask(event)
	if err != nil {
		t.Fatalf("CreateFirstTask failed: %v", err)
	}

	if task == nil {
		t.Fatal("CreateFirstTask should not return nil task")
	}

	if task.TaskName != "basic_ci_all" {
		t.Errorf("TaskName = %v, want 'basic_ci_all'", task.TaskName)
	}

	if task.ExecuteOrder != 1 {
		t.Errorf("ExecuteOrder = %d, want 1", task.ExecuteOrder)
	}

	if task.EventID != 1 {
		t.Errorf("EventID = %d, want 1", task.EventID)
	}

	if task.RequestURL == "" {
		t.Error("RequestURL should not be empty")
	}

	if task.TaskID == "" {
		t.Error("TaskID should be auto-generated")
	}
}

func TestCreateNextTask(t *testing.T) {
	creator := newMockTaskCreator()
	event := &api.Event{
		ID:          1,
		EventType:   "push",
		EventStatus: "pending",
	}

	tests := []struct {
		currentOrder int
		wantName     string
		wantOrder    int
		wantErr      string
	}{
		{1, "deployment_deployment", 2, ""},
		{2, "specialized_tests_api_test", 3, ""},
		{3, "specialized_tests_module_e2e", 4, ""},
		{4, "specialized_tests_agent_e2e", 5, ""},
		{5, "specialized_tests_ai_e2e", 6, ""},
		{6, "", -1, "no more tasks to create"},
	}

	for _, tt := range tests {
		task, err := creator.CreateNextTask(1, tt.currentOrder, event)

		if tt.wantErr != "" {
			if err == nil {
				t.Errorf("CreateNextTask(%d) expected error %q, but got nil", tt.currentOrder, tt.wantErr)
			}
			continue
		}

		if err != nil {
			t.Errorf("CreateNextTask(%d) error: %v", tt.currentOrder, err)
			continue
		}

		if task == nil {
			t.Errorf("CreateNextTask(%d) should not return nil", tt.currentOrder)
			continue
		}

		if task.TaskName != tt.wantName {
			t.Errorf("CreateNextTask(%d) TaskName = %v, want %v", tt.currentOrder, task.TaskName, tt.wantName)
		}

		if task.ExecuteOrder != tt.wantOrder {
			t.Errorf("CreateNextTask(%d) ExecuteOrder = %d, want %d", tt.currentOrder, task.ExecuteOrder, tt.wantOrder)
		}
	}
}

func TestCreateTaskForEvent(t *testing.T) {
	creator := newMockTaskCreator()

	event := &api.Event{
		ID:          1,
		EventType:   "push",
		EventStatus: "pending",
	}

	tests := []struct {
		order    int
		wantName string
	}{
		{1, "basic_ci_all"},
		{2, "deployment_deployment"},
		{3, "specialized_tests_api_test"},
		{4, "specialized_tests_module_e2e"},
		{5, "specialized_tests_agent_e2e"},
		{6, "specialized_tests_ai_e2e"},
	}

	for _, tt := range tests {
		task, err := creator.CreateTaskForEvent(event, tt.order)
		if err != nil {
			t.Errorf("CreateTaskForEvent(%d) error: %v", tt.order, err)
			continue
		}

		if task.TaskName != tt.wantName {
			t.Errorf("CreateTaskForEvent(%d) TaskName = %v, want %v", tt.order, task.TaskName, tt.wantName)
		}
	}
}

func TestGetMockURLs(t *testing.T) {
	creator := newMockTaskCreator()

	tests := []struct {
		taskName string
		wantPath string
	}{
		{"basic_ci_all", "/mock/basic-ci"},
		{"deployment_deployment", "/mock/deployment"},
		{"specialized_tests_api_test", "/mock/api-test"},
		{"specialized_tests_module_e2e", "/mock/module-e2e"},
		{"specialized_tests_agent_e2e", "/mock/agent-e2e"},
		{"specialized_tests_ai_e2e", "/mock/ai-e2e"},
	}

	for _, tt := range tests {
		requestURL := creator.getMockURL(tt.taskName, 1)

		if requestURL == "" {
			t.Errorf("getMockURL(%s) requestURL is empty", tt.taskName)
		}
	}
}

func TestShouldCreateNextTask(t *testing.T) {
	creator := newMockTaskCreator()
	store := newMockTaskStorage()

	event := &api.Event{ID: 1}
	task, _ := creator.CreateFirstTask(event)
	store.CreateTask(task)

	task.Status = models.TaskStatusPassed

	should := creator.ShouldCreateNextTask(1, task)
	if !should {
		t.Error("ShouldCreateNextTask should return true for passed task")
	}

	task.Status = models.TaskStatusFailed
	should = creator.ShouldCreateNextTask(1, task)
	if should {
		t.Error("ShouldCreateNextTask should return false for failed task")
	}

	task.Status = models.TaskStatusPassed
	task.ExecuteOrder = 6
	should = creator.ShouldCreateNextTask(1, task)
	if should {
		t.Error("ShouldCreateNextTask should return false for last task")
	}
}

func TestIsLastTask(t *testing.T) {
	creator := newMockTaskCreator()

	tests := []struct {
		order    int
		expected bool
	}{
		{1, false},
		{2, false},
		{3, false},
		{4, false},
		{5, false},
		{6, true},
		{7, true},
	}

	for _, tt := range tests {
		result := creator.IsLastTask(tt.order)
		if result != tt.expected {
			t.Errorf("IsLastTask(%d) = %v, want %v", tt.order, result, tt.expected)
		}
	}
}

func TestTaskExecutionOrder(t *testing.T) {
	creator := newMockTaskCreator()

	event := &api.Event{
		ID:          1,
		EventType:   "push",
		EventStatus: "pending",
	}

	task1, _ := creator.CreateFirstTask(event)
	if task1.ExecuteOrder != 1 {
		t.Errorf("First task ExecuteOrder = %d, want 1", task1.ExecuteOrder)
	}

	task2, _ := creator.CreateNextTask(1, 1, event)
	if task2.ExecuteOrder != 2 {
		t.Errorf("Second task ExecuteOrder = %d, want 2", task2.ExecuteOrder)
	}

	task3, _ := creator.CreateNextTask(1, 2, event)
	if task3.ExecuteOrder != 3 {
		t.Errorf("Third task ExecuteOrder = %d, want 3", task3.ExecuteOrder)
	}
}
