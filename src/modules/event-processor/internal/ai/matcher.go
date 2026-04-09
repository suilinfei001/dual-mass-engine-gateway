package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github-hub/event-processor/internal/models"
	"github-hub/event-processor/internal/prompt"
	"github-hub/event-processor/internal/storage"
)

var (
	ErrAINotConfigured = errors.New("AI_NOT_CONFIGURED")
)

type AIConfigProvider interface {
	GetAIConfig() (*storage.AIConfig, error)
}

type aiConfigAdapter struct {
	*storage.MySQLConfigStorage
}

func (a *aiConfigAdapter) GetAIConfig() (*storage.AIConfig, error) {
	return a.MySQLConfigStorage.GetAIConfig()
}

type AIClient struct {
	configStorage AIConfigProvider
	client        *http.Client
}

func NewAIClient(configStorage *storage.MySQLConfigStorage) *AIClient {
	return &AIClient{
		configStorage: &aiConfigAdapter{configStorage},
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func NewAIClientWithProvider(provider AIConfigProvider) *AIClient {
	return &AIClient{
		configStorage: provider,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// ChatRequest represents a generic chat completion request
type ChatRequest struct {
	SystemPrompt string
	UserPrompt   string
	Temperature  float64
	MaxTokens    int
}

// ChatResponse represents the response from a chat completion
type ChatResponse struct {
	Content string
	RawBody string
}

// Chat sends a chat completion request to the AI API
func (c *AIClient) Chat(req *ChatRequest) (*ChatResponse, error) {
	config, err := c.configStorage.GetAIConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get AI config: %w", err)
	}

	if config.IP == "" || config.Model == "" || config.Token == "" {
		return nil, ErrAINotConfigured
	}

	temperature := req.Temperature
	if temperature == 0 {
		temperature = 1.0
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 10000
	}

	requestBody := map[string]interface{}{
		"model":             config.Model,
		"temperature":       temperature,
		"top_p":             float64(1),
		"max_tokens":        maxTokens,
		"top_k":             int(1),
		"presence_penalty":  float64(0),
		"frequency_penalty": float64(0),
		"stream":            false,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": req.SystemPrompt,
			},
			{
				"role":    "user",
				"content": req.UserPrompt,
			},
		},
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	log.Printf("[AI Client] Request Body:\n%s", string(jsonData))

	url := fmt.Sprintf("http://%s/api/mf-model-api/v1/chat/completions", config.IP)

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+config.Token)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	log.Printf("[AI Client] AI Response: %s", result.Choices[0].Message.Content)

	return &ChatResponse{
		Content: result.Choices[0].Message.Content,
		RawBody: string(body),
	}, nil
}

type AIMatcher struct {
	aiClient *AIClient
}

func NewAIMatcher(configStorage *storage.MySQLConfigStorage) *AIMatcher {
	return &AIMatcher{
		aiClient: NewAIClient(configStorage),
	}
}

func NewAIMatcherWithProvider(provider AIConfigProvider) *AIMatcher {
	return &AIMatcher{
		aiClient: NewAIClientWithProvider(provider),
	}
}

type MatchRequest struct {
	TaskName     string                       `json:"task_name"`
	EventDetail  map[string]interface{}       `json:"event_detail"`
	Resources    []*models.ExecutableResource `json:"resources"`
	SystemPrompt string                       `json:"system_prompt"`
}

type MatchResult struct {
	ResourceID   int     `json:"resource_id"`
	ResourceName string  `json:"resource_name"`
	RequestURL   string  `json:"request_url"`
	Confidence   float64 `json:"confidence"`
	Reasoning    string  `json:"reasoning"`
}

func (m *AIMatcher) MatchResource(req *MatchRequest) (*MatchResult, error) {
	userQuery := m.buildUserQuery(req)
	log.Printf("[AI Matcher] Task: %s", req.TaskName)
	log.Printf("[AI Matcher] User Query:\n%s", userQuery)
	log.Printf("[AI Matcher] Available Resources: %d", len(req.Resources))

	// Use system prompt from request or default
	systemPrompt := req.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = m.GetDefaultSystemPrompt()
	}

	// Call AI using the generic client
	chatResp, err := m.aiClient.Chat(&ChatRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userQuery,
		Temperature:  1.0,
		MaxTokens:    10000,
	})
	if err != nil {
		return nil, err
	}

	var matchResult MatchResult
	if err := json.Unmarshal([]byte(chatResp.Content), &matchResult); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	log.Printf("[AI Matcher] Parsed Result - ResourceID: %d, ResourceName: %s, Confidence: %.2f, Reasoning: %s",
		matchResult.ResourceID, matchResult.ResourceName, matchResult.Confidence, matchResult.Reasoning)

	return &matchResult, nil
}

