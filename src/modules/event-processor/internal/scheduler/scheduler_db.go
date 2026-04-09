package scheduler

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github-hub/event-processor/internal/api"
	"github-hub/event-processor/internal/models"
	"github-hub/event-processor/internal/storage"
)

type SchedulerWithStorage struct {
	client             *api.Client
	creator            *TaskCreator
	storage            storage.TaskStorage
	resourceStorage    storage.ResourceStorage
	aiMatcher          AIMatcherInterface
	prHandler          *PRHandler
	eventCache         map[int]*api.Event
	eventCacheMu       sync.RWMutex
	resourcePoolClient *api.ResourcePoolClient
}

func NewSchedulerWithStorage(
	client *api.Client,
	store storage.TaskStorage,
	resourceStorage storage.ResourceStorage,
	aiMatcher AIMatcherInterface,
) *SchedulerWithStorage {
	s := &SchedulerWithStorage{
		client:             client,
		creator:            NewTaskCreator(client, store, resourceStorage, aiMatcher),
		storage:            store,
		resourceStorage:    resourceStorage,
		aiMatcher:          aiMatcher,
		eventCache:         make(map[int]*api.Event),
		resourcePoolClient: api.NewResourcePoolClient(),
	}
	s.prHandler = NewPRHandler(client, s.creator, s)
	return s
}

func (s *SchedulerWithStorage) FetchEvents() ([]api.Event, error) {
	events, err := s.client.GetEvents()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch events: %w", err)
	}

	s.eventCacheMu.Lock()
	for i := range events {
		event := events[i]
		s.eventCache[event.ID] = &event
	}
	s.eventCacheMu.Unlock()

	return events, nil
}

func (s *SchedulerWithStorage) refreshEventCache(eventID int) (*api.Event, error) {
	event, err := s.client.GetEvent(eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch event %d: %w", eventID, err)
	}

	s.eventCacheMu.Lock()
	s.eventCache[eventID] = event
	s.eventCacheMu.Unlock()

	return event, nil
}

// serializeExtra 将 map[string]interface{} 序列化为 JSON 字符串
func serializeExtra(extra map[string]interface{}) string {
	if extra == nil || len(extra) == 0 {
		return ""
	}
	data, err := json.Marshal(extra)
	if err != nil {
		log.Printf("[serializeExtra] Failed to marshal extra data: %v", err)
		return ""
	}
	return string(data)
}

func (s *SchedulerWithStorage) ProcessEvents(events []api.Event) error {
	for _, event := range events {
		if s.prHandler.IsPREvent(&event) {
			if err := s.prHandler.HandlePREvent(&event); err != nil {
				log.Printf("Failed to handle PR event %d: %v", event.ID, err)
			}
			continue
		}

		// Only process events that are in "pending" status
		// Events with "processing", "completed", "failed", or "cancelled" status
		// should not be re-processed as they already have tasks created
		if event.EventStatus != "pending" {
			continue
		}

		existingTasks, err := s.storage.GetTasksByEventID(event.ID)
		if err == nil && len(existingTasks) > 0 {
			log.Printf("Tasks already exist for event %d, skipping", event.ID)
			continue
		}

		task, err := s.creator.CreateFirstTask(&event)
		if err != nil {
			log.Printf("Failed to create first task for event %d: %v", event.ID, err)
			continue
		}

		if err := s.storage.CreateTask(task); err != nil {
			log.Printf("Failed to save task %s to database: %v", task.TaskName, err)
			continue
		}

		log.Printf("Created first task %s for event %d (status: %s)", task.TaskName, event.ID, task.Status)

		// Immediately update event status to "processing" to prevent duplicate task creation
		now := time.Now()
		if err := s.client.UpdateEventStatus(event.ID, "processing", now.Format(time.RFC3339)); err != nil {
			log.Printf("Failed to update event status to processing: %v", err)
		} else {
			log.Printf("Updated event %d status to processing", event.ID)
		}

		// 如果任务状态是 no-resource，直接调用 CompleteTask 处理
		if task.Status == models.TaskStatusNoResource {
			log.Printf("Task %s for event %d is in no-resource status, calling CompleteTask directly", task.TaskName, event.ID)
			if err := s.CompleteTask(task, nil); err != nil {
				log.Printf("Failed to complete no-resource task %s for event %d: %v", task.TaskName, event.ID, err)
			}
		}
	}

	return nil
}

