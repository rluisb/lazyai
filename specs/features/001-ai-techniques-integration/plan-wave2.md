# Feature 001 — Wave 2 Plan: Quality + Context

**Branch/worktree:** `/Users/ricardo/projects/teachable/ai-setup/.worktrees/feature-001-ai-techniques-wave2`  
**Inputs:** `research.md`, `plan.md`, `spec.md`, `tasks.md`, `tasks-w1b.md`, task files `001`–`013`, W1.B source outputs  
**Output:** planning artifacts only; no implementation in this wave-planning session

---

## Summary

Wave 2 should be implemented as a conservative sequence of content/skill/schema improvements before any runtime changes. The safest immediate value is in installed library guidance: generated knowledge, environment-aware planning, TillDone checks, causal RCA, safe recovery policy, structured feedback schema, and static lifecycle vocabulary. D4 chain-of-verification can ship as a skill/report first and only be inserted into the feature chain after an explicit human chain-shape approval. D9 has one bounded runtime option (persist rejected-gate structured feedback) but should be gated separately. D5 runtime auto-recovery and D8 runtime agent-state tracking are deferred.

The prior Wave 1 plan recommended a two-week W1 bake period before Wave 2. The user has approved planning now. Implementation tasks must still verify W1.B stability/evidence before touching chain/runtime behavior.

---

## Technical Context

| Area | Context |
|---|---|
| Languages | Go for installer/scaffold/library embedding; TypeScript for TS installer/orchestrator/runtime tests. |
| Runtime/orchestration | Existing `feature.json` is a sequential `steps` chain. W1.B added `plan-quality`, `plan-gate`, `red-team-plan`, and merged gate reporting. |
| W1.B evidence constraint | T8 proved sequential gates are supported; runtime conditionals, template-rendered chain JSON, and parallel blocks are not supported. |
| Primary safe surfaces | `packages/ai-setup-go/library/skills/*.md`, `library/fragments/*.md`, `library/rules/*.md`, `library/templates/*.md`, `library/agents/*.md`, `library/orchestration/chains/feature.json` only if approved. |
| Runtime-sensitive surfaces | `packages/orchestrator/src/chain-machine.ts`, `types.ts`, `tool-handlers.ts`, persistence. Touch only for approved D9 bounded feedback persistence; do not use for D5/D8 in Wave 2. |
| Testing | Existing pattern is test-first: frontmatter/snapshot/schema tests for library content, chain-machine tests for runtime behavior, Go/TS scaffold parity tests where install output changes. |
| Constraints | No code implementation in this planning task; no Wave 3/4 scope; no workflow-engine rewrite; no MCP config changes. |
| Performance goals | No new runtime background work in default Wave 2. Content changes should not affect chain execution latency. Optional chain verification adds one sequential step only if approved. |

---

## Decision Protocol

### Decision 1 — Wave 2 implementation shape

| Option | Pros | Cons | Effort | Decision |
|---|---|---|---|---|
| A. Runtime-first: implement D5/D8/D9 in orchestrator | More automation/observability | High schema/persistence risk; violates conservative scope; W1 bake not observed | M-L | Rejected |
| B. Content/skill-first with narrow approval gates | Low risk; aligns with existing library patterns; defers unsafe automation | Some techniques remain advisory | S-M | **Selected** |
| C. Defer all Wave 2 until bake complete | Lowest risk | No progress despite user approval | XS | Rejected for planning; implementation can still gate runtime tasks |

### Decision 2 — D4 chain-of-verification integration

| Option | Pros | Cons | Decision |
|---|---|---|---|
| Manual skill only | No chain risk; useful immediately | Not automatically enforced | Default first slice |
| Sequential `chain-verify` step | Stronger completion gate; uses existing chain model | Changes default chain; active-run compatibility risk | Requires human approval |
| Parallel verifier | Faster, but unsupported | T8 forbids parallel assumptions | Rejected |

### Decision 3 — D9 structured feedback

| Option | Pros | Cons | Decision |
|---|---|---|---|
| Schema + prompt guidance only | Safe, no runtime schema changes | Feedback not automatically propagated | Ship first |
| Persist feedback in gate/step output via existing `output` | Bounded runtime value | Needs tests and compatibility review | Approval-gated task |
| New gate engine/outcome taxonomy | More expressive | Scope creep and runtime rewrite | Rejected |

### Decision 4 — D5/D8 runtime capabilities

| Item | Wave2-compatible | Deferred runtime capability |
|---|---|---|
| D5 Auto-recovery | Static safe-recovery allowlist and human-gated policy | Failure classifier, autonomous recovery execution, audit automation |
| D8 Agent state machine | Static lifecycle vocabulary in reports/handoffs | Persisted per-agent lifecycle states in `ChainState`/`StepState` and `get_status` |

---

## Constitution Check

