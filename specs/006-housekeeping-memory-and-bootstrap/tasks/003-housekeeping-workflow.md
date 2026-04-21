# Spec 006 — Task 003: Housekeeping Workflow

> Scope: Specify pre-task, inline, and post-task housekeeping workflows. This task is documentation-only and preserves approval-first behavior for memory writes, cleanup, and qmd/codegraph sync actions.

---

## Objective

Define the housekeeping workflow that keeps project context, memory artifacts, and technical indexes aligned across task execution without allowing silent maintenance mutations.

---

## Normative Inputs and Defaults

- Research: `specs/006-housekeeping-memory-and-bootstrap/research.md`
- Plan: `specs/006-housekeeping-memory-and-bootstrap/plan.md`
- Data model: `specs/006-housekeeping-memory-and-bootstrap/data-model.md`
- Metadata migration checklist: `specs/006-housekeeping-memory-and-bootstrap/tasks/001-metadata-migration.md`
- Sync state path: `.ai/housekeeping/sync-state.json`
- Metadata enforcement: strict for new artifacts; warn/migrate for legacy artifacts
- Approval model default: task-scoped approval; optional session-scoped; standing approvals hard-expire at 30 days while allowing an in-flight task to finish
- qmd index path: project-local by default

---

## Housekeeping Modes

Housekeeping is split into three checkpoints:

1. **Pre-task housekeeping** — before the next task or chain step executes
2. **Inline memory extraction checkpoint** — after RPI/workflow transitions
3. **Post-task housekeeping** — after a task or chain step completes

All three checkpoints must preserve the same approval-first model.

---

## 1. Pre-Task Housekeeping

### 1.1 Purpose

Pre-task housekeeping ensures the next step begins with the freshest approved context available, while keeping drift inspection read-only until approval is granted.

### 1.2 Sequence

1. **Check contract status**
   - confirm whether task-scoped, session-scoped, or standing approval is active
   - verify expiry and permitted actions
2. **Run read-only drift inspection**
   - inspect qmd freshness
   - inspect codegraph freshness
   - inspect sync-state and any `staleAcked` records
   - inspect whether newer memory/spec artifacts exist that may change the plan
3. **Evaluate required context actions**
   - determine whether memory/spec loads are needed for the next step
   - determine whether stale indexes should be proposed for sync
4. **Request approval when needed**
   - prefer task-scoped approval when new mutating or approval-gated actions are required
5. **Load approved context and report gaps**

### 1.3 Pre-Task Read-Only Checks

Pre-task housekeeping may always:

- inspect relevant memory/spec artifact timestamps or query results
- inspect qmd/codegraph freshness
- inspect sync-state contents
- detect expired contracts
- detect stale but previously acknowledged drift

Pre-task housekeeping may **not** without approval or contract coverage:

- write memory entries
- create or update `.ai/housekeeping/sync-state.json`
- rebuild qmd indexes
- rebuild codegraph indexes
- clean up temporary or stray artifacts

### 1.4 Pre-Task Output

Pre-task housekeeping should report:

- contract status and expiry
- approved actions available under contract
- newly discovered context that may affect the task
- stale vs stale_acked vs fresh index state
- deferred maintenance items carried forward

---

## 2. Inline Memory Extraction Checkpoint

### 2.1 Trigger

Inline memory extraction runs after meaningful workflow transitions, especially after RPI transitions such as:

- Research → Plan
- Plan → Implement
- Implement → Review
- any equivalent orchestrator workflow transition that completes a meaningful step

### 2.2 Purpose

This checkpoint captures lessons while they are fresh instead of relying only on end-of-task recall.

### 2.3 Extraction Scope

Inline extraction should look for:

- new discoveries that alter future setup or workflow choices
- decisions that narrow future implementation paths
- warnings, pitfalls, or patterns that should persist
- corrections or supersessions of earlier guidance

### 2.4 Approval Behavior

- If an active contract covers `memory_write`, proposed memory entries may be written immediately.
- If no contract covers `memory_write`, the system should stage the proposed entries for later approval and include them in post-task housekeeping.
- Inline extraction itself may happen without mutation; only writing the extracted memory requires approval.

### 2.5 Memory Format Constraints

- New memory artifacts must follow the standardized metadata schema and append-only timeline model.
- Legacy memory artifacts may be read, but any migration or metadata backfill remains a mutating action requiring approval.
- Conflicts are handled by appending superseding entries rather than rewriting older entries.

---

## 3. Post-Task Housekeeping

### 3.1 Purpose

Post-task housekeeping performs the mandatory final sweep for durable memory capture, cleanup proposals, and technical sync proposals after the task outcome is known.

### 3.2 Sequence

1. **Mandatory memory extraction sweep**
   - capture anything missed by inline extraction
   - convert staged lessons/decisions into approval-ready proposals
2. **Cleanup proposal**
   - identify temporary artifacts, workspace organization opportunities, and related housekeeping work
