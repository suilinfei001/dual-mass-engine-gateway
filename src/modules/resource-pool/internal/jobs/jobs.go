package jobs

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/hugoh/go-designs/resource-pool/internal/deployer"
	"github.com/hugoh/go-designs/resource-pool/internal/models"
	"github.com/hugoh/go-designs/resource-pool/internal/storage"
)

// AutoExpireJob 自动过期回收任务
type AutoExpireJob struct {
	allocationStorage storage.AllocationStorage
	testbedStorage    storage.TestbedStorage
	resourceStorage   storage.ResourceInstanceStorage
	taskStorage       storage.ResourceInstanceTaskStorage
	quotaStorage      storage.QuotaPolicyStorage
	deployer          deployer.DeployService
	interval          time.Duration
	stopChan          chan struct{}
}

// NewAutoExpireJob 创建自动过期任务
func NewAutoExpireJob(
	allocationStorage storage.AllocationStorage,
	testbedStorage storage.TestbedStorage,
	resourceStorage storage.ResourceInstanceStorage,
	taskStorage storage.ResourceInstanceTaskStorage,
	quotaStorage storage.QuotaPolicyStorage,
	deployer deployer.DeployService,
) *AutoExpireJob {
	return &AutoExpireJob{
		allocationStorage: allocationStorage,
		testbedStorage:    testbedStorage,
		resourceStorage:   resourceStorage,
		taskStorage:       taskStorage,
		quotaStorage:      quotaStorage,
		deployer:          deployer,
		interval:          1 * time.Minute,
		stopChan:          make(chan struct{}),
	}
}

// Start 启动自动过期任务
func (j *AutoExpireJob) Start() {
	log.Printf("[AutoExpireJob] Starting with interval %v", j.interval)
	ticker := time.NewTicker(j.interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				j.run()
			case <-j.stopChan:
				ticker.Stop()
				log.Printf("[AutoExpireJob] Stopped")
				return
			}
		}
	}()
}

// Stop 停止自动过期任务
func (j *AutoExpireJob) Stop() {
	close(j.stopChan)
}

// run 执行一次过期检查
func (j *AutoExpireJob) run() {
	log.Printf("[AutoExpireJob] Running expired check")

	// 1. 处理已过期的分配（已分配的 testbed）
	j.checkExpiredAllocations()

	// 2. 处理未分配但已过期的 testbed（available 状态但超过策略生命周期）
	j.checkExpiredUnallocatedTestbeds()
}

// checkExpiredAllocations 检查并处理已过期的分配
func (j *AutoExpireJob) checkExpiredAllocations() {
	log.Printf("[AutoExpireJob] Checking expired allocations")

	// 获取所有已过期的分配
	expiredAllocations, err := j.allocationStorage.ListExpiredAllocations()
	if err != nil {
		log.Printf("[AutoExpireJob] Failed to list expired allocations: %v", err)
		return
	}

	if len(expiredAllocations) == 0 {
		log.Printf("[AutoExpireJob] No expired allocations found")
		return
	}

	log.Printf("[AutoExpireJob] Found %d expired allocations", len(expiredAllocations))

	// 处理每个过期的分配
	for _, allocation := range expiredAllocations {
		err := j.expireAllocation(allocation)
		if err != nil {
			log.Printf("[AutoExpireJob] Failed to expire allocation %s: %v", allocation.UUID, err)
		} else {
			log.Printf("[AutoExpireJob] Successfully expired allocation %s", allocation.UUID)
		}
	}
}

