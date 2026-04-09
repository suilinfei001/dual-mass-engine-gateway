// Package azure provides Azure DevOps client for executor service.
package azure

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

const (
	// Azure DevOps API 版本
	apiVersion = "7.0"
)

// AzureClient Azure DevOps 客户端
type AzureClient struct {
	auth      *AzureAuthConfig
	client    *http.Client
	logger    *logger.Logger
}

// NewAzureClient 创建 Azure DevOps 客户端
func NewAzureClient(auth *AzureAuthConfig, log *logger.Logger) *AzureClient {
	return &AzureClient{
		auth: auth,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		logger: log,
	}
}

// RunPipeline 运行 Pipeline
func (c *AzureClient) RunPipeline(ctx context.Context, req *PipelineRunRequest) (*PipelineRunResponse, error) {
	// 构建请求 URL
	url := fmt.Sprintf("%s/%s/_apis/pipelines/%s/runs?api-version=%s",
		c.auth.OrganizationURL,
		req.Project,
		fmt.Sprint(req.PipelineID),
		apiVersion,
	)

	// 构建请求体
	requestBody := map[string]interface{}{
		"resources": map[string]interface{}{
			"repositories": map[string]interface{}{
				"self": map[string]interface{}{
					"refName": fmt.Sprintf("refs/heads/%s", req.SourceBranch),
				},
			},
		},
		"variables": map[string]interface{}{},
	}

	// 添加 commit SHA
	if req.CommitSHA != "" {
		requestBody["resources"].(map[string]interface{})["repositories"].(map[string]interface{})["self"].(map[string]interface{})["version"] = req.CommitSHA
	}

	// 添加自定义参数
	for key, value := range req.Parameters {
		requestBody["variables"].(map[string]interface{})[key] = map[string]interface{}{
			"isSecret": false,
			"value":    value,
		}
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	c.logger.Info("Running Azure DevOps pipeline",
		logger.String("url", url),
		logger.String("project", req.Project),
		logger.Int("pipeline_id", req.PipelineID),
		logger.String("branch", req.SourceBranch),
	)

	// 创建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Basic %s", c.getBasicAuth()))

	// 发送请求
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var response struct {
		RunID    int64  `json:"id"`
		RunURL   string `json:"url"`
		State    string `json:"state"`
		Status   string `json:"status"`
		QueuedAt string `json:"createdDate"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	queuedAt, _ := time.Parse(time.RFC3339, response.QueuedAt)

	return &PipelineRunResponse{
		RunID:    response.RunID,
		RunURL:   response.RunURL,
		Status:   PipelineStatus(response.Status),
		QueuedAt: queuedAt,
	}, nil
}

// GetPipelineStatus 获取 Pipeline 状态
func (c *AzureClient) GetPipelineStatus(ctx context.Context, organization, project string, runID int64) (*PipelineStatusResponse, error) {
	url := fmt.Sprintf("%s/%s/_apis/pipelines/runs/%s?api-version=%s",
		c.auth.OrganizationURL,
		project,
		fmt.Sprint(runID),
		apiVersion,
	)

	c.logger.Debug("Getting pipeline status",
		logger.String("url", url),
		logger.Int64("run_id", runID),
	)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Basic %s", c.getBasicAuth()))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		RunID       int64  `json:"id"`
		State       string `json:"state"`
		Status      string `json:"status"`
		Result      string `json:"result"`
		QueuedAt    string `json:"createdDate"`
		StartedAt   string `json:"startedDate,omitempty"`
		CompletedAt string `json:"finishedDate,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	statusResp := &PipelineStatusResponse{
		RunID:  response.RunID,
		Status: PipelineStatus(response.Status),
		Result: PipelineResult(response.Result),
	}

	queuedAt, _ := time.Parse(time.RFC3339, response.QueuedAt)
	statusResp.QueuedAt = queuedAt

	if response.StartedAt != "" {
		startedAt, _ := time.Parse(time.RFC3339, response.StartedAt)
		statusResp.StartedAt = &startedAt
	}

	if response.CompletedAt != "" {
		completedAt, _ := time.Parse(time.RFC3339, response.CompletedAt)
		statusResp.CompletedAt = &completedAt
	}

	statusResp.Finished = statusResp.Status == PipelineStatusCompleted ||
		statusResp.Status == PipelineStatusFailed ||
		statusResp.Status == PipelineStatusCanceled

	return statusResp, nil
}

// GetPipelineLogs 获取 Pipeline 日志
func (c *AzureClient) GetPipelineLogs(ctx context.Context, organization, project string, runID int64) ([]string, error) {
	url := fmt.Sprintf("%s/%s/_apis/pipelines/runs/%s/logs?api-version=%s",
		c.auth.OrganizationURL,
		project,
		fmt.Sprint(runID),
		apiVersion,
	)

	c.logger.Debug("Getting pipeline logs",
		logger.String("url", url),
		logger.Int64("run_id", runID),
	)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Basic %s", c.getBasicAuth()))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	// 解析日志列表
	var logList []struct {
		LogID int64  `json:"id"`
		URL   string `json:"url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&logList); err != nil {
		return nil, fmt.Errorf("failed to decode log list: %w", err)
	}

	// 获取每个日志内容
	var logs []string
	for _, logRef := range logList {
		logContent, err := c.getLogContent(ctx, logRef.URL)
		if err != nil {
			c.logger.Warn("Failed to get log content",
				logger.Int64("log_id", logRef.LogID),
				logger.Err(err),
			)
			continue
		}
		logs = append(logs, logContent)
	}

	return logs, nil
}

// getLogContent 获取单个日志内容
func (c *AzureClient) getLogContent(ctx context.Context, logURL string) (string, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", logURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Basic %s", c.getBasicAuth()))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(content), nil
}

// CancelPipeline 取消 Pipeline
func (c *AzureClient) CancelPipeline(ctx context.Context, organization, project string, runID int64) error {
	url := fmt.Sprintf("%s/%s/_apis/pipelines/runs/%s?api-version=%s",
		c.auth.OrganizationURL,
		project,
		fmt.Sprint(runID),
		apiVersion,
	)

	c.logger.Info("Canceling pipeline",
		logger.String("project", project),
		logger.Int64("run_id", runID),
	)

	// PATCH 请求设置状态为 canceling
	requestBody := map[string]string{
		"status": "canceling",
	}
	jsonBody, _ := json.Marshal(requestBody)

	httpReq, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Basic %s", c.getBasicAuth()))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// getBasicAuth 获取 Basic Auth 头
func (c *AzureClient) getBasicAuth() string {
	// Azure DevOps 使用 PAT 作为 Basic Auth 的密码
	// 格式: BASE64(":PAT")
	return ":" + c.auth.PAT
}
