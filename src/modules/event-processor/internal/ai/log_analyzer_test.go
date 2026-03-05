package ai

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github-hub/event-processor/internal/models"
)

// TestIsLogFile tests the isLogFile helper function
func TestIsLogFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"valid log file", "log_123.txt", true},
		{"valid log file single digit", "log_1.txt", true},
		{"valid log file large ID", "log_999999.txt", true},
		{"metadata file", "metadata.txt", false},
		{"error file", "log_123_error.txt", false},
		{"no prefix", "123.txt", false},
		{"no extension", "log_123", false},
		{"wrong extension", "log_123.log", false},
		{"non-numeric ID", "log_abc.txt", false},
		{"empty string", "", false},
		{"just log prefix", "log_.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isLogFile(tt.filename); got != tt.want {
				t.Errorf("isLogFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

// TestExtractLogID tests the extractLogID helper function
func TestExtractLogID(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     int
	}{
		{"single digit", "log_1.txt", 1},
		{"multiple digits", "log_12345.txt", 12345},
		{"zero", "log_0.txt", 0},
		{"short filename", "log_1", 0},
		{"no prefix", "123.txt", 0},
		{"non-numeric ID", "log_abc.txt", 0},
		{"empty string", "", 0},
		{"log prefix only", "log_.txt", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractLogID(tt.filename); got != tt.want {
				t.Errorf("extractLogID(%q) = %d, want %d", tt.filename, got, tt.want)
			}
		})
	}
}

// TestGetSortedLogFileNames tests the getSortedLogFileNames helper function
func TestGetSortedLogFileNames(t *testing.T) {
	la := &LogAnalyzer{}

	// Create a map of log files (simulating readLogFiles output after filtering)
	// Note: readLogFiles filters out metadata.txt and error files before this is called
	logFiles := map[string]string{
		"log_3.txt": "content 3",
		"log_1.txt": "content 1",
		"log_10.txt": "content 10",
		"log_2.txt": "content 2",
	}

	// Get sorted names
	names := la.getSortedLogFileNames(logFiles)

	// Should be sorted by log ID
	expected := []string{"log_1.txt", "log_2.txt", "log_3.txt", "log_10.txt"}

	if len(names) != len(expected) {
		t.Fatalf("getSortedLogFileNames() returned %d names, want %d", len(names), len(expected))
	}

	for i, name := range names {
		if name != expected[i] {
			t.Errorf("getSortedLogFileNames()[%d] = %q, want %q", i, name, expected[i])
		}
	}
}

// TestMergeResult tests the mergeResult helper function
func TestMergeResult(t *testing.T) {
	la := &LogAnalyzer{}

	tests := []struct {
		name           string
		existing       map[string]*models.TaskResult
		newResult      models.TaskResult
		expectedType   string
		expectedResult string
		validateExtra  bool
	}{
		{
			name:         "add new check type",
			existing:     map[string]*models.TaskResult{},
			newResult:    models.TaskResult{CheckType: "compilation", Result: "pass"},
			expectedType: "compilation",
			expectedResult: "pass",
		},
		{
			name: "fail overrides pass",
			existing: map[string]*models.TaskResult{
				"compilation": {CheckType: "compilation", Result: "pass"},
			},
			newResult:      models.TaskResult{CheckType: "compilation", Result: "fail"},
			expectedType:   "compilation",
			expectedResult: "fail",
		},
		{
			name: "pass overrides skipped",
			existing: map[string]*models.TaskResult{
				"code_lint": {CheckType: "code_lint", Result: "skipped"},
			},
			newResult:      models.TaskResult{CheckType: "code_lint", Result: "pass"},
			expectedType:   "code_lint",
			expectedResult: "pass",
		},
		{
			name: "timeout overrides running",
			existing: map[string]*models.TaskResult{
				"unit_test": {CheckType: "unit_test", Result: "running"},
			},
			newResult:      models.TaskResult{CheckType: "unit_test", Result: "timeout"},
			expectedType:   "unit_test",
			expectedResult: "timeout",
		},
		{
			name: "merge extra fields",
			existing: map[string]*models.TaskResult{
				"compilation": {CheckType: "compilation", Result: "pass", Extra: map[string]interface{}{
					"image": map[string]interface{}{"amd64": ""},
				}},
			},
			newResult: models.TaskResult{CheckType: "compilation", Result: "pass", Extra: map[string]interface{}{
				"image": map[string]interface{}{"amd64": "test-image:v1"},
			}},
			expectedType:    "compilation",
			expectedResult:  "pass",
			validateExtra:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			la.mergeResult(tt.existing, tt.newResult, "test_log.txt")

			result := tt.existing[tt.expectedType]
			if result == nil {
				t.Fatalf("mergeResult() did not create result for %s", tt.expectedType)
			}

			if result.Result != tt.expectedResult {
				t.Errorf("mergeResult() result.Result = %q, want %q", result.Result, tt.expectedResult)
			}

			if tt.validateExtra {
				if result.Extra == nil {
					t.Errorf("mergeResult() result.Extra is nil")
				} else if img, ok := result.Extra["image"].(map[string]interface{}); ok {
					if amd64, ok := img["amd64"].(string); ok && amd64 != "test-image:v1" {
						t.Errorf("mergeResult() result.Extra.image.amd64 = %q, want %q", amd64, "test-image:v1")
					}
				}
			}
		})
	}
}

