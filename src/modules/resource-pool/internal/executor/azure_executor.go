package executor

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AzureConfig 配置 Azure DevOps 连接
type AzureConfig struct {
	Organization string
	Project      string
	PAT          string
	BaseURL      string // 默认: https://devops.aishu.cn
}

// AzureDeployExecutor Azure DevOps 部署执行器
type AzureDeployExecutor struct {
	config *AzureConfig
	client *http.Client
}

// NewAzureDeployExecutor 创建 Azure 部署执行器
func NewAzureDeployExecutor(config *AzureConfig) *AzureDeployExecutor {
	if config.BaseURL == "" {
		config.BaseURL = "https://devops.aishu.cn"
	}
	return &AzureDeployExecutor{
		config: config,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// createAuthHeader 创建基本认证头
func (e *AzureDeployExecutor) createAuthHeader() string {
	authString := fmt.Sprintf(":%s", e.config.PAT)
	return base64.StdEncoding.EncodeToString([]byte(authString))
}

// doRequest 发送 HTTP 请求
func (e *AzureDeployExecutor) doRequest(method, url string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", e.createAuthHeader()))

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// PipelineRunResult Pipeline 运行结果
type PipelineRunResult struct {
	BuildID     int
	BuildNumber string
	WebURL      string
	Status      string
}

// PipelineStatus Pipeline 状态
type PipelineStatus struct {
	BuildID     int
	BuildNumber string
	Status      string // inProgress, completed, canceled, failed, etc.
	Result      string // succeeded, failed, canceled, etc.
	FinishTime  string
}

// RunPipeline 运行 Azure Pipeline
func (e *AzureDeployExecutor) RunPipeline(ctx context.Context, pipelineID int, params map[string]interface{}, branch string) (*PipelineRunResult, error) {
	if branch == "" {
		branch = "refs/heads/test"
	}

	url := fmt.Sprintf("%s/%s/%s/_apis/pipelines/%d/runs?api-version=6.0-preview.1",
		e.config.BaseURL, e.config.Organization, e.config.Project, pipelineID)

	payload := map[string]interface{}{
		"resources": map[string]interface{}{
			"repositories": map[string]interface{}{
				"self": map[string]interface{}{
					"refName": branch,
				},
			},
		},
	}

	if len(params) > 0 {
		payload["templateParameters"] = params
	}

	respBody, err := e.doRequest("POST", url, payload)
	if err != nil {
		return nil, err
	}

	var result struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		State string `json:"state"`
		Links struct {
			Web struct {
				Href string `json:"href"`
			} `json:"web"`
		} `json:"_links"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &PipelineRunResult{
		BuildID:     result.ID,
		BuildNumber: result.Name,
		WebURL:      result.Links.Web.Href,
		Status:      result.State,
	}, nil
}

// GetStatus 获取 Pipeline 运行状态
func (e *AzureDeployExecutor) GetStatus(ctx context.Context, buildID int) (*PipelineStatus, error) {
	url := fmt.Sprintf("%s/%s/%s/_apis/build/builds/%d?api-version=6.0",
		e.config.BaseURL, e.config.Organization, e.config.Project, buildID)

	respBody, err := e.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		ID          int    `json:"id"`
		BuildNumber string `json:"buildNumber"`
		Status      string `json:"status"`
		Result      string `json:"result"`
		FinishTime  string `json:"finishTime"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &PipelineStatus{
		BuildID:     result.ID,
		BuildNumber: result.BuildNumber,
		Status:      result.Status,
		Result:      result.Result,
		FinishTime:  result.FinishTime,
	}, nil
}

// LogEntry 日志条目
type LogEntry struct {
	LogID     int
	LineCount int
}

// GetLogList 获取 Pipeline 的日志列表
func (e *AzureDeployExecutor) GetLogList(ctx context.Context, buildID int) ([]LogEntry, error) {
	url := fmt.Sprintf("%s/%s/%s/_apis/build/builds/%d/logs?api-version=6.0",
		e.config.BaseURL, e.config.Organization, e.config.Project, buildID)

	respBody, err := e.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Value []struct {
			ID        int `json:"id"`
			LineCount int `json:"lineCount"`
		} `json:"value"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	entries := make([]LogEntry, len(result.Value))
	for i, v := range result.Value {
		entries[i] = LogEntry{
			LogID:     v.ID,
			LineCount: v.LineCount,
		}
	}

	return entries, nil
}

// GetLogContent 获取指定日志的内容
func (e *AzureDeployExecutor) GetLogContent(ctx context.Context, buildID int, logID int) (string, error) {
	url := fmt.Sprintf("%s/%s/%s/_apis/build/builds/%d/logs/%d?api-version=6.0",
		e.config.BaseURL, e.config.Organization, e.config.Project, buildID, logID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "text/plain")
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", e.createAuthHeader()))

	resp, err := e.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(content), nil
}

// IsCompleted 判断 Pipeline 是否已完成
func (e *AzureDeployExecutor) IsCompleted(status string) bool {
	return status == "completed" || status == "canceled" || status == "failed"
}

// IsSuccess 判断 Pipeline 是否成功
func (e *AzureDeployExecutor) IsSuccess(result string) bool {
	return result == "succeeded"
}

// TimelineRecord 时间线记录
type TimelineRecord struct {
	Name   string
	Type   string
	State  string
	Result string
	LogID  *int
}

// Timeline 时间线结果
type Timeline struct {
	Records []TimelineRecord
}

// GetTimeline 获取 Pipeline 时间线
func (e *AzureDeployExecutor) GetTimeline(ctx context.Context, buildID int) (*Timeline, error) {
	url := fmt.Sprintf("%s/%s/%s/_apis/build/builds/%d/timeline?api-version=6.0",
		e.config.BaseURL, e.config.Organization, e.config.Project, buildID)

	respBody, err := e.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Records []struct {
			Name   string `json:"name"`
			Type   string `json:"type"`
			State  string `json:"state"`
			Result string `json:"result"`
			LogID  *int   `json:"logId,omitempty"`
		} `json:"records"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	records := make([]TimelineRecord, 0, len(result.Records))
	for _, r := range result.Records {
		records = append(records, TimelineRecord{
			Name:   r.Name,
			Type:   r.Type,
			State:  r.State,
			Result: r.Result,
			LogID:  r.LogID,
		})
	}

	return &Timeline{Records: records}, nil
}

// Cancel 取消 Pipeline 运行
func (e *AzureDeployExecutor) Cancel(ctx context.Context, buildID int) error {
	url := fmt.Sprintf("%s/%s/%s/_apis/build/builds/%d?api-version=6.0",
		e.config.BaseURL, e.config.Organization, e.config.Project, buildID)

	payload := map[string]interface{}{
		"status": "cancelling",
	}

	_, err := e.doRequest("PATCH", url, payload)
	return err
}
