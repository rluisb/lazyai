# Spec 006 — Housekeeping, Memory, and Bootstrap Workflows

> Scope: Formalize bootstrap, housekeeping, and memory lifecycle workflows; standardize metadata across RPI and orchestrator; clean up AGENTS.md sprawl; define approval contracts for qmd/codegraph sync. No implementation code.

---

## Approved Decisions (User-Ratified)

| # | Decision | Implications |
|---|----------|--------------|
| D-1 | Memory extraction can happen both **inline during workflow execution** and **during post-task housekeeping** — system runs it whenever required | RPI workflow gains an optional "memory extraction" checkpoint after each step; post-task housekeeping gains a mandatory extraction sweep. Neither is exclusive. |
| D-2 | Conflicting memories reconcile by **appending rather than destructive editing**, using a **timeline/history model** within the same memory file; newer entries may supersede older ones while preserving history | Memory file format must support ordered entries with timestamps. File-size / compaction risk is real — plan must include a compaction strategy and a soft size limit with warning. |
| D-3 | Standardize metadata across RPI and all workflows. Minimum candidate fields: `id`, `title`, `ticket_number`, `created_at`, `updated_at`, `created_by`, `updated_by`, `risk_level`, `size_points`/`complexity_level`, `session_id`, `workflow_id`, `workflow_run_id`, `team_id`, `chain_id`, `owner_agent`/`assignee`, `step_ids`/`workflow_steps`, `related_document_refs` | All spec subdirectories, memory files, and chain/handoff artifacts converge on one metadata schema. Backward-compatible additions allowed; removals forbidden without migration. |
| D-4 | qmd and codegraph must stay synced, but the system preserves **approval-first behavior** using an **explicit standing approval / maintenance contract model** rather than silent mutation | A formal "maintenance contract" concept is introduced. Read-only drift/staleness checks are always allowed. Actual sync/repair requires either per-action approval or an active contract. |

---

## Acceptance Criteria

| # | AC | Verified By |
|---|----|-------------|
| AC-1 | Bootstrap workflow is specified as a deterministic sequence: discovery → drift/staleness check (read-only) → approval request → context load → bootstrap report | Spec review + manual walkthrough |
| AC-2 | Housekeeping (pre- and post-task) workflow is specified with explicit memory extraction points (inline + post-task), cleanup proposal, and sync proposal steps; all mutating actions require approval or active contract | Spec review + manual walkthrough |
| AC-3 | Memory file format supports timeline/history model with append-only entries, timestamps, and supersession indicators; a compaction strategy is defined with soft size limits (warn ≥ 200 lines, hard cap at 500 lines) | Schema review + manual test with synthetic memory file |
| AC-4 | A standardized metadata schema is defined and applied to: spec dirs, memory files, chain/handoff artifacts, and RPI task files; field names and types are documented | Schema document review + spot check against existing files |
| AC-5 | Maintenance contract model is specified: contract types (session-scoped, task-scoped, standing), permitted actions per contract, drift/staleness detection behavior (always read-only), and revocation | Spec review + manual walkthrough |
| AC-6 | qmd/codegraph drift detection, repair, and approval behaviors are specified: when to check, what triggers a proposal, how approval works, what happens on rejection | Spec review |
| AC-7 | Install-time UX options for Obsidian (`ob`), qmd, and codegraph are specified with approval-model descriptions per integration | Spec review + wizard flow walkthrough |
| AC-8 | AGENTS.md cleanup/migration is specified: which files to remove, what replaces them, how `compose_agent` injects context, and root-only enforcement | File inventory diff + manual verification |
| AC-9 | Clear ownership boundaries: what lands in ai-setup CLI/wizard vs orchestrator runtime vs memory/spec conventions | Architectural diagram / boundary table review |
| AC-10 | All open decisions and blockers are preserved in the plan with explicit "DECISION NEEDED" markers | Plan review |

---

## Architecture Boundaries

