# Plan: Orchestrator `init` One-Shot Recommendation Command

**Date:** 2026-04-29
**Status:** Proposed — waiting for approval before implementation
**Research:** `research.md`

---

## Goal

Add a one-shot `ai-setup-orchestrator init` command that discovers available orchestration meta and recommends a fitting orchestration approach for a user task.

The command should help a user answer:

> "Given my current agents, skills, host CLI, and project context, should I use a direct agent, chain, team, or workflow?"

---

## MVP Scope

### In scope

1. Fix runtime catalog correctness so DB-created chains/teams/workflows are actually runnable.
2. Add a shared runtime catalog loader used by both handlers and the new CLI command.
3. Add one-shot `init` CLI command.
4. Add deterministic rule-based recommendations.
5. Add JSON and human-readable output.
6. Add tests.

### Out of scope for MVP

- Interactive TUI wizard.
- LLM-powered `--smart` recommendations.
- `--scaffold` creation of catalog definitions.
- `--start` automatic run launch.
- MCP tool version of `init`.
- Baked-in fallback catalog defaults.

---

## Acceptance Criteria

1. `ai-setup-orchestrator init` prints a read-only inventory report.
2. `ai-setup-orchestrator init --task "build auth from scratch" --host claude-code` prints a recommendation.
3. `--host claude-code` can prefer teams when appropriate because Claude supports parallel teams.
4. `--host opencode` can prefer chains for multi-step work because OpenCode does not support parallel teams.
5. `--json` emits machine-readable JSON with the same core information.
6. Root context files are detected: `AGENTS.md`, `CLAUDE.md`.
7. Runtime catalog loading includes active DB definitions for agents, skills, chains, teams, and workflows.
8. Existing MCP handlers continue to work after refactor.
9. Tests pass:
   - catalog resolver tests
   - init CLI tests
   - full test suite
   - typecheck

---

## Implementation Phases

### Phase 1 — Runtime catalog correctness

Fix the discovered gap where internal DB catalog definitions for `chain`, `team`, and `workflow` are not merged into runtime `OrchestrationCatalog`.

Why first:

- `init` recommendations must reflect definitions that can actually run.
- Future `--scaffold` should safely create DB definitions and immediately use them.
- The current `pedro-ernesto-e-banda` team exists in internal catalog but may not be visible to `build_team` until this is fixed.

Files:

- `src/catalog/resolver.ts`
- `src/__tests__/catalog-resolver.case.ts`

### Phase 2 — Shared runtime catalog helper

Extract the `loadCatalog()` + `resolveCatalog()` flow from `OrchestratorToolHandlers.getCatalog()` into a reusable helper.

Suggested file:

- `src/catalog/runtime.ts`

Suggested API:

```ts
export interface RuntimeCatalogOptions {
  projectRoot: string
  libraryOrchestrationRoot?: string
  libraryAgentsRoot?: string
  hostCli?: HostCli
  db?: Db
}

export function loadRuntimeCatalog(options: RuntimeCatalogOptions): OrchestrationCatalog
```

Then `OrchestratorToolHandlers.getCatalog()` delegates to it.

Files:

- `src/catalog/runtime.ts` (new)
- `src/tool-handlers.ts`
- tests as needed

### Phase 3 — `init` CLI command

Add a new CLI module and wire it into `src/index.ts`.

Suggested flags:

```text
--task <text>       Task/request to recommend orchestration for
--host <host>       claude-code | codex | opencode | gemini | copilot (default: opencode)
--project <path>    Project root (default: cwd)
--json              Machine-readable JSON output
--verbose           Include definition source/path details
-h, --help          Show help
```

Human-readable output sections:

1. Project/context
2. Host CLI capability
3. Catalog inventory
4. Root instruction files
5. Recommendation, if task provided
6. Examples, if no task provided

Files:

- `src/cli/init.ts` (new)
- `src/index.ts`
- `src/__tests__/init-cli.case.ts` (new)

### Phase 4 — Recommendation engine

Keep recommendation logic deterministic and local.

Can live inside `src/cli/init.ts` initially, or split later if it grows.

Rules:

1. Review/audit/security task:
   - prefer team if host supports parallel teams and reviewer/red-team/team exists
   - otherwise direct reviewer or review chain
2. Design/architecture/from-scratch task:
   - prefer workflow/chain with architect/planner/implementor
3. Build/implement/refactor task:
   - prefer RPI workflow/chain if available
   - otherwise direct implementor-senior
4. No task:
   - no recommendation; print examples

Files:

- `src/cli/init.ts`
- `src/__tests__/init-cli.case.ts`

---

## Task Breakdown

Detailed task files:

- `tasks/001-runtime-catalog-db-definitions.md`
- `tasks/002-shared-runtime-catalog-helper.md`
- `tasks/003-init-cli-command.md`
- `tasks/004-init-recommendation-rules.md`
- `tasks/005-verification.md`

---

## Risks and Mitigations

| Risk | Mitigation |
|---|---|
| Init report diverges from runtime handlers | Shared `loadRuntimeCatalog()` helper |
| DB catalog body parsing may fail for malformed JSON chains/teams/workflows | Validate defensively; skip invalid active definitions with clear error or throw in tests |
| Recommendation rules overfit keywords | Keep confidence labels and alternatives |
| Host detection ambiguity | Explicit `--host`; default to `opencode` |
| Scope creep into scaffold/start | Defer flags; read-only MVP |

---

## Suggested Implementation Order

1. Implement Phase 1 and run resolver tests.
2. Implement Phase 2 and run handler/tool tests.
3. Implement Phase 3 with inventory-only output and tests.
4. Implement Phase 4 recommendations and tests.
5. Run full suite + typecheck.

---

## Budget Estimate

| Area | Estimated LOC | Risk |
|---|---:|---|
| Resolver DB chain/team/workflow merge | 80 | Medium |
| Shared runtime catalog helper | 40 | Low |
| Init CLI parser/output | 160 | Medium |
| Recommendation rules | 100 | Medium |
| Tests | 250 | Medium |
| **Total** | **~630 LOC** | **Medium** |

---

## Approval Gate

This is a multi-file, >100 LOC feature. Implementation should start only after plan approval.
