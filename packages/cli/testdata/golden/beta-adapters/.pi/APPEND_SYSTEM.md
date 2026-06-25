# Pi Appended System Prompt (LazyAI managed)

This file is appended to Pi's built-in default system prompt when present at
`.pi/APPEND_SYSTEM.md`. Use it to add project-specific instructions without
replacing the default prompt — Pi still auto-discovers `AGENTS.md`/`CLAUDE.md`
context files, skills, and extensions.

## LazyAI conventions

- Canonical agents live in `.pi/agents/<name>.md`.
- Skills live in `.pi/skills/<name>/SKILL.md`.
- Prompt templates live in `.pi/prompts/<name>.md`.
- Safety hooks ship as TypeScript extensions at `.pi/extensions/*.ts`.

Edit the `.ai/` source tree and recompile — do not hand-edit these generated
files.