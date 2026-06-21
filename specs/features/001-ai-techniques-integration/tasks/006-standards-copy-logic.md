> **HISTORICAL (2026-06-20) — parts of this plan target the removed `packages/ai-setup-ts` package; the CLI is now Go-only (`packages/cli`). Retained for historical context.**
>
> This document is historical; `packages/ai-setup-ts` is no longer present.

# Task 006: Standards File-Level Copy Logic

**Phase:** W1.A — Constitution + Coverage + Standards Seed  
**User Story:** P2  
**Status:** TODO  
**Depends on:** T5  
**Parallel with:** none

---

## Objective

Seed starter standards into target repositories using file-level idempotency: copy each missing starter file independently and never overwrite user-authored files.

## Spec References

- AC-N11-001 — standard/full installation creates exactly five starter standards.
- AC-N11-003 — `.gitkeep` or `README.md` alone does not block missing starter files from being copied.
- AC-N11-004 — existing user-authored files at starter-standard paths are not overwritten.
- FR-007 — file-level idempotency and no overwrite behavior.
- `plan.md` §Internal Contracts — Standards seed copy.

## Files to Change

- `packages/ai-setup-go/internal/scaffold/specs.go`
- `packages/ai-setup-ts/src/scaffold/specs.ts`

## Files to Create

- Package-appropriate Go scaffold tests for standards copy/idempotency.
- Package-appropriate TS scaffold tests mirroring Go standards copy/idempotency.
- Fixtures for target `specs/standards/starter/` containing only `.gitkeep`, only `README.md`, and pre-existing same-path user files.

## Files to NOT Touch

- `packages/ai-setup-go/cmd/init.go`
- `packages/ai-setup-go/internal/compiler/template.go`
- `packages/ai-setup-go/internal/scaffold/root.go`
- `packages/ai-setup-go/internal/scaffold/constitution.go`
- `packages/ai-setup-go/internal/store/migrations/006-add-constitution-fields.sql`
- `packages/ai-setup-go/library/root/AGENTS.template.md`
- `packages/ai-setup-go/library/rules/testing.md`
- `packages/ai-setup-go/library/skills/plan/SKILL.md`
- `packages/ai-setup-go/library/orchestration/chains/feature.json`
- `packages/ai-setup-ts/src/wizard/phase2-features.ts`
- `packages/ai-setup-ts/src/wizard/planner.ts`
- `packages/ai-setup-ts/src/scaffold/compiled-root.ts`
- `packages/ai-setup-ts/src/compiler/fragment-resolver.ts`
- `packages/ai-setup-ts/src/presets.ts`
- `specs/features/001-ai-techniques-integration/spec.md`
- `specs/features/001-ai-techniques-integration/plan.md`

## Test-First Order

1. Add failing Go scaffold test: empty target receives all five T5 starter files.
2. Add failing Go scaffold tests: target with `.gitkeep` and/or `README.md` still receives missing starter files.
3. Add failing Go scaffold test: pre-existing user-authored starter file is preserved byte-for-byte while other missing files copy.
4. Add mirrored failing TS tests.
5. Implement Go copy logic first, then TS parity.

## Done When

- [ ] Red copy/idempotency tests exist before production changes.
- [ ] File-level copy checks each approved starter file independently.
- [ ] `.gitkeep` and `README.md` do not suppress copying missing starter standards.
- [ ] Existing user-authored files are never overwritten.
- [ ] Standard/full installation surfaces exactly the five starter standards from T5.
- [ ] Go-first implementation has TS parity test evidence.

## Risks

- **Standards seed clashes with existing user standards:** mitigated by file-level copy and never-overwrite fixtures.
- **TS/Go drift:** mitigated by mirrored scaffold tests and Go-first/TS-follow implementation.

## Constitution Check

- **Article I:** Reuse existing scaffold file-copy utilities where present.
- **Article II:** Idempotency tests fail before copy logic changes.
- **Article III:** File-level idempotency is required by `spec.md` and `plan.md`.
- **Article IV:** Do not add standards sync/update/delete behavior.
- **Article V:** Direct per-file copy is simpler than a standards management subsystem.
- **Article VI:** Avoid extracting new copy abstractions unless existing utilities already support the need.