| Layer | Owns | Does NOT Own |
|-------|------|-------------|
| **ai-setup CLI / wizard** | Install-time UX (optional `ob`, `qmd`, `codegraph` toggles), memory path configuration, `AGENTS.md` generation/scaffolding, bootstrap/housekeeping *specification* | Runtime execution of bootstrap/housekeeping, tool adapter internals |
| **Orchestrator runtime** | Session lifecycle hooks (bootstrap at chain start, housekeeping at chain end), maintenance contract state management, `compose_agent` context injection | Install detection, wizard UI, memory file format specification |
| **Memory / spec conventions** | Memory file format schema, metadata schema, handoff format, promotion/deletion lifecycle | Tool implementations, CLI code, orchestrator internals |

---

## Implementation Phases

### Phase 1 — Memory File Format & Metadata Schema (Foundation)

**Task 1-1**: Define memory file format schema
- Output: `specs/006-housekeeping-memory-and-bootstrap/data-model.md` — Appendix section for memory file format
- Format: YAML frontmatter (standardized metadata per D-3) + markdown body with append-only timeline entries
- Each entry: `## [timestamp] [author]` heading, content body, optional `supersedes: <ref>` field
- Compaction strategy: warn at ≥ 200 lines, propose compaction at ≥ 500 lines. Compaction merges superseded entries into a summary, preserving the most recent entry in full and linking to historical snapshots.
- File-size risk: called out explicitly — memory files are advisory, not archival. A `compacted_from` header field links to the pre-compaction snapshot stored in `specs/memory/archive/`.
- Touch: `specs/006-housekeeping-memory-and-bootstrap/data-model.md`

**Task 1-2**: Define standardized metadata schema
- Output: `specs/006-housekeeping-memory-and-bootstrap/data-model.md` — Core schema section
- All fields from D-3 with types, optionality, and validation rules
- Schema version field for forward compat (`schema_version: 1`)
- Applied to: spec dirs, memory files, chain/handoff artifacts, RPI task files
- Naming convention: snake_case for YAML, camelCase for JSON/programmatic use
- Touch: `specs/006-housekeeping-memory-and-bootstrap/data-model.md`

**Task 1-3**: Map existing files against metadata schema
- Audit current spec dirs, memory files, and chain artifacts
- Identify gaps (which existing files are missing required metadata)
- Produce a migration checklist in `specs/006-housekeeping-memory-and-bootstrap/tasks/001-metadata-migration.md`
- Touch: `specs/006-housekeeping-memory-and-bootstrap/tasks/001-metadata-migration.md`

**depends_on**: none
**parallel**: `[P]` — independent of all other phases

---

### Phase 2 — Maintenance Contract & Approval Model

**Task 2-1**: Define maintenance contract schema
- Output: `specs/006-housekeeping-memory-and-bootstrap/data-model.md` — Contract schema section
- Contract types:
  - `session_scoped`: valid for current session only, expires on session end
  - `task_scoped`: valid for a specific task/workflow run, expires on completion
  - `standing`: persistent until explicitly revoked, requires periodic re-affirmation (default 30-day TTL)
- Permitted actions per contract type (matrix: memory load, memory write, qmd sync, codegraph sync, cleanup, repair)
- Drift/staleness checks: **always read-only, always allowed** (no contract needed). Only mutation requires contract or per-action approval.
- Contract revocation: user can revoke at any time; revocation triggers a "deferred maintenance" notice in next bootstrap report
- Touch: `specs/006-housekeeping-memory-and-bootstrap/data-model.md`

**Task 2-2**: Define qmd/codegraph drift detection, repair, and approval behavior
- Output: `specs/006-housekeeping-memory-and-bootstrap/data-model.md` — Sync strategy section
- Drift detection triggers:
  - Pre-task housekeeping (always)
  - Bootstrap (always, if tools enabled)
  - Post-file-write (if contract active)
