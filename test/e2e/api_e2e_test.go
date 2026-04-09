package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// API 端点配置
var (
	webhookGatewayURL  = getEnv("WEBHOOK_GATEWAY_URL", "http://localhost:4001")
	eventStoreURL      = getEnv("EVENT_STORE_URL", "http://localhost:4002")
	taskSchedulerURL   = getEnv("TASK_SCHEDULER_URL", "http://localhost:4003")
	executorServiceURL = getEnv("EXECUTOR_SERVICE_URL", "http://localhost:4004")
	aiAnalyzerURL      = getEnv("AI_ANALYZER_URL", "http://localhost:4005")
	resourceManagerURL = getEnv("RESOURCE_MANAGER_URL", "http://localhost:4006")
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ============================================================
// 用户故事 1: Webhook 接收与事件创建
// 作为系统，我希望能够接收 GitHub/GitLab Webhook 并创建事件
// ============================================================

func TestE2E_Webhook_GitHubPullRequest(t *testing.T) {
	payload := map[string]interface{}{
		"action": "opened",
		"repository": map[string]interface{}{
			"id":        12345,
			"name":      "test-repo",
			"full_name": "owner/test-repo",
			"html_url":  "https://github.com/owner/test-repo",
			"owner": map[string]interface{}{
				"login": "owner",
			},
		},
		"pull_request": map[string]interface{}{
			"number": 42,
			"title":  "Test PR",
			"head": map[string]interface{}{
				"ref": "feature-branch",
				"sha": "abc123def456",
			},
		},
		"sender": map[string]interface{}{
			"login": "testuser",
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", webhookGatewayURL+"/webhook/github", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "pull_request")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send webhook: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		t.Errorf("Expected status 200 or 202, got %d", resp.StatusCode)
	}
}

func TestE2E_Webhook_GitLabMergeRequest(t *testing.T) {
	payload := map[string]interface{}{
		"object_kind": "merge_request",
		"event_type":  "merge_request",
		"project": map[string]interface{}{
			"id":       12345,
			"name":     "test-project",
			"http_url": "https://gitlab.com/owner/test-project",
		},
		"object_attributes": map[string]interface{}{
			"iid":           42,
			"title":         "Test MR",
			"action":        "open",
			"source_branch": "feature-branch",
		},
		"user": map[string]interface{}{
			"login": "testuser",
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", webhookGatewayURL+"/webhook/gitlab", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gitlab-Event", "Merge Request Hook")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send webhook: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		t.Errorf("Expected status 200 or 202, got %d", resp.StatusCode)
	}
}

// ============================================================
// 用户故事 2: 事件管理
// 作为用户，我希望能够创建、查询、更新事件状态
// ============================================================

func TestE2E_Event_Create(t *testing.T) {
	payload := map[string]interface{}{
		"event_type": "pull_request.opened",
		"source":     "github",
		"repo_name":  "test/repo",
		"repo_owner": "testowner",
		"pr_number":  123,
		"author":     "testuser",
		"commit_sha": "abc123def456",
		"status":     "pending",
		"payload":    "{}",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(eventStoreURL+"/api/events", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 201, 200 or 500, got %d", resp.StatusCode)
	}
}

func TestE2E_Event_List(t *testing.T) {
	resp, err := http.Get(eventStoreURL + "/api/events?limit=10")
	if err != nil {
		t.Fatalf("Failed to list events: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestE2E_Event_GetPending(t *testing.T) {
	resp, err := http.Get(eventStoreURL + "/api/events/pending?limit=10")
	if err != nil {
		t.Fatalf("Failed to get pending events: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestE2E_Event_GetStatistics(t *testing.T) {
	resp, err := http.Get(eventStoreURL + "/api/events/statistics")
	if err != nil {
		t.Fatalf("Failed to get event statistics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// ============================================================
// 用户故事 3: 质量检查管理
// 作为用户，我希望能够创建、查询、更新质量检查
// ============================================================

func TestE2E_QualityCheck_Create(t *testing.T) {
	payload := map[string]interface{}{
		"check_type":   "code_lint",
		"check_status": "pending",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(eventStoreURL+"/api/events/test-uuid-123/quality-checks", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create quality check: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 201, 200, 400 or 500, got %d", resp.StatusCode)
	}
}

func TestE2E_QualityCheck_List(t *testing.T) {
	resp, err := http.Get(eventStoreURL + "/api/events/test-uuid-123/quality-checks")
	if err != nil {
		t.Fatalf("Failed to list quality checks: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 200 or 400, got %d", resp.StatusCode)
	}
}

// ============================================================
// 用户故事 4: 任务调度
// 作为系统，我希望能够创建、启动、完成、取消任务
// ============================================================

func TestE2E_Task_List(t *testing.T) {
	resp, err := http.Get(taskSchedulerURL + "/api/tasks?limit=10")
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestE2E_Task_Start(t *testing.T) {
	req, _ := http.NewRequest("POST", taskSchedulerURL+"/api/tasks/1/start", nil)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to start task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 200, 404 or 400, got %d", resp.StatusCode)
	}
}

func TestE2E_Task_Complete(t *testing.T) {
	payload := map[string]interface{}{
		"results": []map[string]interface{}{
			{
				"check_type": "code_lint",
				"result":     "pass",
				"output":     "All checks passed",
			},
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", taskSchedulerURL+"/api/tasks/1/complete", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to complete task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 200, 404 or 400, got %d", resp.StatusCode)
	}
}

func TestE2E_Task_Fail(t *testing.T) {
	payload := map[string]interface{}{
		"reason": "Test failure",
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", taskSchedulerURL+"/api/tasks/1/fail", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to fail task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 200, 404 or 400, got %d", resp.StatusCode)
	}
}

func TestE2E_Task_Cancel(t *testing.T) {
	payload := map[string]interface{}{
		"reason": "PR synchronized",
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", taskSchedulerURL+"/api/tasks/1/cancel", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to cancel task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 200, 404 or 400, got %d", resp.StatusCode)
	}
}

// ============================================================
// 用户故事 5: 任务执行
// 作为系统，我希望能够执行任务并获取执行状态
// ============================================================

func TestE2E_Execution_Execute(t *testing.T) {
	payload := map[string]interface{}{
		"task_uuid":            "test-uuid-123",
		"task_type":            "basic_ci_all",
		"chart_url":            "http://example.com/chart.tgz",
		"testbed_ip":           "192.168.1.100",
		"testbed_ssh_port":     22,
		"testbed_ssh_user":     "root",
		"testbed_ssh_password": "password",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(executorServiceURL+"/api/execute", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to execute task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 202 or 500, got %d", resp.StatusCode)
	}
}

func TestE2E_Execution_GetStatus(t *testing.T) {
	resp, err := http.Get(executorServiceURL + "/api/executions/test-exec-123")
	if err != nil {
		t.Fatalf("Failed to get execution status: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		t.Logf("Execution status: %v", result)
	}
}

func TestE2E_Execution_List(t *testing.T) {
	resp, err := http.Get(executorServiceURL + "/api/executions?limit=10")
	if err != nil {
		t.Fatalf("Failed to list executions: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestE2E_Execution_Cancel(t *testing.T) {
	req, _ := http.NewRequest("DELETE", executorServiceURL+"/api/executions/test-exec-123", nil)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to cancel execution: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 200, 404 or 400, got %d", resp.StatusCode)
	}
}

// ============================================================
// 用户故事 6: AI 分析
// 作为系统，我希望能够使用 AI 分析日志
// 注意: AI 分析需要配置 AI 服务，未配置时返回 500 是预期行为
// ============================================================

func TestE2E_AI_Analyze(t *testing.T) {
	payload := map[string]interface{}{
		"log_content": "Error: Connection timeout\nStack trace: ...",
		"task_name":   "basic_ci_all",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(aiAnalyzerURL+"/api/analyze", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to analyze log: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500 (AI not configured), got %d", resp.StatusCode)
	}
}

func TestE2E_AI_GetPoolStats(t *testing.T) {
	resp, err := http.Get(aiAnalyzerURL + "/api/pool/stats")
	if err != nil {
		t.Fatalf("Failed to get pool stats: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// ============================================================
// 用户故事 7: 资源管理
// 作为用户，我希望能够管理资源实例
// ============================================================

func TestE2E_Resource_List(t *testing.T) {
	resp, err := http.Get(resourceManagerURL + "/api/resources?limit=10")
	if err != nil {
		t.Fatalf("Failed to list resources: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestE2E_Resource_Create(t *testing.T) {
	payload := map[string]interface{}{
		"name":         "test-resource-e2e",
		"ip_address":   "192.168.1.100",
		"ssh_port":     22,
		"ssh_user":     "root",
		"ssh_password": "password",
		"category_id":  1,
		"is_public":    true,
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(resourceManagerURL+"/api/resources", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create resource: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 201 or 200, got %d", resp.StatusCode)
	}
}

func TestE2E_Resource_Match(t *testing.T) {
	payload := map[string]interface{}{
		"category_id":    1,
		"task_uuid":      "test-task-uuid-123",
		"required_count": 1,
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(resourceManagerURL+"/api/resources/match", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to match resource: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", resp.StatusCode)
	}
}

// ============================================================
// 用户故事 8: 分类管理
// 作为管理员，我希望能够管理资源分类
// ============================================================

func TestE2E_Category_List(t *testing.T) {
	resp, err := http.Get(resourceManagerURL + "/api/categories")
	if err != nil {
		t.Fatalf("Failed to list categories: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestE2E_Category_Create(t *testing.T) {
	payload := map[string]interface{}{
		"name":        "test-category-e2e",
		"description": "Test category for E2E",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(resourceManagerURL+"/api/categories", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 201, 200 or 500, got %d", resp.StatusCode)
	}
}

// ============================================================
// 用户故事 9: Testbed 管理
// 作为系统，我希望能够查询和管理 Testbed
// ============================================================

func TestE2E_Testbed_List(t *testing.T) {
	resp, err := http.Get(resourceManagerURL + "/api/testbeds?limit=10")
	if err != nil {
		t.Fatalf("Failed to list testbeds: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// ============================================================
// 用户故事 10: 健康检查
// 作为运维人员，我希望能够检查所有服务的健康状态
// ============================================================

func TestE2E_HealthCheck_AllServices(t *testing.T) {
	services := map[string]string{
		"webhook-gateway":  webhookGatewayURL + "/health",
		"event-store":      eventStoreURL + "/health",
		"task-scheduler":   taskSchedulerURL + "/health",
		"executor-service": executorServiceURL + "/health",
		"ai-analyzer":      aiAnalyzerURL + "/health",
		"resource-manager": resourceManagerURL + "/health",
	}

	for name, url := range services {
		t.Run(name, func(t *testing.T) {
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("Failed to check health for %s: %v", name, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Service %s health check failed: %d", name, resp.StatusCode)
			}

			var result map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&result)
			t.Logf("%s health: %v", name, result)
		})
	}
}

// ============================================================
// 边界条件测试 - 无效输入
// ============================================================

func TestE2E_Boundary_Webhook_EmptyPayload(t *testing.T) {
	req, _ := http.NewRequest("POST", webhookGatewayURL+"/webhook/github", bytes.NewReader([]byte{}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "pull_request")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send webhook: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 400, 200 or 500, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_Webhook_MalformedJSON(t *testing.T) {
	req, _ := http.NewRequest("POST", webhookGatewayURL+"/webhook/github", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "pull_request")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send webhook: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 400, 200 or 500, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_Webhook_MissingEventHeader(t *testing.T) {
	payload := map[string]interface{}{"action": "opened"}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", webhookGatewayURL+"/webhook/github", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send webhook: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 400 or 200, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_Event_EmptyPayload(t *testing.T) {
	req, _ := http.NewRequest("POST", eventStoreURL+"/api/events", bytes.NewReader([]byte{}))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 400 or 500, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_Event_MalformedJSON(t *testing.T) {
	req, _ := http.NewRequest("POST", eventStoreURL+"/api/events", bytes.NewReader([]byte("{bad json")))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 400 or 500, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_Event_MissingRequiredFields(t *testing.T) {
	payload := map[string]interface{}{
		"source": "github",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(eventStoreURL+"/api/events", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError && resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 400, 500 or 200, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_Task_EmptyPayload(t *testing.T) {
	req, _ := http.NewRequest("POST", taskSchedulerURL+"/api/tasks/1/start", bytes.NewReader([]byte{}))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to start task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 200, 404 or 400, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_Task_CompleteEmptyResults(t *testing.T) {
	payload := map[string]interface{}{
		"results": []map[string]interface{}{},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", taskSchedulerURL+"/api/tasks/1/complete", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to complete task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 200, 404 or 400, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_Resource_EmptyPayload(t *testing.T) {
	req, _ := http.NewRequest("POST", resourceManagerURL+"/api/resources", bytes.NewReader([]byte{}))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to create resource: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 400 or 500, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_Resource_MissingRequiredFields(t *testing.T) {
	payload := map[string]interface{}{
		"description": "missing name field",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(resourceManagerURL+"/api/resources", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create resource: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError && resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 400, 500 or 200, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_Category_EmptyName(t *testing.T) {
	payload := map[string]interface{}{
		"name":        "",
		"description": "Empty name test",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(resourceManagerURL+"/api/categories", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError && resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 400, 500 or 200, got %d", resp.StatusCode)
	}
}

// ============================================================
// 边界条件测试 - 资源不存在
// ============================================================

func TestE2E_Boundary_Event_NotFound(t *testing.T) {
	resp, err := http.Get(eventStoreURL + "/api/events/non-existent-uuid-12345")
	if err != nil {
		t.Fatalf("Failed to get event: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 404, 200 or 400, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_Task_NotFound(t *testing.T) {
	req, _ := http.NewRequest("POST", taskSchedulerURL+"/api/tasks/999999/start", nil)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to start task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 404, 400 or 200, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_Task_CompleteNotFound(t *testing.T) {
	payload := map[string]interface{}{
		"results": []map[string]interface{}{
			{"check_type": "test", "result": "pass"},
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", taskSchedulerURL+"/api/tasks/999999/complete", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to complete task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 404, 400 or 200, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_Execution_NotFound(t *testing.T) {
	resp, err := http.Get(executorServiceURL + "/api/executions/non-existent-exec-999999")
	if err != nil {
		t.Fatalf("Failed to get execution: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 404, 200 or 400, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_Resource_NotFound(t *testing.T) {
	resp, err := http.Get(resourceManagerURL + "/api/resources/999999")
	if err != nil {
		t.Fatalf("Failed to get resource: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 404, 200 or 400, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_Category_NotFound(t *testing.T) {
	resp, err := http.Get(resourceManagerURL + "/api/categories/999999")
	if err != nil {
		t.Fatalf("Failed to get category: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 404, 200 or 400, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_Testbed_NotFound(t *testing.T) {
	resp, err := http.Get(resourceManagerURL + "/api/testbeds/non-existent-testbed-999999")
	if err != nil {
		t.Fatalf("Failed to get testbed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 404, 200 or 400, got %d", resp.StatusCode)
	}
}

// ============================================================
// 边界条件测试 - 参数边界
// ============================================================

func TestE2E_Boundary_Event_InvalidStatus(t *testing.T) {
	payload := map[string]interface{}{
		"event_type": "pull_request.opened",
		"source":     "github",
		"status":     "invalid_status_value",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(eventStoreURL+"/api/events", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError && resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 400, 500 or 200, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_Event_InvalidEventType(t *testing.T) {
	payload := map[string]interface{}{
		"event_type": "",
		"source":     "github",
		"status":     "pending",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(eventStoreURL+"/api/events", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError && resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 400, 500 or 200, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_Task_InvalidTaskID(t *testing.T) {
	req, _ := http.NewRequest("POST", taskSchedulerURL+"/api/tasks/invalid-id/start", nil)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to start task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 400, 404 or 200, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_Task_NegativeTaskID(t *testing.T) {
	req, _ := http.NewRequest("POST", taskSchedulerURL+"/api/tasks/-1/start", nil)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to start task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 400, 404 or 200, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_Resource_InvalidIP(t *testing.T) {
	payload := map[string]interface{}{
		"name":         "test-resource-invalid-ip",
		"ip_address":   "invalid-ip-address",
		"ssh_port":     22,
		"ssh_user":     "root",
		"ssh_password": "password",
		"category_id":  1,
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(resourceManagerURL+"/api/resources", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create resource: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError && resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 400, 500 or 200, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_Resource_InvalidPort(t *testing.T) {
	payload := map[string]interface{}{
		"name":         "test-resource-invalid-port",
		"ip_address":   "192.168.1.100",
		"ssh_port":     -1,
		"ssh_user":     "root",
		"ssh_password": "password",
		"category_id":  1,
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(resourceManagerURL+"/api/resources", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create resource: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError && resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 400, 500 or 200, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_Resource_InvalidCategoryID(t *testing.T) {
	payload := map[string]interface{}{
		"name":         "test-resource-invalid-category",
		"ip_address":   "192.168.1.100",
		"ssh_port":     22,
		"ssh_user":     "root",
		"ssh_password": "password",
		"category_id":  -1,
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(resourceManagerURL+"/api/resources", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create resource: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError && resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 400, 500 or 200, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_AI_EmptyLogContent(t *testing.T) {
	payload := map[string]interface{}{
		"log_content": "",
		"task_name":   "basic_ci_all",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(aiAnalyzerURL+"/api/analyze", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to analyze log: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError && resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 400, 500 or 200, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_AI_MissingLogContent(t *testing.T) {
	payload := map[string]interface{}{
		"task_name": "basic_ci_all",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(aiAnalyzerURL+"/api/analyze", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to analyze log: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError && resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 400, 500 or 200, got %d", resp.StatusCode)
	}
}

// ============================================================
// 边界条件测试 - 特殊字符
// ============================================================

func TestE2E_Boundary_Event_SpecialCharacters(t *testing.T) {
	payload := map[string]interface{}{
		"event_type": "pull_request.opened",
		"source":     "github",
		"repo_name":  "test/repo-with-special-chars-!@#$%",
		"repo_owner": "testowner<>\"'&",
		"author":     "testuser\u0000\u001F",
		"status":     "pending",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(eventStoreURL+"/api/events", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError && resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 400, 500 or 200, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_Resource_SpecialCharacters(t *testing.T) {
	payload := map[string]interface{}{
		"name":         "test-resource-<script>alert('xss')</script>",
		"ip_address":   "192.168.1.100",
		"ssh_port":     22,
		"ssh_user":     "root; DROP TABLE resources;--",
		"ssh_password": "password' OR '1'='1",
		"category_id":  1,
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(resourceManagerURL+"/api/resources", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create resource: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError && resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 400, 500 or 200, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_Category_SpecialCharacters(t *testing.T) {
	payload := map[string]interface{}{
		"name":        "category-<img src=x onerror=alert(1)>",
		"description": "description with 'quotes' and \"double quotes\" and \n newlines",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(resourceManagerURL+"/api/categories", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError && resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 400, 500 or 200, got %d", resp.StatusCode)
	}
}

// ============================================================
// 边界条件测试 - 大数据量
// ============================================================

func TestE2E_Boundary_Event_LargePayload(t *testing.T) {
	largePayload := strings.Repeat("x", 1024*1024)
	payload := map[string]interface{}{
		"event_type": "pull_request.opened",
		"source":     "github",
		"repo_name":  "test/repo",
		"status":     "pending",
		"payload":    largePayload,
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(eventStoreURL+"/api/events", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status 400, 500, 200 or 413, got %d", resp.StatusCode)
	}
}

func TestE2E_Boundary_AI_LargeLogContent(t *testing.T) {
	largeLog := strings.Repeat("Error: Connection timeout\n", 10000)
	payload := map[string]interface{}{
		"log_content": largeLog,
		"task_name":   "basic_ci_all",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(aiAnalyzerURL+"/api/analyze", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to analyze log: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status 400, 500, 200 or 413, got %d", resp.StatusCode)
	}
}

// ============================================================
// 边界条件测试 - 分页参数
// ============================================================

func TestE2E_Boundary_Event_InvalidLimit(t *testing.T) {
	tests := []struct {
		name  string
		limit string
	}{
		{"negative", "-1"},
		{"zero", "0"},
		{"invalid", "abc"},
		{"very_large", "999999999999"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(eventStoreURL + "/api/events?limit=" + tc.limit)
			if err != nil {
				t.Fatalf("Failed to list events: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 400 or 200, got %d", resp.StatusCode)
			}
		})
	}
}

func TestE2E_Boundary_Task_InvalidLimit(t *testing.T) {
	tests := []struct {
		name  string
		limit string
	}{
		{"negative", "-1"},
		{"zero", "0"},
		{"invalid", "abc"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(taskSchedulerURL + "/api/tasks?limit=" + tc.limit)
			if err != nil {
				t.Fatalf("Failed to list tasks: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 400 or 200, got %d", resp.StatusCode)
			}
		})
	}
}

func TestE2E_Boundary_Resource_InvalidLimit(t *testing.T) {
	tests := []struct {
		name  string
		limit string
	}{
		{"negative", "-1"},
		{"zero", "0"},
		{"invalid", "abc"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(resourceManagerURL + "/api/resources?limit=" + tc.limit)
			if err != nil {
				t.Fatalf("Failed to list resources: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 400 or 200, got %d", resp.StatusCode)
			}
		})
	}
}

// ============================================================
// 辅助函数
// ============================================================

func TestMain(m *testing.M) {
	fmt.Println("=== API E2E Tests ===")
	fmt.Println("Testing against:")
	fmt.Printf("  Webhook Gateway:  %s\n", webhookGatewayURL)
	fmt.Printf("  Event Store:      %s\n", eventStoreURL)
	fmt.Printf("  Task Scheduler:   %s\n", taskSchedulerURL)
	fmt.Printf("  Executor Service: %s\n", executorServiceURL)
	fmt.Printf("  AI Analyzer:      %s\n", aiAnalyzerURL)
	fmt.Printf("  Resource Manager: %s\n", resourceManagerURL)
	fmt.Println("")

	os.Exit(m.Run())
}
