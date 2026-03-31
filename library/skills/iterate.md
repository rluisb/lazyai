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