func (s *SchedulerWithStorage) AcquireTestbedForDeployment(task *models.Task) error {
	if task.TaskName != "deployment_deployment" {
		return nil
	}

	log.Printf("[AcquireTestbed] Attempting to acquire robot testbed for deployment task (event_id=%d)", task.EventID)

	resp, err := s.resourcePoolClient.AcquireRobotTestbed()
	if err != nil {
		return fmt.Errorf("failed to acquire testbed: %w", err)
	}

	if resp.Testbed == nil {
		return fmt.Errorf("no testbed returned from resource pool")
	}

	task.TestbedUUID = resp.Testbed.UUID
	task.TestbedIP = resp.Testbed.IPAddress
	task.SSHUser = resp.Testbed.SSHUser
	task.SSHPassword = resp.Testbed.SSHPassword
	task.AllocationUUID = resp.AllocationUUID

	log.Printf("[AcquireTestbed] Response - TestbedUUID: %s, AllocationUUID: %s, IP: %s",
		resp.Testbed.UUID, resp.AllocationUUID, resp.Testbed.IPAddress)

	if err := s.storage.UpdateTask(task); err != nil {
		return fmt.Errorf("failed to update task with testbed info: %w", err)
	}

	log.Printf("[AcquireTestbed] Successfully acquired testbed %s for task %s (event_id=%d)",
		resp.Testbed.UUID, task.TaskName, task.EventID)

	return nil
}

func (s *SchedulerWithStorage) ReleaseTestbed(task *models.Task) error {
	if task.TestbedUUID == "" && task.AllocationUUID == "" {
		log.Printf("[ReleaseTestbed] No testbed to release (both testbed_uuid and allocation_uuid are empty)")
		return nil
	}

	log.Printf("[ReleaseTestbed] Releasing testbed (testbed_uuid=%s, allocation_uuid=%s) for task %s (event_id=%d)",
		task.TestbedUUID, task.AllocationUUID, task.TaskName, task.EventID)

	// Use allocation_uuid if available, otherwise fall back to testbed_uuid
	releaseUUID := task.AllocationUUID
	if releaseUUID == "" {
		releaseUUID = task.TestbedUUID
		log.Printf("[ReleaseTestbed] WARNING: allocation_uuid is empty, using testbed_uuid: %s", releaseUUID)
	}

	if err := s.resourcePoolClient.ReleaseTestbed(releaseUUID); err != nil {
		log.Printf("[ReleaseTestbed] Failed to release testbed %s: %v", releaseUUID, err)
		return fmt.Errorf("failed to release testbed: %w", err)
	}

	log.Printf("[ReleaseTestbed] Successfully released testbed %s", task.TestbedUUID)
	return nil
}

func (s *SchedulerWithStorage) ReleaseTestbedByEventID(eventID int) error {
	tasks, err := s.storage.GetTasksByEventID(eventID)
	if err != nil {
		return fmt.Errorf("failed to get tasks for event %d: %w", eventID, err)
	}

	for _, task := range tasks {
		if task.TestbedUUID != "" {
			if err := s.ReleaseTestbed(task); err != nil {
				log.Printf("[ReleaseTestbedByEventID] Failed to release testbed for task %s: %v",
					task.TaskName, err)
			}
		}
	}

	return nil
}

func (s *SchedulerWithStorage) hasPendingQualityChecks(event api.Event) bool {
	for _, check := range event.QualityChecks {
		if check.CheckStatus == "pending" {
			return true
		}
	}
	return false
}

