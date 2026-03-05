package executor

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github-hub/event-processor/internal/api"
	"github-hub/event-processor/internal/ai"
	"github-hub/event-processor/internal/models"
	"github-hub/event-processor/internal/storage"
)

// TaskExecutionService handles task execution using Azure DevOps pipelines
type TaskExecutionService struct {
	configStorage  *storage.MySQLConfigStorage
	resourceStorage *storage.MySQLResourceStorage
	taskStorage    *storage.MySQLTaskStorage
	eventClient    *api.Client
	logAnalyzer    *ai.LogAnalyzer
}

// NewTaskExecutionService creates a new TaskExecutionService
func NewTaskExecutionService(
	configStorage *storage.MySQLConfigStorage,
	resourceStorage *storage.MySQLResourceStorage,
	taskStorage *storage.MySQLTaskStorage,
	eventClient *api.Client,
) *TaskExecutionService {
	return &TaskExecutionService{
		configStorage:  configStorage,
		resourceStorage: resourceStorage,
		taskStorage:    taskStorage,
		eventClient:    eventClient,
		logAnalyzer:    ai.NewLogAnalyzer(configStorage),
	}
}

// ExecuteTask executes a task using Azure DevOps pipeline
func (s *TaskExecutionService) ExecuteTask(ctx context.Context, task *models.Task, resource *models.ExecutableResource) error {
	log.Printf("[TaskExecutionService] Executing task %s (event_id: %d)", task.TaskName, task.EventID)

	// Get resource if not provided
	if resource == nil {
		var err error
		resource, err = s.getTaskResource(task)
		if err != nil {
			return fmt.Errorf("failed to get task resource: %w", err)
		}
	}

	// Check if the resource allows skipping
	if resource.AllowSkip {
		log.Printf("[TaskExecutionService] Resource allows skip, marking task %s as skipped", task.TaskName)
		reason := fmt.Sprintf("Resource %s allows skipping", resource.ResourceName)
		if err := s.taskStorage.MarkTaskSkipped(task, reason); err != nil {
			return fmt.Errorf("failed to mark task skipped: %w", err)
		}
		// Update task status in memory
		task.Status = models.TaskStatusSkipped
		task.EndTime = &models.LocalTime{Time: time.Now()}
		task.ErrorMessage = &reason
		return nil
	}

	// Get Azure PAT from config storage
	azureConfig, err := s.configStorage.GetAzureConfig()
	if err != nil {
		return fmt.Errorf("failed to get Azure config: %w", err)
	}
	if azureConfig.PAT == "" {
		return fmt.Errorf("Azure PAT not configured")
	}

	// Get event to extract branch information
	branch := "refs/heads/main" // default branch

	// For basic_ci_all, use refs/heads/develop as the Azure DevOps branch
	if task.TaskName == "basic_ci_all" {
		branch = "refs/heads/develop"
		log.Printf("[TaskExecutionService] Using default Azure branch for basic_ci_all: %s", branch)
	} else if s.eventClient != nil {
		// For other task types, try to get the branch from the event
		event, err := s.eventClient.GetEvent(task.EventID)
		if err == nil && event != nil {
			// Use the branch from the event
			if event.Branch != "" {
				// Format branch as refs/heads/{branch} if not already formatted
				if !strings.HasPrefix(event.Branch, "refs/") {
					branch = "refs/heads/" + event.Branch
				} else {
					branch = event.Branch
				}
				log.Printf("[TaskExecutionService] Using branch from event: %s", branch)
			}
		} else {
			log.Printf("[TaskExecutionService] Failed to get event, using default branch: %v", err)
		}
	}

	// Build executor config from resource
	execConfig := &ExecutorConfig{
		Organization: resource.Organization,
		Project:      resource.Project,
		PAT:          azureConfig.PAT,
		Branch:       branch,
		VerifySSL:    true,
	}

	// Create executor
	exec := NewAzureDevOpsExecutor(execConfig)

	// Prepare pipeline parameters
	var params map[string]interface{}
	if resource.PipelineParams != nil {
		params = resource.PipelineParams
	}

	// Run the pipeline
	result, err := exec.Run(ctx, resource.PipelineID, params)
	if err != nil {
		return fmt.Errorf("failed to run pipeline: %w", err)
	}

	log.Printf("[TaskExecutionService] Pipeline started: build_id=%d, build_number=%s", result.BuildID, result.BuildNumber)

	// Update task with build information
	task.BuildID = result.BuildID
	task.RequestURL = result.WebURL

	// Mark task as running
	now := time.Now()
	task.Status = models.TaskStatusRunning
	task.StartTime = &models.LocalTime{Time: now}
	task.UpdatedAt = now

	if err := s.taskStorage.UpdateTask(task); err != nil {
		return fmt.Errorf("failed to update task with build info: %w", err)
	}

	return nil
}

