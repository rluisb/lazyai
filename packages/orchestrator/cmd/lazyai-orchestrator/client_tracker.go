package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	mcpSessionIDHeader        = "Mcp-Session-Id"
	defaultMCPSessionPruneTTL = 10 * time.Minute
)

type clientTracker struct {
	mu           sync.Mutex
	nextID       uint64
	sessionTTL   time.Duration
	active       map[uint64]clientRecord
	mcpSessions  map[string]*clientRecord
	lastActivity time.Time
}

type clientRecord struct {
	ID             string `json:"id"`
	Kind           string `json:"kind"`
	Method         string `json:"method,omitempty"`
	Path           string `json:"path,omitempty"`
	RemoteAddr     string `json:"remoteAddr,omitempty"`
	ActiveRequests int    `json:"activeRequests"`
	StartedAt      string `json:"startedAt"`
	LastSeen       string `json:"lastSeen"`

	startedAt time.Time
	lastSeen  time.Time
}

type clientSnapshot struct {
	Count             int            `json:"count"`
	ActiveRequests    int            `json:"activeRequests"`
	ActiveMCPSessions int            `json:"activeMcpSessions"`
	RecentMCPSessions int            `json:"recentMcpSessions"`
	LastActivity      string         `json:"lastActivity"`
	Tracking          string         `json:"tracking"`
	Details           []clientRecord `json:"details,omitempty"`

	lastActivityTime time.Time
}

func newClientTracker(sessionTTL time.Duration) *clientTracker {
	now := time.Now().UTC()
	if sessionTTL <= 0 {
		sessionTTL = defaultMCPSessionPruneTTL
	}
	return &clientTracker{
		sessionTTL:   sessionTTL,
		active:       map[uint64]clientRecord{},
		mcpSessions:  map[string]*clientRecord{},
		lastActivity: now,
	}
}

func (t *clientTracker) trackHTTP(kind string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		done := t.begin(r, kind)
		defer done()
		defer func() {
			if kind == "mcp" {
				t.touchMCPSession(sessionIDFromHeaders(w.Header()), r)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (t *clientTracker) begin(r *http.Request, kind string) func() {
	now := time.Now().UTC()
	sessionID := ""
	if kind == "mcp" {
		sessionID = sessionIDFromHeaders(r.Header)
	}

	t.mu.Lock()
	t.nextID++
	requestID := t.nextID
	t.lastActivity = now
	record := clientRecord{
		ID:             fmt.Sprintf("request-%d", requestID),
		Kind:           kind,
		Method:         r.Method,
		Path:           r.URL.Path,
		RemoteAddr:     r.RemoteAddr,
		ActiveRequests: 1,
		StartedAt:      now.Format(time.RFC3339),
		LastSeen:       now.Format(time.RFC3339),
		startedAt:      now,
		lastSeen:       now,
	}
	t.active[requestID] = record
	if sessionID != "" {
		t.upsertMCPSessionLocked(sessionID, r, now).ActiveRequests++
	}
	t.mu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			finishedAt := time.Now().UTC()
			t.mu.Lock()
			delete(t.active, requestID)
			t.lastActivity = finishedAt
			if sessionID != "" {
				if session := t.mcpSessions[sessionID]; session != nil {
					if session.ActiveRequests > 0 {
						session.ActiveRequests--
					}
					touchRecord(session, finishedAt)
				}
			}
			t.mu.Unlock()
		})
	}
}

func (t *clientTracker) touchMCPSession(sessionID string, r *http.Request) {
	if sessionID == "" {
		return
	}
	now := time.Now().UTC()
	t.mu.Lock()
	t.upsertMCPSessionLocked(sessionID, r, now)
	t.lastActivity = now
	t.mu.Unlock()
}

func (t *clientTracker) snapshot() clientSnapshot {
	return t.snapshotAt(time.Now().UTC())
}

func (t *clientTracker) snapshotAt(now time.Time) clientSnapshot {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.pruneExpiredSessionsLocked(now)

	snapshot := clientSnapshot{
		ActiveRequests:   len(t.active),
		LastActivity:     t.lastActivity.Format(time.RFC3339),
		Tracking:         "active HTTP requests plus MCP sessions observed via mcp-session-id; Streamable HTTP clients without a session header are counted while their request is active",
		lastActivityTime: t.lastActivity,
	}
	for _, record := range t.active {
		snapshot.Details = append(snapshot.Details, record)
	}
	idleRecentSessions := 0
	for _, session := range t.mcpSessions {
		snapshot.RecentMCPSessions++
		if session.ActiveRequests > 0 {
			snapshot.ActiveMCPSessions++
		} else {
			idleRecentSessions++
		}
		snapshot.Details = append(snapshot.Details, *session)
	}
	snapshot.Count = snapshot.ActiveRequests + idleRecentSessions
	return snapshot
}

func (t *clientTracker) setLastActivity(at time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastActivity = at.UTC()
}

func (t *clientTracker) upsertMCPSessionLocked(sessionID string, r *http.Request, now time.Time) *clientRecord {
	session := t.mcpSessions[sessionID]
	if session == nil {
		session = &clientRecord{
			ID:        sessionID,
			Kind:      "mcp-session",
			StartedAt: now.Format(time.RFC3339),
			startedAt: now,
		}
		t.mcpSessions[sessionID] = session
	}
	session.Method = r.Method
	session.Path = r.URL.Path
	session.RemoteAddr = r.RemoteAddr
	touchRecord(session, now)
	return session
}

func (t *clientTracker) pruneExpiredSessionsLocked(now time.Time) {
	if t.sessionTTL <= 0 {
		return
	}
	for id, session := range t.mcpSessions {
		if session.ActiveRequests == 0 && !session.lastSeen.IsZero() && now.Sub(session.lastSeen) >= t.sessionTTL {
			delete(t.mcpSessions, id)
		}
	}
}

func touchRecord(record *clientRecord, at time.Time) {
	record.lastSeen = at
	record.LastSeen = at.Format(time.RFC3339)
	if record.startedAt.IsZero() {
		record.startedAt = at
		record.StartedAt = at.Format(time.RFC3339)
	}
}

func sessionIDFromHeaders(header http.Header) string {
	for _, name := range []string{mcpSessionIDHeader, strings.ToLower(mcpSessionIDHeader)} {
		if value := strings.TrimSpace(header.Get(name)); value != "" {
			return value
		}
	}
	return ""
}
