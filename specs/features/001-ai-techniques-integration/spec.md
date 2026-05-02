# Spec: 001-ai-techniques-integration

**Feature ID:** 001
**Feature name:** ai-techniques-integration
**Date:** 2026-05-01
**Status:** Draft — created to correct Article III gap before implementation
**Owner:** Planner agent
**Constitution:** [`.ai/constitution/constitution.md`](../../../.ai/constitution/constitution.md)

> **Purpose.** This specification defines the verifiable Wave 1 contract for integrating five AI workflow techniques into `ai-setup`: N8 Constitution Population, N4 Coverage Thresholds, N11 Standards-as-Code, D6 Plan Validation, and D17 Adversarial Self-Play During Design. Implementation approach belongs in `plan.md`; this file is the acceptance contract.

---

## User Scenarios

### P1 — Generate a populated project harness
**As a** developer adopting `ai-setup`  
**I want** `ai-setup init` to capture or infer project profile values  
**So that** generated instructions are usable without manual placeholder cleanup.

**Acceptance criteria**
- [ ] AC-N8-001: Given all project-profile questions are answered, when `ai-setup init` completes, then generated `AGENTS.md` contains no `[YOUR_*]` markers for fields collected by the wizard.
- [ ] AC-N8-002: Given non-interactive init or skipped optional fields, when templates are compiled, then unresolved fields preserve their literal fallback markers (for example `[YOUR_PROJECT_OVERVIEW]`) and never render as empty string, `null`, or `undefined`.
- [ ] AC-N8-003: Given an existing hand-edited `AGENTS.md`, when `ai-setup update` runs, then only recognized placeholders/value slots are patched and custom sections/rules/reordered structure are preserved verbatim.
- [ ] AC-N8-004: Given a target repo with top-level directories, when init generates the codebase map, then known ignored directories (`node_modules`, `dist`, `.git`, `vendor`) are excluded and responsibility cells remain `[WHAT_IT_DOES]` until a human fills them.
- [ ] AC-N8-005: Given identical wizard/config input, when Go and TS compile root instructions, then generated `AGENTS.md` output is byte-identical for the fields in scope.
- [ ] AC-N8-006: Given a TS lowdb v1 store, when the new CLI reads it, then missing v2 fields are backfilled with safe defaults without crashing.
- [ ] AC-N8-007: Given a v2 lowdb store must be used by an older strict CLI, when the documented downgrade helper/procedure is run, then v2-only fields are removed and the store can be read as v1.

### P2 — Make quality thresholds and standards concrete
**As a** project maintainer  
**I want** coverage thresholds and starter standards generated during setup  
**So that** quality gates and pattern reviews have concrete local contracts.

**Acceptance criteria**
- [ ] AC-N4-001: Given a supported stack is detected, when the coverage prompt is shown, then it displays a non-zero stack-aware default; unknown stacks default to 80.
- [ ] AC-N4-002: Given a coverage threshold is accepted or supplied by config, when scaffolding completes, then the value appears in generated `AGENTS.md` and in `specs/rules/testing.md` if that rule is selected.
- [ ] AC-N4-003: Given the coverage prompt is skipped, when scaffolding completes, then the default threshold is used rather than `0%` or an unresolved variable.
- [ ] AC-N4-004: Given an out-of-range threshold, when config/wizard validation runs, then values outside 1–100 are rejected or reset to the documented default.
- [ ] AC-N11-001: Given a standard/full installation, when scaffolding completes, then exactly five starter standards exist under `specs/standards/starter/`: orchestration patterns, test patterns, error handling, agent security, and context loading.
- [ ] AC-N11-002: Given each starter standard file, when the existing standard frontmatter validator runs, then every file passes validation.
- [ ] AC-N11-003: Given `specs/standards/starter/` contains only `.gitkeep` or `README.md`, when init runs, then missing starter files are still copied.
- [ ] AC-N11-004: Given a user-authored file already exists at the same starter-standard path, when init/update runs, then that file is not overwritten.
- [ ] AC-N11-005: Given Quality Gate 4 Pattern Consistency in the constitution, when a reviewer loads standards context, then at least one concrete starter standard is available to cite.

### P3 — Review plans before implementation approval
**As a** human approver  
**I want** plan-quality and optional adversarial findings surfaced before approving implementation  
**So that** design flaws and incomplete plans are caught before code is written.

