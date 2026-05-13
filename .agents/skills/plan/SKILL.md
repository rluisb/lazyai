---
name: plan
description: Plan implementation approach before writing code.
argument-hint: "[feature-or-task]"
trigger: /plan
phase: plan
---

# Plan Skill

## Workflow
1. Read research-rpi output — confirm understanding of scope and findings
2. Define acceptance criteria — what does "done" look like?
3. Break into phases — group related changes, order by dependency
4. Define tasks — one task per file, each with "Done When" criteria
5. Identify risks — what could go wrong, how to mitigate
6. Produce plan — write plan.md and task files

## Integration
- Agent: Planner
- Requires: research-rpi output (research.md)
- Output feeds into: `implement` skill
- Output location: specs/features/NNN-name/plan.md + tasks/
