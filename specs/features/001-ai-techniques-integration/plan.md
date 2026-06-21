> **HISTORICAL (2026-06-20) — parts of this plan target the removed `packages/ai-setup-ts` package; the CLI is now Go-only (`packages/cli`). Retained for historical context.**
>
> This document is historical; `packages/ai-setup-ts` is no longer present.

# Plan: AI Techniques Integration into ai-setup Scaffolder

**Feature ID:** 001
**Spec:** [./spec.md](./spec.md)
**Research:** [./research.md](./research.md), [./ai-techniques-additions-analysis.md](./ai-techniques-additions-analysis.md), [./ai-techniques-patterns-review.md](./ai-techniques-patterns-review.md)
**Date:** 2026-05-01
**Status:** Revised Draft — review rejection addressed; awaiting human re-review and approval
**Owner:** Planner agent (handoff to Builder once approved)
**Constitution:** [`.ai/constitution/constitution.md`](../../../.ai/constitution/constitution.md)

> **Purpose.** This plan describes how the Wave 1 requirements in [`spec.md`](./spec.md) will be built. `spec.md` is now the acceptance contract; this plan contains implementation approach, sequencing, risk handling, and task-generation guidance.

---

## Summary

We integrate five Wave 1 techniques into the `ai-setup` scaffolder so a fresh `ai-setup init` produces a more complete project harness and the feature workflow presents plan-quality evidence before implementation approval. The Wave 1 scope is fixed to: **N8 Constitution Population**, **N4 Coverage Thresholds**, **N11 Standards-as-Code**, **D6 Plan Validation**, and **D17 Adversarial Self-Play During Design**.

This revision corrects the rejected draft by: creating `spec.md`; keeping `feature.json` as a sequential state machine; moving the `user_approval` gate from `plan` to an explicit sequential gate step; inlining D6 validation into the existing planner skill instead of creating a standalone `plan-validate` skill; changing `update` from full template re-emission to targeted replacement; adding downgrade and mid-flight rollback policy; and making red-team failures soft-fail so an API outage does not block the chain.

Wave 1 remains two shippable phases:
1. **W1.A — Constitution + Coverage + Standards Seed** (N8 + N4 + N11)
2. **W1.B — Plan Quality + Adversarial Design Review** (D6 + D17)

---

## Technical Context

| Aspect | Decision | Rationale |
|---|---|---|
| Language(s) | Go 1.26 canonical via `packages/ai-setup-go/`; TypeScript 5.x mirror via `packages/ai-setup-ts/` | Existing dual implementation; Go remains canonical and TS parity is verified task-by-task. |
| Library root | `packages/ai-setup-go/library/` (repo-root `library` symlink points here) | Reuses existing install content surface. |
| Wizard | `packages/ai-setup-go/cmd/init.go` and `packages/ai-setup-ts/src/wizard/phase2-features.ts` | Existing prompt drivers collect project profile and feature flags; no new wizard engine. |
| Compiler | Existing Go `internal/compiler/` and TS `src/compiler/` `FragmentContext` | Extend the existing template context and `{{VAR}}`/`{{#if features.x}}` behavior; no new template engine. |
| Storage | Go SQLite store; TS lowdb JSON store | Add nullable/optional fields to existing stores; no new datastore. |
| Orchestration | Existing sequential `feature.json` state machine | No graph or parallel-block primitive is assumed. W1.B uses sequential steps and an explicit gate step. |
| Testing | Existing Go and TS unit/integration/snapshot harnesses | Tests are written before production changes per Article II. W1.A starts with schema and compiler tests before wizard integration. |
| Telemetry | None new | Gate 5 is satisfied by rollback/detection instructions for this scaffold-level change; no runtime metrics system is added in Wave 1. |

**External dependencies (new):**
- None. All Wave 1 work is content, schema, compiler, wizard, and chain-definition work using existing packages.

**External dependencies (rejected):**
- Markdown/NLP validation frameworks — rejected. D6 uses conservative inline planner checks and structured findings; no semantic parser dependency.
- Runtime guardrails, RAG, model routing, benchmark infrastructure — rejected for Wave 1 per `research.md` sequencing and Article IV.

---

## Constitution Check

