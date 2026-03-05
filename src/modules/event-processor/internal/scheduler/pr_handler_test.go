package scheduler

import (
	"testing"
	"time"

	"github-hub/event-processor/internal/ai"
	"github-hub/event-processor/internal/api"
	"github-hub/event-processor/internal/models"
)

type testMockResourceStorage struct{}

func (m *testMockResourceStorage) CreateResource(resource *models.ExecutableResource) error {
	return nil
}

func (m *testMockResourceStorage) GetResource(id int) (*models.ExecutableResource, error) {
	return nil, nil
}

func (m *testMockResourceStorage) GetAllResources() ([]*models.ExecutableResource, error) {
	return nil, nil
}

func (m *testMockResourceStorage) GetResourcesByCreator(creatorID int) ([]*models.ExecutableResource, error) {
	return nil, nil
}

func (m *testMockResourceStorage) UpdateResource(resource *models.ExecutableResource) error {
	return nil
}

func (m *testMockResourceStorage) DeleteResource(id int) error {
	return nil
}

func (m *testMockResourceStorage) DeleteAllResources() error {
	return nil
}

type testMockAIMatcher struct{}

func (m *testMockAIMatcher) MatchResource(req *ai.MatchRequest) (*ai.MatchResult, error) {
	return nil, nil
}

func (m *testMockAIMatcher) MatchResourceStream(req *ai.MatchRequest, callback func(string)) (*ai.MatchResult, error) {
	return nil, nil
}

func newTestMockTaskCreator() *TaskCreator {
	client := api.NewClient()
	store := newMockTaskStorage()
	resourceStore := &testMockResourceStorage{}
	aiMatcher := &testMockAIMatcher{}
	return NewTaskCreator(client, store, resourceStore, aiMatcher)
}

func newTestMockScheduler() *SchedulerWithStorage {
	client := api.NewClient()
	store := newMockTaskStorage()
	resourceStore := &testMockResourceStorage{}
	aiMatcher := &testMockAIMatcher{}
	return NewSchedulerWithStorage(client, store, resourceStore, aiMatcher)
}

func TestNewPRHandler(t *testing.T) {
	client := api.NewClient()
	sched := newTestMockScheduler()
	creator := newTestMockTaskCreator()

	handler := NewPRHandler(client, creator, sched)
	if handler == nil {
		t.Fatal("NewPRHandler should not return nil")
	}
}

func TestPRHandlerGetPRAction(t *testing.T) {
	sched := newTestMockScheduler()
	creator := newTestMockTaskCreator()
	client := api.NewClient()
	handler := NewPRHandler(client, creator, sched)

	tests := []struct {
		eventType string
		payload   map[string]interface{}
		expected  string
	}{
		{"pull_request", map[string]interface{}{"action": "opened"}, "opened"},
		{"pull_request", map[string]interface{}{"action": "synchronize"}, "synchronize"},
		{"pull_request", map[string]interface{}{}, ""},
		{"push", map[string]interface{}{"action": "opened"}, ""},
	}

	for _, tt := range tests {
		event := &api.Event{
			EventType: tt.eventType,
			Payload:   tt.payload,
		}

		result := handler.getPRAction(event)
		if result != tt.expected {
			t.Errorf("getPRAction() = %v, want %v", result, tt.expected)
		}
	}
}

func TestPRHandlerGetPRURL(t *testing.T) {
	sched := newTestMockScheduler()
	creator := newTestMockTaskCreator()
	client := api.NewClient()
	handler := NewPRHandler(client, creator, sched)

	event := &api.Event{
		Payload: map[string]interface{}{
			"pr_url": "https://github.com/owner/repo/pull/1",
		},
	}

	url := handler.getPRURL(event)
	if url != "https://github.com/owner/repo/pull/1" {
		t.Errorf("getPRURL() = %v, want 'https://github.com/owner/repo/pull/1'", url)
	}

	event.Payload = map[string]interface{}{}
	url = handler.getPRURL(event)
	if url != "" {
		t.Errorf("getPRURL() = %v, want empty string", url)
	}
}

