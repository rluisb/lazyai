<template>
  <name>Tasks — Implementation Task List</name>
  <output>specs/features/NNN-feature-name/tasks/tasks.md</output>
  <input>PRD + TechSpec</input>
  <phase>Plan — Step 3</phase>
</template>

# Tasks: [Feature Name]

**Feature:** NNN-feature-name
**Date:** YYYY-MM-DD
**PRD:** [link to prd.md]
**TechSpec:** [link to techspec.md]

---

## Task Format

```
- [ ] TNNN [P] [US-N] Description — path/to/file
```

- `TNNN` — task number (execution order)
- `[P]` — can run in parallel (different files, no dependency on other [P] tasks)
- `[US-N]` — user story this task serves (traceability)
- `path/` — exact file(s) this task touches

---

## Dependency Graph

<!-- ASCII visualization of phase dependencies and parallel opportunities.
     Update this as tasks are added or reordered. -->

```
Phase 1 (Sequential):     T001 → T002 → T003

Phase 2 (Parallel):
         ┌→ T004 ─┐
T003 ────┤        ├───→ T007
         └→ T005 ─┘
              └→ T006 ─┘

Phase 3 (Sequential):     T007 → T008 → T009
```

---

## Phase 1: Setup

**Goal:** Foundation that everything else depends on.

- [ ] T001 [US-1] [description] — [path]
- [ ] T002 [P] [US-1] [description] — [path]
- [ ] T003 [P] [US-1] [description] — [path]

---

## Phase 2: [User Story 1 Title] (P1) ⭐ MVP

**Goal:** [what this delivers — independently testable]
**Standards:** [specs/standards files to follow]

- [ ] T004 [US-1] [description] — [path]
- [ ] T005 [P] [US-1] [description] — [path]
- [ ] T006 [US-1] [description] — [path]
- [ ] T007 [US-1] Tests for US-1 — [test path]

**MVP Checkpoint:** After this phase, US-1 works end-to-end. Stop and validate.

---

## Phase 3: [User Story 2 Title] (P2)

**Goal:** [what this delivers]
**Standards:** [specs/standards files to follow]

- [ ] T008 [US-2] [description] — [path]
- [ ] T009 [P] [US-2] [description] — [path]
- [ ] T010 [US-2] Tests for US-2 — [test path]

**Checkpoint:** US-1 + US-2 both work. Validate before continuing.

---

## Phase N: Polish

**Goal:** Cross-cutting. Only after all target user stories pass.

- [ ] TNNN [P] Documentation updates — specs/
- [ ] TNNN Code cleanup — [paths]
- [ ] TNNN Run full test suite and fix regressions

---

## Execution Rules

1. **MVP first.** Ship P1 before starting P2. A working P1 > a broken P1+P2.
2. **One task = one commit.** Atomic, reviewable, revertable.
3. **Tests before moving on.** Each phase ends with passing tests.
4. **[P] tasks can run in parallel** by different agents or in separate sessions.
5. **Sequential tasks MUST wait.** Don't start T005 if T004 isn't done.
6. **If blocked:** mark `[BLOCKED: reason]`, move to next unblocked task.
7. **Follow project standards.** Every task references specs/standards/ patterns.
8. **Update progress.md** after completing each task.

---

<!-- PRINCIPLES CHECK — verify before approving:
- [ ] Every task maps to a user story (traceability)
- [ ] P1 tasks come before P2, P2 before P3
- [ ] MVP checkpoint exists after P1 phase
- [ ] Parallel tasks correctly marked [P]
- [ ] No task touches more than 3 files (split if bigger)
- [ ] Dependency graph matches task ordering
- [ ] Standards references included per phase
- [ ] No speculative tasks (YAGNI — only what stories need)
- [ ] Each task is small enough for one session (do one thing well)
-->
