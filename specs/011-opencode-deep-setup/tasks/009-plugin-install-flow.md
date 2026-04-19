# Task 009 — Plugin install flow via `opencode plugin <module>`

**Phase:** 5
**Status:** ✅ complete (2026-04-19)
**Depends on:** 007, 008

## Scope

Optional wizard step for opencode plugins. For each selected plugin, shell out to `opencode plugin <module>` with appropriate flags per scope.

## Changes

- `library/opencode/plugins.json` (new): curated list of plugin npm module names with short descriptions.
- `tui/wizard/`:
  - New step: "OpenCode plugins" — multi-select from `plugins.json`. Shown only if (a) opencode selected, (b) `exec.LookPath("opencode")` succeeds.
- `internal/db/`:
  - Add `OpenCodePlugins []string` to the selection schema.
- `internal/adapter/opencode.go`:
  - After the main `Install()` finishes, if plugins selected, iterate and run:
    - Global scope: `opencode plugin <module> -g`.
    - Project/workspace: `opencode plugin <module>` with `cwd` = target dir.
  - Shell-out uses the same injectable `CmdRunner` pattern as task 008.
  - Failures → non-fatal warnings via logger.

## Tests

- `internal/adapter/opencode_plugin_test.go` (new):
  - Mocked runner captures invocation → assert correct flags per scope.
  - Binary absent → no-op.
  - Plugin install failure → warning logged, init exits 0.
- `internal/db/store_test.go`: roundtrip for `OpenCodePlugins`.

## Definition of Done

- Plugin step shown only when opencode is on PATH and opencode tool is selected.
- Correct scope-based flags verified in tests.
- Failures are non-fatal.
