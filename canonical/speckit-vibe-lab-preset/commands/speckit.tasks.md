---
description: Generate a dependency-ordered, test-first tasks.md that follows the vibe-lab task template.
handoffs:
  - label: Analyze For Consistency
    agent: speckit.analyze
    prompt: Run a project analysis for consistency
    send: true
  - label: Implement Project
    agent: speckit.implement
    prompt: Start the implementation in phases
    send: true
scripts:
  sh: scripts/bash/setup-tasks.sh --json
  ps: scripts/powershell/setup-tasks.ps1 -Json
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Pre-Execution Checks

Check `.specify/extensions.yml` for `hooks.before_tasks`.
- Skip silently if the file or hook list is missing.
- Ignore entries where `enabled: false`.
- Do not evaluate hook `condition` expressions.
- For executable mandatory hooks emit:
  - `EXECUTE_COMMAND: {command}`
- For executable optional hooks present them as optional follow-up commands.

## Outline

1. Run `{SCRIPT}` from repo root and parse `FEATURE_DIR`, `TASKS_TEMPLATE`, and `AVAILABLE_DOCS`.
2. Load `spec.md`, `plan.md`, `TASKS_TEMPLATE` when provided, and `/memory/constitution.md` if it exists.
3. Fill the vibe-lab tasks template.
4. Generate tasks using **dependency and test-first sequence**.
   - Start with the smallest red test or failing check.
   - Follow with the smallest source change.
   - Add refactor only when the focused check is green.
   - Keep docs/cleanup last, after smoke verification.
5. Every task must include an observable **acceptance check**.
6. For one-surface, low-risk work, use the small-change path instead of phase-heavy ceremony.
7. Preserve existing tests; never add tasks that weaken or delete them without explicit approval.
8. Write the completed task list to `FEATURE_DIR/tasks.md`.

## Post-Execution Hooks

Check `.specify/extensions.yml` for `hooks.after_tasks`.
- Use the same mandatory/optional dispatch rules as the pre-execution hooks.

## Completion Report

Report:
- `FEATURE_DIR/tasks.md`
- task count
- first red check
- dependency order summary
- suggested next command
