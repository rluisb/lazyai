# Task 019: TillDone Completion Enforcement Guidance [P]

**Phase:** W2.B — Verification + completion  
**User Story:** [US1] Verification + completion  
**Status:** TODO  
**Depends on:** T014  
**Parallel with:** T015, T016, T017, T020, T022, T023

---

## Objective

Strengthen implement/iterate/review guidance so agents cannot declare completion until all approved task criteria have evidence or a blocker is documented.

## Spec References

- FR-W2-005 through FR-W2-007.
- AC-N12-001 through AC-N12-003.
- `spec-wave2.md` `CompletionEnforcementReport` contract.

## Files to Change/Create

- `packages/ai-setup-go/library/skills/implement.md`
- `packages/ai-setup-go/library/skills/iterate.md`
- `packages/ai-setup-go/library/skills/review.md`
- `packages/ai-setup-go/library/rules/workflow.md`
- Existing skill/rule snapshot tests.

## Files NOT to Touch

- Orchestrator runtime files.
- Feature chain JSON.
- Existing W1 task files `001`–`013`.
- Broad task runner/session management code.

## Test-First Order

1. Add failing snapshots asserting implement and iterate skills require a Done When/evidence checklist before completion.
2. Add failing review guidance assertions for early-stop detection and out-of-scope change detection.
3. Add failing assertions that TillDone allows a documented blocker and preserves one-task-per-session limits.
4. Add content/schema tests for `CompletionEnforcementReport` if report output is required by the selected guidance.
5. Edit only the targeted skill/rule markdown.

## Done When

- [ ] Implement/iterate guidance requires evidence for every task Done When item.
- [ ] Review guidance checks early stop and missing evidence.
- [ ] Blocked status is explicit and acceptable when criteria cannot be met.
- [ ] One task per session remains intact.

## Risks

- **Agents overrun scope:** mitigate by stating TillDone applies only to the approved task.
- **Checklist fatigue:** keep output concise and evidence-based.

## Constitution Check

- **Article I:** Reuse existing skills/rules.
- **Article II:** Snapshot tests precede content edits.
- **Article III:** Criteria map to Wave 2 contract.
- **Article IV:** No runtime enforcement engine.
- **Article V:** Checklist guidance is simpler than automated semantic completion detection.
- **Article VI:** No new generic completion framework.
