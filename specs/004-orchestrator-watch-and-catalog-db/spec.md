# 004 — Orchestrator Watch & Catalog DB: Spec

## Date
2026-04-20

## Context

The orchestrator MCP (`@ai-setup/orchestrator`) currently:

- Loads agent/skill/chain/team/workflow definitions from disk on startup (file-based catalog).
- Persists run state as JSON files under `.ai/orchestration/state/{chains,teams,workflows,plans,handoffs}/`.
- Appends errors to `error-journal.jsonl`.
- Communicates over stdio, one process per client.
- Has no logs, no live event stream, and no way for one CLI instance to observe what another is doing.

For multi-client workflows (running 1 Claude Code + 1 OpenCode + 1 Codex against the same project, or several Claude Code instances against different projects), this model breaks down:

- No shared visibility — each client only sees its own runs.
- No way to watch a chain from a separate terminal.
- Catalog edits require touching files; no audit trail, no rollback.
- User-installed agents/skills (in `~/.config/opencode/`, `~/.claude/`, etc.) are invisible to the orchestrator unless symlinked into its library path.

## Goals

1. **Single source of truth** for orchestrator state via SQLite, while keeping host-CLI agent/skill files sacred (read-only, never written).
2. **BYO agents and skills** — when a chain references `agent: reviewer` and the user has `~/.config/opencode/agents/reviewer.md`, use theirs; otherwise fall back to the orchestrator's internal versioned catalog.
3. **Versioned internal catalog** — every internal agent/skill/chain/team/workflow/mode change creates a new immutable version row. Active version pointer moves; nothing is overwritten.
4. **Observability without coupling** — structured JSONL logs viewable by `ai-setup-orchestrator tail` from a separate terminal, independent of any MCP client.
5. **Live state across clients** — chain runs, gate prompts, budget warnings, and step transitions broadcast to subscribed clients via MCP notifications over a shared HTTP/SSE transport (stdio remains supported for single-client mode).
6. **Triggerable agents** — any client can invoke a named agent at a pinned version; invocation is recorded so other clients can observe.
7. **Durable scheduling** — in-memory message queue with SQLite-backed checkpoints so retries and gates survive restarts.

## Non-goals

- Multi-user / network-exposed deployment. The HTTP transport binds to a Unix socket; localhost only.
- Cross-machine sync. One DB per machine.
- Replacing host CLI catalogs. The orchestrator never writes to `~/.config/opencode/`, `~/.claude/`, or `~/.codex/` unless the user explicitly runs an export command targeting that path.
- Rewriting in Go. TypeScript handles the projected concurrency (small number of CLI clients) comfortably.

## Scope

### Storage

- SQLite via `better-sqlite3` (sync API, WAL mode).
- Cross-project DB at `~/.local/share/ai-setup-orchestrator/orchestrator.db` (XDG `$XDG_DATA_HOME` if set).
- Per-project state remains addressable via a `project_root` column on run tables.
- Migrations directory; numbered sequential SQL files; tracked in `schema_migrations` table.

### Catalog versioning

- Monotonic integer versions per `(kind, name)` pair (no semver).
- `definition_versions` rows are immutable; updates insert a new row and move an `active_version_id` pointer on `definitions`.
- Strict frontmatter validation per kind (Zod schemas mirroring the per-host shapes documented in `research.md`).

### Resolution priority (when a chain step references an agent or skill)

1. Explicit pin in the call (`agent: reviewer@7` or `agent: reviewer/internal@7`) → resolve directly.
2. Project-level host override (`<project>/.claude/agents/reviewer.md`) → pass through, mark `source: user_project`, never store.
3. User-global host override (`~/.config/opencode/agents/reviewer.md`, `~/.claude/agents/reviewer.md`) → pass through, mark `source: user_global`, never store.
4. Internal catalog active version → resolve from DB.

Codex is not file-resolvable (its catalog lives in its own SQLite); steps requesting codex-resolved definitions always fall through to the internal catalog.

### Logs

- Structured JSONL at `~/.local/state/ai-setup-orchestrator/logs/orchestrator-YYYY-MM-DD.log`.
- Levels: trace, debug, info, warn, error.
- `ai-setup-orchestrator tail [--filter run-id=... | --filter level=warn]` subcommand for live viewing in a separate terminal.
- Separate from the in-DB `events` table, which is the audit/replay log driving watch/subscribe.

### Watch / subscribe

- In-process `EventEmitter` fed by state-machine transitions and DB writes.
- `events` table append-only; powers replay-from-cursor for late subscribers.
- MCP notifications: `notifications/chain/updated`, `notifications/budget/warning`, `notifications/gate/required`, `notifications/catalog/changed`.
- Topics: `chain:<id>`, `team:<id>`, `workflow:<id>`, `catalog:<kind>:<name>`.
- Available only over HTTP/SSE transport; stdio clients can poll via existing `get_status` / `get_budget`.

### Transport

- stdio remains the default for single-client use.
- HTTP/SSE over Unix socket at `~/.local/share/ai-setup-orchestrator/orchestrator.sock` for shared-process / watch scenarios.
- No bearer tokens; Unix socket permissions (0600) gate access.

### Catalog management

Both CLI subcommands and MCP tools, sharing the same handlers:

- `catalog list [--kind ...] [--name ...] [--include-versions]`
- `catalog show <kind> <name> [--version N]`
- `catalog diff <kind> <name> <fromVersion> <toVersion>`
- `catalog create-version <kind> <name> --from <file>` (never overwrites; new row, new version number)
- `catalog set-active <kind> <name> <version>`
- `catalog import-from-files [--host opencode|claude-code]` (one-time bootstrap; idempotent — checksums determine whether a new version is needed)
- `catalog export-version <kind> <name> <version> <target-path>` (the only orchestrator-initiated write to a host catalog path; explicit and user-driven)

### Triggerable agents

- `agent_invoke(name, version?, prompt, context?)` MCP tool + CLI subcommand.
- Resolves via the standard priority list, composes via existing `composer.ts`, runs as a single-step chain so existing budget/journal/event machinery applies.
- Invocation recorded in DB; observable by any subscribed client.

### Message queue

- In-memory job queue keyed by `topic` with SQLite-backed checkpointing.
- Workers process chain steps, retries, and gate-resume events.
- Survives restarts; on boot, in-flight jobs from `queue_jobs` are re-queued.

## Success criteria

- All existing `*.case.ts` tests pass without modification.
- A second CLI process can `subscribe` to a chain started by a first process and receive every state transition.
- `ai-setup-orchestrator tail` shows live structured logs in a separate terminal during a chain run.
- Editing `~/.config/opencode/agents/reviewer.md` causes the next chain run to pick up the new content without restarting the orchestrator (mtime invalidation).
- Calling `catalog create-version` twice with identical content is a no-op (checksum dedup); with different content creates version N+1 and leaves version N retrievable.
- Stopping the orchestrator mid-chain, restarting, and resuming the chain produces the same final state as not stopping.

## Open risks

- **SQLite write contention** under HTTP transport with many concurrent step completions. Mitigation: WAL mode + single writer thread + queued writes. Re-evaluate at phase 4.
- **Host config scan cost** if a user has hundreds of agents. Mitigation: mtime-based memoization; full rescan only when a directory's mtime changes.
- **Migration safety** on first install over an existing file-based state. Mitigation: phase 2 includes a one-shot importer that reads existing JSON files and inserts them as DB rows; original files retained until user deletes them.
