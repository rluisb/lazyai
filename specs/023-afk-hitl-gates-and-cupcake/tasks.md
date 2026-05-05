# Tasks: 023-afk-hitl-gates-and-cupcake

**Feature ID:** 023
**Spec:** GitHub issue #169 + approved planning prompt scope (no local `spec.md` present in this worktree)
**Plan:** [./plan.md](./plan.md)
**Date:** 2026-05-05
**Status:** Draft

> **Purpose.** This is the executable breakdown of `plan.md`. Tasks are grouped by phase, labeled by user story, and marked `[P]` when they can run in parallel. Each task has its own per-task harness pointer (`tasks/NNN-name.md` generated from `task-harness-template.md`) for implementation details.

---

## Conventions

- **Task ID:** `T###` — globally unique within this file.
- **Phase label:** `[Phase N]` — corresponds to phases in `plan.md`.
- **User story label:** `[US1]`, `[US2]`, `[US3]` — links the task back to the three bounded slices in the request.
- **Parallel marker:** `[P]` — task can run concurrently with siblings in the same phase after dependencies are complete and file touch-maps do not overlap.
- **AFK/HITL marker:** `[AFK]` means Cupcake/prompt policy allows autonomous execution after dependencies and required approvals; `[HITL]` means human approval or interpretation is required before proceeding.
- **Dependency arrow:** `← T###` — depends on the listed task(s).
- **Status:** ☐ TODO · 🔄 IN PROGRESS · ✅ DONE · ⛔ BLOCKED.
- **Bounded task budget:** each task is capped at ≤100 LOC changed and ≤4 new/changed test cases. If a task exceeds either bound, stop and split it before continuing.

### AFK/HITL classification rules for this task set

- `[AFK]` applies to read-only work, writing `specs/`, running existing tests, adding focused content tests, and editing approved prompt-library files after plan approval.
- `[HITL]` applies to phase transition approval, terminology decisions, merge readiness interpretation, destructive/config/dependency actions, or any action Cupcake/pre-commit would block without a valid human gate.
- AFK/HITL markers do **not** replace Cupcake, pre-commit, CI, or tool-native permission checks.

---

## User Stories

| Story | Slice | Outcome |
|---|---|---|
| [US1] | AFK/HITL markers in speckit-tasks and orchestrate | Task ledgers and wave dispatch guidance classify work using existing Cupcake signals. |
| [US2] | Bugfix feedback-loop gate | Bugfix workflow requires an automated pass/fail signal before hypothesis/RCA/fix planning. |
| [US3] | Terminology enforcement | Clarify and orchestrate workflows push/align domain vocabulary through `KNOWLEDGE_MAP.md`. |

---

## Phase 0 — Implementation Gate

**Goal:** Ensure implementation does not start until the plan is human-approved by the repository gate process.
**Exit criterion:** Human approval exists outside AI-generated text; implementation tasks may be dispatched.

### Task Ledger

| Task | Status | Phase | Story | Markers | Title | Files | Depends on | Done When / Evidence | Harness |
|---|---|---|---|---|---|---|---|---|---|
| **T001** | ☐ | [Phase 0] | [US1][US2][US3] | [HITL] | Obtain implementation gate approval | `specs/023-afk-hitl-gates-and-cupcake/plan.md` | none | Human approval is recorded by the normal repo process before any non-spec file edits begin. Evidence: human-authored gate/approval visible to implementor. | `tasks/T001-implementation-gate.md` |

---

## Phase 1 — AFK/HITL Classification

**Goal:** Add AFK/HITL task markers and dispatch/wave gating guidance without changing enforcement code.
**Exit criterion:** Content tests pass and `speckit-tasks`, `orchestrate`, and `parallel-execution` document prompt-level AFK/HITL behavior.

### Task Ledger

