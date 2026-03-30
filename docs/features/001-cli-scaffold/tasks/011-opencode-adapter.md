# Task 011: Create OpenCode Adapter

**Phase:** 4
**User Story:** US-1
**Status:** TODO
**Depends on:** T009
**Parallel with:** T010

---

## Objective

Create the OpenCode adapter that copies agents and commands to OpenCode's expected directories with OpenCode-specific formatting.

## Subtasks

- [ ] Create `src/adapters/opencode.ts` implementing `ToolAdapter`
- [ ] `install()`:
  - Copy library/agents/ → `.opencode/agents/` (no format change)
  - Transform library skills → `.opencode/commands/` (add opencode frontmatter: description only)
- [ ] `remove()`:
  - Delete `.opencode/agents/`, `.opencode/commands/`
- [ ] Update .ai-setup.json with opencode-specific file hashes

## Files to Touch

| File | Action |
|------|--------|
| `src/adapters/opencode.ts` | Create |

## Done When

- [ ] After init with OpenCode selected: `.opencode/agents/` has 6 files, `.opencode/commands/` has 4 files
- [ ] Commands have correct OpenCode frontmatter (description only)
- [ ] Agents are exact copies