**Acceptance criteria**
- [ ] AC-D6-001: Given a feature plan is produced, when the feature chain advances, then an inline planner quality check runs before the human approval gate and emits `PlanQualityReport` JSON.
- [ ] AC-D6-002: Given plan-quality findings, when the approval gate is displayed, then pass/warn/fail verdict and findings are visible to the human approver.
- [ ] AC-D6-003: Given malformed Markdown or ambiguous structural parsing, when plan-quality checks run, then those parsing issues produce `warn` findings, not `fail` findings.
- [ ] AC-D6-004: Given a `fail` verdict, when the chain advances, then the chain proceeds to the human approval gate with the report; it does not automatically loop back to `plan`.
- [ ] AC-D6-005: Given `spec.md` AC/FR entries, when R1 runs, then it checks coverage in `plan.md` phase/item/downstream-contract sections only; it does not require `tasks.md` because tasks do not exist yet.
- [ ] AC-D6-006: Given any finding, when JSON is produced, then `location` follows `{file, section, lineStart, lineEnd}` with repo-relative file path and 1-based line numbers when available.
- [ ] AC-D17-001: Given preset defaults, when flags are resolved, then `adversarialDesign` is `false` for minimal and `true` for standard/full.
- [ ] AC-D17-002: Given custom preset selection, when the user toggles adversarial design, then generated chain content reflects that choice.
- [ ] AC-D17-003: Given `adversarialDesign=true`, when the feature chain runs, then `red-team-plan` runs sequentially after plan quality and before the explicit approval gate.
- [ ] AC-D17-004: Given `adversarialDesign=false`, when the feature chain is generated or run, then no unverified runtime conditional behavior is assumed; the red-team step is omitted or skipped by a tested mechanism.
- [ ] AC-D17-005: Given red-team provider/API failure, when the red-team step runs, then it emits `status: "soft_fail"` and the approval gate still appears.
- [ ] AC-D17-006: Given both reports exist, when the approval gate is displayed, then findings are merged into one human-readable gate report.

---

## Functional Requirements

| ID | Requirement | Priority | Source story |
|---|---|---|---|
| FR-001 | The system MUST collect or infer project profile fields needed to populate root instructions. | P1 | P1 |
| FR-002 | The system MUST preserve literal fallback placeholders for skipped or unknown values. | P1 | P1 |
| FR-003 | The system MUST update existing `AGENTS.md` files by targeted replacement only, preserving unrecognized user edits. | P1 | P1 |
| FR-004 | The system MUST persist new profile, coverage, and feature-flag fields in existing Go/TS stores with backward-compatible defaults. | P1 | P1 |
| FR-005 | The system MUST provide a downgrade path for TS lowdb v2 stores before rollback to older strict CLIs. | P1 | P1 |
| FR-006 | The system MUST capture and render coverage thresholds with values constrained to 1–100. | P2 | P2 |
| FR-007 | The system MUST seed five starter standards using file-level idempotency and never overwrite existing user files. | P2 | P2 |
| FR-008 | The system MUST run inline planner quality checks before implementation approval. | P3 | P3 |
| FR-009 | The system MUST define plan-quality findings with stable rule IDs and location metadata. | P3 | P3 |
| FR-010 | The system MUST avoid automatic plan-quality fail loops; human approval/rejection remains the control point. | P3 | P3 |
| FR-011 | The system MUST support `adversarialDesign` preset defaults and custom configuration. | P3 | P3 |
| FR-012 | The system MUST run red-team plan review sequentially, not via a parallel block, when enabled. | P3 | P3 |
| FR-013 | The system MUST soft-fail red-team outages and surface the outage in the gate report. | P3 | P3 |
| FR-014 | The system MUST verify conditional red-team inclusion/skipping before relying on it. | P3 | P3 |

---

## Key Entities and Data Model Schemas

### `WizardConfig` / store schema v2

```json
{
  "schemaVersion": 2,
  "projectOverview": "string | null",
  "namingConventions": "string | null",
  "errorHandling": "string | null",
  "apiConventions": "string | null",
  "importOrder": "string | null",
  "protectedBranch": "string | null",
  "testCommand": "string | null",
  "lintCommand": "string | null",
  "buildCommand": "string | null",
  "coverageThreshold": "number (integer 1-100, default 80)",
  "featureFlags": {
    "adversarialDesign": "boolean"
  }
}
```

Lifecycle: created/updated during `init` and `update`; v1 stores are read with defaults; v2 stores can be downgraded to v1 for rollback.

### `FragmentContext` extension

