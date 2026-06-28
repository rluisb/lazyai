---
name: create-agent
description: Use when asked to create or revise a LazyAI agent source definition while keeping generated tool-native agent files derived, not hand-edited.
---

# Create Agent

## When to Use

Use this skill when the user asks to:
- "Create an agent for X"
- "Scaffold a new agent"
- "Write an agent definition for Y"
- "Add an agent that does Z"

Do not use when the user is merely discussing an agent idea. Do not use when the behavior is better expressed as a skill.

## Rule

A canonical agent is one markdown file under `packages/cli/library/canonical/agents/<name>.md`. Host-specific model/provider settings do not belong in canonical agent frontmatter.

## Template

Use `packages/cli/library/canonical/agents/` existing agent files as the shape reference.

## Workflow

1. Confirm the four points: WHAT, HOW, DON'T WANT, VALIDATE.
2. Choose a kebab-case name matching existing agent naming.
3. Write `packages/cli/library/canonical/agents/<name>.md` with `name`, `description`, and markdown system prompt sections.
4. Define when to delegate to the agent and what it must never do.
5. Keep tool needs minimal and explicit; do not create a general-purpose wrapper.
6. Run `lazyai-cli compile` to regenerate tool-native agent files.
7. Run `lazyai-cli doctor` to verify output consistency.

## Constraints

- One agent per focused change.
- Do not add `model`, provider, color, or host-only settings to canonical frontmatter.
- Do not duplicate skills inside an agent prompt; reference skill names when useful.
- OMP/Pi agent adapter support is documented as limited unless `lazyai-cli doctor` verifies it.

## Verification Checklist

- [ ] Canonical file is `packages/cli/library/canonical/agents/<name>.md`.
- [ ] `name` and filename match.
- [ ] `description` is one sentence and trigger-specific.
- [ ] Role has explicit non-goals.
- [ ] Generated Claude and OpenCode adapters exist after `lazyai-cli compile`.
