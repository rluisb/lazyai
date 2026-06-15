# P0-8: Rollback Procedure

**Status:** Partial — update-self tag rollback is verified; DB restore command exists; Phase 1/2 use a coarse tracked rollback point because no clean local phase-boundary commit exists
**Owner:** Ricardo Conceicao  
**Date:** 2026-06-14  
**Linked from:** `plan.md` Phase 0, P0-8 (this document + `plan.md`)

---

## Purpose

Rollback procedure per phase — developer and operator paths. Every destructive phase must have a verified rollback that can be performed without reading source code.

## Rollback Command Readiness

| Command | File | Status |
|---|---|---|
| `lazyai update-self --version <tag>` | `cmd/update-self.go` | Verified for exact tag lookup, including slash-containing tags |
| `lazyai restore-runtime-db <backup-path>` | `cmd/restore_runtime_db.go` | Implemented — restores `.specify/session.db` from backup |

## Phase Rollback Matrix

| Phase | Trigger | Developer rollback | Operator rollback | Verification |
|---|---|---|---|---|
| **Phase 1** — Adapter decouple | Adapter tests fail against neutral contract | `git checkout 98e83d24cf5aefb64a51b3a6e3370bfac5369d4d -- packages/cli/internal/adapter/ packages/cli/cmd/session.go packages/cli/internal/runtime/session/session.go packages/cli/internal/runtime/schema.go` | N/A — adapter changes are dev-only, not in binary | `go test ./packages/cli/internal/adapter ./packages/cli/internal/runtime/session ./packages/cli/cmd -run 'Adapter|OpenCode|Claude|Copilot|Session|Config|Helpers|Init' -count=1` |
| **Phase 2** — CLI excision | Command breakage discovered post-excision | `git checkout 98e83d24cf5aefb64a51b3a6e3370bfac5369d4d -- packages/cli/cmd/ packages/orchestrator/ packages/cli/internal/orchestrator/ packages/cli/library/fortnite/ packages/cli/internal/runtime/workflow/ packages/cli/internal/runtime/taskqueue/ packages/cli/internal/runtime/dispatch/ packages/cli/internal/runtime/session/ packages/cli/library/embed.go go.work` | `lazyai update-self --version <known-good-release-tag>`; exact tag fetch is verified, but the release/tag must exist in GitHub Releases | Affected command tests + `go test ./packages/cli/...` + `go build ./packages/cli/...` |
| **Phase 3** — Schema migration | V2 migration corrupts data | `git checkout 98e83d24cf5aefb64a51b3a6e3370bfac5369d4d -- packages/cli/internal/runtime/schema.go packages/cli/cmd/runtime_helper.go packages/cli/cmd/runtime_helper_test.go packages/cli/internal/runtime/session/dispatch.go packages/cli/internal/runtime/session/session_test.go` | `lazyai restore-runtime-db .specify/session.db.backup` | `go test ./packages/cli/internal/runtime ./packages/cli/internal/runtime/session ./packages/cli/cmd -run 'Schema|Migration|RuntimeDB|Restore|SessionManager|Dispatch' -count=1` + `go build ./packages/cli/...` |
| **Phase 4** — Handoff | Handoff format incompatible | `git checkout 98e83d24cf5aefb64a51b3a6e3370bfac5369d4d -- packages/cli/internal/handoff/ packages/cli/cmd/session.go packages/cli/cmd/session_test.go` | `lazyai update-self --version pre-refactor-025-phase-4` after a human-reviewed Phase 4 release/tag exists | `go test ./packages/cli/internal/handoff ./packages/cli/cmd -run 'Handoff|SessionEndWritesHandoff' -count=1` |
| **Phase 5** — Library curation | Canonical library or token-rent gate breaks installs | `git checkout 98e83d24cf5aefb64a51b3a6e3370bfac5369d4d -- packages/cli/library/canonical/ packages/cli/library/embed.go packages/cli/internal/tokenrent/ packages/cli/internal/adapter/ packages/cli/internal/compiler/agent_validate.go packages/cli/internal/compiler/contract_validator.go packages/cli/internal/library/embed.go packages/cli/internal/library/integration_test.go packages/cli/internal/types/types.go packages/cli/cmd/add.go .githooks/pre-commit .github/workflows/token-rent.yml` | `lazyai update-self --version pre-refactor-025-phase-5` after a human-reviewed Phase 5 release/tag exists | `go run ./packages/cli/internal/tokenrent/cmd/token-rent-check` + `go test ./packages/cli/...` + `go build ./packages/cli/...` |

## Tagging Policy

- Tag before each future destructive phase: `pre-refactor-025-phase-N`.
- Archive tag for removed packages: `archive/pre-refactor-025-orchestrator`.
- Tags are immutable; no force-push after tagging.
- Phase 1/2 local work has no clean phase-boundary commit. Do not create a misleading `pre-refactor-025-phase-2` tag after the fact; use `98e83d24cf5aefb64a51b3a6e3370bfac5369d4d` as the coarse rollback point or create a human-reviewed implementation commit/tag before Phase 3.

## Verification Per Phase

| Phase | Verification Command |
|---|---|
| Phase 1 | `go test ./packages/cli/internal/adapter ./packages/cli/internal/runtime/session ./packages/cli/cmd -run 'Adapter|OpenCode|Claude|Copilot|Session|Config|Helpers|Init' -count=1` |
| Phase 2 | Affected command tests + `go test ./packages/cli/...` + `go build ./packages/cli/...`; root `go build ./...` is not a valid verification command in the current `go.work` layout |
| Phase 3 | `go test ./packages/cli/internal/runtime ./packages/cli/internal/runtime/session ./packages/cli/cmd -run 'Schema|Migration|RuntimeDB|Restore|SessionManager|Dispatch' -count=1` + `go test ./packages/cli/...` + `go build ./packages/cli/...` |
| Phase 4 | `go test ./packages/cli/internal/handoff ./packages/cli/cmd -run 'Handoff|SessionEndWritesHandoff' -count=1` + `go test ./packages/cli/...` + `go build ./packages/cli/...` |
| Phase 5 | `go test ./packages/cli/internal/tokenrent -run 'TestCheckPassesUnderBudget|TestCheckFailsOverBudgetWithoutOverride|TestCheckPassesWithValidOverride|TestCheckFailsWithInvalidOverrideReason|TestCheckExcludesGitkeepFromBudget|TestCurrentCanonicalLibraryWithinBudget' -count=1` + `go run ./packages/cli/internal/tokenrent/cmd/token-rent-check` + `go test ./packages/cli/...` + `go build ./packages/cli/...`; root `go build ./...` remains invalid in the current `go.work` layout |

## Gate

⛔ Human must approve this rollback procedure before any future destructive phase begins.
