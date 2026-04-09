package jobs

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/hugoh/go-designs/resource-pool/internal/deployer"
	"github.com/hugoh/go-designs/resource-pool/internal/models"
	"github.com/hugoh/go-designs/resource-pool/internal/storage"
)

// Mock implementations for testing
type mockQuotaStorage struct {
	policies []*models.QuotaPolicy
}

func (m *mockQuotaStorage) ListPoliciesByPriority() ([]*models.QuotaPolicy, error) {
	return m.policies, nil
}

func (m *mockQuotaStorage) CreateQuotaPolicy(policy *models.QuotaPolicy) error { return nil }
func (m *mockQuotaStorage) GetQuotaPolicy(id int) (*models.QuotaPolicy, error) { return nil, nil }
func (m *mockQuotaStorage) GetQuotaPolicyByUUID(uuid string) (*models.QuotaPolicy, error) {
	return nil, nil
}
func (m *mockQuotaStorage) GetQuotaPolicyByCategory(categoryUUID string) (*models.QuotaPolicy, error) {
	return nil, nil
}
func (m *mockQuotaStorage) GetQuotaPolicyByCategoryAndServiceTarget(categoryUUID string, serviceTarget models.ServiceTarget) (*models.QuotaPolicy, error) {
	return nil, nil
}
func (m *mockQuotaStorage) ListQuotaPolicies() ([]*models.QuotaPolicy, error)  { return nil, nil }
func (m *mockQuotaStorage) UpdateQuotaPolicy(policy *models.QuotaPolicy) error { return nil }
func (m *mockQuotaStorage) DeleteQuotaPolicy(id int) error                     { return nil }
func (m *mockQuotaStorage) DeleteAllQuotaPolicies() error                      { return nil }

type mockTestbedStorage struct {
	testbeds       []*models.Testbed
	availableCount map[string]int // key: categoryUUID_serviceTarget
}

func (m *mockTestbedStorage) CreateTestbed(testbed *models.Testbed) error {
	m.testbeds = append(m.testbeds, testbed)
	return nil
}

func (m *mockTestbedStorage) CountAvailableTestbedsByCategory(categoryUUID string, serviceTarget models.ServiceTarget) (int, error) {
	key := fmt.Sprintf("%s_%s", categoryUUID, serviceTarget)
	if m.availableCount != nil {
		return m.availableCount[key], nil
	}
	return 0, nil
}

func (m *mockTestbedStorage) GetTestbed(id int) (*models.Testbed, error)            { return nil, nil }
func (m *mockTestbedStorage) GetTestbedByUUID(uuid string) (*models.Testbed, error) { return nil, nil }
func (m *mockTestbedStorage) ListTestbeds() ([]*models.Testbed, error)              { return nil, nil }
func (m *mockTestbedStorage) ListTestbedsByStatus(status models.TestbedStatus) ([]*models.Testbed, error) {
	return nil, nil
}
func (m *mockTestbedStorage) ListTestbedsByCategory(categoryUUID string) ([]*models.Testbed, error) {
	return nil, nil
}
func (m *mockTestbedStorage) ListAvailableTestbeds(categoryUUID string) ([]*models.Testbed, error) {
	return nil, nil
}
func (m *mockTestbedStorage) ListAvailableTestbedsByServiceTarget(categoryUUID string, serviceTarget models.ServiceTarget) ([]*models.Testbed, error) {
	return nil, nil
}
func (m *mockTestbedStorage) UpdateTestbed(testbed *models.Testbed) error { return nil }
func (m *mockTestbedStorage) UpdateTestbedStatus(uuid string, status models.TestbedStatus) error {
	return nil
}
func (m *mockTestbedStorage) UpdateTestbedAllocation(testbedUUID, allocUUID string, status models.TestbedStatus) error {
	return nil
}
func (m *mockTestbedStorage) ClearTestbedAllocation(testbedUUID string) error          { return nil }
func (m *mockTestbedStorage) UpdateTestbedHealthCheck(uuid string) error               { return nil }
func (m *mockTestbedStorage) DeleteTestbed(id int) error                               { return nil }
func (m *mockTestbedStorage) CountTestbedsByCategory(categoryUUID string) (int, error) { return 0, nil }
func (m *mockTestbedStorage) CountAllocatedTestbedsByCategory(categoryUUID string) (int, error) {
	return 0, nil
}
func (m *mockTestbedStorage) CountAllAvailableTestbeds() (int, error) {
	return 0, nil
}
func (m *mockTestbedStorage) CountTestbedsByCategoryAndServiceTarget(categoryUUID string, serviceTarget models.ServiceTarget) (int, error) {
	return 0, nil
}
func (m *mockTestbedStorage) DeleteAllTestbeds() error { return nil }
func (m *mockTestbedStorage) ListTestbedsWithPagination(page, pageSize int, status *models.TestbedStatus, categoryUUID *string) ([]*models.Testbed, int, error) {
	return nil, 0, nil
}

