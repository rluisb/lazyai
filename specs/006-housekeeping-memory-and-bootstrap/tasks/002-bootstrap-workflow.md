# Spec 006 — Task 002: Bootstrap Workflow

> Scope: Specify the deterministic bootstrap workflow that runs at chain start. This task is documentation-only and preserves approval-first behavior for all mutating maintenance actions.

---

## Objective

Define the bootstrap sequence for project-context discovery, read-only drift inspection, approval-gated context loading, and bootstrap reporting so the orchestrator can begin work with the best available context without silently mutating project state.

---

## Normative Inputs and Defaults

- Research: `specs/006-housekeeping-memory-and-bootstrap/research.md`
- Plan: `specs/006-housekeeping-memory-and-bootstrap/plan.md`
- Data model: `specs/006-housekeeping-memory-and-bootstrap/data-model.md`
- Sync state path: `.ai/housekeeping/sync-state.json`
- Metadata enforcement: strict for new artifacts; warn/migrate for legacy artifacts
- `compose_agent` integration assumption: use existing `stepInstructions` / current supported fields first
- Approval model default: task-scoped approval; optional session-scoped; standing approvals hard-expire at 30 days while allowing an in-flight task to finish
- qmd index path: project-local by default

---

## Bootstrap Sequence

Bootstrap runs as a deterministic pre-step before the first chain step executes.

1. **Discovery**
2. **Read-only drift and staleness inspection**
3. **Approval and contract evaluation**
4. **Context load**
5. **Bootstrap report**

Bootstrap is advisory and non-blocking unless a later spec explicitly declares a hard stop.

---

## 1. Discovery

### 1.1 Discovery Goals

Bootstrap discovery answers only these questions:

- Where should memory/spec context be searched?
- Which optional integrations are enabled and reachable?
- Which read-only checks can run safely before asking for approval?

### 1.2 Discovery Sequence

Bootstrap resolves paths and integration availability in this order:

1. **Project-local configured memory path** from project config when present.
2. **Spec-local default memory path** under `specs/memory` when no explicit override exists.
3. **Spec corpus root** under `specs/` for related research/plan/task lookup.
4. **qmd project-local index path** when qmd is enabled.
5. **codegraph project-local data path** when codegraph is enabled.
6. **Obsidian/vault hints** only for discovery/config awareness when enabled; Obsidian is not the primary retrieval engine.

### 1.3 Memory Path Resolution Rules

- The canonical memory path remains project-local and should resolve before any external or user-environment-specific path.
- If no explicit memory path is configured, bootstrap falls back to `specs/memory`.
- If the memory directory does not exist, bootstrap reports "memory path unavailable" and continues without memory context.
- Missing paths do not trigger directory creation during bootstrap; creation would be a mutating maintenance action and is out of scope here.

### 1.4 Discovery Outputs

Discovery produces a read-only working set:

- resolved memory path or absence reason
- `specs/` availability status
- qmd enabled/disabled/unavailable status
- codegraph enabled/disabled/unavailable status
- sync-state presence/absence at `.ai/housekeeping/sync-state.json`
- any obvious contract record needed for approval evaluation

---

## 2. Read-Only Drift and Staleness Inspection

### 2.1 Allowed Checks

The following checks are always allowed because they are read-only:

- existence/readability of memory/spec paths
- presence of `.ai/housekeeping/sync-state.json`
- qmd freshness inspection against source mtimes, hashes, tool state, or last index timestamp
- codegraph freshness inspection against source mtimes, hashes, tool state, or last index timestamp
- inspection of existing `staleAcked` and repair proposal records in sync state
- detection of expired or missing maintenance contract coverage

### 2.2 Forbidden Actions During Read-Only Inspection

The following are explicitly **not** part of drift inspection:

- rebuilding qmd indexes
- rebuilding codegraph indexes
- creating or updating sync-state records
- writing memory artifacts
- deleting or cleaning up files

Those remain mutating maintenance actions and require approval or active contract coverage.

### 2.3 Drift Classification

Bootstrap should classify each integration as one of:

- `fresh`
- `stale`
- `stale_acked`
- `disabled`
- `unavailable`
- `unknown`

Interpretation rules:

- `stale` means drift is newly detected and no matching acknowledged fingerprint suppresses a prompt.
- `stale_acked` means the same stale fingerprint was previously rejected and should be reported without re-prompting.
- `unavailable` means the integration is configured or expected but cannot be used in the current environment.

---

## 3. Approval Request Behavior vs Active Maintenance Contract

### 3.1 Approval Gate

After discovery and read-only checks, bootstrap evaluates whether the next actions require approval.

Approval is required for:

- memory/context loading that materially expands task context
- qmd sync or re-index
- codegraph sync or re-index
- any sync-state writeback
- any repair or cleanup action

### 3.2 Contract Evaluation Order

Bootstrap should evaluate approval coverage in this order:

1. Active **task-scoped** contract for the current task or workflow run
2. Active **session-scoped** contract for the current session
3. Active **standing** contract that has not hard-expired
4. No covering contract → request per-action approval

### 3.3 Contract Behavior Rules

- Task-scoped approval is the default preferred approval mode.
- Session-scoped approval is optional when the user wants broader coverage for the current session.
- Standing approval hard-expires at 30 days.
- If a standing approval expires during an in-flight task, bootstrap for that already-started task may honor it only through task completion; the next task must request fresh approval.
- Contract scope must be action-specific. Missing permissions still require per-action approval.