| Task | Status | Phase | Story | Markers | Title | Files | Depends on | Done When / Evidence | Harness |
|---|---|---|---|---|---|---|---|---|---|
| **T002** | ☐ | [Phase 1] | [US1] | [AFK] [P] | Write AFK/HITL content contract tests | `packages/cli/internal/library/afk_hitl_content_test.go` | T001 | ≤4 tests fail on current content and assert required AFK/HITL markers, Cupcake signal mapping, wave gating, and no new enforcement claims. Evidence: targeted `go test` fails before content edits. | `tasks/T002-afk-hitl-content-tests.md` |
| **T003** | ☐ | [Phase 1] | [US1] | [AFK] | Add AFK/HITL markers to speckit task ledger guidance | `packages/cli/library/skills/speckit-tasks.md` | T002 | Skill explains `[AFK]`/`[HITL]` markers, includes them in task row format, and keeps `[P]` semantics separate. Evidence: T002 relevant assertions pass. | `tasks/T003-speckit-tasks-markers.md` |
| **T004** | ☐ | [Phase 1] | [US1] | [AFK] [P] | Add AFK/HITL wave gating to parallel execution | `packages/cli/library/skills/parallel-execution.md` | T002 | Wave model states AFK tasks may dispatch only when dependency/file-safe and HITL tasks pause/split the wave. Evidence: T002 relevant assertions pass. | `tasks/T004-parallel-wave-gating.md` |
| **T005** | ☐ | [Phase 1] | [US1] | [AFK] | Add Cupcake signal checks before orchestrator dispatch | `packages/cli/library/skills/orchestrate.md` | T002 | Orchestrator guidance maps plan/gate signals, hard blocks, read-only work, specs writes, dependency/config/destructive actions, and bugfix interpretation to AFK/HITL. Evidence: T002 relevant assertions pass. | `tasks/T005-orchestrate-afk-hitl-dispatch.md` |
| **T006** | ☐ | [Phase 1] | [US1] | [AFK] | Verify AFK/HITL slice | `packages/cli/internal/library/afk_hitl_content_test.go` | T003, T004, T005 | `cd packages/cli && go test ./internal/library -run 'AFK|HITL|Wave|Dispatch'` passes; scope audit confirms no Cupcake/Rego/hook files changed. | `tasks/T006-verify-afk-hitl.md` |

---

## Phase 2 — Bugfix Feedback-Loop Gate

**Goal:** Require an automated pass/fail signal before bugfix hypothesis and fix planning.
**Exit criterion:** Content tests pass and `bugfix.md` plus the RCA template include feedback-loop guidance.

### Task Ledger

| Task | Status | Phase | Story | Markers | Title | Files | Depends on | Done When / Evidence | Harness |
|---|---|---|---|---|---|---|---|---|---|
| **T007** | ☐ | [Phase 2] | [US2] | [AFK] [P] | Write bugfix feedback-loop content tests | `packages/cli/internal/library/bugfix_feedback_loop_content_test.go` | T006 | ≤4 tests fail on current content and assert Step 0, automated pass/fail signal, block-before-hypothesis wording, RCA template section, and no external evaluation infrastructure. Evidence: targeted `go test` fails before content edits. | `tasks/T007-bugfix-feedback-tests.md` |
| **T008** | ☐ | [Phase 2] | [US2] | [AFK] | Add Step 0 feedback loop to bugfix skill | `packages/cli/library/skills/bugfix.md` | T007 | `bugfix.md` includes Step 0 before current Step 1/pre-flight flow and requires constructing an objective pass/fail signal before hypothesizing. Evidence: T007 relevant assertions pass. | `tasks/T008-bugfix-step-zero.md` |
| **T009** | ☐ | [Phase 2] | [US2] | [AFK] | Add feedback-loop section to bugfix RCA template | `packages/cli/library/templates/bugfix-rca-template.md` | T007 | RCA template records signal command/source, expected fail/pass behavior, current result, and escalation if no signal can be built. Evidence: T007 relevant assertions pass. | `tasks/T009-rca-feedback-loop-template.md` |
| **T010** | ☐ | [Phase 2] | [US2] | [AFK] | Verify bugfix feedback-loop slice | `packages/cli/internal/library/bugfix_feedback_loop_content_test.go` | T008, T009 | `cd packages/cli && go test ./internal/library -run 'Bugfix.*Feedback|FeedbackLoop'` passes; scope audit confirms no bugfix runtime engine or new skill was added. | `tasks/T010-verify-bugfix-feedback.md` |

---

## Phase 3 — Terminology Enforcement

**Goal:** Add lightweight vocabulary alignment guidance to clarify, knowledge map, and orchestrator dispatch.
**Exit criterion:** Content tests pass and terminology alignment is documented without new top-level skills or glossary infrastructure.

### Task Ledger

