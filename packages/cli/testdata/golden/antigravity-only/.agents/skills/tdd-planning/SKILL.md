---
name: tdd-planning
description: Use during research or planning before implementation to choose a TDD mode, define red tests, preserve existing tests, and produce the test-first artifact for a feature, bugfix, or refactor.
---

# TDD Planning

## When to Use

Use this skill when:
- Planning any implementation or behavior-affecting code change.
- Research reveals branches, edge cases, or regressions that need tests.
- The user asks for TDD, tests first, or no test deletion.
- A task is critical enough to need medium, heavy-aggressive, or required TDD.

Do not use for documentation-only changes unless docs generation has executable behavior.

## Rule

Every implementation plan chooses a TDD mode from `canonical/tdd-planning.md` before code is written.

## Workflow

1. Identify the behavior contract in one sentence.
2. Select mode: lightweight, medium, heavy-aggressive, or required.
3. Name the red test file(s), test case(s), and expected failure.
4. Record existing tests that must remain intact.
5. Define green criteria and the smallest verification command.
6. Add the TDD plan to the task plan or standalone `specs/tdd/<slug>.md` for heavy-aggressive/required work.

## TDD Plan Template

```markdown
## TDD Plan

Mode: <lightweight | medium | heavy-aggressive | required>
Behavior: <one sentence>
Red tests:
- <test file>::<test name> — fails because <missing behavior>
Existing tests preserved:
- <test file or pattern>
Green criteria:
- <observable behavior>
Verification:
- Red: <command>
- Green: <command>
Exemption: <none | approved exemption block>
```

## Constraints

- Existing tests must not be removed, skipped, or weakened without explicit user/spec/plan approval.
- Obsolete tests require replacement coverage or a documented reason no replacement is needed.
- Required mode blocks implementation until the red test exists or an exemption is approved.
- Do not test private plumbing when a public seam can observe behavior.

## Verification Checklist

- [ ] TDD mode chosen and justified.
- [ ] Red test path and expected failure named.
- [ ] Existing tests preservation recorded.
- [ ] Verification command is headless and focused.
- [ ] Exemption block exists when no red test is practical.
