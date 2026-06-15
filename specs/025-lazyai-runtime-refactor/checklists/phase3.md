# Checklist: Phase 3 V2 Schema Migration

**Phase:** Phase 3  
**Exit:** V1 runtime data migrates to V2 and can be restored from backup under test.

## Preconditions

- [x] Phase 2 verification output is recorded.
- [x] Human gate approves Phase 2 command and rollback evidence.
- [x] Orchestrator/runtime workflow packages are removed from active build paths.
- [x] `lazyai restore-runtime-db` command behavior is understood and covered by tests.

## RED Tests

- [ ] Synthetic FK-saturated V1 migration test fails before V2 migration exists.
- [ ] Empty database migration test fails before V2 migration exists.
- [ ] Legacy default migration test fails while `loop-driver` survives.
- [ ] Backup-restore round-trip test fails before backup/restore wiring is complete.


**Process note:** these RED checkpoints were not separately captured before same-session implementation. The implemented tests now cover each scenario, but no pre-green failing output was recorded.

## Implementation Checks

- [x] `runtime.SchemaV2` contains `sessions`, `dispatches`, `handoff`, `agent_defaults`, `ledger_refs`, and migration metadata.
- [x] `openRuntimeDB()` detects SchemaV1 before applying V2.
- [x] Migration creates `.specify/session.db.backup` before schema changes.
- [x] Migration fails safely and leaves the backup untouched on error.
- [x] V1 `sessions` data is preserved where supported by V2.
- [x] V1 `dispatches` data is preserved where supported by V2.
- [x] V1 `ledger_refs` data is preserved where supported by V2.
- [x] Legacy `loop-driver` defaults become `primary-agent` defaults in V2.
- [x] Dropped tables are explicitly dropped/replaced according to `schema-v2.md`.

## Verification

- [x] FK-saturated data migration passes.
- [x] Empty DB migration passes.
- [x] Legacy default migration passes.
- [x] Backup -> migrate -> restore -> verify round trip passes.
- [x] Session/dispatch/handoff round trip passes.
- [x] `go test ./packages/cli/internal/runtime ./packages/cli/internal/runtime/session ./packages/cli/cmd -run 'Schema|Migration|RuntimeDB|Restore|SessionManager|Dispatch' -count=1` passes; `go test ./packages/cli/...` and `go build ./packages/cli/...` also pass.

## Rollback Record

- [x] Equivalent tracked rollback point exists: commit `98e83d24cf5aefb64a51b3a6e3370bfac5369d4d` remains the coarse pre-implementation rollback point until a human-reviewed Phase 3 commit/tag exists.
- [x] `lazyai restore-runtime-db .specify/session.db.backup` verification output is recorded.
- [x] Developer rollback command from `rollback.md` is updated for touched migration files.


**Observed verification commands**

- `go test ./packages/cli/internal/runtime ./packages/cli/internal/runtime/session ./packages/cli/cmd -run 'Schema|Migration|RuntimeDB|Restore|SessionManager|Dispatch' -count=1` — passed.
- `go test ./packages/cli/...` — passed.
- `go build ./packages/cli/...` — passed.
- Migration failure path: backup preserved and V1 DB remained readable.
- Restore round trip: `.specify/session.db.pre-restore` preserved the migrated V2 DB while `.specify/session.db` returned to the pre-migration V1 schema/data.

## Gate

- [x] Human approves migration, backup, restore, and FK-saturated test evidence before Phase 4 handoff implementation begins.
