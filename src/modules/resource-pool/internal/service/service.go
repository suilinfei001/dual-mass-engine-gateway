package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hugoh/go-designs/resource-pool/internal/deployer"
	"github.com/hugoh/go-designs/resource-pool/internal/models"
	"github.com/hugoh/go-designs/resource-pool/internal/storage"
)

// ResourcePoolService 资源池服务接口
type ResourcePoolService interface {
	// AcquireTestbed 获取 Testbed
	AcquireTestbed(ctx context.Context, categoryUUID, requester string) (*models.Allocation, *models.Testbed, error)

	// AcquireTestbedForRobot 获取 Robot 专用的 Testbed（无需 category）
	AcquireTestbedForRobot(ctx context.Context) (*models.Allocation, *models.Testbed, error)

	// ReleaseTestbed 释放 Testbed
	ReleaseTestbed(ctx context.Context, allocationUUID string) error

	// ExtendAllocation 延长 Allocation 时间
	ExtendAllocation(ctx context.Context, allocationUUID string, additionalSeconds int) error

	// GetAllocation 获取 Allocation
	GetAllocation(uuid string) (*models.Allocation, error)

	// ListMyAllocations 列出我的 Allocation
	ListMyAllocations(requester string) ([]*models.Allocation, error)

	// GetTestbed 获取 Testbed
	GetTestbed(uuid string) (*models.Testbed, error)

	// ListTestbeds 列出 Testbed
	ListTestbeds(status *models.TestbedStatus, categoryUUID *string) ([]*models.Testbed, error)

	// ListTestbedsWithPagination 分页列出 Testbed
	ListTestbedsWithPagination(page, pageSize int, status *models.TestbedStatus, categoryUUID *string) ([]*models.Testbed, int, error)

	// GetCategory 获取 Category
	GetCategory(uuid string) (*models.Category, error)

	// ListCategories 列出 Category
	ListCategories() ([]*models.Category, error)

	// CreateCategory 创建 Category
	CreateCategory(category *models.Category) error

	// UpdateCategory 更新 Category
	UpdateCategory(category *models.Category) error

	// DeleteCategory 删除 Category
	DeleteCategory(uuid string) error

	// GetQuotaPolicy 获取配额策略
	GetQuotaPolicy(categoryUUID string) (*models.QuotaPolicy, error)

	// SetQuotaPolicy 设置配额策略
	SetQuotaPolicy(policy *models.QuotaPolicy) error

	// ListQuotaPolicies 列出配额策略
	ListQuotaPolicies() ([]*models.QuotaPolicy, error)

	// ProvisionTestbed 手动部署 Testbed
	ProvisionTestbed(ctx context.Context, resourceInstanceUUID string) (*models.Testbed, error)

	// GetTask 获取任务
	GetTask(uuid string) (*models.ResourceInstanceTask, error)

	// ListTasksByResourceInstance 列出资源实例的任务
	ListTasksByResourceInstance(resourceInstanceUUID string, page, pageSize int) ([]*models.ResourceInstanceTask, int, error)

	// HasRunningTasksByResourceInstance 检查资源实例是否有运行中的任务
	HasRunningTasksByResourceInstance(resourceInstanceUUID string) (bool, error)

	// ListRecentTasks 列出最近的任务
	ListRecentTasks(limit int) ([]*models.ResourceInstanceTask, error)

	// GetTaskStatistics 获取任务统计
	GetTaskStatistics() (*storage.TaskStatistics, error)

	// ReplenishCategory 手动触发补充指定类别
	ReplenishCategory(categoryUUID string) error
}

// ResourcePoolServiceImpl 资源池服务实现
type ResourcePoolServiceImpl struct {
	testbedStorage     storage.TestbedStorage
	allocationStorage  storage.AllocationStorage
	categoryStorage    storage.CategoryStorage
	quotaStorage       storage.QuotaPolicyStorage
	resourceStorage    storage.ResourceInstanceStorage
	taskStorage        storage.ResourceInstanceTaskStorage
	deployer           deployer.DeployService
	defaultMaxLifetime int // 默认最大生命周期（秒）
}

