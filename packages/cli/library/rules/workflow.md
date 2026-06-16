# Rule: Workflow

**Category:** Process
**Status:** Active

---

## Rule

Follow the standard development workflow for all changes. Every task must pass through the Purpose Gate before implementation begins.

## Rationale

Ensures consistency, quality, traceability, and prevents speculative implementation (YAGNI).

## Workflow

1. **Research:** Researcher agent explores the codebase and existing patterns
2. **Plan:** Planner agent creates a plan with acceptance criteria
3. **Implement:** Implementer agent executes the plan
4. **Review:** Reviewer agent verifies the implementation against the spec
5. **Merge:** Human review and merge

## Purpose Gate (Before Implementation)

Before writing any code, the implementer must verify:

1. **Is this behavior in the spec or task?** → If NO, don't build it
2. **Is this the simplest way to satisfy the spec?** → If NO, simplify
3. **Would removing this code cause a spec'd test to fail?** → If NO, remove it
4. **Does this belong in the current task's scope?** → If NO, flag it

If any answer indicates speculation, invoke the anti-speculation skill's Halt Protocol.

## TillDone Protocol

Each workflow step runs until its exit criteria are met — not until a time limit or iteration count:

- **Research**: Done when all questions about existing behavior are answered
- **Plan**: Done when every behavior in the spec has a task with test strategy
- **Implement**: Done when all task tests pass and quality gates are green
- **Review**: Done when all spec behaviors are verified and no scope violations exist

TillDone applies only to the approved task. It preserves one task per session and does not authorize scope expansion, speculative cleanup, runtime workflow changes, or extra features. If satisfying the task requires work outside the accepted contract, stop and ask for approval.

Before declaring a task done, the agent must check:

1. **Requirements complete:** every task Done When item has evidence.
2. **Verification evidence:** tests, quality gates, or justified checks are listed.
3. **Unresolved risks/assumptions:** risks and assumptions are closed or explicitly reported.
4. **Scope drift:** changed files and behavior remain inside the approved task.
5. **Blocker/handoff:** unmet criteria are reported as a documented blocker/handoff, not as done.

Use completion status precisely: `done` only when all criteria are met with evidence; `blocked` when completion requires an external decision or unavailable dependency; `not-done` when work remains inside the approved task.

## Human Gates

The following transitions require explicit human approval:

- Plan → Implement (human approves the plan)
- Review → Merge (human approves the PR)
- Any scope expansion beyond the original spec

## Anti-Speculation Integration

This rule works with the anti-speculation skill:
- **Workflow rule**: Defines the Purpose Gate (always-active constraint)
- **Anti-speculation skill**: Provides the 5 detection patterns and Halt Protocol (invoked when speculation detected)

## Enforcement

- Purpose Gate check at the start of every implementation task
- PR review checklist includes scope verification
- CI gates for tests and linting
- Reviewer agent checks for scope leakage in diffs
