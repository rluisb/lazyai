# Plan: 182-workspace-scope-fix

**Feature ID:** 182  
**Spec:** GitHub issue #182  
**Date:** 2026-05-11  
**Status:** Draft — awaiting human approval for implementation  
**Owner:** AI-assisted  
**Constitution:** Repository AGENTS.md process gates

> **Purpose.** Fix workspace-scope scaffolding so planning artifacts (`.specify/`, `specs/`, `KNOWLEDGE_MAP.md`, `CODEOWNERS`) land in the planning/documentation repo while AI tool configs stay at the workspace root. Also prevent global scope from creating project-specific planning artifacts.

---

## Summary

Issue #182 is caused by workspace scope collapsing two different roots into one path. Existing adapter contracts already describe the intended model: in workspace scope, `TargetDir` is the planning repo and `WorkspaceRoot` is the workspace root where tool configs live. The current init/scaffold path does not consistently preserve that split: `ScaffoldAll()` routes planning artifacts through `ctx.TargetDir` without scope guards, `buildScaffoldContext()` sets `PlanningRepoPath` to `HomeDir`, `runInit()` declares but does not read `--workspace-root`, and store persistence omits `WorkspaceRoot` / `PlanningRepoPath`.

The implementation should use existing fields (`TargetDir`, `WorkspaceRoot`, `PlanningRepoPath`) rather than inventing a second planning concept. Route planning/project artifacts through a single planning target helper, pass `WorkspaceRoot` into adapters/MCP compile, and skip project-specific artifacts for global scope.

---

## Technical Context

| Aspect | Decision | Rationale |
|---|---|---|
| Language(s) | Go 1.26 | Existing CLI implementation |
| Framework(s) | Cobra + internal scaffold/adapter packages | Existing command/scaffold architecture |
| Storage | Existing SQLite store fields | `types.Config` and DB already have `workspace_root` and `planning_repo_path` |
| Deployment | CLI binary | No runtime service changes |
| Telemetry | Existing logs only | No new telemetry required |

**External dependencies (new):** None.

**External dependencies (rejected):** None.

---

## Constitution Check

| Article | Verdict | Justification |
|---|---|---|
| I — Library-First | PASS | Uses existing scaffold, adapter, and store primitives. |
| II — Test-First (NON-NEGOTIABLE) | PASS | Add regression tests for workspace routing and global-scope skipping before production edits. |
| III — Docs as Source of Truth | PASS | This plan documents the issue-specific implementation path; issue #182 remains the behavior contract. |
| IV — Anti-Speculation (YAGNI) | PASS | Does not add a full docs-repo picker; only wires existing `--workspace-root`/planning-root model and scope guards. |
| V — Simplicity Over Abstraction | PASS | Reuses existing `PlanningRepoPath` and `WorkspaceRoot` instead of adding `PlanningTargetDir`. |
| VI — Anti-Overengineering (NON-NEGOTIABLE) | PASS | Changes are limited to routing, persistence, and focused tests. |

**Verdict:** PASS — implementation still requires explicit human approval after this plan.

---

## Project Structure

```text
packages/cli/
├── cmd/
│   ├── init.go                         ← read --workspace-root, pass split roots to context/compile
│   ├── helpers.go                      ← derive and persist WorkspaceRoot / PlanningRepoPath
│   ├── init_test.go                    ← context wiring regression tests
│   └── helpers_store_test.go           ← store persistence regression test
└── internal/scaffold/
    ├── artifacts.go                    ← pass WorkspaceRoot to adapters
    ├── scaffold.go                     ← route planning/project artifacts and repo ledgers through planning root
    ├── infra.go                        ← add scope guard for global project-specific files
    └── scaffold_test.go                ← workspace/global scaffold regression tests
```

---

## Data Model

No schema migration is required.

Existing fields already support the fix:

| Entity | Storage | Fields | Notes |
|---|---|---|---|
| Config | SQLite config table / `types.Config` | `target_dir`, `workspace_root`, `planning_repo_path` | Already present in migrations/store read/write |
| ScaffoldContext | In-memory | `TargetDir`, `PlanningRepoPath` | Add `WorkspaceRoot` to match adapter/compile contracts |

