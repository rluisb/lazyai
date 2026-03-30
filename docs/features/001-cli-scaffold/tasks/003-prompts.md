# Task 003: Create Interactive Prompts

**Phase:** 1
**User Story:** US-1
**Status:** TODO
**Depends on:** T002

---

## Objective

Create the interactive prompt flow using @clack/prompts. Must support both interactive and non-interactive (flag-based) modes.

## Subtasks

- [ ] Create `src/prompts.ts`
- [ ] Implement setup type prompt: Project vs Workspace
- [ ] Implement tool selection prompt: multi-select (Pi, OpenCode + future placeholders)
- [ ] Implement project name prompt (default: current directory name)
- [ ] Create `src/types.ts` with `SetupConfig` interface
- [ ] If flags provided (--type, --tools, --name) → skip corresponding prompts
- [ ] If `--no-interactive` → require all flags, error if missing
- [ ] Add intro banner with @clack/prompts intro()
- [ ] Add outro with next-steps message
- [ ] Verify: running init shows the full prompt flow
- [ ] Verify: running init with all flags skips prompts

## Files to Touch

| File | Action |
|------|--------|
| `src/prompts.ts` | Create |
| `src/types.ts` | Create |
| `src/commands/init.ts` | Modify (wire prompts) |

## Done When

- [ ] Interactive mode: shows setup type → tools → name prompts
- [ ] Non-interactive: `init --type project --tools pi,opencode --name test --no-interactive` skips all prompts
- [ ] Config object is correctly built from either path
