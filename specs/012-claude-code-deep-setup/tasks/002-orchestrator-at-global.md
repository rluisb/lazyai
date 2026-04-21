# Task 002 — Orchestrator agent at global scope

**Phase:** 1 (structural fixes, no CLI)
**Estimated LOC:** ~15

## Goal

Remove the `!isGlobal` gate that silently skips the orchestrator agent at global scope. Orchestrator ships whenever `IsOrchestratorEnabled(ctx)` is true, regardless of scope. Per research decision Q2 this was a bug, not a design choice.

## Files to touch

| File | Change |
|---|---|
| `internal/adapter/claudecode.go` | At ~L95, change `if !isGlobal && IsOrchestratorEnabled(ctx)` to `if IsOrchestratorEnabled(ctx)`. Ensure the destination path now resolves to `~/.claude/agents/orchestrator.md` at global scope (depends on Task 001 landing first). |

## Acceptance criteria

- [ ] Orchestrator agent file is emitted at global scope when `EnableServers` contains the orchestrator token
- [ ] Not emitted when `EnableServers` does not contain it (parity with project scope)
- [ ] Frontmatter of the global orchestrator matches the project one byte-for-byte (same tools, same description)

## Test plan

- Extend the scope-parity test from Task 001: add a case that enables the orchestrator and asserts `agents/orchestrator.md` exists at all three scopes.
- Add a negative case: orchestrator disabled → no file at any scope.

## Dependencies

- Task 001 (this change assumes agents live under `agents/` at global scope).
