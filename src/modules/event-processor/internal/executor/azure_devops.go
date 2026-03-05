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

type AzureDevOpsExecutor struct {
	config *ExecutorConfig
	client *http.Client
}

func NewAzureDevOpsExecutor(config *ExecutorConfig) *AzureDevOpsExecutor {
	if config.Branch == "" {
		config.Branch = "refs/heads/main"
	}
	return &AzureDevOpsExecutor{
		config: config,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (e *AzureDevOpsExecutor) createAuthHeader() string {
	authString := fmt.Sprintf(":%s", e.config.PAT)
	return base64.StdEncoding.EncodeToString([]byte(authString))
}

func (e *AzureDevOpsExecutor) doRequest(method, url string, body interface{}) ([]byte, error) {
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

func (e *AzureDevOpsExecutor) Run(ctx context.Context, pipelineID int, params map[string]interface{}) (*RunResult, error) {
	url := fmt.Sprintf("https://devops.aishu.cn/%s/%s/_apis/pipelines/%d/runs?api-version=6.0-preview.1",
		e.config.Organization, e.config.Project, pipelineID)

	payload := map[string]interface{}{
		"resources": map[string]interface{}{
			"repositories": map[string]interface{}{
				"self": map[string]interface{}{
					"refName": e.config.Branch,
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
		ID       int    `json:"id"`
		Name     string `json:"name"`
		State    string `json:"state"`
		Pipeline struct {
			Name string `json:"name"`
		} `json:"pipeline"`
		Links struct {
			Web struct {
				Href string `json:"href"`
			} `json:"web"`
		} `json:"_links"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &RunResult{
		BuildID:     result.ID,
		BuildNumber: result.Name,
		Status:      TaskStatus(result.State),
		WebURL:      result.Links.Web.Href,
	}, nil
}

func (e *AzureDevOpsExecutor) Cancel(ctx context.Context, buildID int) error {
	url := fmt.Sprintf("https://devops.aishu.cn/%s/%s/_apis/build/builds/%d?api-version=6.0",
		e.config.Organization, e.config.Project, buildID)

	payload := map[string]interface{}{
		"status": "cancelling",
	}

	_, err := e.doRequest("PATCH", url, payload)
	return err
}

func (e *AzureDevOpsExecutor) GetStatus(ctx context.Context, buildID int) (*StatusResult, error) {
	url := fmt.Sprintf("https://devops.aishu.cn/%s/%s/_apis/build/builds/%d?api-version=6.0",
		e.config.Organization, e.config.Project, buildID)

	respBody, err := e.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		ID         int    `json:"id"`
		BuildNumber string `json:"buildNumber"`
		Status     string `json:"status"`
		Result     string `json:"result"`
		FinishTime string `json:"finishTime"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &StatusResult{
		BuildID:     result.ID,
		BuildNumber: result.BuildNumber,
		Status:      TaskStatus(result.Status),
		Result:      TaskResult(result.Result),
		FinishTime:  result.FinishTime,
	}, nil
}

func (e *AzureDevOpsExecutor) GetLogs(ctx context.Context, buildID int, logID int) (*LogResult, error) {
	url := fmt.Sprintf("https://devops.aishu.cn/%s/%s/_apis/build/builds/%d/logs/%d?api-version=6.0",
		e.config.Organization, e.config.Project, buildID, logID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "text/plain")
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", e.createAuthHeader()))

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return &LogResult{
		LogID:   logID,
		Content: string(content),
	}, nil
}

func (e *AzureDevOpsExecutor) GetLogList(ctx context.Context, buildID int) ([]LogResult, error) {
	url := fmt.Sprintf("https://devops.aishu.cn/%s/%s/_apis/build/builds/%d/logs?api-version=6.0",
		e.config.Organization, e.config.Project, buildID)

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

	logs := make([]LogResult, len(result.Value))
	for i, v := range result.Value {
		logs[i] = LogResult{
			LogID:     v.ID,
			LineCount: v.LineCount,
		}
	}

	return logs, nil
}

func (e *AzureDevOpsExecutor) GetTimeline(ctx context.Context, buildID int) (*TimelineResult, error) {
	url := fmt.Sprintf("https://devops.aishu.cn/%s/%s/_apis/build/builds/%d/timeline?api-version=6.0",
		e.config.Organization, e.config.Project, buildID)

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
			LogID  *int   `json:"logId,omitempty"` // Use pointer to distinguish 0 from nil
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
			Result: TaskResult(r.Result),
			LogID:  r.LogID,
		})
	}

	return &TimelineResult{Records: records}, nil
}

func (e *AzureDevOpsExecutor) IsCompleted(status TaskStatus) bool {
	return status == TaskStatusCompleted || status == TaskStatusCanceled || status == TaskStatusFailed
}

func (e *AzureDevOpsExecutor) IsSuccess(result TaskResult) bool {
	return result == TaskResultSucceeded
}
