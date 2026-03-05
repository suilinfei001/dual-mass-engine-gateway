package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type ResourceType string

const (
	ResourceTypeBasicCIAll          ResourceType = "basic_ci_all"
	ResourceTypeDeployment          ResourceType = "deployment_deployment"
	ResourceTypeAPITest             ResourceType = "specialized_tests_api_test"
	ResourceTypeModuleE2E           ResourceType = "specialized_tests_module_e2e"
	ResourceTypeAgentE2E            ResourceType = "specialized_tests_agent_e2e"
	ResourceTypeAIE2E               ResourceType = "specialized_tests_ai_e2e"
)

var ValidResourceTypes = []ResourceType{
	ResourceTypeBasicCIAll,
	ResourceTypeDeployment,
	ResourceTypeAPITest,
	ResourceTypeModuleE2E,
	ResourceTypeAgentE2E,
	ResourceTypeAIE2E,
}

func IsValidResourceType(rt string) bool {
	for _, t := range ValidResourceTypes {
		if string(t) == rt {
			return true
		}
	}
	return false
}

type ExecutableResource struct {
	ID               int                    `json:"id"`
	ResourceName     string                 `json:"resource_name"`
	ResourceType     ResourceType           `json:"resource_type"`
	AllowSkip        bool                   `json:"allow_skip"`
	Organization     string                 `json:"organization"`
	Project          string                 `json:"project"`
	PipelineID       int                    `json:"pipeline_id"`
	PipelineParams   map[string]interface{} `json:"pipeline_params,omitempty"`
	MicroserviceName string                 `json:"microservice_name,omitempty"`
	PodName          string                 `json:"pod_name,omitempty"`
	RepoPath         string                 `json:"repo_path"`
	Description      string                 `json:"description,omitempty"`
	CreatorID        int                    `json:"creator_id"`
	CreatorName      string                 `json:"creator_name"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

func (r *ExecutableResource) GetPipelineParamsJSON() (string, error) {
	if r.PipelineParams == nil {
		return "{}", nil
	}
	data, err := json.Marshal(r.PipelineParams)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r *ExecutableResource) SetPipelineParamsFromJSON(jsonStr string) error {
	if jsonStr == "" {
		r.PipelineParams = make(map[string]interface{})
		return nil
	}
	return json.Unmarshal([]byte(jsonStr), &r.PipelineParams)
}

// RequestURL 构建资源的请求 URL
func (r *ExecutableResource) RequestURL() string {
	// 如果有 Azure 配置，使用 Azure URL 格式
	if r.Organization != "" && r.Project != "" && r.PipelineID > 0 {
		return fmt.Sprintf("azure://devops.aishu.cn/%s/%s/pipeline/%d", r.Organization, r.Project, r.PipelineID)
	}

	// 否则，使用微服务或 pod 名称（向后兼容）
	baseURL := ""
	if r.MicroserviceName != "" {
		baseURL = fmt.Sprintf("http://%s", r.MicroserviceName)
	} else if r.PodName != "" {
		baseURL = fmt.Sprintf("http://%s", r.PodName)
	} else {
		baseURL = fmt.Sprintf("http://pipeline-%d", r.PipelineID)
	}

	var path string
	switch r.ResourceType {
	case ResourceTypeBasicCIAll:
		path = "/mock/basic-ci"
	case ResourceTypeDeployment:
		path = "/mock/deployment"
	case ResourceTypeAPITest:
		path = "/mock/api-test"
	case ResourceTypeModuleE2E:
		path = "/mock/module-e2e"
	case ResourceTypeAgentE2E:
		path = "/mock/agent-e2e"
	case ResourceTypeAIE2E:
		path = "/mock/ai-e2e"
	default:
		path = "/mock/unknown"
	}

	return baseURL + path
}

type ExecutableResourceCreateRequest struct {
	ResourceName     string                 `json:"resource_name"`
	ResourceType     string                 `json:"resource_type"`
	AllowSkip        bool                   `json:"allow_skip"`
	Organization     string                 `json:"organization"`
	Project          string                 `json:"project"`
	PipelineID       int                    `json:"pipeline_id"`
	PipelineParams   map[string]interface{} `json:"pipeline_params,omitempty"`
	MicroserviceName string                 `json:"microservice_name,omitempty"`
	PodName          string                 `json:"pod_name,omitempty"`
	RepoPath         string                 `json:"repo_path"`
	Description      string                 `json:"description,omitempty"`
}

type ExecutableResourceUpdateRequest struct {
	AllowSkip        bool                    `json:"allow_skip,omitempty"`
	Organization     string                 `json:"organization,omitempty"`
	Project          string                 `json:"project,omitempty"`
	PipelineID       int                    `json:"pipeline_id,omitempty"`
	PipelineParams   map[string]interface{} `json:"pipeline_params,omitempty"`
	MicroserviceName string                 `json:"microservice_name,omitempty"`
	PodName          string                 `json:"pod_name,omitempty"`
	RepoPath         string                 `json:"repo_path,omitempty"`
	Description      string                 `json:"description,omitempty"`
}

func NewExecutableResource(req *ExecutableResourceCreateRequest, creatorID int, creatorName string) *ExecutableResource {
	now := time.Now()

	// Auto-append skip note to description if allow_skip is true
	description := req.Description
	if req.AllowSkip && description == "" {
		description = "此检查项可以跳过"
	} else if req.AllowSkip && !strings.Contains(description, "可以跳过") {
		description = description + " (此检查项可以跳过)"
	}

	return &ExecutableResource{
		ResourceName:     req.ResourceName,
		ResourceType:     ResourceType(req.ResourceType),
		AllowSkip:        req.AllowSkip,
		Organization:     req.Organization,
		Project:          req.Project,
		PipelineID:       req.PipelineID,
		PipelineParams:   req.PipelineParams,
		MicroserviceName: req.MicroserviceName,
		PodName:          req.PodName,
		RepoPath:         req.RepoPath,
		Description:      description,
		CreatorID:        creatorID,
		CreatorName:      creatorName,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}
