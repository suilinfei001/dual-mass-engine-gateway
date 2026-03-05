package executor

import (
	"testing"
)

func TestBuildAzureURL(t *testing.T) {
	tests := []struct {
		name        string
		host        string
		org         string
		project     string
		pipelineID  int
		expectedURL string
	}{
		{
			name:        "Standard URL",
			host:        "devops.aishu.cn",
			org:         "AISHUDevOps",
			project:     "DIP",
			pipelineID:  6903,
			expectedURL: "azure://devops.aishu.cn/AISHUDevOps/DIP/pipeline/6903",
		},
		{
			name:        "Default host",
			host:        "",
			org:         "AISHUDevOps",
			project:     "DIP",
			pipelineID:  3875,
			expectedURL: "azure://devops.aishu.cn/AISHUDevOps/DIP/pipeline/3875",
		},
		{
			name:        "Different project",
			host:        "devops.aishu.cn",
			org:         "AISHUDevOps",
			project:     "AnyShareFamily",
			pipelineID:  3875,
			expectedURL: "azure://devops.aishu.cn/AISHUDevOps/AnyShareFamily/pipeline/3875",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildAzureURL(tt.host, tt.org, tt.project, tt.pipelineID)
			if result != tt.expectedURL {
				t.Errorf("BuildAzureURL() = %v, want %v", result, tt.expectedURL)
			}
		})
	}
}

func TestParseAzureURL(t *testing.T) {
	tests := []struct {
		name        string
		azureURL    string
		wantOrg     string
		wantProject string
		wantID      int
		wantHost    string
		wantErr     bool
	}{
		{
			name:        "Valid URL with host",
			azureURL:    "azure://devops.aishu.cn/AISHUDevOps/DIP/pipeline/6903",
			wantOrg:     "AISHUDevOps",
			wantProject: "DIP",
			wantID:      6903,
			wantHost:    "devops.aishu.cn",
			wantErr:     false,
		},
		{
			name:        "Valid URL default host",
			azureURL:    "azure://AISHUDevOps/DIP/pipeline/6903",
			wantOrg:     "AISHUDevOps",
			wantProject: "DIP",
			wantID:      6903,
			wantHost:    "devops.aishu.cn", // Default host
			wantErr:     false,
		},
		{
			name:     "Empty URL",
			azureURL: "",
			wantErr:  true,
		},
		{
			name:     "Invalid scheme",
			azureURL: "http://devops.aishu.cn/AISHUDevOps/DIP/pipeline/6903",
			wantErr:  true,
		},
		{
			name:     "Invalid path format",
			azureURL: "azure://devops.aishu.cn/AISHUDevOps/DIP",
			wantErr:  true,
		},
		{
			name:     "Invalid pipeline ID",
			azureURL: "azure://devops.aishu.cn/AISHUDevOps/DIP/pipeline/abc",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAzureURL(tt.azureURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAzureURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Organization != tt.wantOrg {
					t.Errorf("ParseAzureURL() Organization = %v, want %v", got.Organization, tt.wantOrg)
				}
				if got.Project != tt.wantProject {
					t.Errorf("ParseAzureURL() Project = %v, want %v", got.Project, tt.wantProject)
				}
				if got.PipelineID != tt.wantID {
					t.Errorf("ParseAzureURL() PipelineID = %v, want %v", got.PipelineID, tt.wantID)
				}
				if got.Host != tt.wantHost {
					t.Errorf("ParseAzureURL() Host = %v, want %v", got.Host, tt.wantHost)
				}
			}
		})
	}
}

func TestBuildRunURL(t *testing.T) {
	info := &AzurePipelineInfo{
		Host:        "devops.aishu.cn",
		Organization: "AISHUDevOps",
		Project:     "DIP",
		PipelineID:  6903,
	}

	expected := "https://devops.aishu.cn/AISHUDevOps/DIP/_apis/pipelines/6903/runs"
	result := info.BuildRunURL()
	if result != expected {
		t.Errorf("BuildRunURL() = %v, want %v", result, expected)
	}
}

func TestBuildStatusURL(t *testing.T) {
	info := &AzurePipelineInfo{
		Host:        "devops.aishu.cn",
		Organization: "AISHUDevOps",
		Project:     "DIP",
		PipelineID:  6903,
	}

	buildID := 12345
	expected := "https://devops.aishu.cn/AISHUDevOps/DIP/_apis/build/builds/12345"
	result := info.BuildStatusURL(buildID)
	if result != expected {
		t.Errorf("BuildStatusURL() = %v, want %v", result, expected)
	}
}

func TestBuildWebURL(t *testing.T) {
	info := &AzurePipelineInfo{
		Host:        "devops.aishu.cn",
		Organization: "AISHUDevOps",
		Project:     "DIP",
		PipelineID:  6903,
	}

	buildID := 12345
	expected := "https://devops.aishu.cn/AISHUDevOps/DIP/_build/results?buildId=12345"
	result := info.BuildWebURL(buildID)
	if result != expected {
		t.Errorf("BuildWebURL() = %v, want %v", result, expected)
	}
}

func TestIsAzureURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "Azure URL",
			url:      "azure://devops.aishu.cn/AISHUDevOps/DIP/pipeline/6903",
			expected: true,
		},
		{
			name:     "HTTP URL",
			url:      "http://localhost:8090/mock",
			expected: false,
		},
		{
			name:     "HTTPS URL",
			url:      "https://devops.aishu.cn/AISHUDevOps/DIP/_apis/pipelines/6903/runs",
			expected: false,
		},
		{
			name:     "Empty string",
			url:      "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAzureURL(tt.url)
			if result != tt.expected {
				t.Errorf("IsAzureURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRoundTripAzureURL(t *testing.T) {
	// Test that we can build a URL and then parse it back correctly
	originalInfo := &AzurePipelineInfo{
		Host:        "devops.aishu.cn",
		Organization: "AISHUDevOps",
		Project:     "DIP",
		PipelineID:  6903,
	}

	// Build URL
	azureURL := BuildAzureURL(originalInfo.Host, originalInfo.Organization, originalInfo.Project, originalInfo.PipelineID)

	// Parse it back
	parsedInfo, err := ParseAzureURL(azureURL)
	if err != nil {
		t.Fatalf("ParseAzureURL() error = %v", err)
	}

	// Compare
	if parsedInfo.Host != originalInfo.Host {
		t.Errorf("Host mismatch: got %v, want %v", parsedInfo.Host, originalInfo.Host)
	}
	if parsedInfo.Organization != originalInfo.Organization {
		t.Errorf("Organization mismatch: got %v, want %v", parsedInfo.Organization, originalInfo.Organization)
	}
	if parsedInfo.Project != originalInfo.Project {
		t.Errorf("Project mismatch: got %v, want %v", parsedInfo.Project, originalInfo.Project)
	}
	if parsedInfo.PipelineID != originalInfo.PipelineID {
		t.Errorf("PipelineID mismatch: got %v, want %v", parsedInfo.PipelineID, originalInfo.PipelineID)
	}
}
