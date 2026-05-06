# Plan: 023-afk-hitl-gates-and-cupcake

**Feature ID:** 023
**Spec:** GitHub issue #169 + approved planning prompt scope (no local `spec.md` present in this worktree)
**Date:** 2026-05-05
**Status:** Draft
**Owner:** Planner Agent
**Constitution:** Repository workflow constitution Articles I–VI as encoded by `packages/cli/library/templates/plan-template.md` and `packages/cli/library/fragments/rpi-workflow.md`

> **Purpose.** A plan describes *how* the system in `spec.md` will be built. It selects the tech stack, names the major modules, evaluates the design against the constitution, and breaks the work into tasks. Acceptance criteria stay in the spec; behavior contracts live here only when they are implementation details (data model, internal APIs).

---

## Summary

Implement three bounded prompt-library adaptations from `mattpocock/skills`: AFK/HITL task classification, a bugfix feedback-loop gate, and lightweight terminology enforcement. The approach is documentation-and-content-contract only: update existing skill/template files and add focused embedded-library content tests, without changing Cupcake Rego, pre-commit hooks, runtime orchestration state, or adding new skills. The work satisfies three user stories: classify work before dispatch, force bugfixes to build an objective pass/fail loop before hypothesizing, and keep domain vocabulary aligned before agents fan out.

---

## Technical Context

| Aspect | Decision | Rationale |
|---|---|---|
| Language(s) | Go 1.26.1 for tests; Markdown for library skill/template content | `go.work` and `packages/cli/go.mod` declare Go 1.26.1; target files are Markdown library artifacts. |
| Framework(s) | Standard Go `testing`; existing embedded-library FS helpers | Existing `packages/cli/internal/library/*_content_test.go` tests validate bundled Markdown content with simple string assertions. |
| Storage | None | No persisted application data or schema changes. |
| Deployment | Bundled CLI library content | `packages/cli/library/**` is shipped through existing library embedding/scaffold mechanisms. |
| Telemetry | None | Prompt guidance only; no runtime instrumentation is in scope. |

**External dependencies (new):**
- None.

**External dependencies (rejected):**
- `mattpocock/skills` as a runtime or vendored dependency — rejected because this spec adapts small prompt patterns only.
- New terminology/glossary package — rejected because `KNOWLEDGE_MAP.md` and existing skill files are sufficient.
- Cupcake/Rego policy extensions — rejected because AFK/HITL is prompt-level awareness, not a new enforcement layer.

---

## Constitution Check

| Article | Verdict | Justification |
|---|---|---|
| I — Library-First | PASS | Reuses existing Markdown skill system, existing templates, and existing Go content-test pattern; no custom parser, classifier, or runtime gate is introduced. |
| II — Test-First (NON-NEGOTIABLE) | PASS | Each slice begins with a focused content-contract test task that must fail before the corresponding Markdown guidance is added; final verification runs `go test ./internal/library` from `packages/cli`. |
| III — Docs as Source of Truth | PASS | The implementation changes the source-of-truth prompt library and `KNOWLEDGE_MAP.md`; plan/tasks map directly to the requested slices. |
| IV — Anti-Speculation (YAGNI) | PASS | Deliberately not adding workspace-configure, architectural-deepening, diagnose, grill-me, Rego changes, pre-commit changes, new top-level skills, runtime state, or new enforcement semantics. |
| V — Simplicity Over Abstraction | PASS | Simple prose guidance plus content tests is chosen over a generic AFK/HITL classifier, glossary engine, or orchestration state machine. |
| VI — Anti-Overengineering (NON-NEGOTIABLE) | PASS | No new abstractions, no dependencies, no DRY extraction below three uses; repeated phrase checks stay local and explicit in small tests. |

**Verdict:** APPROVED FOR PLANNING / PENDING HUMAN IMPLEMENTATION GATE — plan is bounded and constitution-compliant, but implementation must not begin until human approval is recorded by the repository's normal gate process.

---

## Project Structure

Where the new code goes. Use the actual paths in this repository.

