# Spec 006 — Quickstart and Verification Runbook

> Scope: Manual verification runbook for Spec 006 Phases 1-6. This document is documentation-only and validates the intended behavior without introducing runtime implementation requirements.

---

## Objective

Provide a single manual walkthrough that verifies the approved defaults, workflow contracts, and documentation outputs defined across Spec 006 Phases 1-6.

---

## Normative Inputs and Defaults

- Research: `specs/006-housekeeping-memory-and-bootstrap/research.md`
- Plan: `specs/006-housekeeping-memory-and-bootstrap/plan.md`
- Data model: `specs/006-housekeeping-memory-and-bootstrap/data-model.md`
- Prior tasks:
  - `specs/006-housekeeping-memory-and-bootstrap/tasks/001-metadata-migration.md`
  - `specs/006-housekeeping-memory-and-bootstrap/tasks/002-bootstrap-workflow.md`
  - `specs/006-housekeeping-memory-and-bootstrap/tasks/003-housekeeping-workflow.md`
  - `specs/006-housekeeping-memory-and-bootstrap/tasks/004-agents-cleanup.md`
  - `specs/006-housekeeping-memory-and-bootstrap/tasks/005-install-ux.md`
- Sync state path: `.ai/housekeeping/sync-state.json`
- Default memory path: `specs/memory`
- Metadata enforcement: strict for new artifacts; warn-and-migrate for legacy artifacts
- `compose_agent` integration assumption: use existing supported fields first and inject library/spec context through `stepInstructions`
- Approval default: task-scoped first; session-scoped optional; standing approvals hard-expire at 30 days while allowing an in-flight task to finish
- qmd index path: project-local by default

---

## Verification Principles

This runbook verifies documentation contracts, expected operator-visible behavior, and approval boundaries.

- It does **not** require runtime implementation in this spec slice.
- Read-only observations may always be checked.
- Any mutating example below remains subject to explicit approval or active contract coverage.
- If a command or UI surface is not implemented yet, record the gap as a verification note rather than inventing behavior.

---

## Scenario Matrix

| Scenario | Optional Tooling | Approval State | Expected Outcome |
|---|---|---|---|
| A | none enabled | no contract | Bootstrap and housekeeping fall back to direct file-based behavior; no qmd/codegraph sync actions are available |
| B | qmd enabled only | no contract | qmd search/drift checks are read-only; any index build/rebuild requires approval |
| C | qmd + codegraph enabled | task-scoped contract | Approved task may load memory and perform sync/repair actions covered by the contract without re-prompting |
| D | qmd + codegraph enabled | session-scoped contract | Same behavior as C for the current session; later sessions require fresh approval |
| E | qmd + codegraph enabled | standing contract near expiry | In-flight task may finish; next task must renew after 30-day hard expiry |
| F | Obsidian enabled with any other mix | no contract | Vault discovery remains advisory/read-only; vault or config writes still require explicit approval |

Use Scenario C for the main walkthrough, then spot-check A/B/F for fallback and optionality.

---

## Recommended Manual Walkthrough

### 1. Verify install-time UX choices

Goal: validate Phase 6 documentation for optional `ob`, qmd, and codegraph setup.

Checklist:

1. Start `ai-setup init` in a test project or dry-run environment.
2. Confirm the wizard asks for the project memory path before or alongside optional integrations.
3. Confirm the default memory path is `specs/memory`.
4. Confirm the wizard presents Obsidian (`ob`), qmd, and codegraph as **optional** integrations.
5. Confirm the wizard distinguishes:
   - read-only discovery
   - config writes during setup
   - later runtime reads
   - later approval-gated maintenance writes
6. Enable all optional integrations for the primary walkthrough.

Expected result:

- Tooling remains opt-in.
- qmd and codegraph enablement do **not** imply blanket approval for future index writes.
- Obsidian is described as discovery/config awareness, not the primary retrieval engine.

Relevant ACs: AC-5, AC-6, AC-7, AC-9.

### 2. Verify memory path configuration