// CheckAndSkipIfNeeded checks if the task's resource allows skipping and marks it as skipped if so
// Returns true if the task was skipped, false otherwise
func (s *TaskExecutionService) CheckAndSkipIfNeeded(task *models.Task) (bool, error) {
	// Get the resource for this task
	resource, err := s.getTaskResource(task)
	if err != nil {
		return false, fmt.Errorf("failed to get task resource: %w", err)
	}
	if resource == nil {
		return false, nil // No resource found, cannot skip
	}

	// Check if resource allows skipping
	if resource.AllowSkip {
		log.Printf("[TaskExecutionService] Resource allows skip, marking task %s as skipped (resource: %s)",
			task.TaskName, resource.ResourceName)
		reason := fmt.Sprintf("Resource %s allows skipping", resource.ResourceName)
		if err := s.taskStorage.MarkTaskSkipped(task, reason); err != nil {
			return false, fmt.Errorf("failed to mark task skipped: %w", err)
		}
		// Update task status in memory
		task.Status = models.TaskStatusSkipped
		task.EndTime = &models.LocalTime{Time: time.Now()}
		task.ErrorMessage = &reason
		return true, nil
	}

	return false, nil // Resource does not allow skipping
}

// PollTaskStatus polls the status of a running task
// Returns the current status and results if the task is complete
func (s *TaskExecutionService) PollTaskStatus(ctx context.Context, task *models.Task) (models.TaskStatus, []models.TaskResult, error) {
	if task.BuildID == 0 {
		// No build ID, task may not have been executed yet
		return models.TaskStatusPending, nil, nil
	}

	// Get the resource associated with this task
	resource, err := s.getTaskResource(task)
	if err != nil {
		return models.TaskStatusFailed, nil, fmt.Errorf("failed to get task resource: %w", err)
	}

	// Get Azure PAT from config storage
	azureConfig, err := s.configStorage.GetAzureConfig()
	if err != nil {
		return models.TaskStatusFailed, nil, fmt.Errorf("failed to get Azure config: %w", err)
	}

	// Build executor config from resource
	execConfig := &ExecutorConfig{
		Organization: resource.Organization,
		Project:      resource.Project,
		PAT:          azureConfig.PAT,
		Branch:       "refs/heads/main",
		VerifySSL:    true,
	}

	// Create executor
	exec := NewAzureDevOpsExecutor(execConfig)

	// Get status from Azure DevOps
	statusResult, err := exec.GetStatus(ctx, task.BuildID)
	if err != nil {
		log.Printf("[TaskExecutionService] Failed to get status for build_id=%d: %v", task.BuildID, err)
		return models.TaskStatusRunning, nil, nil
	}

	// Map Azure status to task status
	taskStatus, results := s.mapAzureStatusToTaskStatus(statusResult, exec, ctx)

	return taskStatus, results, nil
}

// CancelTask cancels a running task
func (s *TaskExecutionService) CancelTask(ctx context.Context, task *models.Task) error {
	if task.BuildID == 0 {
		return fmt.Errorf("cannot cancel task without build ID")
	}

	// Get the resource associated with this task
	resource, err := s.getTaskResource(task)
	if err != nil {
		return fmt.Errorf("failed to get task resource: %w", err)
	}

	// Get Azure PAT from config storage
	azureConfig, err := s.configStorage.GetAzureConfig()
	if err != nil {
		return fmt.Errorf("failed to get Azure config: %w", err)
	}

	// Build executor config from resource
	execConfig := &ExecutorConfig{
		Organization: resource.Organization,
		Project:      resource.Project,
		PAT:          azureConfig.PAT,
		Branch:       "refs/heads/main",
		VerifySSL:    true,
	}

	// Create executor
	exec := NewAzureDevOpsExecutor(execConfig)

	// Cancel the pipeline
	if err := exec.Cancel(ctx, task.BuildID); err != nil {
		return fmt.Errorf("failed to cancel pipeline: %w", err)
	}

	log.Printf("[TaskExecutionService] Cancelled build_id=%d for task %s", task.BuildID, task.TaskName)
	return nil
}