```json
{
  "projectName": "string",
  "features": {
    "adversarialDesign": "boolean"
  },
  "constitution": {
    "projectOverview": "string | null",
    "stack": {
      "language": "string | null",
      "framework": "string | null",
      "database": "string | null",
      "orm": "string | null",
      "testing": "string | null",
      "packageManager": "string | null"
    },
    "conventions": {
      "naming": "string | null",
      "errorHandling": "string | null",
      "apiResponses": "string | null",
      "importOrder": "string | null"
    },
    "commands": {
      "test": "string | null",
      "lint": "string | null",
      "build": "string | null"
    },
    "protectedBranch": "string | null",
    "coverageThreshold": "number | null",
    "codebaseMap": [
      { "path": "string", "responsibility": "string" }
    ]
  }
}
```

Lifecycle: assembled during compile/scaffold; not persisted as a separate entity.

### Template fallback map

```json
{
  "PROJECT_OVERVIEW": "[YOUR_PROJECT_OVERVIEW]",
  "NAMING_CONVENTIONS": "[YOUR_NAMING_CONVENTION]",
  "ERROR_HANDLING": "[YOUR_ERROR_PATTERN]",
  "API_CONVENTIONS": "[YOUR_API_CONVENTION]",
  "IMPORT_ORDER": "[YOUR_IMPORT_ORDER]",
  "PROTECTED_BRANCH": "[YOUR_PROTECTED_BRANCH]",
  "TEST_COMMAND": "<!-- fill-in: test command -->",
  "LINT_COMMAND": "[YOUR_LINT_COMMAND]",
  "BUILD_COMMAND": "<!-- fill-in: build command -->",
  "COVERAGE_THRESHOLD": "[YOUR_COVERAGE_THRESHOLD]"
}
```

Lifecycle: used at compile/update time; values are emitted exactly when source data is missing.

---

## Internal Contracts

| Contract | Full JSON shape |
|---|---|
| `PlanQualityReport` | `{"schemaVersion":"plan-quality-report/v1","verdict":"pass|warn|fail","findings":[{"rule":"R1|R2|R3|R4","severity":"info|warn|fail","message":"string","location":{"file":"string","section":"string|null","lineStart":"number|null","lineEnd":"number|null"}}],"checkedAgainst":{"spec":"string","plan":"string","research":"string|null","tasks":null}}` |
| `RedTeamPlanReport` | `{"schemaVersion":"red-team-plan-report/v1","status":"ok|soft_fail|skipped","findings":[{"category":"scope|security|feasibility|rollback|edge-case|assumption|operational","severity":"low|medium|high|critical","message":"string","recommendation":"string","location":{"file":"string","section":"string|null","lineStart":"number|null","lineEnd":"number|null"}}]}` |
| `MergedGateReport` | `{"schemaVersion":"plan-gate-report/v1","summary":{"planVerdict":"pass|warn|fail","redTeamStatus":"ok|soft_fail|skipped","blockingCount":"number","warningCount":"number"},"planQuality":"PlanQualityReport","adversarialReview":"RedTeamPlanReport|null"}` |
| `TargetedUpdatePatch` | `{"file":"AGENTS.md","replacements":[{"field":"string","oldText":"string","newText":"string","location":{"section":"string|null","lineStart":"number|null","lineEnd":"number|null"}}],"warnings":["string"],"preservedUnrecognizedContent":true}` |
| Sequential `feature` chain shape | `{"steps":[{"id":"research"},{"id":"plan","gate":null},{"id":"plan-quality"},{"id":"red-team-plan","optionalByFeature":"adversarialDesign"},{"id":"plan-gate","gate":"user_approval"},{"id":"implement"},{"id":"review"},{"id":"fix"},{"id":"document"}]}` |

---

## Feature Flags and Preset Defaults

| Flag | Type | Minimal | Standard | Full | Custom behavior |
|---|---|---:|---:|---:|---|
| `adversarialDesign` | boolean | false | true | true | Ask user during custom preset / feature selection. |

Rules:
- The flag controls whether `red-team-plan` is included or skipped by a tested mechanism.
- No runtime conditional execution may be assumed until tested.
- If conditional support cannot be verified, generated chains must be static per selected flag value.

---

## Verification Approach Summary

