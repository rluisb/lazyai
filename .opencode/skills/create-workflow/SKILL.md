---
name: create-workflow
description: Use when asked to create or revise a LazyAI workflow artifact as documentation/catalog guidance without reintroducing a runtime workflow engine.
---

# Create Workflow

## When to Use

Use this skill when the user asks to:
- define a repeatable multi-step process
- wire skills, agents, hooks, and verification gates into one documented flow
- add or revise a workflow in the shipped catalog
- make a process deterministic and reviewable

Do not use for one-off task lists. Do not use to smuggle back a runtime workflow/task/orchestration system.

## Rule

LazyAI workflow artifacts are documentation/catalog surfaces unless the user explicitly approves a supported runtime surface. In this repository the active catalog source lives under `packages/cli/library/workflows/`.

## Workflow

1. Confirm WHAT, HOW, DON'T WANT, and VALIDATE.
2. Name the purpose and exit gate before writing steps.
3. Choose markdown unless a stronger machine-readable contract is already approved.
4. Write or update the workflow source under `packages/cli/library/workflows/`.
5. Map each step to a concrete owner: skill, agent, hook/plugin, command, or human gate.
6. Mark unsupported adapter/runtime behavior explicitly instead of inventing it.
7. Verify the workflow remains docs-only unless a supported runtime surface is explicitly in scope.

## Constraints

- No workflow daemon, queue, or orchestration framework without explicit approval.
- No hidden side effects; every step names input, output, and failure mode.
- Hooks/plugins enforce only objective checks they can actually observe.
- Keep the workflow greppable and small.

## Verification Checklist

- [ ] Purpose and exit gate are explicit.
- [ ] Every step has a concrete owner.
- [ ] Unsupported behavior is documented instead of implied.
- [ ] The artifact does not reintroduce retired runtime workflow surfaces.
