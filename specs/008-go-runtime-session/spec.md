# Spec 008: Go Runtime — Session Management

## Chunk 2: Session Lifecycle, Dispatches, Parallel Tasks

**Status:** Draft  
**Author:** Ricardo Conceicao  
**Date:** 2026-05-26  
**Depends on:** Spec 007 (Foundation)

---

## §G Goal

Implement the session management layer that replaces `session-db.sh` commands:
- `init` — Start a new session
- `list` — List all sessions
- `show` — Show session details with dispatches
- `end` — End a session
- `dispatch` — Record an agent dispatch
- `parallel` — Track parallel task execution
- `message` — Inter-agent messaging
- `barrier` — Synchronization barriers
- `lock` — Resource locks

---

## §C Constraints

1. **Depends on Spec 007** — Requires `runtime.DB` and schema
2. **Session ID format** — `ses_<unixtimestamp>` (backward compatible)
3. **Dispatch sequencing** — `seq` auto-increments per session
4. **Atomic operations** — All state changes within transactions
5. **Ledger integration** — Every significant event appends to ledger

---

## §I Interfaces

### §I.1 Session Manager

```go
package session

import "github.com/rluisb/lazyai/packages/cli/internal/runtime"

// Manager handles session lifecycle operations
type Manager struct {
    db *runtime.DB
}

// NewManager creates a session manager
func NewManager(db *runtime.DB) *Manager

// Start creates a new session
func (m *Manager) Start(goal string, opts StartOptions) (*Session, error)

// End terminates a session
func (m *Manager) End(sessionID string) error

// Get retrieves a session by ID
func (m *Manager) Get(sessionID string) (*Session, error)

// List returns all sessions ordered by start time
func (m *Manager) List() ([]Session, error)

// Dispatch records an agent dispatch
func (m *Manager) Dispatch(sessionID string, opts DispatchOptions) (*Dispatch, error)

// ListDispatches returns dispatches for a session
func (m *Manager) ListDispatches(sessionID string) ([]Dispatch, error)
```

### §I.2 Data Types

```go
// Session represents an AI agent session
type Session struct {
    ID        string
    StartedAt time.Time
    EndedAt   *time.Time
    Agent     string
    Model     string
    Goal      string
    Repo      string
    Worktree  string
    Status    SessionStatus
    TokenTotal int
    Summary   string
    Tags      []string
}

// SessionStatus enum
type SessionStatus string
const (
    SessionActive   SessionStatus = "active"
    SessionEnded    SessionStatus = "ended"
    SessionFailed   SessionStatus = "failed"
)

// Dispatch represents an agent dispatch record
type Dispatch struct {
    ID           int
    SessionID    string
    Seq          int
    ParentID     *int
    Agent        string
    Model        string
    Task         string
    Phase        string
    Workflow     string
    Mode         string
    StartedAt    *time.Time
    EndedAt      *time.Time
    Result       string
    TokenUsed    int
    ErrorMessage string
    Summary      string
    FilesTouched []string
}
```

---

## §V Invariants

1. **Session ID uniqueness** — `ses_<timestamp>` must be unique (nanosecond precision)
2. **Dispatch sequence monotonicity** — `seq` increments by 1 per session, no gaps
3. **Active session limit** — Only one active session per worktree (configurable)
4. **Parent dispatch validity** — `parent_id` must reference a dispatch in the same session
5. **Token accumulation** — `token_total` in sessions is the sum of all dispatch `token_used`

---

## §T Tasks

### Task 1: Session CRUD
**File:** `packages/cli/internal/runtime/session/session.go`

Implement:
- `Start()` — Insert into `sessions`, return Session
- `End()` — Update `status='ended'`, set `ended_at`
- `Get()` — Select by ID with all fields
- `List()` — Select all, order by `started_at DESC`
- `UpdateSummary()` — Update summary field
- `AddTags()` — Append tags

### Task 2: Dispatch Tracking
**File:** `packages/cli/internal/runtime/session/dispatch.go`

Implement:
- `Dispatch()` — Insert into `dispatches` with auto-seq
- `CompleteDispatch()` — Update `ended_at`, `result`, `token_used`
- `ListDispatches()` — Select by session_id
- `GetDispatch()` — Select by id
- `GetLastDispatch()` — Select latest by seq

