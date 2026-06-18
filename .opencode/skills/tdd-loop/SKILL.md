---
name: tdd-loop
description: Execute test-driven development loop: red, green, refactor.
argument-hint: "[test-or-behavior]"
trigger: /tdd-loop
phase: implement
---
## Quick Reference

| | |
|---|---|
| **Use when** | [When to use this skill] |
| **Do not use when** | [When NOT to use this skill] |
| **Primary agent** | [Which agent uses this] |
| **Runtime risk** | [Low/Medium/High] |
| **Outputs** | [What this skill produces] |
| **Validation** | [How to validate output] |
| **Deep mode trigger** | [How to trigger full mode] |



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