// checkExpiredUnallocatedTestbeds 检查并处理未分配但已过期的 testbed
func (j *AutoExpireJob) checkExpiredUnallocatedTestbeds() {
	log.Printf("[AutoExpireJob] Checking expired unallocated testbeds")

	// 获取所有策略，用于计算过期时间
	policies, err := j.quotaStorage.ListPoliciesByPriority()
	if err != nil {
		log.Printf("[AutoExpireJob] Failed to list quota policies: %v", err)
		return
	}

	if len(policies) == 0 {
		log.Printf("[AutoExpireJob] No quota policies found")
		return
	}

	// 构建策略映射表：key = category_uuid + "|" + service_target
	policyMap := make(map[string]*models.QuotaPolicy)
	for _, p := range policies {
		key := p.CategoryUUID + "|" + string(p.ServiceTarget)
		policyMap[key] = p
	}

	// 获取所有可用的 testbed
	testbeds, err := j.testbedStorage.ListTestbeds()
	if err != nil {
		log.Printf("[AutoExpireJob] Failed to list testbeds: %v", err)
		return
	}

	now := time.Now()
	expiredCount := 0

	for _, testbed := range testbeds {
		// 只处理 available 状态且未被分配的 testbed
		if testbed.Status != models.TestbedStatusAvailable || testbed.CurrentAllocUUID != nil {
			continue
		}

		// 获取对应的策略
		policyKey := testbed.CategoryUUID + "|" + string(testbed.ServiceTarget)
		policy, ok := policyMap[policyKey]
		if !ok {
			continue
		}

		// 如果策略没有设置生命周期，跳过
		if policy.MaxLifetimeSeconds <= 0 {
			continue
		}

		// 计算 testbed 的过期时间：创建时间 + 策略生命周期
		expiryTime := testbed.CreatedAt.Add(time.Duration(policy.MaxLifetimeSeconds) * time.Second)

		// 检查是否过期
		if now.After(expiryTime) {
			log.Printf("[AutoExpireJob] Testbed %s (category=%s, service_target=%s) expired at %s",
				testbed.UUID, testbed.CategoryUUID, testbed.ServiceTarget, expiryTime.Format(time.RFC3339))

			// 标记 Testbed 为释放中
			testbed.MarkReleasing()
			err = j.testbedStorage.UpdateTestbed(testbed)
			if err != nil {
				log.Printf("[AutoExpireJob] Failed to mark testbed %s as releasing: %v", testbed.UUID, err)
				continue
			}

			// 异步触发快照回滚（即使未分配也需要回滚，因为可能已部署产品）
			go j.restoreTestbed(testbed)
			expiredCount++
		}
	}

	if expiredCount > 0 {
		log.Printf("[AutoExpireJob] Processing %d expired unallocated testbeds for snapshot restore", expiredCount)
	} else {
		log.Printf("[AutoExpireJob] No expired unallocated testbeds found")
	}
}

// expireAllocation 过期单个分配
func (j *AutoExpireJob) expireAllocation(allocation *models.Allocation) error {
	// 标记分配为过期状态
	err := j.allocationStorage.MarkAllocationExpired(allocation.UUID)
	if err != nil {
		return err
	}

	// 获取关联的 Testbed
	testbed, err := j.testbedStorage.GetTestbedByUUID(allocation.TestbedUUID)
	if err != nil {
		log.Printf("[AutoExpireJob] Testbed not found for allocation %s: %v", allocation.UUID, err)
		return nil
	}

	// 标记 Testbed 为释放中
	testbed.MarkReleasing()
	err = j.testbedStorage.UpdateTestbed(testbed)
	if err != nil {
		log.Printf("[AutoExpireJob] Failed to mark testbed %s as releasing: %v", testbed.UUID, err)
	}

	// 异步触发快照回滚
	go j.restoreTestbed(testbed)

	return nil
}

// restoreTestbed 恢复 ResourceInstance（快照回滚），然后标记 Testbed 为 deleted
// Testbed 是一次性的，释放后应删除，不会回到 available 状态
func (j *AutoExpireJob) restoreTestbed(testbed *models.Testbed) {
	log.Printf("[AutoExpireJob] Starting restore for testbed %s", testbed.UUID)

	// 获取 ResourceInstance
	resourceInstance, err := j.resourceStorage.GetResourceInstanceByUUID(testbed.ResourceInstanceUUID)
	if err != nil {
		log.Printf("[AutoExpireJob] Failed to get resource instance: %v", err)
		j.markTestbedDeleted(testbed) // 即使失败也标记为删除
		return
	}

	// 创建回滚任务记录
	task := models.NewRollbackTask(
		resourceInstance.UUID,
		testbed.UUID,
		"",
		models.TriggerSourceAutoExpire,
	)
	err = j.taskStorage.CreateTask(task)
	if err != nil {
		log.Printf("[AutoExpireJob] Failed to create rollback task: %v", err)
	}

	// 标记任务为运行中
	task.MarkRunning()
	_ = j.taskStorage.UpdateTask(task)

	// 如果是虚拟机，执行快照回滚
	if resourceInstance.IsVirtualMachine() && resourceInstance.SnapshotID != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		err = j.deployer.RestoreSnapshot(ctx, *resourceInstance.SnapshotInstanceUUID, *resourceInstance.SnapshotID)
		if err != nil {
			log.Printf("[AutoExpireJob] Snapshot restore failed: %v", err)
			// 标记任务为失败
			task.MarkFailed("RESTORE_FAILED", err.Error())
			_ = j.taskStorage.UpdateTask(task)
			// 即使失败也标记为删除，让管理员处理
		} else {
			log.Printf("[AutoExpireJob] Snapshot restore completed for resource instance %s", resourceInstance.UUID)
			// 标记任务为成功
			task.MarkCompleted(true)
			_ = j.taskStorage.UpdateTask(task)
		}
	} else {
		// 非虚拟机或没有快照，标记任务为成功（无需回滚）
		task.MarkCompleted(true)
		_ = j.taskStorage.UpdateTask(task)
	}

	// 标记 Testbed 为已删除（一次性使用）
	j.markTestbedDeleted(testbed)
}

