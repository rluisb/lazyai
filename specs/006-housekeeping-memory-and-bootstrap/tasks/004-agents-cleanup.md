# Spec 006 — Task 004: AGENTS.md Cleanup and Migration

> Scope: Specify the cleanup of stray `AGENTS.md` files, the root-only target state, and the library-backed replacement strategy. This task is documentation-only and does not implement runtime, CLI, or migration code.

---

## Objective

Define the inventory, target state, migration path, and safety constraints for reducing `AGENTS.md` sprawl while preserving existing instruction coverage through root-level context plus library-backed `compose_agent` injection.

---

## Normative Inputs and Defaults

- Research: `specs/006-housekeeping-memory-and-bootstrap/research.md`
- Plan: `specs/006-housekeeping-memory-and-bootstrap/plan.md`
- Data model: `specs/006-housekeeping-memory-and-bootstrap/data-model.md`
- Prior tasks:
  - `specs/006-housekeeping-memory-and-bootstrap/tasks/001-metadata-migration.md`
  - `specs/006-housekeeping-memory-and-bootstrap/tasks/002-bootstrap-workflow.md`
  - `specs/006-housekeeping-memory-and-bootstrap/tasks/003-housekeeping-workflow.md`
- Sync state path: `.ai/housekeeping/sync-state.json`
- Metadata enforcement: strict for new artifacts; warn/migrate for legacy artifacts
- `compose_agent` integration assumption: use the existing API first and inject library/spec context through `stepInstructions` and current supported fields
- Approval default: task-scoped approval first; optional session-scoped; persistent standing approvals hard-expire at 30 days while allowing an in-flight task to finish
- `ai-setup doctor` behavior for stray `AGENTS.md`: report-only by default; no silent deletion
- qmd index path: project-local by default

---

## Inventory: AGENTS.md Files Relevant to Cleanup and Migration

### 1. Canonical project-level file to keep

| Path | Current Role | Phase 5 Decision |
|---|---|---|
| `AGENTS.md` | Canonical repo-wide operating rules and project context | **Keep** as the only project-local `AGENTS.md` in the target state |

### 2. Active stray files in live `specs/` paths

| Path | Current Role | Matching library replacement | Phase 5 Decision | Justification |
|---|---|---|---|---|
| `specs/adrs/AGENTS.md` | ADR-specific authoring and permanence rules | `library/specs-agents/adrs.md` | **Remove after migration** | Content already exists in library form and should be injected deliberately instead of auto-loaded by directory presence |
| `specs/features/AGENTS.md` | Shared workflow rules for feature/bugfix/refactor/tech-debt work | `library/specs-agents/features.md` plus related work-type files as needed | **Remove after migration** | Directory-scoped context should move to library-backed injection; root file remains the only ambient repo file |

### 3. Expected historical/spec-pattern files named in the plan

These are relevant to the migration contract even when they are not present in the current repo snapshot.

| Path Pattern | Current Repo Evidence | Planned Replacement | Migration Handling |
|---|---|---|---|
| `specs/rules/AGENTS.md` | Not present in current snapshot | `library/specs-agents/rules.md` | If found later, treat as stray and migrate/remove |
| `specs/templates/AGENTS.md` | Not present in current snapshot | `library/specs-agents/templates.md` | If found later, treat as stray and migrate/remove |
| `specs/standards/AGENTS.md` | Not present in current snapshot | `library/specs-agents/standards.md` | If found later, treat as stray and migrate/remove |
| `specs/memory/AGENTS.md` | Not present in current snapshot | `library/specs-agents/memory.md` | If found later, treat as legacy memory guidance and migrate/remove carefully |
| `specs/prompts/AGENTS.md` | Not present in current snapshot | `library/specs-agents/prompts.md` | If found later, treat as stray and migrate/remove |
| `specs/tech-debt/AGENTS.md` | Not present in current snapshot | `library/specs-agents/tech-debt.md` | If found later, treat as stray and migrate/remove |
| `specs/refactors/AGENTS.md` | Not present in current snapshot | `library/specs-agents/refactors.md` | If found later, treat as stray and migrate/remove |
| `specs/bugfixes/AGENTS.md` | Not present in current snapshot | `library/specs-agents/bugfixes.md` | If found later, treat as stray and migrate/remove |

