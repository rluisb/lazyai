# Plan: 025-lazyai-runtime-refactor

**Spec:** `specs/025-lazyai-runtime-refactor/spec.md`
**ADR:** `specs/adrs/003-lazyai-runtime-boundary.md` (Proposed — human-verifiable acceptance pending)
**Date:** 2026-06-13
**Status:** Pre-approval draft — pending Phase 0 prerequisite completion and human approval. This plan is a Phase 0 deliverable (P0-8), not a post-approval artifact. It is revised as prerequisite evidence arrives.
**Owner:** Ricardo Conceicao
**Constitution:** `.specify/memory/constitution.md` Articles I–VI

> **Purpose.** This plan describes *how* the LazyAI runtime refactor will be executed. It names the prerequisite artifacts, selects the tech stack, evaluates the design against the constitution, and breaks the work into gated phases. Acceptance criteria stay in the spec; this plan defines implementation details, data models, and internal APIs.

---

## Summary

Destructive refactor of the LazyAI runtime: remove Fortnite/orchestrator bloat, shrink the runtime's 27-table and orchestrator's 11-table schemas to a compact V2 (5 runtime data tables plus migration metadata), rewrite adapters against a neutral contract, implement handoff capability, and enforce a 50KB token-rent budget on the canonical library. Execution follows the corrected order (Audit → Adapter rewrite → Test rewrite → Excision → Schema migration → Handoff → Library curation) across 6 gated phases. No implementation begins until Phase 0 prerequisites are approved and confidence reaches 85%.

---

## Technical Context

| Aspect | Decision | Rationale |
|---|---|---|
| Framework(s) | Standard Go `testing`; runtime schema migration via `runtime.SchemaV2` in `packages/cli/internal/runtime/schema.go` | Runtime schema is self-contained; applied by `openRuntimeDB()` in `runtime_helper.go` |
| Storage | SQLite (V2 schema, 5 runtime data tables plus migration metadata) | Constitution Article I: reuse existing runtime SQLite layer; no new DB dependency |
| Deployment | Bundled CLI binary | Runtime ships inside `lazyai` CLI; no separate service |
| Telemetry | None added | Out of scope; usage survey uses existing GitHub issues + direct outreach |
| Adapter contract | Markdown spec + Go test fixtures | Neutral contract defined in prose; tests assert canonical behavior per adapter mode |

**External dependencies (new):**
- None. All work uses existing Go stdlib + project-internal packages.

**External dependencies (rejected):**
- New DB driver or ORM — rejected; existing SQLite layer suffices for ~5 tables.
- New test framework — rejected; `testing` + table-driven tests cover all contracts.
- External schema migration tool — rejected; runtime schema is self-contained in `runtime.SchemaV2`, applied by `openRuntimeDB()`.
---

## Constitution Check

| Article | Verdict | Justification |
|---|---|---|
| I — Library-First | PASS | Reuses existing SQLite schema layer, existing CLI command patterns, existing Go test conventions. No new external dependencies. Two new internal packages (`handoff`, `tokenrent`) added — minimal, constitution-compliant additions. |
| II — Test-First (NON-NEGOTIABLE) | PASS | Every code phase (1–5) begins with contract tests. RED→GREEN→REFACTOR enforced per workstream. P0 produces doc-only artifacts; no code before confidence ≥85%. |
| III — Docs as Source of Truth | PASS | Spec, ADR, and this plan are approved before code phases. Adapter contract, handoff schema, V2 schema are doc artifacts before code. |
| IV — Anti-Speculation (YAGNI) | PASS | Removes Fortnite/orchestrator surfaces; does not add features, config knobs, or abstractions. Token-rent is a gate, not a feature. Handoff is a markdown writer, not a protocol. |
| V — Simplicity Over Abstraction | PASS | V2 schema targets ~5 runtime data tables plus migration metadata from 27 runtime tables. Adapter contract is concrete expected-output tables. Token-rent is `wc -c` + text file (Go package wraps the check for testability). |
| VI — Anti-Overengineering (NON-NEGOTIABLE) | PASS | Two new packages (`handoff`, `tokenrent`) — each single-file with test; no DI, no interfaces, no DRY extraction. Handoff writes markdown; tokenrent runs `wc -c`. |

