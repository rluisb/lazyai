---
description: Generate a requirement-quality checklist shaped by anti-speculation and verification truth.
scripts:
  sh: scripts/bash/check-prerequisites.sh --json
  ps: scripts/powershell/check-prerequisites.ps1 -Json
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Pre-Execution Checks

Check `.specify/extensions.yml` for `hooks.before_checklist`.
- Skip silently if the file or hook list is missing.
- Ignore entries where `enabled: false`.
- Do not evaluate hook `condition` expressions.
- For executable mandatory hooks emit:
  - `EXECUTE_COMMAND: {command}`
- For executable optional hooks present them as optional follow-up commands.

## Outline

1. Run `{SCRIPT}` from repo root and parse `FEATURE_DIR` and `AVAILABLE_DOCS`.
2. Load `spec.md`, `plan.md`, `tasks.md` when present, `.specify/templates/checklist-template.md`, and `/memory/constitution.md` if it exists.
3. Generate a checklist file under `FEATURE_DIR/checklists/requirements.md` unless the user explicitly asked for a different domain name.
4. Build the checklist around requirement quality, not implementation QA.
5. The checklist must reinforce these vibe-lab concerns:
   - ambiguity is explicit, not hidden
   - anti-speculation: no inferred features beyond the request
   - unsupported runtime claims are called out
   - verification truth: every important claim names a verified command, test, or scenario
6. Reuse the shipped checklist template headings and add focused checkbox items only.
7. Do not silently invent risk areas. If scope is unclear, ask a focused follow-up before finalizing the checklist.
8. Write the checklist file.

## Post-Execution Hooks

Check `.specify/extensions.yml` for `hooks.after_checklist`.
- Use the same mandatory/optional dispatch rules as the pre-execution hooks.

## Completion Report

Report:
- checklist path
- which requirement-quality areas were covered
- any unresolved ambiguity still blocking a trustworthy checklist
- suggested next command
