package ai

import (
	"context"
	"sync"
	"testing"
	"time"
)

// helper function to create a test pool with pre-filled semaphore
func newTestPool(size int) *AIRequestPool {
	pool := &AIRequestPool{
		semaphore: make(chan struct{}, size),
		totalSize: size,
		available: size,
	}
	// Pre-fill the semaphore with tokens
	for i := 0; i < size; i++ {
		pool.semaphore <- struct{}{}
	}
	return pool
}

// TestNewRequestPool_NoConfig tests creating a pool without config storage
func TestNewRequestPool_NoConfig(t *testing.T) {
	pool := newTestPool(20)

	if pool.totalSize != 20 {
		t.Errorf("Expected total size 20, got %d", pool.totalSize)
	}

	if pool.available != 20 {
		t.Errorf("Expected available 20, got %d", pool.available)
	}
}

// TestAcquireTokens tests basic token acquisition and release
func TestAcquireTokens(t *testing.T) {
	pool := newTestPool(10)

	// Acquire 3 tokens
	release, err := pool.AcquireTokens(context.Background(), 3)
	if err != nil {
		t.Fatalf("Failed to acquire tokens: %v", err)
	}

	pool.mu.Lock()
	avail := pool.available
	pool.mu.Unlock()

	if avail != 7 {
		t.Errorf("Expected 7 available tokens, got %d", avail)
	}

	// Release tokens
	release()

	pool.mu.Lock()
	avail = pool.available
	pool.mu.Unlock()

	if avail != 10 {
		t.Errorf("Expected 10 available tokens after release, got %d", avail)
	}
}

// TestAcquireTokens_All tests acquiring all tokens
func TestAcquireTokens_All(t *testing.T) {
	pool := newTestPool(5)

	// Acquire all tokens
	release, err := pool.AcquireTokens(context.Background(), 5)
	if err != nil {
		t.Fatalf("Failed to acquire tokens: %v", err)
	}

	pool.mu.Lock()
	avail := pool.available
	pool.mu.Unlock()

	if avail != 0 {
		t.Errorf("Expected 0 available tokens, got %d", avail)
	}

	release()

	pool.mu.Lock()
	avail = pool.available
	pool.mu.Unlock()

	if avail != 5 {
		t.Errorf("Expected 5 available tokens after release, got %d", avail)
	}
}

// TestAcquireTokens_Blocking tests that acquisition blocks when pool is exhausted
func TestAcquireTokens_Blocking(t *testing.T) {
	pool := newTestPool(3)

	// Acquire all tokens
	release1, _ := pool.AcquireTokens(context.Background(), 3)

	// Try to acquire more - should block
	acquired := make(chan bool)
	go func() {
		release2, err := pool.AcquireTokens(context.Background(), 1)
		if err != nil {
			t.Errorf("Failed to acquire tokens: %v", err)
			return
		}
		defer release2()
		acquired <- true
	}()

	// Should not acquire immediately
	select {
	case <-acquired:
		release1() // Clean up before failing
		t.Fatal("Should have blocked waiting for tokens")
	case <-time.After(100 * time.Millisecond):
		// Expected to block
	}

	// Release first batch
	release1()

	// Now second acquire should succeed
	select {
	case <-acquired:
		// Success
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Should have acquired tokens after release")
	}
}