**Verdict:** CONSTITUTION-COMPLIANT — corrected per two adversarial reviews (GPT-5.5 xhigh). Two new packages (`handoff`, `tokenrent`) acknowledged and justified. Plan status: Draft, pending human approval.

---
## Decisions Embedded from Clarify

| Decision | Resolution | Plan impact |
|---|---|---|
| Orchestrator paradox | B: Redesigned primary-agent path | Phase 1 rewrites non-Fortnite adapters to wire to a redesigned primary agent. No `orchestrator.md` shim. Phase 2 removes orchestrator packages entirely. |
| Usage survey source | GitHub issues + direct outreach | Phase 0 prerequisite P0-2: query `fortnite`/`opencode` tagged issues for quantitative baseline; Slack/Discourse/github discussions for qualitative blockers. |
| Token budget | 50KB project-wide, single `wc -c` gate | Phase 0 prerequisite P0-7: design enforcement pipeline. Phase 5: implement CI/pre-commit check + `.lazyai/token-rent-override` escape hatch. |

---

## Project Structure

Where changes land. Paths are actual repository paths.

```
[repo-root]/
├── packages/cli/
│   ├── cmd/                              ← Phase 2: rewrite/remove affected commands
│   │   ├── task.go                       ← rewrite or remove
│   │   ├── workflow.go                   ← rewrite or remove
│   │   ├── orchestration.go              ← remove
│   │   ├── ... (13 files with refs)      ← per audit disposition
│   ├── internal/
│   │   ├── runtime/
│   │   │   └── schema.go                 ← Phase 3: V2 migration (SchemaV2 const)
│   │   ├── db/                           ← config/setup DB only; not phase 3
│   │   ├── orchestrator/                 ← Phase 2: remove or archive
│   │   │   ├── catalog.go
│   │   │   ├── catalog_db.go
│   │   │   └── catalog_db_test.go
│   │   ├── adapter/                      ← Phase 1: rewritten neutral adapter tests
│   │   │   ├── opencode_test.go
│   │   │   ├── claudecode_test.go
│   │   │   └── copilot_test.go
│   │   ├── handoff/                      ← Phase 4: new handoff writer
│   │   │   ├── writer.go
│   │   │   └── writer_test.go
│   │   └── tokenrent/                    ← Phase 5: new budget enforcer
│   │       ├── check.go
│   │       └── check_test.go
│   └── library/
│       ├── fortnite/                     ← Phase 2: archive, then remove
│       └── canonical/                    ← Phase 5: curated replacement
│           ├── agents/
│           ├── skills/
│           ├── hooks/
│           └── commands/
├── packages/orchestrator/                ← Phase 2: archive, then remove
├── specs/
│   ├── 025-lazyai-runtime-refactor/
│   │   ├── spec.md                       ← this spec
│   │   ├── research.md                   ← research findings
│   │   ├── plan.md                       ← this file
│   │   ├── tasks.md                      ← generated after plan approval
│   │   ├── analysis.md                   ← generated after plan approval
│   │   └── checklists/                   ← gate checklists
│   │       ├── phase0.md
│   │       ├── phase1.md
│   │       ├── phase2.md
│   │       ├── phase3.md
│   │       ├── phase4.md
│   │       └── phase5.md
│   └── adrs/
│       └── 003-lazyai-runtime-boundary.md ← proposed; approval pending tracked human commit
├── .lazyai/
│   └── token-rent-override               ← Phase 5: documented exception file
└── .github/
    └── workflows/
        └── token-rent.yml                ← Phase 5: CI enforcement
```

---

## Phase 0 — Prerequisites and Approval Gate

**Estimate:** 5–7 working days
**Exit:** All eight prerequisite artifacts exist, are linked from this plan, and are approved by a human.
**Gate:** ⛔ Human must approve all P0 artifacts before Phase 1 begins.

