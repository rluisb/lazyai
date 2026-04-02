# Plan Prompt

**Feature:** [Feature Name]
**PRD:** [Link to PRD]
**Tech Spec:** [Link to Tech Spec]

---

## Instructions

0. **Think Step-by-Step (CoT):** Privately reason step-by-step before writing the final plan, but only output concise conclusions.
1. **Translate Requirements:** Convert the PRD and Tech Spec into a phased task list.
2. **Identify Dependencies:** Explicitly state dependencies for each task.
3. **Define Done Criteria:** Define "done when" criteria for each task.
4. **Size Tasks Correctly:** Each task should be completable in one focused session.
5. **Mark Parallelism:** Identify tasks that can run in parallel.

## Few-Shot Mini Example (Generic)

Use this pattern as a guide:

```
Input (summary): Build password reset flow with email token.
Output (shape):
- Phase 1 (Sequential): token generation + persistence
- Phase 2 (Parallel): API endpoint + email template
- Phase 3 (Sequential): integration tests + rollout checks
```

## Common Mistakes to Avoid
- ❌ Planning implementation details before understanding requirements
- ❌ Skipping the "risks" or "unknowns" section
- ❌ Creating plans that modify files outside the stated scope

## Output Format

```
## Plan: [Feature Name]

### Phases
**Phase N: [Name]** — [Sequential | Parallel]
- T001: [Task name] (depends: none)
  - Done when: [criteria]
- T002: [Task name] (depends: T001)
  - Done when: [criteria]

### Risks
- [Risk and mitigation]
```