func TestPRHandlerIsPREvent(t *testing.T) {
	sched := newTestMockScheduler()
	creator := newTestMockTaskCreator()
	client := api.NewClient()
	handler := NewPRHandler(client, creator, sched)

	prEvent := &api.Event{EventType: "pull_request"}
	if !handler.IsPREvent(prEvent) {
		t.Error("IsPREvent should return true for pull_request event")
	}

	pushEvent := &api.Event{EventType: "push"}
	if handler.IsPREvent(pushEvent) {
		t.Error("IsPREvent should return false for push event")
	}
}

func TestPRHandlerHandlePROpened(t *testing.T) {
	store := newMockTaskStorage()
	resourceStore := &testMockResourceStorage{}
	aiMatcher := &testMockAIMatcher{}
	client := api.NewClient()
	sched := NewSchedulerWithStorage(client, store, resourceStore, aiMatcher)
	creator := NewTaskCreator(client, store, resourceStore, aiMatcher)
	handler := NewPRHandler(client, creator, sched)

	event := &api.Event{
		ID:          1,
		EventType:   "pull_request",
		EventStatus: "pending",
		Payload: map[string]interface{}{
			"action": "opened",
		},
	}

	err := handler.handlePROpened(event)
	if err != nil {
		t.Fatalf("handlePROpened failed: %v", err)
	}

	tasks, _ := store.GetTasksByEventID(1)
	if len(tasks) != 1 {
		t.Errorf("Should have 1 task after handlePROpened, got %d", len(tasks))
	}

	if tasks[0].TaskName != "basic_ci_all" {
		t.Errorf("Task name = %v, want 'basic_ci_all'", tasks[0].TaskName)
	}
}

func TestPRHandlerFilterEvents(t *testing.T) {
	sched := newTestMockScheduler()
	creator := newTestMockTaskCreator()
	client := api.NewClient()
	handler := NewPRHandler(client, creator, sched)

	now := time.Now().Format(time.RFC3339)

	tests := []struct {
		name              string
		relatedEvents     []*api.Event
		currentEvent      *api.Event
		wantTargetID      int
		wantCancelCount   int
		wantCompleteCount int
	}{
		{
			name: "single processing event",
			relatedEvents: []*api.Event{
				{ID: 1, EventStatus: "processing", CreatedAt: now},
			},
			currentEvent:      &api.Event{ID: 1, EventStatus: "processing", CreatedAt: now},
			wantTargetID:      1,
			wantCancelCount:   0,
			wantCompleteCount: 0,
		},
		{
			name: "processing with newer pending",
			relatedEvents: []*api.Event{
				{ID: 1, EventStatus: "processing", CreatedAt: "2024-01-01T00:00:00Z"},
				{ID: 2, EventStatus: "pending", CreatedAt: "2024-01-02T00:00:00Z"},
			},
			currentEvent:      &api.Event{ID: 2, EventStatus: "pending", CreatedAt: "2024-01-02T00:00:00Z"},
			wantTargetID:      2,
			wantCancelCount:   1,
			wantCompleteCount: 0,
		},
		{
			name: "multiple pending events",
			relatedEvents: []*api.Event{
				{ID: 1, EventStatus: "pending", CreatedAt: "2024-01-01T00:00:00Z"},
				{ID: 2, EventStatus: "pending", CreatedAt: "2024-01-02T00:00:00Z"},
			},
			currentEvent:      &api.Event{ID: 2, EventStatus: "pending", CreatedAt: "2024-01-02T00:00:00Z"},
			wantTargetID:      2,
			wantCancelCount:   0,
			wantCompleteCount: 1,
		},
		{
			name: "completed and failed events should be ignored",
			relatedEvents: []*api.Event{
				{ID: 1, EventStatus: "completed", CreatedAt: now},
				{ID: 2, EventStatus: "failed", CreatedAt: now},
				{ID: 3, EventStatus: "pending", CreatedAt: now},
			},
			currentEvent:      &api.Event{ID: 3, EventStatus: "pending", CreatedAt: now},
			wantTargetID:      3,
			wantCancelCount:   0,
			wantCompleteCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, eventsToCancel, eventsToComplete := handler.filterEvents(tt.relatedEvents, tt.currentEvent)

			if target == nil {
				if tt.wantTargetID != 0 {
					t.Errorf("target should not be nil, want ID %d", tt.wantTargetID)
				}
				return
			}

			if target.ID != tt.wantTargetID {
				t.Errorf("target ID = %d, want %d", target.ID, tt.wantTargetID)
			}

			if len(eventsToCancel) != tt.wantCancelCount {
				t.Errorf("eventsToCancel count = %d, want %d", len(eventsToCancel), tt.wantCancelCount)
			}

			if len(eventsToComplete) != tt.wantCompleteCount {
				t.Errorf("eventsToComplete count = %d, want %d", len(eventsToComplete), tt.wantCompleteCount)
			}
		})
	}
}