// NewResourcePoolService 创建资源池服务
func NewResourcePoolService(
	testbedStorage storage.TestbedStorage,
	allocationStorage storage.AllocationStorage,
	categoryStorage storage.CategoryStorage,
	quotaStorage storage.QuotaPolicyStorage,
	resourceStorage storage.ResourceInstanceStorage,
	taskStorage storage.ResourceInstanceTaskStorage,
	deployer deployer.DeployService,
	defaultMaxLifetime int,
) ResourcePoolService {
	if defaultMaxLifetime == 0 {
		defaultMaxLifetime = 86400 // 默认 24 小时
	}

	return &ResourcePoolServiceImpl{
		testbedStorage:     testbedStorage,
		allocationStorage:  allocationStorage,
		categoryStorage:    categoryStorage,
		quotaStorage:       quotaStorage,
		resourceStorage:    resourceStorage,
		taskStorage:        taskStorage,
		deployer:           deployer,
		defaultMaxLifetime: defaultMaxLifetime,
	}
}

// AcquireTestbed 获取 Testbed（基于 service_target 的独立资源池）
//
// 分配逻辑：
// 1. robot 用户只能获取 service_target='robot' 的 testbed
// 2. 普通用户只能获取 service_target='normal' 的 testbed
// 3. 两种资源池互不影响，独立配额管理
func (s *ResourcePoolServiceImpl) AcquireTestbed(ctx context.Context, categoryUUID, requester string) (*models.Allocation, *models.Testbed, error) {
	// 1. 验证 Category 存在且启用
	category, err := s.categoryStorage.GetCategoryByUUID(categoryUUID)
	if err != nil {
		return nil, nil, fmt.Errorf("category not found: %w", err)
	}
	if !category.IsEnabled() {
		return nil, nil, fmt.Errorf("category is disabled")
	}

	// 2. 确定请求者的服务对象类型
	isRobot := IsRobotUser(requester)
	var serviceTarget models.ServiceTarget
	if isRobot {
		serviceTarget = models.ServiceTargetRobot
	} else {
		serviceTarget = models.ServiceTargetNormal
	}

	// 3. 获取该 service_target 对应的配额策略
	policy, err := s.quotaStorage.GetQuotaPolicyByCategoryAndServiceTarget(categoryUUID, serviceTarget)
	if err != nil {
		return nil, nil, fmt.Errorf("quota policy not found for service_target=%s: %w", serviceTarget, err)
	}

	// 4. 检查配额（基于 service_target 的独立配额）
	activeCount, err := s.allocationStorage.CountActiveAllocationsByCategoryAndServiceTarget(categoryUUID, serviceTarget)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count active allocations: %w", err)
	}

	if !policy.CanAllocate(activeCount) {
		return nil, nil, fmt.Errorf("quota exceeded for %s: %d/%d", serviceTarget, activeCount, policy.MaxInstances)
	}

	log.Printf("[AcquireTestbed] %s request: service_target=%s, current=%d/%d",
		requester, serviceTarget, activeCount, policy.MaxInstances)

	// 5. 查找可用的 Testbed（按 service_target 筛选）
	availableTestbeds, err := s.testbedStorage.ListAvailableTestbedsByServiceTarget(categoryUUID, serviceTarget)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list available testbeds: %w", err)
	}

	if len(availableTestbeds) == 0 {
		return nil, nil, fmt.Errorf("no available testbeds for service_target=%s", serviceTarget)
	}

	return s.allocateTestbedWithList(ctx, availableTestbeds, categoryUUID, requester, policy.MaxLifetimeSeconds)
}

// AcquireTestbedForRobot 获取 Robot 专用的 Testbed（无需 category）
// 自动查找有可用 robot testbed 的 category
func (s *ResourcePoolServiceImpl) AcquireTestbedForRobot(ctx context.Context) (*models.Allocation, *models.Testbed, error) {
	// 1. 获取所有 category
	categories, err := s.categoryStorage.ListCategories()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list categories: %w", err)
	}

	// 2. 遍历每个 category，找到有可用 robot testbed 的
	for _, category := range categories {
		if !category.IsEnabled() {
			continue
		}

		// 检查该 category 的 robot 配额
		policy, err := s.quotaStorage.GetQuotaPolicyByCategoryAndServiceTarget(category.UUID, models.ServiceTargetRobot)
		if err != nil {
			continue // 该 category 没有 robot 配额
		}

		// 检查配额
		activeCount, err := s.allocationStorage.CountActiveAllocationsByCategoryAndServiceTarget(category.UUID, models.ServiceTargetRobot)
		if err != nil {
			continue
		}

		if !policy.CanAllocate(activeCount) {
			continue // 配额已满
		}

		// 查找可用的 robot testbed
		availableTestbeds, err := s.testbedStorage.ListAvailableTestbedsByServiceTarget(category.UUID, models.ServiceTargetRobot)
		if err != nil || len(availableTestbeds) == 0 {
			continue
		}

		log.Printf("[AcquireTestbedForRobot] Found category %s with %d available robot testbeds",
			category.Name, len(availableTestbeds))

		return s.allocateTestbedWithList(ctx, availableTestbeds, category.UUID, "robot", policy.MaxLifetimeSeconds)
	}

	return nil, nil, fmt.Errorf("no available robot testbeds in any category")
}

