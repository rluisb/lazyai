# Task 008 — Post-install validation via `opencode debug *`

**Phase:** 4
**Status:** pending
**Depends on:** 003, 006

## Scope

Opt-in, gated post-install validator. Runs `opencode debug config` and `opencode debug agent <name>` when the opencode binary is on PATH. Non-fatal.

## Changes

- `internal/adapter/opencode_validate.go` (new):
  - `type ValidationWarning struct { Scope, Item, Reason string }`.
  - `ValidateOpenCodeInstall(ctx *AdapterContext) ([]ValidationWarning, error)`:
    - `exec.LookPath("opencode")` — if missing, return `(nil, nil)`.
    - Run `opencode debug config` (stdout JSON or best-effort parse) at the install root; assert our mcp entries and dir layout are present.
    - List installed agents from `<root>/agents/*.md`; for each, run `opencode debug agent <name>` and capture parse errors.
    - Return structured warnings.
  - Use an injectable runner interface (`type CmdRunner func(name string, args ...string) ([]byte, error)`) so tests can mock.
- `internal/scaffold/scaffold.go`:
  - After adapter install loop, if opencode was installed, call `ValidateOpenCodeInstall` and log warnings via the existing logger at WARN level.

## Tests

- `internal/adapter/opencode_validate_test.go` (new):
  - Canned stdout → extracts expected warnings.
  - Binary absent → no-op, no error.
  - `opencode debug agent` returning non-zero → warning captured, not propagated as error.

## Definition of Done

- Validator runs only when binary is present.
- Non-fatal; init still exits 0.
- Covered by unit tests with mocked runner.
