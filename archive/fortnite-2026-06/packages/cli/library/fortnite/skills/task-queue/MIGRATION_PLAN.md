# Task Queue Migration Plan — Phase 1C

**Status**: Planning Complete ✅
**Phase**: 1C (Documentation Only)
**Target**: shmig-backed schema migration system for SQLite task queue and workflow/session DB tables

---

## 1. Current Schema Snapshot

### Core Tables

```sql
-- task_queue: Primary task storage
CREATE TABLE IF NOT EXISTS task_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    topic TEXT NOT NULL,
    task TEXT NOT NULL,
    status TEXT DEFAULT 'open',
    max_agents INTEGER NOT NULL DEFAULT 1,
    dedupe_key TEXT,
    created_at TEXT NOT NULL
);

-- task_claims: Agent claims on tasks (composite PK)
CREATE TABLE IF NOT EXISTS task_claims (
    task_id INTEGER NOT NULL REFERENCES task_queue(id),
    agent TEXT NOT NULL,
    claimed_at TEXT NOT NULL,
    PRIMARY KEY (task_id, agent)
);

-- task_dlq: Dead-letter queue for failed tasks
CREATE TABLE IF NOT EXISTS task_dlq (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER REFERENCES task_queue(id),
    failed_agent TEXT,
    error_message TEXT,
    context_dump TEXT,
    failed_at TEXT NOT NULL
);

-- task_messages: Task-scoped collaboration
CREATE TABLE IF NOT EXISTS task_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL REFERENCES task_queue(id),
    from_agent TEXT NOT NULL,
    to_agent TEXT,
    body TEXT NOT NULL,
    created_at TEXT NOT NULL
);
```

### Indexes

```sql
-- Standard indexes
CREATE INDEX IF NOT EXISTS idx_task_queue_session ON task_queue(session_id);
CREATE INDEX IF NOT EXISTS idx_task_queue_topic ON task_queue(topic);
CREATE INDEX IF NOT EXISTS idx_task_queue_status ON task_queue(status);
CREATE INDEX IF NOT EXISTS idx_task_claims_task ON task_claims(task_id);
CREATE INDEX IF NOT EXISTS idx_task_claims_agent ON task_claims(agent);
CREATE INDEX IF NOT EXISTS idx_task_dlq_task ON task_dlq(task_id);
CREATE INDEX IF NOT EXISTS idx_task_messages_task ON task_messages(task_id);

-- Partial unique index for active dedupe keys
CREATE UNIQUE INDEX IF NOT EXISTS idx_task_queue_active_dedupe
ON task_queue(dedupe_key)
WHERE dedupe_key IS NOT NULL AND status IN ('open','claimed');
```

### Workflow Integration

**Dedupe Key Format**: `workflow:<workflow_instance_id>:step:<step_order>`

**Example**: `workflow:wi_abc123:step:1`

This ensures idempotent enqueueing of workflow steps.

---

## 2. Recommended Migration Strategy

### shmig Integration

**Wrapper Script**: `scripts/db-migrate.sh`

```bash
#!/usr/bin/env bash
# db-migrate.sh — shmig wrapper for SQLite migrations
# Usage: ./db-migrate.sh [migrate|rollback|status|init]

set -euo pipefail

DB_PATH="${DB_PATH:-${OPENCODE_WORKSPACE:-.}/.specify/session.db}"
MIGRATIONS_DIR="${MIGRATIONS_DIR:-$(dirname "$0")/migrations}"
SHMIG="${SHMIG:-shmig}"

case "${1:-migrate}" in
    init)
        mkdir -p "$MIGRATIONS_DIR"
        echo "✅ Migrations directory initialized: $MIGRATIONS_DIR"
        ;;
    migrate)
        $SHMIG -t sqlite3 -d "$DB_PATH" -m "$MIGRATIONS_DIR" migrate
        ;;
    rollback)
        $SHMIG -t sqlite3 -d "$DB_PATH" -m "$MIGRATIONS_DIR" rollback
        ;;
    status)
        $SHMIG -t sqlite3 -d "$DB_PATH" -m "$MIGRATIONS_DIR" status
        ;;
    *)
        echo "Usage: db-migrate.sh [migrate|rollback|status|init]"
        exit 1
        ;;
esac
```

