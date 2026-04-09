package service

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/hugoh/go-designs/resource-pool/internal/deployer"
	"github.com/hugoh/go-designs/resource-pool/internal/models"
	"github.com/hugoh/go-designs/resource-pool/internal/storage"
)

// TestResourcePoolService_BasicOperations tests basic service operations with real database
// These are integration-level tests that verify the service works with storage
func TestResourcePoolService_BasicOperations(t *testing.T) {
	// This test requires database connection
	// Skip if database not available
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()
	cleanupTestData(t, db)

	// Create storage instances
	testbedStorage := storage.NewMySQLTestbedStorage(db)
	allocationStorage := storage.NewMySQLAllocationStorage(db)
	categoryStorage := storage.NewMySQLCategoryStorage(db)
	quotaStorage := storage.NewMySQLQuotaPolicyStorage(db)
	resourceStorage := storage.NewMySQLResourceInstanceStorage(db)
	taskStorage := storage.NewMySQLResourceInstanceTaskStorage(db)
	mockDeployer := deployer.NewMockDeployService()

	service := NewResourcePoolService(
		testbedStorage,
		allocationStorage,
		categoryStorage,
		quotaStorage,
		resourceStorage,
		taskStorage,
		mockDeployer,
		3600,
	)

	t.Run("Create and Get Category", func(t *testing.T) {
		category := models.NewCategory("test-service-category-"+fmt.Sprint(time.Now().UnixNano()), "Test category for service")

		err := service.CreateCategory(category)
		if err != nil {
			t.Fatalf("Failed to create category: %v", err)
		}

		retrieved, err := service.GetCategory(category.UUID)
		if err != nil {
			t.Fatalf("Failed to get category: %v", err)
		}

		if retrieved.Name != category.Name {
			t.Errorf("Name mismatch: got %s, want %s", retrieved.Name, category.Name)
		}
	})

	t.Run("Set and Get Quota Policy", func(t *testing.T) {
		category := models.NewCategory("test-quota-category-"+fmt.Sprint(time.Now().UnixNano()), "Test quota")
		err := service.CreateCategory(category)
		if err != nil {
			t.Fatalf("Failed to create category: %v", err)
		}

		policy := models.NewQuotaPolicy(category.UUID, 1, 5, 1, 7200)

		err = service.SetQuotaPolicy(policy)
		if err != nil {
			t.Fatalf("Failed to set quota policy: %v", err)
		}

		retrieved, err := service.GetQuotaPolicy(category.UUID)
		if err != nil {
			t.Fatalf("Failed to get quota policy: %v", err)
		}

		if retrieved.MaxInstances != 5 {
			t.Errorf("MaxInstances mismatch: got %d, want 5", retrieved.MaxInstances)
		}
	})

	t.Run("List Categories", func(t *testing.T) {
		categories, err := service.ListCategories()
		if err != nil {
			t.Fatalf("Failed to list categories: %v", err)
		}

		if len(categories) < 2 {
			t.Errorf("Expected at least 2 categories, got %d", len(categories))
		}
	})
}