### P0-1: CLI Command Import Audit

**Deliverable:** `specs/025-lazyai-runtime-refactor/audit-cli-imports.md`

Matrix covering every `packages/cli/cmd/*.go` file (73 files total: 50 non-test, 23 test). Each file classified as:

| Classification | Meaning |
|---|---|
| `breakage` | Import will break on excision; requires rewrite or removal |
| `rewrite` | Can be rewritten to remove dependency |
| `keep` | No orchestrator/Fortnite dependency; unchanged |
| `remove` | Entire command is Fortnite-only; safe to delete |
| `defer-with-owner` | Cannot decide now; named owner + deadline |

Target imports to audit:
- `runtime/workflow`
- `runtime/taskqueue`
- `runtime/dispatch`
- `internal/orchestrator`
- `packages/orchestrator`
- `library/fortnite`

**Evidence required per file:** import chain, call path, user-facing command impact, test coverage.

### P0-2: Fortnite/OpenCode Usage Survey

**Deliverable:** `specs/025-lazyai-runtime-refactor/survey-usage.md`

Per Clarify resolution: no CLI telemetry exists. Authoritative sources:

1. **Quantitative baseline (primary):** GitHub issues tagged `fortnite` or `opencode`. If labels return zero results (confirmed 2026-06-13), fall back to: full-text search for "fortnite" and "opencode" in issue titles/bodies; `git log -- packages/cli/library/fortnite/` for contributor count; `gh api` searches for Fortnite/OpenCode references.
2. **Quantitative baseline (secondary):** If GitHub returns zero across all queries, record "zero detectable usage" as the finding — this is a valid survey result. Do not assume non-zero usage.
3. **Qualitative blockers:** Direct outreach via Slack/Discourse/github discussions — migration sentiment, must-keep features, acceptable deprecation timeline.


### P0-3: V2 Runtime Schema Design

**Deliverable:** `specs/025-lazyai-runtime-refactor/schema-v2.md`

**Actual baseline:** The current Go runtime schema is `packages/cli/internal/runtime/schema.go` (`SchemaV1` const, 27 CREATE TABLE statements). The orchestrator has its own schema in `packages/orchestrator/internal/db/migrations.go` (11 CREATE TABLE statements, not 26). These are **two separate schemas**. P0-3 must inventory both.

**Runtime V1 tables (27):** schema_migrations, sessions, dispatches, decisions, artifacts, memories, token_log, parallel_tasks, messages, barriers, locks, teams, workflows, workflow_instances, workflow_steps, model_calls, tool_calls, gate_results, ledger_refs, cost_snapshots, checkpoints, eval_runs, eval_results, task_queue, task_claims, task_dlq, task_messages.

**Orchestrator tables (11):** schema_migrations, chain_runs, team_runs, workflow_runs, execution_plans, handoffs, error_journal, definitions, definition_versions, queue_jobs, run_events — all in `packages/orchestrator/internal/db/migrations.go` (not `sqlite.go`).

**Migration design:** V2 migration is a new step in `packages/cli/internal/runtime/schema.go` (the actual runtime schema layer). `runtime.SchemaV2` const replaces `runtime.SchemaV1`. The migration is applied by `openRuntimeDB()` in `packages/cli/cmd/runtime_helper.go` — not `packages/cli/internal/db/migrations.go` (which is config/setup only). No `lazyai migrate down` CLI command exists; this must be built or replaced with backup-restore.

Must include:
- ER diagram (Mermaid or ASCII) for V2
- DDL for all V2 tables
- Migration up SQL (SchemaV1 tables → SchemaV2 tables) — applied by `openRuntimeDB()`
- Migration down: backup `.specify/session.db` before migration; restore from backup to revert
- `loop-driver` default handling across 3 locations: `SchemaV1` DDL default, `session.go:107`, `runtime/session/session.go:64`
- Synthetic test data covering FK cascades, partial unique indexes, legacy defaults, empty databases



### P0-4: Handoff Markdown Schema

