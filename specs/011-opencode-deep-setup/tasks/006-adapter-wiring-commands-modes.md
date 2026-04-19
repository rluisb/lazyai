# Task 006 — Adapter wiring for commands + modes

**Phase:** 3
**Status:** pending
**Depends on:** 005

## Scope

Hook `library/opencode/commands/` and `library/opencode/modes/` into `OpenCodeAdapter.Install` with scope-parity across project/workspace/global.

## Changes

- `internal/adapter/opencode.go`:
  - Add `files.EnsureDir(filepath.Join(ocDir, "modes"))` (already have `commands` from prior patch).
  - Two new `CopyLibraryDirectory` calls:
    - SourceSubdir: `opencode/commands`, SelectionKey: `"opencodeCommands"`, ToDestPath: `<root>/commands/<file>`.
    - SourceSubdir: `opencode/modes`, SelectionKey: `"opencodeModes"`, ToDestPath: `<root>/modes/<file>`.
  - WarnOnSkip: true for both.

## Tests

- `internal/adapter/opencode_test.go`:
  - After install at each scope, assert expected command files under `<root>/commands/` and mode files under `<root>/modes/`.
  - Selection-filter test: passing `"opencodeCommands": ["review"]` only installs `review.md`.
- `internal/adapter/scope_test.go`: no change (paths unchanged from §3.1 in research).

## Definition of Done

- Scope-parity test covers commands + modes at all 3 scopes.
- Selection filter honored for both.
