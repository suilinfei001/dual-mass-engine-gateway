package models

// 系统配置键常量
const (
	// AI 配置
	ConfigKeyAIIP    = "ai_ip"
	ConfigKeyAIModel = "ai_model"
	ConfigKeyAIToken = "ai_token"

	// Azure DevOps 配置
	ConfigKeyAzureOrganization = "azure_organization"
	ConfigKeyAzureProject      = "azure_project"
	ConfigKeyAzurePAT          = "azure_pat"
	ConfigKeyAzureBaseURL      = "azure_base_url"

	// CMP 配置
	ConfigKeyCMPAccessKey = "cmp_access_key"
	ConfigKeyCMPSecretKey = "cmp_secret_key"
)

// 有效配置键列表（用于验证）
var ValidConfigKeys = []string{
	ConfigKeyAIIP,
	ConfigKeyAIModel,
	ConfigKeyAIToken,
	ConfigKeyAzureOrganization,
	ConfigKeyAzureProject,
	ConfigKeyAzurePAT,
	ConfigKeyAzureBaseURL,
	ConfigKeyCMPAccessKey,
	ConfigKeyCMPSecretKey,
}

// IsValidConfigKey 检查配置键是否有效
func IsValidConfigKey(key string) bool {
	for _, validKey := range ValidConfigKeys {
		if key == validKey {
			return true
		}
	}
	return false
}
