> **HISTORICAL (2026-06-20) — parts of this plan target the removed `packages/ai-setup-ts` package; the CLI is now Go-only (`packages/cli`). Retained for historical context.**
>
> This document is historical; `packages/ai-setup-ts` is no longer present.

# Task 002: FragmentContext Fallbacks

**Phase:** W1.A — Constitution + Coverage + Standards Seed  
**User Story:** P1/P2  
**Status:** TODO  
**Depends on:** T1  
**Parallel with:** T5

---

## Objective

Extend compiler/template behavior for `FragmentContext.constitution` so populated values render into root instructions while skipped or unknown values emit the documented literal fallbacks exactly.

## Spec References

- AC-N8-001 — collected fields remove matching `[YOUR_*]` markers from generated `AGENTS.md`.
- AC-N8-002 — skipped/unknown fields preserve fallback markers and never render empty string, `null`, or `undefined`.
- AC-N8-004 — generated codebase map excludes ignored directories and keeps `[WHAT_IT_DOES]` responsibilities.
- AC-N8-005 — Go and TS compiled root instructions are byte-identical for fields in scope.
- AC-N4-002 — coverage threshold appears in `AGENTS.md` and selected `specs/rules/testing.md`.
- AC-N4-003 — skipped coverage prompt uses default threshold, not `0%` or unresolved variable.
- `spec.md` §FragmentContext extension.
- `spec.md` §Template fallback map.

## Files to Change

- `packages/ai-setup-go/internal/compiler/template.go`
- `packages/ai-setup-go/internal/scaffold/constitution.go`
- `packages/ai-setup-go/library/root/AGENTS.template.md`
- `packages/ai-setup-go/library/rules/testing.md`
- `packages/ai-setup-ts/src/compiler/fragment-resolver.ts`
- `packages/ai-setup-ts/src/scaffold/compiled-root.ts`

## Files to Create

- Package-appropriate Go compiler/scaffold tests for `FragmentContext.constitution` fields, fallback markers, coverage threshold substitution, and codebase map filtering.
- Package-appropriate TS compiler/scaffold tests mirroring the Go cases.
- Optional snapshot fixtures only if they are local to this compiler/fallback behavior; full parity snapshots belong to T4.

## Files to NOT Touch

- `packages/ai-setup-go/cmd/init.go`
- `packages/ai-setup-go/internal/store/migrations/006-add-constitution-fields.sql`
- `packages/ai-setup-go/internal/scaffold/root.go`
- `packages/ai-setup-go/internal/scaffold/specs.go`
- `packages/ai-setup-go/library/standards/starter/*`
- `packages/ai-setup-go/library/skills/plan/SKILL.md`
- `packages/ai-setup-go/library/orchestration/chains/feature.json`
- `packages/ai-setup-ts/src/wizard/phase2-features.ts`
- `packages/ai-setup-ts/src/wizard/planner.ts`
- `packages/ai-setup-ts/src/scaffold/specs.ts`
- `packages/ai-setup-ts/src/presets.ts`
- `specs/features/001-ai-techniques-integration/spec.md`
- `specs/features/001-ai-techniques-integration/plan.md`

## Test-First Order

1. Add failing Go tests for each fallback mapping in `spec.md`: project overview, conventions, commands, protected branch, coverage threshold, and codebase map responsibility placeholders.
2. Add failing Go tests for populated values replacing only the matching placeholders.
3. Add failing TS tests with identical input/output expectations.
4. Add tests proving missing coverage resolves to default/fallback behavior required by the spec.
5. Implement Go compiler/template changes first, then TS parity changes.

## Done When

- [ ] Red tests exist before template/compiler production changes.
- [ ] Empty, null, or omitted fields render documented fallback literals exactly.
- [ ] Populated fields render without leaving corresponding collected-field `[YOUR_*]` markers.
- [ ] `coverageThreshold` renders in `AGENTS.md` and `specs/rules/testing.md` when selected.
- [ ] Generated codebase map excludes `node_modules`, `dist`, `.git`, and `vendor`.
- [ ] Codebase map responsibility cells remain `[WHAT_IT_DOES]`.
- [ ] Go-first implementation is verified against mirrored TS tests.

## Risks

- **TS/Go drift on fields and defaults:** mitigated by mirrored tests and parity-ready expectations.
- **Wizard expansion fatigues users:** mitigated indirectly by preserving placeholders for skipped fields rather than forcing answers.

## Constitution Check

- **Article I:** Reuse existing compiler and template mechanisms; no new template engine.
- **Article II:** Failing compiler/scaffold tests precede production changes.
- **Article III:** `spec.md` fallback map is authoritative; do not invent placeholders.
- **Article IV:** Do not add extra project-profile fields beyond W1.A.
- **Article V:** Use a single explicit field-to-placeholder map as planned.
- **Article VI:** Avoid helper extraction unless existing duplication exceeds project rules and has multiple real callers.
