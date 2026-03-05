package executor

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// LogParser handles parsing of Azure DevOps pipeline logs
type LogParser struct {
	executor *AzureDevOpsExecutor
}

// NewLogParser creates a new log parser
func NewLogParser(executor *AzureDevOpsExecutor) *LogParser {
	return &LogParser{
		executor: executor,
	}
}

// FetchAndStoreLogs fetches logs from Azure DevOps timeline and stores them in separate files
// Returns the directory path containing all log files
func (lp *LogParser) FetchAndStoreLogs(ctx context.Context, buildID int, tempDir string) (string, error) {
	// Create a directory for this build's logs
	buildDir := filepath.Join(tempDir, fmt.Sprintf("build_%d", buildID))
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create build directory: %w", err)
	}

	// First, try to get logs from timeline to get meaningful names
	timeline, err := lp.executor.GetTimeline(ctx, buildID)
	var logIDsToFetch []int
	var recordNames []string

	if err == nil {
		log.Printf("[LogParser] Timeline has %d records", len(timeline.Records))
		for i, record := range timeline.Records {
			// Log first few records to debug
			if i < 10 || record.LogID != nil {
				log.Printf("[LogParser] Record %d: Type=%s, Name=%s, LogID=%v, State=%s", i, record.Type, record.Name, record.LogID, record.State)
			}
			// Fetch logs for records that have a LogID set
			if record.LogID != nil {
				logIDsToFetch = append(logIDsToFetch, *record.LogID)
				// Create a readable filename from the record name
				safeName := sanitizeFileName(record.Name)
				recordNames = append(recordNames, safeName)
			}
		}
	}

	// If timeline doesn't have log IDs, use the log list API
	if len(logIDsToFetch) == 0 {
		log.Printf("[LogParser] No log IDs in timeline, fetching log list directly")
		logList, err := lp.executor.GetLogList(ctx, buildID)
		if err != nil {
			return "", fmt.Errorf("failed to get log list: %w", err)
		}
		log.Printf("[LogParser] Got %d logs from log list API", len(logList))
		for _, logEntry := range logList {
			logIDsToFetch = append(logIDsToFetch, logEntry.LogID)
			recordNames = append(recordNames, fmt.Sprintf("log_%d", logEntry.LogID))
		}
	}

	if len(logIDsToFetch) == 0 {
		return "", fmt.Errorf("no logs found for build %d", buildID)
	}

	// Fetch and store each log as a separate file
	for _, logID := range logIDsToFetch {
		logResult, err := lp.executor.GetLogs(ctx, buildID, logID)
		if err != nil {
			// Create an error file to indicate fetching failed
			errorFilePath := filepath.Join(buildDir, fmt.Sprintf("log_%d_error.txt", logID))
			errorContent := fmt.Sprintf("Error fetching log %d: %v\n", logID, err)
			if writeErr := os.WriteFile(errorFilePath, []byte(errorContent), 0644); writeErr != nil {
				return "", fmt.Errorf("failed to write error file: %w", writeErr)
			}
			continue
		}

		// Write log content to a file with log ID for analyzer compatibility
		// The log analyzer expects files named log_ID.txt format
		logFileName := fmt.Sprintf("log_%d.txt", logID)
		logFilePath := filepath.Join(buildDir, logFileName)
		if err := os.WriteFile(logFilePath, []byte(logResult.Content), 0644); err != nil {
			return "", fmt.Errorf("failed to write log file %d: %w", logID, err)
		}
	}

	// Create a metadata file with build information
	metadataPath := filepath.Join(buildDir, "metadata.txt")
	metadata := fmt.Sprintf("Build ID: %d\nTotal Logs: %d\n", buildID, len(logIDsToFetch))
	if err := os.WriteFile(metadataPath, []byte(metadata), 0644); err != nil {
		return "", fmt.Errorf("failed to write metadata file: %w", err)
	}

	return buildDir, nil
}

// sanitizeFileName converts a string to a safe filename
func sanitizeFileName(name string) string {
	// Replace problematic characters with underscores
	result := name
	for _, c := range []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " ", "\t", "\n", "\r"} {
		result = strings.ReplaceAll(result, c, "_")
	}
	// Limit length
	if len(result) > 50 {
		result = result[:50]
	}
	return result
}

// ReadLogFiles reads all log files from a directory
// Returns a map of log file name to content (truncated to 10K if needed)
const MaxSingleLogFileSize = 10240 // 10KB per log file

func (lp *LogParser) ReadLogFiles(logDir string) (map[string]string, error) {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read log directory: %w", err)
	}

	logFiles := make(map[string]string)

	for _, entry := range entries {
		// Skip directories and non-log files
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		// Only process log_*.txt files, skip metadata.txt and error files
		if !isLogFilename(filename) {
			continue
		}

		filePath := filepath.Join(logDir, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			// Log the error but continue with other files
			fmt.Printf("Failed to read file %s: %v\n", filename, err)
			continue
		}

		// Truncate if content exceeds MaxSingleLogFileSize
		// Keep the last MaxSingleLogFileSize characters (important info is at the end)
		if len(content) > MaxSingleLogFileSize {
			content = content[len(content)-MaxSingleLogFileSize:]
		}

		logFiles[filename] = string(content)
	}

	return logFiles, nil
}

// isLogFilename checks if a filename is a valid log file
func isLogFilename(filename string) bool {
	// Skip metadata, error files, and files that don't match log_*.txt pattern
	if filename == "metadata.txt" {
		return false
	}
	if len(filename) < 4 || filename[len(filename)-4:] != ".txt" {
		return false
	}
	// Check if it starts with "log_"
	if len(filename) < 8 || filename[:4] != "log_" {
		return false
	}
	// Check if the middle part is a number (log_ID.txt format)
	idStr := filename[4 : len(filename)-4]
	_, err := strconv.Atoi(idStr)
	return err == nil
}

// GetLogFileNames returns sorted log file names from a directory
func (lp *LogParser) GetLogFileNames(logDir string) ([]string, error) {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read log directory: %w", err)
	}

	var logNames []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filename := entry.Name()
		if isLogFilename(filename) {
			logNames = append(logNames, filename)
		}
	}

	// Sort by log ID for consistent ordering
	sort.Slice(logNames, func(i, j int) bool {
		idI := extractLogID(logNames[i])
		idJ := extractLogID(logNames[j])
		return idI < idJ
	})

	return logNames, nil
}

// extractLogID extracts the numeric ID from a log filename
func extractLogID(filename string) int {
	// filename format: log_ID.txt
	if len(filename) < 9 {
		return 0
	}
	idStr := filename[4 : len(filename)-4]
	id, _ := strconv.Atoi(idStr)
	return id
}