// allocateTestbedWithList 尝试分配可用的 Testbed
func (s *ResourcePoolServiceImpl) allocateTestbedWithList(ctx context.Context, availableTestbeds []*models.Testbed, categoryUUID, requester string, maxLifetimeSeconds int) (*models.Allocation, *models.Testbed, error) {
	// 计算分配过期时间
	maxLifetime := s.defaultMaxLifetime
	if maxLifetimeSeconds > 0 {
		maxLifetime = maxLifetimeSeconds
	}

	// 7. 尝试分配可用的 Testbed，支持并发场景下的重试
	// 当多个请求并发时，CAS 可能失败，需要尝试下一个可用的 Testbed
	var lastErr error

	for _, testbed := range availableTestbeds {
		// 先创建 Allocation 记录（需要 UUID 用于 CAS）
		allocation := models.NewAllocation(testbed.UUID, categoryUUID, requester, maxLifetime)
		err := s.allocationStorage.CreateAllocation(allocation)
		if err != nil {
			lastErr = fmt.Errorf("failed to create allocation: %w", err)
			continue // 尝试下一个 testbed
		}

		// 尝试原子性地分配 Testbed（CAS 操作）
		success, err := s.tryAllocateTestbed(testbed.UUID, allocation.UUID)
		if err != nil {
			lastErr = fmt.Errorf("failed to allocate testbed: %w", err)
			// 清理创建的 allocation 记录
			_ = s.allocationStorage.DeleteAllocationByUUID(allocation.UUID)
			continue // 尝试下一个 testbed
		}

		if success {
			// CAS 成功，分配完成

			// 更新 Allocation 状态为活跃
			allocation.MarkActive(allocation.ExpiresAt)
			err = s.allocationStorage.UpdateAllocation(allocation)
			if err != nil {
				log.Printf("[AcquireTestbed] failed to update allocation status: %v", err)
			}

			// 重新获取完整的 Testbed 信息（包含更新的状态和 CurrentAllocUUID）
			updatedTestbed, err := s.testbedStorage.GetTestbedByUUID(testbed.UUID)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get testbed after allocation: %w", err)
			}

			log.Printf("[AcquireTestbed] Successfully allocated testbed %s to %s (service_target=robot)", testbed.UUID, requester)
			return allocation, updatedTestbed, nil
		}

		// CAS 失败，说明此 testbed 已被其他请求抢占
		// 清理创建的 allocation 记录，尝试下一个 testbed
		log.Printf("[AcquireTestbed] Testbed %s was taken by another request, trying next available testbed", testbed.UUID)
		_ = s.allocationStorage.DeleteAllocationByUUID(allocation.UUID)
	}

	// 所有 testbed 都分配失败
	if lastErr != nil {
		return nil, nil, fmt.Errorf("failed to allocate testbed after %d attempts: %w", len(availableTestbeds), lastErr)
	}
	return nil, nil, fmt.Errorf("no available testbeds (all %d were taken by other concurrent requests)", len(availableTestbeds))
}

