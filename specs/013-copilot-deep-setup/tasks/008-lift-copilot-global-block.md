# Task 008 — Lift the Copilot × global scope block

**Phase:** 3 (scope flip)
**Estimated LOC:** ~60

## Goal

Flip the scope gating for Copilot so global scope becomes supported. Wire the probe from task 007 so the adapter silently no-ops (with a single-line warning) when both the `copilot` binary and `~/.copilot/` are absent at global scope.

## Files to touch

| File | Change |
|---|---|
| `internal/adapter/scope.go` | Remove L28's `if tool == ToolIdCopilot && scope == SetupScopeGlobal { return false }`. `IsScopeSupported` now returns true for (copilot, global). |
| `internal/adapter/scope.go` | `projectSubdir(ToolIdCopilot)` stays `.github`. Global resolution flows through `globalpaths.ResolveGlobalToolTargetDir`. |
| `internal/globalpaths/globalpaths.go` | L71-82: `IsGlobalInstallSupported(ToolIdCopilot) = true`. `ResolveGlobalToolTargetDir` returns `<home>/.copilot` for Copilot (new case in the switch). |
| `internal/adapter/copilot.go` | In `Install`, when `ctx.SetupScope == SetupScopeGlobal`, probe via task 007 helpers. If neither binary nor `~/.copilot/` present, log one-line warning and return `ctx.FileRecords, nil`. Otherwise continue to task 009's global emitters. |
| `internal/globalpaths/globalpaths_test.go` | Update `TestIsGlobalInstallSupported` expectations for Copilot. |
| `internal/adapter/scope_test.go` | Update assertions at L29-31 and L72-74 (Copilot × global previously expected unsupported). |

## Acceptance criteria

- [ ] `IsScopeSupported(ToolIdCopilot, SetupScopeGlobal) == true`
- [ ] `ResolveGlobalToolTargetDir(ToolIdCopilot, homeDir)` returns `<homeDir>/.copilot`
- [ ] Install at global scope is a silent no-op (log only) when probe reports both absent
- [ ] Install at global scope proceeds when probe reports either binary or `~/.copilot/` present
- [ ] Existing scope-parity tests updated; new (copilot, global) entry pending task 012

## Test plan

- `TestCopilot_GlobalScope_NoCopilotPresent_NoWrites`
- `TestCopilot_GlobalScope_BinaryPresent_EmitsFiles` (stub emitter or assert downstream called — full emit test is task 009)
- `TestCopilot_GlobalScope_HomeDirPresent_EmitsFiles`

## Notes

- Warning string: `"[copilot] global scope requested but neither 'copilot' binary nor ~/.copilot/ present; skipping"`. Matches tone of existing log lines in the adapter.
- After this task, Phase 4 fills in the global-scope emitters — the probe is gate, not content.
