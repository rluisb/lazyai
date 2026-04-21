# Task 002 — `instructions` key resolution at all scopes

**Phase:** 1
**Status:** ✅ complete (2026-04-19)
**Depends on:** 001

## Implementation Notes

- No code change required: `InstallToolContextFiles` at `shared.go:378-383` already writes `<ToolDir>/AGENTS.md` in addition to the per-subdir ones (`agents/AGENTS.md`, `skills/AGENTS.md`). The `instructions: ["AGENTS.md"]` entry in `opencode.jsonc` resolves to `<root>/AGENTS.md` which is present at every scope.
- Added table-driven scope-parity test `TestOpenCodeAdapter_InstructionsKeyResolves` covering project/workspace/global. Each subtest parses the installed `opencode.jsonc`, resolves every `instructions` entry relative to the config directory, and asserts the resolved file exists and is non-empty.
- Also extended `TestOpenCodeAdapter_Install_FromFS` key-files list to include `.opencode/AGENTS.md`.

## Verification

- `go test ./... -count=1` — PASS
- `go vet ./...` — clean

## Scope

Ensure the `instructions` key in `opencode.jsonc` points at a file that actually exists at install time, at every scope. Workspace scope must not mirror workspace memory docs into `.opencode/`; the key resolves to an AGENTS.md that lives inside `.opencode/` itself.

## Changes

- `internal/adapter/opencode.go`:
  - Confirm `InstallToolContextFiles` writes `<root>/AGENTS.md` (not just `agents/AGENTS.md` and `skills/AGENTS.md`). Inspect `InstallToolContextFilesOption` — if the root AGENTS.md isn't emitted, add it.
  - Keep `"instructions": ["AGENTS.md"]` — relative path is resolved against the config file's directory (i.e., `<root>/.opencode/`), so this points at `<root>/.opencode/AGENTS.md`.

## Tests

- `internal/adapter/opencode_test.go`:
  - After install at each scope, assert `<root>/AGENTS.md` exists and is non-empty.
  - Parse `opencode.jsonc`; assert `instructions == ["AGENTS.md"]`.
  - Resolve the `instructions` entry relative to config path; assert the resolved file exists.

## Definition of Done

- All 3 scope test cases green.
- The `instructions` key always points at a real file after install.
