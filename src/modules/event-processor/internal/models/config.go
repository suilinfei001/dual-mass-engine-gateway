package models

import "time"

type SystemConfig struct {
	ID          int       `json:"id"`
	ConfigKey   string    `json:"config_key"`
	ConfigValue string    `json:"config_value"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SystemConfigUpdateRequest struct {
	ConfigKey   string `json:"config_key"`
	ConfigValue string `json:"config_value"`
}

const (
	ConfigKeyAIIP               = "ai_ip"
	ConfigKeyAIModel            = "ai_model"
	ConfigKeyAIToken            = "ai_token"
	ConfigKeyAzurePAT           = "azure_pat"
	ConfigKeyEventReceiverIP    = "event_receiver_ip"
	ConfigKeyLogRetentionDays   = "log_retention_days"
	ConfigKeyAIConcurrency      = "ai_concurrency"
	ConfigKeyAIRequestPoolSize  = "ai_request_pool_size"
	ConfigKeyCMPAccessKey       = "cmp_access_key"
	ConfigKeyCMPSecretKey       = "cmp_secret_key"
)

func NewSystemConfig(key, value, description string) *SystemConfig {
	now := time.Now()
	return &SystemConfig{
		ConfigKey:   key,
		ConfigValue: value,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