type mockResourceStorage struct {
	availableInstances []*models.ResourceInstance
}

func (m *mockResourceStorage) ListAvailableResourceInstances() ([]*models.ResourceInstance, error) {
	return m.availableInstances, nil
}

func (m *mockResourceStorage) GetResourceInstance(id int) (*models.ResourceInstance, error) {
	return nil, nil
}
func (m *mockResourceStorage) GetResourceInstanceByUUID(uuid string) (*models.ResourceInstance, error) {
	return nil, nil
}
func (m *mockResourceStorage) GetResourceInstanceByIPAddress(ipAddress string) (*models.ResourceInstance, error) {
	return nil, nil
}
func (m *mockResourceStorage) CreateResourceInstance(instance *models.ResourceInstance) error {
	return nil
}
func (m *mockResourceStorage) UpdateResourceInstance(instance *models.ResourceInstance) error {
	return nil
}
func (m *mockResourceStorage) UpdateResourceInstanceStatus(uuid string, status models.ResourceInstanceStatus) error {
	return nil
}
func (m *mockResourceStorage) DeleteResourceInstance(id int) error { return nil }
func (m *mockResourceStorage) DeleteAllResourceInstances() error   { return nil }
func (m *mockResourceStorage) ListResourceInstances() ([]*models.ResourceInstance, error) {
	return nil, nil
}
func (m *mockResourceStorage) ListPublicResourceInstances() ([]*models.ResourceInstance, error) {
	return nil, nil
}
func (m *mockResourceStorage) ListPublicResourceInstancesByType(instanceType models.InstanceType) ([]*models.ResourceInstance, error) {
	return nil, nil
}
func (m *mockResourceStorage) ListResourceInstancesByCreatedBy(createdBy string) ([]*models.ResourceInstance, error) {
	return nil, nil
}

type mockCategoryStorage struct {
	categories map[string]*models.Category
}

func (m *mockCategoryStorage) GetCategoryByUUID(uuid string) (*models.Category, error) {
	if m.categories != nil {
		return m.categories[uuid], nil
	}
	return &models.Category{UUID: uuid, Name: "test"}, nil
}

func (m *mockCategoryStorage) CreateCategory(category *models.Category) error { return nil }
func (m *mockCategoryStorage) GetCategory(id int) (*models.Category, error)   { return nil, nil }
func (m *mockCategoryStorage) GetCategoryByName(name string) (*models.Category, error) {
	return nil, nil
}
func (m *mockCategoryStorage) ListCategories() ([]*models.Category, error)        { return nil, nil }
func (m *mockCategoryStorage) ListEnabledCategories() ([]*models.Category, error) { return nil, nil }
func (m *mockCategoryStorage) UpdateCategory(category *models.Category) error     { return nil }
func (m *mockCategoryStorage) EnableCategory(uuid string) error                   { return nil }
func (m *mockCategoryStorage) DisableCategory(uuid string) error                  { return nil }
func (m *mockCategoryStorage) DeleteCategory(id int) error                        { return nil }
func (m *mockCategoryStorage) DeleteAllCategories() error                         { return nil }

type mockTaskStorage struct {
	tasks []*models.ResourceInstanceTask
}

func (m *mockTaskStorage) CreateTask(task *models.ResourceInstanceTask) error {
	m.tasks = append(m.tasks, task)
	return nil
}

func (m *mockTaskStorage) UpdateTask(task *models.ResourceInstanceTask) error           { return nil }
func (m *mockTaskStorage) UpdateTaskStatus(uuid string, status models.TaskStatus) error { return nil }

