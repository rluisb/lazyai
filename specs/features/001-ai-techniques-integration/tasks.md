> **HISTORICAL (2026-06-20) — parts of this plan target the removed `packages/ai-setup-ts` package; the CLI is now Go-only (`packages/cli`). Retained for historical context.**
>
> This document is historical; `packages/ai-setup-ts` is no longer present.

# Tasks: Feature 001 — Phase W1.A

## Task Index

| # | ID | Name | User Story | Files | Dependencies | [P] |
|---|----|------|------------|-------|--------------|-----|
| 1 | T1 | Schema + migration tests | P1/P2 | Go/TS store schema, Go migration, TS v2→v1 helper | — | — |
| 2 | T2 | FragmentContext fallback coverage | P1/P2 | Go/TS compiler, root/testing templates, constitution scaffold | T1 | — |
| 3 | T3 | Wizard prompts + config wiring | P1/P2 | Go init wizard, TS phase2 wizard/planner, store config writes | T1 | — |
| 4 | T4 | Targeted update + parity snapshots | P1/P2 | Go/TS root scaffold/update path, parity snapshots | T1 | — |
| 5 | T5 | Starter standards content + frontmatter | P2 | Five starter standard markdown files and validators | — | [P] |
| 6 | T6 | Standards file-level copy logic | P2 | Go/TS specs scaffold copy logic | T5 | — |
| 7 | T7 | W1.A integration/golden tests | P1/P2 | End-to-end init/update fixtures and golden outputs | T4, T6 | — |

## Dependency Graph

```text
start
  ├── T1 schema + migration tests
  │     ├── T2 FragmentContext fallback coverage
  │     ├── T3 wizard prompts + config wiring
  │     └── T4 targeted update + parity snapshots ─┐
  │                                                │
  └── T5 starter standards content [P] ───► T6 ────┤
                                                   ▼
                                  T7 W1.A integration/golden tests
```

Hard dependencies from `plan.md`:
- T1 blocks all schema/config writes.
- T2, T3, and T4 depend on T1.
- T5 is independent and can run in parallel with T2–T4.
- T6 depends on T5.
- T7 depends on T4 and T6.

## Task Details

### T1 — Schema + migration tests

- **Objective:** Add test-first coverage for W1.A store schema changes before any wizard/compiler writes depend on them, including TS v1→v2 defaults and v2→v1 downgrade behavior.
- **User story:** P1 — Generate a populated project harness; P2 — Make quality thresholds concrete.
- **Files:** `packages/ai-setup-go/internal/store/migrations/006-add-constitution-fields.sql`, `packages/ai-setup-go/internal/migration/json_bridge.go`, `packages/ai-setup-ts/src/store/schema.ts`, `packages/ai-setup-ts/src/store/migrations/v2-to-v1.ts`, plus package-appropriate tests.
- **Parallel marker:** none; this is the W1.A schema gate.
- **Done When:** failing tests are written first; Go/TS schema defaults are verified; lowdb downgrade removes v2-only fields; coverage threshold validation/default is tested; no wizard/compiler production writes are made before this gate passes.

### T2 — FragmentContext fallback coverage

- **Objective:** Extend compiler/template tests and implementation for `FragmentContext.constitution`, legacy literal fallback markers, coverage threshold rendering, and codebase map rows.
- **User story:** P1/P2.
- **Files:** `packages/ai-setup-go/internal/compiler/template.go`, `packages/ai-setup-go/internal/scaffold/constitution.go`, `packages/ai-setup-go/library/root/AGENTS.template.md`, `packages/ai-setup-go/library/rules/testing.md`, `packages/ai-setup-ts/src/compiler/fragment-resolver.ts`, `packages/ai-setup-ts/src/scaffold/compiled-root.ts`, package tests.
- **Parallel marker:** none; starts after T1. T5 may run in parallel because it is content-only.
- **Done When:** tests prove populated values replace collected placeholders, skipped values render documented fallbacks exactly, generated codebase map excludes ignored directories, and Go/TS output for fields in scope is parity-ready.

### T3 — Wizard prompts + config wiring

