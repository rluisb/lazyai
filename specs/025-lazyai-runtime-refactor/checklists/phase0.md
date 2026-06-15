# Checklist: Phase 0 Prerequisites and Approval

**Phase:** Phase 0  
**Exit:** All prerequisite artifacts exist, are linked, and have human-tracked approval evidence.

## Required Artifacts

- [ ] `specs/adrs/003-lazyai-runtime-boundary.md` is tracked.
- [ ] `specs/025-lazyai-runtime-refactor/spec.md` is tracked.
- [ ] `specs/025-lazyai-runtime-refactor/plan.md` is tracked.
- [ ] `specs/025-lazyai-runtime-refactor/research.md` is tracked.
- [ ] `specs/025-lazyai-runtime-refactor/audit-cli-imports.md` is tracked.
- [ ] `specs/025-lazyai-runtime-refactor/survey-usage.md` is tracked.
- [ ] `specs/025-lazyai-runtime-refactor/schema-v2.md` is tracked.
- [ ] `specs/025-lazyai-runtime-refactor/schema-handoff.md` is tracked.
- [ ] `specs/025-lazyai-runtime-refactor/library-canonical.md` is tracked.
- [ ] `specs/025-lazyai-runtime-refactor/contract-adapter.md` is tracked.
- [ ] `specs/025-lazyai-runtime-refactor/design-token-rent.md` is tracked.
- [ ] `specs/025-lazyai-runtime-refactor/rollback.md` is tracked.

## Post-Approval Generated Artifacts

- [ ] `specs/025-lazyai-runtime-refactor/tasks.md` exists.
- [ ] `specs/025-lazyai-runtime-refactor/analysis.md` exists.
- [ ] `specs/025-lazyai-runtime-refactor/checklists/phase0.md` exists.
- [ ] `specs/025-lazyai-runtime-refactor/checklists/phase1.md` exists.
- [ ] `specs/025-lazyai-runtime-refactor/checklists/phase2.md` exists.
- [ ] `specs/025-lazyai-runtime-refactor/checklists/phase3.md` exists.
- [ ] `specs/025-lazyai-runtime-refactor/checklists/phase4.md` exists.
- [ ] `specs/025-lazyai-runtime-refactor/checklists/phase5.md` exists.

## Approval Evidence

- [ ] Human-authored approval commit is recorded.
- [ ] Approval commit includes ADR/spec/plan/P0 artifacts.
- [ ] Phase 0 approval does not claim Phase 1-5 code verification.
- [ ] Any pending P0 caveat is carried forward into later phase gates.

## Known Carry-Forward Caveats

- [ ] Qualitative Fortnite/OpenCode outreach is complete or explicitly waived before Phase 2.
- [ ] Missing design-system CSS fixture is repaired before CLI-wide test evidence is used.
- [ ] `update-self --version <tag>` rollback readiness is fixed before destructive operator rollback is claimed.

## Gate

- [ ] Human approves advancing to Phase 1 code work after reviewing `tasks.md`, `analysis.md`, and checklists.