Goal: validate the canonical memory-path default and setup persistence behavior.

Checklist:

1. Accept the default memory path of `specs/memory`.
2. Confirm the chosen path is the project-local path persisted by setup.
3. Confirm qmd scope includes the configured memory path when qmd is enabled.
4. Confirm codegraph configuration references the same project-local memory path when codegraph is enabled.

Expected result:

- `specs/memory` remains the default canonical path.
- Project-local configuration is preferred over any external or vault-specific path.
- No alternate default is introduced.

Relevant ACs: AC-1, AC-7, AC-9.

### 3. Verify bootstrap behavior

Goal: validate the Phase 3 sequence: discovery → read-only drift check → approval evaluation → context load → bootstrap report.

Checklist:

1. Start a chain or equivalent task execution flow.
2. Confirm bootstrap discovery resolves:
   - memory path
   - `specs/` availability
   - qmd availability/status
   - codegraph availability/status
   - sync-state presence at `.ai/housekeeping/sync-state.json`
3. Confirm drift and staleness checks are reported as read-only.
4. If no contract exists, confirm approval is requested before approval-gated memory load or sync/repair.
5. Confirm bootstrap can continue in degraded mode if tooling is unavailable.
6. Confirm a bootstrap report summarizes loaded context, stale items, contract status, and deferred maintenance.

Expected result:

- Bootstrap is deterministic and non-blocking.
- Read-only drift checks happen before any mutation.
- Context loading and sync actions remain approval-gated.

Relevant ACs: AC-1, AC-5, AC-6, AC-9.

### 4. Verify housekeeping behavior

Goal: validate pre-task, inline, and post-task housekeeping checkpoints from Phase 4.

Checklist:

1. Advance to a next step and confirm pre-task housekeeping checks contract status and drift/read-only freshness.
2. Complete a meaningful workflow transition and confirm inline memory extraction is proposed.
3. Complete the task/step and confirm post-task housekeeping performs:
   - mandatory memory extraction sweep
   - cleanup proposal
   - sync proposal
   - approval gate before any writes
4. Confirm deferred maintenance is reported when proposals are rejected or postponed.

Expected result:

- Housekeeping runs at the documented checkpoints.
- Memory extraction may happen inline and also in the post-task sweep.
- Cleanup, sync, and sync-state updates are explicit proposals rather than silent actions.

Relevant ACs: AC-2, AC-3, AC-5, AC-6, AC-9.

### 5. Verify approval-first contract behavior

Goal: validate contract scope, reuse, expiry, and revocation.

Checklist:

1. Start with no contract and confirm the system requests the smallest scope first, preferring task-scoped approval.
2. Approve a task-scoped or session-scoped contract covering `memory_load`, `memory_write`, `qmd_sync`, and `codegraph_sync`.
3. Repeat a sync-eligible action and confirm the covered action does not re-prompt within scope.
4. Revoke or expire the contract mid-task and confirm the in-flight task may finish while future work becomes deferred or requires renewal.
5. Confirm standing approvals are documented with a hard 30-day expiry.

Expected result:

- Approval-first behavior is preserved.
- Contracts reduce prompt repetition without allowing silent mutation outside approved scope.
- Revocation/expiry is surfaced in future reporting.

Relevant ACs: AC-2, AC-5, AC-6.

### 6. Verify stale/drift detection and repair proposals

Goal: validate read-only detection plus approval-gated repair behavior.

Checklist:

1. Change a source file or otherwise simulate index drift.
2. Re-run bootstrap or pre-task housekeeping.
3. Confirm the system reports qmd/codegraph state as `fresh`, `stale`, `stale_acked`, `disabled`, `unavailable`, or `unknown`.
4. Reject a proposed repair once and confirm the stale fingerprint is treated as acknowledged for repeat reporting.
5. Confirm the stale condition still appears in reports even when duplicate prompts are suppressed.
6. Confirm the repair path is a full re-index proposal, not a silent background mutation.

Expected result:

- Drift detection is always read-only.
- Rejection produces deferred maintenance behavior and `staleAcked` semantics at the sync-state level.
- Future reports warn about stale indexes and potential context mismatch.

