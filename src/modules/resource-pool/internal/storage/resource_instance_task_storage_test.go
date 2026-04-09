package storage

import (
	"database/sql"
	"testing"
	"time"

	"github.com/hugoh/go-designs/resource-pool/internal/models"
)

// TestResourceInstanceTaskStorage_CreateAndRetrieve 测试创建和获取任务
func TestResourceInstanceTaskStorage_CreateAndRetrieve(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	storage := NewMySQLResourceInstanceTaskStorage(db)

	task := models.NewDeployTask(
		"test-resource-instance-uuid",
		"test-category-uuid",
		models.TriggerSourceManual,
		"test-user",
	)

	err := storage.CreateTask(task)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	if task.ID == 0 {
		t.Error("Expected task ID to be set after creation")
	}

	retrieved, err := storage.GetTaskByUUID(task.UUID)
	if err != nil {
		t.Fatalf("Failed to retrieve task: %v", err)
	}

	if retrieved.UUID != task.UUID {
		t.Errorf("Expected UUID %s, got %s", task.UUID, retrieved.UUID)
	}

	if retrieved.TaskType != models.TaskTypeDeploy {
		t.Errorf("Expected task type %s, got %s", models.TaskTypeDeploy, retrieved.TaskType)
	}

	if retrieved.TriggerSource != models.TriggerSourceManual {
		t.Errorf("Expected trigger source %s, got %s", models.TriggerSourceManual, retrieved.TriggerSource)
	}

	if retrieved.TriggerUser == nil || *retrieved.TriggerUser != "test-user" {
		t.Error("Expected trigger user to be test-user")
	}
}

// TestResourceInstanceTaskStorage_UpdateStatus 测试更新任务状态
func TestResourceInstanceTaskStorage_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	storage := NewMySQLResourceInstanceTaskStorage(db)

	task := models.NewDeployTask(
		"test-resource-instance-uuid",
		"test-category-uuid",
		models.TriggerSourceManual,
		"test-user",
	)

	err := storage.CreateTask(task)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// 测试状态转换
	task.MarkRunning()
	err = storage.UpdateTask(task)
	if err != nil {
		t.Fatalf("Failed to update task to running: %v", err)
	}

	task.MarkCompleted(true)
	err = storage.UpdateTask(task)
	if err != nil {
		t.Fatalf("Failed to update task to completed: %v", err)
	}

	retrieved, err := storage.GetTaskByUUID(task.UUID)
	if err != nil {
		t.Fatalf("Failed to retrieve task: %v", err)
	}

	if retrieved.Status != models.TaskStatusCompleted {
		t.Errorf("Expected status %s, got %s", models.TaskStatusCompleted, retrieved.Status)
	}

	if retrieved.Success == nil || !*retrieved.Success {
		t.Error("Expected task to be marked as successful")
	}

	if retrieved.StartedAt == nil || retrieved.CompletedAt == nil {
		t.Error("Expected StartedAt and CompletedAt to be set")
	}

	if retrieved.DurationMs == nil {
		t.Error("Expected DurationMs to be set")
	}
}

// TestResourceInstanceTaskStorage_ListByResourceInstance 测试按资源实例列出任务
func TestResourceInstanceTaskStorage_ListByResourceInstance(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	storage := NewMySQLResourceInstanceTaskStorage(db)

	resourceUUID := "test-resource-instance-uuid"

	// 创建多个任务
	for i := 0; i < 3; i++ {
		task := models.NewDeployTask(
			resourceUUID,
			"test-category-uuid",
			models.TriggerSourceManual,
			"test-user",
		)
		err := storage.CreateTask(task)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}
	}

	tasks, err := storage.ListTasksByResourceInstance(resourceUUID)
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}

	if len(tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(tasks))
	}

	// 验证顺序（应该按创建时间降序）
	for i, task := range tasks {
		if i > 0 && tasks[i-1].CreatedAt.Before(task.CreatedAt) {
			t.Error("Tasks should be ordered by created_at DESC")
		}
	}
}

