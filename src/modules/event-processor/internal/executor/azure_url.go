package executor

import (
	"fmt"
	"net/url"
	"strings"
)

// AzureURL represents the internal Azure DevOps pipeline URL format
// Format: azure://devops.aishu.cn/{organization}/{project}/pipeline/{pipeline_id}
// Example: azure://devops.aishu.cn/AISHUDevOps/DIP/pipeline/6903
const (
	AzureURLScheme = "azure"
	AzureBaseHost  = "devops.aishu.cn"
)

// AzurePipelineInfo contains the parsed information from an internal Azure URL
type AzurePipelineInfo struct {
	Host        string // e.g., "devops.aishu.cn"
	Organization string // e.g., "AISHUDevOps"
	Project     string // e.g., "DIP"
	PipelineID  int    // e.g., 6903
}

// BuildAzureURL creates an internal Azure URL from the given parameters
// Format: azure://{host}/{organization}/{project}/pipeline/{pipeline_id}
func BuildAzureURL(host, organization, project string, pipelineID int) string {
	if host == "" {
		host = AzureBaseHost
	}
	return fmt.Sprintf("%s://%s/%s/%s/pipeline/%d",
		AzureURLScheme, host, organization, project, pipelineID)
}

// ParseAzureURL parses an internal Azure URL and extracts the pipeline information
// Supports formats:
// - azure://devops.aishu.cn/{org}/{project}/pipeline/{id}
// - azure://{org}/{project}/pipeline/{id} (host defaults to devops.aishu.cn)
func ParseAzureURL(azureURL string) (*AzurePipelineInfo, error) {
	if azureURL == "" {
		return nil, fmt.Errorf("empty Azure URL")
	}

	// Parse the URL
	u, err := url.Parse(azureURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL format: %w", err)
	}

	if u.Scheme != AzureURLScheme {
		return nil, fmt.Errorf("invalid scheme: %s, expected: %s", u.Scheme, AzureURLScheme)
	}

	info := &AzurePipelineInfo{
		Host: u.Host,
	}

	// Build the full path for parsing
	// If host is empty or doesn't look like a domain, treat it as part of the path
	fullPath := u.Path
	if info.Host == "" || !strings.Contains(info.Host, ".") {
		// Host is likely the organization, reset it to default
		if info.Host != "" && !strings.Contains(info.Host, ".") {
			fullPath = "/" + info.Host + u.Path
		}
		info.Host = AzureBaseHost
	}

	// Parse path: /{organization}/{project}/pipeline/{pipeline_id}
	path := strings.TrimPrefix(fullPath, "/")
	parts := strings.Split(path, "/")

	if len(parts) < 4 || parts[2] != "pipeline" {
		return nil, fmt.Errorf("invalid Azure URL path format: %s, expected: /{org}/{project}/pipeline/{id}", fullPath)
	}

	info.Organization = parts[0]
	info.Project = parts[1]

	// Parse pipeline ID
	_, err = fmt.Sscanf(parts[3], "%d", &info.PipelineID)
	if err != nil {
		return nil, fmt.Errorf("invalid pipeline ID: %s", parts[3])
	}

	return info, nil
}

// BuildRunURL converts the internal Azure URL to the actual Azure DevOps pipeline run API URL
// Returns: https://{host}/{organization}/{project}/_apis/pipelines/{pipeline_id}/runs
func (info *AzurePipelineInfo) BuildRunURL() string {
	return fmt.Sprintf("https://%s/%s/%s/_apis/pipelines/%d/runs",
		info.Host, info.Organization, info.Project, info.PipelineID)
}

// BuildStatusURL returns the Azure DevOps build status API URL
// Returns: https://{host}/{organization}/{project}/_apis/build/builds/{build_id}
func (info *AzurePipelineInfo) BuildStatusURL(buildID int) string {
	return fmt.Sprintf("https://%s/%s/%s/_apis/build/builds/%d",
		info.Host, info.Organization, info.Project, buildID)
}

// BuildCancelURL returns the Azure DevOps build cancel API URL
// Returns: https://{host}/{organization}/{project}/_apis/build/builds/{build_id}
func (info *AzurePipelineInfo) BuildCancelURL(buildID int) string {
	return info.BuildStatusURL(buildID)
}

// BuildTimelineURL returns the Azure DevOps build timeline API URL
// Returns: https://{host}/{organization}/{project}/_apis/build/builds/{build_id}/timeline
func (info *AzurePipelineInfo) BuildTimelineURL(buildID int) string {
	return fmt.Sprintf("https://%s/%s/%s/_apis/build/builds/%d/timeline",
		info.Host, info.Organization, info.Project, buildID)
}

// BuildLogsURL returns the Azure DevOps build logs API URL
// Returns: https://{host}/{organization}/{project}/_apis/build/builds/{build_id}/logs
func (info *AzurePipelineInfo) BuildLogsURL(buildID int) string {
	return fmt.Sprintf("https://%s/%s/%s/_apis/build/builds/%d/logs",
		info.Host, info.Organization, info.Project, buildID)
}

// BuildWebURL returns the Azure DevOps web UI URL for the build
// Returns: https://{host}/{organization}/{project}/_build/results?buildId={build_id}
func (info *AzurePipelineInfo) BuildWebURL(buildID int) string {
	return fmt.Sprintf("https://%s/%s/%s/_build/results?buildId=%d",
		info.Host, info.Organization, info.Project, buildID)
}

// IsAzureURL checks if a URL is an internal Azure URL
func IsAzureURL(requestURL string) bool {
	return strings.HasPrefix(requestURL, AzureURLScheme+"://")
}
