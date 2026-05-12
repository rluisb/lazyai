# Plan: 182-workspace-scope-fix

**Feature ID:** 182
**Spec:** N/A (bugfix)
**Date:** 2026-05-11
**Status:** Draft
**Owner:** TBD
**Constitution:** [`.specify/memory/constitution.md`](../../.specify/memory/constitution.md)

> **Purpose.** Fix workspace scope scaffolding so planning artifacts (`.specify/`, `specs/`, `KNOWLEDGE_MAP.md`, `CODEOWNERS`) land in the designated documentation/planning repo instead of the workspace root. Also fix global scope to avoid generating meaningless project-scoped artifacts.

---

## Summary

When installing with workspace scope, `ScaffoldAll()` passes `ctx.TargetDir` (the workspace root) to every scaffold function. Tool configs correctly route to `WorkspaceRoot` via `adapter.ResolveToolRoot()`, but planning artifacts all land at the workspace root instead of the documentation repo. The `PlanningRepoPath` field exists but defaults to `config.HomeDir` (wrong) and is only used in Step 10 as an afterthought.

The fix introduces a `PlanningTargetDir` concept that routes planning steps to the documentation repo for workspace scope, and adds scope-gating to skip project-scoped files in global scope.

---

## Technical Context

| Aspect | Decision | Rationale |
|---|---|---|
| Language(s) | Go 1.26 | Existing codebase |
| Framework(s) | None (standard library + existing scaffold package) | No new dependencies |
| Storage | N/A | Filesystem-only changes |
| Deployment | N/A | CLI binary |
| Telemetry | N/A | No new telemetry |

**External dependencies (new):** None.
**External dependencies (rejected):** None.

---

## Constitution Check

| Article | Verdict | Justification |
|---|---|---|
| I — Library-First | PASS | Uses existing scaffold functions and types; no new libraries needed. |
| II — Test-First (NON-NEGOTIABLE) | PASS | Tests written first for workspace + global scope routing. Each scaffold function gets a test verifying correct target directory. |
| III — Docs as Source of Truth | PASS | This plan documents the fix approach; issue #182 body documents the bug. |
| IV — Anti-Speculation (YAGNI) | PASS | Only fixes the documented bug. No new features, no refactoring beyond what's needed. |
| V — Simplicity Over Abstraction | PASS | Adds one field (`PlanningTargetDir`) to `ScaffoldContext` and routes existing functions through it. No new abstractions. |
| VI — Anti-Overengineering (NON-NEGOTIABLE) | PASS | Minimal changes: 1 new field, 4-6 function signature updates, scope guards in 2 functions. |

**Verdict:** APPROVED

---

## Project Structure

```
packages/cli/
├── internal/scaffold/
│   ├── types.go                          ← modified (add PlanningTargetDir)
│   ├── scaffold.go                       ← modified (route steps through PlanningTargetDir)
│   ├── infra.go                          ← modified (add SetupScope param, global guard)
│   ├── constitution.go                   ← modified (accept planningDir param)
│   ├── specs.go                          ← modified (accept planningDir param)
│   └── *_test.go                         ← new/modified (workspace + global scope tests)
├── cmd/
│   └── helpers.go                        ← modified (wire PlanningRepoPath from wizard selection)
└── internal/types/
    └── types.go                          ← possibly modified (if SetupScope needs new helpers)
```

---

## Root Cause Analysis

### Problem 1: Single `TargetDir` for dual-path concerns

`ScaffoldContext` has one `TargetDir` used for everything. In workspace scope:
- Tool configs (`.claude/`, `.opencode/`, `.github/`) → should go to `WorkspaceRoot` ✅ (already works via `adapter.ResolveToolRoot()`)
- Planning artifacts (`.specify/`, `specs/`, `KNOWLEDGE_MAP.md`, `CODEOWNERS`) → should go to documentation repo ❌ (currently goes to `TargetDir` = workspace root)

### Problem 2: `PlanningRepoPath` defaults to `config.HomeDir`

At `cmd/helpers.go:215`: `PlanningRepoPath: config.HomeDir` — this is wrong. It should be the documentation repo path selected during the wizard.

### Problem 3: No scope-gating on infra files

`ScaffoldInfra` takes only `targetDir` with no `SetupScope` parameter. It unconditionally writes `KNOWLEDGE_MAP.md`, `CODEOWNERS`, etc. even in global scope where they have no meaningful content.

---

## Fix Design

### Step 1: Add `PlanningTargetDir` to `ScaffoldContext`

```go
type ScaffoldContext struct {
    // ... existing fields ...
    
    // PlanningTargetDir is where planning artifacts (.specify/, specs/, 
    // KNOWLEDGE_MAP.md, CODEOWNERS) are installed. For project scope, 
    // equals TargetDir. For workspace scope, equals the documentation 
    // repo path. For global scope, empty (artifacts skipped).
    PlanningTargetDir string
}
```

### Step 2: Wire `PlanningTargetDir` in `buildScaffoldContext`

In `cmd/helpers.go`, set `PlanningTargetDir` based on scope:
- **Project scope:** `PlanningTargetDir = config.TargetDir`
- **Workspace scope:** `PlanningTargetDir = <documentation repo path from wizard>`
- **Global scope:** `PlanningTargetDir = ""` (signals skip)