// TestResourceInstanceTaskStorage_ListByStatus 测试按状态列出任务
func TestResourceInstanceTaskStorage_ListByStatus(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	storage := NewMySQLResourceInstanceTaskStorage(db)

	// 创建不同状态的任务
	pendingTask := models.NewDeployTask("uuid1", "cat1", models.TriggerSourceManual, "user")
	storage.CreateTask(pendingTask)

	runningTask := models.NewDeployTask("uuid2", "cat2", models.TriggerSourceManual, "user")
	runningTask.MarkRunning()
	storage.CreateTask(runningTask)

	completedTask := models.NewDeployTask("uuid3", "cat3", models.TriggerSourceManual, "user")
	completedTask.MarkCompleted(true)
	storage.CreateTask(completedTask)

	// 测试按状态查询
	pendingTasks, err := storage.ListTasksByStatus(models.TaskStatusPending)
	if err != nil {
		t.Fatalf("Failed to list pending tasks: %v", err)
	}

	if len(pendingTasks) != 1 {
		t.Errorf("Expected 1 pending task, got %d", len(pendingTasks))
	}

	completedTasks, err := storage.ListTasksByStatus(models.TaskStatusCompleted)
	if err != nil {
		t.Fatalf("Failed to list completed tasks: %v", err)
	}

	if len(completedTasks) != 1 {
		t.Errorf("Expected 1 completed task, got %d", len(completedTasks))
	}
}

// TestResourceInstanceTaskStorage_CountByStatus 测试按状态统计任务
func TestResourceInstanceTaskStorage_CountByStatus(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	storage := NewMySQLResourceInstanceTaskStorage(db)

	// 创建多个任务
	for i := 0; i < 3; i++ {
		task := models.NewDeployTask("uuid", "cat", models.TriggerSourceManual, "user")
		if i == 0 {
			task.MarkRunning()
		} else if i == 1 {
			task.MarkCompleted(true)
		}
		err := storage.CreateTask(task)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}
	}

	pendingCount, err := storage.CountTasksByStatus(models.TaskStatusPending)
	if err != nil {
		t.Fatalf("Failed to count pending tasks: %v", err)
	}

	if pendingCount != 1 {
		t.Errorf("Expected 1 pending task, got %d", pendingCount)
	}

	completedCount, err := storage.CountTasksByStatus(models.TaskStatusCompleted)
	if err != nil {
		t.Fatalf("Failed to count completed tasks: %v", err)
	}

	if completedCount != 1 {
		t.Errorf("Expected 1 completed task, got %d", completedCount)
	}
}

// TestResourceInstanceTaskStorage_TryStartTask 测试 CAS 操作启动任务
func TestResourceInstanceTaskStorage_TryStartTask(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	storage := NewMySQLResourceInstanceTaskStorage(db)

	task := models.NewDeployTask("uuid", "cat", models.TriggerSourceManual, "user")
	err := storage.CreateTask(task)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// 第一次尝试启动应该成功
	started, err := storage.TryStartTask(task.UUID)
	if err != nil {
		t.Fatalf("Failed to start task: %v", err)
	}

	if !started {
		t.Error("Expected first TryStartTask to succeed")
	}

	// 第二次尝试启动应该失败（已经被启动）
	started, err = storage.TryStartTask(task.UUID)
	if err != nil {
		t.Fatalf("Failed on second TryStartTask: %v", err)
	}

	if started {
		t.Error("Expected second TryStartTask to fail (already running)")
	}
}

// TestResourceInstanceTaskStorage_DeleteOldTasks 测试删除旧任务
func TestResourceInstanceTaskStorage_DeleteOldTasks(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	storage := NewMySQLResourceInstanceTaskStorage(db)

	// 创建一个已完成的老任务
	oldTask := models.NewDeployTask("uuid", "cat", models.TriggerSourceManual, "user")
	oldTask.CreatedAt = time.Now().Add(-30 * 24 * time.Hour) // 30天前
	oldTask.MarkCompleted(true)
	err := storage.CreateTask(oldTask)
	if err != nil {
		t.Fatalf("Failed to create old task: %v", err)
	}

	// 创建一个新任务
	newTask := models.NewDeployTask("uuid2", "cat", models.TriggerSourceManual, "user")
	err = storage.CreateTask(newTask)
	if err != nil {
		t.Fatalf("Failed to create new task: %v", err)
	}

	// 删除7天前的任务
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	count, err := storage.DeleteOldTasks(cutoff)
	if err != nil {
		t.Fatalf("Failed to delete old tasks: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected to delete 1 old task, deleted %d", count)
	}

	// 验证新任务还在
	_, err = storage.GetTaskByUUID(newTask.UUID)
	if err != nil {
		t.Error("New task should still exist after cleanup")
	}

	// 验证老任务已被删除
	_, err = storage.GetTaskByUUID(oldTask.UUID)
	if err == nil {
		t.Error("Old task should be deleted after cleanup")
	}
}

