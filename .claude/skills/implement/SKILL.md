---
name: implement
description: Implement requested changes safely with test-first workflow.
argument-hint: "[task-or-scope]"
trigger: /implement
phase: implement
---

# Implement Skill

## Workflow
1. Read task — confirm scope, acceptance criteria, and constraints
2. Write failing tests first — one behavior at a time
3. Implement minimum change to pass tests
4. Refactor safely — no behavior changes
5. Run quality gates — lint, typecheck, test, build. Fix before continuing.
6. Update progress — mark task complete, record outcomes

## Completion Enforcement Checklist

Do not declare the implementation done until the approved task contract is satisfied. Continue within the approved task until every task Done When item has evidence, or stop with a documented blocker.

Before reporting completion, verify and record:

1. **Requirements complete:** every task Done When item and acceptance criterion is met, with no skipped criteria.
2. **Verification evidence:** tests, quality gates, build/typecheck/lint, or justified smoke checks are listed with exact commands or artifact references.
3. **Risks and assumptions:** unresolved risks/assumptions are closed, explicitly accepted, or listed as blockers.
4. **Scope drift:** changed files match the approved task scope; any out-of-scope change is reverted or reported.
5. **Blocker/handoff path:** if criteria cannot be met, report `blocked` with the unmet criteria, evidence gathered, and the next required decision.

Use a concise `CompletionEnforcementReport` in the final response when the task has explicit Done When criteria: `status: done | blocked | not-done`, criteria with evidence, blockers, and out-of-scope changes.

## Trace Protocol (complex implementations only)
1. **Thought**: key consideration for this step
2. **Action**: test write / code edit / command run
3. **Observation**: concrete outcome
4. **Decision**: continue / revise / escalate

## Integration
- Agent: Builder
- Requires: plan output (plan.md + tasks/)
- On failure feeds into: `iterate` skill
- On success feeds into: `memory-write` skill (if lessons learned)