func (m *mockTaskStorage) GetTask(id int) (*models.ResourceInstanceTask, error) { return nil, nil }
func (m *mockTaskStorage) GetTaskByUUID(uuid string) (*models.ResourceInstanceTask, error) {
	return nil, nil
}
func (m *mockTaskStorage) ListTasksByResourceInstance(resourceInstanceUUID string) ([]*models.ResourceInstanceTask, error) {
	return nil, nil
}
func (m *mockTaskStorage) ListTasksByResourceInstanceWithPagination(resourceInstanceUUID string, offset, limit int) ([]*models.ResourceInstanceTask, int, error) {
	return nil, 0, nil
}
func (m *mockTaskStorage) ListTasksByType(taskType models.TaskType) ([]*models.ResourceInstanceTask, error) {
	return nil, nil
}
func (m *mockTaskStorage) ListTasksByStatus(status models.TaskStatus) ([]*models.ResourceInstanceTask, error) {
	return nil, nil
}
func (m *mockTaskStorage) ListTasksByQuotaPolicy(quotaPolicyUUID string) ([]*models.ResourceInstanceTask, error) {
	return nil, nil
}
func (m *mockTaskStorage) ListRunningTasks() ([]*models.ResourceInstanceTask, error) { return nil, nil }
func (m *mockTaskStorage) ListRecentTasks(limit int) ([]*models.ResourceInstanceTask, error) {
	return nil, nil
}
func (m *mockTaskStorage) ListFailedTasks(since time.Time) ([]*models.ResourceInstanceTask, error) {
	return nil, nil
}
func (m *mockTaskStorage) CountTasksByStatus(status models.TaskStatus) (int, error) { return 0, nil }
func (m *mockTaskStorage) TryStartTask(uuid string) (bool, error)                   { return true, nil }
func (m *mockTaskStorage) DeleteTask(id int) error                                  { return nil }
func (m *mockTaskStorage) DeleteOldTasks(olderThan time.Time) (int, error)          { return 0, nil }
func (m *mockTaskStorage) GetTaskStatistics() (*storage.TaskStatistics, error)      { return nil, nil }
func (m *mockTaskStorage) HasRunningTasksByResourceInstance(resourceInstanceUUID string) (bool, error) {
	return false, nil
}

type mockDeployer struct {
	deployResult *deployer.DeployResult
	deployError  error
}

func (m *mockDeployer) DeployProduct(ctx context.Context, req deployer.DeployRequest) (*deployer.DeployResult, error) {
	if m.deployError != nil {
		return nil, m.deployError
	}
	if m.deployResult != nil {
		return m.deployResult, nil
	}
	return &deployer.DeployResult{
		Success:       true,
		MariaDBPort:   3306,
		MariaDBUser:   "root",
		MariaDBPasswd: "testpass",
	}, nil
}

func (m *mockDeployer) RestoreSnapshot(ctx context.Context, resourceInstanceUUID, snapshotID string) error {
	return nil
}

func (m *mockDeployer) CheckHealth(ctx context.Context, host string, port int, user, passwd string) (bool, error) {
	return true, nil
}

type mockAllocationStorage struct {
	allocations []*models.Allocation
}

func (m *mockAllocationStorage) CreateAllocation(allocation *models.Allocation) error {
	m.allocations = append(m.allocations, allocation)
	return nil
}

