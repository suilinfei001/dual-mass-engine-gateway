package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_NewClient(t *testing.T) {
	client := NewClient()

	if client == nil {
		t.Fatal("Expected client to be created")
	}

	if client.BaseURL != DefaultEventReceiverAPI {
		t.Errorf("Expected BaseURL to be '%s', got '%s'", DefaultEventReceiverAPI, client.BaseURL)
	}

	if client.HTTPClient == nil {
		t.Error("Expected HTTPClient to be initialized")
	}
}

func TestClient_NewClientWithURL(t *testing.T) {
	customURL := "http://custom-server:8080"
	client := NewClientWithURL(customURL)

	if client.BaseURL != customURL {
		t.Errorf("Expected BaseURL '%s', got '%s'", customURL, client.BaseURL)
	}
}

func TestClient_SetAPIToken(t *testing.T) {
	client := NewClient()
	token := "test-token-123"

	client.SetAPIToken(token)

	if client.apiToken != token {
		t.Errorf("Expected apiToken '%s', got '%s'", token, client.apiToken)
	}
}

func TestClient_GetEvents_Success(t *testing.T) {
	events := []Event{
		{ID: 1, EventID: "evt-1", EventType: "push", EventStatus: "pending"},
		{ID: 2, EventID: "evt-2", EventType: "pull_request", EventStatus: "processing"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/events" {
			t.Errorf("Expected path '/api/events', got '%s'", r.URL.Path)
		}

		resp := APIResponse{
			Success: true,
			Data:    events,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	result, err := client.GetEvents()

	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 events, got %d", len(result))
	}

	if result[0].EventID != "evt-1" {
		t.Errorf("Expected EventID 'evt-1', got '%s'", result[0].EventID)
	}
}

func TestClient_GetEvent_Success(t *testing.T) {
	event := Event{
		ID:          1,
		EventID:     "evt-1",
		EventType:   "push",
		EventStatus: "pending",
		Repository:  "test/repo",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/events/1" {
			t.Errorf("Expected path '/api/events/1', got '%s'", r.URL.Path)
		}

		resp := APIResponse{
			Success: true,
			Data:    event,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	result, err := client.GetEvent(1)

	if err != nil {
		t.Fatalf("GetEvent failed: %v", err)
	}

	if result.EventID != "evt-1" {
		t.Errorf("Expected EventID 'evt-1', got '%s'", result.EventID)
	}

	if result.Repository != "test/repo" {
		t.Errorf("Expected Repository 'test/repo', got '%s'", result.Repository)
	}
}

func TestClient_UpdateEventStatus_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/events/1/status" {
			t.Errorf("Expected path '/api/events/1/status', got '%s'", r.URL.Path)
		}

		if r.Method != "PUT" {
			t.Errorf("Expected PUT method, got '%s'", r.Method)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	err := client.UpdateEventStatus(1, "completed", "2024-01-01T00:00:00Z")

	if err != nil {
		t.Fatalf("UpdateEventStatus failed: %v", err)
	}
}

func TestClient_UpdateQualityCheck_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/quality-checks/1" {
			t.Errorf("Expected path '/api/quality-checks/1', got '%s'", r.URL.Path)
		}

		if r.Method != "PUT" {
			t.Errorf("Expected PUT method, got '%s'", r.Method)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	err := client.UpdateQualityCheck(1, "passed", "", "", 10.5, "", "")

	if err != nil {
		t.Fatalf("UpdateQualityCheck failed: %v", err)
	}
}

func TestClient_BatchUpdateQualityChecks_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/events/1/quality-checks/batch" {
			t.Errorf("Expected path '/api/events/1/quality-checks/batch', got '%s'", r.URL.Path)
		}

		if r.Method != "PUT" {
			t.Errorf("Expected PUT method, got '%s'", r.Method)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	checks := []QualityCheckUpdate{
		{ID: 1, CheckStatus: "passed"},
		{ID: 2, CheckStatus: "failed", ErrorMessage: "test error"},
	}

	err := client.BatchUpdateQualityChecks(1, checks)

	if err != nil {
		t.Fatalf("BatchUpdateQualityChecks failed: %v", err)
	}
}

