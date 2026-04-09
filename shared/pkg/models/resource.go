// Package models provides shared data models for all microservices.
package models

import "time"

// ResourceType represents the type of resource.
type ResourceType string

const (
	ResourceTypeBasicCI    ResourceType = "basic_ci_all"
	ResourceTypeDeployment ResourceType = "deployment_deployment"
	ResourceTypeAPITest    ResourceType = "specialized_tests_api_test"
	ResourceTypeModuleE2E  ResourceType = "specialized_tests_module_e2e"
	ResourceTypeAgentE2E   ResourceType = "specialized_tests_agent_e2e"
	ResourceTypeAIAnalysis ResourceType = "specialized_tests_ai_e2e"
)

// ParseResourceType parses a string into a ResourceType.
func ParseResourceType(s string) (ResourceType, error) {
	switch s {
	case "basic_ci_all":
		return ResourceTypeBasicCI, nil
	case "deployment_deployment":
		return ResourceTypeDeployment, nil
	case "specialized_tests_api_test":
		return ResourceTypeAPITest, nil
	case "specialized_tests_module_e2e":
		return ResourceTypeModuleE2E, nil
	case "specialized_tests_agent_e2e":
		return ResourceTypeAgentE2E, nil
	case "specialized_tests_ai_e2e":
		return ResourceTypeAIAnalysis, nil
	default:
		return "", ErrInvalidResourceType
	}
}

// DisplayName returns a human-readable display name for the resource type.
func (rt ResourceType) DisplayName() string {
	switch rt {
	case ResourceTypeBasicCI:
		return "Basic CI All"
	case ResourceTypeDeployment:
		return "Deployment All"
	case ResourceTypeAPITest:
		return "API Test"
	case ResourceTypeModuleE2E:
		return "Module E2E"
	case ResourceTypeAgentE2E:
		return "Agent E2E"
	case ResourceTypeAIAnalysis:
		return "AI Analysis"
	default:
		return string(rt)
	}
}

// ToQualityCheckType converts ResourceType to QualityCheckType.
func (rt ResourceType) ToQualityCheckType() QualityCheckType {
	return QualityCheckType(rt)
}

// Resource represents a test resource (e.g., Azure DevOps agent).
type Resource struct {
	ID                 int64          `json:"id" db:"id"`
	UUID               string         `json:"uuid" db:"uuid"`
	ResourceType       ResourceType   `json:"resource_type" db:"resource_type"`
	Name               string         `json:"name" db:"name"`
	Description        string         `json:"description" db:"description"`
	AllowSkip          bool           `json:"allow_skip" db:"allow_skip"`
	Organization       string         `json:"organization" db:"organization"`
	Project            string         `json:"project" db:"project"`
	PipelineID         int            `json:"pipeline_id" db:"pipeline_id"`
	PipelineParameters map[string]any `json:"pipeline_parameters" db:"pipeline_parameters"`
	RepoPath           string         `json:"repo_path" db:"repo_path"`
	IsPublic           bool           `json:"is_public" db:"is_public"`
	CreatorID          int64          `json:"creator_id" db:"creator_id"`
	CreatedAt          time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at" db:"updated_at"`
}

// IsAvailable returns true if the resource is available for use.
func (r *Resource) IsAvailable() bool {
	return r.UUID != ""
}