**Deliverable:** `specs/025-lazyai-runtime-refactor/schema-handoff.md`

Contract for session handoff files written by the runtime:

- **Frontmatter keys:** `goal`, `constraints`, `progress` (done/in-progress/pending), `decisions`, `critical_context`, `next_steps`
- **Required sections:** Goal, Constraints & Preferences, Progress, Key Decisions, Critical Context, Next Steps
- **Path conventions:** `specs/memory/handoffs/YYYY-MM-DD-[topic].md`
- **Ownership model:** session-scoped; one handoff per session end
- **Append/update semantics:** write-on-close; no mid-session append; repeated close writes replace the same session handoff atomically
- **Round-trip criteria:** write → read → parse → assert all sections present
### P0-5: Canonical Library Specification

**Deliverable:** `specs/025-lazyai-runtime-refactor/library-canonical.md`

Inventory of every agent, skill, hook, and command that survives the refactor:

| Field | Requirement |
|---|---|
| Name | Unique identifier |
| Type | agent / skill / hook / command |
| Path | Location in `packages/cli/library/canonical/` |
| Justification | Why it stays (usage evidence or architectural necessity) |
| Byte budget | Estimated size contribution toward 50KB cap |
| Owner | Maintainer responsible for updates |

**Exclusion criteria:** Fortnite-specific, orchestrator-dependent, unused in active workflows, superseded by redesigned primary-agent path.

**Test-first requirement:** P0-5 MUST name the surviving agents (including `primary-agent`) and their expected tool sets so that Phase 1 adapter test fixtures can reference canonical agent names before the library files are physically populated in Phase 5.

### P0-6: Adapter Test Contract

**Deliverable:** `specs/025-lazyai-runtime-refactor/contract-adapter.md`

Neutral contract for each registered adapter mode (OpenCode, Claude Code, Copilot). Gemini and Codex are out of scope — no adapter implementations exist for them in the current codebase.

| Field | Requirement |
|---|---|
| Expected files | Which files the adapter generates |
| Agents | Agent definitions produced |
| Defaults | Default agent, skill, hook, command values |
| Generated commands | Slash commands produced |
| Config merge semantics | How adapter config merges with user config |
| Validation behavior | What the adapter validates and when |
| Exclusions | Fortnite-specific names, orchestration defaults, hidden fallbacks |

**Test strategy:** table-driven Go tests. Each adapter mode gets a test fixture; each test asserts canonical output, not Fortnite library behavior.

### P0-7: Token-Rent Enforcement Design

**Deliverable:** `specs/025-lazyai-runtime-refactor/design-token-rent.md`

Per Clarify resolution: 50KB project-wide, single budget.

| Aspect | Design |
|---|---|
| Budget threshold | 50,000 bytes (50KB) total across agents + skills + hooks + commands |
| Byte-counting rule | `wc -c` on all files in `packages/cli/library/canonical/` (recursive) |
| CI integration | GitHub Actions workflow: `token-rent.yml`, runs on PR, fails if total > 50,000 bytes |
| Pre-commit integration | Hook in `.githooks/pre-commit` or `pre-commit` config |
| Failure output | `Library budget exceeded: X / 50000 bytes. Override: add .lazyai/token-rent-override with justification.` |
| Override format | `.lazyai/token-rent-override` — YAML or plain text with `reason:` field |
| Override semantics | Documented exception; auditable; does not silently bypass |
### P0-8: Corrected Phase Plan and Rollback Procedure

**Deliverable:** This document (`plan.md`) plus `specs/025-lazyai-runtime-refactor/rollback.md`.

Rollback procedure per phase — developer and operator paths:

| Phase | Trigger | Developer rollback | Operator rollback | Verification |
|---|---|---|---|---|
| 1 | Adapter tests fail | `git checkout tags/...` adapter files | N/A — adapter changes are dev-only, not in binary | `go test ./internal/adapter/...` |
| 2 | CLI command removal breaks workflow | `git checkout tags/...` full restore; rebuild | `lazyai update-self --version <prev-tag>` after Phase 2 implements tag-specific release fetch | `go build ./...` + integration tests |
| 3 | V2 migration corrupts data | `git checkout tags/...` runtime files | `lazyai restore-runtime-db .specify/session.db.backup` (added in Phase 3) | Round-trip test |
| 4 | Handoff format incompatible | `git checkout tags/...` handoff + session files | `lazyai update-self --version <prev-tag>` after Phase 2 implements tag-specific release fetch | Handoff round-trip test |
| 5 | Token-rent blocks legitimate addition | Add `.lazyai/token-rent-override` | Add `.lazyai/token-rent-override` | CI passes with override |

**Operator rollback design (current code readiness):**
- `lazyai update-self --version <tag>` — flag exists in `cmd/update-self.go`, but tag-specific GitHub Release fetch is not implemented yet. Phase 2 MUST implement and test `fetchReleaseByTag` before operator rollback can rely on this command.
- `lazyai restore-runtime-db <path>` — command exists in `cmd/restore_runtime_db.go`; it restores `.specify/session.db` from backup. Used for Phase 3+ DB rollback.
- Phase 1 has no binary surface impact — developer git checkout is sufficient.

---

## Phase 1 — Adapter Decouple and Test Rewrite

**Estimate:** 10–14 working days
**Exit:** Adapter behavior is neutral, tested, and no longer depends on Fortnite assumptions.
**Gate:** ⛔ Human approves adapter contract tests and confirms confidence ≥85% (per spec §V.1) before CLI excision begins.
**Depends on:** P0-1 (audit), P0-6 (adapter contract), Clarify Q1 (orchestrator paradox), confidence floor ≥85%

### Workstreams

1. **Write adapter test contract fixtures** — table-driven tests per adapter mode. Each fixture encodes expected files, agents, defaults, commands, config merge, and validation behavior from P0-6.

2. **Rewrite adapter tests** — replace Fortnite-encoding tests with neutral contract tests. Remove `FortniteMode` paths, orchestrator defaults, and hidden fallbacks.

3. **Apply orchestrator paradox resolution** — non-Fortnite adapters wire to redesigned primary-agent path. Remove any `orchestrator.md` agent installation.

4. **Remove dead Fortnite mode/default paths** — delete `FortniteMode` enum values, Fortnite-specific config branches, and orchestrator default references from adapter code.

5. **Replace loop-driver defaults in runtime** — three locations must be updated:
   - `packages/cli/cmd/session.go:107`: `agentName = "loop-driver"` → `agentName = "primary-agent"`
   - `packages/cli/internal/runtime/session/session.go:64`: `agent = "loop-driver"` → `agent = "primary-agent"`
   - `packages/cli/internal/runtime/schema.go:20`: `DEFAULT 'loop-driver'` → `DEFAULT 'primary-agent'` (SchemaV1 migration; SchemaV2 removes entirely)
   This is runtime command + session manager work, not adapter-only.

### Rollback

- Tag: `pre-refactor-025-phase-1`
- Trigger: adapter tests fail against neutral contract
- Command: `git checkout tags/pre-refactor-025-phase-1 -- packages/cli/internal/adapter/ packages/cli/cmd/session.go packages/cli/internal/runtime/session/session.go packages/cli/internal/runtime/schema.go`
- Verify: `go test ./internal/adapter/... ./internal/runtime/session/...` + `go build ./cmd/...`

---

## Phase 2 — CLI Command Rewrite and Excision

**Estimate:** 8–10 working days
**Exit:** Affected CLI commands are rewritten or removed, heavy orchestration packages are gone, user-facing command behavior is covered.
**Gate:** ⛔ Human approves CLI audit dispositions and rewritten command tests before schema migration.
**Depends on:** P0-1 (audit), P0-2 (usage survey), Phase 1 (adapters decoupled)

### Workstreams

