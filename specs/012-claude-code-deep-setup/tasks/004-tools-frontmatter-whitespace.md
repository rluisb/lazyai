# Task 004 — Agent `tools` frontmatter: whitespace delimiter for Claude

**Phase:** 1 (structural fixes, no CLI)
**Estimated LOC:** ~25

## Goal

Claude Code expects the agent frontmatter `tools` field to be whitespace-separated (`tools: Bash Read Edit`). Today `NormalizeToolsFrontmatter("comma")` emits comma-separated form for Claude, which is not the documented contract. Change the delimiter to whitespace for Claude while keeping OpenCode's current behavior intact.

## Files to touch

| File | Change |
|---|---|
| `internal/adapter/shared.go` | `NormalizeToolsFrontmatter` at ~L608-661 — add a `"space"` (or `"whitespace"`) delimiter option. Don't remove the comma option (OpenCode depends on it). |
| `internal/adapter/claudecode.go` | Change the call site from `NormalizeToolsFrontmatter(source, "comma")` to `NormalizeToolsFrontmatter(source, "space")`. |
| `internal/adapter/shared_test.go` (or equivalent) | Add test cases for the new delimiter. |

## Acceptance criteria

- [ ] Every agent file written at every scope has `tools: Bash Read Edit` style (space-separated)
- [ ] OpenCode's agent emitter is unchanged (still comma-separated or whatever OpenCode's canonical form is — verify by reading OpenCode docs / existing test fixtures)
- [ ] Unit test covers both delimiter options

## Test plan

- Unit test on `NormalizeToolsFrontmatter`: given `tools: Bash, Read, Edit`, output with `"space"` = `tools: Bash Read Edit`, with `"comma"` = `tools: Bash, Read, Edit`.
- Integration: assert written orchestrator agent file contains whitespace-separated tools.

## Notes

- Verify OpenCode's canonical form before touching shared.go — I'm assuming comma-separated because that's what `"comma"` suggests, but confirm against OpenCode docs or its existing test fixtures. If OpenCode also wants whitespace, the fix is narrower.
