---
description: Execute explicit tasks with smallest failing checks first and cleanup only after smoke verification.
scripts:
  sh: scripts/bash/check-prerequisites.sh --json --require-tasks --include-tasks
  ps: scripts/powershell/check-prerequisites.ps1 -Json -RequireTasks -IncludeTasks
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Pre-Execution Checks

Check `.specify/extensions.yml` for `hooks.before_implement`.
- Skip silently if the file or hook list is missing.
- Ignore entries where `enabled: false`.
- Do not evaluate hook `condition` expressions.
- For executable mandatory hooks emit:
  - `EXECUTE_COMMAND: {command}`
- For executable optional hooks present them as optional follow-up commands.

## Outline

1. Run `{SCRIPT}` from repo root and parse `FEATURE_DIR`, `AVAILABLE_DOCS`, and `TASKS`.
2. Load `FEATURE_DIR/tasks.md`, `FEATURE_DIR/plan.md` when present, and `/memory/constitution.md` if it exists.
3. Execute only tasks that are explicitly written in `tasks.md` or explicitly requested by the user.
   - Do not invent repository housekeeping, ignore-file updates, dependency audits, refactors, documentation, or formatting tasks.
   - Do not run fixed phase checklists that are absent from the task list.
   - If a task is ambiguous enough that execution would require guessing scope, stop and report the blocker.
4. For behavior-affecting work, use the smallest test-first loop:
   - identify the smallest failing check that proves the next explicit task;
   - run only that focused check first and record the observed failure;
   - make the smallest source change that targets the failure;
   - re-run the same focused check before widening verification.
5. Preserve existing tests and verification truth.
   - Never delete, skip, weaken, or relabel an existing test to pass.
   - Do not claim integration, runtime, parity, performance, or gate coverage unless that exact check was observed.
6. Keep execution dependency-ordered.
   - Complete prerequisite tasks before dependent tasks.
   - Keep each change tied to one task and its acceptance check.
7. Only after the requested behavior works and a smoke verification has passed, perform explicit cleanup/docs tasks from `tasks.md`.
   - If `tasks.md` contains no cleanup/docs task, do not add one.
   - If cleanup changes behavior, return to the smallest failing-check loop.
8. Mark completed tasks in `tasks.md` only after their acceptance checks have been observed.

## Post-Execution Hooks

Check `.specify/extensions.yml` for `hooks.after_implement`.
- Use the same mandatory/optional dispatch rules as the pre-execution hooks.

## Completion Report

Report:
- completed task IDs or task text
- files changed
- focused failing checks observed before fixes
- smoke verification run after behavior worked
- cleanup/docs tasks completed, if explicitly present
- any blocked task with the exact missing input