### 3.4 Request Behavior

If no active contract covers the next mutating or approval-gated context action, bootstrap should:

1. summarize what it wants to load or repair
2. distinguish read-only findings from proposed mutations
3. ask for the smallest approval scope first, preferring task-scoped approval
4. proceed only with the approved subset of actions

If approval is denied:

- bootstrap proceeds without the denied context or maintenance action
- stale state is reported
- future prompts follow `staleAcked` suppression rules when the same fingerprint remains unchanged

---

## 4. Context Loading Order

Bootstrap loads context only after approval or contract coverage is confirmed.

### 4.1 Required Loading Order

1. **Memory and spec artifacts first**
   - load the most relevant memory notes and spec artifacts for the incoming task
   - if qmd is enabled and available, use it as the preferred retrieval layer for markdown corpora
   - if qmd is unavailable, fall back to direct file reads of the relevant memory/spec documents
2. **qmd-derived retrieval context second**
   - use qmd results to narrow which markdown artifacts are loaded into task context
   - qmd read/search is read-only; index rebuilds are not part of this step unless separately approved
3. **codegraph context third**
   - load structural code context only when the task is code-relevant and codegraph is enabled/available
   - stale codegraph indexes may be reported, but not rebuilt unless approved or contract-covered

### 4.2 Loading Principles

- Prefer the smallest relevant context set.
- Do not block bootstrap on qmd or codegraph availability.
- Do not invent alternate storage or retrieval paths beyond the approved defaults.
- Treat memory/spec loading and sync repair as separate decisions when contract coverage differs.

---

## 5. Bootstrap Report Contents

The bootstrap report should be concise, deterministic, and sufficient for the chain to start with explicit awareness of gaps.

### 5.1 Required Report Sections

- **Discovery summary**
  - resolved memory path
  - spec path availability
  - qmd/codegraph enabled and availability status
- **Drift summary**
  - fresh vs stale vs stale_acked status for qmd/codegraph
  - sync-state presence and any relevant snapshot age
- **Approval/contract summary**
  - active contract type and expiry when present
  - actions covered by contract
  - approvals requested, granted, denied, or deferred
- **Context loaded**
  - memory files loaded
  - spec/task/research docs loaded
  - code context loaded or skipped
- **Deferred maintenance**
  - stale indexes not repaired
  - rejected loads or syncs
  - expired contract follow-up needed for the next task
- **Fallbacks used**
  - qmd unavailable → direct file reads
  - codegraph unavailable → no structural code context
  - missing memory path → proceed without memory context

### 5.2 Reporting Constraints

- Bootstrap report generation is read-only.
- Bootstrap report does not itself imply approval for later mutation.
- Any missing context must be surfaced explicitly rather than hidden.

---

## 6. Orchestrator Integration at Chain Start

### 6.1 Trigger Point

Bootstrap is owned by the orchestrator runtime and runs when `start_chain` is invoked, before the first chain step executes.

### 6.2 Runtime Boundary

The orchestrator runtime owns:

- invoking bootstrap at chain start
- checking active contract coverage
- deciding whether approval prompts are required before context load or sync
- attaching bootstrap results to the chain/session state for downstream visibility

The orchestrator runtime does **not** own:

- install-time tool detection UX
- qmd or codegraph implementation internals
- memory/spec schema definition
- new `compose_agent` API design beyond current supported fields

### 6.3 `compose_agent` Integration

- Bootstrap should inject discovered library/spec context through existing `compose_agent` inputs first.
- Preferred mechanism: pass summarized bootstrap context through `stepInstructions` and any currently supported fields rather than defining new API surface in this phase.
- If bootstrap context is partial, the injected instructions should say so explicitly.

---

## 7. Fallback and Non-Blocking Behavior

Bootstrap must remain non-blocking when optional integrations are unavailable.

### 7.1 Fallback Rules

- **No memory path found** → continue without memory context and report the gap.
- **qmd unavailable** → use direct file reads for relevant markdown artifacts when approved.
- **codegraph unavailable** → continue without structural code context.
- **sync-state missing** → treat drift status as `unknown` until a later approved maintenance action creates state.
- **approval denied** → continue with reduced context and note intentional incompleteness.
- **contract expired before bootstrap** → request fresh approval before any covered action.

### 7.2 Non-Blocking Principle

Bootstrap failure, partial success, or missing optional tooling should not prevent `start_chain` from creating the chain unless a future spec explicitly introduces a hard prerequisite.

---

## 8. Boundary Summary

| Concern | Orchestrator Runtime | ai-setup CLI / Wizard | Spec / Convention Layer |
|---|---|---|---|
| Chain-start bootstrap hook | Owns | Does not own | Defines expected behavior |
| Memory path defaults/config surface | Reads config | Owns install-time configuration UX | Defines canonical defaults |
| qmd/codegraph drift inspection policy | Executes read-only checks | Does not own runtime checks | Defines approval-first rules |
| qmd/codegraph sync execution policy | Enforces contract/approval | Describes install-time option | Defines sync-state semantics |
| `compose_agent` context injection | Uses existing fields | Does not own runtime prompt assembly | Defines what context should be injected |

---

## Completion Notes

- This workflow intentionally stops short of runtime implementation details.
- Later phases may reference this document, but this task does not define AGENTS cleanup or install-time UX behavior beyond the boundaries noted above.
