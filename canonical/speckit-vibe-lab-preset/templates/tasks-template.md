# Tasks: {{task_list_name}}

## WHAT

One-sentence goal. Link to the spec/plan this task list serves.

## HOW

Order work by dependency and test-first sequence. Each task must have an observable acceptance check.

For one-surface, low-risk work, use the small-change path: one red check, one smallest change, one focused verification.

| Order | Task | Acceptance Check |
|-------|------|-----------------|
| 1 | Write red test before source change | Test fails with expected error |
| 2 | Make smallest source change | Test passes |
| 3 | Refactor if needed | Tests still pass; no behavior change |
| 4 | Add docs/cleanup only after smoke verification | Docs match current behavior |

## EXIT GATE

- [ ] Purpose is fulfilled or explicitly blocked.
- [ ] Validation evidence is observed.
- [ ] Out-of-scope discoveries are reported, not silently absorbed.

## DON'T WANT

- Tasks without observable acceptance checks.
- Cleanup or docs before the behavior is proven.
- Scope creep: anything not in the plan is a new spec.

## VALIDATE

- [ ] Every task has an acceptance check.
- [ ] Tasks are ordered by dependency.
- [ ] Tests are written before source changes (test-first).
- [ ] Existing tests are preserved.
- [ ] Cleanup/docs are last, not first.
- [ ] One-surface, low-risk work uses the small-change path instead of full ceremony.