// TestAcquireTokens_ContextCancellation tests context cancellation
func TestAcquireTokens_ContextCancellation(t *testing.T) {
	pool := newTestPool(2)

	// Acquire all tokens
	release1, _ := pool.AcquireTokens(context.Background(), 2)
	defer release1()

	// Create cancelable context
	ctx, cancel := context.WithCancel(context.Background())

	// Start goroutine that tries to acquire
	errChan := make(chan error, 1)
	go func() {
		_, err := pool.AcquireTokens(ctx, 1)
		errChan <- err
	}()

	// Cancel context immediately
	cancel()

	// Should get context cancelled error
	select {
	case err := <-errChan:
		if err == nil {
			t.Fatal("Expected error from cancelled context")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Should have received cancellation error")
	}
}

// TestTryAcquireTokens_Success tests successful non-blocking acquisition
func TestTryAcquireTokens_Success(t *testing.T) {
	pool := newTestPool(10)

	success, release := pool.TryAcquireTokens(5)
	if !success {
		t.Fatal("Expected TryAcquireTokens to succeed")
	}
	defer release()

	pool.mu.Lock()
	avail := pool.available
	pool.mu.Unlock()

	if avail != 5 {
		t.Errorf("Expected 5 available tokens, got %d", avail)
	}
}

// TestTryAcquireTokens_Failure tests failed non-blocking acquisition
func TestTryAcquireTokens_Failure(t *testing.T) {
	pool := newTestPool(3)

	// Acquire all tokens
	release1, _ := pool.AcquireTokens(context.Background(), 3)
	defer release1()

	// Try to acquire more - should fail
	success, _ := pool.TryAcquireTokens(1)
	if success {
		t.Fatal("Expected TryAcquireTokens to fail")
	}
}

// TestGetStats tests pool statistics
func TestGetStats(t *testing.T) {
	pool := newTestPool(20)

	stats := pool.GetStats()

	if stats.TotalSize != 20 {
		t.Errorf("Expected TotalSize 20, got %d", stats.TotalSize)
	}

	if stats.Available != 20 {
		t.Errorf("Expected Available 20, got %d", stats.Available)
	}

	if stats.InUse != 0 {
		t.Errorf("Expected InUse 0, got %d", stats.InUse)
	}

	// Acquire some tokens
	release, _ := pool.AcquireTokens(context.Background(), 5)

	stats = pool.GetStats()
	if stats.Available != 15 {
		t.Errorf("Expected Available 15, got %d", stats.Available)
	}

	if stats.InUse != 5 {
		t.Errorf("Expected InUse 5, got %d", stats.InUse)
	}

	if stats.UsagePercentage() != 25.0 {
		t.Errorf("Expected UsagePercentage 25.0, got %f", stats.UsagePercentage())
	}

	release()

	stats = pool.GetStats()
	if stats.Available != 20 {
		t.Errorf("Expected Available 20 after release, got %d", stats.Available)
	}
}

// TestConcurrentAcquisitions tests concurrent token acquisitions
func TestConcurrentAcquisitions(t *testing.T) {
	pool := newTestPool(10)

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	// Start 20 goroutines trying to acquire 1 token each
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			release, err := pool.AcquireTokens(context.Background(), 1)
			if err != nil {
				return
			}
			defer release()

			mu.Lock()
			successCount++
			mu.Unlock()

			// Hold token for a short time
			time.Sleep(10 * time.Millisecond)
		}()
	}

	wg.Wait()

	if successCount != 20 {
		t.Errorf("Expected all 20 acquisitions to succeed eventually, got %d", successCount)
	}

	// All tokens should be available now
	pool.mu.Lock()
	avail := pool.available
	pool.mu.Unlock()

	if avail != 10 {
		t.Errorf("Expected 10 available tokens after all releases, got %d", avail)
	}
}

// TestPoolStats_UsagePercentage tests usage percentage calculation
func TestPoolStats_UsagePercentage(t *testing.T) {
	tests := []struct {
		name        string
		totalSize   int
		inUse       int
		expectedPct float64
	}{
		{"Empty pool", 10, 0, 0},
		{"Half full", 10, 5, 50.0},
		{"Full", 10, 10, 100.0},
		{"Zero total", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := PoolStats{
				TotalSize: tt.totalSize,
				InUse:     tt.inUse,
			}

			pct := stats.UsagePercentage()
			if pct != tt.expectedPct {
				t.Errorf("Expected usage percentage %f, got %f", tt.expectedPct, pct)
			}
		})
	}
}

// TestAcquireTokens_Zero tests acquiring zero tokens
func TestAcquireTokens_Zero(t *testing.T) {
	pool := newTestPool(10)

	release, err := pool.AcquireTokens(context.Background(), 0)
	if err != nil {
		t.Fatalf("Failed to acquire 0 tokens: %v", err)
	}

	if release == nil {
		t.Fatal("Release function should not be nil even for 0 tokens")
	}

	// Calling release should be safe
	release()

	pool.mu.Lock()
	avail := pool.available
	pool.mu.Unlock()

	if avail != 10 {
		t.Errorf("Pool should be unchanged after acquiring 0 tokens, got %d available", avail)
	}
}

// TestIncrementDecrementAvailable tests internal counter management
func TestIncrementDecrementAvailable(t *testing.T) {
	pool := newTestPool(10)

	// Test decrement
	pool.decrementAvailable()
	pool.mu.Lock()
	avail := pool.available
	pool.mu.Unlock()

	if avail != 9 {
		t.Errorf("Expected 9 available after decrement, got %d", avail)
	}

	// Test increment
	pool.incrementAvailable()
	pool.mu.Lock()
	avail = pool.available
	pool.mu.Unlock()

	if avail != 10 {
		t.Errorf("Expected 10 available after increment, got %d", avail)
	}

	// Test decrement below zero (should not go negative)
	for i := 0; i < 15; i++ {
		pool.decrementAvailable()
	}
	pool.mu.Lock()
	avail = pool.available
	pool.mu.Unlock()

	if avail < 0 {
		t.Errorf("Available should not go negative, got %d", avail)
	}
}
