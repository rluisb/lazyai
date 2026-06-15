---
description: Analyze four-point artifacts for consistency, speculative drift, unsupported runtime claims, and verification truth.
handoffs:
  - label: Implement Project
    agent: speckit.implement
    prompt: Start implementation only after resolving blocking analysis findings
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

Check `.specify/extensions.yml` for `hooks.before_analyze`.
- Skip silently if the file or hook list is missing.
- Ignore entries where `enabled: false`.
- Do not evaluate hook `condition` expressions.
- For executable mandatory hooks emit:
  - `EXECUTE_COMMAND: {command}`
- For executable optional hooks present them as optional follow-up commands.

## Scope

This command is a read-only consistency review. It reviews the current feature artifacts and reports findings; it does not rewrite the spec, plan, or tasks.

## Outline

1. Run `{SCRIPT}` from repo root and parse `FEATURE_DIR` and `AVAILABLE_DOCS`.
2. Load these artifacts when present:
   - `FEATURE_DIR/spec.md`
   - `FEATURE_DIR/plan.md`
   - `FEATURE_DIR/tasks.md`
   - `/memory/constitution.md` if it exists
3. Treat missing optional artifacts as context gaps, not as permission to invent requirements.
4. Analyze consistency across **WHAT / WHY / HOW / DON'T WANT / VALIDATE**.
   - `WHAT` must match the user-visible goal carried into the plan and tasks.
   - `WHY` must explain the motivation without becoming hidden scope.
   - `HOW` must stay aligned with the chosen approach and existing-pattern constraints.
   - `DON'T WANT` must be preserved as hard boundaries in the plan and task list.
   - `VALIDATE` must name observable checks that prove the requested behavior.
5. Report speculative drift.
   - Flag inferred features, APIs, integrations, runtime support, generic parity, extension behavior, hook behavior, workflow-runner behavior, or publication claims that are not grounded in the artifacts.
   - Flag tasks that implement outside `WHAT` or ignore `DON'T WANT`.
   - Flag plans that turn assumptions into requirements without explicit source text.
6. Report unsupported runtime claims.
   - Distinguish observed evidence from intended behavior.
   - Preserve the current boundary that Claude fixture generation is the only verified runtime seam unless the artifacts cite stronger evidence.
   - Do not claim generic command-directory parity, generated adapter parity, extension lifecycle behavior, or runtime ownership without explicit verification evidence.
7. Do **not** require legacy requirement, success, user-story, or entity inventories as inputs.
   - If those sections exist, use them only as supplementary context.
   - The required review frame is the five four-point sections plus plan/tasks alignment.
8. Produce an analysis report with this structure:
   - `## Summary`
   - `## Blocking Inconsistencies`
   - `## Speculative Drift`
   - `## Unsupported Runtime Claims`
   - `## Verification Gaps`
   - `## Non-blocking Notes`
   - `## Suggested Next Command`
9. Rank each finding as `BLOCKER`, `WARN`, or `INFO`.
10. For every `BLOCKER` or `WARN`, cite the artifact and section name that supports the finding. If no supporting artifact exists, state `Evidence: missing`.
11. Recommend `speckit.implement` only when there are no unresolved `BLOCKER` findings.

## Post-Execution Hooks

Check `.specify/extensions.yml` for `hooks.after_analyze`.
- Use the same mandatory/optional dispatch rules as the pre-execution hooks.

## Completion Report

Report:
- artifacts reviewed
- count of `BLOCKER`, `WARN`, and `INFO` findings
- the highest-risk speculative drift item, if any
- the highest-risk unsupported runtime claim, if any
- suggested next command