// markTestbedDeleted 标记 Testbed 为已删除
func (j *AutoExpireJob) markTestbedDeleted(testbed *models.Testbed) {
	testbed.MarkDeleted()
	err := j.testbedStorage.UpdateTestbed(testbed)
	if err != nil {
		log.Printf("[AutoExpireJob] Failed to mark testbed deleted: %v", err)
	}
}

// SetInterval 设置检查间隔
func (j *AutoExpireJob) SetInterval(interval time.Duration) {
	j.interval = interval
}

// ReplenishJob 自动补充任务
type ReplenishJob struct {
	testbedStorage  storage.TestbedStorage
	taskStorage     storage.ResourceInstanceTaskStorage
	quotaStorage    storage.QuotaPolicyStorage
	resourceStorage storage.ResourceInstanceStorage
	categoryStorage storage.CategoryStorage
	deployer        deployer.DeployService
	interval        time.Duration
	stopChan        chan struct{}
}

// NewReplenishJob 创建自动补充任务
func NewReplenishJob(
	testbedStorage storage.TestbedStorage,
	taskStorage storage.ResourceInstanceTaskStorage,
	quotaStorage storage.QuotaPolicyStorage,
	resourceStorage storage.ResourceInstanceStorage,
	categoryStorage storage.CategoryStorage,
	deployer deployer.DeployService,
) *ReplenishJob {
	return &ReplenishJob{
		testbedStorage:  testbedStorage,
		taskStorage:     taskStorage,
		quotaStorage:    quotaStorage,
		resourceStorage: resourceStorage,
		categoryStorage: categoryStorage,
		deployer:        deployer,
		interval:        1 * time.Minute,
		stopChan:        make(chan struct{}),
	}
}

// Start 启动自动补充任务
func (j *ReplenishJob) Start() {
	log.Printf("[ReplenishJob] Starting with interval %v", j.interval)
	ticker := time.NewTicker(j.interval)

	go func() {
		// 首次启动立即执行一次检查
		j.run()

		for {
			select {
			case <-ticker.C:
				j.run()
			case <-j.stopChan:
				ticker.Stop()
				log.Printf("[ReplenishJob] Stopped")
				return
			}
		}
	}()
}

// Stop 停止自动补充任务
func (j *ReplenishJob) Stop() {
	close(j.stopChan)
}

// run 执行一次补充检查
func (j *ReplenishJob) run() {
	log.Printf("[ReplenishJob] Running replenish check")

	// 获取所有配额策略（已按优先级升序排列，数值越小优先级越高）
	policies, err := j.quotaStorage.ListPoliciesByPriority()
	if err != nil {
		log.Printf("[ReplenishJob] Failed to list quota policies: %v", err)
		return
	}

	// 获取当前可用的资源实例数量
	availableInstances, err := j.resourceStorage.ListAvailableResourceInstances()
	if err != nil {
		log.Printf("[ReplenishJob] Failed to list available resource instances: %v", err)
		return
	}
	totalAvailableInstances := len(availableInstances)

	// 第一轮：检查所有高优先级策略是否已满足
	// 只处理需要补充的策略，计算总需求量
	totalNeeded := 0
	for _, policy := range policies {
		if !policy.AutoReplenish {
			continue
		}

		availableCount, err := j.testbedStorage.CountAvailableTestbedsByCategory(policy.CategoryUUID, policy.ServiceTarget)
		if err != nil {
			continue
		}

		if policy.ShouldReplenish(availableCount) {
			needed := policy.ReplenishThreshold - availableCount
			if needed < 1 {
				needed = 1
			}
			totalNeeded += needed
		}
	}

	// 如果总需求超过可用资源实例数，按优先级顺序处理，资源用完即止
	usedInstances := 0
	for _, policy := range policies {
		if !policy.AutoReplenish {
			continue
		}

		// 检查是否还有可用资源实例
		if usedInstances >= totalAvailableInstances {
			log.Printf("[ReplenishJob] No more available resource instances, skipping lower priority policies")
			break
		}

		// 计算当前策略还能使用多少个资源实例
		remainingInstances := totalAvailableInstances - usedInstances
		err := j.checkAndReplenishWithLimit(policy, remainingInstances, &usedInstances)
		if err != nil {
			log.Printf("[ReplenishJob] Failed to replenish category %s: %v", policy.CategoryUUID, err)
		}
	}
}

