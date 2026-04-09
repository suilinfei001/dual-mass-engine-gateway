// +build integration

package e2e

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hugoh/go-designs/resource-pool/internal/deployer"
	"github.com/hugoh/go-designs/resource-pool/internal/models"
	"github.com/hugoh/go-designs/resource-pool/internal/service"
	"github.com/hugoh/go-designs/resource-pool/internal/storage"
)

const (
	testDSN = "root:root123456@tcp(localhost:3307)/event_processor?parseTime=true&loc=Local"
)

// TestMain sets up and tears down the test database
func TestMain(m *testing.M) {
	// Check if database is available
	db, err := sql.Open("mysql", testDSN)
	if err != nil {
		fmt.Printf("Cannot connect to test database: %v\n", err)
		fmt.Println("Skipping E2E tests - database not available")
		os.Exit(0)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Printf("Cannot ping test database: %v\n", err)
		fmt.Println("Skipping E2E tests - database not accessible")
		os.Exit(0)
	}

	// Run tests
	code := m.Run()

	// Cleanup test data
	cleanupTestData(db)

	os.Exit(code)
}

// cleanupTestData removes all test data
func cleanupTestData(db *sql.DB) {
	// Delete allocations first (due to foreign key constraints)
	_, _ = db.Exec("DELETE FROM allocations WHERE testbed_uuid IN (SELECT uuid FROM testbeds WHERE name LIKE 'e2e-%')")

	// Delete testbeds by name pattern
	_, _ = db.Exec("DELETE FROM testbeds WHERE name LIKE 'e2e-%'")

	// Delete resource instances created by e2e tests
	_, _ = db.Exec("DELETE FROM resource_instances WHERE created_by = 'e2e-test'")

	// Delete quota policies for e2e categories
	_, _ = db.Exec("DELETE FROM quota_policies WHERE category_uuid IN (SELECT uuid FROM categories WHERE name LIKE 'e2e-%')")

	// Delete e2e categories by name pattern
	_, _ = db.Exec("DELETE FROM categories WHERE name LIKE 'e2e-%'")
}

// setupTestEnvironment creates a complete test environment
func setupTestEnvironment(t *testing.T) (*sql.DB, service.ResourcePoolService, *deployer.MockDeployService) {
	db, err := sql.Open("mysql", testDSN)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	db.SetMaxIdleConns(2)
	db.SetMaxOpenConns(5)

	// Clean up any existing test data
	cleanupTestData(db)

	// Create storage instances
	testbedStorage := storage.NewMySQLTestbedStorage(db)
	allocationStorage := storage.NewMySQLAllocationStorage(db)
	categoryStorage := storage.NewMySQLCategoryStorage(db)
	quotaStorage := storage.NewMySQLQuotaPolicyStorage(db)
	resourceStorage := storage.NewMySQLResourceInstanceStorage(db)

	// Create mock deployer with short delays for testing
	mockDeployer := deployer.NewMockDeployService()
	mockDeployer.SetDelays(50*time.Millisecond, 50*time.Millisecond)

	// Create service
	poolService := service.NewResourcePoolService(
		testbedStorage,
		allocationStorage,
		categoryStorage,
		quotaStorage,
		resourceStorage,
		mockDeployer,
		3600, // 1 hour default lifetime
	)

	return db, poolService, mockDeployer
}

