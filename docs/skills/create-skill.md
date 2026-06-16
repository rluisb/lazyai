---
name: create-skill
description: Use when asked to create, scaffold, or write a new vibe-lab Agent Skill. Generates an Agent Skills compatible SKILL.md with optional scripts, references, assets, adapter symlinks, and verification.
---

# Create Skill

## When to Use

Use this skill when the user asks to:
- "Create a skill for X"
- "Scaffold a new skill"
- "Write a SKILL.md for Y"
- "Add a skill that does Z"

Do not use when the user is only discussing a skill idea. Use `skill-authoring` when modifying an existing skill.

## Rule

A skill is `.agents/skills/<name>/SKILL.md`; optional `scripts/`, `references/`, and `assets/` are allowed only when they reduce repeated agent work.

## Template

Use `canonical/skill-template.md`.

Minimum shape:

```text
.agents/skills/<name>/
├── SKILL.md
├── scripts/      # optional executable helpers
├── references/   # optional on-demand docs
└── assets/       # optional templates/resources
```

## Workflow

1. Confirm the four points: WHAT, HOW, DON'T WANT, VALIDATE.
2. Choose a kebab-case name matching Agent Skills rules: lowercase letters, numbers, hyphens; no leading/trailing/consecutive hyphen.
3. Write `SKILL.md` from `canonical/skill-template.md` with `name` and `description` frontmatter.
4. If the skill needs reusable executable logic, add a non-interactive script under `scripts/` and document usage plus dependencies.
5. If the skill needs long detail, move it to `references/` and say exactly when to read it.
6. Run `lazyai-cli compile` to regenerate adapters and generated docs.
7. Run `lazyai-cli doctor` to verify output consistency.

## Script Rules

- Prefer single-file markdown. Add scripts only when the agent would otherwise repeat fragile logic.
- Supported script languages: Bash, TypeScript/JavaScript, Python, Ruby.
- Scripts must accept flags/stdin/env, never interactive prompts.
- Scripts must provide `--help`, useful error messages, and structured output when practical.
- Shared utilities belong in `bin/` or `scripts/`, not inside one skill.

## Verification Checklist

- [ ] `SKILL.md` has valid `name` and `description` frontmatter.
- [ ] `name` matches the directory and Agent Skills naming constraints.
- [ ] Description is trigger-specific and under 1024 characters.
- [ ] Optional files are referenced by relative path from the skill root.
- [ ] `bin/inject`, `bin/doctor`, and `tests/test-provenance-drift.sh` pass.
