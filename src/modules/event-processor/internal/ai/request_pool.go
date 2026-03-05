package ai

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github-hub/event-processor/internal/storage"
)

// AIRequestPool manages a global pool of AI request tokens
// Multiple events can compete for tokens from this pool
type AIRequestPool struct {
	mu           sync.Mutex
	semaphore    chan struct{}
	totalSize    int
	available    int
	configStorage *storage.MySQLConfigStorage
}

// global singleton instance
var globalPool *AIRequestPool
var poolOnce sync.Once

// GetGlobalRequestPool returns the singleton AI request pool
func GetGlobalRequestPool(configStorage *storage.MySQLConfigStorage) *AIRequestPool {
	poolOnce.Do(func() {
		globalPool = NewAIRequestPool(configStorage)
	})
	return globalPool
}

// NewAIRequestPool creates a new AI request pool
func NewAIRequestPool(configStorage *storage.MySQLConfigStorage) *AIRequestPool {
	poolSize := 50 // default
	if configStorage != nil {
		if size, err := configStorage.GetAIRequestPoolSize(); err == nil {
			poolSize = size
		}
	}

	log.Printf("[AIRequestPool] Initializing with size: %d", poolSize)

	pool := &AIRequestPool{
		semaphore:    make(chan struct{}, poolSize),
		totalSize:    poolSize,
		available:    poolSize,
		configStorage: configStorage,
	}

	// Pre-fill the semaphore with all available tokens
	for i := 0; i < poolSize; i++ {
		pool.semaphore <- struct{}{}
	}

	return pool
}

// AcquireTokens tries to acquire the specified number of tokens from the pool
// It blocks until enough tokens are available or context is cancelled
// Returns a function that should be called to release the tokens
func (p *AIRequestPool) AcquireTokens(ctx context.Context, count int) (func(), error) {
	if count <= 0 {
		return func() {}, nil
	}

	log.Printf("[AIRequestPool] Event requesting %d tokens (available: %d/%d)",
		count, p.getAvailable(), p.totalSize)

	// Acquire tokens one by one
	acquired := 0
	for i := 0; i < count; i++ {
		select {
		case <-p.semaphore:
			acquired++
			p.decrementAvailable()
		case <-ctx.Done():
			// Context cancelled, release any acquired tokens
			log.Printf("[AIRequestPool] Context cancelled, releasing %d acquired tokens", acquired)
			for j := 0; j < acquired; j++ {
				p.semaphore <- struct{}{}
				p.incrementAvailable()
			}
			return nil, fmt.Errorf("context cancelled while acquiring tokens")
		}
	}

	log.Printf("[AIRequestPool] Acquired %d tokens (remaining: %d/%d)",
		count, p.getAvailable(), p.totalSize)

	// Return a release function
	releaseFunc := func() {
		for i := 0; i < count; i++ {
			p.semaphore <- struct{}{}
			p.incrementAvailable()
		}
		log.Printf("[AIRequestPool] Released %d tokens (available: %d/%d)",
			count, p.getAvailable(), p.totalSize)
	}

	return releaseFunc, nil
}

// TryAcquireTokens tries to acquire tokens without blocking
// Returns true if successful, false if not enough tokens available
func (p *AIRequestPool) TryAcquireTokens(count int) (bool, func()) {
	if count <= 0 {
		return true, func() {}
	}

	p.mu.Lock()
	currentAvailable := p.available
	p.mu.Unlock()

	if currentAvailable < count {
		log.Printf("[AIRequestPool] Not enough tokens: need %d, available %d/%d",
			count, currentAvailable, p.totalSize)
		return false, func() {}
	}

	// Use the blocking AcquireTokens with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	releaseFunc, err := p.AcquireTokens(ctx, count)
	if err != nil {
		return false, func() {}
	}

	return true, releaseFunc
}

// ReloadConfig reloads the pool size from config storage
// This allows runtime adjustment of the pool size
func (p *AIRequestPool) ReloadConfig() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.configStorage == nil {
		return fmt.Errorf("config storage not available")
	}

	newSize, err := p.configStorage.GetAIRequestPoolSize()
	if err != nil {
		return fmt.Errorf("failed to get pool size: %w", err)
	}

	if newSize == p.totalSize {
		return nil // No change
	}

	log.Printf("[AIRequestPool] Reloading pool size: %d -> %d", p.totalSize, newSize)

	if newSize > p.totalSize {
		// Add more tokens
		for i := 0; i < (newSize - p.totalSize); i++ {
			p.semaphore <- struct{}{}
		}
		p.available += (newSize - p.totalSize)
	} else if newSize < p.totalSize {
		// Remove tokens (only available ones)
		toRemove := p.totalSize - newSize
		for i := 0; i < toRemove; i++ {
			select {
			case <-p.semaphore:
				p.available--
			default:
				// No more available tokens to remove
				break
			}
		}
	}

	p.totalSize = newSize
	log.Printf("[AIRequestPool] Pool size reloaded: %d (available: %d)", p.totalSize, p.available)

	return nil
}

// GetStats returns current pool statistics
func (p *AIRequestPool) GetStats() PoolStats {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Count currently in-use by checking semaphore channel length
	inUse := p.totalSize - p.available

	return PoolStats{
		TotalSize: p.totalSize,
		Available: p.available,
		InUse:     inUse,
	}
}

// getAvailable returns the current available token count (thread-safe)
func (p *AIRequestPool) getAvailable() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.available
}

// incrementAvailable increments the available count (thread-safe)
func (p *AIRequestPool) incrementAvailable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.available++
}

// decrementAvailable decrements the available count (thread-safe)
func (p *AIRequestPool) decrementAvailable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.available > 0 {
		p.available--
	}
}

// PoolStats represents the current state of the request pool
type PoolStats struct {
	TotalSize int `json:"total_size"`
	Available int `json:"available"`
	InUse     int `json:"in_use"`
}

// UsagePercentage returns the percentage of pool in use
func (s PoolStats) UsagePercentage() float64 {
	if s.TotalSize == 0 {
		return 0
	}
	return float64(s.InUse) / float64(s.TotalSize) * 100
}

// ReloadGlobalRequestPool reloads the global request pool with new config
// This should be called when the pool size configuration is updated
func ReloadGlobalRequestPool(configStorage *storage.MySQLConfigStorage) {
	if globalPool != nil {
		if err := globalPool.ReloadConfig(); err != nil {
			log.Printf("[AIRequestPool] Failed to reload config: %v", err)
		}
	} else {
		// Initialize if not exists
		GetGlobalRequestPool(configStorage)
	}
}