// mapAzureStatusToTaskStatus maps Azure DevOps status to internal task status
// When task completes successfully, AI analysis will be triggered separately
func (s *TaskExecutionService) mapAzureStatusToTaskStatus(statusResult *StatusResult, exec *AzureDevOpsExecutor, ctx context.Context) (models.TaskStatus, []models.TaskResult) {
	// Check if pipeline is still running
	if !exec.IsCompleted(statusResult.Status) {
		return models.TaskStatusRunning, nil
	}

	// Pipeline is complete, determine result
	// Return empty results - AI analysis will be triggered separately when task completes
	if exec.IsSuccess(statusResult.Result) {
		return models.TaskStatusPassed, nil
	}

	if statusResult.Result == TaskResultFailed {
		return models.TaskStatusFailed, nil
	}

	if statusResult.Result == TaskResultCanceled {
		return models.TaskStatusCancelled, nil
	}

	return models.TaskStatusFailed, nil
}

// getTaskResource finds the resource associated with a task
// First tries to use the saved ResourceID from AI matching, falls back to task name matching
func (s *TaskExecutionService) getTaskResource(task *models.Task) (*models.ExecutableResource, error) {
	// If task has a ResourceID from AI matching, use it directly
	if task.ResourceID > 0 {
		resource, err := s.resourceStorage.GetResource(task.ResourceID)
		if err != nil {
			return nil, fmt.Errorf("failed to get resource by ID %d: %w", task.ResourceID, err)
		}
		if resource == nil {
			return nil, fmt.Errorf("resource with ID %d not found", task.ResourceID)
		}
		return resource, nil
	}

	// Fallback: Map task name to resource type (for old tasks without ResourceID)
	var resourceType models.ResourceType
	switch task.TaskName {
	case "basic_ci_all":
		resourceType = models.ResourceTypeBasicCIAll
	case "deployment_deployment":
		resourceType = models.ResourceTypeDeployment
	case "specialized_tests_api_test":
		resourceType = models.ResourceTypeAPITest
	case "specialized_tests_module_e2e":
		resourceType = models.ResourceTypeModuleE2E
	case "specialized_tests_agent_e2e":
		resourceType = models.ResourceTypeAgentE2E
	case "specialized_tests_ai_e2e":
		resourceType = models.ResourceTypeAIE2E
	default:
		return nil, fmt.Errorf("unknown task type: %s", task.TaskName)
	}

	// Get all resources and find matching one
	resources, err := s.resourceStorage.GetAllResources()
	if err != nil {
		return nil, err
	}

	for _, resource := range resources {
		if resource.ResourceType == resourceType {
			return resource, nil
		}
	}

	return nil, fmt.Errorf("no resource found for type: %s", resourceType)
}

// FetchAndStoreLogs fetches logs from Azure DevOps and stores them in a directory
// Returns the directory path containing all log files
func (s *TaskExecutionService) FetchAndStoreLogs(ctx context.Context, task *models.Task) (string, error) {
	if task.BuildID == 0 {
		return "", fmt.Errorf("task has no build ID")
	}

	// Get the resource associated with this task
	resource, err := s.getTaskResource(task)
	if err != nil {
		return "", fmt.Errorf("failed to get task resource: %w", err)
	}

	// Get Azure PAT from config storage
	azureConfig, err := s.configStorage.GetAzureConfig()
	if err != nil {
		return "", fmt.Errorf("failed to get Azure config: %w", err)
	}

	// Build executor config from resource
	execConfig := &ExecutorConfig{
		Organization: resource.Organization,
		Project:      resource.Project,
		PAT:          azureConfig.PAT,
		Branch:       "refs/heads/main",
		VerifySSL:    true,
	}

	// Create executor
	exec := NewAzureDevOpsExecutor(execConfig)

	// Create log parser
	parser := NewLogParser(exec)

	// Fetch and store logs (now returns a directory path)
	tempDir := "/tmp/event_processor_logs"
	logDirPath, err := parser.FetchAndStoreLogs(ctx, task.BuildID, tempDir)
	if err != nil {
		return "", fmt.Errorf("failed to fetch and store logs: %w", err)
	}

	log.Printf("[TaskExecutionService] Logs stored in directory: %s", logDirPath)

	// Update task with log directory path
	task.LogFilePath = logDirPath
	if err := s.taskStorage.UpdateTask(task); err != nil {
		log.Printf("Failed to update task with log directory path: %v", err)
	}

	return logDirPath, nil
}