// checkAndReplenish 检查并补充指定类别
func (j *ReplenishJob) checkAndReplenish(policy *models.QuotaPolicy) error {
	// 统计当前可用的 Testbed 数量（按服务对象区分）
	availableCount, err := j.testbedStorage.CountAvailableTestbedsByCategory(policy.CategoryUUID, policy.ServiceTarget)
	if err != nil {
		return err
	}

	log.Printf("[ReplenishJob] Category %s (%s): available=%d, threshold=%d",
		policy.CategoryUUID, policy.ServiceTarget, availableCount, policy.ReplenishThreshold)

	// 检查是否需要补充
	if !policy.ShouldReplenish(availableCount) {
		return nil
	}

	// 获取类别信息
	category, err := j.categoryStorage.GetCategoryByUUID(policy.CategoryUUID)
	if err != nil {
		log.Printf("[ReplenishJob] Failed to get category %s: %v", policy.CategoryUUID, err)
		return err
	}

	// 计算需要补充的数量
	needed := policy.ReplenishThreshold - availableCount
	if needed < 1 {
		needed = 1
	}

	log.Printf("[ReplenishJob] Category %s (%s/%s) needs replenishment: %d testbeds",
		policy.CategoryUUID, policy.ServiceTarget, category.Name, needed)

	// 查找可用的 ResourceInstance
	availableInstances, err := j.resourceStorage.ListAvailableResourceInstances()
	if err != nil {
		log.Printf("[ReplenishJob] Failed to list available resource instances: %v", err)
		return err
	}

	if len(availableInstances) == 0 {
		log.Printf("[ReplenishJob] No available resource instances for replenishment")
		return nil
	}

	// 补充指定数量的 Testbed
	// 注意：必须遍历所有可用实例，因为有些实例可能被跳过（有运行任务、不可达、连续失败等）
	successCount := 0
	for i := 0; i < len(availableInstances) && successCount < needed; i++ {
		instance := availableInstances[i]

		// 1. 检查资源实例是否有运行中的任务
		hasRunning, err := j.taskStorage.HasRunningTasksByResourceInstance(instance.UUID)
		if err != nil {
			log.Printf("[ReplenishJob] Failed to check running tasks: %v", err)
			continue
		}
		if hasRunning {
			log.Printf("[ReplenishJob] Skipping resource instance %s, has running tasks", instance.UUID)
			continue
		}

		// 2. 验证资源实例的可达性（部署前必须确认实例是健康的）
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		healthy, healthErr := j.deployer.CheckHealth(ctx, instance.IPAddress, instance.Port, instance.SSHUser, instance.Passwd)
		cancel()
		if !healthy {
			log.Printf("[ReplenishJob] Skipping resource instance %s, health check failed: %v", instance.UUID, healthErr)
			// 更新资源实例状态为不可达
			_ = j.resourceStorage.UpdateResourceInstanceStatus(instance.UUID, models.ResourceInstanceStatusUnreachable)
			continue
		}
		log.Printf("[ReplenishJob] Resource instance %s is healthy, proceeding with deployment", instance.UUID)

		// 检查资源实例的最近部署失败次数（防止无限重试）
		// 统计最近的自动补充部署任务，如果有3次连续失败，则跳过该资源实例
		recentTasks, err := j.taskStorage.ListTasksByResourceInstance(instance.UUID)
		if err == nil {
			consecutiveFailures := 0
			maxConsecutiveFailures := 3 // 最多允许3次连续失败

			for _, t := range recentTasks {
				// 只统计自动补充触发的部署任务
				if t.TaskType == models.TaskTypeDeploy && t.TriggerSource == models.TriggerSourceAutoReplenish {
					if t.Status == models.TaskStatusFailed {
						consecutiveFailures++
						// 如果有成功任务，重置计数
					} else if t.Status == models.TaskStatusCompleted {
						break
					}
				}
				// 只检查最近的几个任务
				if consecutiveFailures >= maxConsecutiveFailures {
					break
				}
			}

			if consecutiveFailures >= maxConsecutiveFailures {
				log.Printf("[ReplenishJob] Skipping resource instance %s, %d consecutive failures detected (max: %d)",
					instance.UUID, consecutiveFailures, maxConsecutiveFailures)
				continue
			}
		}

		log.Printf("[ReplenishJob] Provisioning testbed from resource instance %s", instance.UUID)

		// 创建部署任务记录
		task := models.NewAutoReplenishDeployTask(
			instance.UUID,
			policy.UUID,
			policy.CategoryUUID,
		)
		err = j.taskStorage.CreateTask(task)
		if err != nil {
			log.Printf("[ReplenishJob] Failed to create deploy task: %v", err)
			// 如果是外键约束错误，策略可能已被删除，跳过此部署
			if strings.Contains(err.Error(), "foreign key constraint") {
				log.Printf("[ReplenishJob] Skipping deployment due to foreign key constraint, policy may have been recreated")
				continue
			}
		}

		// 标记任务为运行中
		task.MarkRunning()
		_ = j.taskStorage.UpdateTask(task)

		// 调用部署服务进行产品部署
		deployReq := deployer.DeployRequest{
			ResourceInstanceUUID: instance.UUID,
			IPAddress:            instance.IPAddress,
			Port:                 instance.Port,
			SSHUser:              instance.SSHUser,
			Passwd:               instance.Passwd,
			ProductVersion:       "v1.0.0",
			ConfigFile:           "{}",
			EnvVars:              make(map[string]string),
			Timeout:              10 * time.Minute,
		}

		ctx = context.Background()
		result, err := j.deployer.DeployProduct(ctx, deployReq)
		if err != nil {
			log.Printf("[ReplenishJob] Failed to deploy product on %s: %v", instance.UUID, err)
			task.MarkFailed("DEPLOY_ERROR", err.Error())
			_ = j.taskStorage.UpdateTask(task)
			continue
		}

		if !result.Success {
			log.Printf("[ReplenishJob] Deployment failed on %s: %s", instance.UUID, result.ErrorMessage)
			task.MarkFailed("DEPLOY_FAILED", result.ErrorMessage)
			_ = j.taskStorage.UpdateTask(task)
			continue
		}

		// 创建 Testbed 记录，先生成临时名字
		testbed := models.NewTestbed(
			"temp-testbed", // 临时名字，创建后会更新
			category.UUID,
			policy.ServiceTarget,
			instance.UUID,
			result.MariaDBPort,
			result.MariaDBUser,
			result.MariaDBPasswd,
		)

		err = j.testbedStorage.CreateTestbed(testbed)
		if err != nil {
			log.Printf("[ReplenishJob] Failed to create testbed record: %v", err)
			task.MarkFailed("CREATE_TESTBED_FAILED", err.Error())
			_ = j.taskStorage.UpdateTask(task)
			continue
		}

		// 使用 testbed UUID 生成最终名字
		testbed.Name = models.GenerateTestbedName(category.Name, testbed.UUID)
		err = j.testbedStorage.UpdateTestbed(testbed)
		if err != nil {
			log.Printf("[ReplenishJob] Failed to update testbed name: %v", err)
		}

		// 标记任务为成功
		task.SetResultDetails(map[string]interface{}{
			"testbed_uuid": testbed.UUID,
			"mariadb_port": result.MariaDBPort,
			"mariadb_user": result.MariaDBUser,
		})
		task.MarkCompleted(true)
		_ = j.taskStorage.UpdateTask(task)

		successCount++
		log.Printf("[ReplenishJob] Successfully provisioned testbed %s", testbed.UUID)
	}

	log.Printf("[ReplenishJob] Replenishment completed: %d/%d testbeds created", successCount, needed)
	return nil
}