// allocateTestbed 分配指定的 Testbed
func (s *ResourcePoolServiceImpl) allocateTestbed(ctx context.Context, testbed *models.Testbed, policy *models.QuotaPolicy, requester string) (*models.Allocation, *models.Testbed, error) {
	// 1. 创建 Allocation 记录
	maxLifetime := policy.MaxLifetimeSeconds
	if maxLifetime == 0 {
		maxLifetime = s.defaultMaxLifetime
	}
	allocation := models.NewAllocation(testbed.UUID, testbed.CategoryUUID, requester, maxLifetime)

	err := s.allocationStorage.CreateAllocation(allocation)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create allocation: %w", err)
	}

	// 2. 原子性地分配 Testbed
	success, err := s.tryAllocateTestbed(testbed.UUID, allocation.UUID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to allocate testbed: %w", err)
	}
	if !success {
		// 分配失败，已被其他请求占用
		return nil, nil, fmt.Errorf("testbed was allocated by another request")
	}

	// 3. 更新 Allocation 状态为活跃
	allocation.MarkActive(allocation.ExpiresAt)
	err = s.allocationStorage.UpdateAllocation(allocation)
	if err != nil {
		log.Printf("[allocateTestbed] failed to update allocation status: %v", err)
	}

	// 4. 重新获取完整的 Testbed 信息
	testbed, err = s.testbedStorage.GetTestbedByUUID(testbed.UUID)
	if err != nil {
		return allocation, nil, fmt.Errorf("failed to get testbed after allocation: %w", err)
	}

	log.Printf("[AcquireTestbed] Successfully allocated testbed %s to %s (priority=%d)",
		testbed.UUID, requester, policy.Priority)
	return allocation, testbed, nil
}

// tryAllocateTestbed 尝试分配 Testbed（原子操作）
func (s *ResourcePoolServiceImpl) tryAllocateTestbed(testbedUUID, allocUUID string) (bool, error) {
	// 使用存储层的 CAS 操作
	if tester, ok := s.testbedStorage.(interface {
		TryAllocateTestbed(string, string) (bool, error)
	}); ok {
		return tester.TryAllocateTestbed(testbedUUID, allocUUID)
	}

	// 回退到非原子方式
	testbed, err := s.testbedStorage.GetTestbedByUUID(testbedUUID)
	if err != nil {
		return false, err
	}

	if !testbed.IsAvailable() {
		return false, nil
	}

	testbed.MarkAllocated(allocUUID)
	err = s.testbedStorage.UpdateTestbed(testbed)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ReleaseTestbed 释放 Testbed
func (s *ResourcePoolServiceImpl) ReleaseTestbed(ctx context.Context, allocationUUID string) error {
	// 1. 获取 Allocation
	allocation, err := s.allocationStorage.GetAllocationByUUID(allocationUUID)
	if err != nil {
		return fmt.Errorf("allocation not found: %w", err)
	}

	if allocation.IsReleased() {
		return fmt.Errorf("allocation already released")
	}

	// 2. 获取关联的 Testbed
	testbed, err := s.testbedStorage.GetTestbedByUUID(allocation.TestbedUUID)
	if err != nil {
		return fmt.Errorf("testbed not found: %w", err)
	}

	// 3. 标记 Testbed 为释放中
	testbed.MarkReleasing()
	err = s.testbedStorage.UpdateTestbed(testbed)
	if err != nil {
		return fmt.Errorf("failed to mark testbed as releasing: %w", err)
	}

	// 4. 标记 Allocation 为已释放
	err = s.allocationStorage.MarkAllocationReleased(allocationUUID)
	if err != nil {
		return fmt.Errorf("failed to mark allocation as released: %w", err)
	}

	// 5. 异步触发快照回滚
	go s.restoreTestbed(testbed)

	log.Printf("[ReleaseTestbed] Successfully released testbed %s", testbed.UUID)
	return nil
}

// restoreTestbed 恢复 ResourceInstance（快照回滚），然后标记 Testbed 为 deleted
// Testbed 是一次性的，释放后应删除，不会回到 available 状态
func (s *ResourcePoolServiceImpl) restoreTestbed(testbed *models.Testbed) {
	log.Printf("[restoreTestbed] Starting restore for testbed %s", testbed.UUID)

	// 获取 ResourceInstance
	resourceInstance, err := s.resourceStorage.GetResourceInstanceByUUID(testbed.ResourceInstanceUUID)
	if err != nil {
		log.Printf("[restoreTestbed] Failed to get resource instance: %v", err)
		s.markTestbedDeleted(testbed) // 即使失败也标记为删除
		return
	}

	// 创建回滚任务记录
	task := models.NewRollbackTask(
		resourceInstance.UUID,
		testbed.UUID,
		"",
		models.TriggerSourceAllocationRelease,
	)
	err = s.taskStorage.CreateTask(task)
	if err != nil {
		log.Printf("[restoreTestbed] Failed to create rollback task: %v", err)
	}

	// 标记任务为运行中
	task.MarkRunning()
	_ = s.taskStorage.UpdateTask(task)

	// 如果是虚拟机，执行快照回滚
	if resourceInstance.IsVirtualMachine() && resourceInstance.SnapshotID != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		err = s.deployer.RestoreSnapshot(ctx, *resourceInstance.SnapshotInstanceUUID, *resourceInstance.SnapshotID)
		if err != nil {
			log.Printf("[restoreTestbed] Snapshot restore failed: %v", err)
			// 标记任务为失败
			task.MarkFailed("RESTORE_FAILED", err.Error())
			_ = s.taskStorage.UpdateTask(task)
			// 即使失败也标记为删除，让管理员处理
		} else {
			log.Printf("[restoreTestbed] Snapshot restore completed for resource instance %s", resourceInstance.UUID)
			// 标记任务为成功
			task.MarkCompleted(true)
			_ = s.taskStorage.UpdateTask(task)
		}
	} else {
		// 非虚拟机或没有快照，标记任务为成功（无需回滚）
		task.MarkCompleted(true)
		_ = s.taskStorage.UpdateTask(task)
	}

	// 标记 Testbed 为已删除（一次性使用）
	s.markTestbedDeleted(testbed)
}