| Article | Verdict | Rationale |
|---|---|---|
| I — Library-First | PASS | Prefer existing markdown skill/rule/template surfaces and existing sequential chain runtime before new code. |
| II — TDD | PASS | Each task requires failing schema/snapshot/chain tests before production/library edits. |
| III — Docs as Source of Truth | PASS | `spec-wave2.md` is the Wave 2 acceptance contract; tasks reference FR/AC IDs. |
| IV — YAGNI | PASS | Wave 3/4, RAG, model routing, learning, runtime D5/D8 are out of scope or deferred behind approval. |
| V — Simplicity | PASS | Content-first, one skill/report at a time, optional sequential chain step only. |
| VI — Anti-Overengineering | PASS | No broad validator framework, no new scheduler, no generic chain templating, no one-caller runtime abstractions. |

---

## Project Structure

```text
specs/features/001-ai-techniques-integration/
  spec-wave2.md              # Wave 2 acceptance contract
  plan-wave2.md              # This plan
  tasks-wave2.md             # Master task index/dependencies
  tasks/014-*.md ...         # Individual implementation harnesses

packages/ai-setup-go/library/
  fragments/reasoning-protocol.md       # N2/D14 content target
  skills/plan.md                        # D14 planning guidance target
  skills/implement.md, iterate.md       # N12 completion guidance target
  skills/review.md                      # N12 review check target
  skills/bugfix.md                      # D13 target
  skills/chain-verify.md                # D4 new skill target
  rules/workflow.md                     # N12 target
  rules/auto-recovery.md                # D5 static policy target
  rules/structured-feedback.md          # D9 static schema target
  rules/agent-state.md                  # D8 static taxonomy target
  templates/bugfix-rca-template.md      # D13 target

packages/ai-setup-go/library/orchestration/chains/
  feature.json                          # Optional D4 sequential step target after approval

packages/orchestrator/src/
  chain-machine.ts, types.ts            # D9 runtime propagation target only if approved
```

---

## Phases & Milestones

| Phase | Tasks | Exit Criteria |
|---|---|---|
| W2.0 Baseline + decision gates | T014 | W1.B chain/runtime constraints re-confirmed; approval decisions recorded before runtime/chain work. |
| W2.A Context + causal content | T015, T016 | Generated knowledge, environment snapshot, and causal RCA guidance pass content tests/snapshots. |
| W2.B Verification + completion | T017, T018, T019 | `chain-verify` report contract exists; optional chain integration approved/tested; TillDone checks installed. |
| W2.C Feedback/recovery/state guidance | T020, T021, T022, T023 | Structured feedback schema, safe recovery policy, and lifecycle vocabulary installed; runtime propagation only if approved. |
| W2.D Integration validation | T024 | Wave 2 ACs trace to tests/snapshots; no Wave 3/4 artifacts or unsupported chain constructs introduced. |

### Dependency Graph

```text
T014 baseline
├── T015 generated knowledge + env planning [P]
├── T016 causal reasoning [P]
├── T017 chain-verify skill [P]
│   └── T018 optional feature-chain integration (approval gate)
├── T019 TillDone completion enforcement [P]
├── T020 structured feedback static contract [P]
│   └── T021 structured feedback runtime propagation (approval gate)
├── T022 auto-recovery static policy [P]
├── T023 agent-state static taxonomy [P]
└── T024 integration validation (depends on T015-T023; skips unapproved deferred runtime tasks)
```

---

## Risks & Mitigations

| Risk | Impact | Mitigation | Owner |
|---|---|---|---|
| W1.B has not baked in production use | Chain/runtime regressions could be compounded | T014 requires evidence check; runtime tasks need human approval | Planner / implementor |
| Chain verification becomes a broad semantic evaluator | False failures and overengineering | Start with explicit rule/report contract and warn on ambiguity | T017 owner |
| TillDone conflicts with one-task-per-session | Agents may overrun scope | State that TillDone means complete the approved task or document blocker | T019 owner |
| Structured feedback runtime change breaks existing gate callers | Gate advancement compatibility issue | Keep outcomes approved/rejected; persist feedback as optional output only; approval-gated | T021 owner |
| Auto-recovery compounds errors | Unsafe autonomous edits | Static policy only; runtime automation deferred | T022 owner |
| Agent state machine triggers schema migration | Runtime compatibility risk | Static vocabulary only; runtime state tracking deferred | T023 owner |

---

## Complexity Tracking

| Complexity | Justification | Status |
|---|---|---|
| New `chain-verify` skill/report | Needed for D4; limited to markdown + schema tests | Allowed |
| Optional feature-chain step | Existing sequential chain supports this; no runtime engine change | Requires approval |
| D9 runtime persistence | Small but runtime-facing | Deferred until explicit approval |
| D5 runtime auto-recovery | Requires classifier/policies/audit | Deferred |
| D8 runtime state machine | Requires type/persistence/status changes | Deferred |

---

## Human Decisions Needed

1. Approve or reject adding `chain-verify` as a sequential default feature-chain step after manual skill/report lands.
2. Approve or reject bounded D9 runtime feedback propagation through existing gate output/state.
3. Confirm D5 runtime auto-recovery remains deferred after static policy lands.
4. Confirm D8 runtime state tracking remains deferred after static lifecycle vocabulary lands.
5. Before implementation, decide whether W1.B stability evidence is sufficient or whether runtime-touching tasks should wait.
