package service

import (
	"testing"

	"github.com/quality-gateway/event-store/internal/storage"
	"github.com/quality-gateway/shared/pkg/logger"
)

func TestEventStoreService_NewService(t *testing.T) {
	log := logger.New(logger.Config{Service: "test"})
	service := NewEventStoreService(nil, nil, log)
	if service == nil {
		t.Error("Expected service to be created")
	}
}

func TestEventStatus_Constants(t *testing.T) {
	tests := []struct {
		status   storage.EventStatus
		expected string
	}{
		{storage.EventStatusPending, "pending"},
		{storage.EventStatusProcessing, "processing"},
		{storage.EventStatusCompleted, "completed"},
		{storage.EventStatusFailed, "failed"},
		{storage.EventStatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.status))
			}
		})
	}
}

func TestEventType_Constants(t *testing.T) {
	tests := []struct {
		eventType storage.EventType
		expected  string
	}{
		{storage.EventTypePullRequestOpened, "pull_request.opened"},
		{storage.EventTypePullRequestSynchronized, "pull_request.synchronize"},
		{storage.EventTypePullRequestClosed, "pull_request.closed"},
		{storage.EventTypePush, "push"},
		{storage.EventTypeRelease, "release"},
	}

	for _, tt := range tests {
		t.Run(string(tt.eventType), func(t *testing.T) {
			if string(tt.eventType) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.eventType))
			}
		})
	}
}

func TestQualityCheckStatus_Constants(t *testing.T) {
	tests := []struct {
		status   storage.QualityCheckStatus
		expected string
	}{
		{storage.QualityCheckStatusPending, "pending"},
		{storage.QualityCheckStatusRunning, "running"},
		{storage.QualityCheckStatusPassed, "passed"},
		{storage.QualityCheckStatusFailed, "failed"},
		{storage.QualityCheckStatusSkipped, "skipped"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.status))
			}
		})
	}
}

func TestEvent_Fields(t *testing.T) {
	event := &storage.Event{
		ID:        1,
		UUID:      "test-uuid",
		EventType: storage.EventTypePullRequestOpened,
		Status:    storage.EventStatusPending,
		Source:    "github",
		RepoName:  "test/repo",
		PRNumber:  123,
		Author:    "testuser",
	}

	if event.UUID != "test-uuid" {
		t.Errorf("Expected UUID=test-uuid, got %s", event.UUID)
	}
	if event.EventType != storage.EventTypePullRequestOpened {
		t.Errorf("Expected EventType=pull_request.opened, got %s", event.EventType)
	}
	if event.Status != storage.EventStatusPending {
		t.Errorf("Expected Status=pending, got %s", event.Status)
	}
	if event.RepoName != "test/repo" {
		t.Errorf("Expected RepoName=test/repo, got %s", event.RepoName)
	}
	if event.PRNumber != 123 {
		t.Errorf("Expected PRNumber=123, got %d", event.PRNumber)
	}
}

func TestQualityCheck_Fields(t *testing.T) {
	check := &storage.QualityCheck{
		ID:          1,
		EventUUID:   "event-uuid",
		CheckType:   "code_lint",
		CheckStatus: "pending",
		Result:      "pass",
		Score:       95.0,
	}

	if check.EventUUID != "event-uuid" {
		t.Errorf("Expected EventUUID=event-uuid, got %s", check.EventUUID)
	}
	if check.CheckType != "code_lint" {
		t.Errorf("Expected CheckType=code_lint, got %s", check.CheckType)
	}
	if check.CheckStatus != "pending" {
		t.Errorf("Expected CheckStatus=pending, got %s", check.CheckStatus)
	}
	if check.Score != 95.0 {
		t.Errorf("Expected Score=95.0, got %f", check.Score)
	}
}

func TestEventFilter_Fields(t *testing.T) {
	filter := &storage.EventFilter{
		Status:   storage.EventStatusPending,
		RepoName: "test/repo",
		PRNumber: 123,
		Limit:    10,
		Offset:   0,
	}

	if filter.Status != storage.EventStatusPending {
		t.Errorf("Expected Status=pending, got %s", filter.Status)
	}
	if filter.RepoName != "test/repo" {
		t.Errorf("Expected RepoName=test/repo, got %s", filter.RepoName)
	}
	if filter.PRNumber != 123 {
		t.Errorf("Expected PRNumber=123, got %d", filter.PRNumber)
	}
	if filter.Limit != 10 {
		t.Errorf("Expected Limit=10, got %d", filter.Limit)
	}
}

func TestConfig_DSN(t *testing.T) {
	cfg := storage.Config{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "root",
		Password: "password",
	}

	dsn := cfg.DSN()
	expected := "root:password@tcp(localhost:3306)/testdb?parseTime=true&loc=Local"
	if dsn != expected {
		t.Errorf("Expected DSN=%s, got %s", expected, dsn)
	}
}