func (s *SchedulerWithStorage) GetNextPendingTask() (*models.Task, *api.Event, error) {
	pendingTasks, err := s.storage.GetPendingTasks()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get pending tasks: %w", err)
	}

	if len(pendingTasks) == 0 {
		return nil, nil, nil
	}

	sort.Slice(pendingTasks, func(i, j int) bool {
		return pendingTasks[i].ExecuteOrder < pendingTasks[j].ExecuteOrder
	})

	nextTask := pendingTasks[0]

	s.eventCacheMu.RLock()
	event, exists := s.eventCache[nextTask.EventID]
	s.eventCacheMu.RUnlock()

	if !exists {
		events, err := s.client.GetEvents()
		if err != nil {
			return nil, nil, fmt.Errorf("event not found in cache and failed to fetch: %w", err)
		}

		var found *api.Event
		for i := range events {
			if events[i].ID == nextTask.EventID {
				found = &events[i]
				break
			}
		}

		if found == nil {
			return nil, nil, fmt.Errorf("event not found: %d", nextTask.EventID)
		}

		s.eventCacheMu.Lock()
		s.eventCache[nextTask.EventID] = found
		s.eventCacheMu.Unlock()

		event = found
	}

	return nextTask, event, nil
}

func (s *SchedulerWithStorage) MarkTaskRunning(task *models.Task) error {
	task.MarkRunning()

	if err := s.storage.UpdateTask(task); err != nil {
		return fmt.Errorf("failed to update task in database: %w", err)
	}

	if err := s.updateQualityChecksForRunning(task); err != nil {
		log.Printf("Failed to update quality checks for running: %v", err)
	}

	if task.TaskName == "basic_ci_all" {
		now := time.Now()
		if err := s.client.UpdateEventStatus(task.EventID, "processing", now.Format(time.RFC3339)); err != nil {
			log.Printf("Failed to update event status: %v", err)
		}
	}
	return nil
}

// TryMarkTaskRunning 尝试将任务标记为运行中（原子操作）
// 只有当 status = 'pending' 时才会设置为 'running'，并返回 true 表示成功
// 如果 status 已经不是 'pending'，则返回 false 表示已被其他进程占用
func (s *SchedulerWithStorage) TryMarkTaskRunning(task *models.Task) (bool, error) {
	started, err := s.storage.TryMarkTaskRunning(task.ID)
	if err != nil {
		return false, err
	}
	if !started {
		return false, nil
	}

	// Update task object in memory
	task.MarkRunning()

	// Update quality checks
	if err := s.updateQualityChecksForRunning(task); err != nil {
		log.Printf("Failed to update quality checks for running: %v", err)
	}

	if task.TaskName == "basic_ci_all" {
		now := time.Now()
		if err := s.client.UpdateEventStatus(task.EventID, "processing", now.Format(time.RFC3339)); err != nil {
			log.Printf("Failed to update event status: %v", err)
		}
	}

	return true, nil
}

