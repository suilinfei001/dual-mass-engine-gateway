// Package client provides Event Store client for webhook gateway.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/quality-gateway/shared/pkg/logger"
)

// EventStoreClient Event Store 服务客户端
type EventStoreClient struct {
	baseURL    string
	httpClient *http.Client
	apiToken   string
	logger     *logger.Logger
}

// NewEventStoreClient 创建 Event Store 客户端
func NewEventStoreClient(baseURL string, apiToken string, log *logger.Logger) *EventStoreClient {
	return &EventStoreClient{
		baseURL: baseURL,
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: log,
	}
}

// CreateEvent 创建事件
func (c *EventStoreClient) CreateEvent(ctx context.Context, event *ForwardedEvent) (*WebhookDeliveryResult, error) {
	// 准备请求体
	body, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}

	// 创建 HTTP 请求
	url := fmt.Sprintf("%s/api/events", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	if c.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiToken)
	}

	c.logger.Debug("Sending event to Event Store",
		logger.String("url", url),
		logger.String("event_uuid", event.UUID),
		logger.String("event_type", string(event.EventType)),
	)

	// 发送请求
	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &WebhookDeliveryResult{
			Success:      false,
			EventUUID:    event.UUID,
			ErrorMessage: err.Error(),
			DeliveredAt:  time.Now(),
		}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)
	respBody, _ := io.ReadAll(resp.Body)

	c.logger.Debug("Event Store response",
		logger.Int("status_code", resp.StatusCode),
		logger.Any("duration_ms", duration.Milliseconds()),
	)

	result := &WebhookDeliveryResult{
		EventUUID:   event.UUID,
		StatusCode:  resp.StatusCode,
		DeliveredAt: time.Now(),
	}

	// 检查响应状态
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Success = true
		return result, nil
	}

	result.Success = false
	result.ErrorMessage = string(respBody)
	return result, fmt.Errorf("event store returned status %d: %s", resp.StatusCode, string(respBody))
}

// UpdateEventStatus 更新事件状态
func (c *EventStoreClient) UpdateEventStatus(ctx context.Context, eventUUID string, status string) error {
	body := map[string]string{"status": status}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/events/%s/status", c.baseURL, eventUUID)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	respBody, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("event store returned status %d: %s", resp.StatusCode, string(respBody))
}

// CheckHealth 检查 Event Store 健康状态
func (c *EventStoreClient) CheckHealth(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return fmt.Errorf("event store unhealthy: status %d", resp.StatusCode)
}
