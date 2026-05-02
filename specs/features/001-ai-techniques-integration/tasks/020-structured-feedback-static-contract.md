# Task 020: Structured Feedback Schema + Static Guidance [P]

**Phase:** W2.C — Feedback/recovery/state guidance  
**User Story:** [US3] Actionable feedback/recovery/state  
**Status:** TODO  
**Depends on:** T014  
**Parallel with:** T015, T016, T017, T019, T022, T023

---

## Objective

Define the `StructuredFeedback` schema and update iterate/orchestrate guidance to consume actionable human feedback without changing runtime gate persistence.

## Spec References

- FR-W2-011, FR-W2-012.
- AC-D9-001, AC-D9-002.
- `spec-wave2.md` `StructuredFeedback` contract.

## Files to Change/Create

- Create `packages/ai-setup-go/library/rules/structured-feedback.md` or equivalent existing rule location.
- `packages/ai-setup-go/library/skills/iterate.md`
- `packages/ai-setup-go/library/skills/orchestrate.md`
- Existing rule/skill snapshot or frontmatter tests.

## Files NOT to Touch

- `packages/orchestrator/src/chain-machine.ts`, `types.ts`, `tool-handlers.ts` — T021 only if approved.
- Feature chain JSON.
- Existing W1 task files `001`–`013`.

## Test-First Order

1. Add failing schema fixture tests for valid/invalid `StructuredFeedback` examples.
2. Add failing snapshots proving iterate/orchestrate guidance references required changes, suggestions, priority, evidence, and target phase/task.
3. Add failing assertion that missing required-change detail triggers clarification rather than guessing.
4. Add negative assertions proving no runtime propagation is claimed in this task.
5. Add the rule/guidance content.

## Done When

- [ ] Structured feedback schema is documented and tested.
- [ ] Iterate/orchestrate guidance consumes structured feedback when present.
- [ ] Ambiguous rejection feedback has a clarification path.
- [ ] No runtime gate-state changes are made.

## Risks

- **Schema too heavy for humans:** keep required fields minimal and suggestions optional.
- **Agents assume runtime support:** explicitly state static guidance only until T021 approval.

## Constitution Check

- **Article I:** Reuse rule/skill markdown surfaces.
- **Article II:** Schema/snapshot tests first.
- **Article III:** Schema comes from `spec-wave2.md`.
- **Article IV:** Runtime propagation is separate and approval-gated.
- **Article V:** One schema/rule is simpler than a new gate system.
- **Article VI:** No generic feedback platform.