- Drift detection method: compare file mtime hash or content hash against last-known index timestamp (stored in contract metadata or a `.sync-state.json` marker file)
- Proposal flow: drift detected → proposal generated (list of stale/desynchronized items) → approval requested (unless contract covers it) → action taken → report
- Rejection behavior: drift recorded but not repaired; next bootstrap report notes "stale indexes" and warns about potential context mismatch
- Repair: full re-index of stale items (not incremental for simplicity in v1)
- Touch: `specs/006-housekeeping-memory-and-bootstrap/data-model.md`

**Task 2-3**: Define rejection and staleness propagation rules
- If user rejects sync: mark items as `stale_acked` in `.sync-state.json` so the system doesn't re-prompt until new drift is detected
- If user rejects memory load: proceed without that context, but note in bootstrap report that context may be incomplete
- If contract expires during a running task: task may complete under the expired contract (grandfather), but next task requires new contract or per-action approval
- Touch: `specs/006-housekeeping-memory-and-bootstrap/data-model.md`

**depends_on**: Phase 1 (metadata schema must exist for contract metadata fields)
**parallel**: sequential after Phase 1

---

### Phase 3 — Bootstrap Workflow Specification

**Task 3-1**: Specify bootstrap sequence
- Output: `specs/006-housekeeping-memory-and-bootstrap/tasks/002-bootstrap-workflow.md`
- Sequence:
  1. **Discovery**: Identify memory path (default `specs/memory`, user-configurable in `.opencode.json` or AGENTS.md), check for `specs/` directory, check for Obsidian vault
  2. **Drift/Staleness Check** (read-only, always allowed): check qmd index freshness (if enabled), check codegraph freshness (if enabled), check `.sync-state.json`
  3. **Approval Request**: If memory context or sync is needed, and no active contract covers it, ask user for approval. If contract exists, skip to step 4.
  4. **Context Load**: Load relevant memory files via qmd search (if enabled) or file read. Extract metadata from found files. If codegraph enabled and relevant, load code-context.
  5. **Bootstrap Report**: Summary of: memory files loaded, indexes status (fresh/stale/synced), contract status, deferred maintenance items
- Touch: `specs/006-housekeeping-memory-and-bootstrap/tasks/002-bootstrap-workflow.md`

**Task 3-2**: Specify bootstrap hook integration with orchestrator
- When: `start_chain` is invoked (before first step executes)
- How: Orchestrator runtime calls bootstrap sequence as a pre-step
- Fallback: If bootstrap fails (no memory path, no qmd, etc.), log warning and proceed — bootstrap is non-blocking
- Touch: `specs/006-housekeeping-memory-and-bootstrap/tasks/002-bootstrap-workflow.md`

**depends_on**: Phase 2 (contract model must be defined for approval gating)
**parallel**: Can proceed in parallel with Phase 4 once Phase 2 is complete

---

### Phase 4 — Housekeeping Workflow Specification

**Task 4-1**: Specify pre-task housekeeping
- Output: `specs/006-housekeeping-memory-and-bootstrap/tasks/003-housekeeping-workflow.md`
- Steps:
  1. Verify maintenance contract for next step (if any mutating actions are upcoming)
  2. Check for newer memories/specs that may change approach (read-only drift check)
  3. If contract expired or missing: request approval before memory/context load
  4. Report: what context was loaded, what's stale, what's deferred
- Touch: `specs/006-housekeeping-memory-and-bootstrap/tasks/003-housekeeping-workflow.md`

**Task 4-2**: Specify post-task housekeeping
- Steps:
  1. **Memory Extraction** (mandatory sweep): Identify new lessons, decisions, patterns from the task. Capture as append-only timeline entries in the relevant memory file. Per D-1, extraction can also happen inline — this sweep catches anything missed inline.
  2. **Cleanup Proposal**: Identify temporary artifacts, git-hygiene opportunities, organization work. Propose to user.
  3. **Sync Proposal**: Identify required qmd/codegraph updates. Check drift. Propose to user (unless contract covers it).
  4. **Approval**: Perform writes/cleanup/sync only after explicit approval or within standing approval. Report what changed.
- Touch: `specs/006-housekeeping-memory-and-bootstrap/tasks/003-housekeeping-workflow.md`

