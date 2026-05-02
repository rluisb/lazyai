# Task 012: `adversarialDesign` Flag + Generated Chain Tests

**Phase:** W1.B — Plan Quality + Adversarial Design Review  
**User Story:** P3 — Review plans before implementation approval  
**Status:** TODO  
**Depends on:** T10, T11  
**Parallel with:** none

---

## Objective

Wire the `adversarialDesign` feature flag into presets/custom selection and generated feature-chain content. When enabled, `red-team-plan` runs sequentially after `plan-quality` and before `plan-gate`; when disabled, the generated chain omits or skips it only by the tested mechanism established in T8.

## User Story / Spec References

- P3 — Human approvers see optional adversarial findings before implementation approval.
- AC-D17-001 — Defaults: minimal=false, standard/full=true.
- AC-D17-002 — Custom preset toggle controls generated chain content.
- AC-D17-003 — With `adversarialDesign=true`, `red-team-plan` runs sequentially after plan quality and before approval.
- AC-D17-004 — With `adversarialDesign=false`, no unverified runtime conditional behavior is assumed.
- AC-D17-006 — When both reports exist, findings are merged into one gate report.
- FR-011 through FR-014.
- `spec.md` §Feature Flags and Preset Defaults.
- `plan.md` §D17 and §Sequential chain design for W1.B.

## Files to Change/Create

- `packages/ai-setup-go/library/orchestration/chains/feature.json` — include `red-team-plan` only by the proven T8 mechanism; keep chain sequential.
- `packages/ai-setup-go/internal/scaffold/orchestration.go` and `packages/ai-setup-go/internal/scaffold/orchestration_test.go` — generate/install enabled and disabled chain variants if compile-time/static generation is the fallback.
- `packages/ai-setup-ts/src/scaffold/orchestration.ts` and `packages/ai-setup-ts/src/__tests__/orchestration.test.ts` — mirror generated chain behavior.
- `packages/ai-setup-ts/src/presets.ts` and `packages/ai-setup-ts/src/__tests__/presets.test.ts` — verify defaults if not already covered by W1.A.
- `packages/ai-setup-ts/src/wizard/phase2-features.ts` and `packages/ai-setup-ts/src/__tests__/phase2-features.test.ts` — custom toggle if TS owns this UX surface.
- Go preset/wizard/config files and tests equivalent to the TS surfaces if present in the current implementation.
- `packages/ai-setup-ts/src/store/schema.ts` and Go `packages/ai-setup-go/internal/types/types.go` only if T1/W1.A did not already add the flag with correct defaults.

## Files Not to Touch

- `packages/ai-setup-go/library/skills/plan.md` — T9 owns planner quality.
- `packages/ai-setup-go/library/skills/red-team-plan.md` content — T11 owns skill content; T12 may reference it only from the chain.
- Runtime conditional semantics not proven in T8.
- MCP server configuration.
- W1.A task files under `specs/features/001-ai-techniques-integration/tasks/001-*.md` through `007-*.md`
- `specs/features/001-ai-techniques-integration/spec.md`
- `specs/features/001-ai-techniques-integration/plan.md`

## Test-First Order

1. Add failing preset tests proving `adversarialDesign` defaults are minimal=false, standard=true, full=true.
2. Add failing custom preset/wizard tests proving a user toggle is persisted and passed to chain generation.
3. Add failing chain-generation tests for `adversarialDesign=true` expecting sequential order: `plan → plan-quality → red-team-plan → plan-gate`.
4. Add failing chain-generation tests for `adversarialDesign=false` expecting either:
   - no `red-team-plan` step in the generated chain, if static generation fallback is used; or
   - a tested runtime skip mechanism from T8, if runtime conditionals are proven.
5. Add failing assertions that no parallel block appears in either generated chain.
6. Only after tests fail, wire the flag through the smallest existing preset/wizard/scaffold surfaces needed.

## Done When

- [ ] Preset defaults match `spec.md`: minimal=false, standard/full=true.
- [ ] Custom preset selection can toggle `adversarialDesign` and the generated chain reflects the choice.
- [ ] Enabled generated chain has `red-team-plan` sequentially after `plan-quality` and before `plan-gate`.
- [ ] Disabled generated chain omits or skips `red-team-plan` only by the T8-proven mechanism.
- [ ] No chain contains a parallel block.
- [ ] Go and TS scaffold/generation behavior is parity-tested where both surfaces generate orchestration content.
- [ ] Installed `.ai/orchestration/chains/feature.json` is verified for enabled and disabled cases.

## Risks

- **Unverified runtime conditionals:** mitigated by dependency on T8 and static variants fallback.
- **Go/TS preset drift:** mitigated by mirrored preset and scaffold tests.
- **Chain JSON becomes invalid when omitting a step:** mitigated by tests parsing both generated variants and checking transitions.
- **Red-team inserted before plan quality:** mitigated by strict order assertions.

## Constitution Check

- **Article I:** Reuse existing preset, wizard, store, compiler/scaffold, and orchestration surfaces.
- **Article II:** Defaults, custom toggle, and generated-chain tests fail before wiring changes.
- **Article III:** Defaults and chain behavior trace to `spec.md` and `plan.md`.
- **Article IV:** Do not add flags beyond `adversarialDesign`; do not add untested runtime conditional behavior.
- **Article V:** Static generated variants are preferred if runtime support is not already proven.
- **Article VI:** No feature-flag framework or orchestration abstraction beyond the existing surfaces.