| Article | Verdict | Justification |
|---|---|---|
| I — Library-First | **PASS** | All five Wave 1 items reuse existing compiler, wizard, store, chain, and library surfaces. No new datastore, workflow engine, parser framework, or validation runtime is introduced. |
| II — Test-First (NON-NEGOTIABLE) | **PASS** | W1.A begins with failing schema/compiler/idempotency tests; W1.B begins with failing chain-shape/report-contract tests before `feature.json` and skill prompt changes. |
| III — Docs as Source of Truth | **PASS after revision** | The rejected draft violated Article III by omitting `spec.md`. This revision creates `spec.md` and makes it the acceptance contract before implementation. The process violation is documented and corrected before human approval. |
| IV — Anti-Speculation (YAGNI) | **PASS** | Wave 2/3/4 items remain roadmap-only. Plan validation is limited to current feature planning; no reusable validator framework, RAG, model routing, telemetry, or future-chain abstraction is added. |
| V — Simplicity Over Abstraction | **PASS** | D6 validation is inlined into the existing planner skill/step. The feature chain remains sequential. Standards seed uses file-level copy logic instead of a standards-management subsystem. |
| VI — Anti-Overengineering (NON-NEGOTIABLE) | **PASS** | The removed `plan-validate` skill avoids an abstraction with one caller. New complexity is limited to explicit schema fields, one optional red-team step, and rollback scripts/procedure needed for safe release. |

**Verdict:** **REVISED — READY FOR RE-REVIEW.** The prior plan's false Article III pass, impossible chain graph, speculative validator skill, and destructive update path have been corrected.

---

## Project Structure

Paths are implementation targets for Wave 1. `spec.md` is read-only during implementation except by explicit approved revision.

```
ai-setup/
├── packages/ai-setup-go/
│   ├── cmd/init.go                                           ← MODIFIED (project profile + coverage prompts)
│   ├── internal/compiler/template.go                         ← MODIFIED (FragmentContext fields + fallbacks)
│   ├── internal/scaffold/root.go                             ← MODIFIED (targeted AGENTS updates, not full re-emission)
│   ├── internal/scaffold/constitution.go                     ← MODIFIED (placeholder resolution)
│   ├── internal/scaffold/specs.go                            ← MODIFIED (file-level standards seed copy)
│   ├── internal/store/migrations/006-add-constitution-fields.sql ← NEW (nullable constitution/coverage fields)
│   ├── internal/migration/json_bridge.go                     ← MODIFIED (lowdb→SQLite bridge shape)
│   ├── library/root/AGENTS.template.md                       ← MODIFIED (fallback placeholders retained)
│   ├── library/rules/testing.md                              ← MODIFIED (coverage threshold substitution)
│   ├── library/standards/starter/                            ← NEW directory (5 standards)
│   │   ├── orchestration-patterns.md
│   │   ├── test-patterns.md
│   │   ├── error-handling.md
│   │   ├── agent-security.md
│   │   └── context-loading.md
│   ├── library/skills/plan/SKILL.md                          ← MODIFIED (inline D6 plan quality checks)
│   ├── library/skills/red-team-plan.md                       ← NEW (D17 soft-fail plan critique)
│   └── library/orchestration/chains/feature.json             ← MODIFIED (sequential plan-quality/red-team/gate steps)
├── packages/ai-setup-ts/
│   ├── src/wizard/phase2-features.ts                         ← MODIFIED (mirror Go fields and flag)
│   ├── src/wizard/planner.ts                                 ← MODIFIED (planned file emission for standards)
│   ├── src/scaffold/compiled-root.ts                         ← MODIFIED (pass extended FragmentContext)
│   ├── src/scaffold/specs.ts                                 ← MODIFIED (file-level standards seed copy)
│   ├── src/compiler/fragment-resolver.ts                     ← MODIFIED (fallback substitution)
│   ├── src/presets.ts                                        ← MODIFIED (`adversarialDesign` defaults)
│   ├── src/store/schema.ts                                   ← MODIFIED (v2 optional fields)
│   ├── src/store/migrations/v2-to-v1.ts                      ← NEW (rollback/downgrade helper)
│   └── src/__tests__/                                        ← NEW/MODIFIED tests per task
└── specs/features/001-ai-techniques-integration/
    ├── spec.md                                               ← NEW contract
    ├── research.md                                           ← read-only at implementation time
    ├── ai-techniques-additions-analysis.md                   ← read-only reference
    ├── ai-techniques-patterns-review.md                      ← read-only reference
    ├── plan.md                                               ← THIS FILE
    └── tasks.md                                              ← generated after human approval
```

