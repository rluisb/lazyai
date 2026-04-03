---
name: plan
description: Create structured implementation plan with TechSpec.
trigger: /plan
phase: plan
---

# Plan Skill

**Command:** `/plan [research-file]`
**Goal:** Convert research or PRD input into a phased implementation plan.

---

## Workflow

1. **Read Research/PRD**
   - Extract objectives, constraints, requirements, and open decisions.
2. **Break Into Phases**
   - Organize work into logical, incremental phases.
3. **Define Tasks Per Phase**
   - Write concrete, bounded tasks with clear expected outcomes.
4. **Identify Dependencies**
   - Capture sequencing, shared prerequisites, and parallel opportunities.
5. **Add Checkpoints and Human Gates**
   - Insert review points for risky changes and decision handoffs.

## Output Format

```markdown
## Implementation Plan: [Feature]

### Inputs
- Research/PRD: [file]

### Phase 1: [Name]
- Task P1-T1: [task]
  - Depends on: [none | task IDs]
  - Acceptance Criteria: [verifiable outcome]
- Checkpoint: [review or gate]

### Phase 2: [Name]
- Task P2-T1: [task]
  - Depends on: [task IDs]
  - Acceptance Criteria: [verifiable outcome]
- Checkpoint: [review or gate]

### Cross-Phase Dependencies
- [dependency mapping]

### Risks and Mitigations
- [risk] — [mitigation]
```

## Integration
- **Primary agent:** Planner (planning phase)
- **Triggered by:** `/plan` command or workflow rule step 2
- **Depends on:** Research output from Scout phase (research.md)
- **Feeds into:** implement.md (Builder reads plan output); parallel-execution.md (for multi-file tasks)
- **Related skills:** research (prior phase), anti-speculation (scope guard during planning)
