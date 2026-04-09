package factory

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hugoh/go-designs/resource-pool/internal/models"
)

// TestFixture 测试夹具
type TestFixture struct {
	DB *sql.DB
}

// NewTestFixture 创建测试夹具
func NewTestFixture(db *sql.DB) *TestFixture {
	return &TestFixture{DB: db}
}

// Cleanup 清理所有测试数据
func (f *TestFixture) Cleanup() error {
	// 按依赖关系逆序删除
	tables := []string{
		"allocations",
		"testbeds",
		"resource_instances",
		"quota_policies",
		"categories",
	}

	for _, table := range tables {
		_, err := f.DB.Exec(fmt.Sprintf("DELETE FROM %s WHERE uuid LIKE 'test-%%'", table))
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateCategory 创建测试 Category
func (f *TestFixture) CreateCategory(name string) *models.Category {
	if name == "" {
		name = "test-category-" + uuid.New().String()[:8]
	}
	category := models.NewCategory(name, "Test category")
	category.Enabled = true

	err := f.insertCategory(category)
	if err != nil {
		panic(fmt.Sprintf("Failed to create category: %v", err))
	}

	return category
}

// CreateDisabledCategory 创建禁用的测试 Category
func (f *TestFixture) CreateDisabledCategory(name string) *models.Category {
	category := models.NewCategory(name, "Test disabled category")
	category.Disable()

	err := f.insertCategory(category)
	if err != nil {
		panic(fmt.Sprintf("Failed to create category: %v", err))
	}

	return category
}

// insertCategory 插入 Category 到数据库
func (f *TestFixture) insertCategory(category *models.Category) error {
	query := `
		INSERT INTO categories (uuid, name, description, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := f.DB.Exec(
		query,
		category.UUID, category.Name, category.Description, category.Enabled,
		category.CreatedAt, category.UpdatedAt,
	)
	return err
}

// CreateQuotaPolicy 创建测试 QuotaPolicy
func (f *TestFixture) CreateQuotaPolicy(categoryUUID string, minInstances, maxInstances, priority, maxLifetimeSeconds int) *models.QuotaPolicy {
	policy := models.NewQuotaPolicy(categoryUUID, minInstances, maxInstances, priority, maxLifetimeSeconds)

	err := f.insertQuotaPolicy(policy)
	if err != nil {
		panic(fmt.Sprintf("Failed to create quota policy: %v", err))
	}

	return policy
}

// insertQuotaPolicy 插入 QuotaPolicy 到数据库
func (f *TestFixture) insertQuotaPolicy(policy *models.QuotaPolicy) error {
	query := `
		INSERT INTO quota_policies (
			uuid, category_uuid, min_instances, max_instances, priority,
			auto_replenish, replenish_threshold, max_lifetime_seconds, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := f.DB.Exec(
		query,
		policy.UUID, policy.CategoryUUID, policy.MinInstances, policy.MaxInstances,
		policy.Priority, policy.AutoReplenish, policy.ReplenishThreshold,
		policy.MaxLifetimeSeconds, policy.CreatedAt, policy.UpdatedAt,
	)
	return err
}

// CreateResourceInstance 创建测试 ResourceInstance
func (f *TestFixture) CreateResourceInstance(instanceType models.InstanceType, ipAddress string, port int, passwd, createdBy string) *models.ResourceInstance {
	var instance *models.ResourceInstance

	snapshotID := "test-snapshot-" + uuid.New().String()[:8]
	description := "Test resource instance"

	if instanceType == models.InstanceTypeVirtualMachine {
		instance = models.NewVirtualMachine(ipAddress, port, passwd, snapshotID, createdBy)
	} else {
		instance = models.NewMachine(ipAddress, port, passwd, createdBy, true)
	}

	if instance.Description != nil {
		*instance.Description = description
	}

	err := f.insertResourceInstance(instance)
	if err != nil {
		panic(fmt.Sprintf("Failed to create resource instance: %v", err))
	}

	return instance
}

// insertResourceInstance 插入 ResourceInstance 到数据库
func (f *TestFixture) insertResourceInstance(instance *models.ResourceInstance) error {
	query := `
		INSERT INTO resource_instances (
			uuid, instance_type, snapshot_id, ip_address, port, passwd,
			description, is_public, created_by, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var description sql.NullString
	if instance.Description != nil {
		description = sql.NullString{String: *instance.Description, Valid: true}
	}

	var snapshotID sql.NullString
	if instance.SnapshotID != nil {
		snapshotID = sql.NullString{String: *instance.SnapshotID, Valid: true}
	}

	_, err := f.DB.Exec(
		query,
		instance.UUID, instance.InstanceType, snapshotID, instance.IPAddress,
		instance.Port, instance.Passwd, description, instance.IsPublic,
		instance.CreatedBy, instance.Status, instance.CreatedAt, instance.UpdatedAt,
	)
	return err
}

// CreateTestbed 创建测试 Testbed
func (f *TestFixture) CreateTestbed(categoryUUID, resourceInstanceUUID string, status models.TestbedStatus) *models.Testbed {
	name := "testbed-" + uuid.New().String()[:8]
	serviceTarget := models.ServiceTargetNormal
	mariaDBPort := 3306
	mariaDBUser := "root"
	mariaDBPasswd := "test-passwd"

	testbed := models.NewTestbed(name, categoryUUID, serviceTarget, resourceInstanceUUID, mariaDBPort, mariaDBUser, mariaDBPasswd)
	testbed.Status = status
	testbed.LastHealthCheck = time.Now()

	err := f.insertTestbed(testbed)
	if err != nil {
		panic(fmt.Sprintf("Failed to create testbed: %v", err))
	}

	return testbed
}

// insertTestbed 插入 Testbed 到数据库
func (f *TestFixture) insertTestbed(testbed *models.Testbed) error {
	query := `
		INSERT INTO testbeds (
			uuid, name, category_uuid, service_target, resource_instance_uuid,
			current_alloc_uuid, mariadb_port, mariadb_user, mariadb_passwd,
			status, last_health_check, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var currentAllocUUID sql.NullString
	if testbed.CurrentAllocUUID != nil {
		currentAllocUUID = sql.NullString{String: *testbed.CurrentAllocUUID, Valid: true}
	}

	_, err := f.DB.Exec(
		query,
		testbed.UUID, testbed.Name, testbed.CategoryUUID, testbed.ServiceTarget, testbed.ResourceInstanceUUID,
		currentAllocUUID, testbed.MariaDBPort, testbed.MariaDBUser, testbed.MariaDBPasswd,
		testbed.Status, testbed.LastHealthCheck, testbed.CreatedAt, testbed.UpdatedAt,
	)
	return err
}

// CreateAllocation 创建测试 Allocation
func (f *TestFixture) CreateAllocation(testbedUUID, categoryUUID, requester string, maxLifetimeSeconds int) *models.Allocation {
	allocation := models.NewAllocation(testbedUUID, categoryUUID, requester, maxLifetimeSeconds)

	err := f.insertAllocation(allocation)
	if err != nil {
		panic(fmt.Sprintf("Failed to create allocation: %v", err))
	}

	return allocation
}

// insertAllocation 插入 Allocation 到数据库
func (f *TestFixture) insertAllocation(allocation *models.Allocation) error {
	query := `
		INSERT INTO allocations (
			uuid, testbed_uuid, category_uuid, requester, requester_comment,
			status, expires_at, released_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var requesterComment sql.NullString
	if allocation.RequesterComment != nil {
		requesterComment = sql.NullString{String: *allocation.RequesterComment, Valid: true}
	}

	var expiresAt sql.NullTime
	if allocation.ExpiresAt != nil {
		expiresAt = sql.NullTime{Time: *allocation.ExpiresAt, Valid: true}
	}

	_, err := f.DB.Exec(
		query,
		allocation.UUID, allocation.TestbedUUID, allocation.CategoryUUID,
		allocation.Requester, requesterComment, allocation.Status,
		expiresAt, allocation.ReleasedAt, allocation.CreatedAt, allocation.UpdatedAt,
	)
	return err
}

// GetTestbedByUUID 根据 UUID 获取 Testbed
func (f *TestFixture) GetTestbedByUUID(uuid string) *models.Testbed {
	query := `
		SELECT id, uuid, name, category_uuid, resource_instance_uuid,
			current_alloc_uuid, mariadb_port, mariadb_user, mariadb_passwd,
			status, last_health_check, created_at, updated_at
		FROM testbeds WHERE uuid = ?
	`
	testbed := &models.Testbed{}
	var currentAllocUUID sql.NullString
	var lastHealthCheck sql.NullTime

	err := f.DB.QueryRow(query, uuid).Scan(
		&testbed.ID, &testbed.UUID, &testbed.Name, &testbed.CategoryUUID,
		&testbed.ResourceInstanceUUID, &currentAllocUUID, &testbed.MariaDBPort,
		&testbed.MariaDBUser, &testbed.MariaDBPasswd, &testbed.Status,
		&lastHealthCheck, &testbed.CreatedAt, &testbed.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(fmt.Sprintf("Failed to get testbed: %v", err))
	}

	if currentAllocUUID.Valid {
		testbed.CurrentAllocUUID = &currentAllocUUID.String
	}

	if lastHealthCheck.Valid {
		testbed.LastHealthCheck = lastHealthCheck.Time
	} else {
		testbed.LastHealthCheck = testbed.CreatedAt
	}

	return testbed
}

// GetAllocationByUUID 根据 UUID 获取 Allocation
func (f *TestFixture) GetAllocationByUUID(uuid string) *models.Allocation {
	query := `
		SELECT id, uuid, testbed_uuid, category_uuid, requester, requester_comment,
			status, expires_at, released_at, created_at, updated_at
		FROM allocations WHERE uuid = ?
	`
	allocation := &models.Allocation{}
	var requesterComment sql.NullString
	var expiresAt sql.NullTime
	var releasedAt sql.NullTime

	err := f.DB.QueryRow(query, uuid).Scan(
		&allocation.ID, &allocation.UUID, &allocation.TestbedUUID,
		&allocation.CategoryUUID, &allocation.Requester, &requesterComment,
		&allocation.Status, &expiresAt, &releasedAt,
		&allocation.CreatedAt, &allocation.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(fmt.Sprintf("Failed to get allocation: %v", err))
	}

	if requesterComment.Valid {
		allocation.RequesterComment = &requesterComment.String
	}

	if expiresAt.Valid {
		allocation.ExpiresAt = &expiresAt.Time
	}

	if releasedAt.Valid {
		allocation.ReleasedAt = &releasedAt.Time
	}

	return allocation
}

// CountAvailableTestbeds 统计可用 Testbed 数量
func (f *TestFixture) CountAvailableTestbeds(categoryUUID string) int {
	var count int
	query := `SELECT COUNT(*) FROM testbeds WHERE category_uuid = ? AND status = 'available'`
	err := f.DB.QueryRow(query, categoryUUID).Scan(&count)
	if err != nil {
		panic(fmt.Sprintf("Failed to count testbeds: %v", err))
	}
	return count
}

// CountActiveAllocations 统计活跃 Allocation 数量
func (f *TestFixture) CountActiveAllocations(categoryUUID string) int {
	var count int
	query := `SELECT COUNT(*) FROM allocations WHERE category_uuid = ? AND status = 'active'`
	err := f.DB.QueryRow(query, categoryUUID).Scan(&count)
	if err != nil {
		panic(fmt.Sprintf("Failed to count allocations: %v", err))
	}
	return count
}
