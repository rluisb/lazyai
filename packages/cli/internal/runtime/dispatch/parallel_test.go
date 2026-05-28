package dispatch

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestDispatchWave(t *testing.T) {
	mock := NewMockDispatcher()
	wave := &Wave{
		Tasks: []WaveTask{
			{Agent: "wall-builder", Mode: "standard", Task: "Implement auth"},
			{Agent: "wall-builder", Mode: "standard", Task: "Implement payment"},
			{Agent: "shield-audit", Mode: "review", Task: "Review auth"},
		},
		MaxConcurrent: 2,
		Timeout:       10 * time.Second,
	}

	result, err := DispatchWave(context.Background(), mock, wave)
	if err != nil {
		t.Fatalf("DispatchWave failed: %v", err)
	}

	if result.HasErrors() {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}

	if len(result.Results) != 3 {
		t.Errorf("results = %d, want 3", len(result.Results))
	}
}

func TestDispatchWaveWithError(t *testing.T) {
	mock := NewMockDispatcher()
	mock.Responses["wall-builder:Implement auth"] = &Result{
		Agent: "wall-builder",
		Task:  "Implement auth",
		Error: fmt.Errorf("mock failure"),
	}

	wave := &Wave{
		Tasks: []WaveTask{
			{Agent: "wall-builder", Mode: "standard", Task: "Implement auth"},
			{Agent: "shield-audit", Mode: "review", Task: "Review auth"},
		},
		MaxConcurrent: 2,
	}

	result, err := DispatchWave(context.Background(), mock, wave)
	if err != nil {
		t.Fatalf("DispatchWave failed: %v", err)
	}

	if !result.HasErrors() {
		t.Fatal("expected errors")
	}

	if len(result.Errors) != 1 {
		t.Errorf("errors = %d, want 1", len(result.Errors))
	}

	if len(result.Results) != 1 {
		t.Errorf("results = %d, want 1", len(result.Results))
	}
}

func TestBarrier(t *testing.T) {
	barrier := NewBarrier("test-barrier", 3)

	// Simulate 3 tasks arriving
	go func() {
		time.Sleep(10 * time.Millisecond)
		barrier.Arrive()
	}()
	go func() {
		time.Sleep(20 * time.Millisecond)
		barrier.Arrive()
	}()
	go func() {
		time.Sleep(30 * time.Millisecond)
		barrier.Arrive()
	}()

	// Wait for all
	err := barrier.Wait(1 * time.Second)
	if err != nil {
		t.Fatalf("Barrier wait failed: %v", err)
	}

	arrived, expected, resolved := barrier.Status()
	if arrived != 3 {
		t.Errorf("arrived = %d, want 3", arrived)
	}
	if expected != 3 {
		t.Errorf("expected = %d, want 3", expected)
	}
	if !resolved {
		t.Error("barrier not resolved")
	}
}

func TestBarrierTimeout(t *testing.T) {
	barrier := NewBarrier("timeout-barrier", 3)

	// Only 1 task arrives
	go func() {
		barrier.Arrive()
	}()

	err := barrier.Wait(50 * time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestLockManager(t *testing.T) {
	lm := NewLockManager()

	// Acquire lock
	err := lm.Acquire("test-lock", "agent-1", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	// Verify held
	holder, _, ok := lm.Status("test-lock")
	if !ok {
		t.Fatal("lock not found")
	}
	if holder != "agent-1" {
		t.Errorf("holder = %q, want agent-1", holder)
	}

	// Another agent should timeout
	err = lm.Acquire("test-lock", "agent-2", 50*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout for agent-2")
	}

	// Release
	err = lm.Release("test-lock", "agent-1")
	if err != nil {
		t.Fatalf("Release failed: %v", err)
	}

	// Now agent-2 can acquire
	err = lm.Acquire("test-lock", "agent-2", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Acquire after release failed: %v", err)
	}
}

func TestLockManagerStaleCleanup(t *testing.T) {
	lm := NewLockManager()

	// Create a stale lock
	lm.locks["stale"] = &Lock{
		name:     "stale",
		holder:   "old-agent",
		acquired: time.Now().Add(-1 * time.Hour),
	}

	// New agent should be able to acquire (stale lock)
	err := lm.Acquire("stale", "new-agent", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Acquire should succeed for stale lock: %v", err)
	}

	// Cleanup
	count := lm.CleanupStale(30 * time.Minute)
	if count != 0 {
		t.Errorf("cleanup count = %d, want 0", count)
	}
}