### 4. Related non-target inventories to distinguish from cleanup scope

| Path / Area | Relevance | Phase 5 Decision |
|---|---|---|
| `.opencode/agents/` | Tool-specific agent definitions; not directory auto-context files | **Keep unchanged** |
| `library/specs-agents/*.md` | Replacement source of spec/workflow guidance | **Keep and use as injection source** |
| `docs/AI-Agentic-Setup-Templates/**/AGENTS.md` | Template/example corpus showing historical sprawl patterns | **Do not treat as live project runtime context**; future template maintenance may align separately |
| `.tmp-*`, fixture samples, backups | Test/demo/history artifacts, not canonical runtime policy | **Out of cleanup enforcement unless explicitly targeted by a separate maintenance task** |

### 5. Inventory conclusions

- The live project currently has **one canonical root `AGENTS.md`** and **two active stray `specs/**/AGENTS.md` files**.
- The repo already contains a **library-backed replacement set** under `library/specs-agents/` for the categories named in the Phase 5 plan.
- Cleanup enforcement should focus on **live project paths**, while clearly excluding fixtures, backups, and template archives unless a later maintenance task expands scope.

---

## Root-Only AGENTS.md Target State

### Target rule

After Phase 5 migration lands, the project-local target state is:

1. `AGENTS.md` at repo root remains the **only ambient project `AGENTS.md`**.
2. `specs/**/AGENTS.md` files are treated as **stray legacy artifacts**, not the preferred source of instructions.
3. Additional instruction content lives in **library markdown fragments** and is injected explicitly when needed.
4. `.opencode/agents/` remains valid because it defines agent personas/tools, not directory auto-loading rules.

### Why root-only

- It preserves a single canonical project contract.
- It removes ambiguous directory-based auto-context behavior.
- It aligns with the approved `compose_agent` assumption: inject extra context through existing fields rather than relying on file placement side effects.
- It keeps cleanup policy compatible with report-only doctor behavior and approval-first mutation rules.

---

## Replacement Strategy: Library Content + Current compose_agent Assumptions

### Replacement principle

Directory-specific instruction files are replaced by **explicit context composition**, not by silently recreating equivalent local `AGENTS.md` files.

### Replacement sources already present

The current repo already includes these replacement fragments:

- `library/specs-agents/adrs.md`
- `library/specs-agents/bugfixes.md`
- `library/specs-agents/features.md`
- `library/specs-agents/memory.md`
- `library/specs-agents/prompts.md`
- `library/specs-agents/refactors.md`
- `library/specs-agents/rules.md`
- `library/specs-agents/standards.md`
- `library/specs-agents/tech-debt.md`
- `library/specs-agents/templates.md`
- `library/specs-agents/workflows.md`

### Injection model

`compose_agent` should prefer the existing supported fields and current API shape:

1. Start with root project context from `AGENTS.md`.
2. Add workflow/spec guidance from `library/specs-agents/*.md` according to task type or directory semantics.
3. Pass that guidance through `stepInstructions` and other currently supported compose fields before considering any API expansion.
4. Keep the selection explicit in orchestrator/runtime composition logic rather than implicit in filesystem traversal.

### Mapping examples

| Former directory behavior | Replacement behavior |
|---|---|
| Working under `specs/adrs/**` auto-loaded `specs/adrs/AGENTS.md` | Runtime/library selection injects `library/specs-agents/adrs.md` when the task concerns ADR work |
| Working under `specs/features/**` auto-loaded `specs/features/AGENTS.md` | Runtime/library selection injects `library/specs-agents/features.md` and related work-type guidance as needed |
| Future work under rules/templates/standards/memory directories depended on local `AGENTS.md` presence | Runtime/library selection injects the corresponding `library/specs-agents/*.md` fragment when the workflow requires it |

### Explicit non-goal

Phase 5 does **not** require a new `compose_agent` API. The approved assumption is to use the current API shape first.

---

## Migration Steps

### Step 1 — Freeze the target contract in docs

- Declare root `AGENTS.md` as canonical.
- Declare `specs/**/AGENTS.md` to be legacy stray files.
- Declare `library/specs-agents/*.md` as the replacement content source.