```
[repo-root]/
├── packages/cli/internal/library/
│   ├── afk_hitl_content_test.go              ← new test-only contract
│   ├── bugfix_feedback_loop_content_test.go  ← new test-only contract
│   └── terminology_content_test.go           ← new test-only contract
├── packages/cli/library/skills/
│   ├── speckit-tasks.md                      ← modified
│   ├── orchestrate.md                        ← modified
│   ├── parallel-execution.md                 ← modified
│   ├── bugfix.md                             ← modified
│   └── speckit-clarify.md                    ← modified
├── packages/cli/library/templates/
│   └── bugfix-rca-template.md                ← modified
├── KNOWLEDGE_MAP.md                          ← modified
└── specs/023-afk-hitl-gates-and-cupcake/
    ├── plan.md                               ← this file
    └── tasks.md                              ← generated by speckit-tasks
```

---

## Data Model

No data changes.

**Migrations:**
- None.

**Backfill / data movement:**
- None required.

---

## Internal Contracts

Public APIs, internal interfaces, events, or queues introduced or changed by this plan.

| Contract | Producer | Consumer | Shape |
|---|---|---|---|
| Task ledger AFK/HITL markers | `speckit-tasks.md` | planners, decomposers, implementors, reviewers | Every task ledger entry includes `[AFK]` or `[HITL]` in addition to optional `[P]`; marker is based on Cupcake signals and human-gated categories. |
| Orchestrator dispatch classification | `orchestrate.md` | orchestrator skill users | Before wave/step dispatch, check plan/gate signals, hard blocks, dependency/config/destructive actions, and classify as AFK or HITL; prompt-level only. |
| Parallel wave gating | `parallel-execution.md` | parallel execution planners/builders | AFK tasks may share a wave if file/dependency-safe; HITL tasks pause the wave or split into a human-gated boundary; no shared-file parallelism. |
| Bugfix feedback-loop gate | `bugfix.md` + `bugfix-rca-template.md` | bugfix executors/reviewers | Step 0 records an automated pass/fail signal before hypothesis/RCA; inability to build the signal blocks or escalates. |
| Terminology alignment | `speckit-clarify.md`, `orchestrate.md`, `KNOWLEDGE_MAP.md` | spec clarifiers, orchestrators, downstream agents | New domain vocabulary is pushed into durable docs; orchestrator checks vocabulary alignment before dispatch and pauses on terminology decisions. |

---

## Decision Protocol

| Decision | Option A | Option B | Chosen | Rationale |
|---|---|---|---|---|
| AFK/HITL enforcement layer | Prompt-level awareness in skills | Modify Cupcake Rego/pre-commit enforcement | A | Matches explicit constraint; existing enforcement remains source of physical blocks. |
| Feedback loop placement | Add Step 0 to existing `bugfix.md` and RCA template | Create a new bugfix-loop skill | A | Smallest change; preserves existing bugfix workflow and avoids top-level skill expansion. |
| Terminology home | Add a lightweight section to `KNOWLEDGE_MAP.md` | Create a new glossary system/file tree | A | Satisfies durable vocabulary guidance without new architecture. |
| Tests | Focused string/content contract tests | Snapshot entire Markdown files | A | Content assertions are precise, low-maintenance, and follow existing library test style. |

---

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation | Owner |
|---|---|---|---|---|
| AFK/HITL wording implies new runtime enforcement | M | H | State repeatedly that classification is prompt-level and Cupcake remains the enforcement source; add negative test for forbidden runtime-autonomy claims. | Implementor |
| Task markers drift from Cupcake signals | M | M | Encode only the provided mapping: plan-attested, gate-attested, hard blocks, read-only/spec writes, dependency/config/destructive categories. | Implementor |
| Bugfix feedback loop becomes heavy process | M | M | Keep Step 0 to one automated pass/fail signal and explicit escalation when unavailable; no external infrastructure. | Implementor |
| Terminology enforcement becomes a glossary project | M | M | Use `KNOWLEDGE_MAP.md` concept section and clarify/orchestrate guidance only; no new files unless a human approves scope change. | Implementor |
| Content tests are brittle to wording | M | L | Assert small required phrases/sections, not whole-file snapshots. | Implementor |
| Same file (`orchestrate.md`) touched by two slices causes merge conflict | M | M | Sequence orchestrate tasks: AFK/HITL dispatch update first, terminology alignment second. | Implementor |