// checkAndReplenishWithLimit 检查并补充指定类别，受限于可用资源实例数量
// maxInstances: 当前策略最多可使用的资源实例数量
// usedCounter: 指针，记录已使用的资源实例总数
func (j *ReplenishJob) checkAndReplenishWithLimit(policy *models.QuotaPolicy, maxInstances int, usedCounter *int) error {
	// 统计当前可用的 Testbed 数量（按服务对象区分）
	availableCount, err := j.testbedStorage.CountAvailableTestbedsByCategory(policy.CategoryUUID, policy.ServiceTarget)
	if err != nil {
		return err
	}

	log.Printf("[ReplenishJob] Category %s (%s): available=%d, threshold=%d",
		policy.CategoryUUID, policy.ServiceTarget, availableCount, policy.ReplenishThreshold)

	// 检查是否需要补充
	if !policy.ShouldReplenish(availableCount) {
		return nil
	}

	// 获取类别信息
	category, err := j.categoryStorage.GetCategoryByUUID(policy.CategoryUUID)
	if err != nil {
		log.Printf("[ReplenishJob] Failed to get category %s: %v", policy.CategoryUUID, err)
		return err
	}

	// 计算需要补充的数量
	needed := policy.ReplenishThreshold - availableCount
	if needed < 1 {
		needed = 1
	}

	// 受限于可用的资源实例数量
	if needed > maxInstances {
		needed = maxInstances
		log.Printf("[ReplenishJob] Category %s (%s/%s) limited to %d testbeds due to resource constraints",
			policy.CategoryUUID, policy.ServiceTarget, category.Name, needed)
	} else {
		log.Printf("[ReplenishJob] Category %s (%s/%s) needs replenishment: %d testbeds",
			policy.CategoryUUID, policy.ServiceTarget, category.Name, needed)
	}

	if needed <= 0 {
		return nil
	}

	// 查找可用的 ResourceInstance
	availableInstances, err := j.resourceStorage.ListAvailableResourceInstances()
	if err != nil {
		log.Printf("[ReplenishJob] Failed to list available resource instances: %v", err)
		return err
	}

	if len(availableInstances) == 0 {
		log.Printf("[ReplenishJob] No available resource instances for replenishment")
		return nil
	}

	// 补充指定数量的 Testbed
	// 注意：必须遍历所有可用实例，因为有些实例可能被跳过（有运行任务、不可达、连续失败等）
	successCount := 0
	for i := 0; i < len(availableInstances) && successCount < needed && (*usedCounter) < len(availableInstances); i++ {
		instance := availableInstances[i]

		// 1. 检查资源实例是否有运行中的任务
		hasRunning, err := j.taskStorage.HasRunningTasksByResourceInstance(instance.UUID)
		if err != nil {
			log.Printf("[ReplenishJob] Failed to check running tasks: %v", err)
			continue
		}
		if hasRunning {
			log.Printf("[ReplenishJob] Skipping resource instance %s, has running tasks", instance.UUID)
			continue
		}

		// 2. 验证资源实例的可达性（部署前必须确认实例是健康的）
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		healthy, healthErr := j.deployer.CheckHealth(ctx, instance.IPAddress, instance.Port, instance.SSHUser, instance.Passwd)
		cancel()
		if !healthy {
			log.Printf("[ReplenishJob] Skipping resource instance %s, health check failed: %v", instance.UUID, healthErr)
			// 更新资源实例状态为不可达
			_ = j.resourceStorage.UpdateResourceInstanceStatus(instance.UUID, models.ResourceInstanceStatusUnreachable)
			continue
		}
		log.Printf("[ReplenishJob] Resource instance %s is healthy, proceeding with deployment", instance.UUID)

		// 检查资源实例的最近部署失败次数（防止无限重试）
		// 统计最近的自动补充部署任务，如果有3次连续失败，则跳过该资源实例
		recentTasks, err := j.taskStorage.ListTasksByResourceInstance(instance.UUID)
		if err == nil {
			consecutiveFailures := 0
			maxConsecutiveFailures := 3 // 最多允许3次连续失败

			for _, t := range recentTasks {
				// 只统计自动补充触发的部署任务
				if t.TaskType == models.TaskTypeDeploy && t.TriggerSource == models.TriggerSourceAutoReplenish {
					if t.Status == models.TaskStatusFailed {
						consecutiveFailures++
						// 如果有成功任务，重置计数
					} else if t.Status == models.TaskStatusCompleted {
						break
					}
				}
				// 只检查最近的几个任务
				if consecutiveFailures >= maxConsecutiveFailures {
					break
				}
			}

			if consecutiveFailures >= maxConsecutiveFailures {
				log.Printf("[ReplenishJob] Skipping resource instance %s, %d consecutive failures detected (max: %d)",
					instance.UUID, consecutiveFailures, maxConsecutiveFailures)
				continue
			}
		}

		log.Printf("[ReplenishJob] Provisioning testbed from resource instance %s", instance.UUID)

		// 创建部署任务记录
		task := models.NewAutoReplenishDeployTask(
			instance.UUID,
			policy.UUID,
			policy.CategoryUUID,
		)
		err = j.taskStorage.CreateTask(task)
		if err != nil {
			log.Printf("[ReplenishJob] Failed to create deploy task: %v", err)
			if strings.Contains(err.Error(), "foreign key constraint") {
				log.Printf("[ReplenishJob] Skipping deployment due to foreign key constraint, policy may have been recreated")
				continue
			}
		}

		// 标记任务为运行中
		task.MarkRunning()
		_ = j.taskStorage.UpdateTask(task)

		// 调用部署服务进行产品部署
		deployReq := deployer.DeployRequest{
			ResourceInstanceUUID: instance.UUID,
			IPAddress:            instance.IPAddress,
			Port:                 instance.Port,
			SSHUser:              instance.SSHUser,
			Passwd:               instance.Passwd,
			ProductVersion:       "v1.0.0",
			ConfigFile:           "{}",
			EnvVars:              make(map[string]string),
			Timeout:              10 * time.Minute,
		}

		ctx = context.Background()
		result, err := j.deployer.DeployProduct(ctx, deployReq)
		if err != nil {
			log.Printf("[ReplenishJob] Failed to deploy product on %s: %v", instance.UUID, err)
			task.MarkFailed("DEPLOY_ERROR", err.Error())
			_ = j.taskStorage.UpdateTask(task)
			continue
		}

		if !result.Success {
			log.Printf("[ReplenishJob] Deployment failed on %s: %s", instance.UUID, result.ErrorMessage)
			task.MarkFailed("DEPLOY_FAILED", result.ErrorMessage)
			_ = j.taskStorage.UpdateTask(task)
			continue
		}

		// 创建 Testbed 记录，先生成临时名字
		testbed := models.NewTestbed(
			"temp-testbed", // 临时名字，创建后会更新
			category.UUID,
			policy.ServiceTarget,
			instance.UUID,
			result.MariaDBPort,
			result.MariaDBUser,
			result.MariaDBPasswd,
		)

		err = j.testbedStorage.CreateTestbed(testbed)
		if err != nil {
			log.Printf("[ReplenishJob] Failed to create testbed record: %v", err)
			task.MarkFailed("CREATE_TESTBED_FAILED", err.Error())
			_ = j.taskStorage.UpdateTask(task)
			continue
		}

		// 使用 testbed UUID 生成最终名字
		testbed.Name = models.GenerateTestbedName(category.Name, testbed.UUID)
		err = j.testbedStorage.UpdateTestbed(testbed)
		if err != nil {
			log.Printf("[ReplenishJob] Failed to update testbed name: %v", err)
		}

		// 标记任务为成功
		task.SetResultDetails(map[string]interface{}{
			"testbed_uuid": testbed.UUID,
			"mariadb_port": result.MariaDBPort,
			"mariadb_user": result.MariaDBUser,
		})
		task.MarkCompleted(true)
		_ = j.taskStorage.UpdateTask(task)

		successCount++
		(*usedCounter)++ // 增加已使用的资源实例计数
		log.Printf("[ReplenishJob] Successfully provisioned testbed %s", testbed.UUID)
	}

	log.Printf("[ReplenishJob] Replenishment completed: %d/%d testbeds created", successCount, needed)
	return nil
}