**Task 4-3**: Specify inline memory extraction checkpoint
- Per D-1, memory extraction is not exclusive to post-task; it can happen inline
- Trigger: after any RPI step completes (Research → Plan, Plan → Implement, Implement → Review)
- Behavior: agent proposes memory entries; if contract covers writes, they land immediately; otherwise, batched for post-task approval
- Touch: `specs/006-housekeeping-memory-and-bootstrap/tasks/003-housekeeping-workflow.md`

**Task 4-4**: Specify housekeeping hook integration with orchestrator
- Pre-task: `advance_chain` triggers pre-task housekeeping before executing the next step
- Post-task: `advance_chain` triggers post-task housekeeping after a step completes
- Inline: step completion callback triggers inline extraction checkpoint
- Touch: `specs/006-housekeeping-memory-and-bootstrap/tasks/003-housekeeping-workflow.md`

**depends_on**: Phase 2 (contract model), Phase 1 (memory format)
**parallel**: Can proceed in parallel with Phase 3 once Phase 2 is complete

---

### Phase 5 — AGENTS.md Cleanup & Migration

**Task 5-1**: Inventory all AGENTS.md files for removal
- Current sprawl (10 files in `specs/` subdirectories):
  - `specs/rules/AGENTS.md`
  - `specs/templates/AGENTS.md`
  - `specs/standards/AGENTS.md` (if exists; inferred from pattern)
  - `specs/memory/AGENTS.md`
  - `specs/prompts/AGENTS.md`
  - `specs/tech-debt/AGENTS.md`
  - `specs/adrs/AGENTS.md`
  - `specs/refactors/AGENTS.md`
  - `specs/bugfixes/AGENTS.md`
  - `specs/features/AGENTS.md`
- Plus root `AGENTS.md` (kept as canonical)
- Plus `.opencode/agents/` directory (kept — these are tool-specific agent definitions, not context rules)
- Output: `specs/006-housekeeping-memory-and-bootstrap/tasks/004-agents-cleanup.md` — full removal list with justification per file
- Touch: `specs/006-housekeeping-memory-and-bootstrap/tasks/004-agents-cleanup.md`

**Task 5-2**: Define replacement strategy: `compose_agent` context injection
- Currently: subdirectory AGENTS.md files provide auto-loading context for agents working in that directory
- Replacement: Orchestrator's `compose_agent` tool dynamically injects role-specific instructions from the library
- Migration path:
  1. Extract unique content from each subdirectory AGENTS.md into `library/specs-agents/` files (already partially done — `library/specs-agents/` contains `bugfixes.md`, `features.md`, etc.)
  2. Verify `compose_agent` can inject these files as step instructions
  3. Delete subdirectory AGENTS.md files
  4. Update root AGENTS.md to reference the library-based approach
- Root AGENTS.md: **keep and enhance** — it remains the canonical project-level context
- `.opencode/agents/` directory: **keep unchanged** — these are tool-specific agent definitions
- Touch: `specs/006-housekeeping-memory-and-bootstrap/tasks/004-agents-cleanup.md`

**Task 5-3**: Define enforcement mechanism for root-only AGENTS.md
- `ai-setup doctor` gains a check: "stray AGENTS.md files detected in subdirectories"
- `ai-setup init` / `ai-setup update` should never generate AGENTS.md in subdirectories
- Library template `AGENTS.template.md` should be updated to remove subdirectory-generation logic
- Touch: `specs/006-housekeeping-memory-and-bootstrap/tasks/004-agents-cleanup.md`

**depends_on**: Phase 1 (metadata schema may apply to AGENTS.md replacement content)
**parallel**: `[P]` — independent of Phases 2, 3, 4

---

### Phase 6 — Install-Time UX for Optional Tooling

**Task 6-1**: Specify Obsidian (`ob`) integration options
- Discovery: `ob` detects vault paths, can surface vault config
- Approval model: read-only discovery (no contract needed), vault configuration writes require approval
- Install-time: `ai-setup init` offers "Enable Obsidian integration?" step (Phase 1 or Phase 2 of wizard)
- If enabled: scaffold Obsidian-related config (vault path in `.opencode.json` or memory config)
- Touch: `specs/006-housekeeping-memory-and-bootstrap/tasks/005-install-ux.md`

