package scheduler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	sharedmodels "github.com/quality-gateway/shared/pkg/models"
	"github.com/quality-gateway/task-scheduler/internal/models"
	"github.com/quality-gateway/task-scheduler/internal/storage"
)

// EventStoreClient 事件存储服务客户端
type EventStoreClient interface {
	GetEvent(eventID int) (*Event, error)
	UpdateEventStatus(eventID int, status string) error
	BatchUpdateQualityChecks(eventID int, updates []QualityCheckUpdate) error
}

// Event 事件信息
type Event struct {
	ID            int               `json:"id"`
	EventID       string            `json:"event_id"`
	EventType     string            `json:"event_type"`
	PRNumber      int               `json:"pr_number"`
	SourceBranch  string            `json:"source_branch"`
	TargetBranch  string            `json:"target_branch"`
	RepoURL       string            `json:"repo_url"`
	EventStatus   string            `json:"event_status"`
	QualityChecks []QualityCheck    `json:"quality_checks"`
}

// QualityCheck 质量检查
type QualityCheck struct {
	ID           int               `json:"id"`
	CheckType    string            `json:"check_type"`
	CheckStatus  string            `json:"check_status"`
	Stage        string            `json:"stage"`
	StageOrder   int               `json:"stage_order"`
	Output       string            `json:"output,omitempty"`
	ErrorMessage string            `json:"error_message,omitempty"`
	Extra        map[string]interface{} `json:"extra,omitempty"`
}

