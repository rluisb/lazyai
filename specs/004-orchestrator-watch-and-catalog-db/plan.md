# 004 — Orchestrator Watch & Catalog DB: Plan

## Date
2026-04-20

## Context

See `spec.md` for goals, scope, and decisions. This document sequences the work into delivery phases. Each phase is a separate PR with its own tests and no cross-phase mixing.

## Phasing principles

- **Test contracts are sacred.** All current `orchestrator/src/__tests__/*.case.ts` files must continue to pass without modification through phases 0–3. New test files cover new behavior.
- **One seam per phase.** `persistence.ts` is replaced in phase 2; `loader.ts` extended in phase 3; transport added in phase 4. No phase touches more than one architectural seam.
- **Behavior change is opt-in until phase 5.** Phases 0–3 are observable only via the new tools; existing chains/teams/workflows behave identically.

## Phase 0 — Scaffolding (no behavior change)

Goal: lay down the DB module, migration framework, and config-resolution utilities. No existing code paths change.

Files created:
- `orchestrator/src/db/index.ts` — `openDatabase(path)` returning a configured `better-sqlite3` instance (WAL, foreign keys on, busy timeout).
- `orchestrator/src/db/migrations.ts` — runner that reads `migrations/*.sql` in order, records applied migrations in `schema_migrations`.
- `orchestrator/src/db/migrations/0001_init.sql` — empty `schema_migrations` table only; real schemas land in later phases.
- `orchestrator/src/config/paths.ts` — XDG-aware path resolution (`getDataDir()`, `getStateDir()`, `getDatabasePath()`, `getSocketPath()`, `getLogDir()`).
- `orchestrator/src/__tests__/db.case.ts` — opens an in-`:memory:` DB, runs migrations, asserts tracking row.

Files modified:
- `orchestrator/package.json` — add `better-sqlite3` dependency.

Acceptance: `npm test` passes including the new `db.case.ts`. No existing test changes.

## Phase 1 — Logging + tail CLI

Goal: structured JSONL logging to disk + a separate-terminal tail subcommand. Independent of DB work; can land in parallel with phase 0.

Files created:
- `orchestrator/src/logging/logger.ts` — minimal Pino-compatible logger emitting JSONL with fields `{ts, level, msg, runId?, stepId?, kind?, ...}`.
- `orchestrator/src/logging/sink.ts` — file rotation by date; writes to `~/.local/state/ai-setup-orchestrator/logs/orchestrator-YYYY-MM-DD.log`.
- `orchestrator/src/cli/tail.ts` — `tail` subcommand; opens today's log, follows growth, supports `--filter run-id=...` and `--filter level=warn`.
- `orchestrator/src/__tests__/logging.case.ts` — round-trip a few log entries; assert filtering.

Files modified:
- `orchestrator/src/index.ts` — instantiate logger; route MCP request/response and machine transitions through it.
- `orchestrator/src/tool-handlers.ts` — log tool entry/exit at info, errors at warn/error.
- `orchestrator/src/chain-machine.ts`, `team-machine.ts`, `workflow-machine.ts` — log state transitions at debug.
- `orchestrator/package.json` — add `bin.ai-setup-orchestrator-tail` shim or wire into existing CLI router.

Acceptance: running a chain produces a readable log file; `ai-setup-orchestrator tail` in a second terminal streams entries live.

## Phase 2 — SQLite-backed `persistence.ts`

Goal: replace JSON-file persistence with SQLite, preserving the existing `persistence.ts` interface so all callers and all `*.case.ts` tests continue to work unchanged.

Schema added (`migrations/0002_run_state.sql`):

