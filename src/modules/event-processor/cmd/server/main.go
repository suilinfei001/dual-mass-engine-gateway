package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github-hub/event-processor/internal/ai"
	"github-hub/event-processor/internal/api"
	"github-hub/event-processor/internal/executor"
	"github-hub/event-processor/internal/mock"
	"github-hub/event-processor/internal/models"
	"github-hub/event-processor/internal/monitor"
	"github-hub/event-processor/internal/scheduler"
	"github-hub/event-processor/internal/storage"
)

const (
	EventFetchInterval = 30 * time.Second
	TaskCheckInterval  = 5 * time.Second
	MockServerPort     = 8090
	APIServerPort      = "5002"
	DefaultDB          = "root:root123456@tcp(localhost:3307)/event_processor?parseTime=true"
)

func getDBConnString() string {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if dbHost != "" && dbPort != "" && dbUser != "" && dbPassword != "" && dbName != "" {
		return dbUser + ":" + dbPassword + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName + "?parseTime=true"
	}

	return DefaultDB
}

func main() {
	dbDSN := flag.String("db", getDBConnString(), "MySQL数据库连接字符串")
	mockPort := flag.Int("mock-port", MockServerPort, "Mock服务端口")
	apiPort := flag.String("api-port", APIServerPort, "API服务端口")
	flag.Parse()

	log.SetFlags(0)
	log.Println("Starting Event Processor...")

	store, err := storage.NewMySQLTaskStorage(*dbDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer store.Close()
	log.Println("Database connected successfully")

	userStore := storage.NewMySQLUserStorage(store.DB())
	resourceStore := storage.NewMySQLResourceStorage(store.DB())
	configStore := storage.NewMySQLConfigStorage(store.DB())

	eventReceiverIP, err := configStore.GetEventReceiverIP()
	if err != nil || eventReceiverIP == "" {
		// Use default Event Receiver IP if not configured
		eventReceiverIP = "http://10.4.111.141:5001"
		log.Printf("Event Receiver IP not configured, using default: %s", eventReceiverIP)
		// Try to save the default value to config
		if err := configStore.SetEventReceiverIP(eventReceiverIP); err != nil {
			log.Printf("Warning: Failed to save default Event Receiver IP: %v", err)
		}
	}
	log.Printf("Event Receiver API: %s", eventReceiverIP)

	client := api.NewClientWithURL(eventReceiverIP)

	mockServer := mock.NewMockServer(*mockPort)
	go func() {
		log.Printf("Starting Mock server on port %d", *mockPort)
		if err := mockServer.Start(); err != nil {
			log.Printf("Mock server error: %v", err)
		}
	}()

	sched := scheduler.NewSchedulerWithStorage(client, store, resourceStore, ai.NewAIMatcher(configStore))

	// Create TaskExecutionService for Azure DevOps integration
	taskExecutionService := executor.NewTaskExecutionService(configStore, resourceStore, store, client)

	// Create monitor with executor service for Azure integration
	mon := monitor.NewMonitorWithExecutor(sched, taskExecutionService)
	mon.Start()

	storageAdapter := storage.NewTaskStorageAdapter(store)
	apiServer := api.NewServerWithResources(*apiPort, storageAdapter, userStore, resourceStore, configStore, client, taskExecutionService, store)
	go func() {
		log.Printf("Starting API server on port %s", *apiPort)
		if err := apiServer.Start(); err != nil {
			log.Printf("API server error: %v", err)
		}
	}()

	ctx, stop := setupSignalHandler()
	defer stop()

	eventFetcher := NewEventFetcher(client, sched)
	go eventFetcher.Start(ctx)

	go runTaskExecutor(ctx, sched, mon, taskExecutionService)

	// Start periodic log cleanup (runs daily)
	go runLogCleanup(ctx, taskExecutionService)

	log.Println("Event Processor is running. Press Ctrl+C to stop.")

	<-ctx.Done()

	log.Println("Shutting down...")
	mon.Stop()
	mockServer.Stop()
}

func setupSignalHandler() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received termination signal")
		cancel()
	}()

	return ctx, cancel
}

type EventFetcher struct {
	client    *api.Client
	scheduler *scheduler.SchedulerWithStorage
}

func NewEventFetcher(client *api.Client, sched *scheduler.SchedulerWithStorage) *EventFetcher {
	return &EventFetcher{
		client:    client,
		scheduler: sched,
	}
}

func (ef *EventFetcher) Start(ctx context.Context) {
	ticker := time.NewTicker(EventFetchInterval)
	defer ticker.Stop()

	ef.fetchAndProcess()

	for {
		select {
		case <-ctx.Done():
			log.Println("Event fetcher stopped")
			return
		case <-ticker.C:
			ef.fetchAndProcess()
		}
	}
}

