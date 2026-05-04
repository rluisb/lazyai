# Tasks: [###-feature-slug]

**Feature ID:** ###
**Spec:** [./spec.md](./spec.md)
**Plan:** [./plan.md](./plan.md)
**Date:** YYYY-MM-DD
**Status:** Draft | Approved | In Progress | Complete

> **Purpose.** This is the executable breakdown of `plan.md`. Tasks are grouped by phase, labeled by user story, and marked `[P]` when they can run in parallel. Each task has its own per-task harness file (`tasks/NNN-name.md` produced by `task-harness-template.md`) for implementation details.

---

## Conventions

- **Task ID:** `T###` — globally unique within this file.
- **Phase label:** `[Phase N]` — corresponds to phases in `plan.md`.
- **User story label:** `[US1]`, `[US2]`, `[US3]` — links the task back to a P1/P2/P3 story in `spec.md`.
- **Parallel marker:** `[P]` — task can run concurrently with siblings in the same phase.
- **Dependency arrow:** `← T###` — depends on the listed task(s).
- **Status:** ☐ TODO · 🔄 IN PROGRESS · ✅ DONE · ⛔ BLOCKED.

---

## Phase 1 — [Phase name from plan.md]

**Goal:** [one sentence — what this phase delivers].
**Exit criterion:** [demoable / mergeable state from plan.md].

### Setup
- [ ] **T001** ☐ [Phase 1] [foundation] Set up [scaffolding / migration / fixture]. **Depends on:** none.

### User Story 1 — [US1 title from spec]
- [ ] **T010** ☐ [Phase 1] [US1] [P] Write failing tests for FR-001, FR-002. **Depends on:** T001.
- [ ] **T011** ☐ [Phase 1] [US1] Implement smallest passing code for T010. **Depends on:** T010.
- [ ] **T012** ☐ [Phase 1] [US1] [P] Add edge-case tests for EC-001, EC-002. **Depends on:** T011.

### User Story 2 — [US2 title]
- [ ] **T020** ☐ [Phase 1] [US2] [P] [Task description]. **Depends on:** T011.
- [ ] **T021** ☐ [Phase 1] [US2] [Task description]. **Depends on:** T020.

### Phase 1 Verification
- [ ] **T099** ☐ [Phase 1] Run full test suite + 5-gate ladder for the slice. **Depends on:** all of Phase 1.

---

## Phase 2 — [Phase name]

**Goal:** [one sentence].
**Exit criterion:** [demoable state].

### User Story 3 — [US3 title]
- [ ] **T100** ☐ [Phase 2] [US3] [Task]. **Depends on:** T099.

*(Continue with phases 3+ as needed. Delete this section if only one phase.)*

---

## Dependency Graph

A visual cross-check that the dependency arrows are consistent.

```
T001 ─► T010 ─► T011 ─► T012
                ├─► T020 ─► T021
                └─► T099
```

> If a task appears in the dependency arrows column above but not in this graph, the graph is wrong. Reviewers MUST validate consistency.

---

## Parallelization Plan

Tasks safe to run in parallel within a phase, organized by Phase.

| Phase | Parallel set | Reason |
|---|---|---|
| 1 | { T010, T012 } | independent test files |
| 1 | { T020 } | independent module; no shared state |

> A task is `[P]`-eligible only when it touches files no other in-flight task touches AND its dependencies are complete.

---

## Per-Task Harnesses

Each task has its own implementation harness file. Generated from `task-harness-template.md`.

```
specs/###-slug/tasks/
├── T001-name.md
├── T010-name.md
├── T011-name.md
└── …
```

> A task's harness file is the *only* document the implementor agent reads at execution time. The contract is: harness in, code + tests out.

---

## Constitutional Notes

- **Article II — Test-First:** every implementation task is preceded by a test task that fails on the unimplemented code.
- **Article IV — YAGNI:** tasks not derived from a spec FR-* or AC are inadmissible. Reviewers MUST reject orphan tasks.
- **Article V — Simplicity:** prefer fewer, larger phases over many fine-grained ones. Phases that complete in <1 day are usually too small.

---

## Approvals

| Role | Name | Date | Verdict |
|---|---|---|---|
| Author | [name] | YYYY-MM-DD | [verdict] |
| Plan reviewer | [name] | YYYY-MM-DD | [verdict] |
| Human gate | [name] | YYYY-MM-DD | approved / requested changes |
