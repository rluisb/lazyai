# Task 012 — Scope-parity regression tests for global agents + orchestrator

**Phase:** 4 (verification)
**Estimated LOC:** ~80

## Goal

Lock in the fixes from Tasks 001 and 002 with explicit regression tests. Future refactors that re-flatten global agents or re-gate the orchestrator must fail loudly.

## Files to touch

| File | Change |
|---|---|
| `internal/adapter/adapter_scope_test.go` | Add `TestClaudeCode_GlobalAgentsInSubdir` — asserts no files exist at `~/.claude/<agent>.md` (flat), asserts every expected agent exists at `~/.claude/agents/<agent>.md`. Add `TestClaudeCode_OrchestratorAtAllScopes` — with orchestrator enabled, assert file exists at all three scopes; with it disabled, assert absence. |

## Acceptance criteria

- [ ] Both new tests pass against the post-Task-001/002 code
- [ ] Both tests FAIL if the pre-fix code is reintroduced (verify by reverting once)
- [ ] Test names clearly describe the defect they guard against (scan-readable)

## Test plan

- Run the new tests against current main (should FAIL — proves they're real regressions)
- Run against the Task 001/002 branch (should PASS)

## Dependencies

- Tasks 001, 002.
