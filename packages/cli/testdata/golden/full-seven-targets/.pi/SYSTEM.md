# Pi System Prompt (LazyAI managed)

This file replaces Pi's built-in default system prompt when present at
`.pi/SYSTEM.md`. Use it to fully override the default prompt with
project-specific instructions.

> **Warning:** replacing the default prompt disables Pi's automatic loading of
> context files (`AGENTS.md`, `CLAUDE.md`), skills, and extensions discovered
> from the project. Prefer `APPEND_SYSTEM.md` when you want to *add*
> instructions without losing the default behavior.

## LazyAI conventions

- Canonical agents live in `.pi/agents/<name>.md`.
- Skills live in `.pi/skills/<name>/SKILL.md`.
- Prompt templates live in `.pi/prompts/<name>.md`.
- Safety hooks ship as TypeScript extensions at `.pi/extensions/*.ts`.

Edit the `.ai/` source tree and recompile — do not hand-edit these generated
files.