- **Test-first:** Each implementation task begins by adding failing tests for the contract it changes.
- **Schema/migration:** Go SQLite migration tests and TS lowdb v1→v2/v2→v1 round-trip tests run before integration work.
- **Compiler/template:** Unit tests cover each new `FragmentContext` field and fallback marker.
- **Update safety:** Fixture tests prove targeted update preserves hand-authored `AGENTS.md` sections.
- **Standards:** Frontmatter validation and file-level idempotency tests cover `.gitkeep`, `README.md`, and existing user files.
- **Chain:** Chain-shape tests prove sequential order, explicit `plan-gate`, no parallel block, no auto fail loop, and verified red-team inclusion/skipping.
- **Reports:** JSON contract tests validate `PlanQualityReport`, `RedTeamPlanReport`, `MergedGateReport`, and location format.
- **Parity:** Go/TS snapshot parity verifies identical output for the same input where both packages implement the same feature.

---

## Rollback Paths Summary

| Area | Rollback path |
|---|---|
| N8 profile fields | Nullable SQLite fields and optional lowdb keys can be ignored by current code; TS rollback to older strict CLI requires running the v2→v1 downgrade helper first. |
| N4 coverage threshold | Remove prompt/config use; fallback marker or default remains safe. Downgrade helper removes the v2 key for older TS CLIs. |
| N11 starter standards | Remove bundled starter files and copy path. Already-copied target files remain user-owned and are not deleted. |
| D6 plan quality | Remove `plan-quality` step and inline planner checklist only after active chains drain or no-op compatibility stubs are installed. |
| D17 adversarial design | Set `adversarialDesign=false` in presets/config first; remove step only after active runs drain or compatibility stubs exist. |
| Chain mid-flight | Do not remove/rename steps referenced by active runs. Drain, handoff, abandon with human approval, or leave no-op stubs until completion. |

---

## Edge Cases

- **EC-001 — Existing `AGENTS.md` has custom structure:** update preserves unknown content and reports warnings for unparseable known fields.
- **EC-002 — Standards directory has only metadata files:** starter standards still copy missing files.
- **EC-003 — Plan Markdown malformed:** plan quality emits warnings, not hard failures caused by parser uncertainty.
- **EC-004 — Red-team service unavailable:** gate report includes `soft_fail`; chain reaches approval gate.
- **EC-005 — Older CLI rollback:** downgrade helper/procedure is required before strict v1 readers consume v2 lowdb.

---

## Assumptions

- **A-001:** Existing Go/TS compilers can be extended with optional context fields without changing template syntax — confidence HIGH from current `FragmentContext` usage in plan research.
- **A-002:** Existing tests can run headless init/scaffold flows in both packages — confidence MEDIUM; task generation must verify commands before implementation.
- **A-003:** Conditional chain behavior is not assumed — confidence HIGH; the spec requires verification or static generation fallback.

---

## Out of Scope

- Standalone `plan-validate` skill file or reusable validation framework.
- Parallel blocks or graph execution in `feature.json`.
- Full root-template re-emission during `update`.
- RAG/retrieval preflight, model routing, guardrails runtime, benchmark/eval suite, continuous learning, debate workflows, auto-recovery, agent state machine.
- MCP server configuration changes.

---

## Clarifications

| Date | Question | Answer | Decided by |
|---|---|---|---|
| 2026-05-01 | Should D6 create a reusable `plan-validate` skill? | No. Inline checks in the existing planner skill for Wave 1; extraction requires future duplication and approval. | Review findings + planner revision |
| 2026-05-01 | Should plan quality auto-loop on failure? | No. Findings are surfaced at `user_approval`; human rejection controls revision. | Review findings + planner revision |
| 2026-05-01 | How should existing `AGENTS.md` be updated? | Targeted replacement only; no full template re-emission. | Review findings + planner revision |

---

## Constitutional Notes

- **Article I — Library-First:** Reuses existing compiler, store, wizard, chain machine, red-team role, and standards validator.
- **Article II — Test-First:** The verification summary requires tests before each production change.
- **Article III — Docs as Source of Truth:** This spec is the contract for the revised plan and must be approved before implementation.
- **Article IV — YAGNI:** Defers Wave 2/3/4 techniques and rejects unrequested runtime systems.
- **Article V — Simplicity:** Chooses inline planner checks and sequential chain steps over new abstractions.
- **Article VI — Anti-Overengineering:** Avoids one-caller helper/skill extraction and speculative conditional runtime support.

---

## Downstream Contract

| Produced for | Filename |
|---|---|
| `speckit-plan` | this file → `plan.md` |
| `speckit-tasks` | this file + `plan.md` → `tasks.md` |
| `speckit-analyze` | this file + `plan.md` + `tasks.md` |
