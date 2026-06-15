---
name: create-agent
description: Use when asked to create, scaffold, or write a new vibe-lab agent definition. Generates one canonical .agents/agents/<name>.md file and lets bin/inject produce CLI adapters.
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

A canonical agent is exactly one markdown file: `.agents/agents/<name>.md`. Host-specific model/provider settings do not belong in canonical agent frontmatter.

## Template

Use `canonical/agent-template.md`.

## Workflow

1. Confirm the four points: WHAT, HOW, DON'T WANT, VALIDATE.
2. Choose a kebab-case name.
3. Write `.agents/agents/<name>.md` with `name`, `description`, `role`, `mode`, and markdown system prompt sections.
4. Define when to delegate to the agent and what it must never do.
5. Keep tool needs minimal and explicit; do not create a general-purpose wrapper.
6. Run `bin/inject` to generate `.claude/agents/<name>.md` and `.opencode/agents/<name>.md`.
7. Run `bin/doctor` and `tests/test-provenance-drift.sh`.

## Constraints

- One agent per focused change.
- Do not add `model`, provider, color, or host-only settings to canonical frontmatter.
- Do not duplicate skills inside an agent prompt; reference skill names when useful.
- OMP/Pi agent adapter support is documented as limited unless `bin/doctor` verifies it.

## Verification Checklist

- [ ] Canonical file is `.agents/agents/<name>.md`.
- [ ] `name` and filename match.
- [ ] `description` is one sentence and trigger-specific.
- [ ] Role has explicit non-goals.
- [ ] Generated Claude and OpenCode adapters exist after `bin/inject`.
