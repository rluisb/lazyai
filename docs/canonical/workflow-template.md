# Workflow Template

Workflow support is project-defined. Keep workflows deterministic, adapter-aware, and explicitly outside workflow-runtime ownership.

```markdown
---
name: <workflow-name>
description: <One sentence trigger and outcome.>
status: draft
---

# <Workflow Name>

## Trigger

<When this workflow starts.>

## Inputs

- <Required input.>

## Purpose

<Why this run exists and what outcome it serves.>

## Four Points

- WHAT: <outcome>
- HOW: <approach>
- DON'T WANT: <constraints>
- VALIDATE: <proof>

## Behavior Scenario

- Given <initial state>
- When <action>
- Then <observable result>

Use `n/a` only when the workflow is deliberately read-only or documentation-only.

## TDD Mode

<Required mode or `n/a` when no behavior is implemented.>

## Steps

1. <Skill, agent, hook, or command.>
2. <Next deterministic step.>
3. <Verification gate.>

## Adapters

- Claude Code: <skill / hook / agent mapping>
- OpenCode: <skill / plugin / agent mapping>
- OMP/Pi: <markdown / unsupported mapping>

## Exit Gate

Stop only when:

1. <purpose fulfilled or explicitly blocked>
2. <validation evidence observed>
3. <out-of-scope discoveries called out instead of silently absorbed>

## Failure

<What stops the workflow and what evidence to report.>
```

Do not add workflow runtime until `bin/inject` and `bin/doctor` validate the artifact type.
Hooks and plugins may enforce only objective checks they can actually observe; subjective completion stays advisory.
