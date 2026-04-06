---
name: tdd-loop
trigger: /tdd-loop
phase: implement
---

# TDD Loop Skill

## RED → GREEN → REFACTOR Cycle

### RED — Write a failing test
- One behavior per test
- Test should fail for the right reason (not a syntax error)
- If the test passes immediately, you misunderstood the requirement — STOP

### GREEN — Make it pass
- Write the minimum code to pass the test
- No premature optimization, no extra features
- If the change is >30 lines, you're probably doing too much — break it down

### REFACTOR — Clean up
- Improve names, extract functions, remove duplication
- Tests must still pass after every refactor step
- No behavior changes during refactor

## Iteration Limits
- Max 10 RED→GREEN→REFACTOR cycles per session
- If stuck after 3 cycles on the same issue: STOP and ask

## Integration
- Agent: Builder
- Used by: `implement` skill as its inner loop