func (s *SchedulerWithStorage) CompleteTask(task *models.Task, results []models.TaskResult) error {
	log.Printf("CompleteTask called: task_name=%s, event_id=%d, task_status=%s, execute_order=%d",
		task.TaskName, task.EventID, task.Status, task.ExecuteOrder)

	if task.Status != models.TaskStatusNoResource {
		// 检查子检查项的结果，确定任务状态
		hasFailures := false
		for _, result := range results {
			if result.Result == "fail" {
				hasFailures = true
				log.Printf("Task %s has failed check: %s (event_id=%d)", task.TaskName, result.CheckType, task.EventID)
				break
			}
		}

		if hasFailures {
			task.MarkFailed("one or more quality checks failed")
			task.Results = results
			log.Printf("Task marked as failed: %s (event_id=%d)", task.TaskName, task.EventID)
		} else {
			task.MarkPassed(results)
			log.Printf("Task marked as passed: %s (event_id=%d)", task.TaskName, task.EventID)
		}
	} else {
		log.Printf("Task is in no-resource status, NOT marking as passed: %s (event_id=%d)", task.TaskName, task.EventID)
	}

	if err := s.storage.UpdateTask(task); err != nil {
		return fmt.Errorf("failed to update task in database: %w", err)
	}

	if results != nil && len(results) > 0 {
		if err := s.storage.SaveTaskResults(task.ID, results); err != nil {
			log.Printf("Failed to save task results: %v", err)
		}
	}

	// 更新 quality checks 状态
	if task.Status == models.TaskStatusNoResource {
		log.Printf("Task is in no-resource status, updating quality checks as failed: %s (event_id=%d)", task.TaskName, task.EventID)
		if err := s.updateQualityChecksForNoResource(task); err != nil {
			// If Event Receiver is not configured, log a more helpful message
			if s.client.BaseURL == "" {
				log.Printf("[ERROR] Event Receiver IP is not configured! Please configure it in the admin console. Quality checks will NOT be updated.")
			} else {
				log.Printf("Failed to update quality checks for no-resource task: %v", err)
			}
		}
	} else {
		log.Printf("Updating quality checks for task: %s (event_id=%d)", task.TaskName, task.EventID)
		if err := s.updateQualityChecks(task, results); err != nil {
			// If Event Receiver is not configured, log a more helpful message
			if s.client.BaseURL == "" {
				log.Printf("[ERROR] Event Receiver IP is not configured! Please configure it in the admin console. Quality checks will NOT be updated.")
			} else {
				log.Printf("Failed to update quality checks: %v", err)
			}
		}
	}

	shouldCreateNext := s.creator.ShouldCreateNextTask(task.EventID, task)
	isLastTask := s.creator.IsLastTask(task.ExecuteOrder)
	log.Printf("CompleteTask: shouldCreateNext=%v, isLastTask=%v, task_status=%s", shouldCreateNext, isLastTask, task.Status)

	// 处理失败或无资源的任务：立即标记事件为失败
	// no_resource 状态也需要更新事件状态，否则 event-receiver 看不到失败原因
	if task.Status == models.TaskStatusFailed || task.Status == models.TaskStatusNoResource {
		now := time.Now()
		if err := s.client.UpdateEventStatus(task.EventID, "failed", now.Format(time.RFC3339)); err != nil {
			log.Printf("Failed to update event status to failed: %v", err)
		} else {
			log.Printf("Event %d marked as failed due to task %s (status=%s)", task.EventID, task.TaskName, task.Status)
		}
		// 释放 Testbed
		if err := s.ReleaseTestbed(task); err != nil {
			log.Printf("Failed to release testbed: %v", err)
		}
		return nil
	}

	if shouldCreateNext {
		event := s.GetEventFromCache(task.EventID)
		nextTask, err := s.creator.CreateNextTask(task.EventID, task.ExecuteOrder, event, task)
		if err != nil {
			log.Printf("Failed to create next task: %v", err)
		} else if nextTask != nil {
			log.Printf("[CompleteTaskWithSuccess] Created next task: taskID=%d, taskName=%s, ChartURL=%s, TestbedIP=%s",
				nextTask.ID, nextTask.TaskName, nextTask.ChartURL, nextTask.TestbedIP)
			if err := s.storage.CreateTask(nextTask); err != nil {
				log.Printf("Failed to save next task to database: %v", err)
			} else {
				log.Printf("[CompleteTaskWithSuccess] Saved to DB: taskID=%d, ChartURL='%s'", nextTask.ID, nextTask.ChartURL)
				log.Printf("Created next task %s for event %d", nextTask.TaskName, task.EventID)
			}
		}
	} else if isLastTask {
		// 最后一个任务完成，更新事件状态为 completed
		// 注意：no_resource 和 failed 状态已在前面处理，不会到达这里
		now := time.Now()
		if err := s.client.UpdateEventStatus(task.EventID, "completed", now.Format(time.RFC3339)); err != nil {
			log.Printf("Failed to update event status to completed: %v", err)
		} else {
			log.Printf("Event %d completed successfully", task.EventID)
		}
		// 整个流程结束，释放 Testbed
		if err := s.ReleaseTestbed(task); err != nil {
			log.Printf("Failed to release testbed: %v", err)
		}
	}

	return nil
}

func (s *SchedulerWithStorage) FailTask(task *models.Task, reason string) error {
	task.MarkFailed(reason)

	if err := s.storage.UpdateTask(task); err != nil {
		return fmt.Errorf("failed to update task in database: %w", err)
	}

	if err := s.updateQualityChecksForFailure(task); err != nil {
		log.Printf("Failed to update quality checks: %v", err)
	}

	// 释放 Testbed（如果是 deployment 任务）
	if err := s.ReleaseTestbed(task); err != nil {
		log.Printf("Failed to release testbed: %v", err)
	}

	now := time.Now()
	if err := s.client.UpdateEventStatus(task.EventID, "failed", now.Format(time.RFC3339)); err != nil {
		log.Printf("Failed to update event status to failed: %v", err)
	}

	log.Printf("Event %d marked as failed: %s", task.EventID, reason)

	return nil
}

