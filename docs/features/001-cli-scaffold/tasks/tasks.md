# Tasks: AI Setup CLI

**Feature:** 001-cli-scaffold
**Date:** 2026-03-28
**PRD:** [../prd.md](../prd.md)
**TechSpec:** [../techspec.md](../techspec.md)

---

## Dependency Graph

```
Phase 1 (Sequential):     T001 → T002 → T003

Phase 2 (Parallel):
         ┌→ T004 ─┐
T003 ────┤        ├───→ T007
         └→ T005 ─┘
              └→ T006 ─┘

Phase 3 (Sequential):     T007 → T008 → T009

Phase 4 (Parallel):
         ┌→ T010 ─┐
T009 ────┤        ├───→ T012
         └→ T011 ─┘

Phase 5 (Sequential):     T012 → T013 → T014 → T015
```

---

## Phase 1: Project Setup

**Goal:** Buildable TypeScript project with CLI entry point.

- [ ] T001 [US-1] Initialize package.json, tsconfig, tsup config — `package.json`, `tsconfig.json`, `tsup.config.ts`
- [ ] T002 [US-1] Create CLI entry with commander — `src/cli.ts`, `bin/ai-setup.js`
- [ ] T003 [US-1] Create interactive prompts module — `src/prompts.ts`

---

## Phase 2: Library Content ⭐ MVP core

**Goal:** All library files packaged and accessible.

- [ ] T004 [P] [US-1] Copy all agent definitions to library — `library/agents/*.md`
- [ ] T005 [P] [US-1] Copy all templates + rules + context files to library — `library/templates/*.md`, `library/rules/*.md`, `library/context/*.md`
- [ ] T006 [P] [US-1] Copy prompts + infra files to library — `library/prompts/*.md`, `library/infra/*`
- [ ] T007 [US-1] Create root AGENTS.md template with placeholders — `library/root/AGENTS.template.md`

---

## Phase 3: Scaffold Engine ⭐ MVP core

**Goal:** Core engine that creates docs/ structure and copies shared files.

- [ ] T008 [US-1] Create file utilities (copy, write, hash, exists check) — `src/utils/files.ts`
- [ ] T009 [US-1] Create scaffold engine (docs/ structure + shared files + config) — `src/scaffold.ts`

**MVP Checkpoint:** After this phase, `init` creates the docs/ structure and shared files. No tool-specific content yet.

---

## Phase 4: Tool Adapters ⭐ MVP core

**Goal:** Pi and OpenCode adapters that format and copy tool-specific files.

- [ ] T010 [P] [US-1] Create adapter interface + Pi adapter — `src/adapters/types.ts`, `src/adapters/pi.ts`
- [ ] T011 [P] [US-1] Create OpenCode adapter — `src/adapters/opencode.ts`
- [ ] T012 [US-1] Wire adapters into scaffold engine — `src/scaffold.ts`

**MVP Checkpoint:** After this phase, `init` creates the full setup for Pi and/or OpenCode. US-1 is complete.

---

## Phase 5: Additional Commands + Polish

**Goal:** add, update, doctor, status commands.

- [ ] T013 [US-3] Create `add` command (add tool to existing setup) — `src/commands/add.ts`
- [ ] T014 [US-4] Create `update` command (smart file refresh) — `src/commands/update.ts`
- [ ] T015 [US-5] Create `doctor` command (verify integrity) — `src/commands/doctor.ts`

---

## Phase 6: Testing + Docs

**Goal:** Test coverage and documentation.

- [ ] T016 [P] Unit tests for adapters and file utils — `tests/unit/`
- [ ] T017 [P] Integration test: full init → verify file tree — `tests/integration/`
- [ ] T018 README with usage, examples, and contribution guide — `README.md`
- [ ] T019 Create ADR-001 (TypeScript + @clack decision) — `docs/adrs/001-typescript-clack-cli.md`

---

## Execution Rules

1. **MVP first.** T001-T012 = working `init` command. Ship that before anything else.
2. **One task = one commit.**
3. **Tests in Phase 6** but each task should be manually verifiable.
4. **Non-interactive flags** added to prompts from the start (T003).
5. **Library files are copied from our Templates directory** — not rewritten.