Relevant ACs: AC-5, AC-6.

### 7. Verify report-only doctor behavior for stray `AGENTS.md`

Goal: validate the Phase 5 cleanup contract.

Checklist:

1. Run `ai-setup doctor` or review the intended doctor output contract.
2. Confirm stray `AGENTS.md` files below repo root are reported as findings.
3. Confirm the canonical root `AGENTS.md` is not reported as stray.
4. Confirm the output stays report-only by default.
5. Confirm no silent deletion or migration is implied.
6. Confirm replacement guidance points to `library/specs-agents/` plus current `compose_agent`/`stepInstructions` usage.

Expected result:

- Root-only `AGENTS.md` remains the target state.
- Doctor behavior reports problems without mutating files.
- Cleanup remains a separate approved action.

Relevant ACs: AC-8, AC-9.

### 8. Verify metadata and memory-model expectations

Goal: validate the Phase 1-2 documentation contracts at a schema/spec level.

Checklist:

1. Review the metadata schema in `data-model.md`.
2. Confirm new artifacts require strict metadata, including `schema_version`, `artifact_type`, `id`, timestamps, authorship, and related references.
3. Confirm legacy artifacts are warn-and-migrate rather than blocked on read.
4. Review the memory note format and confirm it uses:
   - YAML frontmatter
   - append-only timeline entries
   - timestamps
   - supersession via `supersedes`
5. Confirm compaction guidance warns at `>= 200` lines and proposes compaction at `>= 500` lines.
6. Confirm sync-state references the canonical path `.ai/housekeeping/sync-state.json`.

Expected result:

- Metadata rules are normalized across specs, tasks, memory, handoffs, and maintenance artifacts.
- Memory files preserve history instead of destructive edits.
- Compaction is archival/advisory, not deletion.

Relevant ACs: AC-3, AC-4, AC-5.

---

## Acceptance-Criteria Mapping

| AC | Covered By This Runbook | Notes |
|---|---|---|
| AC-1 | Sections 2-3 | Confirms default memory path and deterministic bootstrap sequence |
| AC-2 | Sections 4-5 | Confirms pre-task, inline, and post-task housekeeping with approval gates |
| AC-3 | Section 8 | Confirms append-only memory model and compaction thresholds |
| AC-4 | Section 8 | Confirms standardized metadata expectations across artifact classes |
| AC-5 | Sections 1, 3, 5, 6 | Confirms contract types, drift checks, expiry, and revocation behavior |
| AC-6 | Sections 1, 5, 6 | Confirms qmd/codegraph drift detection, repair proposals, and rejection behavior |
| AC-7 | Sections 1-2 | Confirms install-time UX for optional tooling and memory path configuration |
| AC-8 | Section 7 | Confirms root-only `AGENTS.md` target state and report-only doctor behavior |
| AC-9 | Sections 1, 3, 4, 7, 8 | Confirms ownership boundaries implied by wizard, runtime, and spec-layer checks |
| AC-10 | Not directly verified here | Plan-level decision tracking remains a plan-review activity, not a quickstart walkthrough |

---

## Suggested Evidence Capture Template

For each walkthrough step, record:

- scenario used
- command or UI path exercised
- observed output or note
- approval state at the time
- whether behavior matched the spec
- any gap to log as follow-up

Example:

| Step | Scenario | Observation | Matches Spec? | Notes |
|---|---|---|---|---|
| Bootstrap report | C | Report listed memory path, stale qmd, deferred codegraph repair | Yes | Approval prompt appeared before sync |
| Doctor stray AGENTS | A | Stray files reported, no deletion | Yes | Root file excluded |

---

## Validation Notes for This Spec Slice

- This runbook is documentation/spec oriented and should be validated by manual walkthrough plus document review.
- No automated checks are inherently applicable to this Phase 7 documentation slice.
- If implementation surfaces do not yet exist, verify the expected behavior against the Phase 1-6 documents and record gaps without expanding scope.