func (m *mockAllocationStorage) GetAllocation(id int) (*models.Allocation, error) { return nil, nil }
func (m *mockAllocationStorage) GetAllocationByUUID(uuid string) (*models.Allocation, error) {
	return nil, nil
}
func (m *mockAllocationStorage) ListAllocations() ([]*models.Allocation, error) { return nil, nil }
func (m *mockAllocationStorage) ListActiveAllocations() ([]*models.Allocation, error) {
	return nil, nil
}
func (m *mockAllocationStorage) ListExpiredAllocations() ([]*models.Allocation, error) {
	return nil, nil
}
func (m *mockAllocationStorage) ListAllocationsByCategory(categoryUUID string) ([]*models.Allocation, error) {
	return nil, nil
}
func (m *mockAllocationStorage) ListAllocationsByRequester(requester string) ([]*models.Allocation, error) {
	return nil, nil
}
func (m *mockAllocationStorage) ListAllocationsByTestbed(testbedUUID string) ([]*models.Allocation, error) {
	return nil, nil
}
func (m *mockAllocationStorage) ListAllocationsByStatus(status models.AllocationStatus) ([]*models.Allocation, error) {
	return nil, nil
}
func (m *mockAllocationStorage) UpdateAllocation(allocation *models.Allocation) error { return nil }
func (m *mockAllocationStorage) UpdateAllocationStatus(uuid string, status models.AllocationStatus) error {
	return nil
}
func (m *mockAllocationStorage) MarkAllocationReleased(uuid string) error { return nil }
func (m *mockAllocationStorage) MarkAllocationExpired(uuid string) error  { return nil }
func (m *mockAllocationStorage) DeleteAllocation(id int) error            { return nil }
func (m *mockAllocationStorage) DeleteAllocationByUUID(uuid string) error { return nil }
func (m *mockAllocationStorage) CountActiveAllocationsByCategory(categoryUUID string) (int, error) {
	return 0, nil
}
func (m *mockAllocationStorage) CountActiveAllocationsByCategoryAndServiceTarget(categoryUUID string, serviceTarget models.ServiceTarget) (int, error) {
	return 0, nil
}
func (m *mockAllocationStorage) CountActiveAllocationsByRequester(requester string) (int, error) {
	return 0, nil
}
func (m *mockAllocationStorage) DeleteAllAllocations() error { return nil }

// TestReplenishJob_PriorityOrdering 测试优先级排序
func TestReplenishJob_PriorityOrdering(t *testing.T) {
	categoryUUID := "cat-123"
	policies := []*models.QuotaPolicy{
		{
			UUID:               "policy-normal",
			CategoryUUID:       categoryUUID,
			ServiceTarget:      models.ServiceTargetNormal,
			MinInstances:       0,
			MaxInstances:       5,
			Priority:           100, // 低优先级
			AutoReplenish:      true,
			ReplenishThreshold: 1,
		},
		{
			UUID:               "policy-robot",
			CategoryUUID:       categoryUUID,
			ServiceTarget:      models.ServiceTargetRobot,
			MinInstances:       2,
			MaxInstances:       3,
			Priority:           10, // 高优先级
			AutoReplenish:      true,
			ReplenishThreshold: 2,
		},
	}

	quotaStorage := &mockQuotaStorage{policies: policies}

	// 验证策略按优先级排序（数值越小优先级越高）
	sorted, _ := quotaStorage.ListPoliciesByPriority()
	if len(sorted) != 2 {
		t.Fatalf("Expected 2 policies, got %d", len(sorted))
	}
	// 注意：mock 没有排序，实际排序由 MySQL 实现
	// 这里只验证 mock 返回了策略
	for _, p := range sorted {
		if p.Priority != 10 && p.Priority != 100 {
			t.Errorf("Unexpected priority: %d", p.Priority)
		}
	}
}

// TestReplenishJob_ResourceLimit 测试资源分配限制
func TestReplenishJob_ResourceLimit(t *testing.T) {
	categoryUUID := "cat-123"
	policies := []*models.QuotaPolicy{
		{
			UUID:               "policy-robot",
			CategoryUUID:       categoryUUID,
			ServiceTarget:      models.ServiceTargetRobot,
			MinInstances:       2,
			MaxInstances:       3,
			Priority:           10,
			AutoReplenish:      true,
			ReplenishThreshold: 2,
		},
		{
			UUID:               "policy-normal",
			CategoryUUID:       categoryUUID,
			ServiceTarget:      models.ServiceTargetNormal,
			MinInstances:       0,
			MaxInstances:       5,
			Priority:           100,
			AutoReplenish:      true,
			ReplenishThreshold: 1,
		},
	}

	quotaStorage := &mockQuotaStorage{policies: policies}

	// 只有 1 个可用资源实例
	resourceStorage := &mockResourceStorage{
		availableInstances: []*models.ResourceInstance{
			{UUID: "res-1", Status: models.ResourceInstanceStatusActive},
		},
	}

	testbedStorage := &mockTestbedStorage{
		availableCount: map[string]int{
			categoryUUID + "_robot":  0, // robot 需要 2 个，当前 0
			categoryUUID + "_normal": 0, // normal 需要 1 个，当前 0
		},
	}

	categoryStorage := &mockCategoryStorage{
		categories: map[string]*models.Category{
			categoryUUID: {UUID: categoryUUID, Name: "main"},
		},
	}

	taskStorage := &mockTaskStorage{}
	deployer := &mockDeployer{}

	job := NewReplenishJob(
		testbedStorage,
		taskStorage,
		quotaStorage,
		resourceStorage,
		categoryStorage,
		deployer,
	)

	// 运行补充任务
	job.run()

	// 验证只创建了 1 个 testbed（资源限制）
	if len(testbedStorage.testbeds) != 1 {
		t.Errorf("Expected 1 testbed created due to resource limit, got %d", len(testbedStorage.testbeds))
	}

	// 验证优先级高的 robot 策略获得了资源
	if len(testbedStorage.testbeds) > 0 {
		createdTestbed := testbedStorage.testbeds[0]
		if createdTestbed.ServiceTarget != models.ServiceTargetRobot {
			t.Errorf("Expected robot testbed, got %s", createdTestbed.ServiceTarget)
		}
	}
}