// markTestbedDeleted 标记 Testbed 为已删除
func (s *ResourcePoolServiceImpl) markTestbedDeleted(testbed *models.Testbed) {
	testbed.MarkDeleted()
	err := s.testbedStorage.UpdateTestbed(testbed)
	if err != nil {
		log.Printf("[markTestbedDeleted] Failed to mark testbed deleted: %v", err)
	}
}

// ExtendAllocation 延长 Allocation 时间
func (s *ResourcePoolServiceImpl) ExtendAllocation(ctx context.Context, allocationUUID string, additionalSeconds int) error {
	allocation, err := s.allocationStorage.GetAllocationByUUID(allocationUUID)
	if err != nil {
		return fmt.Errorf("allocation not found: %w", err)
	}

	if !allocation.IsActive() {
		return fmt.Errorf("allocation is not active")
	}

	// 延长过期时间
	if allocation.ExpiresAt != nil {
		newExpiresAt := allocation.ExpiresAt.Add(time.Duration(additionalSeconds) * time.Second)
		allocation.ExpiresAt = &newExpiresAt
	}

	err = s.allocationStorage.UpdateAllocation(allocation)
	if err != nil {
		return fmt.Errorf("failed to extend allocation: %w", err)
	}

	log.Printf("[ExtendAllocation] Extended allocation %s by %d seconds", allocationUUID, additionalSeconds)
	return nil
}

// GetAllocation 获取 Allocation
func (s *ResourcePoolServiceImpl) GetAllocation(uuid string) (*models.Allocation, error) {
	return s.allocationStorage.GetAllocationByUUID(uuid)
}

// ListMyAllocations 列出我的 Allocation
func (s *ResourcePoolServiceImpl) ListMyAllocations(requester string) ([]*models.Allocation, error) {
	return s.allocationStorage.ListAllocationsByRequester(requester)
}

// GetTestbed 获取 Testbed
func (s *ResourcePoolServiceImpl) GetTestbed(uuid string) (*models.Testbed, error) {
	return s.testbedStorage.GetTestbedByUUID(uuid)
}

// ListTestbeds 列出 Testbed
func (s *ResourcePoolServiceImpl) ListTestbeds(status *models.TestbedStatus, categoryUUID *string) ([]*models.Testbed, error) {
	if status != nil && categoryUUID != nil {
		// 需要同时过滤状态和类别
		all, err := s.testbedStorage.ListTestbedsByCategory(*categoryUUID)
		if err != nil {
			return nil, err
		}
		var filtered []*models.Testbed
		for _, t := range all {
			if t.Status == *status {
				filtered = append(filtered, t)
			}
		}
		return filtered, nil
	}

	if status != nil {
		return s.testbedStorage.ListTestbedsByStatus(*status)
	}

	if categoryUUID != nil {
		return s.testbedStorage.ListTestbedsByCategory(*categoryUUID)
	}

	return s.testbedStorage.ListTestbeds()
}

