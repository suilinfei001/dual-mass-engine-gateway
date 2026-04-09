// +build integration

package performance

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hugoh/go-designs/resource-pool/internal/deployer"
	"github.com/hugoh/go-designs/resource-pool/internal/models"
	"github.com/hugoh/go-designs/resource-pool/internal/service"
	"github.com/hugoh/go-designs/resource-pool/internal/storage"
)

const benchmarkDSN = "root:root123456@tcp(localhost:3307)/event_processor?parseTime=true&loc=Local"

// Benchmark service for reuse across benchmarks
var benchmarkPoolService service.ResourcePoolService
var benchmarkDB *sql.DB
var benchmarkCategoryUUID string

// TestMain sets up benchmark environment
func TestMain(m *testing.M) {
	// Check if database is available
	db, err := sql.Open("mysql", benchmarkDSN)
	if err != nil {
		fmt.Printf("Cannot connect to test database: %v\n", err)
		fmt.Println("Skipping benchmarks - database not available")
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Printf("Cannot ping test database: %v\n", err)
		fmt.Println("Skipping benchmarks - database not accessible")
		return
	}

	benchmarkDB = db

	// Setup test data
	setupBenchmarkData()

	// Run benchmarks
	m.Run()

	// Cleanup
	cleanupBenchmarkData()
}

func setupBenchmarkData() {
	// First cleanup any existing perf data
	cleanupBenchmarkData()

	// Create storage instances
	testbedStorage := storage.NewMySQLTestbedStorage(benchmarkDB)
	allocationStorage := storage.NewMySQLAllocationStorage(benchmarkDB)
	categoryStorage := storage.NewMySQLCategoryStorage(benchmarkDB)
	quotaStorage := storage.NewMySQLQuotaPolicyStorage(benchmarkDB)
	resourceStorage := storage.NewMySQLResourceInstanceStorage(benchmarkDB)

	// Create category with unique name to avoid conflicts
	category := models.NewCategory(fmt.Sprintf("perf-category-%d", time.Now().UnixNano()), "Performance Test Category")
	if err := categoryStorage.CreateCategory(category); err != nil {
		panic(fmt.Sprintf("Failed to create category: %v", err))
	}
	benchmarkCategoryUUID = category.UUID

	// Create quota policy with high limit
	policy := models.NewQuotaPolicy(category.UUID, 10, 50, 10, 3600)
	quotaStorage.CreateQuotaPolicy(policy)

	// Create multiple resource instances and testbeds
	for i := 0; i < 20; i++ {
		snapshotID := "perf-snapshot-v1.0"
		instance := models.NewVirtualMachine(
			fmt.Sprintf("192.168.1.%d", 200+i%55),
			22,
			"password",
			snapshotID,
			"perf-test",
		)
		resourceStorage.CreateResourceInstance(instance)

		testbed := models.NewTestbed(
			fmt.Sprintf("perf-testbed-%03d", i+1),
			category.UUID,
			instance.UUID,
			3306,
			"root",
			"testpass",
		)
		testbedStorage.CreateTestbed(testbed)
	}

	// Create deployer and service
	mockDeployer := deployer.NewMockDeployService()
	mockDeployer.SetDelays(10*time.Millisecond, 10*time.Millisecond)

	benchmarkPoolService = service.NewResourcePoolService(
		testbedStorage,
		allocationStorage,
		categoryStorage,
		quotaStorage,
		resourceStorage,
		mockDeployer,
		3600,
	)
}

func cleanupBenchmarkData() {
	if benchmarkDB == nil {
		return
	}

	// Clean up in reverse order of dependencies
	benchmarkDB.Exec("DELETE FROM allocations WHERE testbed_uuid IN (SELECT uuid FROM testbeds WHERE name LIKE 'perf-%')")
	benchmarkDB.Exec("DELETE FROM testbeds WHERE name LIKE 'perf-%'")
	benchmarkDB.Exec("DELETE FROM resource_instances WHERE created_by = 'perf-test'")
	benchmarkDB.Exec("DELETE FROM quota_policies WHERE category_uuid IN (SELECT uuid FROM categories WHERE name LIKE 'perf-category-%')")
	benchmarkDB.Exec("DELETE FROM categories WHERE name LIKE 'perf-category-%'")
}

// BenchmarkAcquireTestbed benchmarks the acquire operation
func BenchmarkAcquireTestbed(b *testing.B) {
	if benchmarkPoolService == nil {
		b.Skip("Benchmark service not initialized")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		allocation, testbed, err := benchmarkPoolService.AcquireTestbed(ctx, benchmarkCategoryUUID, fmt.Sprintf("bench-user-%d", i))
		if err != nil {
			b.Fatalf("Failed to acquire: %v", err)
		}
		// Immediately release for next iteration
		_ = benchmarkPoolService.ReleaseTestbed(ctx, allocation.UUID)
		_ = testbed
	}
}

// BenchmarkListTestbeds benchmarks listing testbeds
func BenchmarkListTestbeds(b *testing.B) {
	if benchmarkPoolService == nil {
		b.Skip("Benchmark service not initialized")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := benchmarkPoolService.ListTestbeds(nil, nil)
		if err != nil {
			b.Fatalf("Failed to list: %v", err)
		}
	}
}

// BenchmarkListCategories benchmarks listing categories
func BenchmarkListCategories(b *testing.B) {
	if benchmarkPoolService == nil {
		b.Skip("Benchmark service not initialized")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := benchmarkPoolService.ListCategories()
		if err != nil {
			b.Fatalf("Failed to list categories: %v", err)
		}
	}
}