1. **Rewrite/remove affected commands** — per P0-1 audit disposition:
   - `breakage` → rewrite to remove orchestrator/Fortnite imports
   - `defer-with-owner` → document owner + deadline; command MUST remain functional (no code change that breaks it)
   - `rewrite` → restructure to use redesigned primary-agent path
   - `remove` → delete command file + tests; add migration note

2. **Remove heavy orchestrator packages** — delete `packages/orchestrator/` (all subpackages). Remove `./packages/orchestrator` from `go.work`.

3. **Remove runtime orchestration packages** — delete the following from `packages/cli/internal/runtime/`:
   - `workflow/` (execute.go, fallback.go, interpolate.go, parser.go, parser_test.go, sync.go)
   - `taskqueue/` (lifecycle.go, sweep.go, taskqueue.go, taskqueue_test.go, claim.go)
   - `dispatch/` (parallel.go, parallel_test.go, dispatcher.go)
   - `session/barrier.go`, `session/lock.go`, `session/message.go`, `session/parallel.go` — Fortnite-specific coordination files
   - `session/dispatch.go` — orchestrator coupling
   Update all command callers (`task.go`, `workflow.go`, `orchestration.go`, `session.go`) and remove dead imports from `go.mod`.

4. **Remove CLI orchestrator internal** — delete `packages/cli/internal/orchestrator/`.

5. **Archive Fortnite library** — move `packages/cli/library/fortnite/` to `archive/fortnite-2026-06/`. Remove `all:fortnite` from `packages/cli/library/embed.go` embed directive.

6. **Archive orchestrator and runtime packages** — before deletion, tag and archive:
   - `git tag archive/pre-refactor-025-orchestrator` on `packages/orchestrator/`
   - `archive/` directory holds removed runtime packages (`workflow/`, `taskqueue/`, `dispatch/`, session coord files) for reference
   Spec §A.8: archival alone is not rollback; archive is for reference, tags are for git checkout rollback.

7. **Remove orchestrator from build system:** — remove `./packages/orchestrator` from `go.work` `use` directive. Remove any orchestrator-related targets from `Makefile`, CI configs, and `bin/` scripts. Verify `go work sync` and `go build ./...` pass after removal.

8. **Cleanup stale references** — scan config, helpers, doctor, add, message, validate, server, list, update, and all remaining commands for dangling orchestrator/Fortnite imports.
### Rollback

- Tag: `pre-refactor-025-phase-2`
- Command: `git checkout tags/pre-refactor-025-phase-2 -- packages/cli/cmd/ packages/orchestrator/ packages/cli/internal/orchestrator/ packages/cli/library/fortnite/ packages/cli/internal/runtime/workflow/ packages/cli/internal/runtime/taskqueue/ packages/cli/internal/runtime/dispatch/ packages/cli/internal/runtime/session/`
- Verify: affected command integration tests + `go build ./...`
## Phase 3 — V2 Schema Migration

**Estimate:** 8–12 working days
**Exit:** V1 runtime data migrates to V2 and can be restored or reversed under test.
**Depends on:** P0-3 (V2 schema), Phase 2 (orchestrator tables removed)
**Gate:** ⛔ Human approves migration test evidence (backup, restore, FK-saturated) before handoff implementation.
### Workstreams

1. **Implement `runtime.SchemaV2`** — new const in `packages/cli/internal/runtime/schema.go`. DDL from P0-3: ~5 tables (sessions, dispatches, handoff, agent_defaults, ledger_refs). Drop Fortnite-specific and observability tables from SchemaV1.

2. **Implement runtime migration in `openRuntimeDB()`** — `packages/cli/cmd/runtime_helper.go`: detect SchemaV1 → backup DB to `.specify/session.db.backup` → apply SchemaV2 → verify. No separate `lazyai migrate` command; migration is automatic on first runtime open after upgrade.

3. **Implement backup-restore rollback** — `lazyai restore-runtime-db <backup-path>` command: restores `.specify/session.db` from backup. This is the replacement for "migrate down" (which does not exist).

4. **Backup/restore verification** — automated test: backup V1 DB → migrate up (SchemaV2) → verify V2 → restore from backup → verify V1 matches original. Run on synthetic FK-saturated data.