| Task | Status | Phase | Story | Markers | Title | Files | Depends on | Done When / Evidence | Harness |
|---|---|---|---|---|---|---|---|---|---|
| **T011** | ☐ | [Phase 3] | [US3] | [HITL] | Confirm terminology decision boundary | `specs/023-afk-hitl-gates-and-cupcake/tasks.md` | T010 | Human confirms the lightweight vocabulary terms/source of truth or accepts the plan's `KNOWLEDGE_MAP.md` approach before terminology edits. Evidence: human-authored response or review note. | `tasks/T011-terminology-decision-boundary.md` |
| **T012** | ☐ | [Phase 3] | [US3] | [AFK] [P] | Write terminology content contract tests | `packages/cli/internal/library/terminology_content_test.go` | T011 | ≤4 tests fail on current content and assert clarify vocabulary push, `KNOWLEDGE_MAP.md` terminology section, orchestrator vocabulary alignment, and no new glossary system/skill. Evidence: targeted `go test` fails before content edits. | `tasks/T012-terminology-content-tests.md` |
| **T013** | ☐ | [Phase 3] | [US3] | [AFK] [P] | Add terminology guidance to speckit clarify | `packages/cli/library/skills/speckit-clarify.md` | T012 | Clarify guidance tells agents to surface new domain vocabulary, ask if terms are ambiguous, and record accepted vocabulary durably without adding scope. Evidence: T012 relevant assertions pass. | `tasks/T013-speckit-clarify-terminology.md` |
| **T014** | ☐ | [Phase 3] | [US3] | [AFK] [P] | Add terminology section concept to knowledge map | `KNOWLEDGE_MAP.md` | T012 | `KNOWLEDGE_MAP.md` includes a lightweight terminology/vocabulary section concept and points agents to accepted domain terms. Evidence: T012 relevant assertions pass. | `tasks/T014-knowledge-map-terminology.md` |
| **T015** | ☐ | [Phase 3] | [US3] | [AFK] [P] | Add vocabulary alignment check before orchestrator dispatch | `packages/cli/library/skills/orchestrate.md` | T005, T012 | Orchestrator guidance checks vocabulary alignment before dispatch and pauses for HITL when terms are undecided. Evidence: T012 relevant assertions pass. | `tasks/T015-orchestrate-vocabulary-alignment.md` |
| **T016** | ☐ | [Phase 3] | [US3] | [AFK] | Verify terminology slice | `packages/cli/internal/library/terminology_content_test.go` | T013, T014, T015 | `cd packages/cli && go test ./internal/library -run 'Terminology|Vocabulary'` passes; scope audit confirms no new top-level skill or glossary subsystem was added. | `tasks/T016-verify-terminology.md` |

---

## Phase 4 — Final Verification and Feedback

**Goal:** Prove all slices satisfy the plan and stop before merge-level human review.
**Exit criterion:** Library content tests pass and forbidden scope expansion is absent.

### Task Ledger

| Task | Status | Phase | Story | Markers | Title | Files | Depends on | Done When / Evidence | Harness |
|---|---|---|---|---|---|---|---|---|---|
| **T017** | ☐ | [Phase 4] | [US1][US2][US3] | [AFK] | Run full library verification and scope audit | `packages/cli/internal/library/*_test.go`, touched Markdown files | T006, T010, T016 | `cd packages/cli && go test ./internal/library` passes; grep/audit evidence shows no changes to `policies/`, `.husky/`, `packages/cli/library/hooks/`, dependency files, or excluded skills. | `tasks/T017-final-verification-scope-audit.md` |
| **T018** | ☐ | [Phase 4] | [US1][US2][US3] | [HITL] | Review verification results before merge | Implementation summary / PR notes | T017 | Human reviews test output and scope audit, then decides whether to merge/request changes. Evidence: human-authored review decision. | `tasks/T018-human-feedback-before-merge.md` |

---

## Dependency Graph

A visual cross-check that the dependency arrows are consistent.

```
T001
 ├─► T002 ─┬─► T003 ─┐
 │         ├─► T004 ─┼─► T006 ─► T007 ─┬─► T008 ─┐
 │         └─► T005 ─┘                  └─► T009 ─┴─► T010
 │
 └──────────────────────────────────────────────────────┐
                                                        ▼
T010 ─► T011 ─► T012 ─┬─► T013 ─┐
                      ├─► T014 ─┼─► T016 ─► T017 ─► T018
                      └─► T015 ─┘
T005 ──────────────────────▲
```

> If a task appears in the dependency arrows column above but not in this graph, the graph is wrong. Reviewers MUST validate consistency.

---

## Parallelization Plan

Tasks safe to run in parallel within a phase, organized by Phase.

