---
name: four-point-vibe-coding
description: Use when starting or steering an agent task with four-point communication, project constitution, and fast feedback.
---

# Four-Point Vibe Coding

## When to Use

Use this skill when starting a new task or context window in any CLI agent.

## How to Use

1. Read the root instruction file for the setup (`AGENTS.md`, `CLAUDE.md`, or the tool-equivalent generated root instructions).
2. Read `.specify/memory/constitution.md` when it exists.
3. Confirm the four points: WHAT, HOW, DON'T WANT, VALIDATE.
4. Ask before coding if any point is missing.
5. Choose the smallest verification step that proves the latest change.

## Files

- Root instructions for the active tool
- `.specify/memory/constitution.md` when present
- The nearest spec, task, or rules file that governs the requested change

## Notes

- CLI-agnostic.
- Keep prompts short; lean on the governing files for detail.
- Do not replace explicit project rules with personal preference.
