---
description: Clarify only material unknowns in a four-point specification before planning.
handoffs:
  - label: Build Technical Plan
    agent: speckit.plan
    prompt: Create a plan for the clarified spec. I am building with...
    send: true
scripts:
  sh: scripts/bash/check-prerequisites.sh --json --paths-only
  ps: scripts/powershell/check-prerequisites.ps1 -Json -PathsOnly
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Pre-Execution Checks

Check `.specify/extensions.yml` for `hooks.before_clarify`.
- Skip silently if the file or hook list is missing.
- Ignore entries where `enabled: false`.
- Do not evaluate hook `condition` expressions.
- For executable mandatory hooks emit:
  - `EXECUTE_COMMAND: {command}`
- For executable optional hooks present them as optional follow-up commands.

## Outline

1. Run `{SCRIPT}` from repo root and parse `FEATURE_DIR` and `FEATURE_SPEC`.
2. Load `FEATURE_SPEC` and `/memory/constitution.md` if it exists.
3. Scan the spec against **WHAT / WHY / HOW / DON'T WANT / VALIDATE** only.
   - `WHAT` must identify the user-visible goal or problem.
   - `WHY` must explain why it matters now.
   - `HOW` must stay at requirements and approach level, not unapproved implementation detail.
   - `DON'T WANT` must capture constraints, non-goals, unsupported-runtime boundaries, and prior failed approaches.
   - `VALIDATE` must name observable checks, commands, scenarios, or acceptance signals.
4. **Clarify only material unknowns**.
   - Ask only when the answer changes scope, implementation boundary, task ordering, risk, or validation.
   - Do not ask about style, nice-to-have preferences, or plan-level details that can be resolved from existing context.
   - Do not supply speculative defaults or best-practice recommendations for missing product or technical choices.
5. Ask at most five focused questions, one at a time.
   - Prefer short multiple-choice options only when the choices are genuinely distinct.
   - Allow a short free-form answer when options would invent choices the user did not imply.
   - If the user says to proceed, stop asking and preserve unresolved material items as `[NEEDS CLARIFICATION: specific question]`.
6. Write accepted answers into the four-point spec sections.
   - Add or update `## Clarifications` only when a question was actually answered.
   - Keep each clarification minimal, testable, and tied to one of the five sections.
   - Replace contradicted wording instead of appending conflicting notes.
   - Never expand scope while recording a clarification.
7. Leave non-blocking ambiguity explicit.
   - Use `[NEEDS CLARIFICATION: ...]` for material gaps the user did not answer.
   - Do not hide gaps behind assumptions.
   - Do not create or update separate checklist artifacts during clarify; requirement-quality review belongs to `speckit.checklist`.
8. Write the updated spec back to `FEATURE_SPEC`.

## Post-Execution Hooks

Check `.specify/extensions.yml` for `hooks.after_clarify`.
- Use the same mandatory/optional dispatch rules as the pre-execution hooks.

## Completion Report

Report:
- `FEATURE_SPEC`
- number of questions asked and answered
- four-point sections touched
- open `[NEEDS CLARIFICATION: ...]` items, if any
- validation impact: what changed in `VALIDATE` or what still blocks it
- suggested next command (`speckit.plan` when material ambiguity is resolved)
