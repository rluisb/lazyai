# Spec 009: Go Runtime — Task Queue

## Chunk 3: Atomic Claiming, Dead Letter Queue, Zombie Sweep

**Status:** Draft  
**Author:** Ricardo Conceicao  
**Date:** 2026-05-26  
**Depends on:** Spec 007 (Foundation)

---

## §G Goal

Implement the task queue layer that replaces `task-queue.sh`:
- `enqueue` — Add a task to the queue
- `claim` — Atomically claim a task (FIFO, with parallelism limit)
- `complete` — Mark a task as completed
- `fail` — Move a task to the dead letter queue
- `sweep` — Remove stale claims from crashed agents
- `list` — List tasks in the queue
- `chat` — Task-scoped messaging between agents

---

## §C Constraints

1. **Depends on Spec 007** — Requires `runtime.DB` and schema
2. **Atomic claiming** — Must use CTE with RETURNING (SQLite supports this)
3. **FIFO ordering** — Tasks claimed in creation order
4. **Parallelism limit** — `max_agents` controls concurrent claims
5. **Dedupe** — Same `dedupe_key` prevents duplicate active tasks

---

## §I Interfaces

### §I.1 Task Queue Manager

```go
package taskqueue

import "github.com/rluisb/lazyai/packages/cli/internal/runtime"

// Manager handles task queue operations
type Manager struct {
    db *runtime.DB
}

// NewManager creates a task queue manager
func NewManager(db *runtime.DB) *Manager

// Enqueue adds a task to the queue
func (m *Manager) Enqueue(sessionID string, topic string, task string, opts EnqueueOptions) (*Task, error)

// Claim atomically claims an available task
func (m *Manager) Claim(sessionID string, topic string, agent string) (*Task, error)

// Complete marks a task as completed
func (m *Manager) Complete(taskID int, agent string) error

// Fail moves a task to the DLQ
func (m *Manager) Fail(taskID int, agent string, errorMessage string, contextDump string) error

// Sweep removes stale claims
func (m *Manager) Sweep(timeoutSeconds int) (int, error)

// List returns tasks in the queue
func (m *Manager) List(sessionID string, status string) ([]Task, error)
```

### §I.2 Data Types

```go
// Task represents a task in the queue
type Task struct {
    ID         int
    SessionID  string
    Topic      string
    Task       string
    Status     TaskStatus
    MaxAgents  int
    DedupeKey  *string
    CreatedAt  time.Time
    Claims     []TaskClaim
}

// TaskStatus enum
type TaskStatus string
const (
    TaskOpen      TaskStatus = "open"
    TaskClaimed   TaskStatus = "claimed"
    TaskCompleted TaskStatus = "completed"
    TaskFailed    TaskStatus = "failed"
)

// TaskClaim represents an agent's claim on a task
type TaskClaim struct {
    TaskID    int
    Agent     string
    ClaimedAt time.Time
}

// DLQEntry represents a dead letter queue entry
type DLQEntry struct {
    ID            int
    TaskID        int
    FailedAgent   string
    ErrorMessage  string
    ContextDump   string
    FailedAt      time.Time
}

// TaskMessage represents a task-scoped message
type TaskMessage struct {
    ID         int
    TaskID     int
    FromAgent  string
    ToAgent    *string
    Body       string
    CreatedAt  time.Time
}
```

---

## §V Invariants

1. **Atomic claiming** — Only one agent can claim a task at a time (CTE + RETURNING)
2. **FIFO ordering** — Tasks claimed in `created_at` order
3. **Parallelism limit** — No more than `max_agents` claims per task
4. **Dedupe uniqueness** — Only one active task per `dedupe_key`
5. **Claim cleanup** — Failed/completed tasks have claims removed

---

## §T Tasks

### Task 1: Enqueue
**File:** `packages/cli/internal/runtime/taskqueue/taskqueue.go`

Implement:
- `Enqueue()` — Insert into `task_queue`
- Validate dedupe_key uniqueness (partial unique index handles this)
- Return the created task