// SetInterval 设置检查间隔
func (j *ReplenishJob) SetInterval(interval time.Duration) {
	j.interval = interval
}

// HealthCheckJob 健康检查任务
type HealthCheckJob struct {
	resourceStorage storage.ResourceInstanceStorage
	deployer        deployer.DeployService
	interval        time.Duration
	maxConcurrency  int
	stopChan        chan struct{}
}

// NewHealthCheckJob 创建健康检查任务
func NewHealthCheckJob(resourceStorage storage.ResourceInstanceStorage, deployer deployer.DeployService) *HealthCheckJob {
	return &HealthCheckJob{
		resourceStorage: resourceStorage,
		deployer:        deployer,
		interval:        5 * time.Minute,
		maxConcurrency:  100, // 最大并发数
		stopChan:        make(chan struct{}),
	}
}

// Start 启动健康检查任务
func (j *HealthCheckJob) Start() {
	log.Printf("[HealthCheckJob] Starting with interval %v, max concurrency %d", j.interval, j.maxConcurrency)
	ticker := time.NewTicker(j.interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				j.run()
			case <-j.stopChan:
				ticker.Stop()
				log.Printf("[HealthCheckJob] Stopped")
				return
			}
		}
	}()
}

// Stop 停止健康检查任务
func (j *HealthCheckJob) Stop() {
	close(j.stopChan)
}