### Migration Directory Structure

```
migrations/
├── 1654321000_init_task_queue_tables.sql
├── 1654321001_add_dedupe_key_column.sql
├── 1654321002_create_active_dedupe_index.sql
└── 1654321003_add_workflow_integration.sql
```

### Migration Naming Convention

- **Format**: `<UNIX_TIMESTAMP>-description.sql` where timestamp is epoch time
- **Order**: Applied by timestamp (epoch order)
- **Content**: SQL with shmig section markers (comments allowed)

### Migration File Format

All migration files must include shmig section markers:

```sql
-- ====  UP  ====
CREATE TABLE task_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    topic TEXT NOT NULL,
    task TEXT NOT NULL,
    status TEXT DEFAULT 'open',
    max_agents INTEGER NOT NULL DEFAULT 1,
    dedupe_key TEXT,
    created_at TEXT NOT NULL
);

-- ==== DOWN ====
DROP TABLE IF EXISTS task_queue;
```

### shmig Version Tracking

**Option A**: Use shmig's default `shmig_version` table
**Option B**: Extend migration metadata separately if additional audit fields are needed

```sql
-- Option B: Custom version table (if needed)
CREATE TABLE IF NOT EXISTS shmig_version (
    version INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL,
    description TEXT,
    checksum TEXT
);
```

### Integration Plan

1. **session-db.sh init**: Call `db-migrate.sh migrate` after creating the DB
2. **task-queue.sh init**: Remove inline DDL, call `db-migrate.sh migrate` instead
3. **workflow-run.sh sync**: Add migration check before workflow operations

---

## 3. Phase Boundaries

### Phase 1C (Current)
- ✅ Documentation and planning only
- ✅ No code changes
- ✅ No automatic migration of current runtime DB
- ✅ No commits/pushes

### Phase 2 (Future)
- Create `migrations/` directory with initial schema
- Implement `db-migrate.sh` wrapper
- Update `task-queue.sh init` to use migrations
- Add migration checks to `session-db.sh` and `workflow-run.sh`
- Test migration paths

**Approval Required**: Phase 2 work deferred until explicit user approval.

---

## 4. Rollback Strategy

### SQLite DDL Limitations

- **No ALTER COLUMN**: SQLite doesn't support modifying columns
- **No DROP COLUMN**: SQLite doesn't support dropping columns
- **No Rename Table**: Use `ALTER TABLE RENAME TO` instead
- **No Transactional DDL**: Each migration runs in its own transaction

### Rollback Approach

1. **Backup first**: `cp .specify/session.db .specify/session.db.backup.$(date +%s)`
2. **Down migrations**: Use `shmig rollback` to revert migrations in reverse order
3. **Manual intervention**: For breaking changes, restore from backup

---

## 5. Verification Plan

### Test Scenarios

1. **Fresh DB Migration**
   ```bash
   rm -f .specify/session.db
   ./scripts/db-migrate.sh migrate
   ./scripts/task-queue.sh init
   ./scripts/task-queue.sh add ses_test "test" "test task"
   ```

2. **Existing DB Migration (without dedupe_key)**
   ```bash
   # Create DB with old schema (no dedupe_key)
   sqlite3 .specify/session.db "CREATE TABLE task_queue (id INTEGER PRIMARY KEY, session_id TEXT, topic TEXT, task TEXT, status TEXT);"

   # Run migration
   ./scripts/db-migrate.sh migrate

   # Verify dedupe_key column exists
   sqlite3 .specify/session.db "PRAGMA table_info(task_queue);"
   ```

