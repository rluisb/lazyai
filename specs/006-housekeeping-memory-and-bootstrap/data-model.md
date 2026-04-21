# Spec 006 — Data Model (Phases 1-2)

> Scope: Foundation contracts for memory artifacts, metadata normalization, maintenance approvals, and qmd/codegraph sync state. This document is normative for new artifacts and migration guidance for legacy artifacts.

---

## Approved Defaults Applied

| Default | Decision |
|---|---|
| Sync state path | `.ai/housekeeping/sync-state.json` |
| Metadata enforcement | Strict for new artifacts; warn-and-migrate for legacy artifacts |
| `compose_agent` integration assumption | Use the existing `compose_agent` API first; inject library/spec context through existing `stepInstructions` and current supported fields |
| Standing approval TTL | Default to task-scoped approval; optional session-scoped; persistent standing approvals hard-expire at 30 days, but an in-flight task may finish |
| `ai-setup doctor` behavior for stray `AGENTS.md` | Report-only by default; no silent deletion |
| qmd index location | Project-local by default |

---

## 1. Standardized Metadata Schema

### 1.1 Scope

This schema applies to:

- Spec directories and their primary documents (`research.md`, `plan.md`, optional `spec.md`)
- RPI task files under `tasks/`
- Memory files under `specs/memory/`
- Handoff artifacts under `specs/memory/handoffs/`
- Chain/workflow export artifacts when serialized to markdown or JSON
- Maintenance contract records and sync-state records

### 1.2 Naming Rules

- YAML / markdown frontmatter uses **snake_case**.
- JSON / programmatic payloads use **camelCase**.
- Field removals are forbidden without a migration.
- New fields are additive and must preserve backward compatibility.

### 1.3 Core Fields

| Field | Type | Required | Applies To | Notes |
|---|---|---:|---|---|
| `schema_version` | integer | Yes | all artifacts | Start at `1` |
| `artifact_type` | string | Yes | all artifacts | Examples: `spec_plan`, `spec_task`, `memory_note`, `handoff`, `maintenance_contract`, `sync_state_snapshot` |
| `id` | string | Yes | all artifacts | Stable identifier within the project |
| `title` | string | Yes | markdown artifacts | Human-readable title |
| `ticket_number` | string \| null | No | task/spec/handoff/memory | External ticket or issue reference |
| `status` | string | No | all artifacts | Suggested values: `draft`, `active`, `done`, `superseded`, `archived` |
| `created_at` | datetime | Yes | all artifacts | ISO-8601 UTC |
| `updated_at` | datetime | Yes | all artifacts | ISO-8601 UTC |
| `created_by` | string | Yes | all artifacts | Human or agent identifier |
| `updated_by` | string | Yes | all artifacts | Human or agent identifier |
| `risk_level` | string | No | task/spec/memory/contract | `low`, `medium`, `high` |
| `size_points` | integer \| null | No | planning/task artifacts | Optional numeric sizing |
| `complexity_level` | string \| null | No | all artifacts | `low`, `medium`, `high` |
| `session_id` | string \| null | No | all artifacts | Session that produced or updated the artifact |
| `workflow_id` | string \| null | No | workflow-bound artifacts | Logical workflow family, e.g. `rpi`, `bootstrap`, `housekeeping` |
| `workflow_run_id` | string \| null | No | workflow-bound artifacts | Specific execution run |
| `team_id` | string \| null | No | all artifacts | Owning team or repo area |
| `chain_id` | string \| null | No | orchestrator/handoff/task/memory | Chain identifier when present |
| `owner_agent` | string \| null | No | all artifacts | Responsible agent role |
| `assignee` | string \| null | No | all artifacts | Human or agent currently assigned |
| `step_ids` | array[string] | No | task/memory/handoff/workflow artifacts | Step IDs already completed or referenced |
| `workflow_steps` | array[string] | No | task/spec/workflow artifacts | Human-readable workflow step labels |
| `related_document_refs` | array[string] | No | all artifacts | Relative repo paths to linked specs, rules, code, or archives |
| `approval_scope` | string \| null | No | contracts/tasks | `per_action`, `task_scoped`, `session_scoped`, `standing` |
| `approval_expires_at` | datetime \| null | No | contracts | Required when approval is time-bounded |
| `legacy_metadata_gaps` | array[string] | No | migrated legacy artifacts | Records missing fields during migration |
| `migration_notes` | array[string] | No | migrated legacy artifacts | Documents conversion assumptions |

