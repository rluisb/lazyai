# Tasks: 025 LazyAI Runtime Refactor

**Spec:** `specs/025-lazyai-runtime-refactor/spec.md`  
**Plan:** `specs/025-lazyai-runtime-refactor/plan.md`  
**Approval evidence:** commit `98e83d24cf5aefb64a51b3a6e3370bfac5369d4d` (`Approve LazyAI runtime refactor plan`)  
**Generated:** 2026-06-14

This task list starts after Phase 0 plan approval. It keeps Pi / Oh My Pi out of scope; adapter work covers only the registered LazyAI adapters: OpenCode, Claude Code, and Copilot.

## Rules

- Use RED -> GREEN -> REFACTOR for every behavior-changing code phase.
- Do not pass a human gate with failing targeted verification.
- Do not remove Fortnite/orchestrator code until rollback and migration notes are ready.
- No compatibility shim named `orchestrator` or `loop-driver` may remain as a hidden default.
- Every destructive removal must have a pre-refactor tag or operator rollback path.

## Phase B0 — Baseline Test Repair

These tasks are not part of the runtime refactor behavior. They restore the current repo to a useful verification baseline before adapter/runtime edits.

- [x] B0-1 Restore or relocate `.claude/skills/tui-lazy-ai-design-system/colors_and_type.css` so `TestCSSTokenParity` can read a tracked design-system fixture.
- [x] B0-2 Run `go test ./packages/cli/internal/theme -run TestCSSTokenParity -count=1` and record passing output.
- [x] B0-3 Run `go test ./packages/cli/...` and record the baseline before Phase 1 edits.

## Phase 1 — Adapter Decouple and Test Rewrite

Exit: adapter behavior is neutral, tested, and no longer depends on Fortnite assumptions.

- [x] P1-1 Add adapter contract test fixtures for OpenCode, Claude Code, and Copilot using `primary-agent` as the default.
- [x] P1-2 Rewrite OpenCode adapter tests to assert canonical files, canonical agents, no `FortniteMode`, no `STARTUP.md`, no `loop-driver`, and no `orchestrator` default.
- [x] P1-3 Rewrite Claude Code adapter tests to assert canonical files, canonical agents, and no orchestrator/Fortnite references.
- [x] P1-4 Rewrite Copilot adapter tests to assert canonical files, canonical agents, and no orchestrator/Fortnite references.
- [x] P1-5 Replace OpenCode non-Fortnite and Fortnite default branches with the redesigned `primary-agent` path.
- [x] P1-6 Remove `FortniteMode` from adapter context construction and generated config behavior.
- [x] P1-7 Replace `loop-driver` defaults with `primary-agent` in `packages/cli/cmd/session.go`, `packages/cli/internal/runtime/session/session.go`, and runtime schema defaults until V2 removes the schema default.
- [x] P1-8 Replace `orchestrator` defaults with `primary-agent` in `packages/cli/cmd/config.go` and related config tests.
- [x] P1-9 Verify `go test ./packages/cli/internal/adapter ./packages/cli/internal/runtime/session ./packages/cli/cmd -run 'Adapter|OpenCode|Claude|Copilot|Session|Config|Helpers|Init' -count=1`.
- [x] P1-10 Gate record: Phase 1 test output is saved in `checklists/phase1.md`; rollback uses coarse tracked point `98e83d24cf5aefb64a51b3a6e3370bfac5369d4d` because no clean Phase 1 commit/tag boundary exists.

## Phase 2 — CLI Command Rewrite and Excision

Exit: affected CLI commands are rewritten or removed, heavy orchestration packages are gone, and user-facing command behavior is covered.