func TestClient_GetEvents_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	_, err := client.GetEvents()

	if err == nil {
		t.Error("Expected error for HTTP 500")
	}
}

func TestClient_GetEvent_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	_, err := client.GetEvent(999)

	if err == nil {
		t.Error("Expected error for HTTP 404")
	}
}

func TestClient_GetEvents_ConnectionError(t *testing.T) {
	client := NewClientWithURL("http://nonexistent-server:99999")
	_, err := client.GetEvents()

	if err == nil {
		t.Error("Expected error for connection failure")
	}
}

func TestClient_UpdateEventStatus_WithAPIToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Expected Authorization 'Bearer test-token', got '%s'", auth)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL)
	client.SetAPIToken("test-token")

	err := client.UpdateEventStatus(1, "completed", "")
	if err != nil {
		t.Fatalf("UpdateEventStatus failed: %v", err)
	}
}

func TestResourcePoolClient_NewClient(t *testing.T) {
	client := NewResourcePoolClient()

	if client == nil {
		t.Fatal("Expected client to be created")
	}

	if client.BaseURL != ResourcePoolAPI {
		t.Errorf("Expected BaseURL '%s', got '%s'", ResourcePoolAPI, client.BaseURL)
	}
}

func TestResourcePoolClient_GetCategories_Success(t *testing.T) {
	categories := []CategoryInfo{
		{UUID: "cat-1", Name: "Category 1", Enabled: true},
		{UUID: "cat-2", Name: "Category 2", Enabled: true},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/external/categories" {
			t.Errorf("Expected path '/external/categories', got '%s'", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(categories)
	}))
	defer server.Close()

	client := &ResourcePoolClient{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	}

	result, err := client.GetCategories()

	if err != nil {
		t.Fatalf("GetCategories failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(result))
	}
}

func TestResourcePoolClient_AcquireTestbed_Success(t *testing.T) {
	response := AcquireTestbedResponse{
		AllocationUUID: "alloc-123",
		Testbed: &TestbedInfo{
			UUID:        "tb-1",
			IPAddress:   "192.168.1.100",
			SSHUser:     "root",
			SSHPassword: "password",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/external/testbeds/acquire" {
			t.Errorf("Expected path '/external/testbeds/acquire', got '%s'", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Expected POST method, got '%s'", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &ResourcePoolClient{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	}

	result, err := client.AcquireTestbed("cat-1", "user1")

	if err != nil {
		t.Fatalf("AcquireTestbed failed: %v", err)
	}

	if result.AllocationUUID != "alloc-123" {
		t.Errorf("Expected AllocationUUID 'alloc-123', got '%s'", result.AllocationUUID)
	}

	if result.Testbed.IPAddress != "192.168.1.100" {
		t.Errorf("Expected IPAddress '192.168.1.100', got '%s'", result.Testbed.IPAddress)
	}
}

func TestResourcePoolClient_AcquireRobotTestbed_Success(t *testing.T) {
	response := struct {
		Success    bool           `json:"success"`
		Allocation AllocationInfo `json:"allocation"`
		Testbed    *TestbedInfo   `json:"testbed"`
	}{
		Success:    true,
		Allocation: AllocationInfo{UUID: "alloc-robot"},
		Testbed: &TestbedInfo{
			UUID:        "tb-robot",
			IPAddress:   "192.168.1.200",
			SSHUser:     "root",
			SSHPassword: "robot-pass",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/testbeds/acquire-robot" {
			t.Errorf("Expected path '/internal/testbeds/acquire-robot', got '%s'", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Expected POST method, got '%s'", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &ResourcePoolClient{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	}

	result, err := client.AcquireRobotTestbed()

	if err != nil {
		t.Fatalf("AcquireRobotTestbed failed: %v", err)
	}

	if result.AllocationUUID != "alloc-robot" {
		t.Errorf("Expected AllocationUUID 'alloc-robot', got '%s'", result.AllocationUUID)
	}

	if result.Testbed.IPAddress != "192.168.1.200" {
		t.Errorf("Expected IPAddress '192.168.1.200', got '%s'", result.Testbed.IPAddress)
	}
}