---

## Data Model

Authoritative schema shapes are also summarized in `spec.md`.

| Entity | Storage | Fields | Indexes / constraints |
|---|---|---|---|
| `WizardConfig` | Go SQLite `wizard_config`; TS lowdb schema v2 | `projectOverview`, `namingConventions`, `errorHandling`, `apiConventions`, `importOrder`, `protectedBranch`, `testCommand`, `lintCommand`, `buildCommand`, `coverageThreshold` | New fields are nullable/optional. `coverageThreshold` must be 1–100 when present; default 80. Missing values resolve to documented placeholders. |
| `FeatureFlags` | Existing preset/config store | `adversarialDesign: boolean` | Defaults: minimal=false, standard=true, full=true; custom preset asks user. |
| `PlanQualityReport` | Transient chain artifact | `{schemaVersion, verdict, findings[], checkedAgainst}` | JSON output contract only; not persisted except in run artifact/handoff. |
| `RedTeamPlanReport` | Transient chain artifact | `{schemaVersion, status, findings[]}` | `status=soft_fail` allowed; soft failures proceed to human gate. |

**Migrations:**
- Go: `packages/ai-setup-go/internal/store/migrations/006-add-constitution-fields.sql` adds nullable constitution/profile/coverage columns to `wizard_config`.
- TS: lowdb schema v1→v2 adds optional fields and `featureFlags.adversarialDesign`.
- TS rollback: `packages/ai-setup-ts/src/store/migrations/v2-to-v1.ts` removes v2-only keys and rewrites `schemaVersion: 1` for users who must run an older strict CLI.

**Backfill / data movement:**
- `ai-setup update` MUST NOT re-emit `AGENTS.md` from the root template by default.
- Update performs targeted replacements only:
  1. Read existing `AGENTS.md`.
  2. Parse known fields from known headings when exact markers or clearly delimited values exist.
  3. Replace only recognized `[YOUR_*]` markers and known generated value slots.
  4. Preserve every unrecognized section, custom rule, reordered block, and hand-authored paragraph verbatim.
  5. If a value cannot be parsed safely, leave the existing text unchanged and report a warning.
- A full template refresh is out of Wave 1 scope unless a future approved spec adds an explicit command.

**Template fallback contract:**
- Each new template variable maps to a legacy literal fallback. Example: `{{PROJECT_OVERVIEW}}` resolves to the wizard/store value when non-empty; otherwise it emits `[YOUR_PROJECT_OVERVIEW]` exactly.
- Fallbacks are defined in a single field-to-placeholder map in each implementation, not inferred from string casing at call sites.

---

## Internal Contracts

| Contract | Producer | Consumer | Shape |
|---|---|---|---|
| `FragmentContext` extension | Wizard/store/presets | Go and TS compilers | Adds `constitution` object and `features.adversarialDesign`; all fields optional. See `spec.md` for JSON shape. |
| Targeted root update | `ai-setup update` | Existing target repo `AGENTS.md` | Patch recognized placeholders/value slots only; no full-file re-emission; unrecognized content preserved byte-for-byte. |
| Standards seed copy | Scaffolder | Target repo `specs/standards/starter/` | File-level idempotency: copy each starter file only if that relative file is absent; `.gitkeep`/`README.md` do not suppress copy; existing user files are never overwritten. |
| Sequential feature chain | `library/orchestration/chains/feature.json` | Orchestrator chain machine | `research → plan → plan-quality → red-team-plan? → plan-gate(user_approval) → implement → review → fix/document`. No parallel block. No automatic fail loop. |
| `PlanQualityReport` | Existing planner skill, inline D6 checks | `plan-gate` prompt and human reviewer | `{schemaVersion:"plan-quality-report/v1", verdict:"pass|warn|fail", findings:[{rule,severity,message,location}], checkedAgainst:{spec,plan,research,tasks:null}}`. |
| `Finding.location` | Planner/red-team reports | Gate prompt/renderers/tests | `{file:string, section:string|null, lineStart:number|null, lineEnd:number|null}`. `file` is repo-relative; line numbers are 1-based when available. |
| `RedTeamPlanReport` | `red-team-plan` skill/step | `plan-gate` prompt and human reviewer | `{schemaVersion:"red-team-plan-report/v1", status:"ok|soft_fail|skipped", findings:[{category,severity,message,recommendation,location}]}`. |
| `MergedGateReport` | `plan-gate` step | Human approval gate | Contains `planQuality`, `adversarialReview`, and a summary count. `soft_fail` is displayed as warning, not a blocked chain. |

