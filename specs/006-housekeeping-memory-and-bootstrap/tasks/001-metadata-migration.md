# Spec 006 — Task 001: Metadata Migration Checklist

> Scope: Inventory current artifact gaps against the standardized metadata schema and define the concrete migration sequence for legacy files. This task is documentation-only; applying the migration remains a separate approved action.

---

## Objective

Bring existing spec, memory, handoff, and task artifacts into alignment with the Spec 006 metadata contract while preserving approval-first behavior and avoiding destructive rewrites.

---

## Scope Inventory

### A. Existing spec artifacts in repo

| Artifact Bucket | Current Evidence | Current Gap vs Spec 006 |
|---|---|---|
| `specs/001-store-and-errors/*` | Present | No standardized frontmatter metadata contract |
| `specs/002-simplification-and-restructure/*` | Present | No standardized frontmatter metadata contract |
| `specs/003-post-install-automation-and-integrations/research.md` | Present | No standardized frontmatter metadata contract |
| `specs/004-go-migration/*` | Present | No standardized frontmatter metadata contract |
| `specs/005-setup-flow-fixes/plan.md` | Present | No standardized frontmatter metadata contract |
| `specs/006-housekeeping-memory-and-bootstrap/{research,plan}.md` | Present | Foundation docs exist, but no shared metadata block yet |
| `specs/compliance.md` | Present | Documentation artifact with no standardized metadata block |

### B. Existing memory / handoff artifacts

| Artifact Bucket | Current Evidence | Current Gap vs Spec 006 |
|---|---|---|
| `specs/memory/AGENTS.md` | Present | Legacy rules reference old memory size guidance and no standardized metadata |
| `specs/memory/**/*.md` | None found in current repo snapshot | No memory notes yet to migrate |
| `specs/memory/handoffs/**/*.md` | Referenced by compliance/rules, but none found in current repo snapshot | Directory contract exists conceptually; no current handoff files to backfill |

### C. Existing task artifacts

| Artifact Bucket | Current Evidence | Current Gap vs Spec 006 |
|---|---|---|
| `specs/**/tasks/*.md` | None found in current repo snapshot before this task | No legacy task files to migrate |
| `specs/**/checklists/*.md` | Present in Spec 002 | Checklist files have no standardized metadata block |

### D. Existing workflow / chain artifacts

| Artifact Bucket | Current Evidence | Current Gap vs Spec 006 |
|---|---|---|
| Chain/workflow metadata in markdown specs | Conceptually referenced in plans/research | Not serialized with standardized metadata fields |
| qmd/codegraph sync state | Not yet created | Canonical path now defined as `.ai/housekeeping/sync-state.json` |
| Maintenance contracts | Not yet created | Contract schema now defined, but no live records exist |

### E. Existing AGENTS cleanup-related artifacts

| Artifact Bucket | Current Evidence | Current Gap vs Spec 006 |
|---|---|---|
| `specs/adrs/AGENTS.md` | Present | Stray subdirectory `AGENTS.md`; future doctor check should report it |
| `specs/features/AGENTS.md` | Present | Stray subdirectory `AGENTS.md`; future doctor check should report it |

---

## Migration Checklist

### Phase 1 — Inventory and classify

- [ ] Confirm the full set of spec markdown artifacts under `specs/`.
- [ ] Confirm whether any hidden or newly-added memory notes/handoffs appeared since this inventory.
- [ ] Confirm whether any task files were added outside `specs/**/tasks/*.md` conventions.
- [ ] Classify each artifact as `new-compliant`, `legacy-needs-migration`, or `out-of-scope`.

### Phase 2 — Define metadata backfill per artifact class

- [ ] For primary spec docs (`research.md`, `plan.md`, optional `spec.md`), add the standardized frontmatter block.
- [ ] For checklist artifacts, add the minimum metadata subset: `schema_version`, `artifact_type`, `id`, `title`, `created_at`, `updated_at`, `created_by`, `updated_by`, and `related_document_refs`.
- [ ] For future handoffs, require the cross-session fields on creation: `session_id`, `workflow_id`, `workflow_run_id`, `chain_id`, `owner_agent`, `assignee`.
- [ ] For future memory notes, migrate from free-form notes to the append-only timeline format.
- [ ] For future task files, require strict metadata at creation time.

### Phase 3 — Legacy handling rules

- [ ] Do not block reads of legacy artifacts that lack metadata.
- [ ] Emit a migration warning when a legacy artifact is opened for meaningful edits.
- [ ] Preserve original content while adding metadata; avoid content rewrites unless separately approved.
- [ ] Record any inferred fields in `migration_notes`.
- [ ] Record any still-missing fields in `legacy_metadata_gaps`.

### Phase 4 — Maintenance artifact rollout

- [ ] Create `.ai/housekeeping/` only when a maintenance feature actually needs to write state.
- [ ] Create `.ai/housekeeping/sync-state.json` using the Phase 2 schema when sync tracking begins.
- [ ] Treat sync-state creation/update as a mutating maintenance action requiring approval or contract coverage.
- [ ] Treat qmd/codegraph re-index actions as mutating maintenance actions, not silent background work.

### Phase 5 — Doctor/reporting follow-up

- [ ] Add report-only detection for stray `AGENTS.md` files under `specs/**`.
- [ ] Report legacy metadata gaps without auto-fixing them.
- [ ] Distinguish `warning` (legacy artifact missing metadata) from `error` (new artifact created without required metadata).

---

## Minimum Backfill Recommendations by Artifact Type

| Artifact Type | Recommended Backfill Level | Notes |
|---|---|---|
| Existing plan/research/spec docs | High | They are long-lived and frequently read |
| Existing checklists | Medium | Add metadata when next touched |
| Existing `specs/memory/AGENTS.md` | Low | Legacy rules doc; update only when Phase 5 policy lands |
| Future handoffs | High | Strong continuity value across sessions |
| Future memory notes | High | Must use timeline model from creation |
| Future maintenance contracts | Strict from first use | No legacy contract format exists |
| Future sync-state file | Strict from first use | Machine-managed artifact |

---

## Open Follow-up for the Next Execution Slice

1. Apply standardized metadata to existing spec artifacts in priority order.
2. Draft bootstrap and housekeeping workflow docs using the contract and sync-state schemas defined in `data-model.md`.
3. Specify the report-only doctor output for stray `AGENTS.md` and legacy metadata gaps.