3. **Sync proposal**
   - run read-only drift checks again
   - determine whether qmd/codegraph are stale
   - determine whether sync-state repair or update is needed
4. **Approval gate**
   - request approval for memory writes, cleanup, sync, or sync-state updates unless covered by active contract
5. **Report outcome**
   - summarize what was written, what was deferred, and what remains stale

### 3.3 Approval-First Cleanup/Sync/Memory Proposals

Post-task housekeeping must treat the following as explicit proposals, not automatic actions:

- memory writes or compaction
- qmd sync/re-index
- codegraph sync/re-index
- sync-state creation or update
- cleanup of temporary artifacts
- repair of stale maintenance state

Each proposal should include:

- why the action is needed
- which files or tool states are affected
- whether the current contract already covers it
- what will remain deferred if the user rejects it

### 3.4 Deferred Maintenance Rules

If the user rejects a proposed post-task maintenance action:

- execution still completes
- the stale or deferred condition is recorded for future reporting
- duplicate prompts are suppressed when `staleAcked` already matches the same fingerprint
- the next bootstrap or housekeeping pass should continue surfacing the deferred condition until new drift appears or approval is later granted

---

## 4. Handling Special Conditions

### 4.1 Stale Indexes

- Stale indexes may always be detected in read-only mode.
- Repair requires approval or contract coverage.
- If repair is deferred, housekeeping reports the risk of context mismatch rather than hiding it.

### 4.2 Rejected Syncs

- Rejected qmd/codegraph sync proposals should result in a `staleAcked` record for the current fingerprint when sync-state writes are approved or already covered.
- If recording that acknowledgment is itself not approved, the system should still report the rejection but may not persist suppression state.
- Rejected syncs never justify silent retries.

### 4.3 `stale_acked` Behavior

- `staleAcked` suppresses repeat prompts for the same unchanged stale fingerprint.
- Once the underlying source fingerprint changes, the prior acknowledgment no longer suppresses prompts.
- `stale_acked` is a reporting and prompt-suppression state, not proof of freshness.

### 4.4 Expired Contracts

- If a contract expires before pre-task housekeeping begins, new approval is required for any covered mutating action.
- If a standing approval expires during an in-flight task, the current task may finish, but post-task housekeeping for new actions should request renewal unless the original in-flight allowance still covers those actions.
- The next task must not inherit the expired contract.

### 4.5 Deferred Maintenance

Deferred maintenance includes any approved-later work such as:

- memory writes staged but not yet authorized
- qmd/codegraph sync proposals rejected or postponed
- cleanup proposals left pending
- metadata migrations or legacy backfills intentionally delayed

Deferred maintenance should be surfaced in future bootstrap and housekeeping reports until resolved or explicitly superseded.

---

## 5. Orchestrator Hook Integration

### 5.1 Pre-Step Hook

Before executing the next step, the orchestrator runtime should run pre-task housekeeping as part of the `advance_chain` flow.

### 5.2 Post-Step Hook

After a step completes, the orchestrator runtime should run post-task housekeeping as part of the same `advance_chain` lifecycle.

### 5.3 Inline Checkpoint Hook

When a workflow transition completes, the orchestrator runtime should trigger the inline memory extraction checkpoint and either:

- write memory immediately if covered by contract, or
- stage proposals for later approval in post-task housekeeping

### 5.4 Non-Blocking Behavior

Housekeeping should not block step completion solely because optional integrations are stale, unavailable, or deferred.

---

## 6. Boundary Summary

| Concern | Orchestrator Runtime | ai-setup CLI / Wizard | Spec / Convention Layer |
|---|---|---|---|
| Pre/post-task housekeeping hooks | Owns runtime invocation | Does not own runtime execution | Defines workflow expectations |
| Contract expiry and approval checks | Owns enforcement during execution | May describe options at install time | Defines contract semantics |
| qmd/codegraph drift inspection | Executes read-only checks | Owns install/config UX only | Defines read-only vs mutating boundaries |
| Memory file schema and append-only rules | Uses schema | Does not own schema | Owns format and migration conventions |
| Cleanup proposal policy | Surfaces proposals | May expose future doctor/wizard guidance | Defines approval-first requirement |
| Legacy metadata migration policy | Reports gaps during workflow | May expose doctor output | Owns strict-new / warn-migrate legacy rules |

---

## 7. Consistency Rules for This Spec Slice

- qmd and codegraph sync/index actions are mutating maintenance actions unless the operation is purely a read-only drift check.
- Approval-first behavior is preserved for bootstrap, pre-task housekeeping, inline memory writes, post-task housekeeping, cleanup, and sync.
- This document specifies workflow behavior only; it does not implement runtime hooks, wizard screens, or doctor commands.
- Later phases may build on these workflows, but this task does not define AGENTS cleanup or install-time UX beyond the required boundaries.

---

## Completion Notes

- This task intentionally pairs inline memory extraction with the mandatory post-task sweep rather than choosing one exclusively.
- Deferred maintenance is first-class: rejection or expiry should be visible in future reports without causing silent mutation.