func (s *SchedulerWithStorage) CancelTask(task *models.Task, reason string) error {
	task.MarkCancelled(reason)

	if err := s.storage.UpdateTask(task); err != nil {
		return fmt.Errorf("failed to update task in database: %w", err)
	}

	if err := s.updateQualityChecksForCancelled(task); err != nil {
		log.Printf("Failed to update quality checks: %v", err)
	}

	// 释放 Testbed（如果是 deployment 任务）
	if err := s.ReleaseTestbed(task); err != nil {
		log.Printf("Failed to release testbed: %v", err)
	}

	return nil
}

func (s *SchedulerWithStorage) TimeoutTask(task *models.Task, reason string) error {
	task.MarkTimeout(reason)

	if err := s.storage.UpdateTask(task); err != nil {
		return fmt.Errorf("failed to update task in database: %w", err)
	}

	if err := s.updateQualityChecksForFailure(task); err != nil {
		log.Printf("Failed to update quality checks: %v", err)
	}

	now := time.Now()
	if err := s.client.UpdateEventStatus(task.EventID, "failed", now.Format(time.RFC3339)); err != nil {
		log.Printf("Failed to update event status to failed: %v", err)
	}

	return nil
}

// SaveTaskResults saves task results to storage
func (s *SchedulerWithStorage) SaveTaskResults(taskID int, results []models.TaskResult) error {
	return s.storage.SaveTaskResults(taskID, results)
}

func (s *SchedulerWithStorage) SkipTask(task *models.Task, reason string) error {
	task.MarkSkipped(reason)

	if err := s.storage.UpdateTask(task); err != nil {
		return fmt.Errorf("failed to update task in database: %w", err)
	}

	// 更新 Event Receiver 的质量检查状态
	log.Printf("Updating quality checks for skipped task: %s (event_id=%d)", task.TaskName, task.EventID)
	if err := s.updateQualityChecksForSkipped(task, reason); err != nil {
		// If Event Receiver is not configured, log a more helpful message but don't fail
		if s.client.BaseURL == "" {
			log.Printf("[WARNING] Event Receiver IP is not configured! Quality checks will NOT be updated in Event Receiver.")
		} else {
			// If Event Receiver is configured but update failed, return error
			log.Printf("[ERROR] Failed to update quality checks for skipped task: %v", err)
			return fmt.Errorf("failed to update quality checks in Event Receiver: %w", err)
		}
	}

	if s.creator.ShouldCreateNextTask(task.EventID, task) {
		event := s.GetEventFromCache(task.EventID)
		nextTask, err := s.creator.CreateNextTask(task.EventID, task.ExecuteOrder, event, task)
		if err != nil {
			log.Printf("Failed to create next task: %v", err)
		} else if nextTask != nil {
			if err := s.storage.CreateTask(nextTask); err != nil {
				log.Printf("Failed to save next task to database: %v", err)
			} else {
				log.Printf("Created next task %s for event %d after skip", nextTask.TaskName, task.EventID)
			}
		}
	} else if s.creator.IsLastTask(task.ExecuteOrder) {
		if task.IsBlocking() {
			now := time.Now()
			if err := s.client.UpdateEventStatus(task.EventID, "failed", now.Format(time.RFC3339)); err != nil {
				log.Printf("Failed to update event status to failed: %v", err)
			}
		} else {
			now := time.Now()
			if err := s.client.UpdateEventStatus(task.EventID, "completed", now.Format(time.RFC3339)); err != nil {
				log.Printf("Failed to update event status to completed: %v", err)
			}
		}
	}

	// 释放 Testbed（如果是 deployment 任务）
	if err := s.ReleaseTestbed(task); err != nil {
		log.Printf("Failed to release testbed: %v", err)
	}

	return nil
}

func (s *SchedulerWithStorage) GetRunningTasks() ([]*models.Task, error) {
	return s.storage.GetRunningTasks()
}

func (s *SchedulerWithStorage) GetEventFromCache(eventID int) *api.Event {
	s.eventCacheMu.RLock()
	defer s.eventCacheMu.RUnlock()
	return s.eventCache[eventID]
}

