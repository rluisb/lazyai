---
name: iterate
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

## Reflexion Step (after each iteration)
- What did I try?
- Why didn't it work?
- What's different about my next attempt?

## Integration
- Agent: Builder
- Requires: failing test output
- Escalation: after 5 iterations, describe blocker and ask for help
