> **HISTORICAL (2026-06-20) — parts of this plan target the removed `packages/ai-setup-ts` package; the CLI is now Go-only (`packages/cli`). Retained for historical context.**
>
> This document is historical; `packages/ai-setup-ts` is no longer present.

# Task 001: Schema + Migration Tests

**Phase:** W1.A — Constitution + Coverage + Standards Seed  
**User Story:** P1 primary; P2 for coverage threshold config  
**Status:** IN PROGRESS  
**Depends on:** none  
**Parallel with:** none

---

## Objective

Establish the W1.A schema safety gate before any compiler, wizard, or scaffold code writes new configuration. Cover Go migration 007 (embedded constant in migrations.go), TS lowdb v1→v2 safe defaults, and TS v2→v1 downgrade behavior.

## Spec References

- AC-N8-006 — TS lowdb v1 reads with missing v2 fields backfilled safely.
- AC-N8-007 — documented downgrade helper/procedure removes v2-only fields for older strict CLIs.
- AC-N4-004 — coverage threshold values outside 1–100 are rejected or reset to documented default.
- FR-004 — new profile, coverage, and feature-flag fields are persisted with backward-compatible defaults.
- FR-005 — downgrade path exists for TS lowdb v2 stores.
- FR-006 — coverage thresholds are constrained to 1–100.
- `spec.md` §WizardConfig (version-adjusted: v1→v2, not v2→v3).

## Repo Reality Notes

- Go migrations are **embedded SQL constants** in `internal/db/migrations.go` (not `.sql` files in a `store/migrations/` dir). Current highest version is 6 (`setup_type`). New migration = version 7.
- Go types are in `internal/types/types.go`. `json_bridge.go` is at `internal/db/json_bridge.go` (imports types, no change needed for this task).
- TS `CURRENT_SCHEMA_VERSION` is **1** (not 2). Bump to 2 for new fields.

## Files to Change

- `packages/ai-setup-go/internal/db/migrations.go` — add migration 007 as embedded SQL constant
- `packages/ai-setup-go/internal/types/types.go` — add fields to StoreData/Config structs
- `packages/ai-setup-ts/src/store/schema.ts` — bump `CURRENT_SCHEMA_VERSION` 1→2, add optional fields to configSchema

## Files to Create

- `packages/ai-setup-ts/src/store/migrations/v2-to-v1.ts` — downgrade helper (removes v2-only keys)
- Go test file(s) under `packages/ai-setup-go/internal/db/` for migration coverage
- TS test file(s) under `packages/ai-setup-ts/src/store/` for v1→v2 defaults and v2→v1 downgrade

## Files to NOT Touch

- `packages/ai-setup-go/cmd/init.go`
- `packages/ai-setup-go/internal/compiler/template.go`
- `packages/ai-setup-go/internal/scaffold/root.go`
- `packages/ai-setup-go/internal/scaffold/specs.go`
- `packages/ai-setup-go/library/root/AGENTS.template.md`
- `packages/ai-setup-go/library/rules/testing.md`
- `packages/ai-setup-go/library/skills/plan/SKILL.md`
- `packages/ai-setup-go/library/skills/red-team-plan.md`
- `packages/ai-setup-go/library/orchestration/chains/feature.json`
- `packages/ai-setup-ts/src/wizard/phase2-features.ts`
- `packages/ai-setup-ts/src/wizard/planner.ts`
- `packages/ai-setup-ts/src/scaffold/compiled-root.ts`
- `packages/ai-setup-ts/src/scaffold/specs.ts`
- `packages/ai-setup-ts/src/compiler/fragment-resolver.ts`
- `packages/ai-setup-ts/src/presets.ts`
- `specs/features/001-ai-techniques-integration/spec.md`
- `specs/features/001-ai-techniques-integration/plan.md`

## Test-First Order

1. Add failing Go migration test (in `internal/db/migration_test.go` or similar) proving migration 007 adds nullable constitution/profile/coverage columns to the `config` table without breaking existing data.
2. Add failing TS store tests proving a v1 lowdb document reads as v2 with safe defaults: nullable/optional profile fields, `coverageThreshold: 80`, and `featureFlags.adversarialDesign` defaulted to false.
3. Add failing TS downgrade tests proving v2-only keys are removed and `schemaVersion` becomes `1`.
4. Only after the tests fail for the expected reason, add the Go migration constant + types first, then TS schema/downgrade changes.
5. Run package tests: `cd packages/ai-setup-go && go test ./internal/db/...`, `cd packages/ai-setup-ts && npx vitest run src/store/`

## Done When

- [ ] Red tests are committed/recorded before production changes.
- [ ] Go migration 007 adds nullable profile/coverage fields without destructive changes.
- [ ] TS v1 store data reads with safe v2 defaults and no crash.
- [ ] TS v2→v1 helper removes v2-only fields and preserves unrelated user data.
- [ ] Coverage threshold validation/default behavior is tested for missing, valid, and out-of-range values.
- [ ] Go-first implementation order is documented in the task evidence before TS parity is verified.
- [ ] No wizard/compiler/scaffold writes are introduced in this task.
- [ ] Test evidence cites commands run and passing output.

## Risks

- **v2 lowdb store crashes older strict TS CLIs on rollback:** mitigated by tested downgrade helper and backup/downgrade procedure.
- **TS/Go drift on fields and defaults:** mitigated by starting W1.A with schema validation/round-trip tests in both packages.

## Constitution Check

- **Article I:** Reuse existing Go embedded migration system and TS Zod schema surfaces; do not add a datastore.
- **Article II:** This task is the W1.A RED gate; tests come before schema writes.
- **Article III:** `spec.md` and `plan.md` define the contract (version numbers adjusted to match actual repo state: TS CURRENT_SCHEMA_VERSION is currently 1).
- **Article IV:** Do not implement W1.B `adversarialDesign` behavior here; only store the safe flag default.
- **Article V:** Prefer explicit optional fields over generic metadata maps per plan.
- **Article VI:** Do not extract new migration frameworks or rollback abstractions.
