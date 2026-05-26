# Spec 007: Go Runtime Foundation

## Chunk 1: Unified Schema & Database Layer

**Status:** Draft  
**Author:** Ricardo Conceicao  
**Date:** 2026-05-26  
**Scope:** Foundation for all-Go runtime migration

---

## §G Goal

Replace the bash SQLite runtime (`session-db.sh`, `task-queue.sh`, etc.) with a unified Go database layer that:
1. Supports all 20+ tables from the bash V2 schema
2. Provides atomic transactions and connection pooling
3. Has versioned migrations (up/down)
4. Is tested with >90% coverage
5. Maintains backward compatibility with existing `.specify/session.db` files

---

## §C Constraints

1. **SQLite only** — No external DB dependencies
2. **Backward compatible** — Existing `.specify/session.db` must migrate cleanly
3. **Thread-safe** — All operations must be safe for concurrent agents
4. **Testable** — Every table must have unit tests
5. **No bash dependency** — Pure Go implementation

---

## §I Interfaces

### §I.1 Database Connection

```go
package runtime

// DB wraps sql.DB with runtime-specific operations
type DB struct {
    *sql.DB
    path string
}

// Open creates or opens a runtime database
func Open(path string) (*DB, error)

// Close closes the database connection
func (db *DB) Close() error

// WithTx executes fn within a transaction
func (db *DB) WithTx(fn func(*sql.Tx) error) error

// Migrate runs migrations up to target version
func (db *DB) Migrate(targetVersion int) error

// Version returns current schema version
func (db *DB) Version() (int, error)
```

### §I.2 Migration System

```go
package runtime

// Migration represents a single schema version
type Migration struct {
    Version int
    Up      string   // SQL to apply
    Down    string   // SQL to revert
    Name    string   // human-readable description
}

// Registry holds all migrations
type Registry struct {
    migrations []Migration
}

// Register adds a migration
func (r *Registry) Register(m Migration)

// Migrations returns ordered list
func (r *Registry) Migrations() []Migration
```

---

## §V Invariants

1. **Schema version monotonicity** — Version only increases, never decreases (except via explicit Down migration)
2. **Migration idempotency** — Running the same migration twice is a no-op
3. **Transaction atomicity** — Every migration runs in a single transaction
4. **FK enforcement** — Foreign keys are enabled and enforced
5. **WAL mode** — Write-ahead logging for concurrent reads

---

## §T Tasks

### Task 1: Schema Definition
**File:** `packages/cli/internal/runtime/schema.go`

Define all tables from bash V2 schema:

**Core Session Tables:**
- `sessions` — Session lifecycle
- `dispatches` — Agent dispatch audit log
- `decisions` — Decision records
- `artifacts` — File change tracking
- `memories` — Long-term memory vault
- `token_log` — Token usage telemetry
- `parallel_tasks` — Parallel execution tracking
- `messages` — Inter-agent messaging
- `barriers` — Synchronization barriers
- `locks` — Resource locks

**Workflow Tables:**
- `teams` — Agent team definitions
- `workflows` — Workflow definitions
- `workflow_instances` — Running workflow instances
- `workflow_steps` — Individual step execution

**V2 Observability Tables:**
- `model_calls` — LLM API telemetry
- `tool_calls` — Tool execution log
- `gate_results` — Gate outcomes
- `ledger_refs` — Ledger cross-references
- `cost_snapshots` — Cost aggregation
- `checkpoints` — Session checkpoints
- `eval_runs` — Evaluation runs
- `eval_results` — Evaluation results

**Task Queue Tables:**
- `task_queue` — Core task table
- `task_claims` — Agent claim tracking
- `task_dlq` — Dead letter queue
- `task_messages` — Task-scoped chat

**Indexes:**
All indexes from bash schema must be preserved for query performance.

### Task 2: Migration Framework
**File:** `packages/cli/internal/runtime/migrate.go`

Implement:
- Migration registry
- Version tracking table (`schema_migrations`)
- Up/Down execution
- Dry-run support
- Migration validation (checksums)

### Task 3: Connection Management
**File:** `packages/cli/internal/runtime/db.go`

Implement:
- `Open()` with WAL mode, FK enforcement, busy timeout
- Connection pooling (max 1 writer, N readers for SQLite)
- `WithTx()` helper
- Health check (`Ping`)
- Backup helper (`BackupTo()`)

### Task 4: Tests
**Files:** `packages/cli/internal/runtime/*_test.go`

Required tests:
- `TestOpenCreatesDatabase` — DB creation
- `TestMigrateUpDown` — Migration round-trip
- `TestMigrationIdempotency` — Running twice is no-op
- `TestForeignKeyEnforcement` — FK constraints work
- `TestConcurrentAccess` — Multiple goroutines safe
- `TestBackupRestore` — Backup/restore cycle
- Table-specific tests for each of 20+ tables (CRUD)

---

## §B Backward Compatibility

### Migration Path from Bash DB

1. Detect existing `.specify/session.db`
2. Read current bash schema version (from `PRAGMA user_version`)
3. Map bash version to Go migration version
4. Run remaining migrations
5. Verify data integrity

### Bash Version Mapping

| Bash Version | Go Migration | Notes |
|-------------|-------------|-------|
| 0 (no version) | 1-10 | Full migration |
| 1 (V1 schema) | 11-20 | V1 → V2 upgrade |
| 2 (V2 schema) | 21+ | Future migrations |

---

## §A Acceptance Criteria

1. `go test ./internal/runtime/...` passes with >90% coverage
2. Existing `.specify/session.db` opens without error
3. All 20+ tables are queryable
4. Migrations run in <5 seconds for empty DB
5. Concurrent goroutines can read/write without `database is locked`
6. Backup/restore produces identical database

---

## §N Next Steps

After this spec is approved:
1. Create `packages/cli/internal/runtime/` directory
2. Implement schema.go (Task 1)
3. Implement migrate.go (Task 2)
4. Implement db.go (Task 3)
5. Write tests (Task 4)
6. PR review and merge

Then proceed to **Chunk 2: Session Management**.

---

## §R References

- Bash schema source: `packages/cli/library/fortnite/scripts/session-db.sh` (lines 200-400)
- Bash migration logic: `packages/cli/library/fortnite/scripts/session-db.sh` (lines 50-100)
- Current Go DB: `packages/cli/internal/db/db.go`
- Current Go migrations: `packages/cli/internal/db/migrations.go`