### Sequential chain design for W1.B

The current `feature.json` proves that steps are an array and `gate: "user_approval"` is a property on the `plan` step. Therefore W1.B MUST NOT insert a parallel block before the existing gate. Instead, W1.B changes the chain shape as follows:

1. `plan` produces `spec.md`/`plan.md` artifacts and transitions by `success` to `plan-quality`; it no longer owns `gate: "user_approval"`.
2. `plan-quality` uses the existing `planner` agent and existing `plan` skill with added inline quality-check instructions. It emits `PlanQualityReport` and always transitions forward to the gate path. A `fail` verdict is a gate recommendation, not an automatic transition loop.
3. `red-team-plan` is sequential and pure read-only. It is included only when `adversarialDesign` is enabled by the generated preset/config. If conditional runtime execution is not already supported, the compiler emits two static chain variants by feature flag rather than relying on unverified runtime skipping.
4. `plan-gate` is the only step with `gate: "user_approval"`. It presents `plan.md`, `spec.md`, `PlanQualityReport`, and optional `RedTeamPlanReport` to the human.
5. `approved → implement`; `rejected → plan` with the gate feedback injected into the next planner context.

**Conditional-skip verification gate:** W1.B may not edit `feature.json` until a test proves one of these mechanisms exists and works: (a) runtime step conditions, or (b) compile-time chain rendering via existing `{{#if features.adversarialDesign}}`. If neither is verified, W1.B must keep `red-team-plan` out of generated chains and document it as manual-only until a separate approved ADR/spec adds conditional chain support.

---

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation | Owner |
|---|---|---|---|---|
| `update` corrupts hand-edited `AGENTS.md` | M | H | Use targeted placeholder/value-slot replacement only; preserve unknown content byte-for-byte; add fixture with structural edits and custom sections. | Builder |
| Wizard expansion fatigues users | M | M | Group fields under one "Project Profile" block; auto-detect commands; preserve placeholders for skipped fields. | Builder |
| v2 lowdb store crashes older strict TS CLIs on rollback | M | H | Ship/document `v2-to-v1` downgrade helper and require backup before v2 write; rollback procedure includes downgrade before running older CLI. | Builder |
| Plan quality checks false-positive or parse Markdown unreliably | M | M | Conservative rules only; parser exceptions and ambiguous structural parsing produce `warn`, never `fail`; findings include exact locations. | Builder |
| Automatic fail loop repeats forever | M | H | No automatic `fail → plan` transition. All D6 verdicts proceed to `plan-gate`; human rejects/approves with feedback. | Builder |
| Red-team provider/API outage blocks feature chain | M | M | `red-team-plan` is wrapped as soft-fail; outage emits `status: soft_fail` warning and proceeds to `plan-gate`. | Builder |
| Standards seed clashes with existing user standards | L | M | File-level copy: copy only missing starter files; existing files are never overwritten; `.gitkeep`/`README.md` do not block copy. | Builder |
| TS/Go drift on fields and defaults | M | M | Start W1.A with schema validation/round-trip tests in both packages; extend parity snapshots before integration. | Builder |
| Mid-flight rollback removes steps referenced by active chain runs | L | H | Rollback policy: do not remove/rename chain steps while active runs exist. Drain, handoff, or leave no-op compatibility stubs until runs complete; then change active chain version. | Release owner |
| Conditional red-team skip assumed but unsupported | M | M | W1.B starts with a failing/proving test for runtime or compile-time conditional chain generation; no conditional chain behavior is shipped unverified. | Builder |

