# Task 001 — Global scope: agents/ subdir + context file placement

**Phase:** 1 (structural fixes, no CLI)
**Estimated LOC:** ~80

## Goal

At global scope, write agents under `~/.claude/agents/<name>.md` (matching project scope) instead of flat at `~/.claude/<name>.md`. Relocate the tool-context render target from `~/.claude/CLAUDE.md` (which collides with Claude's personal-conventions file) to `~/.claude/agents/CLAUDE.md`. Preserve any existing `~/.claude/CLAUDE.md` on re-run — template-fill only on first install.

## Files to touch

| File | Change |
|---|---|
| `internal/adapter/claudecode.go` | Remove the `isGlobal ? flat : agents/` branch at ~L82-83. Always use `agents/<name>.md`. Update `InstallToolContextFiles` call site so `AgentsDestDir` is `"agents"` at all scopes (not `"."` at global). |
| `internal/adapter/shared.go` | If `InstallToolContextFiles` has scope-aware branching for Claude, simplify. Add a guard so `~/.claude/CLAUDE.md` is written only if absent (preserves user content on re-run). |
| `internal/adapter/adapter_scope_test.go` | Update expected path assertions: global should have `agents/builder.md` etc., not flat files. |

## Acceptance criteria

- [x] `go run . init` with global scope produces `~/.claude/agents/<name>.md` for every enabled agent
- [x] `~/.claude/agents/CLAUDE.md` (tool-context) exists and reads naturally against its directory
- [x] `~/.claude/CLAUDE.md` is **not** overwritten on re-run; on first install it is template-filled only if the file does not exist
- [x] Flat-layout legacy files at `~/.claude/<name>.md` (from prior buggy installs) are either migrated into `agents/` or left alone with a one-line warning (pick: migrate). Document whichever is chosen in the commit message.
- [x] Scope-parity test covers the new layout

## Test plan

- Update `TestAdapter_ScopeParity` to assert `agents/` subdir exists at global scope and that the set of agent files there matches project scope.
- Add a regression test: pre-seed `$HOME/.claude/builder.md` (legacy flat file), run Install, assert migration to `$HOME/.claude/agents/builder.md`.
- Add a preservation test: pre-seed `$HOME/.claude/CLAUDE.md` with known content, run Install, assert content unchanged.

## Notes

- This is the single most impactful fix in the spec — the research agent verified live that my own `~/.claude/` has flat agents because of this bug.
- Use `t.TempDir()` + `t.Setenv("HOME", ...)` for isolation.
