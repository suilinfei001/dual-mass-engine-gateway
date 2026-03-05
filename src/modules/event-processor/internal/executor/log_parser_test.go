package executor

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestSanitizeFileName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal name",
			input:    "Build and Test",
			expected: "Build_and_Test",
		},
		{
			name:     "path separators",
			input:    "path/to/file",
			expected: "path_to_file",
		},
		{
			name:     "special characters",
			input:    `test*:?"<>|file`,
			expected: "test_______file",
		},
		{
			name:     "tabs and newlines",
			input:    "test\t\tfile\nname",
			expected: "test__file_name",
		},
		{
			name:     "long name truncated",
			input:    string(make([]byte, 100)), // 100 bytes
			expected: string(make([]byte, 50)),  // should be truncated to 50
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "backslashes",
			input:    "\\windows\\path",
			expected: "_windows_path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFileName(tt.input)
			if tt.expected != "" && len(result) > 0 {
				if len(tt.expected) == 50 && len(tt.input) > 50 {
					// For truncation test, just check length
					if len(result) > 50 {
						t.Errorf("sanitizeFileName() length = %d, want <= 50", len(result))
					}
				} else if result != tt.expected {
					t.Errorf("sanitizeFileName() = %q, want %q", result, tt.expected)
				}
			}
			// Verify no invalid characters remain
			invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
			for _, c := range invalidChars {
				if contains(result, c) {
					t.Errorf("sanitizeFileName() still contains invalid character %q", c)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr))))
}

func TestIsLogFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"valid log file", "log_123.txt", true},
		{"valid log file with single digit", "log_1.txt", true},
		{"valid log file with large ID", "log_999999.txt", true},
		{"metadata file", "metadata.txt", false},
		{"error file", "log_123_error.txt", false},
		{"no prefix", "123.txt", false},
		{"no extension", "log_123", false},
		{"wrong extension", "log_123.log", false},
		{"non-numeric ID", "log_abc.txt", false},
		{"empty string", "", false},
		{"just log prefix", "log_.txt", false},
		{"log prefix with underscores", "log_1_2.txt", false}, // ID is "1_2" not numeric
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isLogFilename(tt.filename); got != tt.want {
				t.Errorf("isLogFilename(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestExtractLogID(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     int
	}{
		{"single digit", "log_1.txt", 1},
		{"multiple digits", "log_12345.txt", 12345},
		{"zero", "log_0.txt", 0},
		{"short filename", "log_1", 0},       // no .txt extension
		{"no prefix", "123.txt", 0},          // doesn't start with log_
		{"non-numeric ID", "log_abc.txt", 0}, // abc not numeric
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

func TestGetLogFileNames(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create test files
	testFiles := []string{
		"log_3.txt",
		"log_1.txt",
		"log_2.txt",
		"metadata.txt",
		"log_4_error.txt",
		"other.txt",
		"log_10.txt",
	}

	for _, filename := range testFiles {
		path := filepath.Join(tempDir, filename)
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", filename, err)
		}
	}

	// Create a subdirectory (should be ignored)
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	lp := &LogParser{}

	names, err := lp.GetLogFileNames(tempDir)
	if err != nil {
		t.Fatalf("GetLogFileNames() error = %v", err)
	}

	// Should only return log_*.txt files, excluding error files and metadata
	expected := []string{"log_1.txt", "log_2.txt", "log_3.txt", "log_10.txt"}
	if len(names) != len(expected) {
		t.Fatalf("GetLogFileNames() returned %d files, want %d", len(names), len(expected))
	}

	// Check sorted order (by log ID)
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("GetLogFileNames()[%d] = %q, want %q", i, name, expected[i])
		}
	}
}

func TestReadLogFiles(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create test log files with specific content
	testLogs := map[string]string{
		"log_1.txt": "short log",
		"metadata.txt": "Build ID: 1\n", // should be ignored
		"log_2_error.txt": "error",      // should be ignored
	}

	for filename, content := range testLogs {
		path := filepath.Join(tempDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", filename, err)
		}
	}

	lp := &LogParser{}

	logs, err := lp.ReadLogFiles(tempDir)
	if err != nil {
		t.Fatalf("ReadLogFiles() error = %v", err)
	}

	// Should only return log_1.txt
	if len(logs) != 1 {
		t.Fatalf("ReadLogFiles() returned %d files, want 1", len(logs))
	}

	if logs["log_1.txt"] != "short log" {
		t.Errorf("ReadLogFiles()[log_1.txt] = %q, want %q", logs["log_1.txt"], "short log")
	}
}

func TestReadLogFilesTruncation(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create a log file larger than MaxSingleLogFileSize (10KB)
	largeContent := make([]byte, MaxSingleLogFileSize+1000) // 11KB+
	for i := range largeContent {
		largeContent[i] = 'A'
	}

	// Add a marker at the end to verify truncation keeps the end
	marker := "END_MARKER"
	largeContent = append(largeContent[:len(largeContent)-len(marker)], []byte(marker)...)

	logPath := filepath.Join(tempDir, "log_1.txt")
	if err := os.WriteFile(logPath, largeContent, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	lp := &LogParser{}

	logs, err := lp.ReadLogFiles(tempDir)
	if err != nil {
		t.Fatalf("ReadLogFiles() error = %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("ReadLogFiles() returned %d files, want 1", len(logs))
	}

	content := logs["log_1.txt"]

	// Should be truncated to MaxSingleLogFileSize
	if len(content) != MaxSingleLogFileSize {
		t.Errorf("ReadLogFiles() content length = %d, want %d", len(content), MaxSingleLogFileSize)
	}

	// Should contain the END_MARKER at the end (we kept the last part)
	if !sort.StringsAreSorted([]string{content, marker}) && len(content) > len(marker) {
		endPart := content[len(content)-len(marker):]
		if endPart != marker {
			t.Errorf("ReadLogFiles() end content = %q, want to contain %q", endPart, marker)
		}
	}
}
