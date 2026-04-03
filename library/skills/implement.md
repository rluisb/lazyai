---
name: implement
description: Execute one approved plan phase with TDD discipline.
trigger: /implement
phase: implement
---

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

## Trace Protocol (ReAct, complex tasks only)

For multi-step or risky work, keep a concise trace:

1. **Thought:** key implementation consideration
2. **Action:** test/edit/command executed
3. **Observation:** concrete outcome
4. **Decision:** continue, revise, or escalate

Skip this for trivial, direct edits.

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

## Integration
- **Primary agent:** Builder (implementation phase)
- **Triggered by:** `/implement` command or workflow rule step 3
- **Depends on:** Plan output from Planner phase (techspec.md, tasks.md)
- **Feeds into:** iterate.md if tests fail; memory-write.md on completion
- **Related skills:** tdd-loop (test-first enforcement), anti-speculation (scope guard)
