package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
)

func TestClientTrackerCountsActiveRequests(t *testing.T) {
	tracker := newClientTracker(time.Minute)
	started := make(chan struct{})
	release := make(chan struct{})

	handler := tracker.trackHTTP("mcp", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		w.WriteHeader(http.StatusNoContent)
	}))

	done := make(chan struct{})
	go func() {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/mcp", nil))
		close(done)
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("handler did not start")
	}

	snapshot := tracker.snapshot()
	if snapshot.Count != 1 || snapshot.ActiveRequests != 1 {
		t.Fatalf("expected one active request, got %+v", snapshot)
	}

	close(release)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("handler did not finish")
	}

	snapshot = tracker.snapshot()
	if snapshot.Count != 0 || snapshot.ActiveRequests != 0 {
		t.Fatalf("expected no active clients after response close, got %+v", snapshot)
	}
	if snapshot.LastActivity == "" {
		t.Fatalf("expected last activity to be recorded")
	}
}

func TestClientTrackerKeepsRecentMCPSessionsUntilTTL(t *testing.T) {
	tracker := newClientTracker(20 * time.Millisecond)
	handler := tracker.trackHTTP("mcp", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set(mcpSessionIDHeader, "session-1")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	snapshot := tracker.snapshot()
	if snapshot.RecentMCPSessions != 1 || snapshot.Count != 1 {
		t.Fatalf("expected recent MCP session to be counted, got %+v", snapshot)
	}

	time.Sleep(35 * time.Millisecond)
	snapshot = tracker.snapshot()
	if snapshot.RecentMCPSessions != 0 || snapshot.Count != 0 {
		t.Fatalf("expected MCP session to expire, got %+v", snapshot)
	}
}

func TestClientTrackerUsesDefaultPruneTTLWhenIdleTimeoutDisabled(t *testing.T) {
	tracker := newClientTracker(0)
	if tracker.sessionTTL != defaultMCPSessionPruneTTL {
		t.Fatalf("expected default MCP session prune TTL, got %s", tracker.sessionTTL)
	}

	seenAt := time.Now().UTC().Add(-defaultMCPSessionPruneTTL)
	tracker.mcpSessions["session-1"] = &clientRecord{
		ID:       "session-1",
		Kind:     "mcp-session",
		lastSeen: seenAt,
	}

	snapshot := tracker.snapshotAt(time.Now().UTC())
	if snapshot.RecentMCPSessions != 0 || snapshot.Count != 0 {
		t.Fatalf("expected disabled idle timeout tracker to still prune MCP sessions, got %+v", snapshot)
	}
}

func TestIdleManagerRequiresNoClientsAndNoActiveRuns(t *testing.T) {
	tracker := newClientTracker(time.Minute)
	var activeRuns db.ActiveRunCounts
	shutdownCalled := false
	manager := newIdleManager(idleManagerOptions{
		Timeout: 10 * time.Millisecond,
		Tracker: tracker,
		ActiveRuns: func(context.Context) (db.ActiveRunCounts, error) {
			return activeRuns, nil
		},
		Shutdown: func(string) { shutdownCalled = true },
	})

	tracker.setLastActivity(time.Now().Add(-time.Minute))
	activeRuns.Chains = 1
	ready, status := manager.shouldShutdown(context.Background(), time.Now())
	if ready || len(status.BlockingReasons) == 0 {
		t.Fatalf("expected active runs to block shutdown, ready=%v status=%+v", ready, status)
	}

	activeRuns = db.ActiveRunCounts{}
	done := tracker.begin(httptest.NewRequest(http.MethodGet, "/health", nil), "health")
	ready, status = manager.shouldShutdown(context.Background(), time.Now().Add(time.Minute))
	done()
	if ready || len(status.BlockingReasons) == 0 {
		t.Fatalf("expected active request to block shutdown, ready=%v status=%+v", ready, status)
	}

	ready, status = manager.shouldShutdown(context.Background(), time.Now().Add(time.Minute))
	if !ready {
		t.Fatalf("expected idle daemon to be ready for shutdown, status=%+v", status)
	}
	manager.shutdown("idle timeout")
	if !shutdownCalled {
		t.Fatalf("expected shutdown callback")
	}
}