```sql
CREATE TABLE execution_plans (
  id TEXT PRIMARY KEY,
  kind TEXT NOT NULL CHECK (kind IN ('chain','team','workflow')),
  definition_name TEXT NOT NULL,
  definition_version TEXT,
  project_root TEXT,
  plan_json TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE chain_runs (
  id TEXT PRIMARY KEY,
  definition_name TEXT NOT NULL,
  definition_version TEXT,
  state TEXT NOT NULL,
  current_step_id TEXT,
  project_root TEXT,
  steps_json TEXT NOT NULL,
  budget_json TEXT NOT NULL,
  meta_json TEXT NOT NULL,         -- entryStepId, completedStepIds, handoffPath, etc.
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE team_runs (
  id TEXT PRIMARY KEY,
  definition_name TEXT NOT NULL,
  definition_version TEXT,
  state TEXT NOT NULL,
  project_root TEXT,
  tasks_json TEXT NOT NULL,
  budget_json TEXT NOT NULL,
  meta_json TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE workflow_runs (
  id TEXT PRIMARY KEY,
  definition_name TEXT NOT NULL,
  definition_version TEXT,
  state TEXT NOT NULL,
  current_phase_id TEXT,
  project_root TEXT,
  phases_json TEXT NOT NULL,
  child_runs_json TEXT NOT NULL,
  budget_json TEXT NOT NULL,
  runtime_json TEXT NOT NULL,
  meta_json TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE handoffs (
  id TEXT PRIMARY KEY,
  run_id TEXT NOT NULL,
  run_kind TEXT NOT NULL,
  doc_json TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE error_journal (
  id TEXT PRIMARY KEY,
  run_id TEXT,
  run_kind TEXT,
  definition_name TEXT,
  step_id TEXT,
  category TEXT NOT NULL,
  code TEXT NOT NULL,
  message TEXT NOT NULL,
  context_json TEXT NOT NULL,
  resolution_json TEXT,
  lesson_json TEXT,
  created_at TEXT NOT NULL
);

CREATE INDEX idx_chain_runs_state ON chain_runs(state);
CREATE INDEX idx_chain_runs_updated ON chain_runs(updated_at);
CREATE INDEX idx_error_journal_run ON error_journal(run_id);
```

Files modified:
- `orchestrator/src/persistence.ts` — internals rewritten to use prepared statements against SQLite. Public function signatures unchanged.

Files created:
- `orchestrator/src/persistence/importer.ts` — one-shot scan of `<project>/.ai/orchestration/state/` JSON files; insert as rows if missing. Idempotent.
- `orchestrator/src/__tests__/persistence-sqlite.case.ts` — round-trip tests against `:memory:` DB; asserts importer correctly migrates a sample on-disk state tree.

Acceptance: all existing `persistence.case.ts` tests pass against the new backend; importer test passes; first run on an existing project picks up prior state.

## Phase 3 — Catalog DB + versioning + resolution + management tools

Goal: introduce internal versioned catalog, host-config resolver, and management surface (CLI + MCP).

Schema added (`migrations/0003_catalog.sql`):

```sql
CREATE TABLE definitions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  kind TEXT NOT NULL CHECK (kind IN ('agent','skill','chain','team','workflow','mode','command')),
  name TEXT NOT NULL,
  active_version_id INTEGER,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(kind, name)
);

CREATE TABLE definition_versions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  definition_id INTEGER NOT NULL REFERENCES definitions(id),
  version INTEGER NOT NULL,
  frontmatter_json TEXT NOT NULL,
  body TEXT NOT NULL,
  checksum TEXT NOT NULL,
  created_at TEXT NOT NULL,
  created_by TEXT,
  UNIQUE(definition_id, version)
);

CREATE INDEX idx_def_versions_checksum ON definition_versions(checksum);
```

Files created:
- `orchestrator/src/catalog/schemas.ts` — Zod frontmatter schemas per kind (agent, skill, chain, team, workflow, mode, command).
- `orchestrator/src/catalog/store.ts` — DB-backed CRUD + version operations (create-version, set-active, list, show, diff).
- `orchestrator/src/catalog/resolver.ts` — implements the 4-step resolution priority (explicit pin → project host → user-global host → internal active).
- `orchestrator/src/catalog/host-scanner.ts` — opencode + claude-code scanners with mtime-based memoization.
- `orchestrator/src/catalog/importer.ts` — bulk import from `~/.config/opencode/`, `~/.claude/`, and the existing `library/` path; checksum dedup.
- `orchestrator/src/cli/catalog.ts` — `catalog` subcommand router (`list`, `show`, `diff`, `create-version`, `set-active`, `import-from-files`, `export-version`).
- `orchestrator/src/catalog-tools.ts` — MCP tools mirroring the CLI subcommands.
- `orchestrator/src/__tests__/catalog-store.case.ts`, `catalog-resolver.case.ts`, `host-scanner.case.ts`.

Files modified:
- `orchestrator/src/loader.ts` — when DB is available, prefer DB internal catalog; fall back to file library if DB empty (allows first-run import to succeed).
- `orchestrator/src/compiler.ts` — when looking up an agent or skill during compilation, route through `resolver.ts` instead of the in-memory catalog map directly.
- `orchestrator/src/server.ts` — register the catalog MCP tools.

Acceptance: round-trip create→show→diff→set-active works; resolution returns user file when present and DB version when not; `loader.case.ts` still passes (resolver behaves identically to the old in-memory map when neither host file nor DB row exists, falling through to the file library).

## Phase 4 — HTTP/SSE transport over Unix socket

