> **HISTORICAL (2026-06-20) — parts of this plan target the removed `packages/ai-setup-ts` package; the CLI is now Go-only (`packages/cli`). Retained for historical context.**
>
> This document is historical; `packages/ai-setup-ts` is no longer present.

# Task 005: Starter Standards Content + Frontmatter [P]

**Phase:** W1.A — Constitution + Coverage + Standards Seed  
**User Story:** P2  
**Status:** TODO  
**Depends on:** none  
**Parallel with:** T2, T3, T4

---

## Objective

Create the five starter standards required by N11 and validate them with the existing frontmatter/standard validator before any scaffold copy logic depends on them.

## Spec References

- AC-N11-001 — standard/full installation creates exactly five starter standards under `specs/standards/starter/`.
- AC-N11-002 — every starter standard passes existing standard frontmatter validation.
- AC-N11-005 — at least one concrete starter standard is available to cite for Constitution Quality Gate 4 Pattern Consistency.
- FR-007 — seed five starter standards using file-level idempotency and never overwrite existing user files.
- `plan.md` §N11 Standards-as-Code scope.

## Files to Change

- Existing frontmatter/standard validator tests only if needed to add fixture coverage for the new starter standards.

## Files to Create

- `packages/ai-setup-go/library/standards/starter/orchestration-patterns.md`
- `packages/ai-setup-go/library/standards/starter/test-patterns.md`
- `packages/ai-setup-go/library/standards/starter/error-handling.md`
- `packages/ai-setup-go/library/standards/starter/agent-security.md`
- `packages/ai-setup-go/library/standards/starter/context-loading.md`
- Package-appropriate frontmatter/standard validation tests or fixtures for these five files.

## Files to NOT Touch

- `packages/ai-setup-go/cmd/init.go`
- `packages/ai-setup-go/internal/compiler/template.go`
- `packages/ai-setup-go/internal/scaffold/root.go`
- `packages/ai-setup-go/internal/scaffold/specs.go`
- `packages/ai-setup-go/internal/store/migrations/006-add-constitution-fields.sql`
- `packages/ai-setup-go/library/root/AGENTS.template.md`
- `packages/ai-setup-go/library/rules/testing.md`
- `packages/ai-setup-go/library/skills/plan/SKILL.md`
- `packages/ai-setup-go/library/skills/red-team-plan.md`
- `packages/ai-setup-go/library/orchestration/chains/feature.json`
- `packages/ai-setup-ts/src/scaffold/specs.ts`
- `packages/ai-setup-ts/src/wizard/*`
- `packages/ai-setup-ts/src/compiler/*`
- `packages/ai-setup-ts/src/presets.ts`
- `specs/features/001-ai-techniques-integration/spec.md`
- `specs/features/001-ai-techniques-integration/plan.md`

## Test-First Order

1. Add failing frontmatter/standard validation coverage that expects exactly the five N11 starter files.
2. Add failing validation for required frontmatter fields using the existing validator.
3. Author only the five approved starter standards.
4. Re-run validation after confirming the exact package test command.

## Done When

- [ ] Red frontmatter/standard validation test exists before content is added.
- [ ] Exactly five starter standards are present with the approved filenames.
- [ ] Each starter standard passes the existing frontmatter validator.
- [ ] Content is concrete enough for a reviewer to cite at least one standard for Gate 4 Pattern Consistency.
- [ ] No scaffold copy logic is changed in this task.
- [ ] No additional standards beyond the approved five are introduced.

## Risks

- **Standards seed clashes with existing user standards:** mitigated later by T6 file-level copy; this task limits itself to source content.
- **Scope creep into a standards-management subsystem:** mitigated by authoring only concrete markdown files and validator coverage.

## Constitution Check

- **Article I:** Reuse existing standard/frontmatter validator.
- **Article II:** Validator tests fail before content is added.
- **Article III:** The approved set of five comes from `spec.md` and `plan.md`.
- **Article IV:** Do not add extra standards or Wave 2 quality systems.
- **Article V:** Markdown starter files are simpler than a standards registry.
- **Article VI:** No standards framework or abstraction is created.