5. **Session/dispatch/handoff round-trip tests** — write session → write dispatch → close session → write handoff → read all back → assert all fields survive migration.

6. **Empty database migration** — test `openRuntimeDB()` on a fresh `.specify/session.db` with zero rows.

7. **Legacy defaults migration** — test that `loop-driver` defaults in V1 (SchemaV1 DDL, `session.go:107`, `session/session.go:64`) become primary-agent defaults in V2 without data loss.

### Rollback

- Tag: `pre-refactor-025-phase-3`
- Trigger: V2 migration corrupts data
- Operator command: `lazyai restore-runtime-db .specify/session.db.backup` (restores from pre-migration backup)
- Developer command: `git checkout tags/pre-refactor-025-phase-3 -- packages/cli/internal/runtime/schema.go packages/cli/cmd/runtime_helper.go`
- Verify: round-trip test on production-shaped synthetic data

---

## Phase 4 — Handoff Implementation

**Estimate:** 3–4 working days
**Exit:** Runtime writes handoffs matching the approved markdown schema.
**Gate:** ⛔ Human approves handoff round-trip test evidence before library curation.
**Depends on:** P0-4 (handoff schema), Phase 3 (V2 schema with handoff table)

### Workstreams

1. **Implement handoff writer** — `packages/cli/internal/handoff/writer.go`. Writes markdown files to `specs/memory/handoffs/YYYY-MM-DD-[topic].md` matching P0-4 schema. Uses V2 handoff table for metadata.

2. **Session integration** — wire handoff write to session-close lifecycle. One handoff per session end.

3. **Round-trip validation** — test: create session → add progress → close session → read handoff file → parse frontmatter → assert all required sections present and populated.

### Rollback

- Tag: `pre-refactor-025-phase-4`
- Trigger: handoff format incompatible with readers
- Command: revert to previous session-close behavior (no handoff write)
- Verify: handoff round-trip test

---

## Phase 5 — Library Curation and Enforcement

**Estimate:** 3–4 working days
**Exit:** Canonical library stays within 50KB budget or fails with documented override instructions.
**Gate:** ⛔ Human approves token-rent CI evidence before refactor is declared complete.
**Depends on:** P0-5 (canonical library spec), P0-7 (token-rent design), Phase 2 (Fortnite library archived)

### Workstreams

1. **Curate canonical library** — per P0-5 inventory. Populate `packages/cli/library/canonical/agents/`, `skills/`, `hooks/`, `commands/` with surviving items. Remove Fortnite-only items.
2. **Wire canonical library to embed** — add `all:canonical` to the embed directive in `packages/cli/library/embed.go`. Update adapter code to read from `canonical/` instead of root `agents/`/`skills/`/orchestrator helpers.
3. **Update adapter canonical paths** — migrate adapter file-generation code to target `canonical/` subdirs instead of root library paths.

3. **Implement pre-commit hook** — `.githooks/pre-commit` (or `pre-commit` config): same `wc -c` check, same failure message.

4. **Implement `.lazyai/token-rent-override` support** — CI and pre-commit skip the budget check when override file exists with non-empty `reason:` field. Override is auditable (committed to repo).

5. **Verify enforcement** — test: add a file that pushes total > 50KB → CI fails → add override → CI passes. Test: remove override → CI fails again.

### Rollback

- Tag: `pre-refactor-025-phase-5`
- Trigger: token-rent gate blocks legitimate addition
- Command: add `.lazyai/token-rent-override` with justification
- Verify: CI passes with override present

## Dependency Order