func (ef *EventFetcher) fetchAndProcess() {
	log.Println("Fetching events...")

	events, err := ef.scheduler.FetchEvents()
	if err != nil {
		log.Printf("Failed to fetch events: %v", err)
		return
	}

	log.Printf("Fetched %d events", len(events))

	if err := ef.scheduler.ProcessEvents(events); err != nil {
		log.Printf("Failed to process events: %v", err)
		return
	}
}

func runTaskExecutor(ctx context.Context, sched *scheduler.SchedulerWithStorage, mon *monitor.Monitor, execService *executor.TaskExecutionService) {
	ticker := time.NewTicker(TaskCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Task executor stopped")
			return
		case <-ticker.C:
			go executeNextTask(sched, mon, execService)
		}
	}
}

func executeNextTask(sched *scheduler.SchedulerWithStorage, mon *monitor.Monitor, execService *executor.TaskExecutionService) {
	task, _, err := sched.GetNextPendingTask()
	if err != nil {
		log.Printf("Failed to get next task: %v", err)
		return
	}

	if task == nil {
		return
	}

	// 检查 Task 是否处于 no-resource 状态，如果是则跳过执行
	if task.Status == models.TaskStatusNoResource {
		log.Printf("Task %s (event_id: %d) is in no-resource status, skipping execution and calling CompleteTask directly",
			task.TaskName, task.EventID)
		sched.CompleteTask(task, nil)
		return
	}

	log.Printf("Executing task: %s (event_id: %d, order: %d, status: %s)",
		task.TaskName, task.EventID, task.ExecuteOrder, task.Status)

	// Try to mark task as running using CAS operation
	// This prevents multiple goroutines from processing the same task
	started, err := sched.TryMarkTaskRunning(task)
	if err != nil {
		log.Printf("Failed to try mark task running: %v", err)
		return
	}
	if !started {
		// Another goroutine already took this task
		log.Printf("Task %s (event_id: %d) already taken by another goroutine, skipping", task.TaskName, task.EventID)
		return
	}

	// Check if the task's resource allows skipping (before executing)
	// This applies to both Azure and non-Azure tasks
	wasSkipped, err := execService.CheckAndSkipIfNeeded(task)
	if err != nil {
		log.Printf("Failed to check skip status for task: %v", err)
		// Continue with normal execution if skip check fails
	} else if wasSkipped {
		// Task was marked as skipped, skip it and move to next task
		log.Printf("Task %s was skipped, skipping and moving to next task", task.TaskName)
		reason := "Resource allows skipping"
		if err := sched.SkipTask(task, reason); err != nil {
			log.Printf("Failed to skip task %s: %v", task.TaskName, err)
		}
		return
	}

	// Check if this is an Azure task
	isAzureTask := executor.IsAzureURL(task.RequestURL)

	// 对于 deployment_deployment 任务，需要先获取 Testbed
	if task.TaskName == "deployment_deployment" {
		log.Printf("[executeNextTask] Deployment task requires testbed, acquiring...")
		for {
			// 检查任务状态，如果已经不是 running 则退出
			currentTask, err := sched.GetTaskByID(task.ID)
			if err != nil {
				log.Printf("[executeNextTask] Failed to get task status: %v", err)
			} else if currentTask.Status != models.TaskStatusRunning {
				log.Printf("[executeNextTask] Task status changed to %s, stopping testbed acquisition", currentTask.Status)
				return
			}

			if err := sched.AcquireTestbedForDeployment(task); err != nil {
				log.Printf("[executeNextTask] Failed to acquire testbed: %v, retrying in 1 minute...", err)
				time.Sleep(1 * time.Minute)
				continue
			}
			break
		}
		log.Printf("[executeNextTask] Testbed acquired: ip=%s, ssh_user=%s",
			task.TestbedIP, task.SSHUser)
	}

	var results []models.TaskResult

	if isAzureTask {
		// Use TaskExecutionService for Azure tasks
		log.Printf("Executing Azure task: %s", task.RequestURL)
		if err := execService.ExecuteTask(context.Background(), task, nil); err != nil {
			log.Printf("Azure task execution failed: %v", err)
			sched.FailTask(task, err.Error())
			return
		}
		// Task is already updated by ExecuteTask
		// Check if task was marked as skipped (e.g., resource allows skip)
		if task.Status == models.TaskStatusSkipped {
			log.Printf("Task %s was skipped, skipping and moving to next task", task.TaskName)
			reason := "Resource allows skipping"
			if err := sched.SkipTask(task, reason); err != nil {
				log.Printf("Failed to skip task %s: %v", task.TaskName, err)
			}
			// Reset analyzing flag if applicable
			if task.BuildID > 0 {
				sched.UpdateTaskAnalyzing(task.ID, false)
			}
			return
		}
	} else {
		// Use mock execution for non-Azure tasks
		results, err = monitor.ExecuteTask(task)
		if err != nil {
			log.Printf("Task execution failed: %v", err)
			sched.FailTask(task, err.Error())
			return
		}

		if task.TaskID != "" {
			if err := sched.UpdateTask(task); err != nil {
				log.Printf("Failed to update task TaskID: %v", err)
			}
		}
	}

	for {
		time.Sleep(5 * time.Second)

		// 检查数据库中的任务状态，如果已经被其他地方（如 monitor）标记为非 running 状态则退出
		currentTask, err := sched.GetTaskByID(task.ID)
		if err != nil {
			log.Printf("[executeNextTask] Failed to get task status from database: %v", err)
		} else if currentTask.Status != models.TaskStatusRunning {
			log.Printf("[executeNextTask] Task status in database changed to %s, stopping polling", currentTask.Status)
			return
		}

		status, queryResults, err := mon.QueryTaskStatus(task)
		if err != nil {
			log.Printf("Failed to check task status: %v", err)
			continue
		}

		switch status {
		case models.TaskStatusPassed:
			log.Printf("Task completed successfully: %s", task.TaskName)

			// For Azure tasks (BuildID > 0), fetch and analyze logs BEFORE completing task
			// This ensures quality checks are updated with actual AI analysis results
			var completeResults []models.TaskResult
			if task.BuildID > 0 {
				log.Printf("Fetching and analyzing logs for Azure task: build_id=%d, task_id=%d", task.BuildID, task.ID)
				analyzeResults, err := execService.FetchAndAnalyzeLogs(context.Background(), task)
				if err != nil {
					// FetchAndAnalyzeLogs now waits for concurrent analysis to complete
					// If it returns an error, it's a real error (analysis failed or timeout)
					log.Printf("Failed to fetch/analyze logs: %v", err)
				} else {
					log.Printf("Log analysis completed successfully")
					if analyzeResults != nil {
						// Save analysis results to database
						if err := sched.SaveTaskResults(task.ID, analyzeResults); err != nil {
							log.Printf("Failed to save analysis results: %v", err)
						} else {
							log.Printf("Analysis results saved for task_id=%d", task.ID)
						}
						completeResults = analyzeResults
					}
				}
			}

			// Use the analysis results for Azure tasks, otherwise use query/mock results
			if completeResults != nil {
				sched.CompleteTask(task, completeResults)
			} else if queryResults != nil {
				sched.CompleteTask(task, queryResults)
			} else {
				sched.CompleteTask(task, results)
			}

			// Reset analyzing flag after task is completed to allow manual re-analysis
			if task.BuildID > 0 {
				sched.UpdateTaskAnalyzing(task.ID, false)
			}
			return

		case models.TaskStatusFailed:
			log.Printf("Task failed: %s", task.TaskName)

			// For Azure tasks (BuildID > 0), fetch and analyze logs BEFORE marking as failed
			// This helps get error details from the logs
			if task.BuildID > 0 {
				log.Printf("Fetching and analyzing logs for failed Azure task: build_id=%d, task_id=%d", task.BuildID, task.ID)
				analyzeResults, err := execService.FetchAndAnalyzeLogs(context.Background(), task)
				if err != nil {
					// FetchAndAnalyzeLogs now waits for concurrent analysis to complete
					log.Printf("Failed to fetch/analyze logs for failed task: %v", err)
				} else {
					log.Printf("Log analysis completed for failed task")
					if analyzeResults != nil {
						// Save analysis results to database
						if err := sched.SaveTaskResults(task.ID, analyzeResults); err != nil {
							log.Printf("Failed to save analysis results: %v", err)
						}
					}
				}
			}

			sched.FailTask(task, "Task execution failed")

			// Reset analyzing flag after task is failed to allow manual re-analysis
			if task.BuildID > 0 {
				sched.UpdateTaskAnalyzing(task.ID, false)
			}
			return

		case models.TaskStatusTimeout:
			log.Printf("Task timed out: %s", task.TaskName)
			sched.TimeoutTask(task, "Task execution timed out")
			return

		case models.TaskStatusCancelled:
			log.Printf("Task cancelled: %s", task.TaskName)
			sched.CancelTask(task, "Task was cancelled")
			return

		case models.TaskStatusNoResource:
			log.Printf("Task has no resource: %s", task.TaskName)
			sched.CompleteTask(task, nil)
			return

		case models.TaskStatusRunning:
			log.Printf("Task still running: %s (task_id: %s, build_id: %d)", task.TaskName, task.TaskID, task.BuildID)
		}
	}
}

// runLogCleanup runs periodic log cleanup task
// Cleans up log files older than the configured retention period
func runLogCleanup(ctx context.Context, execService *executor.TaskExecutionService) {
	// Run cleanup once at startup (after a short delay)
	go func() {
		time.Sleep(1 * time.Minute)
		if err := execService.CleanupOldLogs(); err != nil {
			log.Printf("Initial log cleanup failed: %v", err)
		}
	}()

	// Run cleanup daily
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Log cleanup stopped")
			return
		case <-ticker.C:
			if err := execService.CleanupOldLogs(); err != nil {
				log.Printf("Scheduled log cleanup failed: %v", err)
			}
		}
	}
}