// TestResourceInstanceTaskStorage_GetTaskStatistics 测试获取任务统计
func TestResourceInstanceTaskStorage_GetTaskStatistics(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	storage := NewMySQLResourceInstanceTaskStorage(db)

	// 创建不同状态和类型的任务
	task1 := models.NewDeployTask("uuid1", "cat", models.TriggerSourceManual, "user")
	storage.CreateTask(task1)

	task2 := models.NewDeployTask("uuid2", "cat", models.TriggerSourceAutoReplenish, "system")
	task2.MarkCompleted(true)
	storage.CreateTask(task2)

	task3 := models.NewRollbackTask("uuid3", "testbed", "allocation", models.TriggerSourceAllocationRelease)
	task3.MarkFailed("ERROR", "test error")
	storage.CreateTask(task3)

	stats, err := storage.GetTaskStatistics()
	if err != nil {
		t.Fatalf("Failed to get statistics: %v", err)
	}

	if stats.Total != 3 {
		t.Errorf("Expected total 3, got %d", stats.Total)
	}

	if stats.Pending != 1 {
		t.Errorf("Expected 1 pending, got %d", stats.Pending)
	}

	if stats.Completed != 1 {
		t.Errorf("Expected 1 completed, got %d", stats.Completed)
	}

	if stats.Failed != 1 {
		t.Errorf("Expected 1 failed, got %d", stats.Failed)
	}

	if stats.ByType[models.TaskTypeDeploy] != 2 {
		t.Errorf("Expected 2 deploy tasks, got %d", stats.ByType[models.TaskTypeDeploy])
	}

	if stats.ByType[models.TaskTypeRollback] != 1 {
		t.Errorf("Expected 1 rollback task, got %d", stats.ByType[models.TaskTypeRollback])
	}

	if stats.ByTrigger[models.TriggerSourceManual] != 1 {
		t.Errorf("Expected 1 manual trigger, got %d", stats.ByTrigger[models.TriggerSourceManual])
	}
}

// TestResourceInstanceTask_ModelMethods 测试模型方法
func TestResourceInstanceTask_ModelMethods(t *testing.T) {
	task := models.NewDeployTask("uuid", "cat", models.TriggerSourceManual, "user")

	// 测试 DisplayName
	if task.TaskType.DisplayName() != "部署" {
		t.Errorf("Expected display name '部署', got '%s'", task.TaskType.DisplayName())
	}

	if task.Status.DisplayName() != "等待中" {
		t.Errorf("Expected status display name '等待中', got '%s'", task.Status.DisplayName())
	}

	if task.TriggerSource.DisplayName() != "手动触发" {
		t.Errorf("Expected trigger source display name '手动触发', got '%s'", task.TriggerSource.DisplayName())
	}

	// 测试状态转换
	task.MarkRunning()
	if task.Status != models.TaskStatusRunning {
		t.Error("Expected status to be running after MarkRunning")
	}
	if task.StartedAt == nil {
		t.Error("Expected StartedAt to be set after MarkRunning")
	}

	task.MarkCompleted(true)
	if task.Status != models.TaskStatusCompleted {
		t.Error("Expected status to be completed after MarkCompleted")
	}
	if task.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set after MarkCompleted")
	}
	if task.Success == nil || !*task.Success {
		t.Error("Expected Success to be true after MarkCompleted(true)")
	}

	// 测试失败任务
	failTask := models.NewDeployTask("uuid2", "cat", models.TriggerSourceManual, "user")
	failTask.MarkFailed("TEST_ERROR", "test error message")
	if failTask.Status != models.TaskStatusFailed {
		t.Error("Expected status to be failed after MarkFailed")
	}
	if failTask.ErrorCode == nil || *failTask.ErrorCode != "TEST_ERROR" {
		t.Error("Expected ErrorCode to be set after MarkFailed")
	}

	// 测试取消任务
	cancelTask := models.NewDeployTask("uuid3", "cat", models.TriggerSourceManual, "user")
	cancelTask.MarkCancelled("test cancel reason")
	if cancelTask.Status != models.TaskStatusCancelled {
		t.Error("Expected status to be cancelled after MarkCancelled")
	}

	// 测试重试逻辑
	retryTask := models.NewDeployTask("uuid4", "cat", models.TriggerSourceManual, "user")
	retryTask.MarkFailed("ERROR", "error")
	if !retryTask.CanRetry() {
		t.Error("Expected failed task to be retryable")
	}

	retryTask.IncrementRetry()
	if retryTask.RetryCount != 1 {
		t.Errorf("Expected retry count to be 1 after IncrementRetry, got %d", retryTask.RetryCount)
	}
	if retryTask.Status != models.TaskStatusPending {
		t.Error("Expected status to be pending after IncrementRetry")
	}

	// 测试终态检查
	if !models.TaskStatusCompleted.IsTerminal() {
		t.Error("Expected Completed to be a terminal state")
	}
	if !models.TaskStatusFailed.IsTerminal() {
		t.Error("Expected Failed to be a terminal state")
	}
	if models.TaskStatusRunning.IsTerminal() {
		t.Error("Expected Running to not be a terminal state")
	}
}

