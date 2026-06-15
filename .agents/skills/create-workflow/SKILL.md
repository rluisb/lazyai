---
name: create-workflow
description: Use when asked to create or design a vibe-lab workflow artifact that coordinates skills, agents, hooks, plugins, and verification gates across Claude Code, OpenCode, and OMP/Pi.
---

# Create Workflow

## When to Use

Use this skill when the user asks to:
- "Create a workflow for X"
- "Wire these skills/hooks/agents together"
- "Define a repeatable implementation flow"
- "Make this process deterministic"

Do not use for one-off task lists. Use a workflow only when multiple artifacts or CLI adapters need a shared contract.

## Rule

A workflow is a deterministic artifact, not a runtime framework. Do not add workflow runtime until `bin/inject` and `bin/doctor` validate the artifact type.

## Template

Use `canonical/workflow-template.md`.

## Workflow

1. Confirm the four points: WHAT, HOW, DON'T WANT, VALIDATE.
2. Name the workflow Purpose and the concrete Exit Gate before writing steps.
3. Choose whether the workflow is markdown-only or needs YAML structure.
4. Write `.agents/workflows/<name>.md` for human/agent workflow or `.agents/workflows/<name>.yml` for machine-readable workflow.
5. Include a Behavior Scenario unless the workflow is deliberately read-only/documentation-only; mark `n/a` explicitly when skipped.
6. Record a TDD mode or `n/a` depending on whether the workflow changes behavior.
7. Map each step to a skill, agent, hook/plugin, command, or verification gate.
8. Document adapter behavior for Claude Code, OpenCode, and OMP/Pi.
9. Mark unsupported adapter behavior explicitly.
10. Run `bin/inject`, `bin/doctor`, and `tests/test-provenance-drift.sh` after workflow support is wired.

## Constraints

- No workflow daemon, queue, or orchestration framework unless explicitly approved.
- No hidden side effects. Every step names its input, output, and failure condition.
- Hooks/plugins enforce only objective checks they can actually observe; workflows describe the deterministic sequence.
- Keep workflows greppable and small.

## Verification Checklist

- [ ] Workflow includes Four Points.
- [ ] Workflow names its Purpose and Exit Gate.
- [ ] Behavior Scenario is present or explicitly `n/a`.
- [ ] TDD mode is named or explicitly `n/a`.
- [ ] Every step has a concrete owner: skill, agent, hook/plugin, command, or human gate.
- [ ] Adapter support is documented.
- [ ] Failure behavior is explicit.
