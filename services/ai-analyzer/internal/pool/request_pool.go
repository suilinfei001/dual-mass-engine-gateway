package pool

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/quality-gateway/ai-analyzer/internal/types"
)

// AIRequestPool manages a global pool of AI request tokens
// Multiple events can compete for tokens from this pool
type AIRequestPool struct {
	mu        sync.Mutex
	semaphore chan struct{}
	totalSize int
	available int
}

// NewAIRequestPool creates a new AI request pool
func NewAIRequestPool(size int) *AIRequestPool {
	if size <= 0 {
		size = 50 // default
	}

	log.Printf("[AIRequestPool] Initializing with size: %d", size)

	pool := &AIRequestPool{
		semaphore: make(chan struct{}, size),
		totalSize: size,
		available: size,
	}

	// Pre-fill the semaphore with all available tokens
	for i := 0; i < size; i++ {
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

	log.Printf("[AIRequestPool] Requesting %d tokens (available: %d/%d)",
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

// Resize adjusts the pool size
func (p *AIRequestPool) Resize(newSize int) error {
	if newSize <= 0 {
		newSize = 50 // default
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if newSize == p.totalSize {
		return nil // No change
	}

	log.Printf("[AIRequestPool] Resizing pool: %d -> %d", p.totalSize, newSize)

	// Calculate current in-use tokens
	inUse := p.totalSize - p.available

	// Create new semaphore with new size
	newSem := make(chan struct{}, newSize)

	// Add available tokens to new semaphore
	newAvailable := newSize - inUse
	for i := 0; i < newAvailable; i++ {
		newSem <- struct{}{}
	}

	// Replace old semaphore
	p.semaphore = newSem
	p.totalSize = newSize
	p.available = newAvailable

	log.Printf("[AIRequestPool] Pool size resized: %d (available: %d)", p.totalSize, p.available)

	return nil
}

// GetStats returns current pool statistics
func (p *AIRequestPool) GetStats() types.PoolStats {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Count currently in-use
	inUse := p.totalSize - p.available

	return types.PoolStats{
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