// BenchmarkGetQuotaPolicy benchmarks getting quota policy
func BenchmarkGetQuotaPolicy(b *testing.B) {
	if benchmarkPoolService == nil {
		b.Skip("Benchmark service not initialized")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := benchmarkPoolService.GetQuotaPolicy(benchmarkCategoryUUID)
		if err != nil {
			b.Fatalf("Failed to get quota: %v", err)
		}
	}
}

// BenchmarkConcurrentAcquire benchmarks concurrent acquire operations
func BenchmarkConcurrentAcquire(b *testing.B) {
	if benchmarkPoolService == nil {
		b.Skip("Benchmark service not initialized")
	}

	ctx := context.Background()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			allocation, _, err := benchmarkPoolService.AcquireTestbed(ctx, benchmarkCategoryUUID, fmt.Sprintf("parallel-user-%d", i))
			if err != nil {
				// Testbed might be exhausted, skip this iteration
				continue
			}
			// Release immediately
			_ = benchmarkPoolService.ReleaseTestbed(ctx, allocation.UUID)
			i++
		}
	})
}

// TestLoadAllocation simulates sustained load with concurrent allocations
func TestLoadAllocation(t *testing.T) {
	if benchmarkPoolService == nil {
		t.Skip("Benchmark service not initialized")
	}

	ctx := context.Background()
	const duration = 5 * time.Second
	const workers = 10

	var wg sync.WaitGroup
	results := make(chan int, workers*100)
	startTime := time.Now()

	// Launch workers
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			successCount := 0

			for time.Since(startTime) < duration {
				allocation, _, err := benchmarkPoolService.AcquireTestbed(ctx, benchmarkCategoryUUID, fmt.Sprintf("load-user-%d-%d", workerID, successCount))
				if err != nil {
					// Testbed exhausted or other error, retry after brief pause
					time.Sleep(10 * time.Millisecond)
					continue
				}

				successCount++
				// Simulate some work
				time.Sleep(50 * time.Millisecond)

				// Release
				_ = benchmarkPoolService.ReleaseTestbed(ctx, allocation.UUID)
			}

			results <- successCount
		}(w)
	}

	// Wait for all workers to complete
	wg.Wait()
	close(results)

	// Collect results
	totalOperations := 0
	for count := range results {
		totalOperations += count
	}

	operationsPerSecond := float64(totalOperations) / duration.Seconds()

	t.Logf("Load test results:")
	t.Logf("  Duration: %v", duration)
	t.Logf("  Workers: %d", workers)
	t.Logf("  Total operations: %d", totalOperations)
	t.Logf("  Operations per second: %.2f", operationsPerSecond)

	// Performance assertion: should handle at least 5 ops/sec
	if operationsPerSecond < 5 {
		t.Errorf("Performance too low: %.2f ops/sec, expected at least 5 ops/sec", operationsPerSecond)
	}
}

// BenchmarkAllocateReleaseCycle benchmarks the full allocate-use-release cycle
func BenchmarkAllocateReleaseCycle(b *testing.B) {
	if benchmarkPoolService == nil {
		b.Skip("Benchmark service not initialized")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Allocate
		allocation, testbed, err := benchmarkPoolService.AcquireTestbed(ctx, benchmarkCategoryUUID, fmt.Sprintf("cycle-user-%d", i))
		if err != nil {
			// Testbed exhausted, reset test data
			cleanupBenchmarkData()
			setupBenchmarkData()
			continue
		}

		// Simulate some work
		_ = testbed.Status

		// Release
		_ = benchmarkPoolService.ReleaseTestbed(ctx, allocation.UUID)
	}
}

// BenchmarkExtendAllocation benchmarks the extend operation
func BenchmarkExtendAllocation(b *testing.B) {
	if benchmarkPoolService == nil {
		b.Skip("Benchmark service not initialized")
	}

	ctx := context.Background()

	// Pre-allocate for the benchmark
	allocations := make([]string, 100)
	for i := 0; i < 100; i++ {
		allocation, _, err := benchmarkPoolService.AcquireTestbed(ctx, benchmarkCategoryUUID, fmt.Sprintf("extend-setup-%d", i))
		if err != nil {
			b.Skip("Could not setup allocations for benchmark")
			return
		}
		allocations[i] = allocation.UUID
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := i % len(allocations)
		err := benchmarkPoolService.ExtendAllocation(ctx, allocations[idx], 1800)
		if err != nil {
			b.Fatalf("Failed to extend: %v", err)
		}
	}

	// Cleanup
	for _, uuid := range allocations {
		_ = benchmarkPoolService.ReleaseTestbed(ctx, uuid)
	}
}

// BenchmarkStorageGetTestbed benchmarks storage layer get operation
func BenchmarkStorageGetTestbed(b *testing.B) {
	if benchmarkDB == nil {
		b.Skip("Benchmark database not initialized")
	}

	testbedStorage := storage.NewMySQLTestbedStorage(benchmarkDB)

	// Get a test testbed UUID
	testbed, _ := testbedStorage.ListTestbeds()
	if len(testbed) == 0 {
		b.Skip("No testbed available for benchmark")
	}
	uuid := testbed[0].UUID

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := testbedStorage.GetTestbedByUUID(uuid)
		if err != nil {
			b.Fatalf("Failed to get testbed: %v", err)
		}
	}
}

// BenchmarkStorageListAvailable benchmarks listing available testbeds
func BenchmarkStorageListAvailable(b *testing.B) {
	if benchmarkDB == nil {
		b.Skip("Benchmark database not initialized")
	}

	testbedStorage := storage.NewMySQLTestbedStorage(benchmarkDB)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := testbedStorage.ListAvailableTestbeds(benchmarkCategoryUUID)
		if err != nil {
			b.Fatalf("Failed to list available: %v", err)
		}
	}
}