**Task 6-2**: Specify qmd integration options
- Purpose: preferred retrieval/index engine for markdown corpora
- Approval model:
  - Read-only: qmd searches always allowed
  - Index builds/rebuilds: require approval or contract (they write to the index)
  - Drift checks: always allowed (read-only mtime comparison)
- Install-time: `ai-setup init` offers "Enable qmd for markdown retrieval?" step
- If enabled: configure qmd index path, add to `--memory-path` resolution chain
- Touch: `specs/006-housekeeping-memory-and-bootstrap/tasks/005-install-ux.md`

**Task 6-3**: Specify codegraph integration options
- Purpose: structural code analysis and code-context preparation
- Approval model:
  - Read-only: codegraph queries always allowed
  - Index builds/rebuilds: require approval or contract
  - Drift checks: always allowed
- Install-time: `ai-setup init` offers "Enable codegraph for structural analysis?" step
- If enabled: configure codegraph data path, add to bootstrap context-load sequence
- Touch: `specs/006-housekeeping-memory-and-bootstrap/tasks/005-install-ux.md`

**Task 6-4**: Specify memory path configuration in wizard
- `ai-setup init` asks: "Where should project memory be stored?" (default: `specs/memory`)
- Memory path is persisted to project config (`.opencode.json` or `AGENTS.md` frontmatter)
- If qmd enabled: memory path is added to qmd index
- If codegraph enabled: memory path is added to codegraph scan scope
- Touch: `specs/006-housekeeping-memory-and-bootstrap/tasks/005-install-ux.md`