// QualityCheckUpdate 质量检查更新
type QualityCheckUpdate struct {
	ID           int    `json:"id"`
	CheckStatus  string `json:"check_status"`
	StartedAt    string `json:"started_at,omitempty"`
	CompletedAt  string `json:"completed_at,omitempty"`
	Output       string `json:"output,omitempty"`
	Extra        string `json:"extra,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// Scheduler 任务调度器
type Scheduler struct {
	storage       storage.TaskStorage
	eventStore    EventStoreClient
	eventCache    map[int]*Event
	eventCacheMu  sync.RWMutex
}

// NewScheduler 创建调度器
func NewScheduler(store storage.TaskStorage, eventStore EventStoreClient) *Scheduler {
	return &Scheduler{
		storage:    store,
		eventStore: eventStore,
		eventCache: make(map[int]*Event),
	}
}

// CreateTask 创建任务
func (s *Scheduler) CreateTask(eventID int, taskName, checkType, stage string,
	stageOrder, checkOrder, executeOrder int, requestURL string) (*models.Task, error) {

	task := models.NewTask(eventID, taskName, checkType, stage,
		stageOrder, checkOrder, executeOrder, requestURL)

	if err := s.storage.CreateTask(task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	log.Printf("[Scheduler] Created task %s (id=%d) for event %d",
		taskName, task.ID, eventID)

	return task, nil
}

// GetTask 获取任务
func (s *Scheduler) GetTask(id int) (*models.Task, error) {
	return s.storage.GetTask(id)
}

// GetTasksByEventID 获取事件的所有任务
func (s *Scheduler) GetTasksByEventID(eventID int) ([]*models.Task, error) {
	return s.storage.GetTasksByEventID(eventID)
}

// GetPendingTasks 获取待执行任务
func (s *Scheduler) GetPendingTasks() ([]*models.Task, error) {
	tasks, err := s.storage.GetPendingTasks()
	if err != nil {
		return nil, err
	}

	// 按执行顺序排序
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ExecuteOrder < tasks[j].ExecuteOrder
	})

	return tasks, nil
}

// GetRunningTasks 获取运行中的任务
func (s *Scheduler) GetRunningTasks() ([]*models.Task, error) {
	return s.storage.GetRunningTasks()
}

// ListTasks 分页获取任务列表
func (s *Scheduler) ListTasks(limit, offset int) ([]*models.Task, int, error) {
	tasks, err := s.storage.ListTasks(limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// 获取总数
	var total int
	for _, status := range []sharedmodels.TaskStatus{
		sharedmodels.TaskStatusPending,
		sharedmodels.TaskStatusRunning,
		sharedmodels.TaskStatusPassed,
		sharedmodels.TaskStatusFailed,
		sharedmodels.TaskStatusSkipped,
		sharedmodels.TaskStatusCancelled,
		sharedmodels.TaskStatusTimeout,
		sharedmodels.TaskStatusNoResource,
	} {
		count, _ := s.storage.GetTaskCountByStatus(status)
		total += count
	}

	return tasks, total, nil
}

// StartTask 启动任务
func (s *Scheduler) StartTask(taskID int) (*models.Task, error) {
	task, err := s.storage.GetTask(taskID)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	// 使用 CAS 操作确保只有一个 goroutine 能启动任务
	started, err := s.storage.TryMarkTaskRunning(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to mark task running: %w", err)
	}
	if !started {
		return nil, fmt.Errorf("task %d is already being processed", taskID)
	}

	// 重新获取更新后的任务
	task, err = s.storage.GetTask(taskID)
	if err != nil {
		return nil, err
	}

	// 更新 Event Store 的质量检查状态
	if err := s.updateQualityChecksForRunning(task); err != nil {
		log.Printf("[Scheduler] Failed to update quality checks for running: %v", err)
	}

	// 如果是 basic_ci_all 任务，更新事件状态
	if task.TaskName == "basic_ci_all" {
		if err := s.eventStore.UpdateEventStatus(task.EventID, "processing"); err != nil {
			log.Printf("[Scheduler] Failed to update event status: %v", err)
		}
	}

	log.Printf("[Scheduler] Started task %s (id=%d) for event %d",
		task.TaskName, task.ID, task.EventID)

	return task, nil
}

// CompleteTask 完成任务
func (s *Scheduler) CompleteTask(taskID int, results []models.TaskResult) error {
	task, err := s.storage.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// 检查是否有失败的子检查项
	hasFailures := false
	for _, result := range results {
		if result.Result == "fail" {
			hasFailures = true
			break
		}
	}

	if hasFailures {
		task.MarkFailed("one or more quality checks failed")
	} else {
		task.MarkPassed()
	}

	// 更新任务状态
	if err := s.storage.UpdateTask(task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// 保存任务结果
	if len(results) > 0 {
		if err := s.storage.SaveTaskResults(taskID, results); err != nil {
			log.Printf("[Scheduler] Failed to save task results: %v", err)
		}
	}

	// 更新质量检查
	if err := s.updateQualityChecks(task, results); err != nil {
		log.Printf("[Scheduler] Failed to update quality checks: %v", err)
	}

	// 创建下一个任务或完成事件
	if err := s.handleTaskCompletion(task); err != nil {
		log.Printf("[Scheduler] Failed to handle task completion: %v", err)
	}

	log.Printf("[Scheduler] Completed task %s (id=%d) for event %d",
		task.TaskName, task.ID, task.EventID)

	return nil
}

// FailTask 标记任务失败
func (s *Scheduler) FailTask(taskID int, reason string) error {
	task, err := s.storage.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	task.MarkFailed(reason)

	if err := s.storage.UpdateTask(task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// 更新质量检查
	if err := s.updateQualityChecksForFailure(task); err != nil {
		log.Printf("[Scheduler] Failed to update quality checks: %v", err)
	}

	// 释放 testbed（如果有）
	if task.TestbedUUID != "" {
		if err := s.storage.ReleaseTestbed(taskID); err != nil {
			log.Printf("[Scheduler] Failed to release testbed: %v", err)
		}
	}

	// 更新事件状态为失败
	if err := s.eventStore.UpdateEventStatus(task.EventID, "failed"); err != nil {
		log.Printf("[Scheduler] Failed to update event status: %v", err)
	}

	log.Printf("[Scheduler] Failed task %s (id=%d) for event %d: %s",
		task.TaskName, task.ID, task.EventID, reason)

	return nil
}

// CancelTask 取消任务
func (s *Scheduler) CancelTask(taskID int, reason string) error {
	task, err := s.storage.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	task.MarkCancelled(reason)

	if err := s.storage.UpdateTask(task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// 更新质量检查
	if err := s.updateQualityChecksForCancelled(task); err != nil {
		log.Printf("[Scheduler] Failed to update quality checks: %v", err)
	}

	// 释放 testbed（如果有）
	if task.TestbedUUID != "" {
		if err := s.storage.ReleaseTestbed(taskID); err != nil {
			log.Printf("[Scheduler] Failed to release testbed: %v", err)
		}
	}

	log.Printf("[Scheduler] Cancelled task %s (id=%d) for event %d",
		task.TaskName, task.ID, task.EventID)

	return nil
}

// CancelEventTasks 取消事件的所有任务
func (s *Scheduler) CancelEventTasks(eventID int, reason string) (int, error) {
	log.Printf("[Scheduler] CancelEventTasks called for event %d, reason: %s", eventID, reason)

	count, err := s.storage.CancelTasksByEventID(eventID, reason)
	if err != nil {
		log.Printf("[Scheduler] CancelTasksByEventID failed for event %d: %v", eventID, err)
		return 0, fmt.Errorf("failed to cancel tasks: %w", err)
	}

	log.Printf("[Scheduler] CancelTasksByEventID cancelled %d tasks for event %d", count, eventID)

	if err := s.eventStore.UpdateEventStatus(eventID, "cancelled"); err != nil {
		log.Printf("[Scheduler] UpdateEventStatus failed for event %d: %v", eventID, err)
		return count, fmt.Errorf("failed to update event status: %w", err)
	}

	log.Printf("[Scheduler] Event %d status updated to cancelled", eventID)

	return count, nil
}

// handleTaskCompletion 处理任务完成后的逻辑
func (s *Scheduler) handleTaskCompletion(task *models.Task) error {
	// 如果任务失败或无资源，更新事件状态为失败
	if task.Status == sharedmodels.TaskStatusFailed || task.Status == sharedmodels.TaskStatusNoResource {
		return s.eventStore.UpdateEventStatus(task.EventID, "failed")
	}

	// 获取事件信息（保留用于未来扩展）
	_, err := s.getEvent(task.EventID)
	if err != nil {
		return err
	}

	// 检查是否是最后一个任务
	if task.ExecuteOrder >= 6 { // MaxExecuteOrder
		// 最后一个任务完成
		return s.eventStore.UpdateEventStatus(task.EventID, "completed")
	}

	// 创建下一个任务（这里简化处理，实际应该根据任务定义）
	// 这个逻辑应该在 Task Creator 中处理

	return nil
}

// updateQualityChecks 更新质量检查状态
func (s *Scheduler) updateQualityChecks(task *models.Task, results []models.TaskResult) error {
	event, err := s.getEvent(task.EventID)
	if err != nil {
		return err
	}

	now := time.Now()
	var updates []QualityCheckUpdate

	if task.TaskName == "basic_ci_all" {
		// 处理 basic_ci_all 的多个检查
		for _, result := range results {
			for _, check := range event.QualityChecks {
				if check.Stage == "basic_ci" && check.CheckType == result.CheckType {
					status := "passed"
					if result.Result == "fail" {
						status = "failed"
					}

					updates = append(updates, QualityCheckUpdate{
						ID:          check.ID,
						CheckStatus: status,
						CompletedAt: now.Format(time.RFC3339),
						Output:      result.Output,
						Extra:       result.Extra, // Already a JSON string
					})
				}
			}
		}
	} else {
		// 处理单个检查的任务
		for _, result := range results {
			for _, check := range event.QualityChecks {
				if check.CheckType == result.CheckType {
					status := "passed"
					if result.Result == "fail" {
						status = "failed"
					}

					updates = append(updates, QualityCheckUpdate{
						ID:          check.ID,
						CheckStatus: status,
						CompletedAt: now.Format(time.RFC3339),
						Output:      result.Output,
						Extra:       result.Extra, // Already a JSON string
					})
				}
			}
		}
	}

	if len(updates) > 0 {
		return s.eventStore.BatchUpdateQualityChecks(task.EventID, updates)
	}

	return nil
}

// updateQualityChecksForRunning 更新运行中的质量检查
func (s *Scheduler) updateQualityChecksForRunning(task *models.Task) error {
	event, err := s.getEvent(task.EventID)
	if err != nil {
		return err
	}

	now := time.Now()
	var updates []QualityCheckUpdate

	for _, check := range event.QualityChecks {
		if check.Stage == task.Stage && check.CheckStatus == "pending" {
			updates = append(updates, QualityCheckUpdate{
				ID:          check.ID,
				CheckStatus: "running",
				StartedAt:   now.Format(time.RFC3339),
			})
		}
	}

	if len(updates) > 0 {
		return s.eventStore.BatchUpdateQualityChecks(task.EventID, updates)
	}

	return nil
}

// updateQualityChecksForFailure 更新失败的质量检查
func (s *Scheduler) updateQualityChecksForFailure(task *models.Task) error {
	event, err := s.getEvent(task.EventID)
	if err != nil {
		return err
	}

	now := time.Now()
	var updates []QualityCheckUpdate

	errorMsg := ""
	if task.ErrorMessage != nil {
		errorMsg = *task.ErrorMessage
	}

	for _, check := range event.QualityChecks {
		if check.Stage == task.Stage &&
			(check.CheckStatus == "pending" || check.CheckStatus == "running") {
			updates = append(updates, QualityCheckUpdate{
				ID:           check.ID,
				CheckStatus:  "failed",
				CompletedAt:  now.Format(time.RFC3339),
				ErrorMessage: errorMsg,
			})
		}
	}

	if len(updates) > 0 {
		return s.eventStore.BatchUpdateQualityChecks(task.EventID, updates)
	}

	return nil
}

// updateQualityChecksForCancelled 更新取消的质量检查
func (s *Scheduler) updateQualityChecksForCancelled(task *models.Task) error {
	event, err := s.getEvent(task.EventID)
	if err != nil {
		return err
	}

	now := time.Now()
	var updates []QualityCheckUpdate

	errorMsg := ""
	if task.ErrorMessage != nil {
		errorMsg = *task.ErrorMessage
	}

	for _, check := range event.QualityChecks {
		if check.CheckStatus == "pending" || check.CheckStatus == "running" {
			updates = append(updates, QualityCheckUpdate{
				ID:           check.ID,
				CheckStatus:  "cancelled",
				CompletedAt:  now.Format(time.RFC3339),
				ErrorMessage: errorMsg,
			})
		}
	}

	if len(updates) > 0 {
		return s.eventStore.BatchUpdateQualityChecks(task.EventID, updates)
	}

	return nil
}

// getEvent 获取事件信息（带缓存）
func (s *Scheduler) getEvent(eventID int) (*Event, error) {
	// 先从缓存读取
	s.eventCacheMu.RLock()
	event, exists := s.eventCache[eventID]
	s.eventCacheMu.RUnlock()

	if exists {
		return event, nil
	}

	// 从 Event Store 获取
	event, err := s.eventStore.GetEvent(eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	// 更新缓存
	s.eventCacheMu.Lock()
	s.eventCache[eventID] = event
	s.eventCacheMu.Unlock()

	return event, nil
}

// refreshEventCache 刷新事件缓存
func (s *Scheduler) refreshEventCache(eventID int) (*Event, error) {
	event, err := s.eventStore.GetEvent(eventID)
	if err != nil {
		return nil, err
	}

	s.eventCacheMu.Lock()
	s.eventCache[eventID] = event
	s.eventCacheMu.Unlock()

	return event, nil
}

// serializeExtra 序列化 Extra 字段
func serializeExtra(extra map[string]interface{}) string {
	if extra == nil || len(extra) == 0 {
		return ""
	}
	data, err := json.Marshal(extra)
	if err != nil {
		return ""
	}
	return string(data)
}

// HTTPEventStoreClient HTTP 实现的 Event Store 客户端
type HTTPEventStoreClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHTTPEventStoreClient 创建 HTTP Event Store 客户端
func NewHTTPEventStoreClient(baseURL string) *HTTPEventStoreClient {
	return &HTTPEventStoreClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetEvent 获取事件
func (c *HTTPEventStoreClient) GetEvent(eventID int) (*Event, error) {
	url := fmt.Sprintf("%s/api/events/%d", c.baseURL, eventID)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var response struct {
		Success bool   `json:"success"`
		Data    Event `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response.Data, nil
}

// UpdateEventStatus 更新事件状态
func (c *HTTPEventStoreClient) UpdateEventStatus(eventID int, status string) error {
	url := fmt.Sprintf("%s/api/events/%d/status", c.baseURL, eventID)

	reqBody := map[string]string{"status": status}
	jsonData, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	r, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", r.StatusCode)
	}

	return nil
}

// BatchUpdateQualityChecks 批量更新质量检查
func (c *HTTPEventStoreClient) BatchUpdateQualityChecks(eventID int, updates []QualityCheckUpdate) error {
	url := fmt.Sprintf("%s/api/events/%d/quality-checks/batch", c.baseURL, eventID)

	reqBody := map[string]interface{}{"updates": updates}
	jsonData, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	r, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", r.StatusCode)
	}

	return nil
}