// TestTestbedLifecycle_E2E tests the complete testbed lifecycle
func TestTestbedLifecycle_E2E(t *testing.T) {
	db, poolService, mockDeployer := setupTestEnvironment(t)
	defer db.Close()
	defer cleanupTestData(db)

	var categoryUUID string
	var testbedUUID string

	t.Run("Step1_CreateCategory", func(t *testing.T) {
		category := models.NewCategory("e2e-main-category", "E2E Test Main Category")
		err := poolService.CreateCategory(category)
		if err != nil {
			t.Fatalf("Failed to create category: %v", err)
		}

		// Verify category was created
		retrieved, err := poolService.GetCategory(category.UUID)
		if err != nil {
			t.Fatalf("Failed to retrieve category: %v", err)
		}
		if retrieved.Name != category.Name {
			t.Errorf("Name mismatch: got %s, want %s", retrieved.Name, category.Name)
		}

		// Store for later steps
		categoryUUID = category.UUID
		t.Logf("Created category: %s", categoryUUID)
	})

	t.Run("Step2_SetQuotaPolicy", func(t *testing.T) {
		if categoryUUID == "" {
			t.Fatal("categoryUUID not set from Step1")
		}

		policy := models.NewQuotaPolicy(categoryUUID, 1, 3, 10, 3600)
		policy.AutoReplenish = true
		policy.ReplenishThreshold = 1

		err := poolService.SetQuotaPolicy(policy)
		if err != nil {
			t.Fatalf("Failed to set quota policy: %v", err)
		}

		// Verify policy was set
		retrieved, err := poolService.GetQuotaPolicy(categoryUUID)
		if err != nil {
			t.Fatalf("Failed to retrieve quota policy: %v", err)
		}
		if retrieved.MaxInstances != 3 {
			t.Errorf("MaxInstances mismatch: got %d, want 3", retrieved.MaxInstances)
		}

		t.Logf("Set quota policy for category: %s", categoryUUID)
	})

	t.Run("Step3_CreateResourceInstance", func(t *testing.T) {
		resourceStorage := storage.NewMySQLResourceInstanceStorage(db)

		snapshotID := "e2e-snapshot-main-v1.0"
		instance := models.NewVirtualMachine("192.168.1.100", 22, "password", snapshotID, "e2e-test")

		err := resourceStorage.CreateResourceInstance(instance)
		if err != nil {
			t.Fatalf("Failed to create resource instance: %v", err)
		}

		t.Logf("Created resource instance: %s", instance.UUID)
	})

	t.Run("Step4_CreateTestbed", func(t *testing.T) {
		testbedStorage := storage.NewMySQLTestbedStorage(db)
		resourceStorage := storage.NewMySQLResourceInstanceStorage(db)

		// Get the resource instance
		instances, err := resourceStorage.ListResourceInstances()
		if err != nil || len(instances) == 0 {
			t.Fatal("No resource instances found")
		}
		resourceInstance := instances[0]

		testbed := models.NewTestbed(
			"e2e-testbed-main-001",
			categoryUUID,
			resourceInstance.UUID,
			3306,
			"root",
			"testpass",
		)

		err = testbedStorage.CreateTestbed(testbed)
		if err != nil {
			t.Fatalf("Failed to create testbed: %v", err)
		}

		testbedUUID = testbed.UUID
		t.Logf("Created testbed: %s", testbedUUID)
	})

	t.Run("Step5_AcquireTestbed", func(t *testing.T) {
		ctx := context.Background()
		allocation, testbed, err := poolService.AcquireTestbed(ctx, categoryUUID, "e2e-test-user")
		if err != nil {
			t.Fatalf("Failed to acquire testbed: %v", err)
		}

		if allocation == nil {
			t.Fatal("Expected allocation, got nil")
		}
		if testbed == nil {
			t.Fatal("Expected testbed, got nil")
		}

		if testbed.Status != models.TestbedStatusAllocated {
			t.Errorf("Expected status allocated, got %s", testbed.Status)
		}

		t.Logf("Acquired testbed: %s, allocation: %s", testbed.UUID, allocation.UUID)
	})

	t.Run("Step6_VerifyTestbedStatus", func(t *testing.T) {
		testbed, err := poolService.GetTestbed(testbedUUID)
		if err != nil {
			t.Fatalf("Failed to get testbed: %v", err)
		}

		if testbed.Status != models.TestbedStatusAllocated {
			t.Errorf("Expected status allocated, got %s", testbed.Status)
		}

		if testbed.CurrentAllocUUID == nil {
			t.Error("Expected CurrentAllocUUID to be set")
		}

		t.Logf("Testbed status verified: %s", testbed.Status)
	})

	t.Run("Step7_ExtendAllocation", func(t *testing.T) {
		testbedStorage := storage.NewMySQLTestbedStorage(db)
		allocationStorage := storage.NewMySQLAllocationStorage(db)

		testbed, _ := testbedStorage.GetTestbedByUUID(testbedUUID)
		if testbed.CurrentAllocUUID == nil {
			t.Fatal("No allocation found")
		}

		allocation, err := allocationStorage.GetAllocationByUUID(*testbed.CurrentAllocUUID)
		if err != nil {
			t.Fatalf("Failed to get allocation: %v", err)
		}

		originalExpiry := allocation.ExpiresAt
		additionalSeconds := 1800 // 30 minutes

		ctx := context.Background()
		err = poolService.ExtendAllocation(ctx, allocation.UUID, additionalSeconds)
		if err != nil {
			t.Fatalf("Failed to extend allocation: %v", err)
		}

		// Verify extension
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

		t.Logf("Extended allocation by %d seconds", additionalSeconds)
	})

	t.Run("Step8_ReleaseTestbed", func(t *testing.T) {
		testbedStorage := storage.NewMySQLTestbedStorage(db)
		allocationStorage := storage.NewMySQLAllocationStorage(db)

		testbed, _ := testbedStorage.GetTestbedByUUID(testbedUUID)
		if testbed.CurrentAllocUUID == nil {
			t.Fatal("No allocation found")
		}

		allocationUUID := *testbed.CurrentAllocUUID

		ctx := context.Background()
		err := poolService.ReleaseTestbed(ctx, allocationUUID)
		if err != nil {
			t.Fatalf("Failed to release testbed: %v", err)
		}

		// Verify allocation is released
		allocation, err := allocationStorage.GetAllocationByUUID(allocationUUID)
		if err != nil {
			t.Fatalf("Failed to get allocation after release: %v", err)
		}

		if !allocation.IsReleased() {
			t.Error("Expected allocation to be released")
		}

		t.Logf("Released testbed, allocation status: %s", allocation.Status)
	})

	t.Run("Step9_WaitForSnapshotRestore", func(t *testing.T) {
		// Wait for async restore to complete
		time.Sleep(500 * time.Millisecond)

		testbedStorage := storage.NewMySQLTestbedStorage(db)
		testbed, err := testbedStorage.GetTestbedByUUID(testbedUUID)
		if err != nil {
			t.Fatalf("Failed to get testbed after release: %v", err)
		}

		if testbed.Status != models.TestbedStatusAvailable {
			t.Errorf("Expected status available after release, got %s", testbed.Status)
		}

		if testbed.CurrentAllocUUID != nil {
			t.Error("Expected CurrentAllocUUID to be nil after release")
		}

		// Check deployer stats
		stats := mockDeployer.GetStats()
		restoreCount := stats["restore_count"].(int)
		if restoreCount == 0 {
			t.Error("Expected restore to be called at least once")
		}

		t.Logf("Snapshot restore completed, testbed status: %s, restore count: %d", testbed.Status, restoreCount)
	})

	t.Run("Step10_VerifyQuotaStillEnforced", func(t *testing.T) {
		// Create another resource instance and testbed for quota testing
		resourceStorage := storage.NewMySQLResourceInstanceStorage(db)
		testbedStorage := storage.NewMySQLTestbedStorage(db)

		snapshotID := "e2e-snapshot-main-v1.0"
		instance2 := models.NewVirtualMachine("192.168.1.101", 22, "password", snapshotID, "e2e-test")
		err := resourceStorage.CreateResourceInstance(instance2)
		if err != nil {
			t.Fatalf("Failed to create resource instance 2: %v", err)
		}

		testbed2 := models.NewTestbed(
			"e2e-testbed-main-002",
			categoryUUID,
			instance2.UUID,
			3306,
			"root",
			"testpass",
		)
		err = testbedStorage.CreateTestbed(testbed2)
		if err != nil {
			t.Fatalf("Failed to create testbed 2: %v", err)
		}

		ctx := context.Background()
		_, testbed, err := poolService.AcquireTestbed(ctx, categoryUUID, "e2e-test-user2")
		if err != nil {
			t.Fatalf("Failed to acquire testbed 2: %v", err)
		}

		// Verify we can still get allocation (quota is 3, we've used 1)
		if testbed == nil {
			t.Error("Expected testbed, got nil")
		}

		t.Logf("Quota still enforced after release, acquired: %s", testbed.UUID)
	})
}

