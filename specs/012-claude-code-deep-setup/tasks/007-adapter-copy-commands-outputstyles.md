# Task 007 — Adapter wiring: copy commands/ and output-styles/ at every scope

**Phase:** 2 (starter library assets)
**Estimated LOC:** ~60

## Goal

Walk `library/claudecode/commands/` and `library/claudecode/output-styles/` during `Install()` and copy into `<scope-root>/commands/` and `<scope-root>/output-styles/` at project, workspace, and global scopes. Create the directories even if empty (Claude's discovery tolerates empty dirs).

## Files to touch

| File | Change |
|---|---|
| `internal/adapter/claudecode.go` | Add two new copy steps after agents/skills. Reuse `CopyLibraryDirectory` or equivalent helper; scope-root comes from `ResolveToolRoot`. |
| `internal/adapter/adapter_scope_test.go` | Extend expected-files list per scope to include `commands/{review,test,commit}.md` and `output-styles/{terse,explanatory}.md`. |

## Acceptance criteria

- [ ] `<scope-root>/commands/review.md`, `test.md`, `commit.md` present at all three scopes
- [ ] `<scope-root>/output-styles/terse.md`, `explanatory.md` present at all three scopes
- [ ] Files are byte-identical to the library sources (no transformation)
- [ ] Empty-library case (no files under `library/claudecode/commands/`) still creates an empty `commands/` directory at the scope root

## Test plan

- Scope-parity test update (extend existing cases).
- Byte-equality test: hash library source vs installed file.
- Empty-library edge case: stub the embedded FS with zero commands, assert empty dir exists.

## Dependencies

- Task 005, 006 (assets must exist before the adapter has anything to copy).
