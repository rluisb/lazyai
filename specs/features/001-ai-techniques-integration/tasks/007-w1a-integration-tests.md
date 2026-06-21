> **HISTORICAL (2026-06-20) — parts of this plan target the removed `packages/ai-setup-ts` package; the CLI is now Go-only (`packages/cli`). Retained for historical context.**
>
> This document is historical; `packages/ai-setup-ts` is no longer present.

# Task 007: W1.A Integration + Golden Tests

**Phase:** W1.A — Constitution + Coverage + Standards Seed  
**User Story:** P1/P2  
**Status:** TODO  
**Depends on:** T4, T6  
**Parallel with:** none

---

## Objective

Add final W1.A end-to-end and golden validation proving constitution population, coverage thresholds, targeted update safety, and standards seeding work together across Go and TS surfaces.

## Spec References

- AC-N8-001 through AC-N8-007.
- AC-N4-001 through AC-N4-004.
- AC-N11-001 through AC-N11-005.
- `spec.md` §Verification Approach Summary.
- `plan.md` §Phases & Milestones W1.A exit criterion.

## Files to Change

- Existing Go integration/golden test files under `packages/ai-setup-go/` as appropriate for init/update/scaffold coverage.
- Existing TS integration/golden test files under `packages/ai-setup-ts/src/__tests__/` as appropriate for init/update/scaffold coverage.

## Files to Create

- New Go fixtures/golden outputs for W1.A populated and skipped-field flows if existing fixture locations do not cover them.
- New TS fixtures/golden outputs for W1.A parity if existing fixture locations do not cover them.
- Acceptance trace notes inside test descriptions or comments mapping W1.A AC IDs to evidence.

## Files to NOT Touch

- `packages/ai-setup-go/library/skills/plan/SKILL.md`
- `packages/ai-setup-go/library/skills/red-team-plan.md`
- `packages/ai-setup-go/library/orchestration/chains/feature.json`
- `packages/ai-setup-ts/src/presets.ts` except schema compile-only needs already approved by prior tasks.
- Any Wave 1.B plan-quality/adversarial design files.
- Any Wave 2/3/4 roadmap files.
- MCP server configuration.
- `specs/features/001-ai-techniques-integration/spec.md`
- `specs/features/001-ai-techniques-integration/plan.md`

## Test-First Order

1. Add failing Go integration/golden test for answered profile questions: generated `AGENTS.md` has no collected-field `[YOUR_*]` markers.
2. Add failing Go integration/golden test for skipped/non-interactive fields: documented literal fallback markers are preserved exactly.
3. Add failing Go integration/golden test for coverage threshold in `AGENTS.md` and selected `specs/rules/testing.md`.
4. Add failing Go integration test for starter standards file-level idempotency in a realistic scaffold run.
5. Add mirrored TS integration/golden tests for parity where TS owns the same surface.
6. Only after failures are observed, make the smallest integration adjustments required to connect prior task outputs.

## Done When

- [ ] Red integration/golden tests exist before any integration production changes.
- [ ] Fresh init with answered W1.A profile fields produces `AGENTS.md` without collected-field fallback markers.
- [ ] Non-interactive/skipped fields preserve documented fallback markers exactly.
- [ ] Coverage threshold default/accepted value is present in `AGENTS.md` and selected testing rule.
- [ ] Generated codebase map excludes ignored directories and keeps `[WHAT_IT_DOES]` cells.
- [ ] Existing hand-edited `AGENTS.md` remains preserved under update fixtures.
- [ ] Exactly five starter standards are present after scaffold, and same-path user files are not overwritten.
- [ ] Go/TS parity evidence is captured for shared W1.A outputs.
- [ ] All W1.A AC IDs are traceable to at least one test or documented verification artifact.

## Risks

- **`update` corrupts hand-edited `AGENTS.md`:** integration fixtures must include custom content preservation.
- **Standards seed clashes with existing user standards:** integration fixtures must prove no overwrite behavior.
- **TS/Go drift on fields and defaults:** golden parity evidence closes W1.A.
- **Wizard expansion fatigues users:** integration should verify skipped fields remain safe.

## Constitution Check

- **Article I:** Use existing integration/golden test harnesses.
- **Article II:** Integration/golden tests are written failing before final glue changes.
- **Article III:** W1.A exit criteria and AC IDs are traced from `spec.md` and `plan.md`.
- **Article IV:** Stop at N8/N4/N11; do not begin W1.B chain/report work.
- **Article V:** Prefer end-to-end validation through existing init/update flows over new harnesses.
- **Article VI:** No speculative abstractions; integration changes only connect prior task outputs.