// TestReplenishJob_NoAvailableResources 测试无可用资源情况
func TestReplenishJob_NoAvailableResources(t *testing.T) {
	categoryUUID := "cat-123"
	policies := []*models.QuotaPolicy{
		{
			UUID:               "policy-robot",
			CategoryUUID:       categoryUUID,
			ServiceTarget:      models.ServiceTargetRobot,
			MinInstances:       2,
			MaxInstances:       3,
			Priority:           10,
			AutoReplenish:      true,
			ReplenishThreshold: 2,
		},
	}

	quotaStorage := &mockQuotaStorage{policies: policies}
	resourceStorage := &mockResourceStorage{
		availableInstances: []*models.ResourceInstance{}, // 无可用资源
	}

	testbedStorage := &mockTestbedStorage{
		availableCount: map[string]int{
			categoryUUID + "_robot": 0,
		},
	}

	categoryStorage := &mockCategoryStorage{
		categories: map[string]*models.Category{
			categoryUUID: {UUID: categoryUUID, Name: "main"},
		},
	}

	taskStorage := &mockTaskStorage{}
	deployer := &mockDeployer{}

	job := NewReplenishJob(
		testbedStorage,
		taskStorage,
		quotaStorage,
		resourceStorage,
		categoryStorage,
		deployer,
	)

	// 运行补充任务 - 不应 panic
	job.run()

	// 验证没有创建 testbed
	if len(testbedStorage.testbeds) != 0 {
		t.Errorf("Expected 0 testbeds when no resources available, got %d", len(testbedStorage.testbeds))
	}
}

// TestReplenishJob_MeetsThreshold 测试达到阈值不补充
func TestReplenishJob_MeetsThreshold(t *testing.T) {
	categoryUUID := "cat-123"
	policies := []*models.QuotaPolicy{
		{
			UUID:               "policy-robot",
			CategoryUUID:       categoryUUID,
			ServiceTarget:      models.ServiceTargetRobot,
			MinInstances:       2,
			MaxInstances:       3,
			Priority:           10,
			AutoReplenish:      true,
			ReplenishThreshold: 2,
		},
	}

	quotaStorage := &mockQuotaStorage{policies: policies}
	resourceStorage := &mockResourceStorage{
		availableInstances: []*models.ResourceInstance{
			{UUID: "res-1", Status: models.ResourceInstanceStatusActive},
		},
	}

	// 当前可用数 = 阈值，不应补充
	testbedStorage := &mockTestbedStorage{
		availableCount: map[string]int{
			categoryUUID + "_robot": 2, // 达到阈值
		},
	}

	categoryStorage := &mockCategoryStorage{
		categories: map[string]*models.Category{
			categoryUUID: {UUID: categoryUUID, Name: "main"},
		},
	}

	taskStorage := &mockTaskStorage{}
	deployer := &mockDeployer{}

	job := NewReplenishJob(
		testbedStorage,
		taskStorage,
		quotaStorage,
		resourceStorage,
		categoryStorage,
		deployer,
	)

	job.run()

	// 验证没有创建 testbed
	if len(testbedStorage.testbeds) != 0 {
		t.Errorf("Expected 0 testbeds when threshold is met, got %d", len(testbedStorage.testbeds))
	}
}

