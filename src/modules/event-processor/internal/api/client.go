package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	// DefaultEventReceiverAPI 默认的 Event Receiver API 地址
	DefaultEventReceiverAPI = ""
)

// Client Event Receiver API 客户端
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	apiToken   string // API Token 用于认证
}

// NewClient 创建新的 API 客户端（使用默认地址）
func NewClient() *Client {
	return NewClientWithURL(DefaultEventReceiverAPI)
}

// NewClientWithURL 使用指定地址创建 API 客户端
func NewClientWithURL(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiToken: os.Getenv("EVENT_RECEIVER_API_TOKEN"),
	}
}

// SetAPIToken 设置 API Token
func (c *Client) SetAPIToken(token string) {
	c.apiToken = token
}

// Event 事件数据结构（来自 Event Receiver）
type Event struct {
	ID            int                    `json:"id"`
	EventID       string                 `json:"event_id"`
	EventType     string                 `json:"event_type"`
	EventStatus   string                 `json:"event_status"`
	Repository    string                 `json:"repository"`
	Branch        string                 `json:"branch"`
	CommitSHA     string                 `json:"commit_sha"`
	PRNumber      *int                   `json:"pr_number,omitempty"`
	TargetBranch  string                 `json:"target_branch,omitempty"`
	Payload       map[string]interface{} `json:"payload"`
	Pusher        string                 `json:"pusher,omitempty"`
	Author        string                 `json:"author,omitempty"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
	ProcessedAt   string                 `json:"processed_at,omitempty"`
	QualityChecks []QualityCheck         `json:"quality_checks"`
}

// QualityCheck 质量检查项
type QualityCheck struct {
	ID             int     `json:"id"`
	GitHubEventID string  `json:"github_event_id"`
	CheckType      string  `json:"check_type"`
	CheckStatus    string  `json:"check_status"`
	Stage          string  `json:"stage"`
	StageOrder     int     `json:"stage_order"`
	CheckOrder     int     `json:"check_order"`
	StartedAt      string  `json:"started_at,omitempty"`
	CompletedAt    string  `json:"completed_at,omitempty"`
	DurationSeconds *float64 `json:"duration_seconds,omitempty"`
	ErrorMessage  *string  `json:"error_message,omitempty"`
	Output         *string  `json:"output,omitempty"`
	Extra          *string `json:"extra,omitempty"` // 额外信息（JSON字符串）
	RetryCount     int     `json:"retry_count"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// APIResponse API 响应结构
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// GetEvents 获取所有事件
func (c *Client) GetEvents() ([]Event, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/events")
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 解析 events 数组
	var events []Event
	if dataBytes, err := json.Marshal(apiResp.Data); err == nil {
		json.Unmarshal(dataBytes, &events)
	}

	return events, nil
}

// GetEvent 获取单个事件详情
func (c *Client) GetEvent(id int) (*Event, error) {
	resp, err := c.HTTPClient.Get(fmt.Sprintf("%s/api/events/%d", c.BaseURL, id))
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 解析 event 对象
	var event Event
	if dataBytes, err := json.Marshal(apiResp.Data); err == nil {
		json.Unmarshal(dataBytes, &event)
	}

	return &event, nil
}

// UpdateEventStatus 更新事件状态
func (c *Client) UpdateEventStatus(eventID int, status string, processedAt string) error {
	payload := map[string]interface{}{
		"event_status": status,
	}
	if processedAt != "" {
		payload["processed_at"] = processedAt
	}

	return c.putJSON(fmt.Sprintf("/api/events/%d/status", eventID), payload)
}

// UpdateQualityCheck 更新单个质量检查状态
func (c *Client) UpdateQualityCheck(checkID int, status string, startedAt, completedAt string, duration float64, errorMessage, output string) error {
	payload := map[string]interface{}{
		"check_status": status,
	}
	if startedAt != "" {
		payload["started_at"] = startedAt
	}
	if completedAt != "" {
		payload["completed_at"] = completedAt
	}
	if duration > 0 {
		payload["duration_seconds"] = duration
	}
	if errorMessage != "" {
		payload["error_message"] = errorMessage
	}
	if output != "" {
		payload["output"] = output
	}

	return c.putJSON(fmt.Sprintf("/api/quality-checks/%d", checkID), payload)
}

// BatchUpdateQualityChecks 批量更新质量检查状态
func (c *Client) BatchUpdateQualityChecks(eventID int, checks []QualityCheckUpdate) error {
	payload := map[string]interface{}{
		"quality_checks": checks,
	}

	return c.putJSON(fmt.Sprintf("/api/events/%d/quality-checks/batch", eventID), payload)
}

// QualityCheckUpdate 质量检查更新
type QualityCheckUpdate struct {
	ID           int     `json:"id"`
	CheckStatus   string  `json:"check_status,omitempty"`
	StartedAt     string  `json:"started_at,omitempty"`
	CompletedAt    string  `json:"completed_at,omitempty"`
	Duration      float64 `json:"duration_seconds,omitempty"`
	ErrorMessage  string  `json:"error_message,omitempty"`
	Output        string  `json:"output,omitempty"`
	Extra         string  `json:"extra,omitempty"` // 额外信息（JSON字符串）
}

// putJSON 发送 PUT 请求
func (c *Client) putJSON(path string, payload interface{}) error {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("PUT", c.BaseURL+path, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 添加 API Token 认证
	if c.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiToken)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