// AnalyzeLogs analyzes logs using AI and returns structured results
// Deprecated: Use FetchAndAnalyzeLogs which handles directory-based analysis
func (s *TaskExecutionService) AnalyzeLogs(ctx context.Context, task *models.Task, logContent string) ([]models.TaskResult, error) {
	log.Printf("[TaskExecutionService] Analyzing logs for task %s (task_id: %d, build_id: %d)", task.TaskName, task.ID, task.BuildID)
	log.Printf("[TaskExecutionService] Log content length: %d bytes", len(logContent))

	// Use AI log analyzer for parsing
	results, err := s.logAnalyzer.AnalyzeLogs(logContent)
	if err != nil {
		log.Printf("[TaskExecutionService] AI analysis failed: %v", err)
		return nil, fmt.Errorf("failed to analyze logs with AI: %w", err)
	}

	log.Printf("[TaskExecutionService] AI analysis completed successfully, returned %d results", len(results))

	// Log summary of results
	passCount := 0
	failCount := 0
	for _, r := range results {
		if r.Result == "pass" {
			passCount++
		} else if r.Result == "fail" {
			failCount++
		}
	}
	log.Printf("[TaskExecutionService] Analysis summary: %d passed, %d failed, %d total", passCount, failCount, len(results))

	return results, nil
}