// TestConcurrentAllocation_E2E tests concurrent allocation scenarios
func TestConcurrentAllocation_E2E(t *testing.T) {
	db, poolService, _ := setupTestEnvironment(t)
	defer db.Close()
	defer cleanupTestData(db)

	// Setup: Create category, policy, and multiple testbeds
	category := models.NewCategory("e2e-concurrent-category", "E2E Concurrent Test Category")
	err := poolService.CreateCategory(category)
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}

	policy := models.NewQuotaPolicy(category.UUID, 0, 3, 10, 3600)
	err = poolService.SetQuotaPolicy(policy)
	if err != nil {
		t.Fatalf("Failed to set quota policy: %v", err)
	}

	testbedStorage := storage.NewMySQLTestbedStorage(db)
	resourceStorage := storage.NewMySQLResourceInstanceStorage(db)

	// Create 3 resource instances and testbeds
	for i := 0; i < 3; i++ {
		snapshotID := "e2e-snapshot-concurrent-v1.0"
		instance := models.NewVirtualMachine(
			fmt.Sprintf("192.168.1.%d", 100+i),
			22,
			"password",
			snapshotID,
			"e2e-test",
		)
		err := resourceStorage.CreateResourceInstance(instance)
		if err != nil {
			t.Fatalf("Failed to create resource instance %d: %v", i, err)
		}

		testbed := models.NewTestbed(
			fmt.Sprintf("e2e-testbed-concurrent-%03d", i+1),
			category.UUID,
			instance.UUID,
			3306,
			"root",
			"testpass",
		)
		err = testbedStorage.CreateTestbed(testbed)
		if err != nil {
			t.Fatalf("Failed to create testbed %d: %v", i, err)
		}
	}

	t.Run("ConcurrentAcquire", func(t *testing.T) {
		ctx := context.Background()
		concurrency := 3
		results := make(chan *acquireResult, concurrency)

		// Launch concurrent acquisitions
		for i := 0; i < concurrency; i++ {
			go func(index int) {
				allocation, testbed, err := poolService.AcquireTestbed(ctx, category.UUID, fmt.Sprintf("e2e-concurrent-user-%d", index))
				results <- &acquireResult{
					index:      index,
					allocation: allocation,
					testbed:    testbed,
					err:        err,
				}
			}(i)
		}

		// Collect results and track unique testbed allocations
		successCount := 0
		allocatedTestbedUUIDs := make(map[string]bool)
		for i := 0; i < concurrency; i++ {
			result := <-results
			if result.err == nil && result.allocation != nil && result.testbed != nil {
				successCount++
				allocatedTestbedUUIDs[result.testbed.UUID] = true
			} else if result.err != nil {
				t.Logf("Request %d failed: %v", result.index, result.err)
			}
		}

		// With retry logic, all concurrent requests should succeed (3 testbeds available)
		if successCount != concurrency {
			t.Errorf("Expected all %d acquisitions to succeed with retry logic, got %d", concurrency, successCount)
		}

		if len(allocatedTestbedUUIDs) != successCount {
			t.Errorf("Each successful acquisition should get a unique testbed, got %d allocations but %d unique testbeds",
				successCount, len(allocatedTestbedUUIDs))
		}

		t.Logf("Concurrent acquisition test: %d/%d succeeded, %d unique testbeds", successCount, concurrency, len(allocatedTestbedUUIDs))
	})

	t.Run("ConcurrentAcquireMoreRequestsThanTestbeds", func(t *testing.T) {
		ctx := context.Background()

		// 重置：释放之前的 testbed
		testbedStorage := storage.NewMySQLTestbedStorage(db)
		allocationStorage := storage.NewMySQLAllocationStorage(db)
		testbedList, _ := testbedStorage.ListTestbedsByCategory(category.UUID)
		for _, tb := range testbedList {
			if tb.Status == models.TestbedStatusAllocated || tb.Status == models.TestbedStatusInUse {
				allocs, _ := allocationStorage.ListAllocationsByTestbed(tb.UUID)
				for _, alloc := range allocs {
					_ = poolService.ReleaseTestbed(ctx, alloc.UUID)
				}
			}
		}

		// 5个并发请求，但只有3个testbed
		concurrency := 5
		results := make(chan *acquireResult, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(index int) {
				allocation, testbed, err := poolService.AcquireTestbed(ctx, category.UUID, fmt.Sprintf("e2e-overflow-user-%d", index))
				results <- &acquireResult{
					index:      index,
					allocation: allocation,
					testbed:    testbed,
					err:        err,
				}
			}(i)
		}

		successCount := 0
		failureCount := 0
		allocatedTestbedUUIDs := make(map[string]bool)

		for i := 0; i < concurrency; i++ {
			result := <-results
			if result.err == nil && result.allocation != nil && result.testbed != nil {
				successCount++
				allocatedTestbedUUIDs[result.testbed.UUID] = true
			} else {
				failureCount++
			}
		}

		// 应该只有3个成功（等于可用testbed数量）
		if successCount != 3 {
			t.Errorf("Expected 3 acquisitions to succeed (equal to available testbeds), got %d", successCount)
		}

		if failureCount != 2 {
			t.Errorf("Expected 2 acquisitions to fail (5 requests - 3 testbeds), got %d", failureCount)
		}

		if len(allocatedTestbedUUIDs) != 3 {
			t.Errorf("Expected 3 unique testbeds to be allocated, got %d", len(allocatedTestbedUUIDs))
		}

		t.Logf("Overflow test: %d succeeded, %d failed, %d unique testbeds", successCount, failureCount, len(allocatedTestbedUUIDs))
	})

	t.Run("QuotaExceeded", func(t *testing.T) {
		ctx := context.Background()

		// First, allocate remaining testbeds to fill the quota
		testbedStorage := storage.NewMySQLTestbedStorage(db)
		testbedList, _ := testbedStorage.ListTestbedsByCategory(category.UUID)

		for _, tb := range testbedList {
			if tb.IsAvailable() {
				_, _, _ = poolService.AcquireTestbed(ctx, category.UUID, "e2e-fill-user")
			}
		}

		// Now try to acquire when quota is full
		_, _, err := poolService.AcquireTestbed(ctx, category.UUID, "e2e-exceeded-user")
		if err == nil {
			t.Error("Expected quota exceeded error or no available testbeds error, got nil")
		}

		t.Logf("Quota exceeded correctly: %v", err)
	})
}

