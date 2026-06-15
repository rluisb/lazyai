# Checklist: Phase 2 CLI Command Rewrite and Excision

**Phase:** Phase 2  
**Exit:** Affected CLI commands are rewritten or removed, heavy orchestration packages are gone, and user-facing command behavior is covered.

## Preconditions

- [x] Phase 1 verification output is recorded.
- [x] Human gate approves Phase 1 adapter evidence.
- [x] Qualitative Fortnite/OpenCode outreach is complete or explicitly waived by a human gate.
- [x] `docs/migration/fortnite-orchestrator-removal.md` exists before user-facing removals.
- [x] `lazyai update-self --version <tag>` tag-specific fetch tests are written before implementation.

**Waiver evidence:** 2026-06-14 interactive human approval in this session — `Go ahead`
## RED Tests

- [ ] `update-self --version packages/orchestrator/v1.0.0` fails before tag-specific release fetch is implemented.
- [ ] Command tests fail for direct dependencies on `runtime/taskqueue`, `runtime/workflow`, `runtime/dispatch`, and `internal/orchestrator`.
- [ ] Doctor/MCP tests fail while legacy orchestrator MCP detection remains.

## Implementation Checks

- [x] `task.go` no longer imports `runtime/taskqueue`.
- [x] `workflow.go` and `workflow_test.go` no longer import `runtime/workflow` or `runtime/dispatch`.
- [x] `orchestration.go` is rewritten or removed with migration notes.
- [x] `mcp_setup.go`, `server.go`, `doctor_mcp.go`, `doctor_health.go`, `message.go`, `add.go`, `validate_input.go`, `update.go`, and `init.go` no longer preserve orchestrator/Fortnite assumptions.
- [x] `packages/orchestrator/` is removed from `go.work`.
- [x] `packages/cli/internal/orchestrator/` is removed.
- [x] Runtime `workflow/`, `taskqueue/`, `dispatch/`, and Fortnite session coordination files are removed after callers are clean.
- [x] `packages/cli/library/fortnite/` is archived to the approved path and no longer actively embedded.
- [x] `packages/cli/library/embed.go` no longer embeds active `all:fortnite` content.

## Verification

- [x] Focused command tests for rewritten task/workflow/orchestration paths pass.
- [x] Focused `update-self --version <tag>` tests pass for slash-containing release tags.
- [x] `go build ./packages/cli/...` passes for the active CLI workspace.
- [x] Root `go build ./...` was rechecked and remains invalid in this repo's `go.work` layout: `directory prefix . does not contain modules listed in go.work or their selected dependencies`.
- [x] Search confirms no stale active command/build references remain for `internal/orchestrator`, `packages/orchestrator`, `runtime/workflow`, `runtime/taskqueue`, `runtime/dispatch`, `library/fortnite`, `FortniteMode`, or `loop-driver`.
- [x] Remaining generic `orchestrator` text is intentional: migration/rollback docs, historical design docs, archived reports/spec evidence, legacy cleanup paths, regression tests that assert retired agent files are not generated, arbitrary MCP passthrough/setup-scan tests, and Phase 5 library-curation inputs currently filtered out of adapter output.

**Remaining reference note:** active direct imports/defaults are gone. `packages/cli/library/agents/orchestrator.md`, `packages/cli/library/copilot/agents/orchestrator.agent.yaml`, and `packages/cli/library/skills/orchestrate.md` still exist only as pre-Phase-5 library-curation inputs; adapter output filters exclude them from OpenCode, Claude Code, and Copilot installs.

**Observed verification commands**

- `go test ./packages/cli/...` — passed.
- `go build ./packages/cli/...` — passed.
- `go build ./...` — rechecked and still fails in this repo's existing `go.work` layout with `directory prefix . does not contain modules listed in go.work or their selected dependencies`.
- `python3 -m mkdocs build --strict` — passed; emitted upstream MkDocs 2.0 warning and nav-info messages.
- Search over `packages/cli` + `go.work` for `internal/orchestrator|packages/orchestrator|runtime/workflow|runtime/taskqueue|runtime/dispatch|library/fortnite|FortniteMode|loop-driver` — no matches.

## Rollback Record

- [x] Equivalent tracked rollback point exists: commit `98e83d24cf5aefb64a51b3a6e3370bfac5369d4d` (`Approve LazyAI runtime refactor plan`) is the last tracked pre-implementation point.
- [x] Operator rollback via `lazyai update-self --version <tag>` is verified.
- [x] Developer rollback command from `rollback.md` is updated for the no-tag local boundary.
- [x] Phase 2 verification output is recorded.

**Rollback boundary note:** no clean tracked Phase 1/Phase 2 commit boundary exists; current refactor work is local/uncommitted. Do not create a misleading `pre-refactor-025-phase-2` tag after the fact. Use commit `98e83d24cf5aefb64a51b3a6e3370bfac5369d4d` as the coarse rollback point, or create a human-reviewed implementation commit/tag before any further destructive phase.

## Gate

- [x] Human approves CLI audit dispositions, command test evidence, and rollback evidence before Phase 3 migration begins (`Go ahead`, 2026-06-14).