---

## Complexity Tracking

| Item | Simpler alternative | Why complexity is justified | Cost |
|---|---|---|---|
| Explicit `plan-gate` step instead of keeping gate on `plan` | Keep gate on `plan` and put checks in prose | Required to present generated quality/red-team reports before human approval without inventing a graph or parallel primitive. | One additional sequential chain step. |
| Inline D6 checks in planner skill | Separate `plan-validate` skill | Separate skill was rejected as YAGNI and an abstraction with one caller. Inline checks are simpler and constitutional. | Planner skill grows by a bounded checklist. |
| Typed `constitution` fields instead of `metadata: map[string]string` | Generic map | Type/schema tests catch drift and prevent silent placeholder drops across Go/TS. | Several optional fields per language. |
| Lowdb v2→v1 downgrade helper | Document manual JSON edit only | Older strict CLIs can crash on v2 stores; a safe rollback path is required by Gate 5. | One small migration helper and documentation. |
| Sequential red-team step with soft-fail | Manual red-team only | D17 value is automatic design critique before implementation; soft-fail limits operational risk. | One skill markdown and one optional sequential chain step. |

> Removed complexity from rejected draft: no standalone `plan-validate` skill, no parallel chain block, no automatic fail loop, no full template re-emission.

---

## Phases & Milestones

| Phase | Goal | Exit criterion |
|---|---|---|
| **W1.A — Constitution + Coverage + Standards Seed** | A fresh `ai-setup init` resolves collected project-profile fields, coverage threshold, and seeds 5 starter standards safely. | Schema/round-trip tests pass; AGENTS snapshot has zero unresolved markers for collected fields; skipped fields preserve fallback markers; standards are copied file-by-file without overwriting existing files; Go/TS parity passes. |
| **W1.B — Plan Quality + Adversarial Design Review** | The `feature` chain reaches human approval only after sequential plan-quality and optional red-team reports are prepared. | Chain-shape tests prove no parallel block; gate moved to `plan-gate`; conditional red-team inclusion is verified or omitted; reports match JSON contracts; red-team outage soft-fails. |
| **Wave 2 (roadmap only)** | Quality + Context: D4, N12, N2, D9, D13, D14, D5, D8. | Future plan revision after W1 ships and is observed in real use for at least 2 weeks. |
| **Wave 3 (roadmap only)** | Retrieval & Intelligence: N9 + D1, D2 rule-doc only, N1, N6. | Future spec/plan; `retrievalPreflight` remains out of Wave 1. |
| **Wave 4 (roadmap only)** | Learning & Measurement: D18, D11. | Future spec/plan after real usage data exists. |

> Tasks for W1.A and W1.B will be generated by `speckit-tasks` after human approval. Roadmap waves are intentionally not task-decomposed.

### Wave 1 Item Detail

#### N8 — Constitution Population (P0/M) — Phase W1.A

**Scope.** Capture project profile in the wizard: project overview, naming conventions, error handling, API conventions, import order, protected branch. Auto-detect commands where existing stack detection can do so. Resolve collected fields in root templates with explicit fallback placeholders. Generate codebase map rows from top-level directories with `[WHAT_IT_DOES]` responsibility placeholders for human completion.

**Files affected.** `cmd/init.go`, `internal/compiler/template.go`, `internal/scaffold/root.go`, `internal/scaffold/constitution.go`, `internal/store/migrations/006-add-constitution-fields.sql`, `internal/migration/json_bridge.go`, `library/root/AGENTS.template.md`, TS `src/wizard/phase2-features.ts`, `src/scaffold/compiled-root.ts`, `src/compiler/fragment-resolver.ts`, `src/store/schema.ts`, `src/store/migrations/v2-to-v1.ts`.

**Install surfaces touched.** Wizard prompt + compile-time emitter + library content. No MCP server config.

**Acceptance criteria.** See `spec.md` AC-N8-001 through AC-N8-007.

**Verification approach.** Schema tests first; template fallback unit tests; targeted-update fixture tests; headless init integration tests; Go/TS parity snapshots; manual wizard walkthrough.