func (s *SchedulerWithStorage) updateQualityChecks(task *models.Task, results []models.TaskResult) error {
	if task.TaskName == "basic_ci_all" {
		return s.updateBasicCIChecks(task, results)
	}

	if task.Status == models.TaskStatusNoResource {
		return s.updateQualityChecksForNoResource(task)
	}

	if len(results) == 0 {
		return nil
	}

	result := results[0]
	now := time.Now()

	status := "passed"
	if result.Result == "fail" || result.Result == "timeout" || result.Result == "cancelled" {
		status = "failed"
	} else if result.Result == "skipped" {
		status = "skipped"
	}

	// 总是获取最新的事件信息，而不是依赖缓存
	event, err := s.refreshEventCache(task.EventID)
	if err != nil {
		return fmt.Errorf("event not found in cache and failed to refresh: %w", err)
	}

	log.Printf("DEBUG: Updating quality checks for event %d (task: %s, check_type: %s), event has %d quality checks",
		task.EventID, task.TaskName, result.CheckType, len(event.QualityChecks))

	var checkID int
	var checkStatus string
	for _, check := range event.QualityChecks {
		if check.CheckType == result.CheckType {
			checkID = check.ID
			checkStatus = check.CheckStatus
			log.Printf("DEBUG: Found quality check ID %d for check_type %s with current status: %s", checkID, check.CheckType, checkStatus)
			break
		}
	}

	if checkID == 0 {
		return fmt.Errorf("quality check not found for check_type: %s", result.CheckType)
	}

	// 只更新状态为 pending 或 running 的质量检查，不更新已经被 cancelled 的检查
	if checkStatus == "cancelled" {
		log.Printf("DEBUG: Skipping update for quality check ID %d (already cancelled)", checkID)
		return nil
	}

	updates := []api.QualityCheckUpdate{
		{
			ID:          checkID,
			CheckStatus: status,
			StartedAt:   now.Format(time.RFC3339),
			CompletedAt: now.Format(time.RFC3339),
			Output:      result.Output,
			Extra:       serializeExtra(result.Extra),
		},
	}

	return s.client.BatchUpdateQualityChecks(task.EventID, updates)
}

func (s *SchedulerWithStorage) updateBasicCIChecks(task *models.Task, results []models.TaskResult) error {
	if task.Status == models.TaskStatusNoResource {
		return s.updateQualityChecksForNoResource(task)
	}

	var updates []api.QualityCheckUpdate
	now := time.Now()

	// 总是获取最新的事件信息，而不是依赖缓存
	event, err := s.refreshEventCache(task.EventID)
	if err != nil {
		return fmt.Errorf("event not found in cache and failed to refresh: %w", err)
	}

	if len(event.QualityChecks) == 0 {
		return fmt.Errorf("quality checks not found for event %d", task.EventID)
	}

	basicCIChecks := make(map[string]api.QualityCheck)
	for _, check := range event.QualityChecks {
		if check.Stage == "basic_ci" {
			basicCIChecks[check.CheckType] = check
		}
	}

	for _, result := range results {
		status := "passed"
		if result.Result == "fail" || result.Result == "timeout" || result.Result == "cancelled" {
			status = "failed"
		} else if result.Result == "skipped" {
			status = "skipped"
		}

		check, exists := basicCIChecks[result.CheckType]
		if !exists {
			continue
		}

		// 只更新状态为 pending 或 running 的质量检查，不更新已经被 cancelled 的检查
		if check.CheckStatus == "cancelled" {
			log.Printf("DEBUG: Skipping update for basic CI check ID %d (already cancelled)", check.ID)
			continue
		}

		updates = append(updates, api.QualityCheckUpdate{
			ID:          check.ID,
			CheckStatus: status,
			StartedAt:   now.Format(time.RFC3339),
			CompletedAt: now.Format(time.RFC3339),
			Output:      result.Output,
			Extra:       serializeExtra(result.Extra),
		})
	}

	return s.client.BatchUpdateQualityChecks(task.EventID, updates)
}