### Task 2: Atomic Claim
**File:** `packages/cli/internal/runtime/taskqueue/claim.go`

Implement CTE-based atomic claim:

```sql
WITH available_task AS (
    SELECT q.id
    FROM task_queue q
    LEFT JOIN task_claims c ON q.id = c.task_id
    WHERE q.topic = ? AND q.status = 'open'
      AND q.id NOT IN (SELECT task_id FROM task_claims WHERE agent = ?)
    GROUP BY q.id
    HAVING COUNT(c.agent) < q.max_agents
    ORDER BY q.created_at ASC
    LIMIT 1
)
INSERT INTO task_claims (task_id, agent, claimed_at)
SELECT id, ?, ? FROM available_task
RETURNING task_id;
```

- Retry logic: 3 attempts with 1s backoff
- Handle `database is locked` (SQLite busy)
- Return claimed task or nil if none available

### Task 3: Complete and Fail
**File:** `packages/cli/internal/runtime/taskqueue/lifecycle.go`

Implement:
- `Complete()` — Update status to 'completed', remove claim
- `Fail()` — Update status to 'failed', remove claim, insert into `task_dlq`
- Both in transactions

### Task 4: Zombie Sweep
**File:** `packages/cli/internal/runtime/taskqueue/sweep.go`

Implement:
- `Sweep()` — Delete stale claims where `claimed_at < now - timeout`
- Only sweep claims for tasks still 'open'
- Return count of removed claims

### Task 5: Task-Scoped Messaging
**File:** `packages/cli/internal/runtime/taskqueue/message.go`

Implement:
- `SendMessage()` — Insert into `task_messages`
- `GetMessages()` — Select by task_id, cursor-based (id > last_seen)
- `GetUnreadCount()` — Count messages for agent

### Task 6: Tests
**Files:** `packages/cli/internal/runtime/taskqueue/*_test.go`

Required tests:
- `TestEnqueueAndClaim` — Basic enqueue + claim
- `TestAtomicClaim` — Concurrent claims (race test)
- `TestFIFOOrdering` — Verify claim order
- `TestParallelismLimit` — max_agents enforcement
- `TestDedupe` — Duplicate prevention
- `TestCompleteAndFail` — Lifecycle transitions
- `TestZombieSweep` — Stale claim removal
- `TestMessaging` — Send, receive, cursor

---

## §B Backward Compatibility

### Bash Command Mapping

| Bash Command | Go Method | Notes |
|-------------|-----------|-------|
| `task-queue.sh enqueue` | `taskqueue.Enqueue()` | Same fields |
| `task-queue.sh claim` | `taskqueue.Claim()` | Same CTE logic |
| `task-queue.sh complete` | `taskqueue.Complete()` | Same status |
| `task-queue.sh fail` | `taskqueue.Fail()` | Same DLQ insert |
| `task-queue.sh sweep` | `taskqueue.Sweep()` | Same timeout logic |
| `task-queue.sh list` | `taskqueue.List()` | Same columns |
| `task-queue.sh chat` | `taskqueue.SendMessage()` | Same scoped messaging |

---

## §A Acceptance Criteria

1. `go test ./internal/runtime/taskqueue/...` passes with >90% coverage
2. Concurrent goroutines can claim tasks without race conditions
3. FIFO ordering is preserved under load
4. max_agents limit is enforced
5. Dedupe prevents duplicate active tasks
6. Zombie sweep removes stale claims

---

## §N Next Steps

After this spec is approved:
1. Create `packages/cli/internal/runtime/taskqueue/` directory
2. Implement taskqueue.go (Task 1)
3. Implement claim.go (Task 2)
4. Implement lifecycle.go, sweep.go, message.go (Tasks 3-5)
5. Write tests (Task 6)
6. PR review and merge

Then proceed to **Chunk 4: Workflow Engine**.

---

## §R References

- Bash task queue: `packages/cli/library/fortnite/scripts/task-queue.sh`
- Current Go task: `packages/cli/cmd/task.go`
- Spec 007 (Foundation): `specs/007-go-runtime-foundation/spec.md`
