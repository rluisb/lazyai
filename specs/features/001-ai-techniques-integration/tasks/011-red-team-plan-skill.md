> **HISTORICAL (2026-06-20) — parts of this plan target the removed `packages/ai-setup-ts` package; the CLI is now Go-only (`packages/cli`). Retained for historical context.**
>
> This document is historical; `packages/ai-setup-ts` is no longer present.

# Task 011: `red-team-plan` Skill + Soft-Fail Tests

**Phase:** W1.B — Plan Quality + Adversarial Design Review  
**User Story:** P3 — Review plans before implementation approval  
**Status:** TODO  
**Depends on:** T8  
**Parallel with:** none

---

## Objective

Create a bounded `red-team-plan` skill for adversarial design review and test the `RedTeamPlanReport` contract, including provider/API outage soft-fail behavior. The skill is read-only and must critique plan/spec design before the human gate when enabled.

## User Story / Spec References

- P3 — Human approvers see optional adversarial findings before implementation approval.
- AC-D17-003 — With `adversarialDesign=true`, `red-team-plan` runs sequentially after plan quality and before approval.
- AC-D17-005 — Red-team provider/API failure emits `status: "soft_fail"` and the approval gate still appears.
- AC-D17-006 — Red-team findings are merged into the human-readable gate report when present.
- FR-012 — Red-team plan review runs sequentially, not via a parallel block.
- FR-013 — Red-team outages soft-fail and surface in the gate report.
- `spec.md` §Internal Contracts `RedTeamPlanReport`.
- `plan.md` §D17 — Adversarial Self-Play During Design.

## Files to Change/Create

- `packages/ai-setup-go/library/skills/red-team-plan.md` — new skill source.
- Existing skill/frontmatter tests, likely `packages/ai-setup-ts/src/__tests__/frontmatter.test.ts`, for metadata and markdown shape if applicable.
- New report-contract tests/fixtures under appropriate package test locations, for example:
  - `packages/ai-setup-ts/src/__tests__/red-team-plan-report.test.ts`
  - `packages/orchestrator/src/__tests__/chain-machine.case.ts` only if soft-fail transition behavior is asserted at runtime.
- Optional fixture Markdown plans/specs for adversarial finding categories.

## Files Not to Touch

- `packages/ai-setup-go/library/orchestration/chains/feature.json` — T12 owns chain insertion/conditional inclusion.
- `packages/ai-setup-go/library/skills/plan.md` — T9 owns planner quality.
- `packages/ai-setup-ts/src/presets.ts`
- `packages/ai-setup-ts/src/wizard/phase2-features.ts`
- Runtime provider configuration or MCP server configuration.
- W1.A task files under `specs/features/001-ai-techniques-integration/tasks/001-*.md` through `007-*.md`
- `specs/features/001-ai-techniques-integration/spec.md`
- `specs/features/001-ai-techniques-integration/plan.md`

## Test-First Order

1. Add failing metadata/frontmatter test proving `red-team-plan.md` exists with the expected skill identity and no write permissions implied.
2. Add failing schema tests for `RedTeamPlanReport` with `schemaVersion: "red-team-plan-report/v1"`, `status: "ok|soft_fail|skipped"`, categories, severity, recommendation, and location shape.
3. Add failing fixture test proving outage/provider failure instructions produce `status: "soft_fail"` rather than blocking/failing the chain.
4. Add failing fixture/snapshot test proving the skill is read-only and critiques `spec.md`/`plan.md` across scope, security, feasibility, rollback, edge-case, assumption, and operational categories.
5. Only after tests fail, create `packages/ai-setup-go/library/skills/red-team-plan.md` with the minimal bounded skill content.

## Done When

- [ ] `packages/ai-setup-go/library/skills/red-team-plan.md` exists and passes skill/frontmatter validation.
- [ ] The skill explicitly reads `spec.md`, `plan.md`, and constitution context; it does not write code or modify docs.
- [ ] `RedTeamPlanReport` schema is tested, including location metadata.
- [ ] Outage/provider/API failure is represented as `status: "soft_fail"` with a human-visible warning finding.
- [ ] The skill text states soft-fail reports proceed to `plan-gate`.
- [ ] No chain insertion is done until T12.

## Risks

- **Red-team outage blocks planning:** mitigated by tested `soft_fail` status and gate continuation instructions.
- **Overbroad adversarial review:** mitigated by fixed categories and read-only plan/spec scope.
- **Scope creep into implementation review:** mitigated by limiting this skill to design-time plan/spec critique before code exists.

## Constitution Check

- **Article I:** Reuse existing skill markdown/frontmatter conventions and existing red-team role language.
- **Article II:** Schema, frontmatter, and soft-fail tests are written before creating the skill.
- **Article III:** Report shape and behavior come from `spec.md`; sequencing comes from `plan.md`.
- **Article IV:** No provider integration, model routing, or debate workflow is added.
- **Article V:** One bounded read-only skill is simpler than a multi-agent debate system.
- **Article VI:** No speculative abstractions or generalized adversarial framework.