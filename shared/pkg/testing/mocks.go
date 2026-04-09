// Package testing provides testing utilities for all microservices.
package testing

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

	"github.com/quality-gateway/shared/pkg/models"
)

// MockServer is a mock HTTP server for testing.
type MockServer struct {
	server   *httptest.Server
	handlers map[string]http.HandlerFunc
	mu       sync.RWMutex
}

// NewMockServer creates a new mock server.
func NewMockServer() *MockServer {
	ms := &MockServer{
		handlers: make(map[string]http.HandlerFunc),
	}
	ms.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + ":" + r.URL.Path
		ms.mu.RLock()
		handler, ok := ms.handlers[key]
		ms.mu.RUnlock()

		if !ok {
			// Try with wildcard
			for k, h := range ms.handlers {
				if strings.HasSuffix(k, "*") && strings.HasPrefix(key, strings.TrimSuffix(k, "*")) {
					h(w, r)
					return
				}
			}
			http.NotFound(w, r)
			return
		}

		handler(w, r)
	}))
	return ms
}

// URL returns the mock server URL.
func (ms *MockServer) URL() string {
	return ms.server.URL
}

// Close closes the mock server.
func (ms *MockServer) Close() {
	ms.server.Close()
}

// Handle registers a handler for a method and path.
func (ms *MockServer) Handle(method, path string, handler http.HandlerFunc) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.handlers[method+":"+path] = handler
}

// HandleFunc registers a handler for a method and path.
func (ms *MockServer) HandleFunc(method, path string, handler func(w http.ResponseWriter, r *http.Request)) {
	ms.Handle(method, path, handler)
}

// HandleJSON registers a JSON response handler.
func (ms *MockServer) HandleJSON(method, path string, data interface{}) {
	ms.Handle(method, path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(data)
	})
}

// HandleStatus registers a status response handler.
func (ms *MockServer) HandleStatus(method, path string, status int) {
	ms.Handle(method, path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	})
}

// Reset clears all handlers.
func (ms *MockServer) Reset() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.handlers = make(map[string]http.HandlerFunc)
}

// RequestCount returns the number of requests made.
func (ms *MockServer) RequestCount() int {
	return len(ms.handlers)
}

// NewMockEvent creates a mock event for testing.
func NewMockEvent(overrides ...func(*models.Event)) *models.Event {
	event := &models.Event{
		ID:           1,
		EventID:      "test-event-1",
		EventType:    models.EventTypePROpened,
		PRNumber:     123,
		PRTitle:      "Test PR",
		SourceBranch: "feature/test",
		TargetBranch: "main",
		RepoURL:      "https://github.com/test/repo",
		Sender:       "test-user",
		EventStatus:  models.EventStatusPending,
	}

	for _, override := range overrides {
		override(event)
	}

	return event
}

// NewMockTask creates a mock task for testing.
func NewMockTask(overrides ...func(*models.Task)) *models.Task {
	task := &models.Task{
		ID:           1,
		EventID:      1,
		EventType:    models.EventTypePROpened,
		TaskType:     models.TaskTypeBasicCI,
		Status:       models.TaskStatusPending,
		PRNumber:     123,
		SourceBranch: "feature/test",
		TargetBranch: "main",
		RepoURL:      "https://github.com/test/repo",
		Analyzing:    false,
	}

	for _, override := range overrides {
		override(task)
	}

	return task
}

// NewMockQualityCheck creates a mock quality check for testing.
func NewMockQualityCheck(overrides ...func(*models.QualityCheck)) *models.QualityCheck {
	check := &models.QualityCheck{
		ID:          1,
		EventID:     1,
		CheckType:   models.QualityCheckTypeBasicCI,
		CheckStatus: models.QualityCheckStatusPending,
		StageOrder:  1,
		Result:      make(models.CheckResult),
	}

	for _, override := range overrides {
		override(check)
	}

	return check
}

// TestHelper provides common test helper functions.
type TestHelper struct {
	T interface {
		Cleanup(func())
		Errorf(string, ...interface{})
		Error(...interface{})
		FailNow()
	}
}

// NewTestHelper creates a new test helper.
func NewTestHelper(t interface {
	Cleanup(func())
	Errorf(string, ...interface{})
	Error(...interface{})
	FailNow()
}) *TestHelper {
	return &TestHelper{T: t}
}

// RequireNoError fails the test if err is not nil.
func (h *TestHelper) RequireNoError(err error) {
	if err != nil {
		h.T.Errorf("unexpected error: %v", err)
		h.T.FailNow()
	}
}

// RequireEqual fails the test if expected and actual are not equal.
func (h *TestHelper) RequireEqual(expected, actual interface{}) {
	if expected != actual {
		h.T.Errorf("expected %v, got %v", expected, actual)
		h.T.FailNow()
	}
}

// RequireNotNil fails the test if value is nil.
func (h *TestHelper) RequireNotNil(value interface{}) {
	if value == nil {
		h.T.Error("expected value to be non-nil")
		h.T.FailNow()
	}
}

// RequireTrue fails the test if condition is false.
func (h *TestHelper) RequireTrue(condition bool, msg string) {
	if !condition {
		h.T.Errorf("expected true, got false: %s", msg)
		h.T.FailNow()
	}
}

// RequireFalse fails the test if condition is true.
func (h *TestHelper) RequireFalse(condition bool, msg string) {
	if condition {
		h.T.Errorf("expected false, got true: %s", msg)
		h.T.FailNow()
	}
}

// JSONMatcher provides JSON matching utilities.
type JSONMatcher struct {
	data map[string]interface{}
}

// NewJSONMatcher creates a new JSON matcher.
func NewJSONMatcher(data map[string]interface{}) *JSONMatcher {
	return &JSONMatcher{data: data}
}

// HasField checks if a field exists.
func (m *JSONMatcher) HasField(field string) bool {
	_, ok := m.data[field]
	return ok
}

// GetField gets a field value.
func (m *JSONMatcher) GetField(field string) (interface{}, bool) {
	val, ok := m.data[field]
	return val, ok
}

// StringField gets a string field.
func (m *JSONMatcher) StringField(field string) (string, bool) {
	val, ok := m.data[field]
	if !ok {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

// IntField gets an int field.
func (m *JSONMatcher) IntField(field string) (int, bool) {
	val, ok := m.data[field]
	if !ok {
		return 0, false
	}
	// Handle both int and float64 (from JSON unmarshaling)
	switch v := val.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	}
	return 0, false
}

// NewMockResource creates a mock resource for testing.
func NewMockResource(overrides ...func(*models.Resource)) *models.Resource {
	resource := &models.Resource{
		ID:           1,
		UUID:         "test-resource-uuid",
		ResourceType: models.ResourceTypeBasicCI,
		Name:         "Test Resource",
		Description:  "Test resource description",
		AllowSkip:    false,
		Organization: "test-org",
		Project:      "test-project",
		PipelineID:   123,
		IsPublic:     true,
		CreatorID:    1,
	}

	for _, override := range overrides {
		override(resource)
	}

	return resource
}

// NewMockUser creates a mock user for testing.
func NewMockUser(overrides ...func(*models.User)) *models.User {
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		FullName: "Test User",
	}

	for _, override := range overrides {
		override(user)
	}

	return user
}
