# Task 008: Chain Gate + Conditional Capability Tests

**Phase:** W1.B — Plan Quality + Adversarial Design Review  
**User Story:** P3 — Review plans before implementation approval  
**Status:** DONE — evidence committed in `99741a7 test: document W1B chain capabilities`  
**Depends on:** W1.A merged complete  
**Parallel with:** none

---

## Objective

Characterize the existing chain runtime/scaffold capabilities before editing the feature chain. This task has now proved the current chain supports a sequential approval gate step and does not support runtime conditional red-team execution, compile-time orchestration template rendering, or parallel blocks. Downstream implementation must use explicit scaffold/install-time chain selection for enabled/disabled D17 chains.

## Evidence Produced

T8 is complete. The committed characterization evidence proves:

- Sequential gate support exists.
- Runtime conditional step execution is absent/unsupported.
- Compile-time orchestration template rendering is absent: chain JSON is copied byte-for-byte, so `{{#if features.adversarialDesign}}` remains literal.
- Parallel blocks are not present.

Downstream W1.B tasks must therefore use only sequential chain edits and explicit scaffold/install-time chain selection. They must not attempt runtime conditionals, source-chain template conditionals, or parallel blocks.

## User Story / Spec References

- P3 — Human approvers see plan-quality and optional adversarial findings before approving implementation.
- AC-D17-004 — `adversarialDesign=false` must not rely on unverified runtime conditional behavior; after T8, red-team must be omitted by explicit scaffold/install-time chain selection.
- AC-D6-001 — Plan quality must run before the human approval gate.
- AC-D6-004 — A fail verdict proceeds to the human gate; it does not automatically loop back to `plan`.
- FR-010 — Avoid automatic plan-quality fail loops.
- FR-014 — Verify conditional red-team inclusion/skipping before relying on it.
- `plan.md` §Sequential chain design for W1.B and §Execution Order T8.

## Files to Change/Create

- `packages/orchestrator/src/__tests__/chain-machine.case.ts` — add characterization tests for gate transitions and document absent/unsupported runtime conditional behavior.
- `packages/orchestrator/src/chain-machine.ts` — change only if the failing characterization exposes an existing intended capability that is incomplete; do not add new conditional semantics unless already documented and approved.
- `packages/ai-setup-go/internal/scaffold/orchestration_test.go` — add source/install chain-shape tests as needed.
- `packages/ai-setup-ts/src/__tests__/orchestration.test.ts` — mirror source/install chain-shape tests as needed.
- Test fixtures under existing package test locations if needed.

## Files Not to Touch

- `packages/ai-setup-go/library/orchestration/chains/feature.json` — T8 is a capability gate; do not edit the source chain here.
- `packages/ai-setup-go/library/skills/plan.md`
- `packages/ai-setup-go/library/skills/red-team-plan.md`
- `packages/ai-setup-ts/src/presets.ts`
- `packages/ai-setup-ts/src/wizard/phase2-features.ts`
- W1.A task files under `specs/features/001-ai-techniques-integration/tasks/001-*.md` through `007-*.md`
- `specs/features/001-ai-techniques-integration/spec.md`
- `specs/features/001-ai-techniques-integration/plan.md`

## Test-First Order

1. Add a failing/characterization test proving the chain machine can stop at a step with `gate: "user_approval"` and resume through explicit `approved`/`rejected` transitions.
2. Add a failing/characterization test proving the existing `feature.json` shape is a sequential `steps` array and has no parallel block requirement.
3. Search/test for runtime conditional-step support (for example `condition`, `optionalByFeature`, or equivalent) if present in the chain machine.
4. If runtime condition support exists, add a failing test proving a disabled conditional red-team step is skipped by the runtime.
5. If runtime condition support does not exist, add a failing test documenting that W1.B must use explicit scaffold/install-time chain selection rather than untested runtime skipping or generic template rendering.
6. Only after the tests fail for the expected reason, make the smallest production-test-support change required to pass characterization, or leave production unchanged if the tests already pass.

## Done When

- [x] Tests prove the chain remains sequential; no parallel block is required or introduced.
- [x] Tests prove explicit approval gates are supported as sequential steps.
- [x] Runtime conditional red-team behavior is documented by tests as unsupported/absent.
- [x] Compile-time orchestration template rendering is documented by tests as unsupported/absent; chain JSON is copied byte-for-byte.
- [x] If runtime conditionals are unsupported, the accepted fallback is explicit scaffold/install-time chain selection controlled by preset/config.
- [x] No source `feature.json` changes are made in this task.
- [x] Evidence is captured in commit `99741a7 test: document W1B chain capabilities`.

## Risks

- **Unsupported runtime/template conditional assumptions:** mitigated by making this the first W1.B task and requiring downstream tasks to use explicit chain selection instead of runtime conditionals or source-chain template rendering.
- **Accidentally designing a new workflow engine:** mitigated by limiting T8 to characterization or existing intended capability fixes only.
- **Go/TS scaffold drift:** mitigated by adding both Go and TS source/install chain-shape tests where relevant.

## Constitution Check

- **Article I:** Reuse existing chain machine and orchestration scaffold tests; do not introduce a workflow library.
- **Article II:** Capability tests are written before any feature-chain changes.
- **Article III:** Tests encode the `spec.md` and `plan.md` constraints before implementation tasks depend on them.
- **Article IV:** Do not add runtime conditionals, source-chain template rendering, or parallel blocks speculatively; use explicit scaffold/install-time chain selection downstream.
- **Article V:** Keep validation to sequential gates and concrete capability evidence only.
- **Article VI:** No new abstractions, no parallel engine, no generic chain template renderer, and no one-caller conditional framework.