type acquireResult struct {
	index      int
	allocation *models.Allocation
	testbed    *models.Testbed
	err        error
}

// TestAutoExpire_E2E tests automatic expiration functionality
func TestAutoExpire_E2E(t *testing.T) {
	db, poolService, _ := setupTestEnvironment(t)
	defer db.Close()
	defer cleanupTestData(db)

	// Setup with short lifetime for testing
	category := models.NewCategory("e2e-expire-category", "E2E Auto Expire Test Category")
	err := poolService.CreateCategory(category)
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}

	// Set very short lifetime (5 seconds)
	policy := models.NewQuotaPolicy(category.UUID, 1, 3, 10, 5)
	err = poolService.SetQuotaPolicy(policy)
	if err != nil {
		t.Fatalf("Failed to set quota policy: %v", err)
	}

	testbedStorage := storage.NewMySQLTestbedStorage(db)
	resourceStorage := storage.NewMySQLResourceInstanceStorage(db)

	snapshotID := "e2e-snapshot-expire-v1.0"
	instance := models.NewVirtualMachine("192.168.1.200", 22, "password", snapshotID, "e2e-test")
	err = resourceStorage.CreateResourceInstance(instance)
	if err != nil {
		t.Fatalf("Failed to create resource instance: %v", err)
	}

	testbed := models.NewTestbed(
		"e2e-testbed-expire-001",
		category.UUID,
		instance.UUID,
		3306,
		"root",
		"testpass",
	)
	err = testbedStorage.CreateTestbed(testbed)
	if err != nil {
		t.Fatalf("Failed to create testbed: %v", err)
	}

	t.Run("AcquireAndCheckExpiry", func(t *testing.T) {
		ctx := context.Background()
		allocation, _, err := poolService.AcquireTestbed(ctx, category.UUID, "e2e-expire-user")
		if err != nil {
			t.Fatalf("Failed to acquire testbed: %v", err)
		}

		// Check allocation has expiry time
		if allocation.ExpiresAt == nil {
			t.Fatal("Expected ExpiresAt to be set")
		}

		initialExpiry := allocation.ExpiresAt
		remainingSeconds := allocation.GetRemainingSeconds()

		t.Logf("Allocation created with expiry: %v, remaining seconds: %d", initialExpiry, remainingSeconds)

		// Verify it's not expired yet
		if allocation.IsExpired() {
			t.Error("Allocation should not be expired immediately")
		}

		// Wait for expiration
		t.Logf("Waiting for allocation to expire...")
		time.Sleep(6 * time.Second)

		// Check if expired
		if !allocation.IsExpired() {
			t.Error("Allocation should be expired after 6 seconds")
		}

		t.Logf("Allocation correctly expired after waiting")
	})

	t.Run("ManualReleaseAfterExpiry", func(t *testing.T) {
		testbedStorage := storage.NewMySQLTestbedStorage(db)

		// Get the testbed
		testbed, err := testbedStorage.GetTestbedByUUID(testbed.UUID)
		if err != nil {
			t.Fatalf("Failed to get testbed: %v", err)
		}

		if testbed.CurrentAllocUUID == nil {
			t.Fatal("No allocation found on testbed")
		}

		allocationUUID := *testbed.CurrentAllocUUID

		// Manually release (this should work even if expired)
		ctx := context.Background()
		err = poolService.ReleaseTestbed(ctx, allocationUUID)
		if err != nil {
			t.Logf("Manual release after expiry result: %v", err)
		}

		// Wait for async restore
		time.Sleep(500 * time.Millisecond)

		// Verify testbed is available
		testbed, err = testbedStorage.GetTestbedByUUID(testbed.UUID)
		if err != nil {
			t.Fatalf("Failed to get testbed after release: %v", err)
		}

		if testbed.Status != models.TestbedStatusAvailable {
			t.Errorf("Expected testbed to be available after release, got %s", testbed.Status)
		}

		t.Logf("Testbed correctly available after manual release")
	})
}