// TestResourcePoolService_AcquireRelease tests the acquire and release cycle
func TestResourcePoolService_AcquireRelease(t *testing.T) {
	db := getTestDBForService(t)
	if db == nil {
		return
	}
	defer db.Close()
	cleanupTestData(t, db)

	testbedStorage := storage.NewMySQLTestbedStorage(db)
	allocationStorage := storage.NewMySQLAllocationStorage(db)
	categoryStorage := storage.NewMySQLCategoryStorage(db)
	quotaStorage := storage.NewMySQLQuotaPolicyStorage(db)
	resourceStorage := storage.NewMySQLResourceInstanceStorage(db)
	taskStorage := storage.NewMySQLResourceInstanceTaskStorage(db)
	mockDeployer := deployer.NewMockDeployService()

	// Speed up mock operations
	mockDeployer.SetDelays(100*time.Millisecond, 100*time.Millisecond)

	service := NewResourcePoolService(
		testbedStorage,
		allocationStorage,
		categoryStorage,
		quotaStorage,
		resourceStorage,
		taskStorage,
		mockDeployer,
		3600,
	)

	// Create test data
	category := models.NewCategory("test-acquire-category-"+fmt.Sprint(time.Now().UnixNano()), "Test acquire/release")
	err := service.CreateCategory(category)
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}

	// Create quota policy for robot users (matching the testbed's service_target)
	now := time.Now()
	robotPolicy := &models.QuotaPolicy{
		UUID:               uuid.New().String(),
		CategoryUUID:       category.UUID,
		MinInstances:       1,
		MaxInstances:       5,
		Priority:           1,
		ServiceTarget:      models.ServiceTargetRobot,
		AutoReplenish:      true,
		ReplenishThreshold: 1,
		MaxLifetimeSeconds: 3600,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	err = service.SetQuotaPolicy(robotPolicy)
	if err != nil {
		t.Fatalf("Failed to set quota policy: %v", err)
	}

	// Create resource instance and testbed
	snapshotID := "test-snapshot-acquire"
	instance := models.NewVirtualMachine("192.168.1.100", 22, "password", snapshotID, "test-service")
	instance.SnapshotInstanceUUID = &snapshotID
	err = resourceStorage.CreateResourceInstance(instance)
	if err != nil {
		t.Fatalf("Failed to create resource instance: %v", err)
	}

	testbed := models.NewTestbed("test-service-bed-"+fmt.Sprint(time.Now().UnixNano()), category.UUID, models.ServiceTargetRobot, instance.UUID, 3306, "root", "testpass")
	err = testbedStorage.CreateTestbed(testbed)
	if err != nil {
		t.Fatalf("Failed to create testbed: %v", err)
	}

	t.Run("Acquire Testbed", func(t *testing.T) {
		// Use robot user for testing since it has highest priority
		allocation, retrievedTestbed, err := service.AcquireTestbed(context.Background(), category.UUID, RobotUsername)
		if err != nil {
			t.Fatalf("Failed to acquire testbed: %v", err)
		}

		if allocation == nil {
			t.Fatal("Expected allocation, got nil")
		}
		if retrievedTestbed == nil {
			t.Fatal("Expected testbed, got nil")
		}

		if retrievedTestbed.Status != models.TestbedStatusAllocated {
			t.Errorf("Expected status allocated, got %s", retrievedTestbed.Status)
		}
		if retrievedTestbed.CurrentAllocUUID == nil || *retrievedTestbed.CurrentAllocUUID != allocation.UUID {
			t.Error("CurrentAllocUUID not set correctly")
		}
		if !allocation.IsActive() {
			t.Error("Expected allocation to be active")
		}
	})

	t.Run("Release Testbed", func(t *testing.T) {
		// Get the allocation
		testbed, _ := testbedStorage.ListTestbedsByCategory(category.UUID)
		if len(testbed) == 0 {
			t.Fatal("No testbed found")
		}

		allocations, _ := allocationStorage.ListAllocationsByTestbed(testbed[0].UUID)
		if len(allocations) == 0 {
			t.Fatal("No allocation found")
		}

		err := service.ReleaseTestbed(context.Background(), allocations[0].UUID)
		if err != nil {
			t.Fatalf("Failed to release testbed: %v", err)
		}

		// Wait for async restore to complete
		time.Sleep(500 * time.Millisecond)

		// Check testbed is marked as deleted (one-time use design)
		retrievedTestbed, err := testbedStorage.GetTestbedByUUID(testbed[0].UUID)
		if err != nil {
			t.Fatalf("Failed to get testbed after release: %v", err)
		}

		if retrievedTestbed.Status != models.TestbedStatusDeleted {
			t.Errorf("Expected status deleted after release (one-time use), got %s", retrievedTestbed.Status)
		}

		// Verify allocation is marked as released
		releasedAlloc, err := allocationStorage.GetAllocationByUUID(allocations[0].UUID)
		if err != nil {
			t.Fatalf("Failed to get allocation after release: %v", err)
		}

		if !releasedAlloc.IsReleased() {
			t.Error("Expected allocation to be marked as released")
		}
	})

	t.Run("Extend Allocation", func(t *testing.T) {
		// Create a new testbed since the previous test deleted the original one
		newInstance := models.NewVirtualMachine("192.168.1.101", 22, "password", "test-snapshot-extend", "test-service")
		newInstance.SnapshotInstanceUUID = ptr("test-snapshot-extend")
		err = resourceStorage.CreateResourceInstance(newInstance)
		if err != nil {
			t.Fatalf("Failed to create new resource instance: %v", err)
		}

		newTestbed := models.NewTestbed("test-service-bed-extend-"+fmt.Sprint(time.Now().UnixNano()), category.UUID, models.ServiceTargetRobot, newInstance.UUID, 3306, "root", "testpass")
		err = testbedStorage.CreateTestbed(newTestbed)
		if err != nil {
			t.Fatalf("Failed to create new testbed: %v", err)
		}

		// Acquire with robot user
		allocation, _, err := service.AcquireTestbed(context.Background(), category.UUID, RobotUsername)
		if err != nil {
			t.Fatalf("Failed to acquire testbed: %v", err)
		}

		originalExpiry := allocation.ExpiresAt
		additionalSeconds := 1800 // 30 minutes

		err = service.ExtendAllocation(context.Background(), allocation.UUID, additionalSeconds)
		if err != nil {
			t.Fatalf("Failed to extend allocation: %v", err)
		}

		// Get updated allocation
		updatedAllocation, err := allocationStorage.GetAllocationByUUID(allocation.UUID)
		if err != nil {
			t.Fatalf("Failed to get updated allocation: %v", err)
		}

		if updatedAllocation.ExpiresAt == nil {
			t.Fatal("Expected ExpiresAt to be set")
		}

		expectedExpiry := originalExpiry.Add(time.Duration(additionalSeconds) * time.Second)
		diff := updatedAllocation.ExpiresAt.Sub(expectedExpiry)
		if diff < 0 {
			diff = -diff
		}
		if diff > time.Second {
			t.Errorf("Expiry time not extended correctly: expected ~%v, got %v (diff %v)",
				expectedExpiry, updatedAllocation.ExpiresAt, diff)
		}
	})
}

// getTestDBForService gets test database connection for service tests
func getTestDBForService(t *testing.T) *sql.DB {
	dsn := "root:root123456@tcp(localhost:3307)/event_processor?parseTime=true&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Skipf("Cannot connect to test database: %v", err)
		return nil
	}

	if err := db.Ping(); err != nil {
		t.Skipf("Cannot ping test database: %v", err)
		db.Close()
		return nil
	}

	db.SetMaxIdleConns(2)
	db.SetMaxOpenConns(5)

	return db
}

// cleanupTestData cleans up test data
func cleanupTestData(t *testing.T, db *sql.DB) {
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

func ptr(s string) *string {
	return &s
}