func (s *SchedulerWithStorage) updateQualityChecksForFailure(task *models.Task) error {
	event, err := s.refreshEventCache(task.EventID)
	if err != nil {
		return fmt.Errorf("event not found in cache and failed to refresh: %w", err)
	}

	var updates []api.QualityCheckUpdate
	now := time.Now()

	errorMsg := ""
	if task.ErrorMessage != nil {
		errorMsg = *task.ErrorMessage
	}

	for _, check := range event.QualityChecks {
		if check.Stage == task.Stage && (check.CheckStatus == "pending" || check.CheckStatus == "running") {
			updates = append(updates, api.QualityCheckUpdate{
				ID:           check.ID,
				CheckStatus:  "failed",
				CompletedAt:  now.Format(time.RFC3339),
				ErrorMessage: errorMsg,
			})
		}
	}

	if len(updates) > 0 {
		return s.client.BatchUpdateQualityChecks(task.EventID, updates)
	}

	return nil
}

func (s *SchedulerWithStorage) updateQualityChecksForCancelled(task *models.Task) error {
	// 总是获取最新的事件信息，而不是依赖缓存
	event, err := s.refreshEventCache(task.EventID)
	if err != nil {
		return fmt.Errorf("event not found in cache and failed to refresh: %w", err)
	}

	var updates []api.QualityCheckUpdate
	now := time.Now()

	errorMsg := ""
	if task.ErrorMessage != nil {
		errorMsg = *task.ErrorMessage
	}

	for _, check := range event.QualityChecks {
		if check.CheckStatus == "pending" || check.CheckStatus == "running" {
			updates = append(updates, api.QualityCheckUpdate{
				ID:           check.ID,
				CheckStatus:  "cancelled",
				CompletedAt:  now.Format(time.RFC3339),
				ErrorMessage: errorMsg,
			})
		}
	}

	if len(updates) > 0 {
		return s.client.BatchUpdateQualityChecks(task.EventID, updates)
	}

	return nil
}

func (s *SchedulerWithStorage) updateQualityChecksForRunning(task *models.Task) error {
	// 总是获取最新的事件信息，而不是依赖缓存
	event, err := s.refreshEventCache(task.EventID)
	if err != nil {
		return fmt.Errorf("event not found in cache and failed to refresh: %w", err)
	}

	var updates []api.QualityCheckUpdate
	now := time.Now()

	for _, check := range event.QualityChecks {
		if check.Stage == task.Stage && check.CheckStatus == "pending" {
			updates = append(updates, api.QualityCheckUpdate{
				ID:          check.ID,
				CheckStatus: "running",
				StartedAt:   now.Format(time.RFC3339),
			})
		}
	}

	if len(updates) > 0 {
		return s.client.BatchUpdateQualityChecks(task.EventID, updates)
	}

	return nil
}

func (s *SchedulerWithStorage) GetTasksByEventID(eventID int) ([]*models.Task, error) {
	return s.storage.GetTasksByEventID(eventID)
}

func (s *SchedulerWithStorage) GetTaskByID(id int) (*models.Task, error) {
	return s.storage.GetTask(id)
}

func (s *SchedulerWithStorage) GetLatestTaskByEventID(eventID int) (*models.Task, error) {
	return s.storage.GetLatestTaskByEventID(eventID)
}

func (s *SchedulerWithStorage) UpdateTask(task *models.Task) error {
	return s.storage.UpdateTask(task)
}

func (s *SchedulerWithStorage) UpdateTaskAnalyzing(taskID int, analyzing bool) error {
	return s.storage.UpdateTaskAnalyzing(taskID, analyzing)
}

func (s *SchedulerWithStorage) IsTaskAnalyzing(taskID int) (bool, error) {
	return s.storage.IsTaskAnalyzing(taskID)
}

func (s *SchedulerWithStorage) TryStartAnalysisOrResetStale(taskID int) (bool, error) {
	return s.storage.TryStartAnalysisOrResetStale(taskID)
}

func (s *SchedulerWithStorage) GetTaskResults(taskID int) ([]models.TaskResult, error) {
	return s.storage.GetTaskResults(taskID)
}