```
P0-1 (audit) ─────────────────────────────────────────────────┐
P0-2 (survey) ─────────────────────────────────────────────────┤
P0-3 (V2 schema) ─────────────────────────────────────────────┤
P0-4 (handoff schema) ────────────────────────────────────────┤
P0-5 (canonical library) ─────────────────────────────────────┤
P0-6 (adapter contract) ──────────────────────────────────────┤
P0-7 (token-rent design) ─────────────────────────────────────┤
P0-8 (rollback procedure) ────────────────────────────────────┤
                                                               │
                    ⛔ Phase 0 Gate: all P0 artifacts approved │
                                                               │
                    ⛔ Confidence Gate: human confirms ≥85%     │
                       (no code before this point)              │
                                                               │
Phase 1 (adapters) ◄── P0-1, P0-6, Clarify Q1, confidence≥85% │
     │                                                         │
     ⛔ Phase 1 Gate                                            │
     │                                                         │
Phase 2 (excision) ◄── P0-1, P0-2, Phase 1                    │
     │                                                         │
     ⛔ Phase 2 Gate                                            │
     │                                                         │
Phase 3 (schema) ◄── P0-3, Phase 2                             │
     │                                                         │
     ⛔ Phase 3 Gate                                            │
     │                                                         │
Phase 4 (handoff) ◄── P0-4, Phase 3                            │
     │                                                         │
     ⛔ Phase 4 Gate                                            │
     │                                                         │
Phase 5 (library) ◄── P0-5, P0-7, Phase 2                      │
     │                                                         │
     ⛔ Phase 5 Gate: refactor complete                         │
```

---

## Risk Register

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| CLI audit reveals more breakage than estimated | Medium | Schedule slip 5–10 days | P0-1 is the first prerequisite; plan adjusted before Phase 1 |
| Usage survey shows high Fortnite dependency | Medium | Scope reduction: some Fortnite items retained as canonical | P0-2 gates excision decisions |
| Adapter contract misses edge case | Medium | Adapter tests fail in Phase 1 | Contract reviewed by adapter owners before test rewrite |
| Token-rent budget too tight for canonical library | Medium | Legitimate items blocked | `.lazyai/token-rent-override` escape hatch; budget adjustable |
| Orchestrator removal breaks undiscovered dependency | Low-Medium | Build breakage | P0-1 audit covers all known imports; `go build ./...` gate |

| V2 migration corrupts production data | Medium (untested) | Critical: data loss | Backup-before-migrate invariant; `lazyai restore-runtime-db` command implemented; synthetic migration tests designed in P0-3 but not yet executed |

## Confidence Assessment

| Stage | Confidence | Threshold |
|---|---|---|
| Current (post-research, pre-P0) | ~63% | — |
| After P0 artifacts complete | ~78% | — |
| After P0 + human confidence gate | ~85% | **Code floor** — all code phases (1–5) blocked below this |
| After Phase 1 (adapters decoupled) | ~87% | — |
| After Phase 2 (excision) | ~90% | — |
| After Phase 5 (complete) | ~92% | — |

**Implementation gate (corrected per spec §V.1):** NO implementation (code changes) may begin while confidence is below 85%. P0 artifacts are documentation and research — they do not count as implementation. Phase 1 writes Go test fixtures and rewrites adapter code — that IS implementation, blocked until confidence ≥85%. The 85% gate requires: (a) all P0 artifacts approved, (b) human explicitly confirms confidence ≥85%.

---

## Total Estimate

| Phase | Working days |
|---|---|
| Phase 0 — Prerequisites | 5–7 |
| Phase 1 — Adapter decouple | 10–14 |
| Phase 2 — CLI excision | 8–10 |
| Phase 3 — Schema migration | 8–12 |
| Phase 4 — Handoff | 3–4 |
| Phase 5 — Library curation | 3–4 |
| **Total** | **37–51 working days** |

---

## Next Steps

1. Complete all eight Phase 0 prerequisite artifacts (P0-1 through P0-7) with concrete evidence.
2. Revise this plan (P0-8) to incorporate evidence from completed prerequisites.
3. ⛔ **Human approves the revised plan (Phase 0 gate).**
4. After plan approval, generate `tasks.md` with decomposed work items per phase.
5. Generate `analysis.md` with risk/impact/dependency analysis.
6. Generate `checklists/` gate files for each phase.
7. Begin Phase 1 implementation (no code changes before confidence ≥85%).