func (m *AIMatcher) MatchResourceStream(req *MatchRequest, callback func(string)) (*MatchResult, error) {
	config, err := m.aiClient.configStorage.GetAIConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get AI config: %w", err)
	}

	if config.IP == "" || config.Model == "" || config.Token == "" {
		return nil, ErrAINotConfigured
	}

	userQuery := m.buildUserQuery(req)

	requestBody := map[string]interface{}{
		"model":             config.Model,
		"temperature":       float64(1),
		"top_p":             float64(1),
		"max_tokens":        int(10000),
		"top_k":             int(1),
		"presence_penalty":  float64(0),
		"frequency_penalty": float64(0),
		"stream":            true,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": req.SystemPrompt,
			},
			{
				"role":    "user",
				"content": userQuery,
			},
		},
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("http://%s/api/mf-model-api/v1/chat/completions", config.IP)

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+config.Token)

	resp, err := m.aiClient.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var fullContent strings.Builder
	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			if data == "[DONE]" {
				break
			}

			var chunk map[string]interface{}
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if content, ok := delta["content"].(string); ok {
							fullContent.WriteString(content)
							if callback != nil {
								callback(content)
							}
						}
					}
				}
			}
		}
	}

	var matchResult MatchResult
	if err := json.Unmarshal([]byte(fullContent.String()), &matchResult); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &matchResult, nil
}

func (m *AIMatcher) buildUserQuery(req *MatchRequest) string {
	var sb strings.Builder

	// Build resources JSON for the query
	resourcesJSON, _ := json.MarshalIndent(req.Resources, "", "  ")
	eventDetailsJSON, _ := json.MarshalIndent(req.EventDetail, "", "  ")

	// Format the query according to the new template
	sb.WriteString("/explore/(system_prompt=$sys_pmt)\n\n")
	sb.WriteString("## 当前任务信息\n")
	sb.WriteString(fmt.Sprintf("- task_name: %s\n", req.TaskName))
	sb.WriteString("\n## event_details数据：\n")
	sb.WriteString(string(eventDetailsJSON))
	sb.WriteString("\n\n## 可执行资源数据:\n")
	sb.WriteString(string(resourcesJSON))
	sb.WriteString("\n\n## 匹配要求\n")
	sb.WriteString("请严格按照以下规则匹配：\n")
	sb.WriteString(fmt.Sprintf("1. 当前任务名是 '%s'，只能匹配 resource_type='%s' 的资源\n", req.TaskName, req.TaskName))
	sb.WriteString("2. 如果没有匹配 resource_type 的资源，必须将 resource_id 设为 0\n")
	sb.WriteString("3. 不要随意猜测，严格按 task_name 和 resource_type 精确匹配\n")
	sb.WriteString("\n请返回JSON格式的匹配结果 -> ans")

	return sb.String()
}

// GetDefaultSystemPrompt returns the default system prompt for resource matching
func (m *AIMatcher) GetDefaultSystemPrompt() string {
	return prompt.GetResourceMatcherSystemPrompt()
}