// ListTestbedsWithPagination 分页列出 Testbed
func (s *ResourcePoolServiceImpl) ListTestbedsWithPagination(page, pageSize int, status *models.TestbedStatus, categoryUUID *string) ([]*models.Testbed, int, error) {
	return s.testbedStorage.ListTestbedsWithPagination(page, pageSize, status, categoryUUID)
}

// GetCategory 获取 Category
func (s *ResourcePoolServiceImpl) GetCategory(uuid string) (*models.Category, error) {
	return s.categoryStorage.GetCategoryByUUID(uuid)
}

// ListCategories 列出 Category
func (s *ResourcePoolServiceImpl) ListCategories() ([]*models.Category, error) {
	return s.categoryStorage.ListCategories()
}

// CreateCategory 创建 Category
func (s *ResourcePoolServiceImpl) CreateCategory(category *models.Category) error {
	// 检查名称是否重复
	existing, err := s.categoryStorage.GetCategoryByName(category.Name)
	if err == nil && existing != nil {
		return fmt.Errorf("category name already exists")
	}

	err = s.categoryStorage.CreateCategory(category)
	if err == nil {
		log.Printf("[CreateCategory] Created category %s (uuid=%s, enabled=%v)", category.Name, category.UUID, category.Enabled)
	}
	return err
}

// UpdateCategory 更新 Category
func (s *ResourcePoolServiceImpl) UpdateCategory(category *models.Category) error {
	return s.categoryStorage.UpdateCategory(category)
}

// DeleteCategory 删除 Category
func (s *ResourcePoolServiceImpl) DeleteCategory(uuid string) error {
	// TODO: 检查是否有关联的 Testbed 或配额策略
	category, err := s.categoryStorage.GetCategoryByUUID(uuid)
	if err != nil {
		return err
	}

	err = s.categoryStorage.DeleteCategory(category.ID)
	if err == nil {
		log.Printf("[DeleteCategory] Deleted category %s (uuid=%s)", category.Name, category.UUID)
	}
	return err
}

// GetQuotaPolicy 获取配额策略
func (s *ResourcePoolServiceImpl) GetQuotaPolicy(categoryUUID string) (*models.QuotaPolicy, error) {
	return s.quotaStorage.GetQuotaPolicyByCategory(categoryUUID)
}

// SetQuotaPolicy 设置配额策略
func (s *ResourcePoolServiceImpl) SetQuotaPolicy(policy *models.QuotaPolicy) error {
	// 检查 Category 是否存在
	_, err := s.categoryStorage.GetCategoryByUUID(policy.CategoryUUID)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	// 检查是否已存在相同 category 和 service_target 的策略
	existing, err := s.quotaStorage.GetQuotaPolicyByCategoryAndServiceTarget(policy.CategoryUUID, policy.ServiceTarget)
	if err == nil && existing != nil {
		// 更新现有策略
		policy.ID = existing.ID
		policy.UUID = existing.UUID
		err := s.quotaStorage.UpdateQuotaPolicy(policy)
		if err == nil {
			log.Printf("[SetQuotaPolicy] Updated quota policy for category %s (service_target=%s): min=%d, max=%d, priority=%d",
				policy.CategoryUUID, policy.ServiceTarget, policy.MinInstances, policy.MaxInstances, policy.Priority)
		}
		return err
	}

	// 创建新策略
	err = s.quotaStorage.CreateQuotaPolicy(policy)
	if err == nil {
		log.Printf("[SetQuotaPolicy] Created new quota policy for category %s (service_target=%s): min=%d, max=%d, priority=%d",
			policy.CategoryUUID, policy.ServiceTarget, policy.MinInstances, policy.MaxInstances, policy.Priority)
	}
	return err
}

// ListQuotaPolicies 列出配额策略
func (s *ResourcePoolServiceImpl) ListQuotaPolicies() ([]*models.QuotaPolicy, error) {
	return s.quotaStorage.ListQuotaPolicies()
}

