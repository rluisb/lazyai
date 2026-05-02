# Task 013: W1.B Gate Report Integration Tests

**Phase:** W1.B — Plan Quality + Adversarial Design Review  
**User Story:** P3 — Review plans before implementation approval  
**Status:** TODO  
**Depends on:** T10, T12  
**Parallel with:** none

---

## Objective

Add final end-to-end W1.B validation proving the feature chain reaches human approval only after plan-quality and optional red-team reports are available, and that the gate displays a merged report without blocking on plan-quality fail verdicts or red-team soft failures.

## User Story / Spec References

- P3 — Human approvers see plan-quality and optional adversarial findings before implementation approval.
- AC-D6-001 — Inline plan-quality runs before approval and emits `PlanQualityReport`.
- AC-D6-002 — Approval gate displays pass/warn/fail verdict and findings.
- AC-D6-004 — A fail verdict proceeds to human approval, no automatic plan loop.
- AC-D6-006 — Finding locations use repo-relative file paths and 1-based lines when available.
- AC-D17-003 — Enabled red-team runs sequentially after plan quality and before approval.
- AC-D17-005 — Red-team provider/API failure emits `soft_fail` and gate still appears.
- AC-D17-006 — Plan-quality and red-team findings merge into one human-readable gate report.
- `spec.md` §Internal Contracts `MergedGateReport`.
- `plan.md` W1.B exit criterion.

## Files to Change/Create

- `packages/orchestrator/src/__tests__/chain-machine.case.ts` — end-to-end chain state progression/gate tests if the chain machine owns this behavior.
- `packages/orchestrator/src/__tests__/tool-handlers.case.ts` — gate/report surfacing tests if handler responses present gate data.
- `packages/ai-setup-go/internal/scaffold/orchestration_test.go` — installed source chain integration assertions as needed.
- `packages/ai-setup-ts/src/__tests__/orchestration.test.ts` — mirrored scaffold/install integration assertions as needed.
- New report merge tests/fixtures under appropriate package test locations, for example `packages/ai-setup-ts/src/__tests__/plan-gate-report.test.ts` if report shaping lives in TS.
- Existing snapshots/fixtures that represent gate prompt/report content.

## Files Not to Touch

- W1.A task files under `specs/features/001-ai-techniques-integration/tasks/001-*.md` through `007-*.md`
- `specs/features/001-ai-techniques-integration/spec.md`
- `specs/features/001-ai-techniques-integration/plan.md`
- MCP server configuration.
- Wave 2/3/4 roadmap docs or implementation.
- New standalone `plan-validate` skill or debate workflow files.

## Test-First Order

1. Add failing integration test for `adversarialDesign=false`: `plan → plan-quality → plan-gate`, with `PlanQualityReport` visible at gate.
2. Add failing integration test for plan-quality `fail`: chain enters `plan-gate` rather than auto-transitioning to `plan`.
3. Add failing integration test for `adversarialDesign=true`: `plan → plan-quality → red-team-plan → plan-gate`.
4. Add failing integration test simulating red-team provider/API outage: report has `status: "soft_fail"` and gate still appears.
5. Add failing report-merge test for `MergedGateReport` with summary counts, `planQuality`, and optional `adversarialReview`.
6. Add failing location assertions proving repo-relative paths and 1-based line numbers where available.
7. Only after the red tests fail, make the smallest integration/report wiring changes needed to pass.

## Done When

- [ ] D6-only generated/run chain reaches `plan-gate` with `PlanQualityReport` visible.
- [ ] Plan-quality `fail` verdict reaches human approval and does not automatically loop.
- [ ] D17-enabled generated/run chain includes `red-team-plan` sequentially between `plan-quality` and `plan-gate`.
- [ ] Red-team outage yields `status: "soft_fail"` and the gate still appears.
- [ ] `MergedGateReport` matches `schemaVersion: "plan-gate-report/v1"` and includes correct summary counts.
- [ ] Human-readable gate content includes plan-quality verdict/findings and optional red-team status/findings.
- [ ] No parallel block appears in final chain fixtures or generated installs.
- [ ] W1.B AC-D6-001..006 and AC-D17-001..006 are traceable to tests or documented verification evidence.

## Risks

- **Report artifacts not available to gate step:** mitigated by end-to-end chain/report tests rather than only unit tests.
- **Soft-fail accidentally treated as blocking failure:** mitigated by outage simulation and transition assertions.
- **Enabled/disabled chain drift:** mitigated by testing both `adversarialDesign` paths.
- **Human gate lacks actionable context:** mitigated by human-readable merged report assertions.

## Constitution Check

- **Article I:** Reuse existing orchestrator/scaffold integration harnesses and report structures.
- **Article II:** End-to-end integration and merge tests are written before final glue changes.
- **Article III:** The W1.B exit criterion and all AC IDs are traced from `spec.md` and `plan.md`.
- **Article IV:** Stop at W1.B D6/D17; do not add runtime guardrails, RAG, debate, or auto-recovery.
- **Article V:** Validate the simple sequential chain through existing harnesses.
- **Article VI:** No speculative abstractions; integration work only connects the tested W1.B contracts.