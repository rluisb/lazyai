# Task 010: Create Pi Adapter

**Phase:** 4
**User Story:** US-1
**Status:** TODO
**Depends on:** T009
**Parallel with:** T011

---

## Objective

Create the Pi adapter that copies agents, skills, and prompts to Pi's expected directories with Pi-specific formatting.

## Subtasks

- [ ] Create `src/adapters/types.ts` — `ToolAdapter` interface: `install(targetDir, libraryDir)`, `remove(targetDir)`, `getToolId(): string`
- [ ] Create `src/adapters/pi.ts` implementing `ToolAdapter`
- [ ] `install()`:
  - Copy library/agents/ → `.pi/agents/` (no format change)
  - Copy library/prompts/ → `.pi/templates/` (no format change)
  - Transform library skills → `.pi/skills/` (add pi frontmatter: name, description, usage)
- [ ] `remove()`:
  - Delete `.pi/agents/`, `.pi/skills/`, `.pi/templates/`
- [ ] Update .ai-setup.json with pi-specific file hashes

## Files to Touch

| File | Action |
|------|--------|
| `src/adapters/types.ts` | Create |
| `src/adapters/pi.ts` | Create |

## Done When

- [ ] After init with Pi selected: `.pi/agents/` has 6 files, `.pi/skills/` has 4 files, `.pi/templates/` has 5 files
- [ ] Skills have correct Pi frontmatter (name, description, usage)
- [ ] Agents are exact copies (no transformation)