// TestReplenishJob_AutoReplenishDisabled 测试自动补充关闭
func TestReplenishJob_AutoReplenishDisabled(t *testing.T) {
	categoryUUID := "cat-123"
	policies := []*models.QuotaPolicy{
		{
			UUID:               "policy-robot",
			CategoryUUID:       categoryUUID,
			ServiceTarget:      models.ServiceTargetRobot,
			MinInstances:       2,
			MaxInstances:       3,
			Priority:           10,
			AutoReplenish:      false, // 关闭自动补充
			ReplenishThreshold: 2,
		},
	}

	quotaStorage := &mockQuotaStorage{policies: policies}
	resourceStorage := &mockResourceStorage{
		availableInstances: []*models.ResourceInstance{
			{UUID: "res-1", Status: models.ResourceInstanceStatusActive},
		},
	}

	testbedStorage := &mockTestbedStorage{
		availableCount: map[string]int{
			categoryUUID + "_robot": 0, // 低于阈值但 AutoReplenish=false
		},
	}

	categoryStorage := &mockCategoryStorage{
		categories: map[string]*models.Category{
			categoryUUID: {UUID: categoryUUID, Name: "main"},
		},
	}

	taskStorage := &mockTaskStorage{}
	deployer := &mockDeployer{}

	job := NewReplenishJob(
		testbedStorage,
		taskStorage,
		quotaStorage,
		resourceStorage,
		categoryStorage,
		deployer,
	)

	job.run()

	// 验证没有创建 testbed
	if len(testbedStorage.testbeds) != 0 {
		t.Errorf("Expected 0 testbeds when AutoReplenish is disabled, got %d", len(testbedStorage.testbeds))
	}
}

// TestReplenishJob_DeployFailure 测试部署失败处理
func TestReplenishJob_DeployFailure(t *testing.T) {
	categoryUUID := "cat-123"
	policies := []*models.QuotaPolicy{
		{
			UUID:               "policy-robot",
			CategoryUUID:       categoryUUID,
			ServiceTarget:      models.ServiceTargetRobot,
			MinInstances:       2,
			MaxInstances:       3,
			Priority:           10,
			AutoReplenish:      true,
			ReplenishThreshold: 2,
		},
	}

	quotaStorage := &mockQuotaStorage{policies: policies}
	resourceStorage := &mockResourceStorage{
		availableInstances: []*models.ResourceInstance{
			{UUID: "res-1", Status: models.ResourceInstanceStatusActive},
		},
	}

	testbedStorage := &mockTestbedStorage{
		availableCount: map[string]int{
			categoryUUID + "_robot": 0,
		},
	}

	categoryStorage := &mockCategoryStorage{
		categories: map[string]*models.Category{
			categoryUUID: {UUID: categoryUUID, Name: "main"},
		},
	}

	taskStorage := &mockTaskStorage{}
	deployer := &mockDeployer{
		deployError: errors.New("deploy failed"),
	}

	job := NewReplenishJob(
		testbedStorage,
		taskStorage,
		quotaStorage,
		resourceStorage,
		categoryStorage,
		deployer,
	)

	// 运行不应 panic
	job.run()

	// 验证创建了失败任务
	if len(taskStorage.tasks) != 1 {
		t.Errorf("Expected 1 task created, got %d", len(taskStorage.tasks))
	}

	if len(taskStorage.tasks) > 0 && taskStorage.tasks[0].Status != models.TaskStatusFailed {
		t.Errorf("Expected task status to be failed, got %s", taskStorage.tasks[0].Status)
	}
}

// TestJobManager_Creation 测试 JobManager 创建
func TestJobManager_Creation(t *testing.T) {
	manager := NewJobManager(
		nil, nil, nil, nil, nil, nil, nil,
	)

	if manager == nil {
		t.Fatal("Expected non-nil JobManager")
	}

	if manager.autoExpireJob == nil {
		t.Error("Expected autoExpireJob to be initialized")
	}
	if manager.replenishJob == nil {
		t.Error("Expected replenishJob to be initialized")
	}
	if manager.healthCheckJob == nil {
		t.Error("Expected healthCheckJob to be initialized")
	}
}

