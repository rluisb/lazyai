# Task 014: Wave 2 Baseline Capability + W1 Bake Audit

**Phase:** W2.0 — Baseline + decision gates  
**User Story:** Setup / all Wave 2 stories  
**Status:** DONE  
**Depends on:** W1.A + W1.B merged on `origin/main`  
**Parallel with:** none

---

## Objective

Re-confirm W1.B runtime constraints, collect available W1 bake evidence, and record the human decisions required before any Wave 2 chain/runtime implementation begins.

## Spec References

- `spec-wave2.md` §1 Scope Boundary and §6 Approval Decisions Required.
- `plan-wave2.md` §Technical Context, §Human Decisions Needed.
- T8 evidence: sequential gates supported; runtime conditionals, template-rendered chain JSON, and parallel blocks unsupported.

## Files to Change/Create

- Create a Wave 2 implementation note or test evidence file only if the project already has an accepted evidence location for feature work.
- Add/update tests only if existing W1.B capability tests do not cover the required constraints.

## Files NOT to Touch

- `packages/ai-setup-go/library/skills/*` — content tasks own these.
- `packages/ai-setup-go/library/orchestration/chains/feature.json` — T018 owns optional chain change.
- `packages/orchestrator/src/chain-machine.ts`, `types.ts`, `tool-handlers.ts` — T021 only if approved.
- Existing W1 task files `001`–`013`.
- Wave 3/4 docs or implementation.

## Test-First Order

1. Re-run or add characterization tests proving the feature chain remains sequential and no runtime conditionals/template chain rendering/parallel blocks are assumed.
2. Capture W1.B chain/gate evidence available from merged tests, incidents, or user reports.
3. Record explicit decisions: T018 approved/deferred, T021 approved/deferred, D5 runtime deferred, D8 runtime deferred.
4. If evidence contradicts W1.B assumptions, stop and request revised Wave 2 planning before implementation.

## Evidence Produced

- T8 evidence commit `99741a7` is an ancestor of both `HEAD` and `origin/main`, so the W1.B capability baseline is present in the current worktree and merged mainline.
- Existing characterization tests cover the T014 constraints; no new tests were needed for this audit:
  - `packages/ai-setup-ts/src/__tests__/orchestration.test.ts` proves source and installed feature chains are sequential `steps` arrays, contain no `parallel` blocks, and contain no `{{#if}}`, `optionalByFeature`, or `condition` markers.
  - `packages/orchestrator/src/__tests__/chain-machine.case.ts` proves runtime conditional metadata is not skipped by the chain machine, sequential `plan-gate` approval/rejection transitions work, and W1.B gate report behavior is visible at the human approval gate.
  - `packages/ai-setup-go/internal/scaffold/orchestration_test.go` mirrors Go scaffold chain-shape checks for base and adversarial feature-chain installs.
- Current W1.B source matches the expected baseline:
  - `packages/ai-setup-go/library/orchestration/chains/feature.json` is the base sequential chain with `plan -> plan-quality -> plan-gate` before implementation.
  - `packages/ai-setup-go/library/orchestration/chains/feature-adversarial.json` is the explicit adversarial sequential variant with `red-team-plan` between `plan-quality` and `plan-gate`.
  - `packages/orchestrator/src/chain-machine.ts` keeps explicit `approved`/`rejected` gate handling and does not add D9 runtime feedback propagation, D5 runtime auto-recovery, or D8 runtime state tracking.
- Available bake evidence is merged test/source evidence only. No two-week production bake report, incident log, or user-report evidence was found in the repo during this audit.

## Decision Record

- **T018 D4 chain integration:** Not approved by this audit. Requires a human approve/defer decision before adding `chain-verify` to the default feature chain.
- **T021 D9 runtime feedback propagation:** Not approved by this audit. Requires a human approve/defer decision before touching runtime gate/step state.
- **D5 runtime auto-recovery:** Deferred. Wave 2 may add only static safe-recovery guidance unless a separate approval/ADR authorizes autonomous runtime recovery.
- **D8 runtime state tracking:** Deferred. Wave 2 may add only static lifecycle vocabulary unless a separate approval/ADR authorizes `ChainState`/`StepState` or `get_status` schema changes.
- **W1.B bake:** Merged tests and current-source checks are sufficient for content-only Wave 2 tasks, but runtime-touching or chain-shape tasks should wait for explicit human confirmation that W1.B stability evidence is sufficient.

## Done When

- [x] W1.B constraints are confirmed against current `origin/main`/worktree source.
- [x] A Wave 2 decision record identifies which approval-gated tasks may proceed.
- [x] Any lack of two-week bake evidence is explicitly noted for human review.
- [x] No implementation/library/runtime files are changed unless needed for characterization tests.

## Risks

- **W1.B not baked:** mitigate by documenting evidence and gating runtime changes.
- **Scope drift into implementation:** this task is characterization/decision recording only.

## Constitution Check

- **Article I:** Reuse existing W1.B tests and evidence rather than inventing a new audit harness.
- **Article II:** Any new characterization must be test-first and fail for the expected reason.
- **Article III:** Decisions trace to `spec-wave2.md` and `plan-wave2.md`.
- **Article IV:** Do not begin Wave 2 implementation here.
- **Article V:** A small evidence/decision record is simpler than a new governance system.
- **Article VI:** No workflow abstractions or runtime behavior changes.