**Migrations:** None.

**Backfill / data movement:** None required.

---

## Internal Contracts

| Contract | Producer | Consumer | Shape |
|---|---|---|---|
| Workspace planning root | `buildScaffoldContext` | `ScaffoldAll`, repo ledgers | `planningRoot(ctx)` returns `ctx.PlanningRepoPath` for workspace when set, otherwise `ctx.TargetDir`; returns empty for global planning-only steps. |
| Workspace tool root | `runInit` / `buildScaffoldContext` | adapters, MCP compile | `ctx.WorkspaceRoot` is set from `--workspace-root` for workspace scope, defaulting to `ctx.TargetDir` for backward compatibility. |
| Global project-artifact guard | `ScaffoldAll` / `ScaffoldInfra` | constitution/specs/infra scaffolding | Global scope skips `.specify/`, `specs/`, `KNOWLEDGE_MAP.md`, `CODEOWNERS`, and compliance/project specs. |

---

## Phases & Milestones

| Phase | Goal | Exit criterion |
|---|---|---|
| 1 — Regression tests | Capture #182 failures before code changes | Tests fail under current behavior for workspace planning-root routing and global project-artifact skipping. |
| 2 — Root wiring | Preserve workspace root vs planning repo in init context/store | `buildScaffoldContext` and store tests pass for workspace split roots. |
| 3 — Scaffold routing | Route planning artifacts to planning root and tool artifacts to workspace root | Workspace scaffold test shows `.specify/`, `specs/`, `KNOWLEDGE_MAP.md`, `CODEOWNERS` in planning repo; `.opencode/`, `.claude/`, `.github/`, MCP outputs at workspace root. |
| 4 — Global guard | Suppress project-specific artifacts in global scope | Global scaffold test shows no project specs/knowledge/codeowners in target dir, while global tool outputs still work. |
| 5 — Verification | Run targeted Go tests | `go test ./cmd ./internal/scaffold ./internal/adapter` passes from `packages/cli`. |

---

## Task Breakdown

### T1 — Add failing workspace scaffold regression test

**File:** `packages/cli/internal/scaffold/scaffold_test.go`

Test setup:
- `workspaceRoot := t.TempDir()`
- `planningRepo := filepath.Join(workspaceRoot, "bee-gone")`
- `ctx.SetupScope = workspace`
- `ctx.TargetDir = planningRepo`
- `ctx.PlanningRepoPath = planningRepo`
- `ctx.WorkspaceRoot = workspaceRoot`
- selected tools include OpenCode/Claude as needed

Done when:
- Test expects `.specify/`, `specs/`, `KNOWLEDGE_MAP.md`, `CODEOWNERS` under `planningRepo`.
- Test expects `.opencode/` / `.claude/` under `workspaceRoot`.
- Test expects those planning artifacts not to be created at `workspaceRoot`.

### T2 — Add failing global scaffold regression test

**File:** `packages/cli/internal/scaffold/scaffold_test.go`

Done when:
- Global scope does not create `.specify/`, `specs/`, `KNOWLEDGE_MAP.md`, or `CODEOWNERS` in the project/current dir.
- Global tool artifacts still resolve through `HomeDir` where supported.

### T3 — Add context/store wiring tests

**Files:**
- `packages/cli/cmd/init_test.go`
- `packages/cli/cmd/helpers_store_test.go`

Done when:
- `buildScaffoldContext` sets workspace `PlanningRepoPath` to the planning repo/current target dir, not `HomeDir`.
- `buildScaffoldContext` sets `WorkspaceRoot` from config/flag for workspace scope.
- `writeStoreFromScaffoldResult` persists `WorkspaceRoot` and `PlanningRepoPath`.

### T4 — Wire workspace root flag and scaffold context

**Files:**
- `packages/cli/cmd/init.go`
- `packages/cli/cmd/helpers.go`
- `packages/cli/internal/scaffold/types.go`

