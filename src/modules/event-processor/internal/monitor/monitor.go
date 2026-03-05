package monitor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github-hub/event-processor/internal/executor"
	"github-hub/event-processor/internal/models"
)

const (
	TaskTimeout     = 60 * time.Minute
	PollingInterval = 30 * time.Second
)

type Scheduler interface {
	GetRunningTasks() ([]*models.Task, error)
	CompleteTask(task *models.Task, results []models.TaskResult) error
	FailTask(task *models.Task, reason string) error
	TimeoutTask(task *models.Task, reason string) error
	CancelTask(task *models.Task, reason string) error
	SaveTaskResults(taskID int, results []models.TaskResult) error
	UpdateTask(task *models.Task) error
}

type Monitor struct {
	scheduler         Scheduler
	ctx               context.Context
	cancel            context.CancelFunc
	client            *http.Client
	executionService  *executor.TaskExecutionService
	useAzureExecution bool
}

func NewMonitor(scheduler Scheduler) *Monitor {
	ctx, cancel := context.WithCancel(context.Background())
	return &Monitor{
		scheduler: scheduler,
		ctx:       ctx,
		cancel:    cancel,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		useAzureExecution: false,
	}
}

// NewMonitorWithExecutor creates a new Monitor with TaskExecutionService for Azure execution
func NewMonitorWithExecutor(scheduler Scheduler, executionService *executor.TaskExecutionService) *Monitor {
	ctx, cancel := context.WithCancel(context.Background())
	return &Monitor{
		scheduler: scheduler,
		ctx:       ctx,
		cancel:    cancel,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		executionService:  executionService,
		useAzureExecution: true,
	}
}

func (m *Monitor) Start() {
	log.Println("Starting task monitor...")

	go m.monitorRunningTasks()
}

func (m *Monitor) Stop() {
	log.Println("Stopping task monitor...")
	m.cancel()
}

func (m *Monitor) monitorRunningTasks() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			log.Println("Task monitor stopped")
			return
		case <-ticker.C:
			m.checkRunningTasks()
		}
	}
}

func (m *Monitor) checkRunningTasks() {
	tasks, err := m.scheduler.GetRunningTasks()
	if err != nil {
		log.Printf("Failed to get running tasks: %v", err)
		return
	}

	for _, task := range tasks {
		m.checkTaskStatus(task)
	}
}

func (m *Monitor) checkTaskStatus(task *models.Task) {
	if task.StartTime == nil {
		return
	}

	if time.Since(task.StartTime.Time) > TaskTimeout {
		log.Printf("Task %s (event_id: %d) timed out", task.TaskName, task.EventID)
		if err := m.scheduler.TimeoutTask(task, "Task execution timed out"); err != nil {
			log.Printf("Failed to mark task as timeout: %v", err)
		}
		return
	}

	status, results, err := m.QueryTaskStatus(task)
	if err != nil {
		log.Printf("Failed to query task status: %v", err)
		return
	}

	switch status {
	case models.TaskStatusPassed:
		log.Printf("Task %s completed successfully", task.TaskName)
		// For basic_ci_all, executor (runTaskExecutor) handles the full flow including AI analysis
		// Monitor should NOT process basic_ci_all to avoid duplicate AI analysis calls
		if task.TaskName == "basic_ci_all" && m.executionService != nil && task.BuildID > 0 {
			// Skip processing - executor will handle AI analysis and task completion
			log.Printf("[Monitor] Skipping basic_ci_all task - executor handles AI analysis")
		} else {
			// For other tasks, complete immediately
			if err := m.scheduler.CompleteTask(task, results); err != nil {
				log.Printf("Failed to complete task: %v", err)
			}
			// Trigger AI analysis for completed tasks with build_id
			if m.executionService != nil && task.BuildID > 0 {
				go m.analyzeAndSaveResults(task)
			}
		}
	case models.TaskStatusFailed:
		log.Printf("Task %s failed", task.TaskName)
		if err := m.scheduler.FailTask(task, "Task execution failed"); err != nil {
			log.Printf("Failed to fail task: %v", err)
		}
		// Also trigger AI analysis for failed tasks to get error details
		if m.executionService != nil && task.BuildID > 0 {
			go m.analyzeAndSaveResults(task)
		}
	case models.TaskStatusCancelled:
		log.Printf("Task %s cancelled", task.TaskName)
		if err := m.scheduler.CancelTask(task, "Task was cancelled"); err != nil {
			log.Printf("Failed to cancel task: %v", err)
		}
	case models.TaskStatusRunning:
		log.Printf("Task %s still running", task.TaskName)
	default:
		log.Printf("Task %s has unknown status: %s", task.TaskName, status)
	}
}