### Task 3: Parallel Task Tracking
**File:** `packages/cli/internal/runtime/session/parallel.go`

Implement:
- `CreateParallelTask()` — Insert into `parallel_tasks`
- `UpdateParallelTask()` — Update status, result, completed_at
- `ListParallelTasks()` — Select by session_id or wave_id
- `GetWaveSummary()` — Aggregate results by wave

### Task 4: Messaging
**File:** `packages/cli/internal/runtime/session/message.go`

Implement:
- `SendMessage()` — Insert into `messages`
- `GetMessages()` — Select by session_id, filter by to_agent
- `MarkRead()` — Update `read_at`
- `GetUnreadCount()` — Count unread for agent

### Task 5: Barriers
**File:** `packages/cli/internal/runtime/session/barrier.go`

Implement:
- `CreateBarrier()` — Insert into `barriers`
- `ArriveAtBarrier()` — Increment `arrived_count`, check if resolved
- `GetBarrierStatus()` — Select by barrier_id
- `ResolveBarrier()` — Update `status='resolved'`, set `resolved_at`

### Task 6: Locks
**File:** `packages/cli/internal/runtime/session/lock.go`

Implement:
- `AcquireLock()` — Insert into `locks` if not held
- `ReleaseLock()` — Update `status='released'`, set `released_at`
- `GetLockStatus()` — Select by lock_name
- `ListActiveLocks()` — Select where status='active'

### Task 7: Tests
**Files:** `packages/cli/internal/runtime/session/*_test.go`

Required tests:
- `TestSessionLifecycle` — Start, Get, End
- `TestDispatchSequence` — Verify seq increments
- `TestDispatchParent` — Parent-child relationship
- `TestParallelTasks` — Create, update, list
- `TestMessaging` — Send, receive, mark read
- `TestBarrierResolution` — Create, arrive, resolve
- `TestLockAcquireRelease` — Acquire, release, conflict
- `TestConcurrentAccess` — Multiple goroutines

---

## §B Backward Compatibility

### Bash Command Mapping

| Bash Command | Go Method | Notes |
|-------------|-----------|-------|
| `session-db.sh init` | `session.Start()` | Same session ID format |
| `session-db.sh list` | `session.List()` | Same output columns |
| `session-db.sh show` | `session.Get()` + `ListDispatches()` | Same detail level |
| `session-db.sh end` | `session.End()` | Same status transition |
| `session-db.sh dispatch` | `session.Dispatch()` | Same fields |
| `session-db.sh parallel` | `session.CreateParallelTask()` | Same wave semantics |
| `session-db.sh message` | `session.SendMessage()` | Same priority levels |
| `session-db.sh barrier` | `session.CreateBarrier()` | Same expected_count |
| `session-db.sh lock` | `session.AcquireLock()` | Same lock_name |

### Data Migration

Existing data in `.specify/session.db`:
1. Read existing `sessions` table
2. Verify schema matches V1 or V2
3. If V1: migrate to V2 (add new columns with defaults)
4. If V2: use as-is

---

## §A Acceptance Criteria

1. All bash commands have Go equivalents with identical behavior
2. `go test ./internal/runtime/session/...` passes with >90% coverage
3. Concurrent goroutines can create sessions and dispatches safely
4. Session ID format matches bash: `ses_<timestamp>`
5. Dispatch seq auto-increments correctly per session
6. Parent dispatch references are validated

---

## §N Next Steps

After this spec is approved:
1. Create `packages/cli/internal/runtime/session/` directory
2. Implement session.go (Task 1)
3. Implement dispatch.go (Task 2)
4. Implement parallel.go, message.go, barrier.go, lock.go (Tasks 3-6)
5. Write tests (Task 7)
6. PR review and merge

Then proceed to **Chunk 3: Task Queue**.

---

## §R References

- Bash session commands: `packages/cli/library/fortnite/scripts/session-db.sh` (lines 400-800)
- Bash dispatch logic: `packages/cli/library/fortnite/scripts/session-db.sh` (lines 1000-1200)
- Current Go session: `packages/cli/cmd/session.go`
- Spec 007 (Foundation): `specs/007-go-runtime-foundation/spec.md`