- **Objective:** Add test-first wizard/config wiring for project profile fields and stack-aware coverage threshold defaults, implementing Go first and verifying TS parity after.
- **User story:** P1/P2.
- **Files:** `packages/ai-setup-go/cmd/init.go`, `packages/ai-setup-go/internal/migration/json_bridge.go`, `packages/ai-setup-ts/src/wizard/phase2-features.ts`, `packages/ai-setup-ts/src/wizard/planner.ts`, `packages/ai-setup-ts/src/store/schema.ts`, package tests.
- **Parallel marker:** none; starts after T1. T5 may run in parallel.
- **Done When:** headless wizard/config tests fail first; prompts/config collect or default profile/coverage values; out-of-range thresholds are rejected/reset to default; non-interactive/skipped fields remain safe; Go behavior is implemented before TS parity verification.

### T4 — Targeted update + parity snapshots

- **Objective:** Prove `ai-setup update` patches only recognized placeholders/value slots and preserves hand-authored `AGENTS.md` content, with Go/TS snapshot parity for generated W1.A fields.
- **User story:** P1/P2.
- **Files:** `packages/ai-setup-go/internal/scaffold/root.go`, `packages/ai-setup-go/internal/scaffold/constitution.go`, `packages/ai-setup-go/library/root/AGENTS.template.md`, `packages/ai-setup-go/library/rules/testing.md`, `packages/ai-setup-ts/src/scaffold/compiled-root.ts`, `packages/ai-setup-ts/src/compiler/fragment-resolver.ts`, package tests/snapshots.
- **Parallel marker:** none; starts after T1. T5 may run in parallel.
- **Done When:** targeted-update fixtures fail first; `TargetedUpdatePatch` shape is asserted; custom sections/reordered structure are preserved byte-for-byte; warnings are emitted for unsafe parse cases; Go/TS snapshots are byte-identical for fields in scope.

### T5 — Starter standards content + frontmatter [P]

- **Objective:** Author the five starter standards and validate them with existing frontmatter/standard checks before scaffold copy logic is touched.
- **User story:** P2.
- **Files:** `packages/ai-setup-go/library/standards/starter/orchestration-patterns.md`, `test-patterns.md`, `error-handling.md`, `agent-security.md`, `context-loading.md`, plus frontmatter validator tests.
- **Parallel marker:** [P] — independent content task; can run while T2–T4 proceed after T1.
- **Done When:** frontmatter tests fail first; exactly five starter standards exist; all pass existing standard frontmatter validation; at least one concrete standard is available for Quality Gate 4 citation.

### T6 — Standards file-level copy logic

- **Objective:** Implement standards seeding with file-level idempotency so missing starter files copy even when `.gitkeep`/`README.md` exists and user-authored files are never overwritten.
- **User story:** P2.
- **Files:** `packages/ai-setup-go/internal/scaffold/specs.go`, `packages/ai-setup-ts/src/scaffold/specs.ts`, package tests.
- **Parallel marker:** none; depends on T5 starter content.
- **Done When:** copy/idempotency tests fail first; `.gitkeep` and `README.md` do not block missing starter files; existing same-path user files are preserved; Go implementation precedes TS parity verification.

### T7 — W1.A integration/golden tests

- **Objective:** Add final end-to-end and golden coverage for the W1.A constitution, coverage, and standards seed workflow.
- **User story:** P1/P2.
- **Files:** Go/TS init/update integration tests, golden outputs, and fixtures under existing package test locations.
- **Parallel marker:** none; depends on T4 and T6.
- **Done When:** integration/golden tests fail first; accepted wizard/config input yields no collected-field `[YOUR_*]` markers in generated `AGENTS.md`; skipped fields preserve literal fallbacks; coverage threshold appears in `AGENTS.md` and selected `specs/rules/testing.md`; five standards are present without overwriting user files; W1.A ACs are traceable to evidence.

## Constitution Gate

- **Article I (Library-First):** Tasks reuse existing compiler, store, wizard, scaffold, frontmatter validator, and test harnesses; no new frameworks or dependencies are planned.
- **Article II (TDD):** Every task begins with a failing test before production/content changes.
- **Article III (Docs):** Task AC references come from `spec.md`; file scope and execution order come from `plan.md`.
- **Article IV (YAGNI):** Tasks are limited to W1.A N8/N4/N11; W1.B, Wave 2/3/4, MCP config, and runtime orchestration changes are excluded.
- **Article V (Simplicity):** Tasks keep concrete schema/template/scaffold changes; no new template engine, store, or standards subsystem.
- **Article VI (Anti-Overengineering):** No one-caller abstractions or speculative helper extraction are required by these tasks.
