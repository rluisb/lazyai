# Task 007 — `copilot_cli.go` probe + runner interface

**Phase:** 3 (probe substrate)
**Estimated LOC:** ~90

## Goal

Mirror spec 012's `claude_cli.go` substrate for Copilot: binary probe + `CopilotCLIRunner` interface for injectable testability. Also provide a `CopilotHomePresent(homeDir)` helper so we can treat "`~/.copilot/` already exists" as sufficient probe signal even if the binary isn't on PATH (e.g. Homebrew install that hasn't been launched yet).

## Files to create/touch

| File | Change |
|---|---|
| `internal/adapter/copilot_cli.go` (new) | Declare `CopilotCLIRunner` interface with `Run(ctx, args...) ([]byte, error)` and `LookItUp() (string, error)`. Provide `ExecCopilotCLIRunner` default impl using `exec.Command`. Provide `LookupCopilotBinary() (string, bool)` via `exec.LookPath("copilot")`. Provide `CopilotHomePresent(homeDir string) bool` checking `<homeDir>/.copilot` dir existence. |
| `internal/adapter/copilot_cli_test.go` (new) | Test `LookupCopilotBinary` with `t.Setenv("PATH", "...")`; test `CopilotHomePresent` with `t.TempDir()`. |

## Probe policy (used in task 008)

`ProbeResult { BinaryPath string; HomeDir bool }`
- `binary` = `LookupCopilotBinary()` result
- `home`   = `CopilotHomePresent(ctx.HomeDir)` result
- Global-scope emitters fire when `binary` OR `home` is true.
- Validation (task 011) fires only when `binary` is true.

## Acceptance criteria

- [ ] Interface defined, default impl present, fake impl constructible in tests
- [ ] `LookupCopilotBinary` returns the resolved path or empty + false
- [ ] `CopilotHomePresent` uses `ctx.HomeDir` when set, falls back to `os.UserHomeDir()`
- [ ] `go test ./internal/adapter/... -run CopilotCLI` green

## Test plan

Basic unit tests — this task is plumbing only. Behavior tests live in task 008/011.

## Notes

- Match the surface of `claude_cli.go` as closely as possible so a future refactor can dedupe into a shared substrate.
