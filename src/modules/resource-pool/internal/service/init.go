package service

import (
	"fmt"
	"log"

	"github.com/hugoh/go-designs/resource-pool/internal/models"
	"github.com/hugoh/go-designs/resource-pool/internal/storage"
)

const (
	// MainCategoryName 主线分类名称（所有testbed共享）
	MainCategoryName = "主线"
	// MainCategoryDescription 主线分类描述
	MainCategoryDescription = "主线分类，所有testbed共享的资源池"
	// RobotUsername robot用户名
	RobotUsername = "robot"
	// RobotPassword robot用户密码
	RobotPassword = "qazxswcde123"
	// RobotQuotaMaxInstances robot用户配额最大实例数
	RobotQuotaMaxInstances = 3
	// RobotQuotaPriority robot用户配额优先级（0为最高）
	RobotQuotaPriority = 0
	// RegularUserPriority 普通用户默认优先级
	RegularUserPriority = 100
	// MainCategoryDefaultQuota 主线分类普通用户默认配额
	MainCategoryDefaultQuota = 50
)

// Initializer 资源池初始化器
type Initializer struct {
	service        *ResourcePoolServiceImpl
	userStorage    storage.UserStorage
	categoryStorage storage.CategoryStorage
	quotaStorage    storage.QuotaPolicyStorage
}

// NewInitializer 创建初始化器
func NewInitializer(
	service *ResourcePoolServiceImpl,
	userStorage storage.UserStorage,
	categoryStorage storage.CategoryStorage,
	quotaStorage storage.QuotaPolicyStorage,
) *Initializer {
	return &Initializer{
		service:        service,
		userStorage:    userStorage,
		categoryStorage: categoryStorage,
		quotaStorage:    quotaStorage,
	}
}

// Initialize 初始化资源池
func (i *Initializer) Initialize() error {
	log.Printf("[Initializer] Starting resource pool initialization...")

	// 1. 初始化主线分类（所有testbed共享的单一资源池）
	category, err := i.initializeMainCategory()
	if err != nil {
		return fmt.Errorf("failed to initialize main category: %w", err)
	}

	// 2. 初始化robot用户
	err = i.initializeRobotUser()
	if err != nil {
		return fmt.Errorf("failed to initialize robot user: %w", err)
	}

	// 3. 初始化配额策略（robot用户优先级0，普通用户优先级100）
	// 注意：由于当前数据库schema限制，配额策略是按类别设置的
	// 我们通过服务层的逻辑来实现按用户优先级分配
	// 这里创建默认的配额策略，优先级100用于普通用户
	err = i.initializeQuotaPolicy(category.UUID)
	if err != nil {
		return fmt.Errorf("failed to initialize quota policy: %w", err)
	}

	log.Printf("[Initializer] Resource pool initialization completed successfully")
	log.Printf("[Initializer] Robot user quota: %d instances (priority %d)", RobotQuotaMaxInstances, RobotQuotaPriority)
	log.Printf("[Initializer] Regular user quota: up to %d instances (priority %d)", MainCategoryDefaultQuota, RegularUserPriority)
	return nil
}

// initializeMainCategory 初始化主线分类（所有testbed共享）
func (i *Initializer) initializeMainCategory() (*models.Category, error) {
	// 检查主线分类是否存在
	category, err := i.categoryStorage.GetCategoryByName(MainCategoryName)
	if err == nil && category != nil {
		log.Printf("[Initializer] Main category already exists: %s", category.UUID)
		return category, nil
	}

	// 创建主线分类
	category = models.NewCategory(MainCategoryName, MainCategoryDescription)
	err = i.categoryStorage.CreateCategory(category)
	if err != nil {
		return nil, fmt.Errorf("failed to create main category: %w", err)
	}

	log.Printf("[Initializer] Created main category: %s", category.UUID)
	return category, nil
}

// initializeRobotUser 初始化robot用户
func (i *Initializer) initializeRobotUser() error {
	// 检查robot用户是否存在
	user, err := i.userStorage.GetUserByUsername(RobotUsername)
	if err == nil && user != nil {
		log.Printf("[Initializer] Robot user already exists: %s", user.Username)
		return nil
	}

	// 创建robot用户（使用bcrypt哈希密码）
	hashedPassword, err := storage.BcryptHashPassword(RobotPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	err = i.userStorage.CreateUser(RobotUsername, hashedPassword, "robot")
	if err != nil {
		return fmt.Errorf("failed to create robot user: %w", err)
	}

	log.Printf("[Initializer] Created robot user: %s", RobotUsername)
	return nil
}

// initializeQuotaPolicy 初始化配额策略
func (i *Initializer) initializeQuotaPolicy(categoryUUID string) error {
	// 检查是否已存在robot配额策略
	policies, err := i.quotaStorage.ListPoliciesByPriority()
	if err == nil && len(policies) > 0 {
		// 检查是否已有robot的配额策略
		for _, p := range policies {
			if p.CategoryUUID == categoryUUID && p.ServiceTarget == models.ServiceTargetRobot {
				log.Printf("[Initializer] Robot quota policy already exists for category %s", categoryUUID)
				return nil
			}
		}
	}

	// 只创建robot用户专属配额策略（优先级0）
	robotPolicy := models.NewQuotaPolicy(categoryUUID, 0, RobotQuotaMaxInstances, RobotQuotaPriority, 86400)
	robotPolicy.AutoReplenish = true
	robotPolicy.ReplenishThreshold = 3
	robotPolicy.ServiceTarget = models.ServiceTargetRobot

	err = i.quotaStorage.CreateQuotaPolicy(robotPolicy)
	if err != nil {
		return fmt.Errorf("failed to create robot quota policy: %w", err)
	}

	log.Printf("[Initializer] Created robot quota policy: priority=%d, max_instances=%d, service_target=%s",
		RobotQuotaPriority, RobotQuotaMaxInstances, models.ServiceTargetRobot)

	return nil
}

// GetRobotQuota 获取robot用户的配额配置（用于服务层）
func GetRobotQuota() (maxInstances int, priority int) {
	return RobotQuotaMaxInstances, RobotQuotaPriority
}

// IsRobotUser 检查是否是robot用户
func IsRobotUser(username string) bool {
	return username == RobotUsername
}