**depends_on**: Phase 2 (contract model needed for approval-model descriptions)
**parallel**: Can proceed in parallel with Phases 3, 4 (consumes contract model but doesn't block workflow specs)

---

### Phase 7 — Verification & Quickstart

**Task 7-1**: Create quickstart verification runbook
- Output: `specs/006-housekeeping-memory-and-bootstrap/quickstart.md`
- Manual walkthrough steps:
  1. Run `ai-setup init` with all optional tooling enabled
  2. Verify memory path configured correctly
  3. Simulate bootstrap: start a chain, verify bootstrap report appears
  4. Simulate housekeeping: complete a step, verify memory extraction proposal
  5. Test contract: grant session-scoped contract, verify sync happens without re-prompt
  6. Test contract revocation: revoke mid-task, verify deferred maintenance notice
  7. Test drift: modify a file, run pre-task housekeeping, verify drift detected
  8. Verify no stray AGENTS.md files in subdirectories after `doctor` check
- Touch: `specs/006-housekeeping-memory-and-bootstrap/quickstart.md`

**Task 7-2**: Create research / decisions log
- Output: Update existing `research.md` with cross-references to this plan
- Touch: `specs/006-housekeeping-memory-and-bootstrap/research.md`

**Task 7-3**: Validate all ACs against plan output
- Walk through each AC and verify the plan addresses it
- Document any gaps as "DECISION NEEDED" entries
- Touch: this file (`plan.md`) — AC mapping section

**depends_on**: All previous phases
**parallel**: sequential (final phase)

---

## Dependency Graph

```
Phase 1 (Memory Format + Metadata Schema) ──┬──► Phase 2 (Contracts + Sync) ──┬──► Phase 3 (Bootstrap)
Phase 1                                       │                                 └──► Phase 4 (Housekeeping)
                                              │
Phase 5 (AGENTS.md Cleanup) ─────────────────┤ [P] independent
                                              │
Phase 1 ──► Phase 2 ──► Phase 6 (Install UX) ┤
                                              │
All phases ───────────────────────────────────► Phase 7 (Verification)
```

- Phase 1 is the foundation (all others depend on it except Phase 5)
- Phase 5 is `[P]` — fully independent, can run in parallel with anything
- Phase 3 and Phase 4 can run in parallel once Phase 2 completes
- Phase 6 requires Phase 2 (contract model) but can run in parallel with Phases 3/4
- Phase 7 is final, depends on all others

---

## qmd / Codegraph Sync Strategy (Detail)

### When Sync Checks Occur

| Trigger | Tool | Check Type | Approval Required? |
|---------|------|------------|-------------------|
| Bootstrap | qmd + codegraph | Staleness (mtime diff vs `.sync-state.json`) | No (read-only) |
| Pre-task housekeeping | qmd + codegraph | Staleness | No (read-only) |
| Post-file-write (with contract) | qmd + codegraph | Re-index proposal | Only if no active contract |
| Post-task housekeeping | qmd + codegraph | Full drift + sync proposal | Only if no active contract |

### Sync Repair Flow

```
Drift detected (read-only check)
  → Generate sync proposal (which items are stale, what re-index will do)
  → Check active contract
    → Contract covers sync? → Execute sync → Report
    → No contract? → Request approval
      → Approved? → Execute sync → Report
      → Rejected? → Mark stale_acked → Defer → Next bootstrap warns
```

### Staleness Markers

File: `.sync-state.json` (in project root or memory path)

```json
{
  "qmd": { "last_index_time": "2026-04-17T10:00:00Z", "index_path": ".qmd-index" },
  "codegraph": { "last_index_time": "2026-04-15T08:00:00Z", "data_path": ".codegraph" },
  "stale_acked": {
    "qmd": [],
    "codegraph": []
  }
}
```

---

## Memory File Format Schema (Detail)

### Frontmatter (Standardized Metadata)

```yaml
---
schema_version: 1
id: mem-2026-0417-auth-timeout
title: "Auth service timeout behavior"
ticket_number: PROJ-456
created_at: 2026-04-17T10:00:00Z
updated_at: 2026-04-17T14:30:00Z
created_by: builder-agent
updated_by: builder-agent
risk_level: low
complexity_level: low
session_id: sess-abc123
workflow_id: wf-rpi
workflow_run_id: run-001
team_id: platform
chain_id: chain-xyz
owner_agent: builder
assignee: builder
step_ids: [step-3, step-4]
related_refs:
  - specs/004-go-migration/plan.md
  - specs/rules/security.md
compacted_from: null   # set if file was compacted; points to archive snapshot
---
```

### Body (Timeline/History Model)

```markdown
## [2026-04-17T10:00:00Z] builder-agent

Discovered: The auth service has a 5-second timeout on webhook processing.
This was found during bugfix PROJ-456. The timeout is configurable via
`AUTH_WEBHOOK_TIMEOUT_MS` env var (default: 5000).

---

## [2026-04-17T14:30:00Z] builder-agent
Supersedes: [2026-04-17T10:00:00Z]

Updated: After PROJ-490, the default timeout was raised to 10 seconds.
The env var was renamed to `AUTH_WEBHOOK_TIMEOUT` (no `_MS` suffix).
Old env var still works as fallback with a deprecation log.

The original 5-second discovery is preserved above for historical context.
```

### Compaction Rules

| Condition | Action |
|-----------|--------|
| File < 200 lines | No action |
| File ≥ 200 lines | Log warning at next housekeeping: "Memory file `auth-timeout.md` exceeds 200 lines. Consider compaction." |
| File ≥ 500 lines | Propose forced compaction at next housekeeping (user can defer once) |
| Compaction execution | Merge superseded entries into a single summary line per cluster. Move full pre-compaction file to `specs/memory/archive/YYYY-MM-DD-<id>.md`. Set `compacted_from` in frontmatter. |

---

## Risks & Mitigations

| # | Risk | Severity | Mitigation |
|---|------|----------|------------|
| R-1 | Memory file size bloat from append-only timeline model (D-2) | High | Soft limits (warn at 200, propose at 500), compaction strategy, archive directory. Compaction is not deletion — history preserved in archive. |
| R-2 | Approval gate bottleneck — user asked for approval too frequently | High | Maintenance contracts reduce friction. Three tiers (session/task/standing). Standing contract with 30-day re-affirmation prevents "approve once, forget forever." |
| R-3 | Metadata schema too rigid — fields don't fit all artifact types | Medium | Schema is additive only. Optional fields with sensible defaults. `schema_version` field allows forward-compatible extension. |
| R-4 | AGENTS.md removal breaks existing agent context loading | Medium | `compose_agent` replacement must be verified before deletion. Phased removal: extract → verify → delete. `ai-setup doctor` detects stray files. |
| R-5 | qmd/codegraph not installed — entire housekeeping feature degrades | Medium | All tooling is optional. Without qmd, memory loading falls back to direct file read. Without codegraph, code-context loading is skipped. Bootstrap report notes which tools are unavailable. |
| R-6 | `.sync-state.json` drifts from actual index state (e.g., manual index rebuild) | Low | Staleness check compares mtime of source files vs `last_index_time`. Manual rebuilds update `last_index_time` — if they don't, next check detects the mismatch and proposes sync. |
| R-7 | Standing contract TTL (30 days) too short or too long for different teams | Low | TTL is configurable per project (`.opencode.json`). 30 days is the default, not a mandate. |
| R-8 | Obsidian `ob` CLI not available or has OS-specific pathing | Low | Integration is fully optional. Absence of `ob` only disables vault discovery — no other functionality affected. Bootstrap report notes if `ob` was expected but not found. |

---

## Out of Scope

- Implementation code for any CLI, orchestrator, or runtime component
- Internal logic of qmd or codegraph
- Database/store migration (separate spec after Phase 2 of orchestrator roadmap)
- Full `compose_agent` rewrite (only the specification of how it replaces AGENTS.md loading)
- Non-catalog tool support (only `ob`, `qmd`, `codegraph` are in scope for install UX)
- Headless/CI integration for housekeeping (future consideration)

---

## Open Decisions & Blockers

| # | Item | Status | Impact if Unresolved |
|---|------|--------|---------------------|
| OD-1 | Sync state path resolved to `.ai/housekeeping/sync-state.json` | RESOLVED | Canonical location selected for drift detection and maintenance state |
| OD-2 | Metadata enforcement resolved to **strict for new artifacts; warn/migrate for legacy artifacts** | RESOLVED | New writes can validate strongly without blocking legacy reads |
| OD-3 | `compose_agent` integration assumption: use the existing API first and inject library/spec context through `stepInstructions` / currently supported fields | RESOLVED FOR THIS SPEC SLICE | Unblocks Phase 5 planning without introducing new API surface in Phases 1-2 |
| OD-4 | Standing approvals hard-expire at 30 days, but an in-flight task may finish | RESOLVED | Clarifies TTL behavior without a post-expiry grace window for new tasks |
| OD-5 | `ai-setup doctor` reports stray `AGENTS.md` files by default; no silent deletion | RESOLVED | Keeps enforcement safe and approval-first |
| OD-6 | qmd index location defaults to project-local | RESOLVED | Install UX and sync-state behavior now have a default |

---

## AC Mapping

| AC | Covered In | Status |
|----|-----------|--------|
| AC-1 | Phase 3 (Task 3-1) | ✅ Addressed |
| AC-2 | Phase 4 (Tasks 4-1, 4-2, 4-3) | ✅ Addressed |
| AC-3 | Phase 1 (Task 1-1) | ✅ Addressed |
| AC-4 | Phase 1 (Tasks 1-2, 1-3) | ✅ Addressed |
| AC-5 | Phase 2 (Tasks 2-1, 2-2, 2-3) | ✅ Addressed |
| AC-6 | Phase 2 (Task 2-2) | ✅ Addressed |
| AC-7 | Phase 6 (Tasks 6-1, 6-2, 6-3, 6-4) | ✅ Addressed |
| AC-8 | Phase 5 (Tasks 5-1, 5-2, 5-3) | ✅ Addressed |
| AC-9 | Architecture Boundaries table above | ✅ Addressed |
| AC-10 | Open Decisions section above | ✅ Addressed |