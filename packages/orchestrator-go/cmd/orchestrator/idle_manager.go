package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ricardoborges-teachable/ai-setup/packages/orchestrator-go/internal/db"
)

type activeRunCounter func(context.Context) (db.ActiveRunCounts, error)

type idleManagerOptions struct {
	Timeout       time.Duration
	CheckInterval time.Duration
	Tracker       *clientTracker
	ActiveRuns    activeRunCounter
	Shutdown      func(reason string)
}

type idleManager struct {
	timeout       time.Duration
	checkInterval time.Duration
	tracker       *clientTracker
	activeRuns    activeRunCounter
	shutdownFn    func(reason string)
}

type idleStatus struct {
	Enabled              bool               `json:"enabled"`
	TimeoutSeconds       int64              `json:"timeoutSeconds"`
	IdleForSeconds       int64              `json:"idleForSeconds"`
	ShutdownAfterSeconds int64              `json:"shutdownAfterSeconds"`
	LastActivity         string             `json:"lastActivity"`
	BlockingReasons      []string           `json:"blockingReasons,omitempty"`
	ActiveRuns           db.ActiveRunCounts `json:"activeRuns"`
	ActiveRunsError      string             `json:"activeRunsError,omitempty"`
}

func newIdleManager(options idleManagerOptions) *idleManager {
	interval := options.CheckInterval
	if interval <= 0 {
		interval = defaultIdleCheckInterval(options.Timeout)
	}
	return &idleManager{
		timeout:       options.Timeout,
		checkInterval: interval,
		tracker:       options.Tracker,
		activeRuns:    options.ActiveRuns,
		shutdownFn:    options.Shutdown,
	}
}

func defaultIdleCheckInterval(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return 30 * time.Second
	}
	interval := timeout / 4
	if interval < time.Second {
		return time.Second
	}
	if interval > 15*time.Second {
		return 15 * time.Second
	}
	return interval
}

func (m *idleManager) start(ctx context.Context) {
	if m == nil || m.timeout <= 0 || m.shutdownFn == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(m.checkInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				ready, _ := m.shouldShutdown(ctx, now.UTC())
				if ready {
					m.shutdown("idle timeout")
					return
				}
			}
		}
	}()
}

func (m *idleManager) status(ctx context.Context) idleStatus {
	_, status := m.shouldShutdown(ctx, time.Now().UTC())
	return status
}

func (m *idleManager) shouldShutdown(ctx context.Context, now time.Time) (bool, idleStatus) {
	status := idleStatus{
		Enabled:         m != nil && m.timeout > 0,
		LastActivity:    now.UTC().Format(time.RFC3339),
		ActiveRuns:      db.ActiveRunCounts{},
		BlockingReasons: []string{},
	}
	if m == nil {
		return false, status
	}
	if m.timeout > 0 {
		status.TimeoutSeconds = int64(m.timeout.Seconds())
	}

	var clients clientSnapshot
	if m.tracker != nil {
		clients = m.tracker.snapshotAt(now)
		if !clients.lastActivityTime.IsZero() {
			status.LastActivity = clients.lastActivityTime.Format(time.RFC3339)
			status.IdleForSeconds = int64(now.Sub(clients.lastActivityTime).Seconds())
		}
		if clients.Count > 0 {
			status.BlockingReasons = append(status.BlockingReasons, fmt.Sprintf("%d active/recent client(s)", clients.Count))
		}
	}

	if m.activeRuns != nil {
		counts, err := m.activeRuns(ctx)
		if err != nil {
			status.ActiveRunsError = err.Error()
			status.BlockingReasons = append(status.BlockingReasons, "active run count unavailable")
		} else {
			status.ActiveRuns = counts
			activeTotal := activeRunTotal(counts)
			if activeTotal > 0 {
				status.BlockingReasons = append(status.BlockingReasons, fmt.Sprintf("%d active run/job(s)", activeTotal))
			}
		}
	}

	if m.timeout <= 0 {
		return false, status
	}

	remaining := m.timeout - time.Duration(status.IdleForSeconds)*time.Second
	if remaining > 0 {
		status.ShutdownAfterSeconds = int64(remaining.Seconds())
		status.BlockingReasons = append(status.BlockingReasons, "idle timeout not reached")
	} else {
		status.ShutdownAfterSeconds = 0
	}

	return len(status.BlockingReasons) == 0, status
}

func activeRunTotal(counts db.ActiveRunCounts) int {
	if counts.Total > 0 {
		return counts.Total
	}
	return counts.Chains + counts.Teams + counts.Workflows + counts.QueueJobs
}

func (m *idleManager) shutdown(reason string) {
	if m != nil && m.shutdownFn != nil {
		m.shutdownFn(reason)
	}
}