**Rollback path.** Revert nullable SQLite migration via normal DB backup/restore if needed. For TS lowdb, run the v2→v1 downgrade helper before invoking older strict CLIs. Root templates retain legacy placeholders, so generated files remain readable after rollback.

#### N4 — Coverage Thresholds (P1/XS) — Phase W1.A

**Scope.** Add one coverage-threshold prompt/config value with stack-aware defaults (default 80 when stack is unknown). Resolve coverage placeholders in `AGENTS.md` and `specs/rules/testing.md` when present.

**Files affected.** Same context/compiler/wizard files as N8 plus `library/rules/testing.md`.

**Install surfaces touched.** Wizard + compiler emitter + library content.

**Acceptance criteria.** See `spec.md` AC-N4-001 through AC-N4-004.

**Verification approach.** Unit tests for default selection and bounds; integration test asserting substitution; skipped-field fallback test.

**Rollback path.** Remove the prompt/config key and let fallback placeholders render. For lowdb rollback, the v2→v1 helper removes `coverageThreshold`.

#### N11 — Standards-as-Code (P1/S) — Phase W1.A

**Scope.** Author 5 starter standards in `library/standards/starter/` and scaffold them into `specs/standards/starter/` using file-level idempotency.

**Files affected.** `library/standards/starter/*.md`, `internal/scaffold/specs.go`, TS `src/scaffold/specs.ts`.

**Install surfaces touched.** Library content + compile-time emitter.

**Acceptance criteria.** See `spec.md` AC-N11-001 through AC-N11-005.

**Verification approach.** Frontmatter validator; integration file-count/content tests; idempotency tests with `.gitkeep`, `README.md`, and pre-existing user-authored files.

**Rollback path.** Remove starter source files and copy code path. Existing target repos keep copied standards as ordinary user files.

#### D6 — Plan Validation (P1/S) — Phase W1.B

**Scope.** Add an inline "Plan Quality Check" section to the existing planner skill/step. It runs after `plan.md` is produced and before the explicit `plan-gate`. It emits `PlanQualityReport`; it does not create `library/skills/plan-validate.md`.

Initial rules:
- **R1 — Spec coverage in plan:** Every Wave 1 `spec.md` acceptance criterion / functional requirement has at least one matching plan item detail, phase exit criterion, or downstream contract reference. R1 does **not** inspect `tasks.md`, because tasks do not exist when this check runs; task coverage remains the responsibility of `speckit-tasks`/`speckit-analyze` after tasks are generated.
- **R2 — Phase exit criteria:** Every planned phase has an observable exit criterion.
- **R3 — Risk mitigation completeness:** Every risk row has a mitigation and owner; R3 is not about constitution fields.
- **R4 — Rollback coverage:** Every Wave 1 item has a rollback path.

Structural parser errors, malformed tables, or uncertain Markdown parsing produce `warn` findings with locations when available; they never produce `fail` by themselves. Confident content omissions may produce `fail`, but all verdicts proceed to `plan-gate` for human decision.

**Files affected.** `library/skills/plan/SKILL.md`, `library/orchestration/chains/feature.json`, tests. No new `plan-validate` skill file.

**Install surfaces touched.** Library content + orchestration definition.

**Acceptance criteria.** See `spec.md` AC-D6-001 through AC-D6-006.

**Verification approach.** Report-contract unit tests; fixture checks for pass/warn/fail; chain integration test confirming no automatic loop and gate report surfacing.

**Rollback path.** Remove the `plan-quality` step and the inline planner-skill checklist. Active runs must be drained or compatibility-stubbed before removing the step.

#### D17 — Adversarial Self-Play During Design (P1/S) — Phase W1.B

**Scope.** Add a sequential `red-team-plan` design critique before `plan-gate` when `features.adversarialDesign` is enabled. It reuses the existing red-team role in plan/spec attack mode and emits a soft-failable report.

**Files affected.** `library/skills/red-team-plan.md`, `library/orchestration/chains/feature.json`, `src/presets.ts`, `src/store/schema.ts`, `src/wizard/phase2-features.ts`, Go equivalents for preset/config if present.

**Install surfaces touched.** Library content + feature preset flag + orchestration definition + wizard prompt/compile-time generation.

**Acceptance criteria.** See `spec.md` AC-D17-001 through AC-D17-006.

