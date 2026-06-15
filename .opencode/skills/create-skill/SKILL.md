---
name: create-skill
description: Use when asked to create or revise a LazyAI skill source file while keeping generated tool-native skill surfaces derived, not hand-edited.
---

# Create Skill

## When to Use

Use this skill when the user asks to:
- create a new reusable skill
- rewrite an existing skill's trigger or workflow
- add a skill to a managed source library
- turn repeated guidance into one canonical skill artifact

Do not use when the user is only brainstorming. Use `skill-authoring` when modifying an existing skill in place.

## Rule

A shipped LazyAI skill has one source of truth. In this repository that source is `packages/cli/library/skills/<name>.md`. Generated `.opencode/skills/<name>/SKILL.md`, `.claude/skills/<name>/SKILL.md`, `.pi/skills/<name>/SKILL.md`, and Copilot `.agent.yaml` files are outputs, not authoring locations.

## Workflow

1. Confirm WHAT, HOW, DON'T WANT, and VALIDATE.
2. Choose a kebab-case name.
3. Write or update the canonical skill source markdown with `name` and `description` frontmatter.
4. Keep the source concise and operational: trigger, workflow, constraints, and verification.
5. Prefer one markdown file unless the source library already has an approved multi-file pattern.
6. Recompile affected tool surfaces and inspect generated outputs.
7. Run `lazyai-cli compile` plus `lazyai-cli doctor` in a consuming project, or `go test ./packages/cli/...` plus `go build ./packages/cli/...` when changing LazyAI's embedded skill library or adapters.

## Constraints

- One skill, one clear trigger.
- Do not hand-edit generated tool-native skill files.
- Do not reintroduce retired workflow/task/orchestrator runtime assumptions.
- Keep optional detail proportional; do not hide the main workflow in deep references.

## Verification Checklist

- [ ] The edited file is canonical source, not generated output.
- [ ] `name` matches the filename.
- [ ] Description is trigger-specific.
- [ ] Generated OpenCode/Claude/Pi/Copilot surfaces were refreshed when those tools are in scope.
