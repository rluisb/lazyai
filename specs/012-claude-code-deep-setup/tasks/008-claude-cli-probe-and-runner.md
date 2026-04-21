# Task 008 — Claude CLI probe + injectable runner

**Phase:** 3 (CLI orchestration for MCP)
**Estimated LOC:** ~90

## Goal

Add a thin, testable wrapper around the `claude` binary: a probe that returns its path (or "not found") and a runner interface with an injectable default implementation. Mirrors the pattern used in `opencode_validate.go` (`CmdRunner`).

## Files to create / touch

| File | Change |
|---|---|
| `internal/adapter/claude_cli.go` (new) | Define `ClaudeCLIRunner` interface with `Run(ctx, args ...string) (stdout []byte, stderr []byte, err error)`. Provide `DefaultClaudeCLIRunner` backed by `exec.Command`. Define `LookupClaudeBinary() (path string, found bool)` using `exec.LookPath("claude")`. |
| `internal/adapter/claudecode.go` | Wire an optional `runner ClaudeCLIRunner` field on the adapter struct; default to the real runner but allow tests to inject. |

## Runner interface sketch

```go
type ClaudeCLIRunner interface {
    Run(ctx context.Context, workingDir string, args ...string) (stdout, stderr []byte, err error)
}
```

`workingDir` is set to the workspace / project root when invoking at project scope (so `claude mcp add --scope project` lands in the right `.mcp.json`). Empty at user scope.

## Acceptance criteria

- [ ] `LookupClaudeBinary()` returns `(path, true)` when `claude` is on PATH and `("", false)` when not
- [ ] `ClaudeCLIRunner` can be mocked in tests (interface, not a concrete struct)
- [ ] Default runner handles non-zero exit by returning `err` with stderr attached
- [ ] No existing call site is affected (pure addition)

## Test plan

- Unit test: `LookupClaudeBinary()` with PATH containing a fake `claude` shim, and with PATH empty.
- Unit test: default runner with a known-good command (e.g. `claude --version`), assert stdout contains version string and err is nil.
- Fake runner used by Tasks 009, 010, 013.

## Notes

- Keep this task CLI-invocation-free — no MCP logic here. Just the substrate.