**Verification approach.** Frontmatter/schema test for skill; chain tests for enabled/disabled generated chain shape; integration test for merged report; simulated red-team outage test expecting `soft_fail` and gate continuation.

**Rollback path.** Set `adversarialDesign=false` in all presets and generated config; leave skill markdown available for manual use. Do not remove the step from active chain versions until active runs drain or a no-op compatibility stub is installed.

---

## Execution Order

**W1.A gates W1.B.** W1.B depends on W1.A because plan-quality reports cite `spec.md`, `plan.md`, rollback paths, and standards/constitution expectations. The earlier draft's R3 justification was wrong: R3 checks risk mitigation completeness, not constitution fields.

```
                ┌─────────────────────────────────────────────────────┐
                │ W1.A — Constitution + Coverage + Standards Seed      │
                │                                                     │
 start ────────►│ T1: schema + migration tests (Go/TS, incl. v2→v1)    │
                │       │                                             │
                │       ├──► T2: FragmentContext fallback tests        │
                │       ├──► T3: wizard prompts/config wiring          │
                │       └──► T4: targeted update + parity snapshots    │
                │                                                     │
                │ T5: standards content + frontmatter tests            │
                │ T6: file-level standards copy tests + implementation │
                │ T7: W1.A integration/golden tests                    │
                └─────────────────────────────────────────────────────┘
                                       │
                                       ▼
                ┌─────────────────────────────────────────────────────┐
                │ W1.B — Sequential plan-quality and red-team reports  │
                │                                                     │
                │ T8: verify chain gate/conditional capabilities       │
                │ T9: inline planner quality checklist + report tests  │
                │ T10: move gate to plan-gate sequentially             │
                │ T11: red-team-plan skill + soft-fail tests           │
                │ T12: adversarialDesign flag + generated chain tests  │
                │ T13: end-to-end gate report integration tests        │
                └─────────────────────────────────────────────────────┘
                                       │
                                       ▼
                                  Wave 1 done
```

**Hard dependencies inside W1.A.** T1 blocks all schema/config writes. T2/T3/T4 depend on T1. T5 is independent. T6 depends on T5. T7 depends on T4 and T6.

**Hard dependencies inside W1.B.** T8 blocks `feature.json` edits. T9 blocks T10. T11 can start after T8. T12 depends on T8 and T11. T13 depends on T10 and T12.

**Soft constraint.** W1.A and W1.B should be separate PRs to validate constitution population and schema migration before altering the feature chain gate path.

---

## Resolution of Open Questions (research.md §8)

### Open Question #1 — Standards scope: 3-4 or 5?

**Resolved: 5.** The starter set exactly matches the review recommendation: orchestration patterns, test patterns, error handling, agent security, and context loading. Additional standards are Wave 2+ only after concrete review gaps appear.

### Open Question #2 — Backward compatibility on `update`

**Resolved: targeted replacement.** `ai-setup update` must preserve existing files by default and patch only recognized placeholders/value slots. Full root-template re-emission is explicitly rejected because it would overwrite hand-authored structure and custom rules.

Rollback compatibility is two-part:
1. Existing generated docs remain valid because fallback placeholders are retained.
2. Older strict TS CLIs require the v2→v1 lowdb downgrade helper before rollback.

### Open Question #3 — Extension interaction with new flags

**Resolved as out of Wave 1.** Wave 1 adds only `adversarialDesign`, which controls generated feature-chain content and does not alter agent composition APIs. Extension/RAG interaction is deferred to the future `retrievalPreflight` spec.

---

## Rollback and Release Policy

1. **Pre-release backup:** Before lowdb v2 write, create or document a backup of the v1 JSON store.
2. **TS downgrade:** To roll back to an older strict TS CLI, run `v2-to-v1` downgrade helper first; then install/run the older CLI.
3. **Go rollback:** SQLite fields are nullable; rollback uses DB backup/restore or a later approved down migration if the migration framework supports down files.
4. **Active chains:** Do not remove or rename `plan-quality`, `red-team-plan`, or `plan-gate` while active runs reference them. Drain active runs, hand them off, or leave no-op compatibility stubs until completion.
5. **Preset rollback:** Disable D17 by setting `adversarialDesign=false`; this is preferred over deleting the skill/step immediately.