// FetchAndAnalyzeLogs fetches logs from Azure DevOps and analyzes them
// Each stage's log is analyzed separately, then results are merged
// Log directories are retained based on the configured retention period
// If logs already exist locally (in task.LogFilePath), skips re-downloading
func (s *TaskExecutionService) FetchAndAnalyzeLogs(ctx context.Context, task *models.Task) ([]models.TaskResult, error) {
	// Try to start AI analysis using atomic CAS operation
	// This MUST be done BEFORE fetching logs to prevent duplicate log storage
	// from both runTaskExecutor and monitor
	if s.taskStorage != nil {
		started, err := s.taskStorage.TryStartAnalysisOrResetStale(task.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to try start analysis: %w", err)
		}
		if !started {
			// Another goroutine already started the analysis
			// Wait for it to complete and return the results
			log.Printf("[TaskExecutionService] AI analysis already in progress for task_id=%d, waiting for completion", task.ID)

			// Poll for results from the other goroutine
			for i := 0; i < 60; i++ { // Wait up to 5 minutes (60 * 5 seconds)
				time.Sleep(5 * time.Second)

				// Check if task status changed (no longer running = analysis done)
				updatedTask, err := s.taskStorage.GetTask(task.ID)
				if err == nil && updatedTask != nil {
					// If task is no longer running, analysis should be complete
					if updatedTask.Status != models.TaskStatusRunning {
						// Try to get the saved results
						results, _ := s.taskStorage.GetTaskResults(task.ID)
						if len(results) > 0 {
							log.Printf("[TaskExecutionService] Retrieved %d results from completed analysis for task_id=%d", len(results), task.ID)
							return results, nil
						}
						// No results found, task completed without AI analysis
						log.Printf("[TaskExecutionService] Analysis completed but no results found for task_id=%d", task.ID)
						return nil, fmt.Errorf("analysis completed but no results available")
					}
				}

				// Also check if analyzing flag was reset (analysis failed/aborted)
				analyzing, err := s.taskStorage.IsTaskAnalyzing(task.ID)
				if err == nil && !analyzing {
					log.Printf("[TaskExecutionService] Analysis flag reset, trying to get results for task_id=%d", task.ID)
					results, _ := s.taskStorage.GetTaskResults(task.ID)
					if len(results) > 0 {
						return results, nil
					}
					// Flag reset but no results, analysis failed
					return nil, fmt.Errorf("analysis failed")
				}
			}

			log.Printf("[TaskExecutionService] Wait timeout for task_id=%d analysis", task.ID)
			return nil, fmt.Errorf("waited too long for analysis to complete")
		}
		log.Printf("[TaskExecutionService] Acquired analysis lock for task_id=%d, build_id=%d", task.ID, task.BuildID)
	}

	// Reload task to get the latest LogFilePath in case another goroutine updated it
	// This ensures we use the most recently fetched logs
	updatedTask, err := s.taskStorage.GetTask(task.ID)
	if err == nil && updatedTask != nil && updatedTask.LogFilePath != "" {
		task.LogFilePath = updatedTask.LogFilePath
	}

	var logDirPath string

	// Check if logs already exist locally
	if task.LogFilePath != "" {
		// Verify the log directory still exists and contains log files
		if _, err := os.Stat(task.LogFilePath); err == nil {
			// Check if directory contains log files
			entries, err := os.ReadDir(task.LogFilePath)
			if err == nil {
				hasLogs := false
				for _, entry := range entries {
					if !entry.IsDir() && strings.HasPrefix(entry.Name(), "log_") && strings.HasSuffix(entry.Name(), ".txt") {
						hasLogs = true
						break
					}
				}
				if hasLogs {
					log.Printf("[TaskExecutionService] Using existing logs from directory: %s", task.LogFilePath)
					logDirPath = task.LogFilePath
				}
			}
		}
	}

	// If no existing logs found, fetch from Azure DevOps
	if logDirPath == "" {
		log.Printf("[TaskExecutionService] No existing logs found, fetching from Azure DevOps")
		var err error
		logDirPath, err = s.FetchAndStoreLogs(ctx, task)
		if err != nil {
			return nil, err
		}
	}

	log.Printf("[TaskExecutionService] Analyzing logs from directory: %s", logDirPath)

	// Analyze all log files in the directory
	results, err := s.logAnalyzer.AnalyzeLogDirectory(logDirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze logs: %w", err)
	}

	log.Printf("[TaskExecutionService] Log analysis completed, directory retained: %s", logDirPath)

	return results, nil
}

// CleanupOldLogs removes log files older than the configured retention period
func (s *TaskExecutionService) CleanupOldLogs() error {
	// Get retention period from config
	retentionDays, err := s.configStorage.GetLogRetentionDays()
	if err != nil {
		log.Printf("[TaskExecutionService] Failed to get log retention days: %v", err)
		return err
	}

	log.Printf("[TaskExecutionService] Cleaning up log directories older than %d days", retentionDays)

	tempDir := "/tmp/event_processor_logs"

	// Check if directory exists
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		log.Printf("[TaskExecutionService] Log directory does not exist: %s", tempDir)
		return nil
	}

	// Read all entries in the directory
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return fmt.Errorf("failed to read log directory: %w", err)
	}

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	deletedCount := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Only process build_* directories
		if !strings.HasPrefix(entry.Name(), "build_") {
			continue
		}

		dirPath := filepath.Join(tempDir, entry.Name())
		dirInfo, err := os.Stat(dirPath)
		if err != nil {
			log.Printf("[TaskExecutionService] Failed to stat directory %s: %v", dirPath, err)
			continue
		}

		// Delete directories older than retention period
		if dirInfo.ModTime().Before(cutoffTime) {
			if err := os.RemoveAll(dirPath); err != nil {
				log.Printf("[TaskExecutionService] Failed to delete old log directory %s: %v", dirPath, err)
			} else {
				log.Printf("[TaskExecutionService] Deleted old log directory: %s (modified: %s)", dirPath, dirInfo.ModTime().Format(time.RFC3339))
				deletedCount++
			}
		}
	}

	log.Printf("[TaskExecutionService] Cleanup completed: deleted %d directories", deletedCount)
	return nil
}

