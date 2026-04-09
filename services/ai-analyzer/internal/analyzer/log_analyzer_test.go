package analyzer

import (
	"testing"

	"github.com/quality-gateway/ai-analyzer/internal/types"
)

func TestGetPromptByTaskName(t *testing.T) {
	analyzer := &LogAnalyzer{}

	tests := []struct {
		name     string
		taskName string
		wantType string
	}{
		{
			name:     "basic_ci_all task",
			taskName: "basic_ci_all",
			wantType: "build",
		},
		{
			name:     "basic_ci task",
			taskName: "basic_ci",
			wantType: "build",
		},
		{
			name:     "deployment task",
			taskName: "deployment_deployment",
			wantType: "deployment",
		},
		{
			name:     "deployment short task",
			taskName: "deployment",
			wantType: "deployment",
		},
		{
			name:     "specialized_tests task",
			taskName: "specialized_tests",
			wantType: "specialized",
		},
		{
			name:     "unknown task",
			taskName: "unknown_task",
			wantType: "build", // default
		},
		{
			name:     "empty task",
			taskName: "",
			wantType: "build", // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := analyzer.GetPromptByTaskName(tt.taskName)

			// Basic validation
			if prompt == "" {
				t.Error("prompt should not be empty")
			}

			// Check for expected content based on type
			switch tt.wantType {
			case "build":
				if !contains(prompt, "compilation") || !contains(prompt, "code_lint") {
					t.Error("build prompt should contain compilation and code_lint")
				}
			case "deployment":
				if !contains(prompt, "deployment") {
					t.Error("deployment prompt should contain deployment")
				}
			case "specialized":
				if !contains(prompt, "api_test") || !contains(prompt, "e2e_test") {
					t.Error("specialized prompt should contain api_test and e2e_test")
				}
			}
		})
	}
}

func TestPoolStatsUsagePercentage(t *testing.T) {
	tests := []struct {
		name     string
		stats    types.PoolStats
		expected float64
	}{
		{
			name: "zero total",
			stats: types.PoolStats{TotalSize: 0, InUse: 0},
			expected: 0,
		},
		{
			name: "half used",
			stats: types.PoolStats{TotalSize: 10, InUse: 5},
			expected: 50.0,
		},
		{
			name: "fully used",
			stats: types.PoolStats{TotalSize: 10, InUse: 10},
			expected: 100.0,
		},
		{
			name: "empty",
			stats: types.PoolStats{TotalSize: 10, InUse: 0},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stats.UsagePercentage()
			if got != tt.expected {
				t.Errorf("UsagePercentage() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAIConfigIsConfigured(t *testing.T) {
	tests := []struct {
		name string
		config types.AIConfig
		expected bool
	}{
		{
			name: "fully configured",
			config: types.AIConfig{IP: "192.168.1.1", Model: "gpt-4", Token: "test-token"},
			expected: true,
		},
		{
			name: "missing IP",
			config: types.AIConfig{IP: "", Model: "gpt-4", Token: "test-token"},
			expected: false,
		},
		{
			name: "missing Model",
			config: types.AIConfig{IP: "192.168.1.1", Model: "", Token: "test-token"},
			expected: false,
		},
		{
			name: "missing Token",
			config: types.AIConfig{IP: "192.168.1.1", Model: "gpt-4", Token: ""},
			expected: false,
		},
		{
			name: "all empty",
			config: types.AIConfig{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.IsConfigured()
			if got != tt.expected {
				t.Errorf("IsConfigured() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMergeResult(t *testing.T) {
	analyzer := &LogAnalyzer{}

	t.Run("add new result", func(t *testing.T) {
		allResults := make(map[string]*CheckResult)
		newResult := CheckResult{
			CheckType: "compilation",
			Result:    "pass",
		}

		analyzer.mergeResult(allResults, newResult)

		if len(allResults) != 1 {
			t.Errorf("expected 1 result, got %d", len(allResults))
		}

		if allResults["compilation"].Result != "pass" {
			t.Errorf("expected pass, got %s", allResults["compilation"].Result)
		}
	})

	t.Run("upgrade priority", func(t *testing.T) {
		allResults := map[string]*CheckResult{
			"compilation": {
				CheckType: "compilation",
				Result:    "pass",
			},
		}

		newResult := CheckResult{
			CheckType: "compilation",
			Result:    "fail",
		}

		analyzer.mergeResult(allResults, newResult)

		if allResults["compilation"].Result != "fail" {
			t.Errorf("expected fail (higher priority), got %s", allResults["compilation"].Result)
		}
	})

	t.Run("merge extra fields", func(t *testing.T) {
		allResults := map[string]*CheckResult{
			"compilation": {
				CheckType: "compilation",
				Result:    "pass",
				Extra: map[string]interface{}{
					"chart": "old-chart.tgz",
				},
			},
		}

		newResult := CheckResult{
			CheckType: "compilation",
			Result:    "pass",
			Extra: map[string]interface{}{
				"image": map[string]interface{}{
					"amd64": "test-image:v1",
				},
			},
		}

		analyzer.mergeResult(allResults, newResult)

		extra := allResults["compilation"].Extra
		if extra["chart"] != "old-chart.tgz" {
			t.Errorf("expected to preserve old chart, got %v", extra["chart"])
		}
		if extra["image"] == nil {
			t.Error("expected to add new image field")
		}
	})

	t.Run("append output", func(t *testing.T) {
		allResults := map[string]*CheckResult{
			"compilation": {
				CheckType: "compilation",
				Result:    "pass",
				Output:    "first output",
			},
		}

		newResult := CheckResult{
			CheckType: "compilation",
			Result:    "pass",
			Output:    "second output",
		}

		analyzer.mergeResult(allResults, newResult)

		expected := "first output; second output"
		if allResults["compilation"].Output != expected {
			t.Errorf("expected output '%s', got '%s'", expected, allResults["compilation"].Output)
		}
	})
}

func TestCheckResultPriorityOrder(t *testing.T) {
	// This test verifies the priority order used in mergeResult
	analyzer := &LogAnalyzer{}

	// Test that fail > pass > skipped
	allResults := make(map[string]*CheckResult)

	// Add skipped first
	analyzer.mergeResult(allResults, CheckResult{CheckType: "test", Result: "skipped"})
	if allResults["test"].Result != "skipped" {
		t.Errorf("expected skipped, got %s", allResults["test"].Result)
	}

	// Upgrade to pass
	analyzer.mergeResult(allResults, CheckResult{CheckType: "test", Result: "pass"})
	if allResults["test"].Result != "pass" {
		t.Errorf("expected pass, got %s", allResults["test"].Result)
	}

	// Upgrade to fail
	analyzer.mergeResult(allResults, CheckResult{CheckType: "test", Result: "fail"})
	if allResults["test"].Result != "fail" {
		t.Errorf("expected fail, got %s", allResults["test"].Result)
	}

	// Should not downgrade from fail to pass
	analyzer.mergeResult(allResults, CheckResult{CheckType: "test", Result: "pass"})
	if allResults["test"].Result != "fail" {
		t.Errorf("expected fail (not downgraded), got %s", allResults["test"].Result)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