The wizard already collects the documentation repo path. It just needs to be wired to `PlanningTargetDir` instead of `PlanningRepoPath`.

### Step 3: Route planning steps through `PlanningTargetDir`

Update `ScaffoldAll()` in `scaffold.go`:

| Step | Function | Current param | New param |
|------|----------|---------------|-----------|
| 1 | `ScaffoldConstitution` | `ctx.TargetDir` | `planningDir(ctx)` |
| 4 | `ScaffoldSpecs` | `ctx.TargetDir` | `planningDir(ctx)` |
| 6 | `ScaffoldInfra` | `ctx.TargetDir` | `planningDir(ctx)` + `ctx.SetupScope` |

Where `planningDir(ctx)` returns `ctx.PlanningTargetDir` if set, else `ctx.TargetDir` (backward compat).

### Step 4: Add scope-gating to `ScaffoldInfra`

Add `SetupScope` parameter to `ScaffoldInfra`. In global scope, skip:
- `KNOWLEDGE_MAP.md`
- `CODEOWNERS`
- `compliance.md`

Keep in global scope:
- Pre-commit hook (if applicable)

### Step 5: Update `ScaffoldConstitution` and `ScaffoldSpecs` signatures

Both functions currently take `targetDir string`. Add a `planningDir string` parameter (or change the first parameter to be the planning directory). The simplest approach: change the `targetDir` parameter to mean "planning directory" for these functions, since that's what they actually need.

### Step 6: Update `ScaffoldCompiledRoot` planning directory

`ScaffoldCompiledRoot` already has a `PlanningDir` field in its options struct. Ensure it receives `PlanningTargetDir` for workspace scope.

---

## Task Breakdown

### Task 1: Add `PlanningTargetDir` field and helper function
- Add `PlanningTargetDir string` to `ScaffoldContext` in `types.go`
- Add `func planningDir(ctx *ScaffoldContext) string` helper in `scaffold.go`
- **File:** `packages/cli/internal/scaffold/types.go`, `packages/cli/internal/scaffold/scaffold.go`

### Task 2: Wire `PlanningTargetDir` in `buildScaffoldContext`
- Update `cmd/helpers.go` to set `PlanningTargetDir` based on scope
- For workspace scope: use the documentation repo path from wizard result
- For project scope: use `config.TargetDir`
- For global scope: leave empty
- **File:** `packages/cli/cmd/helpers.go`

### Task 3: Route `ScaffoldConstitution` through planning directory
- Update `ScaffoldConstitution` call in `ScaffoldAll` to use `planningDir(ctx)`
- Update `ScaffoldConstitution` function signature if needed
- **File:** `packages/cli/internal/scaffold/scaffold.go`, `packages/cli/internal/scaffold/constitution.go`

### Task 4: Route `ScaffoldSpecs` through planning directory
- Update `ScaffoldSpecs` call in `ScaffoldAll` to use `planningDir(ctx)`
- **File:** `packages/cli/internal/scaffold/scaffold.go`, `packages/cli/internal/scaffold/specs.go`

### Task 5: Add scope-gating to `ScaffoldInfra`
- Add `scope types.SetupScope` parameter to `ScaffoldInfra`
- Skip `KNOWLEDGE_MAP.md`, `CODEOWNERS`, `compliance.md` in global scope
- Update call in `ScaffoldAll` to pass `ctx.SetupScope`
- **File:** `packages/cli/internal/scaffold/infra.go`, `packages/cli/internal/scaffold/scaffold.go`

### Task 6: Update `ScaffoldCompiledRoot` planning directory
- Ensure `PlanningDir` in `ScaffoldCompiledRootOptions` receives `planningDir(ctx)`
- **File:** `packages/cli/internal/scaffold/scaffold.go`

### Task 7: Tests — workspace scope routing
- Test that workspace scope routes planning artifacts to `PlanningTargetDir`
- Test that tool configs still route to `WorkspaceRoot`
- **File:** `packages/cli/internal/scaffold/scaffold_test.go`

### Task 8: Tests — global scope gating
- Test that global scope skips `KNOWLEDGE_MAP.md`, `CODEOWNERS`, `specs/`
- Test that global scope still creates tool configs at global paths
- **File:** `packages/cli/internal/scaffold/scaffold_test.go`

---

## Risks

| Risk | Level | Mitigation |
|---|---|---|
| Wizard doesn't expose documentation repo path to `buildScaffoldContext` | Medium | Inspect wizard result structure; may need to thread the value through an additional field. |
| Existing tests assume `TargetDir` for all scaffold steps | Medium | Update test fixtures to use `PlanningTargetDir` where appropriate. |
| `ScaffoldRepoRoots` (Step 10) already uses `PlanningRepoPath` | Low | This is correct behavior; just ensure `PlanningTargetDir` and `PlanningRepoPath` are consistent. |

---

## Dependencies

- None. This is a self-contained bugfix.

---

## Open Questions

1. **Wizard result structure:** Does the wizard result already contain the documentation repo path, or does it need to be extracted from the repo selection step?
2. **`PlanningRepoPath` vs `PlanningTargetDir`:** Should these be unified into one field, or kept separate? (Recommendation: unify — they serve the same purpose.)
