# Task 014: Wave 2 Baseline Capability + W1 Bake Audit

**Phase:** W2.0 — Baseline + decision gates  
**User Story:** Setup / all Wave 2 stories  
**Status:** TODO  
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

## Done When

- [ ] W1.B constraints are confirmed against current `origin/main`/worktree source.
- [ ] A Wave 2 decision record identifies which approval-gated tasks may proceed.
- [ ] Any lack of two-week bake evidence is explicitly noted for human review.
- [ ] No implementation/library/runtime files are changed unless needed for characterization tests.

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
