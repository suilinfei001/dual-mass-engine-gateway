package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/quality-gateway/ai-analyzer/internal/types"
)

var (
	// ErrAINotConfigured is returned when AI is not configured
	ErrAINotConfigured = fmt.Errorf("AI_NOT_CONFIGURED")
)

// AIConfigProvider provides AI configuration
type AIConfigProvider interface {
	GetAIConfig() (*types.AIConfig, error)
}

// AIClient handles communication with AI server
type AIClient struct {
	configProvider AIConfigProvider
	client         *http.Client
}

// NewAIClient creates a new AI client
func NewAIClient(configProvider AIConfigProvider) *AIClient {
	return &AIClient{
		configProvider: configProvider,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Chat sends a chat completion request to the AI API
func (c *AIClient) Chat(req *types.ChatRequest) (*types.ChatResponse, error) {
	config, err := c.configProvider.GetAIConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get AI config: %w", err)
	}

	if !config.IsConfigured() {
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

	return &types.ChatResponse{
		Content: result.Choices[0].Message.Content,
		RawBody: string(body),
	}, nil
}

// SetTimeout sets the HTTP client timeout
func (c *AIClient) SetTimeout(timeout time.Duration) {
	c.client.Timeout = timeout
}
