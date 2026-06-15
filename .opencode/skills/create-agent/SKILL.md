---
name: create-agent
description: Use when asked to create or revise a LazyAI agent source definition while keeping generated tool-native agent files derived, not hand-edited.
---

# Create Agent

## When to Use

Use this skill when the user asks to:
- create a new agent
- rewrite an existing agent role
- split one broad agent into focused roles
- add a reusable agent definition to a managed source library

Do not use when the user is only discussing an idea. Do not use when a skill or plain rule is enough.

## Rule

A shipped LazyAI agent has one source of truth. In this repository that source is `packages/cli/library/canonical/agents/<name>.md`. In a consuming setup, edit the canonical source your project compiles from and never hand-edit generated `.opencode/`, `.claude/`, or `.github/` agent files.

## Workflow

1. Confirm WHAT, HOW, DON'T WANT, and VALIDATE.
2. Choose a kebab-case name.
3. Write or update the canonical source file.
4. Keep role, boundaries, and failure modes explicit.
5. Keep tool/provider specifics in generated surfaces, not in the neutral source, unless the target format requires generated frontmatter.
6. Recompile affected tool surfaces and inspect the generated agent files.
7. Run `lazyai-cli compile` plus `lazyai-cli doctor` in a consuming project, or `go test ./packages/cli/...` plus `go build ./packages/cli/...` when changing LazyAI's embedded library or adapters.

## Constraints

- One agent, one focused role.
- Do not hand-edit generated tool-native agent files.
- Do not reintroduce retired orchestrator/Fortnite roles or workflow-runtime assumptions.
- Pi has no agent surface.

## Verification Checklist

- [ ] The edited file is canonical source, not generated output.
- [ ] `name` and filename match.
- [ ] Role, non-goals, and escalation conditions are explicit.
- [ ] Generated OpenCode, Claude Code, and Copilot agent files were refreshed when those tools are in scope.
