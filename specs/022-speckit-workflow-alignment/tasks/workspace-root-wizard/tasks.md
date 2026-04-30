# Tasks: Workspace Root Wizard UX

**Spec:** 022-speckit-workflow-alignment
**Input:** Research findings (back-end WorkspaceRoot exists, wizard needs front-end)
**Prerequisites:** Spec 022 Phase E2.3 (compile.go + workspace_compile.go + types.go already done)

## Format: `[ID] [P?] Description`

- **[P]**: Can run in parallel (different files/languages, no shared state)
- Include exact file paths

---

## Phase 1: CLI + Config (Foundational — blocks everything else)

- [ ] T001 [P] Add `--workspace-root` flag to TS init command
  - File: `packages/ai-setup-ts/src/commands/init.ts`
  - Add `--workspace-root <path>` option, validate path exists when `--scope workspace`
  - Wire into WizardSelections

- [ ] T002 [P] Add `WorkspaceRoot` to TS types + store schema
  - Files: `packages/ai-setup-ts/src/types.ts`, `packages/ai-setup-ts/src/store/schema.ts`
  - Add `workspaceRoot?: string` to SetupConfig type and zod schema

- [ ] T003 [P] Add `workspace-root` flag to Go init command
  - File: `packages/ai-setup-go/cmd/init.go`
  - Add flag, validate for workspace scope

- [ ] T004 [P] Add `WorkspaceRoot` to Go types
  - File: `packages/ai-setup-go/internal/types/types.go`
  - Add `WorkspaceRoot string` to Config struct

---

## Phase 2: Wizard UX (depends on Phase 1)

- [ ] T005 Add workspace root prompt to TS wizard Phase 1
  - File: `packages/ai-setup-ts/src/wizard/phase1-context.ts`
  - After scope selection (line ~480), before planning repo prompt
  - Prompt: "Where should AI tool configs live? (workspace root)" with default = `path.dirname(planningRepoPath)`
  - Store in wizard state, pass through to Planner/Selections

- [ ] T006 Pass WorkspaceRoot through TS wizard planner
  - File: `packages/ai-setup-ts/src/wizard/planner.ts`
  - Ensure workspaceRoot flows to init/compile commands

- [ ] T007 Add workspace root prompt to Go wizard (non-interactive path)
  - File: `packages/ai-setup-go/cmd/init.go` (already modified in T003)
  - Validate that when `--scope workspace`, `--workspace-root` and `--planning-repo` are both set

---

## Phase 3: Adapter Scope + Compile Wiring (depends on Phase 1)

- [ ] T008 [P] Update Go adapter scope to write tools to workspace root
  - File: `packages/ai-setup-go/internal/adapter/scope.go`
  - When scope is workspace AND WorkspaceRoot is set: use WorkspaceRoot for tool config paths
  - Keep backward compat: if WorkspaceRoot empty, fall back to planning repo (current behavior)

- [ ] T009 [P] Update Go compile to pass workspace root through chain
  - Files: `packages/ai-setup-go/cmd/compile.go`, `packages/ai-setup-go/cmd/init.go`
  - Wire the WorkspaceRoot from init's store into compile's CompileContext

- [ ] T010 Update TS adapter scope for workspace root
  - File: `packages/ai-setup-ts/src/adapters/shared.ts` or registry
  - Mirror Go's behavior: workspace root for tool configs, planning repo for specs

---

## Phase 4: Tests + Verification (depends on all above)

- [ ] T011 [P] Add wizard test for workspace root prompt
  - File: `packages/ai-setup-ts/src/__tests__/wizard-phases.test.ts`
  - Test: workspace scope sets workspaceRoot, passes through to planner

- [ ] T012 [P] Add adapter test for workspace root paths
  - File: `packages/ai-setup-go/internal/adapter/adapter_scope_test.go`
  - Test: workspace scope with WorkspaceRoot set writes to correct directory

- [ ] T013 [P] Add init test for --workspace-root flag
  - File: `packages/ai-setup-ts/src/__tests__/cli.test.ts`
  - Test: `ai-setup init --scope workspace --workspace-root /tmp/ws --planning-repo /tmp/ws/bee-gone`

- [ ] T014 Run full verification
  - `pnpm -r run typecheck` — must pass
  - `pnpm -r run test` — must pass (all 425 tests)
  - `go build . && go test ./...` — must pass

---

## Dependencies

```
T001 + T003 + T004  (parallel, CLI + types, TS + Go)
    │
    ▼
T005 + T002 + T007  (wizard UX + schema, depends on T001/T004)
    │
    ▼
T006 + T008 + T009 + T010  (planner + adapter, depends on T002/T004/T005)
    │
    ▼
T011 + T012 + T013  (tests, parallel)
    │
    ▼
T014  (verification)
```

## Parallel Opportunities

- T001 (TS init) ∥ T003 (Go init) ∥ T004 (Go types) — different languages
- T002 (TS types) ∥ T005 (wizard prompt) — different files
- T008 (Go adapter) ∥ T010 (TS adapter) — different languages
- T011 ∥ T012 ∥ T013 — different test files