func TestPRHandlerHandlePREvent(t *testing.T) {
	store := newMockTaskStorage()
	resourceStore := &testMockResourceStorage{}
	aiMatcher := &testMockAIMatcher{}
	client := api.NewClient()
	sched := NewSchedulerWithStorage(client, store, resourceStore, aiMatcher)
	creator := NewTaskCreator(client, store, resourceStore, aiMatcher)
	handler := NewPRHandler(client, creator, sched)

	pushEvent := &api.Event{
		ID:        1,
		EventType: "push",
	}

	err := handler.HandlePREvent(pushEvent)
	if err != nil {
		t.Errorf("HandlePREvent for non-PR event should not error: %v", err)
	}

	prOpenedEvent := &api.Event{
		ID:          2,
		EventType:   "pull_request",
		EventStatus: "pending",
		Payload: map[string]interface{}{
			"action": "opened",
		},
	}

	err = handler.HandlePREvent(prOpenedEvent)
	if err != nil {
		t.Fatalf("HandlePREvent for opened PR failed: %v", err)
	}

	tasks, _ := store.GetTasksByEventID(2)
	if len(tasks) != 1 {
		t.Errorf("Should have 1 task for opened PR, got %d", len(tasks))
	}
}

func TestPRHandlerSynchronizeScenario(t *testing.T) {
	store := newMockTaskStorage()
	resourceStore := &testMockResourceStorage{}
	aiMatcher := &testMockAIMatcher{}
	client := api.NewClient()
	sched := NewSchedulerWithStorage(client, store, resourceStore, aiMatcher)
	creator := NewTaskCreator(client, store, resourceStore, aiMatcher)
	handler := NewPRHandler(client, creator, sched)

	prURL := "https://github.com/owner/repo/pull/1"

	event1 := &api.Event{
		ID:          1,
		EventType:   "pull_request",
		EventStatus: "processing",
		Payload: map[string]interface{}{
			"action": "synchronize",
			"pr_url": prURL,
		},
		CreatedAt: "2024-01-01T00:00:00Z",
	}

	task1 := models.NewBasicCITask(1, 1, "http://test")
	task1.Status = models.TaskStatusRunning
	task1.StartTime = &models.LocalTime{Time: time.Now()}
	store.CreateTask(task1)

	event2 := &api.Event{
		ID:          2,
		EventType:   "pull_request",
		EventStatus: "pending",
		Payload: map[string]interface{}{
			"action": "synchronize",
			"pr_url": prURL,
		},
		CreatedAt: "2024-01-02T00:00:00Z",
	}

	sched.eventCache[1] = event1
	sched.eventCache[2] = event2

	relatedEvents := []*api.Event{event1, event2}
	target, eventsToCancel, eventsToComplete := handler.filterEvents(relatedEvents, event2)

	if target == nil {
		t.Fatal("target should not be nil")
	}

	if target.ID != 2 {
		t.Errorf("target ID = %d, want 2 (newer pending event)", target.ID)
	}

	if len(eventsToCancel) != 1 {
		t.Errorf("eventsToCancel count = %d, want 1", len(eventsToCancel))
	}

	if len(eventsToComplete) != 0 {
		t.Errorf("eventsToComplete count = %d, want 0", len(eventsToComplete))
	}
}
