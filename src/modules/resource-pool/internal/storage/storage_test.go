package storage

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hugoh/go-designs/resource-pool/internal/models"
)

// getTestDB 获取测试数据库连接
// 使用环境变量 RESOURCE_POOL_TEST_DSN，如果没有则使用默认值
func getTestDB(t *testing.T) *sql.DB {
	dsn := "root:root123456@tcp(localhost:3307)/event_processor?parseTime=true&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Skipf("Cannot connect to test database: %v", err)
		return nil
	}

	// Test connection
	if err := db.Ping(); err != nil {
		t.Skipf("Cannot ping test database: %v", err)
		db.Close()
		return nil
	}

	// Set shorter timeout for tests
	db.SetMaxIdleConns(2)
	db.SetMaxOpenConns(5)

	return db
}

// cleanupTestTables 清理测试数据
func cleanupTestTables(t *testing.T, db *sql.DB) {
	tables := []string{
		"allocations",
		"testbeds",
		"resource_instances",
		"quota_policies",
		"categories",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE uuid LIKE 'test-%%'", table))
		if err != nil {
			t.Logf("Warning: failed to cleanup %s: %v", table, err)
		}
	}
}

// TestMySQLTestbedStorage_CreateAndGet 测试创建和获取 Testbed
func TestMySQLTestbedStorage_CreateAndGet(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()
	cleanupTestTables(t, db)

	storage := NewMySQLTestbedStorage(db)

	// Create a category and resource instance first
	categoryUUID := fmt.Sprintf("test-category-%s", time.Now().Format("20060102150405"))
	instanceUUID := fmt.Sprintf("test-instance-%s", time.Now().Format("20060102150405"))

	// Insert category
	_, err := db.Exec(`INSERT INTO categories (uuid, name, description, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())`,
		categoryUUID, "Test Category", "Test", true)
	if err != nil {
		t.Fatalf("Failed to create test category: %v", err)
	}

	// Insert resource instance
	snapshotID := "test-snapshot"
	_, err = db.Exec(`INSERT INTO resource_instances (uuid, instance_type, snapshot_id, ip_address, port, ssh_user, passwd, is_public, created_by, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())`,
		instanceUUID, "VirtualMachine", snapshotID, "192.168.1.100", 22, "root", "password", true, "admin", "active")
	if err != nil {
		t.Fatalf("Failed to create test resource instance: %v", err)
	}

	// Create testbed
	testbed := &models.Testbed{
		UUID:                 fmt.Sprintf("test-testbed-%s", time.Now().Format("20060102150405")),
		Name:                 "test-testbed",
		CategoryUUID:         categoryUUID,
		ServiceTarget:        models.ServiceTargetNormal,
		ResourceInstanceUUID: instanceUUID,
		CurrentAllocUUID:     nil,
		MariaDBPort:          3306,
		MariaDBUser:          "root",
		MariaDBPasswd:        "testpass",
		Status:               models.TestbedStatusAvailable,
		LastHealthCheck:      time.Now(),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err = storage.CreateTestbed(testbed)
	if err != nil {
		t.Fatalf("Failed to create testbed: %v", err)
	}

	if testbed.ID == 0 {
		t.Error("ID should be set after creation")
	}

	// Get by ID
	retrieved, err := storage.GetTestbed(testbed.ID)
	if err != nil {
		t.Fatalf("Failed to get testbed by ID: %v", err)
	}

	if retrieved.UUID != testbed.UUID {
		t.Errorf("UUID mismatch: got %s, want %s", retrieved.UUID, testbed.UUID)
	}

	if retrieved.Name != testbed.Name {
		t.Errorf("Name mismatch: got %s, want %s", retrieved.Name, testbed.Name)
	}

	if retrieved.Status != testbed.Status {
		t.Errorf("Status mismatch: got %s, want %s", retrieved.Status, testbed.Status)
	}

	// Get by UUID
	retrievedByUUID, err := storage.GetTestbedByUUID(testbed.UUID)
	if err != nil {
		t.Fatalf("Failed to get testbed by UUID: %v", err)
	}

	if retrievedByUUID.ID != testbed.ID {
		t.Errorf("ID mismatch when getting by UUID")
	}
}

// TestMySQLTestbedStorage_List 测试列出 Testbed
func TestMySQLTestbedStorage_List(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()
	cleanupTestTables(t, db)

	storage := NewMySQLTestbedStorage(db)

	// Create test data
	categoryUUID := fmt.Sprintf("test-category-%s", time.Now().Format("20060102150405"))
	instanceUUID := fmt.Sprintf("test-instance-%s", time.Now().Format("20060102150405"))

	_, err := db.Exec(`INSERT INTO categories (uuid, name, description, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())`,
		categoryUUID, "Test Category", "Test", true)
	if err != nil {
		t.Fatalf("Failed to create test category: %v", err)
	}

	_, err = db.Exec(`INSERT INTO resource_instances (uuid, instance_type, snapshot_id, ip_address, port, ssh_user, passwd, is_public, created_by, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())`,
		instanceUUID, "VirtualMachine", "snap", "192.168.1.100", 22, "root", "password", true, "admin", "active")
	if err != nil {
		t.Fatalf("Failed to create test resource instance: %v", err)
	}

	// Create 3 testbeds with different statuses
	statuses := []models.TestbedStatus{
		models.TestbedStatusAvailable,
		models.TestbedStatusAllocated,
		models.TestbedStatusInUse,
	}

	for i, status := range statuses {
		testbed := &models.Testbed{
			UUID:                 fmt.Sprintf("test-testbed-%d-%s", i, time.Now().Format("20060102150405")),
			Name:                 fmt.Sprintf("test-bed-%d", i),
			CategoryUUID:         categoryUUID,
			ServiceTarget:        models.ServiceTargetNormal,
			ResourceInstanceUUID: instanceUUID,
			MariaDBPort:          3306,
			MariaDBUser:          "root",
			MariaDBPasswd:        "testpass",
			Status:               status,
			LastHealthCheck:      time.Now(),
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		}
		err := storage.CreateTestbed(testbed)
		if err != nil {
			t.Fatalf("Failed to create testbed %d: %v", i, err)
		}
	}

	// List all
	all, err := storage.ListTestbeds()
	if err != nil {
		t.Fatalf("Failed to list testbeds: %v", err)
	}

	if len(all) < 3 {
		t.Errorf("Expected at least 3 testbeds, got %d", len(all))
	}

	// List by category
	byCategory, err := storage.ListTestbedsByCategory(categoryUUID)
	if err != nil {
		t.Fatalf("Failed to list testbeds by category: %v", err)
	}

	if len(byCategory) < 3 {
		t.Errorf("Expected at least 3 testbeds in category, got %d", len(byCategory))
	}

	// List by status
	available, err := storage.ListTestbedsByStatus(models.TestbedStatusAvailable)
	if err != nil {
		t.Fatalf("Failed to list testbeds by status: %v", err)
	}

	if len(available) < 1 {
		t.Errorf("Expected at least 1 available testbed, got %d", len(available))
	}

	// List available
	availableInCategory, err := storage.ListAvailableTestbeds(categoryUUID)
	if err != nil {
		t.Fatalf("Failed to list available testbeds: %v", err)
	}

	if len(availableInCategory) < 1 {
		t.Errorf("Expected at least 1 available testbed in category, got %d", len(availableInCategory))
	}
}

// TestMySQLTestbedStorage_Update 测试更新 Testbed
func TestMySQLTestbedStorage_Update(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()
	cleanupTestTables(t, db)

	storage := NewMySQLTestbedStorage(db)

	// Create test data
	categoryUUID := fmt.Sprintf("test-category-%s", time.Now().Format("20060102150405"))
	instanceUUID := fmt.Sprintf("test-instance-%s", time.Now().Format("20060102150405"))

	_, err := db.Exec(`INSERT INTO categories (uuid, name, description, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())`,
		categoryUUID, "Test Category", "Test", true)
	if err != nil {
		t.Fatalf("Failed to create test category: %v", err)
	}

	_, err = db.Exec(`INSERT INTO resource_instances (uuid, instance_type, snapshot_id, ip_address, port, ssh_user, passwd, is_public, created_by, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())`,
		instanceUUID, "VirtualMachine", "snap", "192.168.1.100", 22, "root", "password", true, "admin", "active")
	if err != nil {
		t.Fatalf("Failed to create test resource instance: %v", err)
	}

	testbed := &models.Testbed{
		UUID:                 fmt.Sprintf("test-testbed-%s", time.Now().Format("20060102150405")),
		Name:                 "original-name",
		CategoryUUID:         categoryUUID,
		ServiceTarget:        models.ServiceTargetNormal,
		ResourceInstanceUUID: instanceUUID,
		MariaDBPort:          3306,
		MariaDBUser:          "root",
		MariaDBPasswd:        "testpass",
		Status:               models.TestbedStatusAvailable,
		LastHealthCheck:      time.Now(),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err = storage.CreateTestbed(testbed)
	if err != nil {
		t.Fatalf("Failed to create testbed: %v", err)
	}

	// Update
	testbed.Name = "updated-name"
	testbed.Status = models.TestbedStatusDeleted
	testbed.UpdatedAt = time.Now()

	err = storage.UpdateTestbed(testbed)
	if err != nil {
		t.Fatalf("Failed to update testbed: %v", err)
	}

	// Verify
	retrieved, err := storage.GetTestbed(testbed.ID)
	if err != nil {
		t.Fatalf("Failed to get updated testbed: %v", err)
	}

	if retrieved.Name != "updated-name" {
		t.Errorf("Name not updated: got %s", retrieved.Name)
	}

	if retrieved.Status != models.TestbedStatusDeleted {
		t.Errorf("Status not updated: got %s", retrieved.Status)
	}
}

// TestMySQLTestbedStorage_TryAllocateTestbed 测试原子分配操作
func TestMySQLTestbedStorage_TryAllocateTestbed(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()
	cleanupTestTables(t, db)

	storage := NewMySQLTestbedStorage(db)

	// Create test data
	categoryUUID := fmt.Sprintf("test-category-%s", time.Now().Format("20060102150405"))
	instanceUUID := fmt.Sprintf("test-instance-%s", time.Now().Format("20060102150405"))

	_, err := db.Exec(`INSERT INTO categories (uuid, name, description, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())`,
		categoryUUID, "Test Category", "Test", true)
	if err != nil {
		t.Fatalf("Failed to create test category: %v", err)
	}

	_, err = db.Exec(`INSERT INTO resource_instances (uuid, instance_type, snapshot_id, ip_address, port, ssh_user, passwd, is_public, created_by, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())`,
		instanceUUID, "VirtualMachine", "snap", "192.168.1.100", 22, "root", "password", true, "admin", "active")
	if err != nil {
		t.Fatalf("Failed to create test resource instance: %v", err)
	}

	testbed := &models.Testbed{
		UUID:                 fmt.Sprintf("test-testbed-%s", time.Now().Format("20060102150405")),
		Name:                 "test-bed",
		CategoryUUID:         categoryUUID,
		ServiceTarget:        models.ServiceTargetNormal,
		ResourceInstanceUUID: instanceUUID,
		MariaDBPort:          3306,
		MariaDBUser:          "root",
		MariaDBPasswd:        "testpass",
		Status:               models.TestbedStatusAvailable,
		LastHealthCheck:      time.Now(),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err = storage.CreateTestbed(testbed)
	if err != nil {
		t.Fatalf("Failed to create testbed: %v", err)
	}

	allocUUID := fmt.Sprintf("test-alloc-%s", time.Now().Format("20060102150405"))

	// First allocation should succeed
	success, err := storage.TryAllocateTestbed(testbed.UUID, allocUUID)
	if err != nil {
		t.Fatalf("TryAllocateTestbed failed: %v", err)
	}
	if !success {
		t.Error("First allocation should succeed")
	}

	// Second allocation should fail (already allocated)
	success, err = storage.TryAllocateTestbed(testbed.UUID, allocUUID+"-2")
	if err != nil {
		t.Fatalf("Second TryAllocateTestbed failed: %v", err)
	}
	if success {
		t.Error("Second allocation should fail (already allocated)")
	}

	// Verify status
	retrieved, err := storage.GetTestbedByUUID(testbed.UUID)
	if err != nil {
		t.Fatalf("Failed to get testbed: %v", err)
	}

	if retrieved.Status != models.TestbedStatusAllocated {
		t.Errorf("Status should be allocated, got %s", retrieved.Status)
	}

	if retrieved.CurrentAllocUUID == nil || *retrieved.CurrentAllocUUID != allocUUID {
		t.Error("CurrentAllocUUID not set correctly")
	}
}

// TestMySQLTestbedStorage_ClearAllocation 测试清除分配
func TestMySQLTestbedStorage_ClearAllocation(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()
	cleanupTestTables(t, db)

	storage := NewMySQLTestbedStorage(db)

	// Create test data
	categoryUUID := fmt.Sprintf("test-category-%s", time.Now().Format("20060102150405"))
	instanceUUID := fmt.Sprintf("test-instance-%s", time.Now().Format("20060102150405"))

	_, err := db.Exec(`INSERT INTO categories (uuid, name, description, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())`,
		categoryUUID, "Test Category", "Test", true)
	if err != nil {
		t.Fatalf("Failed to create test category: %v", err)
	}

	_, err = db.Exec(`INSERT INTO resource_instances (uuid, instance_type, snapshot_id, ip_address, port, ssh_user, passwd, is_public, created_by, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())`,
		instanceUUID, "VirtualMachine", "snap", "192.168.1.100", 22, "root", "password", true, "admin", "active")
	if err != nil {
		t.Fatalf("Failed to create test resource instance: %v", err)
	}

	testbed := &models.Testbed{
		UUID:                 fmt.Sprintf("test-testbed-%s", time.Now().Format("20060102150405")),
		Name:                 "test-bed",
		CategoryUUID:         categoryUUID,
		ServiceTarget:        models.ServiceTargetNormal,
		ResourceInstanceUUID: instanceUUID,
		MariaDBPort:          3306,
		MariaDBUser:          "root",
		MariaDBPasswd:        "testpass",
		Status:               models.TestbedStatusAllocated,
		LastHealthCheck:      time.Now(),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
	allocUUID := "test-alloc"
	testbed.CurrentAllocUUID = &allocUUID

	err = storage.CreateTestbed(testbed)
	if err != nil {
		t.Fatalf("Failed to create testbed: %v", err)
	}

	// Clear allocation
	err = storage.ClearTestbedAllocation(testbed.UUID)
	if err != nil {
		t.Fatalf("Failed to clear allocation: %v", err)
	}

	// Verify
	retrieved, err := storage.GetTestbedByUUID(testbed.UUID)
	if err != nil {
		t.Fatalf("Failed to get testbed: %v", err)
	}

	if retrieved.Status != models.TestbedStatusAvailable {
		t.Errorf("Status should be available after clear, got %s", retrieved.Status)
	}

	if retrieved.CurrentAllocUUID != nil {
		t.Error("CurrentAllocUUID should be nil after clear")
	}
}

// TestMySQLTestbedStorage_Count 测试统计方法
func TestMySQLTestbedStorage_Count(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()
	cleanupTestTables(t, db)

	storage := NewMySQLTestbedStorage(db)

	// Create test data
	categoryUUID := fmt.Sprintf("test-category-%s", time.Now().Format("20060102150405"))
	instanceUUID := fmt.Sprintf("test-instance-%s", time.Now().Format("20060102150405"))

	_, err := db.Exec(`INSERT INTO categories (uuid, name, description, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())`,
		categoryUUID, "Test Category", "Test", true)
	if err != nil {
		t.Fatalf("Failed to create test category: %v", err)
	}

	_, err = db.Exec(`INSERT INTO resource_instances (uuid, instance_type, snapshot_id, ip_address, port, ssh_user, passwd, is_public, created_by, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())`,
		instanceUUID, "VirtualMachine", "snap", "192.168.1.100", 22, "root", "password", true, "admin", "active")
	if err != nil {
		t.Fatalf("Failed to create test resource instance: %v", err)
	}

	// Create 2 available, 1 allocated, 1 in_use
	for i := 0; i < 4; i++ {
		status := models.TestbedStatusAvailable
		if i == 1 {
			status = models.TestbedStatusAllocated
		} else if i == 2 {
			status = models.TestbedStatusInUse
		}

		testbed := &models.Testbed{
			UUID:                 fmt.Sprintf("test-testbed-%d-%s", i, time.Now().Format("20060102150405")),
			Name:                 fmt.Sprintf("test-bed-%d", i),
			CategoryUUID:         categoryUUID,
			ServiceTarget:        models.ServiceTargetNormal,
			ResourceInstanceUUID: instanceUUID,
			MariaDBPort:          3306,
			MariaDBUser:          "root",
			MariaDBPasswd:        "testpass",
			Status:               status,
			LastHealthCheck:      time.Now(),
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		}
		err := storage.CreateTestbed(testbed)
		if err != nil {
			t.Fatalf("Failed to create testbed %d: %v", i, err)
		}
	}

	// Count all
	total, err := storage.CountTestbedsByCategory(categoryUUID)
	if err != nil {
		t.Fatalf("Failed to count testbeds: %v", err)
	}

	if total != 4 {
		t.Errorf("Expected 4 total testbeds, got %d", total)
	}

	// Count available
	available, err := storage.CountAvailableTestbedsByCategory(categoryUUID, models.ServiceTargetNormal)
	if err != nil {
		t.Fatalf("Failed to count available testbeds: %v", err)
	}

	if available != 2 {
		t.Errorf("Expected 2 available testbeds, got %d", available)
	}

	// Count allocated
	allocated, err := storage.CountAllocatedTestbedsByCategory(categoryUUID)
	if err != nil {
		t.Fatalf("Failed to count allocated testbeds: %v", err)
	}

	if allocated != 2 {
		t.Errorf("Expected 2 allocated testbeds, got %d", allocated)
	}
}

// TestMySQLTestbedStorage_Delete 测试删除 Testbed
func TestMySQLTestbedStorage_Delete(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()
	cleanupTestTables(t, db)

	storage := NewMySQLTestbedStorage(db)

	// Create test data
	categoryUUID := fmt.Sprintf("test-category-%s", time.Now().Format("20060102150405"))
	instanceUUID := fmt.Sprintf("test-instance-%s", time.Now().Format("20060102150405"))

	_, err := db.Exec(`INSERT INTO categories (uuid, name, description, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())`,
		categoryUUID, "Test Category", "Test", true)
	if err != nil {
		t.Fatalf("Failed to create test category: %v", err)
	}

	_, err = db.Exec(`INSERT INTO resource_instances (uuid, instance_type, snapshot_id, ip_address, port, ssh_user, passwd, is_public, created_by, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())`,
		instanceUUID, "VirtualMachine", "snap", "192.168.1.100", 22, "root", "password", true, "admin", "active")
	if err != nil {
		t.Fatalf("Failed to create test resource instance: %v", err)
	}

	testbed := &models.Testbed{
		UUID:                 fmt.Sprintf("test-testbed-%s", time.Now().Format("20060102150405")),
		Name:                 "test-bed",
		CategoryUUID:         categoryUUID,
		ServiceTarget:        models.ServiceTargetNormal,
		ResourceInstanceUUID: instanceUUID,
		MariaDBPort:          3306,
		MariaDBUser:          "root",
		MariaDBPasswd:        "testpass",
		Status:               models.TestbedStatusAvailable,
		LastHealthCheck:      time.Now(),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err = storage.CreateTestbed(testbed)
	if err != nil {
		t.Fatalf("Failed to create testbed: %v", err)
	}

	id := testbed.ID

	// Delete
	err = storage.DeleteTestbed(id)
	if err != nil {
		t.Fatalf("Failed to delete testbed: %v", err)
	}

	// Verify deleted
	_, err = storage.GetTestbed(id)
	if err == nil {
		t.Error("Should get error when getting deleted testbed")
	}
}

// TestMySQLResourceInstanceStorage_ListAvailableResourceInstances 测试列出可用资源实例
// 验证修复后的逻辑：排除被活跃 testbed (available/allocated/in_use) 关联的 resource_instance
// 允许被已删除 (deleted) testbed 关联的 resource_instance 再次使用
func TestMySQLResourceInstanceStorage_ListAvailableResourceInstances(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()
	cleanupTestTables(t, db)

	resourceStorage := NewMySQLResourceInstanceStorage(db)

	// 创建 3 个 ResourceInstance
	instance1UUID := fmt.Sprintf("test-instance-1-%s", time.Now().Format("20060102150405"))
	instance2UUID := fmt.Sprintf("test-instance-2-%s", time.Now().Format("20060102150405"))
	instance3UUID := fmt.Sprintf("test-instance-3-%s", time.Now().Format("20060102150405"))
	categoryUUID := fmt.Sprintf("test-category-%s", time.Now().Format("20060102150405"))

	// 插入 category
	_, err := db.Exec(`INSERT INTO categories (uuid, name, description, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())`,
		categoryUUID, "Test Category", "Test", true)
	if err != nil {
		t.Fatalf("Failed to create test category: %v", err)
	}

	snapshotID := "test-snapshot"

	// 插入 instance1 (将被 available testbed 关联)
	_, err = db.Exec(`INSERT INTO resource_instances (uuid, instance_type, snapshot_id, ip_address, port, ssh_user, passwd, is_public, created_by, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())`,
		instance1UUID, "VirtualMachine", snapshotID, "192.168.1.10", 22, "root", "pass1", true, "admin", "active")
	if err != nil {
		t.Fatalf("Failed to create instance1: %v", err)
	}

	// 插入 instance2 (将被 deleted testbed 关联)
	_, err = db.Exec(`INSERT INTO resource_instances (uuid, instance_type, snapshot_id, ip_address, port, ssh_user, passwd, is_public, created_by, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())`,
		instance2UUID, "VirtualMachine", snapshotID, "192.168.1.11", 22, "root", "pass2", true, "admin", "active")
	if err != nil {
		t.Fatalf("Failed to create instance2: %v", err)
	}

	// 插入 instance3 (无关联 testbed)
	_, err = db.Exec(`INSERT INTO resource_instances (uuid, instance_type, snapshot_id, ip_address, port, ssh_user, passwd, is_public, created_by, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())`,
		instance3UUID, "VirtualMachine", snapshotID, "192.168.1.12", 22, "root", "pass3", true, "admin", "active")
	if err != nil {
		t.Fatalf("Failed to create instance3: %v", err)
	}

	// 创建 testbed1 (available 状态) 关联 instance1 → instance1 不可用
	testbed1UUID := fmt.Sprintf("test-testbed1-%s", time.Now().Format("20060102150405"))
	_, err = db.Exec(`INSERT INTO testbeds (uuid, name, category_uuid, service_target, resource_instance_uuid, mariadb_port, mariadb_user, mariadb_passwd, status, last_health_check, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), NOW())`,
		testbed1UUID, "testbed1", categoryUUID, "robot", instance1UUID, 3306, "root", "pass", "available")
	if err != nil {
		t.Fatalf("Failed to create testbed1: %v", err)
	}

	// 创建 testbed2 (deleted 状态) 关联 instance2 → instance2 可用（可复用）
	testbed2UUID := fmt.Sprintf("test-testbed2-%s", time.Now().Format("20060102150405"))
	_, err = db.Exec(`INSERT INTO testbeds (uuid, name, category_uuid, service_target, resource_instance_uuid, mariadb_port, mariadb_user, mariadb_passwd, status, last_health_check, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), NOW())`,
		testbed2UUID, "testbed2", categoryUUID, "robot", instance2UUID, 3306, "root", "pass", "deleted")
	if err != nil {
		t.Fatalf("Failed to create testbed2: %v", err)
	}

	// 调用 ListAvailableResourceInstances
	available, err := resourceStorage.ListAvailableResourceInstances()
	if err != nil {
		t.Fatalf("Failed to list available instances: %v", err)
	}

	// 验证结果：应该是 instance2 和 instance3
	if len(available) != 2 {
		t.Errorf("Expected 2 available resource instances, got %d", len(available))
	}

	// 验证 instance1 不在结果中（被 available testbed 关联）
	for _, inst := range available {
		if inst.UUID == instance1UUID {
			t.Error("Instance1 should not be available (associated with available testbed)")
		}
	}

	// 验证 instance2 在结果中（关联的是 deleted testbed）
	foundInstance2 := false
	for _, inst := range available {
		if inst.UUID == instance2UUID {
			foundInstance2 = true
			break
		}
	}
	if !foundInstance2 {
		t.Error("Instance2 should be available (associated with deleted testbed)")
	}

	// 验证 instance3 在结果中（无关联 testbed）
	foundInstance3 := false
	for _, inst := range available {
		if inst.UUID == instance3UUID {
			foundInstance3 = true
			break
		}
	}
	if !foundInstance3 {
		t.Error("Instance3 should be available (no associated testbed)")
	}
}

// TestMySQLResourceInstanceStorage_ListAvailableWithInUse 测试关联 in_use 状态的 testbed
func TestMySQLResourceInstanceStorage_ListAvailableWithInUse(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()
	cleanupTestTables(t, db)

	resourceStorage := NewMySQLResourceInstanceStorage(db)

	instanceUUID := fmt.Sprintf("test-instance-inuse-%s", time.Now().Format("20060102150405"))
	categoryUUID := fmt.Sprintf("test-category-inuse-%s", time.Now().Format("20060102150405"))

	// 插入 category
	_, err := db.Exec(`INSERT INTO categories (uuid, name, description, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())`,
		categoryUUID, "Test Category", "Test", true)
	if err != nil {
		t.Fatalf("Failed to create test category: %v", err)
	}

	snapshotID := "test-snapshot"

	// 插入 instance
	_, err = db.Exec(`INSERT INTO resource_instances (uuid, instance_type, snapshot_id, ip_address, port, ssh_user, passwd, is_public, created_by, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())`,
		instanceUUID, "VirtualMachine", snapshotID, "192.168.1.20", 22, "root", "pass", true, "admin", "active")
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}

	// 创建 in_use 状态的 testbed 关联 instance → 不可用
	testbedUUID := fmt.Sprintf("test-testbed-inuse-%s", time.Now().Format("20060102150405"))
	_, err = db.Exec(`INSERT INTO testbeds (uuid, name, category_uuid, service_target, resource_instance_uuid, mariadb_port, mariadb_user, mariadb_passwd, status, last_health_check, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), NOW())`,
		testbedUUID, "testbed-inuse", categoryUUID, "robot", instanceUUID, 3306, "root", "pass", "in_use")
	if err != nil {
		t.Fatalf("Failed to create testbed: %v", err)
	}

	// 调用 ListAvailableResourceInstances
	available, err := resourceStorage.ListAvailableResourceInstances()
	if err != nil {
		t.Fatalf("Failed to list available instances: %v", err)
	}

	// 验证 instance 不在结果中
	for _, inst := range available {
		if inst.UUID == instanceUUID {
			t.Errorf("Instance with in_use testbed should not be available, but found %s", instanceUUID)
		}
	}
}

// TestMySQLResourceInstanceStorage_ListAvailableWithReleasing 测试关联 releasing 状态的 testbed
func TestMySQLResourceInstanceStorage_ListAvailableWithReleasing(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()
	cleanupTestTables(t, db)

	resourceStorage := NewMySQLResourceInstanceStorage(db)

	timestamp := time.Now().Format("20060102150405")
	instanceUUID := fmt.Sprintf("test-inst-rel-%s", timestamp)
	categoryUUID := fmt.Sprintf("test-cat-rel-%s", timestamp)

	// 插入 category
	_, err := db.Exec(`INSERT INTO categories (uuid, name, description, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())`,
		categoryUUID, "Test Category", "Test", true)
	if err != nil {
		t.Fatalf("Failed to create test category: %v", err)
	}

	snapshotID := "test-snapshot"

	// 插入 instance
	_, err = db.Exec(`INSERT INTO resource_instances (uuid, instance_type, snapshot_id, ip_address, port, ssh_user, passwd, is_public, created_by, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())`,
		instanceUUID, "VirtualMachine", snapshotID, "192.168.1.21", 22, "root", "pass", true, "admin", "active")
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}

	// 创建 releasing 状态的 testbed 关联 instance → 可用（释放中不是活跃状态）
	testbedUUID := fmt.Sprintf("test-tb-rel-%s", timestamp)
	_, err = db.Exec(`INSERT INTO testbeds (uuid, name, category_uuid, service_target, resource_instance_uuid, mariadb_port, mariadb_user, mariadb_passwd, status, last_health_check, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), NOW())`,
		testbedUUID, "testbed-releasing", categoryUUID, "robot", instanceUUID, 3306, "root", "pass", "releasing")
	if err != nil {
		t.Fatalf("Failed to create testbed: %v", err)
	}

	// 调用 ListAvailableResourceInstances
	available, err := resourceStorage.ListAvailableResourceInstances()
	if err != nil {
		t.Fatalf("Failed to list available instances: %v", err)
	}

	// 验证 instance 在结果中（releasing 状态不是活跃状态）
	found := false
	for _, inst := range available {
		if inst.UUID == instanceUUID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Instance with releasing testbed should be available, but not found %s", instanceUUID)
	}
}
