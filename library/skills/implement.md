# Implement Skill

**Command:** `/implement [plan-file] [phase]`
**Goal:** Execute one plan phase using TDD and disciplined progress updates.

---

## Workflow

1. **Read Plan Phase**
   - Confirm target tasks, scope boundaries, and acceptance criteria.
2. **Write Failing Tests First**
   - Add or update focused tests for one task at a time.
3. **Implement Minimum Change**
   - Write the smallest production code needed to pass tests.
4. **Refactor Safely**
   - Improve clarity and maintainability without changing behavior.
5. **Run Quality Gates**
   - Execute required checks and fix regressions before continuing.
6. **Update Progress**
   - Mark task status, record outcomes, and note follow-ups.

## Principles

- Small commits
- One task at a time
- Checkpoint after each task

## Output Format

```markdown
## Implementation Run: [Plan File] — [Phase]

### Task Status
- [task ID]: [in progress | done | blocked]

### Test Cycle
- Added failing test: [test name]
- Implementation change: [file + summary]
- Result: [green/red]

### Quality Gates
- [command]: [pass/fail]

### Notes
- [risks, follow-ups, or handoff items]
```
