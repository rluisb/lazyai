---
name: iterate
description: Iterate on implementation based on feedback or new requirements.
argument-hint: "[feedback]"
trigger: /iterate
phase: implement
---

# Iterate Skill

## Workflow (test→fix→verify loop)
1. Run failing test — confirm the failure is reproducible
2. Read error output — understand what went wrong
3. Apply minimal fix — smallest change that addresses the failure
4. Re-run test — verify the fix works
5. Run full suite — check for regressions
6. If still failing: repeat from step 1 (max 5 iterations)
7. If still failing after 5 iterations: STOP and escalate

## StructuredFeedback Consumption

When feedback includes a `StructuredFeedback` object, consume it as static prompt guidance for the approved task only:

1. Triage `requiredChanges` before `suggestions`; address blocking and high `priority` items first.
2. For each required change, restate the source (`requestedBy`), `targetPhaseOrStep` or target task/file, `evidence`/location, and recommended next action before editing.
3. Treat `blocksProgress: true` as a stop condition for reporting success until the required change is fixed or a blocker is documented.
4. Keep suggestions optional unless the human explicitly turns them into required changes.
5. If request_changes/rejected feedback omits required changes, priority, evidence, target phase, target task, or action detail, ask for clarification; do not guess.

If approved T021 runtime feedback is present as rejected-gate `structuredFeedback`, consume it with the same rules above. Do not claim broader runtime propagation, new gate outcomes, measurement hooks, or automated feedback handling.

## Completion Enforcement Checklist

Iteration is complete only when the approved feedback/task contract is satisfied. Continue the fix→verify loop within the max 5 iterations and approved task boundary until every task Done When item has evidence, or stop with a documented blocker.

Before reporting success, verify and record:

1. **Requirements complete:** every requested change, failing check, and task Done When item is addressed.
2. **Verification evidence:** the originally failing test/check is green and relevant quality gates are recorded.
3. **Risks and assumptions:** unresolved risks/assumptions are called out instead of silently treated as complete.
4. **Scope drift:** no unrelated files or behavior changed while fixing the feedback.
5. **Blocker/handoff path:** if the loop limit or external dependency prevents completion, STOP and escalate with a blocker, attempts tried, and next decision needed.

Use a concise `CompletionEnforcementReport` when closing the iteration: `status: done | blocked | not-done`, criteria with evidence, blockers, and out-of-scope changes.

## Reflexion Step (after each iteration)
- What did I try?
- Why didn't it work?
- What's different about my next attempt?

## Integration
- Agent: Builder
- Requires: failing test output
- Escalation: after 5 iterations, describe blocker and ask for help
