> **HISTORICAL (2026-06-20) — parts of this plan target the removed `packages/ai-setup-ts` package; the CLI is now Go-only (`packages/cli`). Retained for historical context.**
>
> This document is historical; `packages/ai-setup-ts` is no longer present.

# Task 003: Wizard Prompts + Config Wiring

**Phase:** W1.A — Constitution + Coverage + Standards Seed  
**User Story:** P1/P2  
**Status:** TODO  
**Depends on:** T1  
**Parallel with:** T5

---

## Objective

Wire project profile and coverage threshold collection through existing Go and TS wizard/config flows, using Go-first implementation and TS parity verification.

## Spec References

- AC-N8-001 — answered profile questions populate generated `AGENTS.md` fields.
- AC-N8-002 — non-interactive/skipped optional fields preserve fallback markers.
- AC-N8-006 — missing v2 fields backfill safely when read.
- AC-N4-001 — supported stacks show non-zero stack-aware coverage defaults; unknown stacks default to 80.
- AC-N4-002 — accepted/supplied coverage threshold appears in generated outputs.
- AC-N4-003 — skipped coverage prompt uses default threshold.
- AC-N4-004 — out-of-range thresholds are rejected or reset to default.
- FR-001, FR-004, FR-006.

## Files to Change

- `packages/ai-setup-go/cmd/init.go`
- `packages/ai-setup-go/internal/migration/json_bridge.go`
- `packages/ai-setup-ts/src/wizard/phase2-features.ts`
- `packages/ai-setup-ts/src/wizard/planner.ts`
- `packages/ai-setup-ts/src/store/schema.ts`

## Files to Create

- Package-appropriate Go wizard/config tests for project profile prompts, non-interactive defaults, and coverage validation.
- Package-appropriate TS wizard/planner/store tests mirroring Go behavior.

## Files to NOT Touch

- `packages/ai-setup-go/internal/compiler/template.go`
- `packages/ai-setup-go/internal/scaffold/root.go`
- `packages/ai-setup-go/internal/scaffold/specs.go`
- `packages/ai-setup-go/library/root/AGENTS.template.md`
- `packages/ai-setup-go/library/rules/testing.md`
- `packages/ai-setup-go/library/standards/starter/*`
- `packages/ai-setup-go/library/skills/plan/SKILL.md`
- `packages/ai-setup-go/library/orchestration/chains/feature.json`
- `packages/ai-setup-ts/src/scaffold/compiled-root.ts`
- `packages/ai-setup-ts/src/scaffold/specs.ts`
- `packages/ai-setup-ts/src/compiler/fragment-resolver.ts`
- `packages/ai-setup-ts/src/presets.ts` except if an existing config type import requires a compile-only update explicitly tied to T1 schema.
- `specs/features/001-ai-techniques-integration/spec.md`
- `specs/features/001-ai-techniques-integration/plan.md`

## Test-First Order

1. Add failing Go headless wizard/config tests for project overview, naming conventions, error handling, API conventions, import order, protected branch, commands, and coverage threshold.
2. Add failing Go tests for stack-aware coverage defaults and unknown-stack default `80`.
3. Add failing Go tests for out-of-range thresholds.
4. Add mirrored failing TS tests for phase2 wizard/planner/store behavior.
5. Implement Go prompt/config wiring first; then implement TS parity.

## Done When

- [ ] Red wizard/config tests are written before production changes.
- [ ] Project Profile prompts are grouped and optional/skippable without unsafe empty rendering.
- [ ] Existing command auto-detection is reused where available; no new detection framework is introduced.
- [ ] Coverage default is non-zero for supported stacks and `80` for unknown stacks.
- [ ] Coverage values outside 1–100 are rejected or reset to documented default.
- [ ] Non-interactive init/config paths use safe defaults and do not crash when v2 fields are missing.
- [ ] Go-first changes have TS parity tests/evidence.

## Risks

- **Wizard expansion fatigues users:** mitigate by grouping fields under one Project Profile block, reusing auto-detection, and preserving placeholders for skipped fields.
- **TS/Go drift on fields and defaults:** mitigate by mirrored tests and Go-first/TS-follow discipline.

## Constitution Check

- **Article I:** Reuse existing wizard drivers and config/store wiring.
- **Article II:** Wizard/config tests fail before prompt implementation.
- **Article III:** Prompted fields match `spec.md` `WizardConfig`; do not add undocumented questions.
- **Article IV:** No Wave 2/3/4 prompts or MCP config changes.
- **Article V:** Keep one concrete Project Profile block; no wizard engine abstraction.
- **Article VI:** Do not add speculative config knobs beyond N8/N4 fields and schema-required feature flag default.