func (m *Monitor) QueryTaskStatus(task *models.Task) (models.TaskStatus, []models.TaskResult, error) {
	// If using Azure execution service and task has a build ID, query Azure DevOps
	if m.useAzureExecution && m.executionService != nil && task.BuildID > 0 {
		return m.executionService.PollTaskStatus(context.Background(), task)
	}

	// Otherwise, fall back to mock server for backward compatibility
	if task.RequestURL == "" {
		return models.TaskStatusRunning, nil, nil
	}

	statusURL := fmt.Sprintf("http://localhost:8090/mock/status/%s", task.TaskID)

	resp, err := m.client.Get(statusURL)
	if err != nil {
		return models.TaskStatusRunning, nil, fmt.Errorf("failed to query status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return models.TaskStatusRunning, nil, nil
	}

	var mockTask struct {
		TaskID  string       `json:"task_id"`
		Status  string       `json:"status"`
		Results []MockResult `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&mockTask); err != nil {
		return models.TaskStatusRunning, nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var results []models.TaskResult
	for _, r := range mockTask.Results {
		results = append(results, models.TaskResult{
			CheckType: r.CheckType,
			Result:    r.Result,
			Extra:     r.Extra,
		})
	}

	switch mockTask.Status {
	case "pass":
		return models.TaskStatusPassed, results, nil
	case "fail":
		return models.TaskStatusFailed, results, nil
	case "cancelled":
		return models.TaskStatusCancelled, results, nil
	case "running":
		return models.TaskStatusRunning, nil, nil
	default:
		return models.TaskStatusRunning, nil, nil
	}
}

type MockResult struct {
	CheckType string                 `json:"check_type"`
	Result    string                 `json:"result"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

// ExecuteTask executes a task using Azure or mock execution
func ExecuteTask(task *models.Task) ([]models.TaskResult, error) {
	log.Printf("Executing task: %s (event_id: %d)", task.TaskName, task.EventID)

	if task.RequestURL == "" {
		return MockExecuteTask(task)
	}

	// Check if this is an Azure URL
	if executor.IsAzureURL(task.RequestURL) {
		log.Printf("Detected Azure URL: %s", task.RequestURL)
		// For Azure URLs, the execution is handled by TaskExecutionService
		// Return empty results - the task will be polled for status
		return nil, nil
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(task.RequestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to execute task: %w", err)
	}
	defer resp.Body.Close()

	var mockTask struct {
		TaskID  string       `json:"task_id"`
		Status  string       `json:"status"`
		Results []MockResult `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&mockTask); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	task.TaskID = mockTask.TaskID

	var results []models.TaskResult
	for _, r := range mockTask.Results {
		results = append(results, models.TaskResult{
			CheckType: r.CheckType,
			Result:    r.Result,
			Extra:     r.Extra,
		})
	}

	return results, nil
}

// CancelTask cancels a running task
func CancelTask(task *models.Task) error {
	// CancelledURL was removed - cancellation now handled through executor service
	if task.BuildID == 0 {
		log.Printf("Task %s has no build ID to cancel", task.TaskName)
		return nil
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("POST", "http://localhost:8090/mock/cancel", bytes.NewBuffer([]byte{}))
	if err != nil {
		return fmt.Errorf("failed to create cancel request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to cancel task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cancel request failed with status: %d", resp.StatusCode)
	}

	log.Printf("Task %s cancelled successfully", task.TaskName)
	return nil
}

func MockExecuteTask(task *models.Task) ([]models.TaskResult, error) {
	log.Printf("Mock executing task: %s (event_id: %d)", task.TaskName, task.EventID)

	time.Sleep(2 * time.Second)

	switch task.TaskName {
	case "basic_ci_all":
		return []models.TaskResult{
			{CheckType: "compilation", Result: "pass"},
			{CheckType: "code_lint", Result: "pass"},
			{CheckType: "security_scan", Result: "pass"},
			{CheckType: "unit_test", Result: "pass", Extra: map[string]interface{}{"score": 95}},
		}, nil
	case "deployment_deployment":
		return []models.TaskResult{
			{CheckType: "deployment", Result: "pass", Extra: map[string]interface{}{
				"node_ip":   "192.168.1.100",
				"node_port": "22",
				"node_user": "deployer",
			}},
		}, nil
	case "specialized_tests_api_test":
		return []models.TaskResult{
			{CheckType: "api_test", Result: "pass"},
		}, nil
	case "specialized_tests_module_e2e":
		return []models.TaskResult{
			{CheckType: "module_e2e", Result: "pass"},
		}, nil
	case "specialized_tests_agent_e2e":
		return []models.TaskResult{
			{CheckType: "agent_e2e", Result: "pass"},
		}, nil
	case "specialized_tests_ai_e2e":
		return []models.TaskResult{
			{CheckType: "ai_e2e", Result: "pass"},
		}, nil
	default:
		return nil, fmt.Errorf("unknown task type: %s", task.TaskName)
	}
}

// analyzeAndSaveResults fetches logs, analyzes them with AI, and saves the results
// This runs asynchronously after a task completes
func (m *Monitor) analyzeAndSaveResults(task *models.Task) {
	log.Printf("[Monitor] Starting AI analysis for task %s (build_id: %d)", task.TaskName, task.BuildID)

	results, err := m.executionService.FetchAndAnalyzeLogs(context.Background(), task)
	if err != nil {
		// FetchAndAnalyzeLogs now waits for concurrent analysis to complete
		// If it returns an error, it's a real error (analysis failed or timeout)
		log.Printf("[Monitor] AI analysis failed for task %s: %v", task.TaskName, err)
		return
	}

	log.Printf("[Monitor] AI analysis completed for task %s, got %d results", task.TaskName, len(results))

	// Save results to storage
	if m.executionService != nil {
		// Use the scheduler's SaveTaskResults method
		if err := m.scheduler.SaveTaskResults(task.ID, results); err != nil {
			log.Printf("[Monitor] Failed to save task results for task %s: %v", task.TaskName, err)
		} else {
			log.Printf("[Monitor] Saved %d results for task %s", len(results), task.TaskName)
		}
	}
}

// analyzeAndCompleteTask fetches logs, analyzes them with AI, saves results, and completes the task
// This is used for basic_ci_all tasks to ensure AI analysis completes before task is marked as done
func (m *Monitor) analyzeAndCompleteTask(task *models.Task) {
	log.Printf("[Monitor] Starting AI analysis for basic_ci_all task %s (build_id: %d)", task.TaskName, task.BuildID)

	results, err := m.executionService.FetchAndAnalyzeLogs(context.Background(), task)
	if err != nil {
		// FetchAndAnalyzeLogs now waits for concurrent analysis to complete
		// If it returns an error, it's a real error (analysis failed or timeout)
		log.Printf("[Monitor] AI analysis failed for basic_ci_all task %s: %v", task.TaskName, err)
		// If analysis fails, still complete the task but with no results
		// Note: Don't reset Analyzing flag - CompleteTask will change status to passed/failed
		if err := m.scheduler.CompleteTask(task, nil); err != nil {
			log.Printf("[Monitor] Failed to complete basic_ci_all task: %v", err)
		}
		return
	}

	log.Printf("[Monitor] AI analysis completed for basic_ci_all task %s, got %d results", task.TaskName, len(results))

	// Save results to storage
	if err := m.scheduler.SaveTaskResults(task.ID, results); err != nil {
		log.Printf("[Monitor] Failed to save task results for basic_ci_all task %s: %v", task.TaskName, err)
	}

	// Now complete the task with the AI analysis results
	// CompleteTask will change status to "passed", so the task won't be returned by GetRunningTasks anymore
	if err := m.scheduler.CompleteTask(task, results); err != nil {
		log.Printf("[Monitor] Failed to complete basic_ci_all task: %v", err)
	} else {
		log.Printf("[Monitor] Completed basic_ci_all task %s with AI analysis results", task.TaskName)
	}
}
