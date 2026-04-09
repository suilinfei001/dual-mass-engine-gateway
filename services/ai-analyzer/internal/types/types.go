package types

import "time"

// AIConfig represents the AI server configuration
type AIConfig struct {
	IP        string    `json:"ip" db:"ai_ip"`
	Model     string    `json:"model" db:"ai_model"`
	Token     string    `json:"token" db:"ai_token"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// IsConfigured returns true if AI is properly configured
func (c *AIConfig) IsConfigured() bool {
	return c.IP != "" && c.Model != "" && c.Token != ""
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

// PoolStats represents the current state of the request pool
type PoolStats struct {
	TotalSize int `json:"total_size"`
	Available int `json:"available"`
	InUse     int `json:"in_use"`
}

// UsagePercentage returns the percentage of pool in use
func (s PoolStats) UsagePercentage() float64 {
	if s.TotalSize == 0 {
		return 0
	}
	return float64(s.InUse) / float64(s.TotalSize) * 100
}