Implementation notes:
- Read existing `--workspace-root` flag in `runInit`.
- Add `CLIWorkspaceRoot` to `wizard.WizardConfig` if needed.
- Add `WorkspaceRoot` to `scaffold.ScaffoldContext`.
- For workspace scope:
  - `TargetDir` remains the planning repo/current dir.
  - `PlanningRepoPath` defaults to `TargetDir` unless explicitly provided later by UX.
  - `WorkspaceRoot` defaults to `--workspace-root`, falling back to `TargetDir` for backward compatibility.
- For project scope: `PlanningRepoPath = TargetDir`, `WorkspaceRoot = ""`.
- For global scope: `PlanningRepoPath = ""`, `WorkspaceRoot = ""`.

### T5 — Route adapters and MCP compile to workspace root

**Files:**
- `packages/cli/internal/scaffold/artifacts.go`
- `packages/cli/cmd/init.go`

Done when:
- Adapter `Install` receives `WorkspaceRoot: ctx.WorkspaceRoot`.
- `compileMCPForInit` receives `WorkspaceRoot: ctx.WorkspaceRoot` instead of always `ctx.TargetDir`.

### T6 — Route planning artifacts and repo ledgers to planning root

**File:** `packages/cli/internal/scaffold/scaffold.go`

Done when:
- `ScaffoldConstitution`, `ScaffoldSpecs`, and `ScaffoldInfra` use `planningRoot(ctx)` where appropriate.
- Workspace repo roots and ledgers use the same planning root.
- Global scope skips planning-only steps when planning root is empty.

### T7 — Add scope guard to infra scaffolding

**File:** `packages/cli/internal/scaffold/infra.go`

Done when:
- `ScaffoldInfra` receives `setupScope` or equivalent scope signal.
- Global scope skips project-specific infra (`Compliance`, `KnowledgeMap`, `Codeowners`).
- Pre-commit remains a no-op unless a real `.git` exists.

### T8 — Targeted verification

**Command:** from `packages/cli`

```bash
go test ./cmd ./internal/scaffold ./internal/adapter
```

Done when:
- Targeted tests pass.
- No unrelated root checkout files are modified.

---

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation | Owner |
|---|---|---|---|---|
| `--workspace-root` semantics differ from existing expectations | M | M | Keep fallback to `TargetDir` when absent; document that split-root workspace mode requires `--workspace-root` or future interactive UX. | implementer |
| Existing tests assume project-scope target behavior | M | L | Only adjust tests that use workspace/global scope; project scope remains unchanged. | implementer |
| Tracked file paths become ambiguous across roots | M | M | Preserve relative paths for planning-root files; existing adapter global/workspace tracking patterns remain unchanged. | implementer |
| Headless populate still reads `AGENTS.md` from planning repo while tool AGENTS may be at workspace root | L | M | Leave headless behavior unchanged unless regression exposes failure; #182 scope is scaffold placement. | implementer |

---

## Complexity Tracking

| Item | Simpler alternative | Why complexity is justified | Cost |
|---|---|---|---|
| Add `WorkspaceRoot` to `ScaffoldContext` | Reuse `TargetDir` for all roots | Existing adapter contracts already require separate planning/tool roots in workspace scope. | Small API surface change |
| Add planning root helper | Inline `if workspace` checks | Centralizes root fallback and avoids repeated inconsistent logic. | Small helper to maintain |

---

## Out of Scope

- Interactive documentation-repo picker — no current wizard field exists; this plan preserves current-dir-as-planning-repo and wires existing `--workspace-root`.
- New database migration — relevant fields already exist.
- Reworking workspace repo discovery or `RepoInfo` semantics.
- Changing Copilot skill/agent conversion behavior.
- Fixing orchestrator MCP schemas (already handled separately in root checkout).

---

## Downstream Contract

| Produces for | Filename |
|---|---|
| implementation | this plan |
| review | targeted tests + diff |

---

## Approvals

| Role | Name | Date | Verdict |
|---|---|---|---|
| Author | AI-assisted | 2026-05-11 | Draft |
| Human gate | Pending | 2026-05-11 | pending implementation approval |