3. **Idempotency Test**
   ```bash
   ./scripts/db-migrate.sh migrate
   ./scripts/db-migrate.sh migrate  # Should be idempotent
   ./scripts/db-migrate.sh status
   ```

4. **Task Queue E2E After Migration**
   ```bash
   # Enqueue, claim, complete cycle
   TASK_ID=$(./scripts/task-queue.sh add ses_test "auth" "test task" 1 "test-dedupe" | grep -oP 'id=\K\d+')
   ./scripts/task-queue.sh claim "auth" "wall-builder"
   ./scripts/task-queue.sh complete "$TASK_ID"
   ```

5. **Workflow Enqueue Dedupe After Migration**
   ```bash
   # Enqueue same workflow step twice (should be idempotent)
   skills/workflow-engine/scripts/workflow-exec.sh enqueue-step wi_test123
   skills/workflow-engine/scripts/workflow-exec.sh enqueue-step wi_test123
   # Should see: "⚠️  Task already queued: id=... dedupe_key=workflow:wi_test123:step:1"
   ```

---

## 6. Risks and Mitigations

### Schema Drift

**Risk**: Inline DDL in scripts vs migration files get out of sync
**Mitigation**: Remove all inline DDL from scripts, use migrations exclusively

### Runtime DB Preservation

**Risk**: Accidental migration of production runtime DB
**Mitigation**: Backup before migration, test on copies first

### Partial Indexes Compatibility

**Risk**: SQLite version compatibility with partial unique indexes
**Mitigation**: Feature-detect in migration, provide fallback

### Bash Portability

**Risk**: macOS vs Linux shell differences
**Mitigation**: Use `#!/usr/bin/env bash`, test on both platforms

---

## 7. Explicit Non-Goals

- ❌ No code changes in Phase 1C
- ❌ No automatic migration of current dirty runtime DB
- ❌ No commits/pushes in Phase 1C
- ❌ No production deployment
- ❌ No breaking changes to existing scripts

---

## 8. Next Steps

### For Phase 2 Approval

```bash
# Create migrations directory
mkdir -p skills/task-queue/migrations

# Create initial migration files
cat > skills/task-queue/migrations/$(date +%s)_init_task_queue_tables.sql << 'SQL'
-- ====  UP  ====
CREATE TABLE task_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    topic TEXT NOT NULL,
    task TEXT NOT NULL,
    status TEXT DEFAULT 'open',
    max_agents INTEGER NOT NULL DEFAULT 1,
    dedupe_key TEXT,
    created_at TEXT NOT NULL
);

CREATE TABLE task_claims (
    task_id INTEGER NOT NULL REFERENCES task_queue(id),
    agent TEXT NOT NULL,
    claimed_at TEXT NOT NULL,
    PRIMARY KEY (task_id, agent)
);

CREATE TABLE task_dlq (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER REFERENCES task_queue(id),
    failed_agent TEXT,
    error_message TEXT,
    context_dump TEXT,
    failed_at TEXT NOT NULL
);

CREATE TABLE task_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL REFERENCES task_queue(id),
    from_agent TEXT NOT NULL,
    to_agent TEXT,
    body TEXT NOT NULL,
    created_at TEXT NOT NULL
);

-- ==== DOWN ====
DROP TABLE IF EXISTS task_messages;
DROP TABLE IF EXISTS task_dlq;
DROP TABLE IF EXISTS task_claims;
DROP TABLE IF EXISTS task_queue;
SQL

# Continue with remaining migrations...
```

---

## Checklist

- [x] Document current schema
- [x] Document dedupe key format
- [x] Propose shmig integration strategy
- [x] Define migration directory structure
- [x] Outline verification plan
- [x] Identify risks and mitigations
- [x] Define phase boundaries
- [x] List explicit non-goals

---

**Created**: 2026-05-20
**Status**: Ready for review ✅