// healthCheckResult 健康检查结果
type healthCheckResult struct {
	uuid      string
	healthy   bool
	newStatus models.ResourceInstanceStatus
	err       error
}

// run 执行一次健康检查
func (j *HealthCheckJob) run() {
	log.Printf("[HealthCheckJob] Running health check")

	// 获取所有资源实例
	instances, err := j.resourceStorage.ListResourceInstances()
	if err != nil {
		log.Printf("[HealthCheckJob] Failed to list resource instances: %v", err)
		return
	}

	if len(instances) == 0 {
		log.Printf("[HealthCheckJob] No resource instances to check")
		return
	}

	log.Printf("[HealthCheckJob] Checking %d resource instances", len(instances))

	// 并发健康检查，最大并发数为 maxConcurrency
	results := j.checkInstancesConcurrent(instances)

	// 更新状态
	updatedCount := 0
	for _, result := range results {
		// 记录错误但继续处理状态更新
		if result.err != nil {
			log.Printf("[HealthCheckJob] Health check error for %s: %v", result.uuid, result.err)
			// 不要 continue，即使有错误也要更新状态
		}

		// 获取当前实例状态
		instance, err := j.resourceStorage.GetResourceInstanceByUUID(result.uuid)
		if err != nil {
			log.Printf("[HealthCheckJob] Failed to get instance %s: %v", result.uuid, err)
			continue
		}

		// 只有状态改变时才更新
		if instance.Status != result.newStatus {
			err = j.resourceStorage.UpdateResourceInstanceStatus(result.uuid, result.newStatus)
			if err != nil {
				log.Printf("[HealthCheckJob] Failed to update status for %s: %v", result.uuid, err)
			} else {
				updatedCount++
				log.Printf("[HealthCheckJob] Updated %s status from %s to %s (healthy=%v)",
					result.uuid, instance.Status, result.newStatus, result.healthy)
			}
		}
	}

	log.Printf("[HealthCheckJob] Health check completed: %d instances checked, %d status updated",
		len(instances), updatedCount)
}