func (s *SchedulerWithStorage) updateQualityChecksForNoResource(task *models.Task) error {
	event, err := s.refreshEventCache(task.EventID)
	if err != nil {
		return fmt.Errorf("event not found in cache and failed to refresh: %w", err)
	}

	var updates []api.QualityCheckUpdate
	now := time.Now()

	errorMsg := ""
	if task.ErrorMessage != nil {
		errorMsg = *task.ErrorMessage
	}

	for _, check := range event.QualityChecks {
		if check.CheckStatus == "pending" || check.CheckStatus == "running" {
			updates = append(updates, api.QualityCheckUpdate{
				ID:           check.ID,
				CheckStatus:  "failed",
				CompletedAt:  now.Format(time.RFC3339),
				ErrorMessage: errorMsg,
			})
		}
	}

	if len(updates) > 0 {
		return s.client.BatchUpdateQualityChecks(task.EventID, updates)
	}

	return nil
}

func (s *SchedulerWithStorage) updateQualityChecksForSkipped(task *models.Task, reason string) error {
	// 总是获取最新的事件信息，而不是依赖缓存
	event, err := s.refreshEventCache(task.EventID)
	if err != nil {
		return fmt.Errorf("event not found in cache and failed to refresh: %w", err)
	}

	var updates []api.QualityCheckUpdate
	now := time.Now()

	log.Printf("[updateQualityChecksForSkipped] task=%s, event_id=%d, total_checks=%d",
		task.TaskName, task.EventID, len(event.QualityChecks))

	if task.TaskName == "basic_ci_all" {
		// basic_ci_all 特殊处理，更新所有 basic_ci 阶段的检查
		for _, check := range event.QualityChecks {
			log.Printf("[updateQualityChecksForSkipped] check: id=%d, type=%s, stage=%q, status=%q",
				check.ID, check.CheckType, check.Stage, check.CheckStatus)
			if check.Stage == "basic_ci" && (check.CheckStatus == "pending" || check.CheckStatus == "running") {
				updates = append(updates, api.QualityCheckUpdate{
					ID:          check.ID,
					CheckStatus: "skipped",
					CompletedAt: now.Format(time.RFC3339),
					Output:      reason,
				})
				log.Printf("[updateQualityChecksForSkipped] adding update for check id=%d", check.ID)
			}
		}
	} else if task.TaskName == "deployment_deployment" {
		// deployment_deployment 特殊处理，更新所有 deployment 阶段的检查
		for _, check := range event.QualityChecks {
			log.Printf("[updateQualityChecksForSkipped] check: id=%d, type=%s, stage=%q, status=%q",
				check.ID, check.CheckType, check.Stage, check.CheckStatus)
			if check.Stage == "deployment" && (check.CheckStatus == "pending" || check.CheckStatus == "running") {
				updates = append(updates, api.QualityCheckUpdate{
					ID:          check.ID,
					CheckStatus: "skipped",
					CompletedAt: now.Format(time.RFC3339),
					Output:      reason,
				})
				log.Printf("[updateQualityChecksForSkipped] adding update for check id=%d", check.ID)
			}
		}
	} else {
		// 其他任务根据 CheckType 更新对应的检查
		log.Printf("[updateQualityChecksForSkipped] task.CheckType=%q", task.CheckType)
		for _, check := range event.QualityChecks {
			log.Printf("[updateQualityChecksForSkipped] check: id=%d, type=%s, status=%q",
				check.ID, check.CheckType, check.CheckStatus)
			if check.CheckType == task.CheckType && (check.CheckStatus == "pending" || check.CheckStatus == "running") {
				updates = append(updates, api.QualityCheckUpdate{
					ID:          check.ID,
					CheckStatus: "skipped",
					CompletedAt: now.Format(time.RFC3339),
					Output:      reason,
				})
				log.Printf("[updateQualityChecksForSkipped] adding update for check id=%d", check.ID)
				break
			}
		}
	}

	if len(updates) == 0 {
		log.Printf("[WARNING] No quality checks found to update for skipped task %s (event_id=%d)",
			task.TaskName, task.EventID)
		return fmt.Errorf("no quality checks found to update for skipped task %s (event_id=%d)",
			task.TaskName, task.EventID)
	}

	log.Printf("[updateQualityChecksForSkipped] updating %d checks for event %d", len(updates), task.EventID)
	if err := s.client.BatchUpdateQualityChecks(task.EventID, updates); err != nil {
		return fmt.Errorf("failed to batch update quality checks: %w", err)
	}

	return nil
}