// ProvisionTestbed 手动部署 Testbed
func (s *ResourcePoolServiceImpl) ProvisionTestbed(ctx context.Context, resourceInstanceUUID string) (*models.Testbed, error) {
	log.Printf("[ProvisionTestbed] Starting manual provisioning for resource instance %s", resourceInstanceUUID)

	// 获取 ResourceInstance
	resourceInstance, err := s.resourceStorage.GetResourceInstanceByUUID(resourceInstanceUUID)
	if err != nil {
		return nil, fmt.Errorf("resource instance not found: %w", err)
	}

	if !resourceInstance.IsActive() {
		return nil, fmt.Errorf("resource instance is not active")
	}

	if !resourceInstance.IsVirtualMachine() {
		return nil, fmt.Errorf("only VirtualMachine can be provisioned")
	}

	// 检查资源实例是否有运行中的任务
	hasRunning, err := s.HasRunningTasksByResourceInstance(resourceInstanceUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to check running tasks: %w", err)
	}
	if hasRunning {
		return nil, fmt.Errorf("resource instance has running tasks, cannot deploy")
	}

	// 创建部署任务记录
	task := models.NewDeployTask(
		resourceInstanceUUID,
		"", // categoryUUID, 稍后设置
		models.TriggerSourceManual,
		"admin", // 手动触发，默认为 admin
	)
	err = s.taskStorage.CreateTask(task)
	if err != nil {
		log.Printf("[ProvisionTestbed] Failed to create deploy task: %v", err)
	}

	// 标记任务为运行中
	task.MarkRunning()
	_ = s.taskStorage.UpdateTask(task)

	// 调用 Deployer 部署产品
	deployReq := deployer.DeployRequest{
		ResourceInstanceUUID: resourceInstance.UUID,
		IPAddress:            resourceInstance.IPAddress,
		Port:                 resourceInstance.Port,
		SSHUser:              resourceInstance.SSHUser,
		Passwd:               resourceInstance.Passwd,
		ProductVersion:       "v1.0.0", // 默认版本
		ConfigFile:           "{}",
		EnvVars:              make(map[string]string),
		Timeout:              10 * time.Minute,
	}

	result, err := s.deployer.DeployProduct(ctx, deployReq)
	if err != nil {
		// 标记任务为失败
		task.MarkFailed("DEPLOY_ERROR", err.Error())
		_ = s.taskStorage.UpdateTask(task)
		return nil, fmt.Errorf("deployment failed: %w", err)
	}

	if !result.Success {
		// 标记任务为失败
		task.MarkFailed("DEPLOY_FAILED", result.ErrorMessage)
		_ = s.taskStorage.UpdateTask(task)
		return nil, fmt.Errorf("deployment failed: %s", result.ErrorMessage)
	}

	// 获取默认类别（第一个启用的类别）
	categoryUUID := "default"
	categories, err := s.categoryStorage.ListCategories()
	if err == nil && len(categories) > 0 {
		for _, cat := range categories {
			if cat.Enabled {
				categoryUUID = cat.UUID
				task.CategoryUUID = &categoryUUID
				_ = s.taskStorage.UpdateTask(task)
				break
			}
		}
	}

	// 获取 category 名称用于生成 testbed 名字
	category, err := s.categoryStorage.GetCategoryByUUID(categoryUUID)
	categoryName := "unknown"
	if err == nil && category != nil {
		categoryName = category.Name
	}

	// 创建 Testbed 记录，先生成临时名字
	testbed := models.NewTestbed(
		"temp-testbed", // 临时名字，创建后会更新
		categoryUUID,
		models.ServiceTargetNormal, // 手动部署默认为普通用户池
		resourceInstanceUUID,
		result.MariaDBPort,
		result.MariaDBUser,
		result.MariaDBPasswd,
	)

	err = s.testbedStorage.CreateTestbed(testbed)
	if err != nil {
		// 标记任务为失败
		task.MarkFailed("CREATE_TESTBED_FAILED", err.Error())
		_ = s.taskStorage.UpdateTask(task)
		return nil, fmt.Errorf("failed to create testbed: %w", err)
	}

	// 使用 testbed UUID 生成最终名字
	testbed.Name = models.GenerateTestbedName(categoryName, testbed.UUID)
	err = s.testbedStorage.UpdateTestbed(testbed)
	if err != nil {
		log.Printf("[ProvisionTestbed] Failed to update testbed name: %v", err)
	}

	// 标记任务为成功，并记录结果详情
	task.SetResultDetails(map[string]interface{}{
		"testbed_uuid":    testbed.UUID,
		"mariadb_port":    result.MariaDBPort,
		"mariadb_user":    result.MariaDBUser,
		"product_version": deployReq.ProductVersion,
	})
	task.MarkCompleted(true)
	_ = s.taskStorage.UpdateTask(task)

	log.Printf("[ProvisionTestbed] Successfully provisioned testbed %s", testbed.UUID)
	return testbed, nil
}

