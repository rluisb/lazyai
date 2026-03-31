# Task 001: Initialize Project

**Phase:** 1
**User Story:** US-1
**Status:** TODO
**Depends on:** None

---

## Objective

Create a buildable TypeScript project with the right package structure for an npx-publishable CLI.

## Subtasks

- [ ] Create `package.json` with name `@ricardoborges-teachable/ai-setup`, bin entry, type module
- [ ] Create `tsconfig.json` with strict mode, ESM output
- [ ] Create `tsup.config.ts` for building to `dist/`
- [ ] Create `bin/ai-setup.js` shebang entry point (`#!/usr/bin/env node`)
- [ ] Create `.gitignore` (node_modules, dist, .ai-setup.json)
- [ ] Install dev deps: `typescript`, `tsup`, `@types/node`
- [ ] Install prod deps: `commander`, `@clack/prompts`
- [ ] Verify: `pnpm build` succeeds
- [ ] Verify: `node bin/ai-setup.js --version` outputs version

## Files to Touch

| File | Action |
|------|--------|
| `package.json` | Create |
| `tsconfig.json` | Create |
| `tsup.config.ts` | Create |
| `bin/ai-setup.js` | Create |
| `.gitignore` | Create |

## Done When

- [ ] `pnpm build` produces `dist/cli.js`
- [ ] `node bin/ai-setup.js --version` prints version
- [ ] `node bin/ai-setup.js --help` shows help text