---

## Out of Scope

- Standalone `library/skills/plan-validate.md` — rejected as premature abstraction.
- Parallel chain execution inside `feature.json` — unsupported by current sequential chain structure.
- Runtime conditional step execution unless verified by tests; compile-time chain generation is the fallback.
- Full `AGENTS.md` template refresh during `update` — rejected due data-loss risk.
- Wave 2 / 3 / 4 items: RAG, knowledge retrieval, model routing, guardrails runtime, evaluation benchmark, continuous learning, debate workflow, auto-recovery, agent state machine.
- MCP server configuration changes — untouched in Wave 1.
- New external parser/NLP libraries.

---

## Plan Validation Self-Check

| Review finding | Resolution in this revision |
|---|---|
| D1 Impossible chain restructuring | Chain is explicitly sequential; no parallel block; gate moved to `plan-gate`; conditional skip must be verified or implemented by compile-time generation. |
| D2 Missing `spec.md` | `spec.md` is created and linked; acceptance criteria moved there. |
| D3 Premature `plan-validate` skill | Removed standalone skill; D6 checks are inline in existing planner skill/step. |
| D4 Destructive `update` re-emission | Replaced with targeted placeholder/value-slot patching; custom structure preserved. |
| H1 lowdb v2→v1 rollback | Added downgrade helper/procedure and rollback policy. |
| H2 parser infinite loop | No auto fail loop; structural parser issues warn; all verdicts proceed to human gate. |
| H3 undefined R1 | R1 checks `spec.md` AC/FR coverage in `plan.md`, not `tasks.md`. |
| H4 missing SQLite migration file | `006-add-constitution-fields.sql` is included in project structure and N8 files. |
| M1 location format | `Finding.location` JSON shape defined. |
| M2 template fallback | Field-to-placeholder fallback map defined. |
| M3 R3 justification | Execution order corrected; R3 is risk mitigation completeness. |
| M4 Quality Gate 4 pointer | N11 references `.ai/constitution/constitution.md` Gate 4 Pattern Consistency. |
| M5 mid-flight rollback | Active-run drain/stub policy added. |
| M6 W1.A vulnerability | W1.A starts with schema/migration validation tests before integration. |
| M7 standards idempotency | Changed from empty-directory check to file-level copy. |
| L1 red-team outage | `red-team-plan` soft-fails and proceeds to gate. |

**Coverage:** Every Wave 1 item has acceptance criteria in `spec.md`, an implementation section here, a verification approach, and a rollback path. Every cross-analysis answer and research open question has either a concrete plan section or documented deferral.

**Constitution self-check:** Article I PASS; Article II PASS; Article III PASS after corrected spec creation; Article IV PASS; Article V PASS; Article VI PASS.

---

## Decisions Requiring Human Approval Before Implementation

1. **Standards count = 5.** Approve or change before task generation.
2. **Targeted update policy.** Approve preserving custom `AGENTS.md` structure and disallowing full re-emission in Wave 1.
3. **Adversarial design defaults:** `minimal=false`, `standard=true`, `full=true`.
4. **W1.A and W1.B as separate PRs.** Approve or request a single PR with higher review burden.
5. **D6 fail semantics:** `fail` is a human-gate recommendation, not an automatic chain loop.
6. **Conditional red-team behavior:** Approve compile-time generated chain variants if runtime conditional skipping is absent.

If approved, the next step is `/speckit-tasks` against `spec.md` + this plan. No implementation begins until approval is recorded.

---

## Downstream Contract

| Produces for | Filename |
|---|---|
| `speckit-tasks` | `spec.md` + this file → `tasks.md` |
| `speckit-analyze` | `spec.md` + this file + `tasks.md` |
| `speckit-implement` | this file (technical context) + per-task harnesses |

---

## Approvals

| Role | Name | Date | Verdict |
|---|---|---|---|
| Author | Planner agent | 2026-05-01 | revised after rejection |
| Constitution check | Planner self-check (§Constitution Check, §Plan Validation Self-Check) | 2026-05-01 | PASS after correction |
| Human gate | (pending) | — | awaiting re-review |
