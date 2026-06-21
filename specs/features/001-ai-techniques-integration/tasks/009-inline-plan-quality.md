> **HISTORICAL (2026-06-20) — parts of this plan target the removed `packages/ai-setup-ts` package; the CLI is now Go-only (`packages/cli`). Retained for historical context.**
>
> This document is historical; `packages/ai-setup-ts` is no longer present.

# Task 009: Inline Planner Quality Checklist + Report Tests

**Phase:** W1.B — Plan Quality + Adversarial Design Review  
**User Story:** P3 — Review plans before implementation approval  
**Status:** TODO  
**Depends on:** T8  
**Parallel with:** none

---

## Objective

Add the D6 plan-quality contract as an inline checklist in the existing planner skill and test the `PlanQualityReport` JSON contract. Do not create a standalone `plan-validate` skill or reusable validation framework.

## User Story / Spec References

- P3 — Human approvers see plan-quality findings before implementation approval.
- AC-D6-001 — Plan quality runs after a feature plan is produced and emits `PlanQualityReport` JSON.
- AC-D6-003 — Malformed Markdown or ambiguous parsing produces `warn`, not `fail`.
- AC-D6-004 — A `fail` verdict proceeds to the human approval gate; no automatic plan loop.
- AC-D6-005 — R1 checks spec AC/FR coverage in `plan.md` phase/item/downstream-contract sections only; it does not require `tasks.md`.
- AC-D6-006 — Findings include `{file, section, lineStart, lineEnd}` with repo-relative file paths and 1-based line numbers when available.
- FR-008 through FR-010.
- `spec.md` §Internal Contracts `PlanQualityReport`.
- `plan.md` §D6 — Plan Validation.

## Files to Change/Create

- `packages/ai-setup-go/library/skills/plan.md` — add the bounded inline “Plan Quality Check” section and JSON output contract.
- Existing skill/frontmatter tests, likely `packages/ai-setup-ts/src/__tests__/frontmatter.test.ts`, if the planner skill metadata is validated there.
- New or existing tests/fixtures for report-contract validation under appropriate package test locations, for example:
  - `packages/ai-setup-ts/src/__tests__/plan-quality-report.test.ts`
  - `packages/orchestrator/src/__tests__/chain-machine.case.ts` only if needed to assert non-looping outcomes at the chain layer.
- Fixture Markdown files for pass/warn/fail cases if the package already uses fixture-based prompt/report tests.

## Files Not to Touch

- `packages/ai-setup-go/library/skills/plan-validate.md` — do not create this file.
- `packages/ai-setup-go/library/orchestration/chains/feature.json` — T10 owns chain edits.
- `packages/ai-setup-go/library/skills/red-team-plan.md` — T11 owns this skill.
- `packages/ai-setup-ts/src/presets.ts`
- `packages/ai-setup-ts/src/wizard/phase2-features.ts`
- W1.A task files under `specs/features/001-ai-techniques-integration/tasks/001-*.md` through `007-*.md`
- `specs/features/001-ai-techniques-integration/spec.md`
- `specs/features/001-ai-techniques-integration/plan.md`

## Test-First Order

1. Add failing report-schema tests for `PlanQualityReport` with `schemaVersion: "plan-quality-report/v1"`, `verdict`, `findings`, `location`, and `checkedAgainst.tasks: null`.
2. Add failing fixture tests for R1–R4:
   - R1 spec AC/FR coverage only in allowed `plan.md` sections.
   - R2 phase exit criteria present.
   - R3 each risk has mitigation and owner.
   - R4 each Wave 1 item has rollback coverage.
3. Add failing fixture tests proving malformed/ambiguous Markdown produces `warn` findings rather than parser-driven `fail` findings.
4. Add failing test or snapshot proving the planner skill instructs the agent to emit JSON and always proceed to gate recommendation instead of looping.
5. Only after the red tests fail, edit `packages/ai-setup-go/library/skills/plan.md` with the minimal inline checklist and contract instructions.

## Done When

- [ ] The planner skill contains an inline “Plan Quality Check” section; no standalone validator skill exists.
- [ ] `PlanQualityReport` schema is covered by tests or snapshots.
- [ ] R1 explicitly excludes `tasks.md` because tasks do not exist when plan quality runs.
- [ ] Parser uncertainty/malformed Markdown maps to `warn`, never automatic `fail`.
- [ ] Findings use repo-relative file paths and 1-based line numbers when available.
- [ ] The skill text states all verdicts proceed to the human gate; `fail` is a recommendation, not an automatic loop.
- [ ] Test evidence and commands are captured.

## Risks

- **False-positive plan failures:** mitigated by conservative rules and parser uncertainty downgrading to `warn`.
- **Scope creep into a validation framework:** mitigated by inlining D6 into the existing planner skill only.
- **Incorrect task coverage expectation:** mitigated by explicitly setting `checkedAgainst.tasks` to `null` and excluding `tasks.md` from R1.

## Constitution Check

- **Article I:** Reuse the existing planner skill and test harnesses.
- **Article II:** Report schema and rule tests are written before skill edits.
- **Article III:** Rules and schema come directly from `spec.md`; implementation approach comes from `plan.md`.
- **Article IV:** No standalone `plan-validate` skill, no semantic parser dependency, and no Wave 2 validation features.
- **Article V:** Concrete R1–R4 checklist is simpler than a general validator.
- **Article VI:** Avoid one-caller abstractions and broad helper extraction.