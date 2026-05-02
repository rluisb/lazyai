# Task 021: Structured Feedback Runtime Propagation

**Phase:** W2.C — Feedback/recovery/state guidance  
**User Story:** [US3] Actionable feedback/recovery/state  
**Status:** DEFERRED TODO — requires explicit human approval after T020  
**Depends on:** T020 and approval for bounded runtime change  
**Parallel with:** none

---

## Objective

If approved, persist structured rejected-gate feedback through existing chain state/output fields so the target step can see actionable feedback without changing gate outcomes or building a new gate engine.

## Spec References

- FR-W2-013.
- AC-D9-003.
- `plan-wave2.md` Decision 3.

## Files to Change/Create

- `packages/orchestrator/src/chain-machine.ts`
- `packages/orchestrator/src/types.ts` only if optional typing is needed.
- `packages/orchestrator/src/__tests__/chain-machine.case.ts`
- `packages/orchestrator/src/__tests__/tool-handlers.case.ts` only if handler behavior exposes the feedback.

## Files NOT to Touch

- Outcome taxonomy beyond existing `approved` / `rejected` gate decisions.
- Feature chain JSON.
- D5 auto-recovery runtime code.
- D8 runtime state tracking.
- Persistence migrations unless explicitly approved by human review.

## Test-First Order

1. Confirm T014/T020 recorded approval for runtime feedback propagation.
2. Add failing tests proving `advance_chain` with rejected gate and `output.structuredFeedback` stores feedback on the gate/step state.
3. Add failing tests proving approved/rejected outcome semantics are unchanged.
4. Add failing tests proving the next target step can access feedback via existing state snapshot/output context.
5. Add backward-compatibility tests: gate decisions without feedback still work.
6. Implement the smallest optional-field propagation needed.

## Done When

- [ ] Human approval for runtime change is recorded.
- [ ] Structured rejected feedback is persisted without changing gate outcomes.
- [ ] Existing callers that omit feedback remain compatible.
- [ ] No new gate engine, conditionals, or recovery automation are introduced.

## Risks

- **Runtime schema compatibility:** mitigate by optional fields and backward-compatible tests.
- **Scope creep:** reject new outcomes, UI, persistence migrations, or feedback analytics.

## Constitution Check

- **Article I:** Reuse existing `output`/state structures where possible.
- **Article II:** Runtime tests fail before code changes.
- **Article III:** Behavior traces to D9 contract only.
- **Article IV:** No new gate engine or D5/D8 runtime work.
- **Article V:** Optional propagation is the simplest runtime enhancement.
- **Article VI:** No broad abstractions or generic feedback framework.
