# Task 002: Create CLI Entry with Commander

**Phase:** 1
**User Story:** US-1
**Status:** TODO
**Depends on:** T001

---

## Objective

Wire up commander with all planned commands (init, add, remove, update, doctor, status). Only `init` will be implemented — others are stubs.

## Subtasks

- [ ] Create `src/cli.ts` with commander program definition
- [ ] Register commands: `init`, `add`, `remove`, `update`, `doctor`, `status`
- [ ] `init` command accepts flags: `--type`, `--tools`, `--name`, `--no-interactive`
- [ ] `add` command accepts `<tool>` argument
- [ ] `remove` command accepts `<tool>` argument
- [ ] Stub unimplemented commands with "coming soon" message
- [ ] Export main function from `src/index.ts`
- [ ] Verify: `node bin/ai-setup.js init --help` shows init options

## Files to Touch

| File | Action |
|------|--------|
| `src/cli.ts` | Create |
| `src/index.ts` | Create |
| `src/commands/init.ts` | Create (handler stub) |
| `src/commands/add.ts` | Create (stub) |
| `src/commands/update.ts` | Create (stub) |
| `src/commands/doctor.ts` | Create (stub) |
| `src/commands/status.ts` | Create (stub) |

## Done When

- [ ] `node bin/ai-setup.js --help` shows all commands
- [ ] `node bin/ai-setup.js init --help` shows init flags
- [ ] `node bin/ai-setup.js update` shows "coming soon"
