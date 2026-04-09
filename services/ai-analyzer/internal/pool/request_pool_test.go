package pool

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewAIRequestPool(t *testing.T) {
	pool := NewAIRequestPool(10)

	if pool.totalSize != 10 {
		t.Errorf("expected total size 10, got %d", pool.totalSize)
	}

	if pool.available != 10 {
		t.Errorf("expected available 10, got %d", pool.available)
	}
}

func TestNewAIRequestPoolZeroSize(t *testing.T) {
	pool := NewAIRequestPool(0)

	// Should default to 50
	if pool.totalSize != 50 {
		t.Errorf("expected total size 50 (default), got %d", pool.totalSize)
	}
}

func TestAcquireTokens(t *testing.T) {
	pool := NewAIRequestPool(5)

	release, err := pool.AcquireTokens(context.Background(), 2)
	if err != nil {
		t.Fatalf("failed to acquire tokens: %v", err)
	}

	if pool.getAvailable() != 3 {
		t.Errorf("expected available 3, got %d", pool.getAvailable())
	}

	release()

	if pool.getAvailable() != 5 {
		t.Errorf("expected available 5 after release, got %d", pool.getAvailable())
	}
}

func TestAcquireTokensZeroCount(t *testing.T) {
	pool := NewAIRequestPool(5)

	release, err := pool.AcquireTokens(context.Background(), 0)
	if err != nil {
		t.Fatalf("failed to acquire tokens: %v", err)
	}

	if pool.getAvailable() != 5 {
		t.Errorf("expected available 5, got %d", pool.getAvailable())
	}

	// Release should be a no-op
	release()

	if pool.getAvailable() != 5 {
		t.Errorf("expected available 5 after release, got %d", pool.getAvailable())
	}
}

func TestAcquireTokensContextCancelled(t *testing.T) {
	pool := NewAIRequestPool(1)

	// Acquire the only token
	release1, _ := pool.AcquireTokens(context.Background(), 1)

	// Try to acquire another token with a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := pool.AcquireTokens(ctx, 1)
	if err == nil {
		t.Error("expected error when context is cancelled")
	}

	// Release the first token
	release1()

	if pool.getAvailable() != 1 {
		t.Errorf("expected available 1, got %d", pool.getAvailable())
	}
}

func TestTryAcquireTokens(t *testing.T) {
	pool := NewAIRequestPool(5)

	success, release := pool.TryAcquireTokens(2)
	if !success {
		t.Error("expected success")
	}

	release()

	if pool.getAvailable() != 5 {
		t.Errorf("expected available 5 after release, got %d", pool.getAvailable())
	}
}

func TestTryAcquireTokensNotEnough(t *testing.T) {
	pool := NewAIRequestPool(2)

	// Acquire all tokens
	release1, _ := pool.AcquireTokens(context.Background(), 2)
	defer release1()

	// Try to acquire more
	success, _ := pool.TryAcquireTokens(1)
	if success {
		t.Error("expected failure when not enough tokens")
	}
}

func TestResizeIncrease(t *testing.T) {
	pool := NewAIRequestPool(5)

	err := pool.Resize(10)
	if err != nil {
		t.Fatalf("failed to resize: %v", err)
	}

	if pool.totalSize != 10 {
		t.Errorf("expected total size 10, got %d", pool.totalSize)
	}

	if pool.getAvailable() != 10 {
		t.Errorf("expected available 10, got %d", pool.getAvailable())
	}
}

func TestResizeDecrease(t *testing.T) {
	pool := NewAIRequestPool(10)

	err := pool.Resize(5)
	if err != nil {
		t.Fatalf("failed to resize: %v", err)
	}

	if pool.totalSize != 5 {
		t.Errorf("expected total size 5, got %d", pool.totalSize)
	}

	if pool.getAvailable() != 5 {
		t.Errorf("expected available 5, got %d", pool.getAvailable())
	}
}

func TestResizeNoChange(t *testing.T) {
	pool := NewAIRequestPool(5)

	err := pool.Resize(5)
	if err != nil {
		t.Fatalf("failed to resize: %v", err)
	}

	if pool.totalSize != 5 {
		t.Errorf("expected total size 5, got %d", pool.totalSize)
	}
}

func TestGetStats(t *testing.T) {
	pool := NewAIRequestPool(10)

	// Acquire some tokens
	release, _ := pool.AcquireTokens(context.Background(), 3)

	stats := pool.GetStats()

	if stats.TotalSize != 10 {
		t.Errorf("expected TotalSize 10, got %d", stats.TotalSize)
	}

	if stats.Available != 7 {
		t.Errorf("expected Available 7, got %d", stats.Available)
	}

	if stats.InUse != 3 {
		t.Errorf("expected InUse 3, got %d", stats.InUse)
	}

	if stats.UsagePercentage() != 30.0 {
		t.Errorf("expected UsagePercentage 30.0, got %f", stats.UsagePercentage())
	}

	release()
}

func TestConcurrentAcquire(t *testing.T) {
	pool := NewAIRequestPool(10)

	var wg sync.WaitGroup
	errors := make(chan error, 20)

	// Spawn 20 goroutines, each trying to acquire 1 token
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			release, err := pool.AcquireTokens(context.Background(), 1)
			if err != nil {
				errors <- err
				return
			}
			// Hold for a bit
			time.Sleep(10 * time.Millisecond)
			release()
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil && err.Error() != "context cancelled while acquiring tokens" {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// All tokens should be released
	if pool.getAvailable() != 10 {
		t.Errorf("expected available 10 after all releases, got %d", pool.getAvailable())
	}
}
