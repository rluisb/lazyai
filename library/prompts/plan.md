# Plan Prompt

**Feature:** [Feature Name]
**PRD:** [Link to PRD]
**Tech Spec:** [Link to Tech Spec]

---

## Instructions

1. **Translate Requirements:** Convert the PRD and Tech Spec into a phased task list.
2. **Identify Dependencies:** Explicitly state dependencies for each task.
3. **Define Done Criteria:** Define "done when" criteria for each task.
4. **Size Tasks Correctly:** Each task should be completable in one focused session.
5. **Mark Parallelism:** Identify tasks that can run in parallel.

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
