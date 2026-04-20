# Task 004 — Adapter: emit `.github/agents/` + `.github/instructions/`

**Phase:** 2 (VS-Code surface parity, project/workspace scope)
**Estimated LOC:** ~100

## Goal

Wire `CopilotAdapter.Install` to copy library agents + instructions into the scope-resolved root. At project/workspace scope this targets `<target>/.github/agents/` and `<target>/.github/instructions/`. Global scope emission is deferred to task 009.

## Files to touch

| File | Change |
|---|---|
| `internal/adapter/copilot.go` | Add `copyCopilotAgents(ctx, root)` and `copyCopilotInstructions(ctx, root)` helpers. Call both inside `Install()`. Honor selection sets (`ctx.Selections.Agents`, `ctx.Selections.Prompts`/Instructions as appropriate). Gate orchestrator on `IsOrchestratorEnabled(ctx)`. |
| `internal/adapter/copilot_test.go` (new) | Table-driven test for project + workspace scope: assert each selected file lands at the expected path; assert orchestrator gating. |

## Acceptance criteria

- [ ] `.github/agents/<name>.agent.yaml` exists for every selected agent from `library/copilot/agents/` at project and workspace scopes
- [ ] `.github/instructions/<name>.instructions.md` exists for every selected instruction at project and workspace scopes
- [ ] `orchestrator.agent.yaml` only emitted when `IsOrchestratorEnabled(ctx)` is true
- [ ] File contents are byte-identical to library source (we do not transform agent YAML)
- [ ] `FileRecords` include the new emissions with `Owner: FileOwnerLibrary`

## Test plan

- `TestCopilot_EmitAgents_ProjectScope`
- `TestCopilot_EmitAgents_WorkspaceScope`
- `TestCopilot_EmitInstructions_ProjectScope`
- `TestCopilot_OrchestratorGated`

## Notes

- Reuse `CopyLibraryDirectory` pattern from the existing chatmodes copy (`copilot.go:68-78`) where possible.
- Do not introduce new selection key types unless the existing `Selections.Agents` set is insufficient — for instructions, piggyback on a new `Selections.Instructions` set or, if that's a larger change, hardcode all three instructions files into the install for now (note in commit).
