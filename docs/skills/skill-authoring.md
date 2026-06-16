---
name: skill-authoring
description: Use when creating or modifying vibe-lab skills or adjacent artifact templates so Agent Skills compatibility, canonical source layout, adapter generation, and verification stay consistent.
---

# Skill Authoring

## When to Use

Use this skill when creating, renaming, or modifying a skill under `packages/cli/library/skills/`, or when updating the templates that skills depend on.

For agents, hooks, policies, and workflows, use the canonical templates in `canonical/` first; use `create-agent`, `create-hook`, or `create-workflow` when scaffolding a new artifact.

## Rule

Skills are progressive-disclosure artifacts: keep `SKILL.md` concise, move optional detail to `references/`, executable repeatable logic to `scripts/`, and reusable templates to `assets/`.

## Canonical Templates

- `canonical/skill-template.md` — Agent Skills compatible skill shape.
- `canonical/agent-template.md` — canonical agent shape.
- `canonical/hook-template.md` — hook policy plus Claude/OpenCode runtime mapping.
- `canonical/policy-template.md` — markdown rule/policy shape.
- `canonical/workflow-template.md` — deterministic workflow shape.

## Required Shape

```text
packages/cli/library/skills/<skill>/
├── SKILL.md       # required
├── scripts/       # optional
├── references/    # optional
└── assets/        # optional
```

`SKILL.md` frontmatter must include `name` and `description`. The `name` value must match the directory name.

## Verification Loop

After changing a skill or template:

1. Run `lazyai-cli compile`.
2. Run `lazyai-cli doctor`.
3. Run targeted tests for any affected adapter output.
4. Confirm generated catalogs describe the artifact clearly.

## Constraints

- Do not duplicate canonical files under CLI-specific directories.
- Do not add helper scripts unless they remove fragile repeated work.
- Do not add a second memory, sandbox, workflow, or adapter system.
- Do not broaden a skill description until trigger behavior is tested or reviewed.
