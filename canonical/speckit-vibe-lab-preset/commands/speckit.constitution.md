---
description: Create or update the project constitution as a four-point operating principles document.
handoffs:
  - label: Build Specification
    agent: speckit.specify
    prompt: Implement the feature specification based on the updated constitution. I want to build...
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Pre-Execution Checks

Check `.specify/extensions.yml` for `hooks.before_constitution`.
- Skip silently if the file or hook list is missing.
- Ignore entries where `enabled: false`.
- Do not evaluate hook `condition` expressions.
- For executable mandatory hooks emit:
  - `EXECUTE_COMMAND: {command}`
- For executable optional hooks present them as optional follow-up commands.

## Outline

1. Load `.specify/memory/constitution.md` if it exists.
   - If it does not exist, copy `.specify/templates/constitution-template.md` into that path first.
2. Load `.specify/templates/constitution-template.md` and treat it as the source shape.
3. Read nearby repo context only as needed to fill the constitution with concrete project rules.
4. Update the constitution as **four-point operating principles**:
   - `## WHAT` — what this project is trying to achieve.
   - `## HOW` — working rules, including risk-calibrated process and simplest viable changes.
   - `## DON'T WANT` — explicit anti-speculation and anti-scope-creep boundaries.
   - `## VALIDATE` — observable checks and truth requirements.
5. Keep the document concrete.
   - Do not turn it into a placeholder-token form.
   - Do not invent governance/versioning ceremony unless the project already asked for it.
   - Do not claim integrations or runtime behavior without observed verification.
6. If the constitution changes any project rule that affects `.specify/templates/*.md` or overridden `speckit.*` command files, update those artifacts to stay aligned.
7. Write the completed constitution back to `.specify/memory/constitution.md`.

## Post-Execution Hooks

Check `.specify/extensions.yml` for `hooks.after_constitution`.
- Use the same mandatory/optional dispatch rules as the pre-execution hooks.

## Completion Report

Report:
- Constitution path.
- Which four-point sections changed.
- Any dependent templates/commands updated to stay consistent.
- Any remaining `[NEEDS CLARIFICATION: ...]` items if the user left a material rule unresolved.
- Suggested next command.