Goal: opt-in shared-process transport so multiple clients can interact with one orchestrator.

Files created:
- `orchestrator/src/transport/http-server.ts` — small HTTP+SSE server bound to `~/.local/share/ai-setup-orchestrator/orchestrator.sock` (mode 0600); MCP framing over SSE.
- `orchestrator/src/cli/serve.ts` — `serve` subcommand that starts the HTTP server (default behavior remains stdio).
- `orchestrator/src/__tests__/http-transport.case.ts` — connect over the socket, call a tool, receive a response.

Files modified:
- `orchestrator/src/index.ts` — `--transport stdio|http` flag; default `stdio`.

Acceptance: a smoke test client can talk to the orchestrator over the socket and call `list_catalog`.

## Phase 5 — Watch / subscribe

Goal: live event stream for chain/team/workflow transitions; MCP notifications over the HTTP transport.

Schema added (`migrations/0004_events.sql`):

```sql
CREATE TABLE events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  topic TEXT NOT NULL,
  run_id TEXT,
  kind TEXT NOT NULL,
  payload_json TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE INDEX idx_events_topic ON events(topic, id);
CREATE INDEX idx_events_run ON events(run_id);
```

Files created:
- `orchestrator/src/events/bus.ts` — in-process `EventEmitter` wrapper with topic semantics; writes every event to `events` table.
- `orchestrator/src/events/notifications.ts` — converts bus events into MCP notifications.
- `orchestrator/src/tools/subscribe.ts` — `subscribe(topic, sinceEventId?)` MCP tool; replays from event id then attaches live.
- `orchestrator/src/__tests__/event-bus.case.ts`, `subscribe.case.ts`.

Files modified:
- `orchestrator/src/chain-machine.ts`, `team-machine.ts`, `workflow-machine.ts` — emit transition events via the bus.
- `orchestrator/src/persistence.ts` — emit `*_run.updated` events after each save.

Acceptance: client A starts a chain; client B subscribed to `chain:<id>` receives every transition + the gate prompt + the completion event.

## Phase 6 — Triggerable agent registry

Goal: any client can invoke a named agent at a pinned version with full observability.

Files created:
- `orchestrator/src/tools/agent-invoke.ts` — `agent_invoke(name, version?, prompt, context?)` MCP tool. Composes via existing composer; runs as a single-step chain so budget/journal/events all apply.
- `orchestrator/src/cli/agent.ts` — `agent invoke <name> [--version N] --prompt ...` subcommand.
- `orchestrator/src/__tests__/agent-invoke.case.ts`.

Acceptance: invocation produces a chain run visible via `subscribe`, persisted in DB, with composed prompt referencing the pinned version's body.

## Phase 7 — Message queue

Goal: durable scheduling so retries, gates, and time-delayed transitions survive restarts.

Schema added (`migrations/0005_queue.sql`):

```sql
CREATE TABLE queue_jobs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  topic TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('pending','running','done','failed','dead')),
  payload_json TEXT NOT NULL,
  attempts INTEGER NOT NULL DEFAULT 0,
  max_attempts INTEGER NOT NULL DEFAULT 3,
  next_run_at TEXT NOT NULL,
  last_error TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE INDEX idx_queue_jobs_pending ON queue_jobs(status, next_run_at);
```

Files created:
- `orchestrator/src/queue/queue.ts` — in-memory job queue with SQLite checkpointing.
- `orchestrator/src/queue/workers.ts` — workers for chain step continuation, retry-with-backoff, gate-resume.
- `orchestrator/src/__tests__/queue.case.ts`.

Files modified:
- `orchestrator/src/chain-machine.ts` — retries and time-delayed transitions enqueue jobs instead of resolving inline.
- `orchestrator/src/index.ts` — boot reclaims in-flight jobs (`status='running'` → `pending`).

Acceptance: a chain with a 5s retry survives an orchestrator restart and completes the retry on schedule.

## Cross-phase concerns

- **Test isolation**: every DB-touching test opens its own `:memory:` DB; no shared state.
- **Migration ordering**: each phase appends to `migrations/`; never edit a previously-shipped migration.
- **Telemetry**: every phase adds to the logger before adding to the events table; logs are the user-visible diagnostic surface, events are the programmatic one.
- **Backwards compatibility**: stdio transport, existing tool names, and existing tool input schemas are preserved through phase 7. New tools use distinct names (`subscribe`, `agent_invoke`, `catalog_*`).

## Out of scope for this plan

- Multi-machine sync.
- Distributed locking.
- A web UI for catalog management (CLI + MCP only).
- Migration to Go (reconfirmed: TS handles projected concurrency).