// GetTask 获取任务
func (s *ResourcePoolServiceImpl) GetTask(uuid string) (*models.ResourceInstanceTask, error) {
	return s.taskStorage.GetTaskByUUID(uuid)
}

// ListTasksByResourceInstance 列出资源实例的任务
func (s *ResourcePoolServiceImpl) ListTasksByResourceInstance(resourceInstanceUUID string, page, pageSize int) ([]*models.ResourceInstanceTask, int, error) {
	offset := (page - 1) * pageSize
	return s.taskStorage.ListTasksByResourceInstanceWithPagination(resourceInstanceUUID, offset, pageSize)
}

// HasRunningTasksByResourceInstance 检查资源实例是否有运行中的任务
func (s *ResourcePoolServiceImpl) HasRunningTasksByResourceInstance(resourceInstanceUUID string) (bool, error) {
	return s.taskStorage.HasRunningTasksByResourceInstance(resourceInstanceUUID)
}

// ListRecentTasks 列出最近的任务
func (s *ResourcePoolServiceImpl) ListRecentTasks(limit int) ([]*models.ResourceInstanceTask, error) {
	return s.taskStorage.ListRecentTasks(limit)
}

// GetTaskStatistics 获取任务统计
func (s *ResourcePoolServiceImpl) GetTaskStatistics() (*storage.TaskStatistics, error) {
	return s.taskStorage.GetTaskStatistics()
}

// ReplenishCategory 手动触发补充指定类别
func (s *ResourcePoolServiceImpl) ReplenishCategory(categoryUUID string) error {
	// 获取类别的配额策略
	policy, err := s.quotaStorage.GetQuotaPolicyByCategory(categoryUUID)
	if err != nil {
		return fmt.Errorf("failed to get quota policy: %w", err)
	}

	if !policy.AutoReplenish {
		return fmt.Errorf("auto-replenish is not enabled for this category")
	}

	// 统计当前可用的 Testbed 数量（按服务对象区分）
	availableCount, err := s.testbedStorage.CountAvailableTestbedsByCategory(categoryUUID, policy.ServiceTarget)
	if err != nil {
		return fmt.Errorf("failed to count available testbeds: %w", err)
	}

	log.Printf("[ReplenishCategory] Category %s (%s): available=%d, threshold=%d",
		categoryUUID, policy.ServiceTarget, availableCount, policy.ReplenishThreshold)

	// 检查是否需要补充
	if availableCount >= policy.ReplenishThreshold {
		log.Printf("[ReplenishCategory] No replenishment needed")
		return nil
	}

	// 获取类别信息
	category, err := s.categoryStorage.GetCategoryByUUID(categoryUUID)
	if err != nil {
		return fmt.Errorf("failed to get category: %w", err)
	}

	// 计算需要补充的数量
	needed := policy.ReplenishThreshold - availableCount
	if needed < 1 {
		needed = 1
	}

	log.Printf("[ReplenishCategory] Category %s (%s) needs replenishment: %d testbeds",
		categoryUUID, category.Name, needed)

	// 查找可用的 ResourceInstance
	availableInstances, err := s.resourceStorage.ListAvailableResourceInstances()
	if err != nil {
		return fmt.Errorf("failed to list available resource instances: %w", err)
	}

	if len(availableInstances) == 0 {
		log.Printf("[ReplenishCategory] No available resource instances for replenishment")
		return fmt.Errorf("no available resource instances")
	}

	// 补充指定数量的 Testbed
	ctx := context.Background()
	successCount := 0
	for i := 0; i < needed && i < len(availableInstances); i++ {
		instance := availableInstances[i]

		log.Printf("[ReplenishCategory] Provisioning testbed from resource instance %s", instance.UUID)

		// 调用 ProvisionTestbed
		_, err := s.ProvisionTestbed(ctx, instance.UUID)
		if err != nil {
			log.Printf("[ReplenishCategory] Failed to provision testbed from %s: %v", instance.UUID, err)
			continue
		}

		successCount++
		log.Printf("[ReplenishCategory] Successfully provisioned testbed")
	}

	log.Printf("[ReplenishCategory] Replenishment completed: %d/%d testbeds created", successCount, needed)
	return nil
}
