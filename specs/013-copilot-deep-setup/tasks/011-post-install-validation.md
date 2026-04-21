# Task 011 — Post-install validation via `copilot --agent`

**Phase:** 5 (validation)
**Estimated LOC:** ~100

## Goal

When the `copilot` binary is on PATH, smoke-test each emitted `.agent.yaml` by invoking `copilot --agent <name> -p "ai-setup validation ping" --allow-all-tools -s` with a short timeout. Non-zero exit → single-line warning to stderr. Never fatal.

## Files to touch

| File | Change |
|---|---|
| `internal/adapter/copilot_validate.go` (new) | `ValidateCopilotInstall(ctx, runner)` — enumerate emitted agents from `ctx.FileRecords`, run the smoke per agent with a 5s timeout. Collect warnings, return them for caller to log. |
| `internal/adapter/copilot.go` | `CopilotAdapter.CanRunHeadless` returns true when `LookupCopilotBinary` succeeds. `RunHeadlessValidation` calls `ValidateCopilotInstall` with the default `ExecCopilotCLIRunner`. |
| `internal/adapter/copilot_validate_test.go` (new) | Inject fake runner recording invocations; assert one call per emitted agent with expected args; assert warning collected when fake returns non-zero. |

## Invocation shape

```
copilot --agent <name> -p "ai-setup validation ping" --allow-all-tools -s
```

With a 5-second timeout via `context.WithTimeout`. Output is discarded on success; captured for the warning on failure.

## Acceptance criteria

- [ ] `CanRunHeadless` returns true iff `copilot` on PATH
- [ ] One smoke invocation per emitted agent file
- [ ] Timeout of 5s per invocation
- [ ] Non-zero exits produce warnings, not errors
- [ ] Warnings include agent name + truncated stderr
- [ ] Tests cover both "fake success" and "fake non-zero" paths via injected runner

## Test plan

- `TestValidateCopilotInstall_AllAgentsPass`
- `TestValidateCopilotInstall_OneAgentFails_WarnsOthersStillRun`
- `TestValidateCopilotInstall_BinaryMissing_NoOp`

## Notes

- The validation ping itself costs a premium request per agent against the user's Copilot quota. Keep the prompt short; document in commit that users can disable via `--skip-validation` flag if we add one (out of scope this spec).
- Alternatively in a follow-up: change the ping to a YAML-parse-only check (e.g. `copilot --agent <name> --help`) — research didn't confirm such a mode exists. Left for Phase 6 polish if we find one.
