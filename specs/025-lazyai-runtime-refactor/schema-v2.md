# P0-3: V2 Runtime Schema Design

**Status:** Complete — target tables and migration path defined  
**Owner:** Ricardo Conceicao  
**Date:** 2026-06-14  
**Linked from:** `plan.md` Phase 0, P0-3

---

## Purpose

Design the V2 runtime schema that replaces the current 27-table SchemaV1 and absorbs the 11-table orchestrator schema. Target: 5 runtime data tables plus migration metadata.

## Baseline Inventory

### Runtime V1 (27 tables in `packages/cli/internal/runtime/schema.go`)

`schema_migrations`, `sessions`, `dispatches`, `decisions`, `artifacts`, `memories`, `token_log`, `parallel_tasks`, `messages`, `barriers`, `locks`, `teams`, `workflows`, `workflow_instances`, `workflow_steps`, `model_calls`, `tool_calls`, `gate_results`, `ledger_refs`, `cost_snapshots`, `checkpoints`, `eval_runs`, `eval_results`, `task_queue`, `task_claims`, `task_dlq`, `task_messages`

### Orchestrator (11 tables in `packages/orchestrator/internal/db/migrations.go`)

`schema_migrations`, `chain_runs`, `team_runs`, `workflow_runs`, `execution_plans`, `handoffs`, `error_journal`, `definitions`, `definition_versions`, `queue_jobs`, `run_events`

## V2 Runtime Data Target (5 tables + migration metadata)

| Table | Purpose | From | Key Columns |
|---|---|---|---|
| `sessions` | Session lifecycle | V1 `sessions` | `id`, `started_at`, `ended_at`, `agent`, `status`, `summary` |
| `dispatches` | Agent dispatch records | V1 `dispatches` | `id`, `session_id`, `agent`, `input`, `output`, `created_at` |
| `handoff` | Session handoff metadata | New (from orchestrator `handoffs`) | `id`, `session_id`, `path`, `goal`, `status`, `created_at` |
| `agent_defaults` | Per-tool agent defaults | New | `tool_id`, `default_agent`, `instructions` |
| `ledger_refs` | Cost/usage ledger | V1 `ledger_refs` | `id`, `session_id`, `event_type`, `metadata`, `created_at` |

Migration metadata: retain/recreate `schema_migrations` as migration bookkeeping. It is not counted as one of the five runtime data tables.

## Dropped or Replaced Tables (34 listed tables + retained runtime metadata)

**Runtime V1 data tables (23 dropped/replaced):** `decisions`, `artifacts`, `memories`, `token_log`, `parallel_tasks`, `messages`, `barriers`, `locks`, `teams`, `workflows`, `workflow_instances`, `workflow_steps`, `model_calls`, `tool_calls`, `gate_results`, `cost_snapshots`, `checkpoints`, `eval_runs`, `eval_results`, `task_queue`, `task_claims`, `task_dlq`, `task_messages`

**Runtime metadata:** `schema_migrations` is retained/recreated for migration bookkeeping and is not counted in the five runtime data tables.

**Orchestrator tables (11 dropped/replaced):** `schema_migrations`, `chain_runs`, `team_runs`, `workflow_runs`, `execution_plans`, `handoffs` (replaced by V2 `handoff`), `error_journal`, `definitions`, `definition_versions`, `queue_jobs`, `run_events`

## ER Diagram (ASCII)

```
┌──────────────┐       ┌──────────────┐
│   sessions   │       │agent_defaults│
│──────────────│       │──────────────│
│ id (PK)      │       │ tool_id (PK) │
│ started_at   │       │ default_agent│
│ ended_at     │       │ instructions │
│ agent        │       └──────────────┘
│ status       │
│ summary      │
└──────┬───────┘
       │ 1:N
       ├──────────────────────────────┐
       │                              │
┌──────┴───────┐              ┌───────┴──────┐
│  dispatches  │              │   handoff    │
│──────────────│              │──────────────│
│ id (PK)      │              │ id (PK)      │
│ session_id FK│              │ session_id FK│
│ agent        │              │ path         │
│ input        │              │ goal         │
│ output       │              │ status       │
│ created_at   │              │ created_at   │
└──────────────┘              └──────────────┘

       │ 1:N
┌──────┴───────┐
│ ledger_refs  │
│──────────────│
│ id (PK)      │
│ session_id FK│
│ event_type   │
│ metadata     │
│ created_at   │
└──────────────┘
```

## Migration Design

- **Up:** `openRuntimeDB()` in `packages/cli/cmd/runtime_helper.go` detects SchemaV1 → backs up to `.specify/session.db.backup` → applies SchemaV2
- **Rollback:** `lazyai restore-runtime-db <backup-path>` restores from backup (command implemented in `cmd/restore_runtime_db.go`)
- **No down migration SQL:** No `lazyai migrate down` command exists; backup-restore is the rollback path

## Default Handling

- `sessions.agent DEFAULT 'loop-driver'` → removed in V2; `agent_defaults` table provides per-tool defaults
- `session.go:107` `agentName = "loop-driver"` → `agentName = "primary-agent"`
- `session/session.go:64` `agent = "loop-driver"` → `agent = "primary-agent"`

## Test Plan

1. Synthetic FK-saturated data migration (create V1 DB with all 27 tables populated, migrate, verify 5 tables)
2. Empty database migration (fresh `.specify/session.db` with zero rows)
3. Legacy defaults migration (`loop-driver` → `primary-agent` in agent_defaults)
4. Backup → migrate → restore → verify round-trip
5. Session/dispatch/handoff round-trip (write session → write dispatch → close → write handoff → read all back)

## Gate

⛔ Human must approve this schema design before Phase 3 migration begins.
