---
description: Create or update the feature specification with four-point framing and explicit ambiguity handling.
handoffs:
  - label: Build Technical Plan
    agent: speckit.plan
    prompt: Create a plan for the spec. I am building with...
  - label: Clarify Spec Requirements
    agent: speckit.clarify
    prompt: Clarify specification requirements
    send: true
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Pre-Execution Checks

Check `.specify/extensions.yml` for `hooks.before_specify`.
- Skip silently if the file or hook list is missing.
- Ignore entries where `enabled: false`.
- Do not evaluate hook `condition` expressions.
- For executable mandatory hooks emit:
  - `EXECUTE_COMMAND: {command}`
- For executable optional hooks present them as optional follow-up commands.

## Outline

Given the feature description, do this:

1. Generate a concise 2-4 word short name.
2. Create the feature directory under `specs/` unless `SPECIFY_FEATURE_DIRECTORY` was explicitly provided.
3. Copy the resolved active `spec-template` into `SPECIFY_FEATURE_DIRECTORY/spec.md` and persist the resolved directory to `.specify/feature.json`.
4. Load the resolved `spec-template` and `/memory/constitution.md` if it exists.
5. Fill the specification using **WHAT / WHY / HOW / DON'T WANT / VALIDATE**.
   - `WHAT` states the user-visible problem or goal.
   - `WHY` states why this matters now.
   - `HOW` stays at requirements level only.
   - `DON'T WANT` captures constraints, non-goals, prior failures, and unsupported-runtime boundaries.
   - `VALIDATE` lists observable acceptance checks only.
6. **Do not silently guess** material product or implementation choices.
   - Do not guess tech-stack, API shape, file structure, integration support, or runtime claims unless the user already supplied them.
   - Use `[NEEDS CLARIFICATION: specific question]` for material unknowns.
   - Keep at most three unresolved clarification markers by impact; if more exist, keep the top three and move the rest into a short assumptions note with explicit risk.
7. Keep scope strict.
   - Anything outside `WHAT` belongs in `DON'T WANT`, not as an inferred requirement.
   - Do not add speculative features.
8. Write the completed spec back to `SPEC_FILE`.

## Post-Execution Hooks

Check `.specify/extensions.yml` for `hooks.after_specify`.
- Use the same mandatory/optional dispatch rules as the pre-execution hooks.

## Completion Report

Report:
- `SPECIFY_FEATURE_DIRECTORY`
- `SPEC_FILE`
- The top open `[NEEDS CLARIFICATION: ...]` items, if any.
- The observable validation path named in `VALIDATE`.
- Suggested next command (`speckit.plan` or `speckit.clarify`).
