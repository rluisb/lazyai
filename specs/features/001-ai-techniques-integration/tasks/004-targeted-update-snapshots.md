# Task 004: Targeted Update + Parity Snapshots

**Phase:** W1.A — Constitution + Coverage + Standards Seed  
**User Story:** P1/P2  
**Status:** TODO  
**Depends on:** T1  
**Parallel with:** T5

---

## Objective

Replace any destructive root-update behavior with targeted placeholder/value-slot patching and add Go/TS snapshot parity for W1.A generated fields.

## Spec References

- AC-N8-003 — existing hand-edited `AGENTS.md` is patched only in recognized placeholders/value slots; custom content is preserved verbatim.
- AC-N8-005 — Go and TS compile root instructions byte-identically for fields in scope.
- AC-N4-002 — coverage threshold appears in generated `AGENTS.md` and selected `specs/rules/testing.md`.
- `spec.md` §TargetedUpdatePatch contract.
- `plan.md` §Data Model targeted update policy.

## Files to Change

- `packages/ai-setup-go/internal/scaffold/root.go`
- `packages/ai-setup-go/internal/scaffold/constitution.go`
- `packages/ai-setup-go/library/root/AGENTS.template.md`
- `packages/ai-setup-go/library/rules/testing.md`
- `packages/ai-setup-ts/src/scaffold/compiled-root.ts`
- `packages/ai-setup-ts/src/compiler/fragment-resolver.ts`

## Files to Create

- Go targeted-update fixture tests for hand-edited `AGENTS.md` preservation.
- TS targeted-update or compiled-root fixture tests mirroring Go behavior where TS owns update/scaffold output.
- Go/TS parity snapshot fixtures for W1.A fields.

## Files to NOT Touch

- `packages/ai-setup-go/cmd/init.go`
- `packages/ai-setup-go/internal/store/migrations/006-add-constitution-fields.sql`
- `packages/ai-setup-go/internal/scaffold/specs.go`
- `packages/ai-setup-go/library/standards/starter/*`
- `packages/ai-setup-go/library/skills/plan/SKILL.md`
- `packages/ai-setup-go/library/skills/red-team-plan.md`
- `packages/ai-setup-go/library/orchestration/chains/feature.json`
- `packages/ai-setup-ts/src/wizard/phase2-features.ts`
- `packages/ai-setup-ts/src/wizard/planner.ts`
- `packages/ai-setup-ts/src/scaffold/specs.ts`
- `packages/ai-setup-ts/src/presets.ts`
- `specs/features/001-ai-techniques-integration/spec.md`
- `specs/features/001-ai-techniques-integration/plan.md`

## Test-First Order

1. Add failing targeted-update fixture tests with custom sections, reordered headings, existing user-authored rules, and known placeholders.
2. Add failing tests asserting `TargetedUpdatePatch` includes `file`, `replacements[]` with `field`, `oldText`, `newText`, `location`, `warnings[]`, and `preservedUnrecognizedContent: true`.
3. Add failing tests for unsafe/unparseable known fields: leave existing text unchanged and emit warning.
4. Add failing Go/TS parity snapshots for identical W1.A input.
5. Implement Go targeted update first; then TS parity/snapshot changes as applicable.

## Done When

- [ ] Red targeted-update and snapshot tests exist before production changes.
- [ ] Update never full-re-emits `AGENTS.md` by default.
- [ ] Recognized placeholders/value slots are updated; unknown content is preserved byte-for-byte.
- [ ] Unsafe parse cases produce warnings and leave content unchanged.
- [ ] `TargetedUpdatePatch` JSON contract is asserted.
- [ ] Go/TS generated output is byte-identical for W1.A fields under identical input.
- [ ] Coverage threshold substitutions remain covered in root/rules output.

## Risks

- **`update` corrupts hand-edited `AGENTS.md`:** mitigate with targeted replacement only and fixtures that preserve structural edits/custom sections byte-for-byte.
- **TS/Go drift on fields and defaults:** mitigate with parity snapshots.

## Constitution Check

- **Article I:** Reuse existing scaffold/update paths and snapshot harnesses.
- **Article II:** Fixtures/snapshots fail before production changes.
- **Article III:** `TargetedUpdatePatch` shape and update policy come from `spec.md`/`plan.md`.
- **Article IV:** Do not implement full template refresh or new update modes.
- **Article V:** Prefer direct targeted replacements over a document-rewrite subsystem.
- **Article VI:** Avoid one-caller parser abstractions; preserve concrete fixture-driven behavior.
