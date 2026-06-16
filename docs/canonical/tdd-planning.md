# TDD Planning

Every implementation or behavior-affecting code change MUST choose a TDD mode during research or planning.

## Test Preservation Rule

- Existing tests MUST NOT be deleted, skipped, weakened, or rewritten to pass unless the user, plan, or spec explicitly authorizes it.
- Obsolete tests require: obsolete behavior, approval source, and replacement coverage.
- Never remove tests just because they fail.

## Modes

### lightweight

Use for low-risk, narrow changes.

- Artifact: test intent in the task plan.
- Minimum: one failing focused check or explicit exemption before implementation.
- Validate: focused test or smallest command covering the change.

### medium

Use for normal feature, bugfix, parser, validation, or API work.

- Artifact: `## TDD Plan` inside the task/feature plan.
- Include: behavior contract, red test names, edge cases, verification command.
- Validate: red failure observed, green focused test, then next-smallest check.

### heavy-aggressive

Use for security, money, data loss, migrations, concurrency, or high-regression-risk work.

 Artifact: standalone `specs/tdd/<slug>.md` or feature spec section.
- Include: unit, integration/contract, failure, boundary, and regression tests.
- Validate: focused tests plus the smallest suite proving the integration boundary.

### required

Use when user/spec/incident demands TDD or when code touches public contracts after a regression.

- Implementation is blocked until the red test exists or exemption is approved.
- Plan must name test file paths and assertions.
- Verification evidence must include red and green outputs.

## Exemption

```markdown
Test-first exemption: <why a failing test is not practical>
Approval source: <user | plan | spec | existing repo constraint>
Validation instead: <command or manual scenario>
Risk: <what remains uncovered>
```
