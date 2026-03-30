# Tasks: [Feature Name]

**Spec:** [Link to techspec]
**Total tasks:** [N]
**MVP tasks:** [T001-TNNN]

---

## Dependency Graph

```
Phase 1 (Sequential): T001 → T002 → T003
Phase 2 (Parallel):   T003 → [T004, T005] → T006
```

## Task List

| Task | Name | Phase | Depends | Status |
|------|------|-------|---------|--------|
| T001 | [Name] | 1 | none | TODO |
| T002 | [Name] | 1 | T001 | TODO |

## Execution Rules

- Complete tasks in dependency order
- Tasks marked parallel can run simultaneously
- Each task gets one commit
- Run quality gates after each phase
