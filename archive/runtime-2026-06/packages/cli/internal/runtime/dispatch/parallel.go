// Package dispatch provides agent execution and parallel coordination primitives.
//
// This replaces the bash dispatch scripts (dispatch-wave.sh, wait-barrier.sh,
// task-barrier.sh, task-lock.sh) with Go-native implementations using goroutines
// and channels for synchronization.
package dispatch

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Wave represents a parallel dispatch wave
type Wave struct {
	Tasks         []WaveTask
	MaxConcurrent int
	Timeout       time.Duration
}

// WaveTask is a single task in a dispatch wave
type WaveTask struct {
	Agent      string
	Mode       string
	Task       string
	OutputPath string
}

// WaveResult contains the outcome of a wave dispatch
type WaveResult struct {
	Results map[string]*Result
	Errors  map[string]error
	mu      sync.RWMutex
}

// NewWaveResult creates a wave result collector
func NewWaveResult() *WaveResult {
	return &WaveResult{
		Results: make(map[string]*Result),
		Errors:  make(map[string]error),
	}
}

// SetResult records a task result
func (wr *WaveResult) SetResult(key string, result *Result) {
	wr.mu.Lock()
	defer wr.mu.Unlock()
	wr.Results[key] = result
}

// SetError records a task error
func (wr *WaveResult) SetError(key string, err error) {
	wr.mu.Lock()
	defer wr.mu.Unlock()
	wr.Errors[key] = err
}

// GetResult retrieves a task result
func (wr *WaveResult) GetResult(key string) (*Result, bool) {
	wr.mu.RLock()
	defer wr.mu.RUnlock()
	r, ok := wr.Results[key]
	return r, ok
}

// HasErrors returns true if any task failed
func (wr *WaveResult) HasErrors() bool {
	wr.mu.RLock()
	defer wr.mu.RUnlock()
	return len(wr.Errors) > 0
}

// DispatchWave executes tasks in parallel with concurrency limit
func DispatchWave(ctx context.Context, d Dispatcher, wave *Wave) (*WaveResult, error) {
	if len(wave.Tasks) == 0 {
		return NewWaveResult(), nil
	}

	result := NewWaveResult()

	// Create semaphore for concurrency control
	maxConcurrent := wave.MaxConcurrent
	if maxConcurrent <= 0 {
		maxConcurrent = 4
	}

	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	// Context with timeout
	if wave.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, wave.Timeout)
		defer cancel()
	}

	// Dispatch each task
	for _, task := range wave.Tasks {
		wg.Add(1)
		go func(t WaveTask) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				result.SetError(t.Agent+":"+t.Task, ctx.Err())
				return
			}

			// Execute task
			r, err := d.Dispatch("", t.Agent, t.Task, t.Mode)
			if err != nil {
				result.SetError(t.Agent+":"+t.Task, err)
				return
			}

			result.SetResult(t.Agent+":"+t.Task, r)
		}(task)
	}

	wg.Wait()
	return result, nil
}

// Barrier is a synchronization primitive for parallel tasks
type Barrier struct {
	id            string
	expectedCount int
	arrivedCount  int
	mu            sync.Mutex
	cond          *sync.Cond
	resolved      bool
}

// NewBarrier creates a barrier for N tasks
func NewBarrier(id string, expectedCount int) *Barrier {
	b := &Barrier{
		id:            id,
		expectedCount: expectedCount,
	}
	b.cond = sync.NewCond(&b.mu)
	return b
}

// Arrive signals that a task has reached the barrier
func (b *Barrier) Arrive() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.resolved {
		return fmt.Errorf("barrier %s already resolved", b.id)
	}

	b.arrivedCount++
	if b.arrivedCount >= b.expectedCount {
		b.resolved = true
		b.cond.Broadcast()
	}

	return nil
}

// Wait blocks until all tasks arrive or timeout
func (b *Barrier) Wait(timeout time.Duration) error {
	// Fast path: already resolved
	b.mu.Lock()
	if b.resolved {
		b.mu.Unlock()
		return nil
	}
	b.mu.Unlock()

	// Poll with timeout
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		b.mu.Lock()
		if b.resolved {
			b.mu.Unlock()
			return nil
		}
		b.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}

	return fmt.Errorf("barrier %s timeout after %v", b.id, timeout)
}

// Status returns current barrier status
func (b *Barrier) Status() (arrived int, expected int, resolved bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.arrivedCount, b.expectedCount, b.resolved
}

// Lock is a named exclusive lock for shared resources
type Lock struct {
	name     string
	holder   string
	acquired time.Time
	mu       sync.Mutex
}

// LockManager manages named locks
type LockManager struct {
	locks map[string]*Lock
	mu    sync.RWMutex
}

// NewLockManager creates a lock manager
func NewLockManager() *LockManager {
	return &LockManager{
		locks: make(map[string]*Lock),
	}
}

// Acquire attempts to acquire a named lock
func (lm *LockManager) Acquire(name string, holder string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if lm.tryAcquire(name, holder) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("lock %s timeout after %v", name, timeout)
}

// tryAcquire attempts to acquire without waiting
func (lm *LockManager) tryAcquire(name string, holder string) bool {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if existing, ok := lm.locks[name]; ok {
		// Check if lock is stale (> 30 min)
		if time.Since(existing.acquired) > 30*time.Minute {
			// Stale lock, take over
			lm.locks[name] = &Lock{
				name:     name,
				holder:   holder,
				acquired: time.Now(),
			}
			return true
		}
		return false
	}

	lm.locks[name] = &Lock{
		name:     name,
		holder:   holder,
		acquired: time.Now(),
	}
	return true
}

// Release releases a named lock
func (lm *LockManager) Release(name string, holder string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lock, ok := lm.locks[name]
	if !ok {
		return fmt.Errorf("lock %s not found", name)
	}

	if lock.holder != holder {
		return fmt.Errorf("lock %s held by %s, not %s", name, lock.holder, holder)
	}

	delete(lm.locks, name)
	return nil
}

// Status returns lock status
func (lm *LockManager) Status(name string) (holder string, acquired time.Time, ok bool) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	lock, ok := lm.locks[name]
	if !ok {
		return "", time.Time{}, false
	}
	return lock.holder, lock.acquired, true
}

// CleanupStale removes locks older than maxAge
func (lm *LockManager) CleanupStale(maxAge time.Duration) int {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	count := 0
	for name, lock := range lm.locks {
		if time.Since(lock.acquired) > maxAge {
			delete(lm.locks, name)
			count++
		}
	}
	return count
}
