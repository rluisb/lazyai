# Task 012: Wire Adapters into Scaffold Engine

**Phase:** 4
**User Story:** US-1
**Status:** TODO
**Depends on:** T010, T011

---

## Objective

Connect the adapter registry to the scaffold engine so `init` runs the correct adapters based on user's tool selection.

## Subtasks

- [ ] Create `src/adapters/registry.ts` — map of tool ID → adapter instance
- [ ] Modify `src/scaffold.ts` to accept tool list and call each adapter's `install()`
- [ ] Modify `src/commands/init.ts` to pass selected tools to scaffold
- [ ] Display per-tool success messages during init
- [ ] Verify: init with Pi only → only .pi/ created
- [ ] Verify: init with OpenCode only → only .opencode/ created
- [ ] Verify: init with both → both directories created

## Files to Touch

| File | Action |
|------|--------|
| `src/adapters/registry.ts` | Create |
| `src/scaffold.ts` | Modify |
| `src/commands/init.ts` | Modify |

## Done When

- [ ] `init --tools pi` → .pi/ exists, .opencode/ doesn't
- [ ] `init --tools opencode` → .opencode/ exists, .pi/ doesn't
- [ ] `init --tools pi,opencode` → both exist
- [ ] .ai-setup.json lists selected tools
- [ ] US-1 acceptance criteria all pass