### 1.4 Validation Rules

- New artifacts must include all required fields at creation time.
- `updated_at` must be greater than or equal to `created_at`.
- `workflow_run_id` must not appear without `workflow_id`.
- `chain_id` should be present when an artifact was produced during orchestrator execution.
- `owner_agent` and `assignee` may match, but both should be explicit for handoffs and tasks.
- `related_document_refs` must be repo-relative paths.
- JSON producers should emit camelCase equivalents of the same schema.

### 1.5 Enforcement Model

| Artifact Age | Enforcement |
|---|---|
| New artifacts created after Spec 006 adoption | **Strict**: invalid metadata blocks creation/update until required fields are supplied |
| Legacy artifacts updated in place | **Warn + migrate**: do not block reads; add missing metadata when the file is next meaningfully edited |
| Legacy artifacts left untouched | **Warn only**: surface gaps in doctor/reporting and migration checklists |

Legacy migration must preserve approval-first behavior: metadata backfills are mutating actions and require user approval when applied to existing files.

---

## 2. Memory File Format

### 2.1 Purpose

Memory files capture durable project learnings without destructive overwrites. Conflicts reconcile by appending newer entries that supersede older entries while preserving the historical trail.

### 2.2 File Shape

Memory files use YAML frontmatter followed by an append-only markdown timeline body.

```yaml
---
schema_version: 1
artifact_type: memory_note
id: mem-auth-timeout
title: Auth service timeout behavior
ticket_number: PROJ-456
status: active
created_at: 2026-04-17T10:00:00Z
updated_at: 2026-04-17T14:30:00Z
created_by: builder-agent
updated_by: builder-agent
risk_level: low
complexity_level: low
session_id: sess-abc123
workflow_id: rpi
workflow_run_id: run-001
team_id: platform
chain_id: chain-xyz
owner_agent: builder
assignee: builder
step_ids:
  - step-3
  - step-4
related_document_refs:
  - specs/004-go-migration/plan.md
  - specs/rules/security.md
compacted_from: null
compaction_status: none
---
```

```markdown
## [2026-04-17T10:00:00Z] builder-agent
entry_id: mem-auth-timeout#001
entry_type: discovery
supersedes: null
related_document_refs:
- specs/004-go-migration/plan.md

Discovered: The auth service has a 5-second timeout on webhook processing.
The timeout is configurable via `AUTH_WEBHOOK_TIMEOUT_MS`.

---

## [2026-04-17T14:30:00Z] builder-agent
entry_id: mem-auth-timeout#002
entry_type: correction
supersedes: mem-auth-timeout#001
related_document_refs:
- specs/004-go-migration/plan.md
- specs/004-go-migration/implementation-plan.md

Updated: The default timeout is now 10 seconds and the preferred env var is
`AUTH_WEBHOOK_TIMEOUT`.
```

### 2.3 Timeline Rules

- Entries append in chronological order.
- Existing entry bodies are not edited except for approved compaction.
- New information that changes prior guidance uses `supersedes` rather than deleting old text.
- Entries may represent `discovery`, `decision`, `warning`, `pattern`, `correction`, or `promotion`.
- Related document references may appear at file level, entry level, or both.

### 2.4 Compaction Strategy

| Threshold | Behavior |
|---|---|
| `< 200` lines | No action |
| `>= 200` lines | Warn at the next housekeeping pass |
| `>= 500` lines | Propose compaction as a mutating maintenance action |

Compaction rules:

1. Compaction requires explicit approval or an active contract that covers memory writes.
2. The original pre-compaction file is archived to `specs/memory/archive/YYYY-MM-DD-<id>.md`.
3. Frontmatter sets `compacted_from` to the archive path and `compaction_status` to `compacted`.
4. Superseded entry clusters may be summarized, but the most recent active entry in each cluster remains intact.
5. Compaction is advisory housekeeping, not deletion; the archive remains the historical record.

### 2.5 Legacy Memory Migration

Existing memory guidance that says "keep each file under 50 lines" is treated as legacy operational advice, not the new normative file contract. Legacy notes may remain readable as-is, but new or migrated memory artifacts should adopt the append-only timeline model above.

---

## 3. Maintenance Contract Schema

### 3.1 Contract Purpose

Maintenance contracts are explicit approval records that pre-authorize bounded housekeeping actions. They reduce repeated prompts without allowing silent mutation.

### 3.2 Contract Types

| Contract Type | Scope | Expiry | Notes |
|---|---|---|---|
| `task_scoped` | One task or workflow run | On task completion | Default standing-approval choice |
| `session_scoped` | Current session only | On session end | Optional when user wants broader coverage |
| `standing` | Persistent approval across sessions | Hard-expire at 30 days | In-flight task may finish; the next task requires renewal |

### 3.3 Contract Shape

```yaml
schema_version: 1
artifact_type: maintenance_contract
id: contract-2026-04-17-bootstrap
status: active
created_at: 2026-04-17T09:00:00Z
updated_at: 2026-04-17T09:00:00Z
created_by: user
updated_by: user
session_id: sess-abc123
workflow_id: bootstrap
workflow_run_id: run-001
team_id: platform
chain_id: chain-xyz
owner_agent: orchestrator
assignee: orchestrator
approval_scope: task_scoped
approval_expires_at: 2026-04-17T18:00:00Z
permitted_actions:
  - memory_load
  - memory_write
  - qmd_sync
  - codegraph_sync
  - cleanup
  - repair
related_document_refs:
  - specs/006-housekeeping-memory-and-bootstrap/plan.md
```

### 3.4 Permitted Action Matrix

| Action | Read-only? | Per-action approval | `task_scoped` | `session_scoped` | `standing` |
|---|---:|---:|---:|---:|---:|
| Drift/staleness check | Yes | Allowed without contract | Allowed | Allowed | Allowed |
| Memory/context load | No | Yes | Yes | Yes | Yes |
| Memory write / compaction | No | Yes | Yes | Yes | Yes |
| qmd sync / re-index | No | Yes | Yes | Yes | Yes |
| codegraph sync / re-index | No | Yes | Yes | Yes | Yes |
| Cleanup of temporary artifacts | No | Yes | Yes | Yes | Yes |
| Repair of stale maintenance state | No | Yes | Yes | Yes | Yes |

Interpretation rules:

- qmd/codegraph drift checks are always read-only and never require a contract.
- qmd/codegraph sync or index rebuilds are mutating maintenance actions.
- Memory loads are treated as approval-gated maintenance actions when they materially expand task context beyond the current request.
- A contract may cover only a subset of actions; absent actions still require per-action approval.

### 3.5 Revocation and Expiry

- Users may revoke any active contract at any time.
- Revocation creates deferred maintenance items for future reporting.
- If a contract expires during an in-flight task, that task may finish under the original approval, but no new task may start under the expired contract.
- Standing approvals do not silently renew.

---

## 4. Sync Strategy for qmd and Codegraph

### 4.1 Sync State Path

The canonical sync state file is:

- `.ai/housekeeping/sync-state.json`

This path is project-local and should be treated as a maintenance artifact, not a user-authored spec.

### 4.2 qmd / Codegraph Index Defaults

- qmd index location defaults to a project-local path.
- codegraph state is likewise tracked project-locally unless a later spec explicitly overrides it.
- Cross-project/global caches are out of scope for Phases 1-2.

### 4.3 Drift Detection Triggers

Read-only drift checks may run at:

- Bootstrap
- Pre-task housekeeping
- Post-task housekeeping
- Post-file-write when a contract is already active

### 4.4 Drift Detection Inputs

Drift detection compares the last-known sync snapshot against current source state using one or more of:

- Source file modification times
- Source file content hashes
- Last successful index timestamp
- Tool enabled/disabled state

### 4.5 Sync State Schema

```json
{
  "schemaVersion": 1,
  "updatedAt": "2026-04-17T10:00:00Z",
  "qmd": {
    "enabled": true,
    "indexPath": ".qmd-index",
    "lastIndexTime": "2026-04-17T10:00:00Z",
    "sourceFingerprint": "sha256:abc123",
    "driftStatus": "fresh"
  },
  "codegraph": {
    "enabled": true,
    "dataPath": ".codegraph",
    "lastIndexTime": "2026-04-17T09:45:00Z",
    "sourceFingerprint": "sha256:def456",
    "driftStatus": "stale"
  },
  "staleAcked": {
    "qmd": [],
    "codegraph": [
      {
        "fingerprint": "sha256:def456",
        "ackedAt": "2026-04-17T10:05:00Z",
        "reason": "user rejected sync during bootstrap"
      }
    ]
  },
  "repairProposals": [
    {
      "tool": "codegraph",
      "proposalId": "repair-001",
      "status": "pending",
      "createdAt": "2026-04-17T10:05:00Z",
      "reason": "index older than source fingerprint"
    }
  ]
}
```

### 4.6 Repair Proposal Flow

1. Detect drift in read-only mode.
2. Create a repair proposal describing stale artifacts and the required re-index action.
3. Check whether the active contract covers the proposed mutation.
4. If covered, execute sync and update sync state.
5. If not covered, request approval.
6. If approved, execute sync and update sync state.
7. If rejected, record the fingerprint under `staleAcked` and report the stale condition in future bootstrap/housekeeping summaries until new drift is detected.

### 4.7 Rejection and `stale_acked` Behavior

- Rejection never mutates user content; it only records acknowledged staleness in sync state.
- `staleAcked` suppresses duplicate prompts for the same fingerprint.
- If the source fingerprint changes again, the prior acknowledgment becomes stale and the system may prompt again.
- If memory load is rejected, execution may continue with a report that context is intentionally incomplete.

---

## 5. Artifact-Specific Notes

### 5.1 Spec Directories and Primary Docs

Primary spec docs may share a common frontmatter block or keep metadata in the lead file for the directory. During migration, at least the lead artifact in each spec directory should establish the canonical metadata set.

### 5.2 RPI Task Files

Task files under `tasks/` should adopt the standardized frontmatter immediately when created under Spec 006. Legacy tasks without metadata should be inventoried and upgraded on touch.

### 5.3 Handoffs

Handoffs under `specs/memory/handoffs/` should use the same metadata core plus explicit `owner_agent`, `assignee`, `session_id`, `workflow_id`, `workflow_run_id`, and `related_document_refs` because they are cross-session continuity artifacts.

### 5.4 `compose_agent` Assumption

Phase 1-2 only assume that existing `compose_agent` inputs are sufficient for injected context. No new API surface is defined here; later phases should pass library/spec context through `stepInstructions` and currently supported fields first.

---

## 6. Approval-First Operating Rules

- Read-only drift checks are always allowed.
- qmd/codegraph sync, index rebuilds, memory writes, compaction, cleanup, and repair are mutating maintenance actions.
- Mutating maintenance actions require explicit user approval unless covered by an active maintenance contract.
- `ai-setup doctor` should report stray `AGENTS.md` files by default; deletion or migration remains a separate approved action.

---

## 7. Phase 1-2 Readiness Notes

This document resolves the defaults needed to unblock later workflow specs:

- canonical sync-state location selected
- enforcement model defined for new vs legacy artifacts
- standing approval expiry clarified
- qmd project-local indexing default selected
- approval-first treatment of sync/index actions made explicit
