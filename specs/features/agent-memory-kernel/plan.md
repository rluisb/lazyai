# Implementation Plan: Agent Memory Kernel — Phase 1

Parent ADR: [ADR-002](../../adrs/002-agent-memory-kernel.md)

ADR accepted. Phase 1 is bounded to deterministic continuity: resumability, handoff persistence, FTS5 lexical search, and a minimal context builder. No vector search, CLI commands, retention, or MCP exposure.

---

## Phase 1.1 — Module Scaffolding

1. Create `packages/agentmemory` directory.
2. Create `packages/agentmemory/go.mod` — module path `github.com/rluisb/lazyai/packages/agentmemory`.
3. Add to `go.work` — `use ./packages/agentmemory`.
4. Run `go mod tidy`.

## Phase 1.2 — Database and Migrations

1. Create `packages/agentmemory/db.go` — `Open(path string) (*DB, error)` using `modernc.org/sqlite`.
2. Create `packages/agentmemory/migrations.go` — hand-rolled migration engine matching orchestrator pattern:
   - Inline Go constants
   - `001_` prefix numbering
   - `Migration{ID, SQL}` struct
   - `RunMigrations()` applies pending migrations in order
   - Multi-statement wrapped in transactions
3. Migration `001_init.sql` — create tables:
   - `tasks` (id, namespace, project_root, task_type, state, current_step, state_json, goal, tags, created_at, updated_at)
   - `task_events` (id, task_id, namespace, run_id, event_type, payload_json, created_at)
   - `checkpoints` (id, task_id, namespace, step_id, summary, state_json, created_at)
   - `artifacts` (id, task_id, namespace, path, content_preview, size_bytes, content_hash, mime_type, tags, created_at)
   - `memories` (id, namespace, content, source_task_id, source_step_id, tags, importance, created_at)
4. Migration `002_fts.sql` — create FTS5 virtual tables for `memories` and `artifacts`.

## Phase 1.3 — Domain Types

Create `packages/agentmemory/domain.go` with types:
- `Task` — mirrors `tasks` table
- `TaskEvent` — mirrors `task_events`
- `Checkpoint` — mirrors `checkpoints`
- `Artifact` — mirrors `artifacts`
- `Memory` — mirrors `memories`
- Enums: `TaskType`, `TaskState`, `EventType`, `Importance`

## Phase 1.4 — Store Implementations

1. `packages/agentmemory/task_store.go` + tests:
   - `CreateTask`, `GetTask`, `UpdateTaskState`, `ListTasks(namespace)`
2. `packages/agentmemory/event_store.go` + tests:
   - `AppendEvent`, `ListEvents(taskID, since, limit)`, `RecentEvents(namespace, limit)`
3. `packages/agentmemory/checkpoint_store.go` + tests:
   - `CreateCheckpoint`, `LatestCheckpoint(taskID)`, `ListCheckpoints(taskID)`
4. `packages/agentmemory/artifact_store.go` + tests:
   - `RecordArtifact`, `ListArtifacts(taskID, namespace)`
5. `packages/agentmemory/memory_store.go` + tests:
   - `SaveMemory`, `SearchMemories(namespace, query)`, `ListMemories(namespace, tags, limit)`

## Phase 1.5 — FTS5 Lexical Search

1. Implement FTS5 search over `memories_fts` and `artifacts_fts`.
2. Escape FTS5 reserved characters in query input.
3. Include snippet extraction from matched content.
4. Tests for: exact match, prefix match, phrase match, empty query, no results.

## Phase 1.6 — Redaction Layer

1. `packages/agentmemory/redact.go` — `Redact(data []byte) []byte`.
2. Apply redaction to `content`, `state_json`, `payload_json`, `content_preview` before writes.
3. Patterns: `sk-*`, `Bearer *`, `Authorization: *`, `-----BEGIN * PRIVATE KEY-----`, env vars matching `*_KEY`, `*_SECRET`, `*_TOKEN`.
4. Tests for: API key redaction, private key redaction, env var redaction, JSON structure preservation, no-op on clean input.

## Phase 1.7 — Context Builder

1. `packages/agentmemory/context.go` — `BuildContext(ctx, taskID) (ContextBundle, error)`.
2. Assembled context: task goal/state + latest checkpoint summary + active handoffs + recent 20 events + top 5 FTS memories.
3. Tests for: full context assembly, empty task, no checkpoints, no events, no memories.

## Phase 1.8 — File Permissions and .gitignore

1. `packages/agentmemory/permissions.go` — `EnsurePermissions(dir string) error`.
2. Apply `0700` on data directory, `0600` on database file.
3. Auto-generate `.gitignore` in `.ai/agent-memory/`.
4. Tests: verify permissions are applied, verify `.gitignore` content.

## Phase 1.9 — Integration and Verification

1. Add `packages/agentmemory` to `go.work`.
2. Run `GOWORK=off go test ./... -count=1` from `packages/agentmemory`.
3. Run acceptance scenario: write task/checkpoint/memory, read them back, search via FTS5, build context.
4. Verify `CGO_ENABLED=0` clean build.

## Out of Scope (Phase 2+)

- CLI inspection commands
- Wiring into orchestrator run lifecycle
- Vector search / semantic recall
- Retention controls
- MCP tool exposure
- Embedding providers
- Workspace/user namespaces
- Compaction/summarization
