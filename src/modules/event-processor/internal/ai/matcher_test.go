package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github-hub/event-processor/internal/models"
	"github-hub/event-processor/internal/storage"
)

type mockConfigProvider struct {
	config *storage.AIConfig
	err    error
}

func (m *mockConfigProvider) GetAIConfig() (*storage.AIConfig, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.config == nil {
		return &storage.AIConfig{}, nil
	}
	return m.config, nil
}

func TestAIMatcher_MatchResource_NoConfig(t *testing.T) {
	mockProvider := &mockConfigProvider{config: &storage.AIConfig{}}
	matcher := NewAIMatcherWithProvider(mockProvider)

	req := &MatchRequest{
		TaskName: "test_task",
		Resources: []*models.ExecutableResource{
			{ID: 1, ResourceName: "test_resource"},
		},
	}

	_, err := matcher.MatchResource(req)
	if err != ErrAINotConfigured {
		t.Errorf("Expected ErrAINotConfigured, got %v", err)
	}
}

func TestAIMatcher_MatchResource_Success(t *testing.T) {
	aiResponse := MatchResult{
		ResourceID:   1,
		ResourceName: "test_resource",
		RequestURL:   "http://test.url",
		Confidence:   0.95,
		Reasoning:    "Test reasoning",
	}

	responseJSON, _ := json.Marshal(aiResponse)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/mf-model-api/v1/chat/completions" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"content": string(responseJSON),
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	mockProvider := &mockConfigProvider{
		config: &storage.AIConfig{
			IP:    strings.TrimPrefix(server.URL, "http://"),
			Model: "test-model",
			Token: "test-token",
		},
	}

	matcher := NewAIMatcherWithProvider(mockProvider)

	req := &MatchRequest{
		TaskName: "test_task",
		EventDetail: map[string]interface{}{
			"event_type": "push",
		},
		Resources: []*models.ExecutableResource{
			{ID: 1, ResourceName: "test_resource", ResourceType: "test_task"},
		},
	}

	result, err := matcher.MatchResource(req)
	if err != nil {
		t.Fatalf("MatchResource failed: %v", err)
	}

	if result.ResourceID != 1 {
		t.Errorf("Expected ResourceID 1, got %d", result.ResourceID)
	}

	if result.ResourceName != "test_resource" {
		t.Errorf("Expected ResourceName 'test_resource', got '%s'", result.ResourceName)
	}

	if result.Confidence != 0.95 {
		t.Errorf("Expected Confidence 0.95, got %f", result.Confidence)
	}
}

func TestAIMatcher_BuildUserQuery(t *testing.T) {
	matcher := &AIMatcher{}

	req := &MatchRequest{
		TaskName: "basic_ci_all",
		EventDetail: map[string]interface{}{
			"event_type": "push",
			"repository": "test/repo",
		},
		Resources: []*models.ExecutableResource{
			{ID: 1, ResourceName: "basic_ci_all_resource", ResourceType: "basic_ci_all"},
		},
	}

	query := matcher.buildUserQuery(req)

	if !strings.Contains(query, "basic_ci_all") {
		t.Error("Query should contain task name")
	}

	if !strings.Contains(query, "event_type") {
		t.Error("Query should contain event details")
	}

	if !strings.Contains(query, "resource_type='basic_ci_all'") {
		t.Error("Query should contain resource_type matching requirement")
	}
}

func TestAIMatcher_GetDefaultSystemPrompt(t *testing.T) {
	matcher := &AIMatcher{}

	prompt := matcher.GetDefaultSystemPrompt()

	if prompt == "" {
		t.Error("Default system prompt should not be empty")
	}
}

func TestAIClient_Chat_NoConfig(t *testing.T) {
	mockProvider := &mockConfigProvider{config: &storage.AIConfig{}}
	client := NewAIClientWithProvider(mockProvider)

	_, err := client.Chat(&ChatRequest{
		SystemPrompt: "test",
		UserPrompt:   "test",
	})

	if err != ErrAINotConfigured {
		t.Errorf("Expected ErrAINotConfigured, got %v", err)
	}
}

func TestAIClient_Chat_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Authorization header 'Bearer test-token', got '%s'", r.Header.Get("Authorization"))
		}

		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"content": "test response",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	mockProvider := &mockConfigProvider{
		config: &storage.AIConfig{
			IP:    strings.TrimPrefix(server.URL, "http://"),
			Model: "test-model",
			Token: "test-token",
		},
	}

	client := NewAIClientWithProvider(mockProvider)

	resp, err := client.Chat(&ChatRequest{
		SystemPrompt: "test system",
		UserPrompt:   "test user",
	})

	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if resp.Content != "test response" {
		t.Errorf("Expected content 'test response', got '%s'", resp.Content)
	}
}

func TestAIClient_Chat_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	mockProvider := &mockConfigProvider{
		config: &storage.AIConfig{
			IP:    strings.TrimPrefix(server.URL, "http://"),
			Model: "test-model",
			Token: "test-token",
		},
	}

	client := NewAIClientWithProvider(mockProvider)

	_, err := client.Chat(&ChatRequest{
		SystemPrompt: "test",
		UserPrompt:   "test",
	})

	if err == nil {
		t.Error("Expected error for HTTP 500")
	}
}

func TestAIClient_Chat_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	mockProvider := &mockConfigProvider{
		config: &storage.AIConfig{
			IP:    strings.TrimPrefix(server.URL, "http://"),
			Model: "test-model",
			Token: "test-token",
		},
	}

	client := NewAIClientWithProvider(mockProvider)

	_, err := client.Chat(&ChatRequest{
		SystemPrompt: "test",
		UserPrompt:   "test",
	})

	if err == nil {
		t.Error("Expected error for empty response")
	}
}
