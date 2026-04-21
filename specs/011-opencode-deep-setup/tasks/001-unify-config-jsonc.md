# Task 001 — Unify opencode config on `.jsonc`

**Phase:** 1
**Status:** ✅ complete (2026-04-19)
**Depends on:** none

## Implementation Notes

- Added package-level `OpenCodeConfigFilename = "opencode.jsonc"` constant in `internal/adapter/opencode.go` shared with `compileOpenCodeMCP` in `mcp_compiler.go`.
- Migration path: if `<root>/opencode.json` pre-exists, copy to `opencode.json.bak` (once — subsequent runs leave the bak alone), copy to `opencode.jsonc` when absent, then remove `.json`.
- Default-config write is now gated on absence of `.jsonc` — prevents silent clobber of user customizations like `permission.edit == "allow"` across re-runs.
- Tests: updated `TestOpenCodeAdapter_Install_FromFS` to assert `.jsonc` and absence of `.json`; added `TestOpenCodeAdapter_Install_MigratesJsonToJsonc`; updated `TestScaffoldAll_OpenCode` to match new filename.

## Verification

- `go test ./... -count=1` — PASS
- `go vet ./...` — clean

## Scope

Collapse opencode's two-config-file split onto a single `opencode.jsonc` target at install-time and compile-time. One-shot migration from any existing `opencode.json`.

## Changes

- `internal/adapter/opencode.go`:
  - Change `configPath` to `opencode.jsonc`.
  - Remove the "skip merge if .jsonc exists" branch.
  - Add migration: if `opencode.json` exists at install time → read it, write `.bak` sidecar of the `.json`, merge contents forward into the new `.jsonc` write, delete the `.json`.
- `internal/adapter/mcp_compiler.go#compileOpenCodeMCP`:
  - Already targets `.jsonc` — verify and factor the filename into a shared constant (`openCodeConfigFilename = "opencode.jsonc"`) used by both install and compile.

## Tests

- `internal/adapter/opencode_test.go`:
  - Install on empty dir → `.opencode/opencode.jsonc` exists, `.json` does not.
  - Install with pre-existing `opencode.json` containing `{"foo":"bar"}` → `.jsonc` exists with `foo:"bar"` preserved; `.json` is gone; `.json.bak` sidecar has original content.
  - Install at each scope (project/workspace/global) — scope-parity check.

## Definition of Done

- All three scopes yield exactly one `opencode.jsonc`; no `opencode.json` remaining.
- `.json` pre-existence path covered by a test.
- `go test ./... -count=1` green.