- [x] P2-1 Complete qualitative Fortnite/OpenCode outreach or record an explicit human waiver before excision.
- [x] P2-2 Create `docs/migration/fortnite-orchestrator-removal.md` with user-facing replacements for FortniteMode, `loop-driver`, and orchestrator paths.
- [x] P2-3 Add failing tests for `lazyai update-self --version <tag>` using tag-shaped releases such as `packages/orchestrator/v1.0.0`.
- [x] P2-4 Implement tag-specific release lookup without corrupting slash-containing release tags.
- [x] P2-5 Rewrite `packages/cli/cmd/task.go` away from `runtime/taskqueue`.
- [x] P2-6 Rewrite `packages/cli/cmd/workflow.go` and `workflow_test.go` away from `runtime/workflow` and `runtime/dispatch`.
- [x] P2-7 Rewrite or remove `packages/cli/cmd/orchestration.go` according to the approved migration note.
- [x] P2-8 Rewrite `mcp_setup.go`, `server.go`, `doctor_mcp.go`, `doctor_health.go`, `message.go`, `add.go`, `validate_input.go`, `update.go`, and `init.go` to remove orchestrator/Fortnite assumptions.
- [x] P2-9 Remove `packages/orchestrator/` from `go.work` and remove package references from build/test tooling.
- [x] P2-10 Archive `packages/cli/library/fortnite/` to the approved archive path and remove `all:fortnite` from active embedding.
- [x] P2-11 Remove `packages/cli/internal/orchestrator/`.
- [x] P2-12 Remove runtime `workflow/`, `taskqueue/`, `dispatch/`, and Fortnite session coordination files listed in `plan.md`.
- [x] P2-13 Search remaining code/docs for stale `orchestrator`, `loop-driver`, `FortniteMode`, `library/fortnite`, `runtime/workflow`, `runtime/taskqueue`, and `runtime/dispatch` references; direct active imports/defaults are gone, and remaining hits are justified in `checklists/phase2.md`.
- [x] P2-14 Verify affected command tests and record the root `go build ./...` workspace limitation.
- [x] P2-15 Gate record: Phase 2 verification output, equivalent rollback point, and operator rollback command evidence are saved before Phase 3; human Phase 2 approval remains the active gate.

## Phase 3 — V2 Schema Migration

Exit: V1 runtime data migrates to V2 and can be restored from backup under test.

- [x] P3-1 Add migration tests for FK-saturated V1 data, empty DB, legacy defaults, and backup-restore round trip.
- [x] P3-2 Implement `runtime.SchemaV2` in `packages/cli/internal/runtime/schema.go` with `sessions`, `dispatches`, `handoff`, `agent_defaults`, `ledger_refs`, and migration metadata.
- [x] P3-3 Update `openRuntimeDB()` to detect V1, create `.specify/session.db.backup`, apply V2, and verify the resulting schema.
- [x] P3-4 Preserve required V1 session, dispatch, and ledger data; fail safely with untouched backup on migration errors.
- [x] P3-5 Verify `loop-driver` legacy defaults migrate to `primary-agent`.
- [x] P3-6 Verify `lazyai restore-runtime-db .specify/session.db.backup` restores the pre-migration DB.
- [x] P3-7 Verify migration round-trip tests and runtime helper tests.
- [x] P3-8 Gate record: backup/restore evidence is saved in `checklists/phase3.md`; human approval remains required before Phase 4.

## Phase 4 — Handoff Implementation

Exit: runtime writes handoffs matching the approved markdown schema.

- [x] P4-1 Add handoff writer tests for frontmatter keys, required sections, path convention, and atomic replace behavior.
- [x] P4-2 Implement `packages/cli/internal/handoff/writer.go`.
- [x] P4-3 Implement a minimal reader/parser inside `packages/cli/internal/handoff/writer.go` for round-trip tests.
- [x] P4-4 Wire handoff write to session-close lifecycle, one handoff per session end.
- [x] P4-5 Store handoff metadata in the V2 `handoff` table.
- [x] P4-6 Verify write -> read -> parse round trip and repeated close replacement.
- [x] P4-7 Gate record: handoff round-trip evidence is saved in `checklists/phase4.md`; human approval remains required before Phase 5.

## Phase 5 — Canonical Library Curation and Token-Rent Enforcement

Exit: canonical library stays within 50KB or fails with documented override instructions.

- [x] P5-1 Populate `packages/cli/library/canonical/agents/` with `primary-agent`, `builder`, `planner`, `reviewer`, and `scout`.
- [x] P5-2 Populate canonical skills, hooks, and commands from `library-canonical.md`.
- [x] P5-3 Update `packages/cli/library/embed.go` to embed `all:canonical` and stop embedding active Fortnite content.
- [x] P5-4 Update adapter file-generation paths to read canonical assets.
- [x] P5-5 Add token-rent tests for under-budget, over-budget, valid override, and invalid override cases.
- [x] P5-6 Implement `packages/cli/internal/tokenrent/check.go` using recursive byte counts over `packages/cli/library/canonical/`.
- [x] P5-7 Add pre-commit token-rent enforcement with the approved failure message.
- [x] P5-8 Add `.github/workflows/token-rent.yml` CI enforcement.
- [x] P5-9 Implement `.lazyai/token-rent-override` support requiring a non-empty `reason:` field.
- [x] P5-10 Verify over-budget failure, valid override pass, invalid override failure, and normal `go test ./packages/cli/...`; root `go build ./...` limitation was rechecked and preserved from Phase 2.
- [x] P5-11 Gate record: token-rent evidence is saved in `checklists/phase5.md`; final human approval remains required before the refactor is declared complete.