// checkInstancesConcurrent 并发检查资源实例
func (j *HealthCheckJob) checkInstancesConcurrent(instances []*models.ResourceInstance) []healthCheckResult {
	results := make([]healthCheckResult, len(instances))

	// 使用信号量控制并发数
	semaphore := make(chan struct{}, j.maxConcurrency)
	var wg sync.WaitGroup

	for i, instance := range instances {
		wg.Add(1)
		go func(idx int, inst *models.ResourceInstance) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			results[idx] = j.checkInstance(inst)
		}(i, instance)
	}

	// 等待所有检查完成
	wg.Wait()

	return results
}

// checkInstance 检查单个资源实例
func (j *HealthCheckJob) checkInstance(instance *models.ResourceInstance) healthCheckResult {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	healthy, err := j.deployer.CheckHealth(ctx, instance.IPAddress, instance.Port, instance.SSHUser, instance.Passwd)

	var newStatus models.ResourceInstanceStatus
	if healthy {
		newStatus = models.ResourceInstanceStatusActive
	} else {
		newStatus = models.ResourceInstanceStatusUnreachable
	}

	return healthCheckResult{
		uuid:      instance.UUID,
		healthy:   healthy,
		newStatus: newStatus,
		err:       err,
	}
}

// SetInterval 设置检查间隔
func (j *HealthCheckJob) SetInterval(interval time.Duration) {
	j.interval = interval
}

// SetMaxConcurrency 设置最大并发数
func (j *HealthCheckJob) SetMaxConcurrency(max int) {
	j.maxConcurrency = max
}

// JobManager 任务管理器
type JobManager struct {
	autoExpireJob  *AutoExpireJob
	replenishJob   *ReplenishJob
	healthCheckJob *HealthCheckJob
}

// NewJobManager 创建任务管理器
func NewJobManager(
	allocationStorage storage.AllocationStorage,
	testbedStorage storage.TestbedStorage,
	taskStorage storage.ResourceInstanceTaskStorage,
	quotaStorage storage.QuotaPolicyStorage,
	resourceStorage storage.ResourceInstanceStorage,
	categoryStorage storage.CategoryStorage,
	deployer deployer.DeployService,
) *JobManager {
	return &JobManager{
		autoExpireJob:  NewAutoExpireJob(allocationStorage, testbedStorage, resourceStorage, taskStorage, quotaStorage, deployer),
		replenishJob:   NewReplenishJob(testbedStorage, taskStorage, quotaStorage, resourceStorage, categoryStorage, deployer),
		healthCheckJob: NewHealthCheckJob(resourceStorage, deployer),
	}
}

// Start 启动所有任务
func (m *JobManager) Start() {
	m.autoExpireJob.Start()
	m.replenishJob.Start()
	m.healthCheckJob.Start()
}

// Stop 停止所有任务
func (m *JobManager) Stop() {
	m.autoExpireJob.Stop()
	m.replenishJob.Stop()
	m.healthCheckJob.Stop()
}

// GetAutoExpireJob 获取自动过期任务
func (m *JobManager) GetAutoExpireJob() *AutoExpireJob {
	return m.autoExpireJob
}

// GetReplenishJob 获取补充任务
func (m *JobManager) GetReplenishJob() *ReplenishJob {
	return m.replenishJob
}

// GetHealthCheckJob 获取健康检查任务
func (m *JobManager) GetHealthCheckJob() *HealthCheckJob {
	return m.healthCheckJob
}
