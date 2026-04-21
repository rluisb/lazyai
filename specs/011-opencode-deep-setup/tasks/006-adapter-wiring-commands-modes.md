# Task 006 — Adapter wiring for commands + modes

**Phase:** 3
**Status:** ✅ complete (2026-04-19)
**Depends on:** 005

## Implementation Notes

- New ID types in `internal/types/types.go`: `OpenCodeCommandId` (with consts for `review`/`test`/`commit`) and `OpenCodeModeId` (with consts for `plan`/`audit`). Distinct keyspaces from Gemini's `CommandId` and Copilot's `ChatModeId`.
- `AdapterSelections` (`internal/adapter/types.go`) gained `OpenCodeCommands []types.OpenCodeCommandId` and `OpenCodeModes []types.OpenCodeModeId`.
- `CopyLibraryDirectoryOption.SelectionKey` now also accepts `"opencodeCommands"` and `"opencodeModes"`; both `copyLibraryDirectoryFromFS` and `copyLibraryDirectoryFromDisk` build the selection sets and apply them in their switch-case filter (keeping the pattern identical to existing `commands`/`chatmodes` cases).
- `OpenCodeAdapter.Install` now `EnsureDir`s `<root>/modes` and issues two additional `CopyLibraryDirectory` calls pointing at `opencode/commands` and `opencode/modes`. Empty selection slices preserve the existing "install all" default.
- Tests:
  - Extended `createTestFS()` with synthetic opencode asset entries so the in-memory FS exercises the new code paths.
  - `TestOpenCodeAdapter_InstallsCommandsAndModes` — 3-scope subtest grid asserting every starter asset lands at the expected path.
  - `TestOpenCodeAdapter_SelectionFiltersCommandsAndModes` — narrows the selection to `review` + `plan`, asserts the other commands/modes are absent.

## Verification

- `go test ./... -count=1` — PASS
- `go vet ./...` — clean

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
