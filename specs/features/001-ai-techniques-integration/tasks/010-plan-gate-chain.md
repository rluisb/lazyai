# Task 010: Move Approval to Sequential `plan-gate`

**Phase:** W1.B — Plan Quality + Adversarial Design Review  
**User Story:** P3 — Review plans before implementation approval  
**Status:** TODO  
**Depends on:** T9  
**Parallel with:** none

---

## Objective

Move the human approval gate off the `plan` step and onto an explicit sequential `plan-gate` step after `plan-quality`, preserving a simple sequential chain and ensuring plan-quality reports are visible before implementation approval.

## User Story / Spec References

- P3 — Human approvers see plan-quality findings before approving implementation.
- AC-D6-001 — `plan-quality` runs before the human approval gate.
- AC-D6-002 — The approval gate displays pass/warn/fail verdict and findings.
- AC-D6-004 — A fail verdict reaches the human approval gate; no automatic loop to `plan`.
- AC-D17-006 — The approval gate eventually displays merged findings when both reports exist.
- FR-008 through FR-010.
- `spec.md` §Internal Contracts sequential `feature` chain shape.
- `plan.md` §Sequential chain design for W1.B.

## Files to Change/Create

- `packages/ai-setup-go/library/orchestration/chains/feature.json` — move `gate: "user_approval"` from `plan` to new sequential `plan-gate`; insert `plan-quality` before `plan-gate`.
- `packages/orchestrator/src/__tests__/chain-machine.case.ts` — assert plan-quality success/fail outcomes route to `plan-gate` rather than `plan`/`implement`.
- `packages/ai-setup-go/internal/scaffold/orchestration_test.go` — assert source/install chain shape and installed `.ai/orchestration/chains/feature.json` includes the new order.
- `packages/ai-setup-ts/src/__tests__/orchestration.test.ts` — mirror chain-shape/install assertions.
- Any existing snapshot/fixture that records library orchestration chain content.

## Files Not to Touch

- `packages/ai-setup-go/library/skills/red-team-plan.md` — T11 owns this.
- `packages/ai-setup-ts/src/presets.ts` and custom flag UX — T12 owns this.
- Runtime conditional semantics beyond what T8 proved.
- W1.A task files under `specs/features/001-ai-techniques-integration/tasks/001-*.md` through `007-*.md`
- `specs/features/001-ai-techniques-integration/spec.md`
- `specs/features/001-ai-techniques-integration/plan.md`

## Test-First Order

1. Add failing chain-shape tests expecting step order: `research → plan → plan-quality → plan-gate → implement → review → fix/document` for the D6-only path.
2. Add failing test proving `plan` no longer has `gate: "user_approval"`.
3. Add failing test proving `plan-gate` is the only approval gate before implementation.
4. Add failing transition test proving `plan-quality` verdicts/outcomes proceed to `plan-gate`; no `fail → plan` automatic loop exists.
5. Add failing install/scaffold test proving generated `.ai/orchestration/chains/feature.json` receives the updated source chain.
6. Only after tests fail, edit source `feature.json` with the smallest sequential change.

## Done When

- [ ] Source `feature.json` remains a sequential `steps` array with no parallel block.
- [ ] `plan` transitions to `plan-quality` and no longer owns `gate: "user_approval"`.
- [ ] `plan-quality` transitions to `plan-gate` for all tested verdicts/outcomes.
- [ ] `plan-gate` owns `gate: "user_approval"` and transitions `approved → implement`, `rejected → plan` with feedback context expected by the chain.
- [ ] Tests prove a fail verdict does not automatically loop to `plan`.
- [ ] Install/scaffold tests prove the updated source chain reaches `.ai/orchestration/chains/feature.json`.

## Risks

- **Mid-flight chain compatibility:** active runs referencing old step IDs may break if changed in-place; mitigation is to follow the plan rollback policy before release.
- **Accidental auto-loop:** mitigated by transition tests for `plan-quality` fail/warn/pass.
- **Gate report not visible:** mitigated by making `plan-gate` description/prompt explicitly reference `PlanQualityReport` and by T13 integration tests.

## Constitution Check

- **Article I:** Reuse existing chain JSON and chain-machine tests.
- **Article II:** Chain-shape and transition tests fail before source chain edits.
- **Article III:** Step order and gate behavior are traced to `spec.md` and `plan.md`.
- **Article IV:** Do not add parallel blocks, graph execution, or auto-recovery loops.
- **Article V:** One explicit `plan-gate` step is the simplest way to surface reports before approval.
- **Article VI:** No workflow-engine abstraction or speculative transition framework.