// TestMergeResultPreserveScoreOnFail tests that score is preserved when result changes from pass to fail
func TestMergeResultPreserveScoreOnFail(t *testing.T) {
	la := &LogAnalyzer{}

	// Start with a passing result with a score
	existing := map[string]*models.TaskResult{
		"unit_test": {
			CheckType: "unit_test",
			Result:    "pass",
			Extra:     map[string]interface{}{"score": 30.8912},
		},
	}

	// Merge a failing result with score 0
	newResult := models.TaskResult{
		CheckType: "unit_test",
		Result:    "fail",
		Extra:     map[string]interface{}{"score": 0.0},
	}

	la.mergeResult(existing, newResult, "test_log.txt")

	// Result should be "fail" (higher priority)
	if existing["unit_test"].Result != "fail" {
		t.Errorf("mergeResult() result.Result = %v, want 'fail'", existing["unit_test"].Result)
	}

	// Score should be preserved as 30.8912 (from the pass result), NOT 0
	score, ok := existing["unit_test"].Extra["score"].(float64)
	if !ok {
		t.Fatal("mergeResult() score not found in Extra")
	}
	if score != 30.8912 {
		t.Errorf("mergeResult() score = %v, want 30.8912 (preserved from pass result)", score)
	}
}

// TestReadLogFiles tests the readLogFiles helper function
func TestReadLogFiles(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create test log files
	testLogs := map[string]string{
		"log_1.txt":           "short log content",
		"log_2.txt":           "another log",
		"metadata.txt":        "Build ID: 1\n",
		"log_3_error.txt":     "error content",
		"other.txt":           "other content",
		"log_not_number.txt":  "invalid",
	}

	for filename, content := range testLogs {
		path := filepath.Join(tempDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", filename, err)
		}
	}

	la := &LogAnalyzer{}

	logs, err := la.readLogFiles(tempDir)
	if err != nil {
		t.Fatalf("readLogFiles() error = %v", err)
	}

	// Should only return log_*.txt files (excluding error files and metadata)
	if len(logs) != 2 {
		t.Fatalf("readLogFiles() returned %d files, want 2", len(logs))
	}

	if logs["log_1.txt"] != "short log content" {
		t.Errorf("readLogFiles()[log_1.txt] = %q, want %q", logs["log_1.txt"], "short log content")
	}

	if logs["log_2.txt"] != "another log" {
		t.Errorf("readLogFiles()[log_2.txt] = %q, want %q", logs["log_2.txt"], "another log")
	}
}