---

## Complexity Tracking

A deliberate ledger of every place this plan accepts complexity above the simplest workable approach. If this section is empty, the simplest approach was chosen.

| Item | Simpler alternative | Why complexity is justified | Cost |
|---|---|---|---|
| Three small content-test files | Manual review only | Article II requires test-first verification; separate files keep each slice ≤4 tests and avoid shared edit conflicts. | Low maintenance; phrase assertions must be updated if headings change. |
| Explicit HITL task for terminology decision | Let implementor choose vocabulary phrasing | User identified terminology decisions as HITL; prevents silent domain-language drift. | One human checkpoint before terminology edits. |
| Two sequential `orchestrate.md` tasks | One combined orchestrate edit | Separating dispatch classification from vocabulary alignment keeps each task bounded and reviewable while preserving file conflict ordering. | Slightly more task bookkeeping. |

---

## Phases & Milestones

A plan is broken into phases. Each phase ends in a demoable, mergeable state.

| Phase | Goal | Exit criterion |
|---|---|---|
| 0 — Implementation gate | Confirm human approval before touching library/test files. | Human gate recorded outside AI-generated text; implementation can start. |
| 1 — AFK/HITL classification | Add task-ledger markers and dispatch/wave classification guidance. | Content tests pass and `speckit-tasks`, `orchestrate`, `parallel-execution` describe AFK/HITL without new enforcement. |
| 2 — Bugfix feedback-loop gate | Require an automated pass/fail signal before bugfix hypothesis/RCA. | Content tests pass and `bugfix` + RCA template include Step 0 / feedback-loop sections. |
| 3 — Terminology enforcement | Add lightweight vocabulary guidance to clarify, knowledge map, and orchestrator dispatch. | Content tests pass and vocabulary alignment is documented without creating new skill or glossary system. |
| 4 — Verification | Prove the slices are bounded and regression-safe. | `go test ./internal/library` passes from `packages/cli`; scope audit confirms no forbidden files/features were added. |

> Tasks for each phase live in `tasks.md`. This table is the high-level roadmap.

---

## Out of Scope

Explicitly deferred. Anti-Speculation (Article IV) is enforced by listing what was *not* built.

- `workspace-configure` skill adaptation — explicitly excluded by scope.
- `architectural-deepening` skill adaptation — explicitly excluded by scope.
- `diagnose` skill adaptation — explicitly excluded by scope.
- `grill-me` skill adaptation — explicitly excluded by scope.
- Cupcake Rego policy changes — AFK/HITL is awareness, not a new physical enforcement layer.
- Pre-commit hook changes — existing anti-forgery and commit-size behavior remain unchanged.
- New top-level skills — all guidance goes into existing skill files.
- Runtime orchestration state, changed `ChainState`/`StepState`, or `get_status` behavior — prompt vocabulary only.
- Dependency, config, secret, or infrastructure changes — not required for Markdown guidance.
- New glossary engine or terminology database — `KNOWLEDGE_MAP.md` section is sufficient.
- Broad library rewrite or template reorganization — only listed files and test-only content contracts are planned.

---

## Downstream Contract

| Produces for | Filename |
|---|---|
| `speckit-tasks` | this file + issue #169 scope → `tasks.md` |
| `speckit-analyze` | this file + `tasks.md` |
| `speckit-implement` | this file (technical context) + task harness pointers |

---

## Approvals

| Role | Name | Date | Verdict |
|---|---|---|---|
| Author | Planner Agent | 2026-05-05 | Draft complete |
| Constitution check | Planner Agent | 2026-05-05 | PASS — pending human review |
| Human gate | TBD human | TBD | pending |
