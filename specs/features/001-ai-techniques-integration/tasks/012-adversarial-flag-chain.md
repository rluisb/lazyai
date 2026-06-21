> **HISTORICAL (2026-06-20) — parts of this plan target the removed `packages/ai-setup-ts` package; the CLI is now Go-only (`packages/cli`). Retained for historical context.**
>
> This document is historical; `packages/ai-setup-ts` is no longer present.

# Task 012: `adversarialDesign` Flag + Explicit Generated Chain Selection Tests

**Phase:** W1.B — Plan Quality + Adversarial Design Review  
**User Story:** P3 — Review plans before implementation approval  
**Status:** TODO  
**Depends on:** T8 evidence, T10, T11  
**Parallel with:** none

---

## Objective

Wire the `adversarialDesign` feature flag into presets/custom selection and generated feature-chain content using the T8-proven implementation constraints. When enabled, scaffold/install logic must generate or copy a sequential chain with `red-team-plan` after `plan-quality` and before `plan-gate`. When disabled, scaffold/install logic must generate or copy a sequential chain that omits `red-team-plan`.

T8 proved runtime conditional step execution and compile-time orchestration template rendering are absent/unsupported. Therefore T12 must not use `{{#if features.adversarialDesign}}`, `optionalByFeature`, `condition`, runtime skip semantics, a generic chain template renderer, or parallel blocks.

## User Story / Spec References

- P3 — Human approvers see optional adversarial findings before implementation approval.
- AC-D17-001 — Defaults: minimal=false, standard/full=true.
- AC-D17-002 — Custom preset toggle controls generated chain content.
- AC-D17-003 — With `adversarialDesign=true`, `red-team-plan` runs sequentially after plan quality and before approval.
- AC-D17-004 — With `adversarialDesign=false`, no unverified runtime conditional behavior is assumed.
- AC-D17-006 — When both reports exist, findings are merged into one gate report.
- FR-011 through FR-014.
- T8 evidence commit `99741a7` — sequential gates are supported; runtime conditionals, compile-time chain template rendering, and parallel blocks are absent/unsupported.
- `spec.md` §Feature Flags and Preset Defaults.
- `plan.md` §D17 and §Sequential chain design for W1.B.

## Files to Change/Create

- `packages/ai-setup-go/library/orchestration/chains/feature.json` — treat as the disabled/base sequential chain from T10 unless current code conventions require a different explicit source name.
- An explicit enabled sequential chain source or equivalent explicit construction path, for example an adjacent `feature-adversarial*.json` source file if that matches scaffold conventions; it must be ordinary JSON, not a templated chain file.
- `packages/ai-setup-go/internal/scaffold/orchestration.go` and `packages/ai-setup-go/internal/scaffold/orchestration_test.go` — choose/copy the enabled or disabled chain into installed `.ai/orchestration/chains/feature.json` based on resolved `adversarialDesign`.
- `packages/ai-setup-ts/src/scaffold/orchestration.ts` and `packages/ai-setup-ts/src/__tests__/orchestration.test.ts` — mirror explicit enabled/disabled chain selection if TS owns orchestration installation.
- `packages/ai-setup-ts/src/presets.ts` and `packages/ai-setup-ts/src/__tests__/presets.test.ts` — verify defaults if not already covered by W1.A.
- `packages/ai-setup-ts/src/wizard/phase2-features.ts` and `packages/ai-setup-ts/src/__tests__/phase2-features.test.ts` — custom toggle if TS owns this UX surface.
- Go preset/wizard/config files and tests equivalent to the TS surfaces if present in the current implementation.
- `packages/ai-setup-ts/src/store/schema.ts` and Go `packages/ai-setup-go/internal/types/types.go` only if T1/W1.A did not already add the flag with correct defaults.

## Files Not to Touch

- `packages/ai-setup-go/library/skills/plan.md` — T9 owns planner quality.
- `packages/ai-setup-go/library/skills/red-team-plan.md` content — T11 owns skill content; T12 may reference it only from the chain.
- Runtime conditional semantics, source-chain template conditionals, generic orchestration template rendering, and parallel blocks — T8 proved these are absent/unsupported.
- MCP server configuration.
- W1.A task files under `specs/features/001-ai-techniques-integration/tasks/001-*.md` through `007-*.md`
- `specs/features/001-ai-techniques-integration/spec.md`
- `specs/features/001-ai-techniques-integration/plan.md`

## Test-First Order

1. Add failing preset tests proving `adversarialDesign` defaults are minimal=false, standard=true, full=true.
2. Add failing custom preset/wizard tests proving a user toggle is persisted and passed to chain generation.
3. Add failing chain-generation tests for `adversarialDesign=true` expecting sequential order: `plan → plan-quality → red-team-plan → plan-gate`.
4. Add failing chain-generation tests for `adversarialDesign=false` expecting no `red-team-plan` step in the generated chain.
5. Add failing assertions that neither generated chain contains a parallel block, `{{#if ...}}` marker, `optionalByFeature`, `condition`, or other runtime/template conditional marker.
6. Add failing install/scaffold tests proving both enabled and disabled selections install to the same destination path, `.ai/orchestration/chains/feature.json`, with the correct sequential content for the selected flag.
7. Only after tests fail, wire the flag through the smallest existing preset/wizard/scaffold surfaces needed.

## Scope Decision Gate

Preferred path: add explicit scaffold/install-time chain selection for enabled vs. disabled D17 chains. If the current scaffold code makes that change larger than a bounded chain-selection patch, stop before implementation and ask for a scope decision: either approve the explicit chain-selection work or defer automatic D17 chain wiring while keeping `red-team-plan` available for manual skill use.

## Done When

- [ ] Preset defaults match `spec.md`: minimal=false, standard/full=true.
- [ ] Custom preset selection can toggle `adversarialDesign` and the generated chain reflects the choice.
- [ ] Enabled generated chain has `red-team-plan` sequentially after `plan-quality` and before `plan-gate`.
- [ ] Disabled generated chain omits `red-team-plan` by explicit scaffold/install-time chain selection, not by runtime skip semantics.
- [ ] No source or generated chain contains `{{#if ...}}`, `optionalByFeature`, `condition`, or other runtime/template conditional marker.
- [ ] No chain contains a parallel block.
- [ ] Go and TS scaffold/generation behavior is parity-tested where both surfaces generate orchestration content.
- [ ] Installed `.ai/orchestration/chains/feature.json` is verified for enabled and disabled cases.

## Risks

- **Unsupported runtime/template conditionals:** mitigated by dependency on T8 and explicit chain selection tests.
- **Go/TS preset drift:** mitigated by mirrored preset and scaffold tests.
- **Chain JSON becomes invalid when omitting a step:** mitigated by tests parsing both generated variants and checking transitions.
- **Red-team inserted before plan quality:** mitigated by strict order assertions.
- **Chain-selection work broadens beyond W1.B:** mitigated by the Scope Decision Gate; defer automatic D17 wiring rather than inventing a generic framework.

## Constitution Check

- **Article I:** Reuse existing preset, wizard, store, compiler/scaffold, and orchestration surfaces.
- **Article II:** Defaults, custom toggle, and generated-chain tests fail before wiring changes.
- **Article III:** Defaults and chain behavior trace to `spec.md` and `plan.md`.
- **Article IV:** Do not add flags beyond `adversarialDesign`; do not add runtime conditional behavior, template-rendered chain files, or parallel blocks.
- **Article V:** Explicit chain selection is the simplest path compatible with T8 evidence.
- **Article VI:** No feature-flag framework, generic orchestration abstraction, or chain template renderer beyond the existing surfaces.
