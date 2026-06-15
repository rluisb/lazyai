---
description: Create an implementation plan that follows the vibe-lab plan template and explicit TDD mode selection.
handoffs:
  - label: Create Tasks
    agent: speckit.tasks
    prompt: Break the plan into tasks
    send: true
  - label: Create Checklist
    agent: speckit.checklist
    prompt: Create a checklist for the following domain...
scripts:
  sh: scripts/bash/setup-plan.sh --json
  ps: scripts/powershell/setup-plan.ps1 -Json
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Pre-Execution Checks

Check `.specify/extensions.yml` for `hooks.before_plan`.
- Skip silently if the file or hook list is missing.
- Ignore entries where `enabled: false`.
- Do not evaluate hook `condition` expressions.
- For executable mandatory hooks emit:
  - `EXECUTE_COMMAND: {command}`
- For executable optional hooks present them as optional follow-up commands.

## Outline

1. Run `{SCRIPT}` from repo root and parse JSON for `FEATURE_SPEC`, `IMPL_PLAN`, `SPECS_DIR`, and `BRANCH`.
2. Load `FEATURE_SPEC`, `.specify/templates/plan-template.md`, and `/memory/constitution.md` if it exists.
3. Fill the plan template with the smallest concrete plan that satisfies the spec.
4. In `## HOW`, make these sections explicit:
   - simplest viable approach
   - existing pattern review
   - rejected alternatives
   - boundaries
   - risk
5. Add a `## TDD Plan` section even if the template did not include one.
   - Choose a **TDD mode**: `lightweight`, `medium`, `heavy-aggressive`, or `required`.
   - Name the red test or focused failing check.
   - Name the green verification command.
   - State which existing tests must be preserved.
6. Keep the plan boring.
   - No new dependency, runtime state, hook, generated adapter, or extra artifact unless the spec forces it now.
   - Do not create `research.md`, `data-model.md`, `contracts/`, or `quickstart.md` unless the spec explicitly requires them.
7. Write the plan back to `IMPL_PLAN`.

## Post-Execution Hooks

Check `.specify/extensions.yml` for `hooks.after_plan`.
- Use the same mandatory/optional dispatch rules as the pre-execution hooks.

## Completion Report

Report:
- `IMPL_PLAN`
- chosen TDD mode
- simplest viable approach summary
- explicit out-of-scope and deferred items
- suggested next command