| Phase | Parallel set | Reason |
|---|---|---|
| 1 | { T003, T004 } after T002 | Different Markdown files (`speckit-tasks.md`, `parallel-execution.md`) and same test contract. |
| 3 | { T013, T014, T015 } after T012 and T005 | Different target files for T013/T014/T015; T015 waits for earlier `orchestrate.md` AFK/HITL edit to avoid same-file conflict. |

> A task is `[P]`-eligible only when it touches files no other in-flight task touches AND its dependencies are complete. `[HITL]` tasks pause dispatch until human input is available.

---

## File Touch-Map

| Task | Expected writes | Conflict notes |
|---|---|---|
| T002 | `packages/cli/internal/library/afk_hitl_content_test.go` | Independent test file. |
| T003 | `packages/cli/library/skills/speckit-tasks.md` | No other task touches this file. |
| T004 | `packages/cli/library/skills/parallel-execution.md` | No other task touches this file. |
| T005 | `packages/cli/library/skills/orchestrate.md` | Must complete before T015. |
| T007 | `packages/cli/internal/library/bugfix_feedback_loop_content_test.go` | Independent test file. |
| T008 | `packages/cli/library/skills/bugfix.md` | No other task touches this file. |
| T009 | `packages/cli/library/templates/bugfix-rca-template.md` | No other task touches this file. |
| T012 | `packages/cli/internal/library/terminology_content_test.go` | Independent test file. |
| T013 | `packages/cli/library/skills/speckit-clarify.md` | No other task touches this file. |
| T014 | `KNOWLEDGE_MAP.md` | No other task touches this file. |
| T015 | `packages/cli/library/skills/orchestrate.md` | Same file as T005; sequenced after T005. |

---

## Per-Task Harnesses

Each task has a harness pointer to be generated from `packages/cli/library/templates/task-harness-template.md` before implementation. The harness must pre-fill objective, excerpts from this plan/tasks file, quality gates, permission scope, and the exact files in play.

```
specs/023-afk-hitl-gates-and-cupcake/tasks/
├── T001-implementation-gate.md
├── T002-afk-hitl-content-tests.md
├── T003-speckit-tasks-markers.md
├── T004-parallel-wave-gating.md
├── T005-orchestrate-afk-hitl-dispatch.md
├── T006-verify-afk-hitl.md
├── T007-bugfix-feedback-tests.md
├── T008-bugfix-step-zero.md
├── T009-rca-feedback-loop-template.md
├── T010-verify-bugfix-feedback.md
├── T011-terminology-decision-boundary.md
├── T012-terminology-content-tests.md
├── T013-speckit-clarify-terminology.md
├── T014-knowledge-map-terminology.md
├── T015-orchestrate-vocabulary-alignment.md
├── T016-verify-terminology.md
├── T017-final-verification-scope-audit.md
└── T018-human-feedback-before-merge.md
```

> A task's harness file is the *only* document the implementor agent reads at execution time. If harness generation is skipped, implementation must not proceed past T001.

---

## Quality Gates

- **Static/content integrity:** Markdown remains readable; Go test files are `gofmt`-formatted.
- **Contract compliance:** AFK/HITL stays prompt-level; no Rego/pre-commit/dependency/config changes.
- **Behavioral validation:** New content tests fail first, then pass after corresponding docs edits.
- **Pattern consistency:** Tests follow `packages/cli/internal/library/*_content_test.go` patterns using `readLibraryFile` and `assertContainsAll`.
- **Scope audit:** Excluded skills (`workspace-configure`, `architectural-deepening`, `diagnose`, `grill-me`) are not added or modified.

---

## Constitutional Notes

- **Article II — Test-First:** T002, T007, and T012 are RED-first tasks and gate their respective content edits.
- **Article IV — YAGNI:** tasks not derived from the three requested slices are inadmissible; reviewers must reject Rego, hook, runtime, dependency, or new-skill work.
- **Article V — Simplicity:** prose guidance plus content tests is the chosen simple design; no classifier or glossary engine.
- **Article VI — Anti-Overengineering:** each task is bounded to ≤100 LOC and ≤4 test cases; no abstractions or helper extraction unless already present.

---

## Approvals

| Role | Name | Date | Verdict |
|---|---|---|---|
| Author | Planner Agent | 2026-05-05 | Draft complete |
| Plan reviewer | TBD | TBD | pending |
| Human gate | TBD human | TBD | pending |
