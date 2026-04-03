---
name: iterate
description: Run focused test-fix-verify loops until targeted scope is green.
trigger: /iterate
phase: iterate
---

# Iterate Skill

**Command:** `/iterate [task-file]`
**Goal:** Run a focused test→fix→verify loop for a single task until green.

---

## Ralph Loop Workflow

1. **Run Tests**
   - Execute the relevant test scope for the task.
2. **Identify Failures**
   - Isolate the first actionable failure.
3. **Fix One Failure at a Time**
   - Apply the smallest change that addresses the root cause.
4. **Re-run Tests**
   - Validate the fix and identify the next failing point.
5. **Repeat Until Green**
   - Continue loop discipline until targeted tests pass.
6. **Run Full Quality Gates**
   - Confirm broader project health before handoff.

## Reflexion Step

When a loop iteration fails:

1. **Verbalize the failure** — Write a brief reflection in the task journal:
   > "This attempt failed because [specific reason]. The approach of [what was tried]
   > did not work due to [root cause]. In the next attempt, I will [adjusted strategy]."

2. **Preserve the reflection** — This verbal reflection MUST be included in session compaction
   as high-priority content. It prevents repeating the same failed approach.

3. **Reset with context** — When starting a fresh session for the next attempt, include:
   - The original requirement
   - The reflection from the failed attempt
   - Any partial progress that should be preserved

4. **Escalate after 2 reflections** — If two consecutive iterations produce failure reflections
   on the same issue, escalate: invoke the Reviewer agent or request human guidance.

## Principles

- Fix-only mode (no feature creep)
- Smallest possible change

## Output Format

```markdown
## Iterate Run: [Task]

### Current Failure
- [test] — [error summary]

### Fix Applied
- [file] — [what changed and why]

### Verification
- Targeted tests: [pass/fail]
- Full gates: [pass/fail]

### Next Step
- [continue loop | ready for review]
```

## Integration
- **Primary agent:** Implementor (debug loop)
- **Triggered by:** `/iterate` command or failing tests during implementation
- **Depends on:** Latest failing test output, error context, and task scope
- **Feeds into:** Continued implementation when green, then `memory-write` for captured learnings