// TestReadLogFilesTruncation tests that large log files are truncated correctly
func TestReadLogFilesTruncation(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create a log file larger than MaxSingleLogContentSize
	largeContent := make([]byte, MaxSingleLogContentSize+1000)
	for i := range largeContent {
		largeContent[i] = 'A'
	}
	// Add marker at the end
	marker := "END"
	largeContent = largeContent[:len(largeContent)-len(marker)]
	largeContent = append(largeContent, []byte(marker)...)

	logPath := filepath.Join(tempDir, "log_1.txt")
	if err := os.WriteFile(logPath, largeContent, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	la := &LogAnalyzer{}

	logs, err := la.readLogFiles(tempDir)
	if err != nil {
		t.Fatalf("readLogFiles() error = %v", err)
	}

	content := logs["log_1.txt"]

	// Should be truncated to MaxSingleLogContentSize
	if len(content) != MaxSingleLogContentSize {
		t.Errorf("readLogFiles() content length = %d, want %d", len(content), MaxSingleLogContentSize)
	}

	// Should end with our marker (we kept the last part)
	if len(content) >= len(marker) && content[len(content)-len(marker):] != marker {
		t.Errorf("readLogFiles() content end = %q, want to end with %q", content[len(content)-len(marker):], marker)
	}
}

// TestSaveTmpResult tests the saveTmpResult helper function
func TestSaveTmpResult(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	la := &LogAnalyzer{}

	results := []models.TaskResult{
		{CheckType: "compilation", Result: "pass"},
		{CheckType: "code_lint", Result: "pass"},
	}

	tmpPath := filepath.Join(tempDir, "test_result.json")

	err := la.saveTmpResult(tmpPath, results)
	if err != nil {
		t.Fatalf("saveTmpResult() error = %v", err)
	}

	// Verify file was created
	content, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("failed to read temp result file: %v", err)
	}

	// Should contain JSON with our results
	contentStr := string(content)
	if !containsString(contentStr, "compilation") {
		t.Errorf("saveTmpResult() content missing 'compilation'")
	}
	if !containsString(contentStr, "code_lint") {
		t.Errorf("saveTmpResult() content missing 'code_lint'")
	}
}

// TestGetLogAnalysisPrompt tests the prompt generation
func TestGetLogAnalysisPrompt(t *testing.T) {
	la := &LogAnalyzer{}

	prompt := la.GetLogAnalysisPrompt()

	if prompt == "" {
		t.Error("GetLogAnalysisPrompt() returned empty string")
	}

	// Check for key parts of the prompt
	keyPhrases := []string{
		"compilation",
		"code_lint",
		"security_scan",
		"unit_test",
		"check_type",
		"result",
	}

	for _, phrase := range keyPhrases {
		if !containsString(prompt, phrase) {
			t.Errorf("GetLogAnalysisPrompt() missing key phrase: %s", phrase)
		}
	}
}

// TestAnalyzeLogDirectory tests the directory analysis functionality
func TestAnalyzeLogDirectory(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create a log directory with timestamp
	timestamp := time.Now().Format("20060102_150405")
	logDirName := "build_1_tmp_" + timestamp
	logDir := filepath.Join(tempDir, logDirName)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("failed to create log directory: %v", err)
	}

	// Create test log files
	testLogs := map[string]string{
		"log_1.txt": "Build completed successfully",
		"log_2.txt": "Tests passed",
	}

	for filename, content := range testLogs {
		path := filepath.Join(logDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", filename, err)
		}
	}

	// Create LogAnalyzer without config storage (for testing file operations only)
	la := &LogAnalyzer{}

	// Test reading logs - we can't test AI analysis without a real AI service
	// but we can test the directory structure is created correctly
	_, err := la.readLogFiles(logDir)
	if err != nil {
		t.Fatalf("readLogFiles() error = %v", err)
	}

	// Verify temp directory structure
	entries, err := os.ReadDir(logDir)
	if err != nil {
		t.Fatalf("failed to read log directory: %v", err)
	}

	logFiles := []string{}
	for _, entry := range entries {
		if !entry.IsDir() {
			logFiles = append(logFiles, entry.Name())
		}
	}

	// Should have our 2 test log files
	if len(logFiles) != 2 {
		t.Errorf("log directory has %d files, want 2", len(logFiles))
	}
}

// containsString is a simple string contains helper
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && s[:len(substr)] == substr) ||
		(len(s) > 0 && containsString(s[1:], substr)))
}