### Step 2 — Verify replacement coverage before removal

Before any deletion is proposed:

- confirm every live stray `AGENTS.md` has equivalent or intentionally consolidated content in `library/specs-agents/`
- confirm root `AGENTS.md` still carries repo-wide rules and does not depend on subdirectory auto-loading to stay valid
- confirm `compose_agent` usage can pass the needed library content through current fields such as `stepInstructions`

### Step 3 — Migrate live stray files

For each live stray file in `specs/**`:

1. compare its unique instructions against the matching library fragment
2. copy any missing normative guidance into the library fragment if needed
3. update the migration inventory/report
4. only then propose deletion of the stray file

### Step 4 — Update generation rules

- `ai-setup init` and `ai-setup update` should stop generating subdirectory `AGENTS.md` files
- template/scaffolding logic should point to root-only `AGENTS.md` plus library fragments
- future spec scaffolds should rely on explicit composition, not local auto-context files

### Step 5 — Add doctor/reporting enforcement

`ai-setup doctor` should:

- detect stray `AGENTS.md` files below the repo root
- classify them as cleanup/migration findings
- report suggested replacement sources when known
- remain **report-only by default**
- never silently delete files

### Step 6 — Preserve auditability during removal

- any actual deletion remains a separate approved mutating action
- migration reports should name the files proposed for removal
- fixture/template/archive copies should be clearly labeled out-of-scope unless an explicit maintenance task includes them

---

## Safety Constraints

### Approval and mutation rules

- Inventorying and reporting stray `AGENTS.md` files is read-only and always allowed.
- Deleting, rewriting, or regenerating `AGENTS.md` files is a mutating action and must remain explicit.
- `ai-setup doctor` must preserve **report-only by default** behavior.
- No silent deletion is allowed.

### Legacy handling rules

- Treat existing subdirectory `AGENTS.md` files as legacy artifacts to migrate, not as automatically invalid files to erase.
- If a live stray file contains unique guidance not yet reflected in `library/specs-agents/`, preserve that guidance before proposing removal.
- Template examples, backups, and fixtures may document historical behavior and should not be swept into the live cleanup path by accident.

### Scope rules for this slice

- This Phase 5 slice is specification-only.
- It does not implement doctor checks, scaffold changes, runtime composition code, or deletion logic.
- It does not define Phase 6 install UX.

---

## Ownership Boundaries

| Concern | Orchestrator Runtime | ai-setup CLI / Wizard | Library / Spec Conventions |
|---|---|---|---|
| Selecting which instruction fragments to inject for a task | Owns runtime composition behavior | Does not own runtime execution | Defines what fragments exist and what they mean |
| Generating only root `AGENTS.md` for new projects | Does not own scaffolding | Owns init/update/scaffold behavior | Defines the target convention |
| Reporting stray subdirectory `AGENTS.md` files | May surface runtime observations if relevant | Owns doctor/report UX | Defines what counts as stray vs canonical |
| Library-backed guidance files under `library/specs-agents/` | Consumes them through current compose fields | May scaffold/update them in templates later | Owns the normative replacement content |
| Deletion/migration policy | Does not silently delete | Must keep report-only default unless user approves mutation | Defines safety constraints and migration order |

---

## Acceptance Check for Phase 5

This task satisfies the Phase 5 documentation slice when all of the following are true:

- the AGENTS inventory distinguishes canonical, live stray, expected legacy-pattern, and out-of-scope files
- root-only `AGENTS.md` target state is explicit
- replacement strategy points to `library/specs-agents/` and current `compose_agent` assumptions
- migration steps describe coverage verification before deletion
- doctor/report behavior stays report-only by default with no silent deletion
- ownership boundaries clearly separate orchestrator runtime, CLI/wizard, and library/spec conventions
- no runtime or CLI implementation is introduced in this task

---

## Deferred Work Outside This Slice

The following work is intentionally deferred:

- implementing `ai-setup doctor` detection/output
- updating scaffold/template generation code
- wiring runtime `compose_agent` selection logic
- executing any actual AGENTS deletion or migration
- Phase 6 install-time UX for optional tooling
