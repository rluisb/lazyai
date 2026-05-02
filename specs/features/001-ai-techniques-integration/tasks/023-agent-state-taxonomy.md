# Task 023: Agent Lifecycle / State Taxonomy Guidance [P]

**Phase:** W2.C — Feedback/recovery/state guidance  
**User Story:** [US3] Actionable feedback/recovery/state  
**Status:** TODO  
**Depends on:** T014  
**Parallel with:** T015, T016, T017, T019, T020, T022

---

## Objective

Add a static agent lifecycle vocabulary for status reports, handoffs, and recovery summaries without adding runtime per-agent state tracking.

## Spec References

- FR-W2-020 through FR-W2-022.
- AC-D8-001 through AC-D8-003.
- `plan-wave2.md` Decision 4.

## Files to Change/Create

- Create `packages/ai-setup-go/library/rules/agent-state.md` or equivalent existing rule location.
- `packages/ai-setup-go/library/skills/orchestrate.md`
- `packages/ai-setup-go/library/templates/task-harness-template.md` only if lifecycle labels belong in task handoffs/status evidence.
- Existing rule/template/skill snapshot tests.

## Files NOT to Touch

- `packages/orchestrator/src/types.ts`, `chain-machine.ts`, `tool-handlers.ts`, persistence, `get_status` behavior.
- Feature chain JSON.
- D5 runtime auto-recovery.
- Existing W1 task files `001`–`013`.

## Test-First Order

1. Add failing content tests for lifecycle labels: `loading_context`, `planning`, `awaiting_approval`, `executing`, `verifying`, `blocked`, `handoff`, `done`, `error`.
2. Add failing snapshots proving orchestrate/handoff guidance uses labels in status/recovery summaries.
3. Add negative assertions that guidance says labels are report vocabulary only in Wave 2.
4. Update rule/skill/template markdown only.

## Done When

- [ ] Lifecycle vocabulary is documented and validated.
- [ ] Status/handoff/recovery guidance explains how to use labels.
- [ ] Runtime state tracking is explicitly deferred.
- [ ] No `ChainState`, `StepState`, persistence, or `get_status` changes are made.

## Risks

- **Semantic mismatch with runtime step state:** mitigate by naming this report vocabulary only.
- **Schema creep:** defer persisted state fields to a future ADR.

## Constitution Check

- **Article I:** Reuse rule/skill/template guidance.
- **Article II:** Snapshot tests precede edits.
- **Article III:** Labels trace to `spec-wave2.md`.
- **Article IV:** No runtime state machine in Wave 2.
- **Article V:** Static vocabulary is the simplest useful increment.
- **Article VI:** No state-machine abstraction or persistence migration.