// TestResourceInstanceTask_NewTasks 测试各种创建任务的方法
func TestResourceInstanceTask_NewTasks(t *testing.T) {
	// 测试 NewDeployTask
	deployTask := models.NewDeployTask("resource-uuid", "category-uuid", models.TriggerSourceManual, "admin")
	if deployTask.TaskType != models.TaskTypeDeploy {
		t.Error("NewDeployTask should create a deploy task")
	}
	if deployTask.TriggerUser == nil || *deployTask.TriggerUser != "admin" {
		t.Error("NewDeployTask should set trigger user")
	}

	// 测试 NewRollbackTask
	rollbackTask := models.NewRollbackTask("resource-uuid", "testbed-uuid", "allocation-uuid", models.TriggerSourceAllocationRelease)
	if rollbackTask.TaskType != models.TaskTypeRollback {
		t.Error("NewRollbackTask should create a rollback task")
	}
	if rollbackTask.TestbedUUID == nil || *rollbackTask.TestbedUUID != "testbed-uuid" {
		t.Error("NewRollbackTask should set testbed UUID")
	}
	if rollbackTask.AllocationUUID == nil || *rollbackTask.AllocationUUID != "allocation-uuid" {
		t.Error("NewRollbackTask should set allocation UUID")
	}

	// 测试 NewHealthCheckTask
	healthCheckTask := models.NewHealthCheckTask("resource-uuid")
	if healthCheckTask.TaskType != models.TaskTypeHealthCheck {
		t.Error("NewHealthCheckTask should create a health check task")
	}

	// 测试 NewAutoReplenishDeployTask
	autoReplenishTask := models.NewAutoReplenishDeployTask("resource-uuid", "policy-uuid", "category-uuid")
	if autoReplenishTask.TaskType != models.TaskTypeDeploy {
		t.Error("NewAutoReplenishDeployTask should create a deploy task")
	}
	if autoReplenishTask.TriggerSource != models.TriggerSourceAutoReplenish {
		t.Error("NewAutoReplenishDeployTask should have auto_replenish trigger source")
	}
	if autoReplenishTask.QuotaPolicyUUID == nil || *autoReplenishTask.QuotaPolicyUUID != "policy-uuid" {
		t.Error("NewAutoReplenishDeployTask should set quota policy UUID")
	}
}

// setupTestDB 设置测试数据库
func setupTestDB(t *testing.T) *sql.DB {
	// 这里需要创建实际的测试数据库连接
	// 为了简化，这里返回 nil，实际测试时需要配置测试数据库
	t.Skip("Skipping test: test database not configured")
	return nil
}

// teardownTestDB 清理测试数据库
func teardownTestDB(t *testing.T, db *sql.DB) {
	// 清理测试数据
}