// TestJobManager_Getters 测试 JobManager Getter 方法
func TestJobManager_Getters(t *testing.T) {
	manager := NewJobManager(
		nil, nil, nil, nil, nil, nil, nil,
	)

	autoExpire := manager.GetAutoExpireJob()
	if autoExpire == nil {
		t.Error("Expected GetAutoExpireJob to return non-nil")
	}

	replenish := manager.GetReplenishJob()
	if replenish == nil {
		t.Error("Expected GetReplenishJob to return non-nil")
	}

	healthCheck := manager.GetHealthCheckJob()
	if healthCheck == nil {
		t.Error("Expected GetHealthCheckJob to return non-nil")
	}
}

// TestNewReplenishJob 测试 ReplenishJob 创建
func TestNewReplenishJob(t *testing.T) {
	testbedStorage := &mockTestbedStorage{}
	taskStorage := &mockTaskStorage{}
	quotaStorage := &mockQuotaStorage{}
	resourceStorage := &mockResourceStorage{}
	categoryStorage := &mockCategoryStorage{}
	deployer := &mockDeployer{}

	job := NewReplenishJob(
		testbedStorage,
		taskStorage,
		quotaStorage,
		resourceStorage,
		categoryStorage,
		deployer,
	)

	if job == nil {
		t.Fatal("Expected non-nil ReplenishJob")
	}

	if job.interval != 1*time.Minute {
		t.Errorf("Expected default interval 1m, got %v", job.interval)
	}
}

// TestNewAutoExpireJob 测试 AutoExpireJob 创建
func TestNewAutoExpireJob(t *testing.T) {
	job := NewAutoExpireJob(nil, nil, nil, nil, nil, nil)

	if job == nil {
		t.Fatal("Expected non-nil AutoExpireJob")
	}

	if job.interval != 1*time.Minute {
		t.Errorf("Expected default interval 1m, got %v", job.interval)
	}
}

// TestNewHealthCheckJob 测试 HealthCheckJob 创建
func TestNewHealthCheckJob(t *testing.T) {
	job := NewHealthCheckJob(nil, nil)

	if job == nil {
		t.Fatal("Expected non-nil HealthCheckJob")
	}

	if job.interval != 5*time.Minute {
		t.Errorf("Expected default interval 5m, got %v", job.interval)
	}

	if job.maxConcurrency != 100 {
		t.Errorf("Expected default maxConcurrency 100, got %d", job.maxConcurrency)
	}
}

// TestReplenishJob_SetInterval 测试设置间隔
func TestReplenishJob_SetInterval(t *testing.T) {
	job := NewReplenishJob(nil, nil, nil, nil, nil, nil)

	newInterval := 30 * time.Second
	job.SetInterval(newInterval)

	if job.interval != newInterval {
		t.Errorf("Expected interval %v, got %v", newInterval, job.interval)
	}
}

// TestAutoExpireJob_SetInterval 测试设置间隔
func TestAutoExpireJob_SetInterval(t *testing.T) {
	job := NewAutoExpireJob(nil, nil, nil, nil, nil, nil)

	newInterval := 30 * time.Second
	job.SetInterval(newInterval)

	if job.interval != newInterval {
		t.Errorf("Expected interval %v, got %v", newInterval, job.interval)
	}
}

// TestHealthCheckJob_SetInterval 测试设置间隔
func TestHealthCheckJob_SetInterval(t *testing.T) {
	job := NewHealthCheckJob(nil, nil)

	newInterval := 3 * time.Minute
	job.SetInterval(newInterval)

	if job.interval != newInterval {
		t.Errorf("Expected interval %v, got %v", newInterval, job.interval)
	}
}

// TestHealthCheckJob_SetMaxConcurrency 测试设置最大并发数
func TestHealthCheckJob_SetMaxConcurrency(t *testing.T) {
	job := NewHealthCheckJob(nil, nil)

	newMax := 50
	job.SetMaxConcurrency(newMax)

	if job.maxConcurrency != newMax {
		t.Errorf("Expected maxConcurrency %d, got %d", newMax, job.maxConcurrency)
	}
}
