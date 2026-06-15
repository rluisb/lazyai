# Analysis: 025 LazyAI Runtime Refactor

**Spec:** `specs/025-lazyai-runtime-refactor/spec.md`  
**Plan:** `specs/025-lazyai-runtime-refactor/plan.md`  
**Tasks:** `specs/025-lazyai-runtime-refactor/tasks.md`  
**Generated:** 2026-06-14

This analysis records the dependency order, unresolved risks, and gate conditions for the approved runtime refactor. It does not expand scope beyond the approved adapters: OpenCode, Claude Code, and Copilot. Pi / Oh My Pi is intentionally deferred.

## Current State

- Phase 0 artifacts are tracked by approval commit `98e83d24cf5aefb64a51b3a6e3370bfac5369d4d`.
- Code implementation has not started in this analysis.
- The repo still contains Fortnite/orchestrator defaults and imports.
- `packages/cli/library/canonical/agents/primary-agent.md` exists, but canonical embedding is not wired.
- The current CLI package baseline has a known unrelated failure: `TestCSSTokenParity` cannot find `.claude/skills/tui-lazy-ai-design-system/colors_and_type.css`.

## Scope Decisions

| Decision | Status | Consequence |
|---|---|---|
| LazyAI owns runtime; vibe-lab supplies principles/expectations | Approved by ADR-003 commit | Do not port vibe-lab runtime ownership into LazyAI. |
| Registered adapters only | Approved scope | OpenCode, Claude Code, Copilot are in; Gemini, Codex, Pi, Oh My Pi are out. |
| Redesigned `primary-agent` path | Approved | No `orchestrator.md` shim; remove orchestrator concept instead. |
| Token-rent budget is 50,000 bytes | Approved | Canonical library growth must be enforced by CI/pre-commit. |
| Backup-restore is rollback for DB migration | Approved | No SQL down migration required; restore command must be verified. |

## Dependency Graph

```text
B0 baseline test repair
  -> Phase 1 adapter contract tests and default rewrites
      -> Phase 2 command rewrite, rollback tag fetch, orchestration excision
          -> Phase 3 V2 schema migration and restore verification
              -> Phase 4 handoff writer and session integration
      -> Phase 5 canonical library curation and token-rent enforcement
```

Phase 5 depends on Phase 2 because Fortnite library embedding/removal must be resolved before canonical assets become the only active library path.

## Risk Register

| Risk | Phase | Likelihood | Impact | Control |
|---|---:|---|---|---|
| Baseline CLI tests remain red for unrelated theme fixture | B0 | High | Medium | Restore tracked CSS fixture before refactor tests. |
| Adapter tests accidentally preserve Fortnite behavior | 1 | Medium | High | Contract fixtures assert `primary-agent` and reject `FortniteMode`, `STARTUP.md`, `loop-driver`, `orchestrator`. |
| `FortniteMode` removal breaks OpenCode users silently | 1-2 | Medium | High | Migration note plus OpenCode contract tests; Phase 2 outreach/waiver gate. |
| Existing command depends on removed runtime package | 2 | Medium | High | Use P0 audit disposition; no package deletion before affected command tests pass. |
| `update-self --version <tag>` cannot fetch slash-containing release tags | 2 | High | High | Add tag endpoint tests before destructive removal; preserve exact tag string. |
| Orchestrator removal leaves build-system references | 2 | Medium | High | Remove `go.work`, Makefile/CI/bin references and verify `go build ./...`. |
| V2 migration drops data needed by active runtime flows | 3 | Medium | Critical | FK-saturated migration tests, backup-before-migrate, restore verification. |
| Backup restore overwrites wrong path | 3 | Low-Medium | Critical | Tests must use temp homes/targets and assert restored `.specify/session.db` bytes/schema. |
| Handoff writer appends duplicate sections on repeated close | 4 | Medium | Medium | Atomic replace test for same-session repeated close. |
| Canonical library exceeds 50KB after curation | 5 | Medium | Medium | Token-rent check with failure message and explicit override file. |
| Deferred Pi / Oh My Pi adapter becomes expected behavior | Cross | Low | Medium | Record out-of-scope decision here and in adapter contract changes if later reopened. |

## Gate Dependencies

### Before Phase 1

- Approval commit exists for ADR/spec/plan/P0 artifacts.
- `tasks.md`, `analysis.md`, and `checklists/` exist.
- Baseline theme fixture repair is either complete or explicitly separated from Phase 1 verification output.

### Before Phase 2

- Phase 1 adapter tests pass.
- Human gate approves adapter contract evidence.
- Qualitative Fortnite/OpenCode outreach is complete, or a human waiver is recorded.
- Migration note path is ready: `docs/migration/fortnite-orchestrator-removal.md`.

### Before Phase 3

- Phase 2 command rewrites pass.
- `lazyai update-self --version <tag>` is verified against tag-shaped GitHub Releases.
- Build system no longer references `packages/orchestrator/`.
- Heavy orchestration package removal is covered by tests and rollback tag.

### Before Phase 4

- V2 migration tests pass for FK-saturated data, empty DB, legacy defaults, and backup restore.
- `lazyai restore-runtime-db` evidence exists.

### Before Phase 5

- Handoff writer round-trip passes.
- Session-close integration writes exactly one handoff per session close and replaces atomically for repeats.

### Completion Gate

- `go test ./packages/cli/...` passes.
- `go build ./...` passes after orchestrator module removal.
- Token-rent CI/pre-commit checks fail over budget and pass with a valid documented override.
- Rollback records exist for every destructive phase.

## Verification Strategy

| Layer | Command or Evidence |
|---|---|
| Baseline theme | `go test ./packages/cli/internal/theme -run TestCSSTokenParity -count=1` |
| Adapter contract | `go test ./packages/cli/internal/adapter -run 'OpenCode|Claude|Copilot|Adapter' -count=1` |
| Session defaults | `go test ./packages/cli/internal/runtime/session ./packages/cli/cmd -run 'Session|Config|Helpers|Init' -count=1` |
| CLI commands | Focused `packages/cli/cmd` tests for rewritten task/workflow/orchestration/update-self paths |
| Migration | Runtime helper/schema tests with synthetic V1 DBs |
| Handoff | `packages/cli/internal/handoff` round-trip tests plus session-close integration |
| Token-rent | Unit tests for byte counting and override validation plus CI/pre-commit dry runs |
| Full package | `go test ./packages/cli/...` |
| Workspace build | `go build ./...` after `go.work` no longer includes orchestrator |

## Open Items

- Qualitative Fortnite/OpenCode outreach remains pending unless the human approval trail explicitly waives it before Phase 2.
- The missing design-system CSS fixture must be repaired before CLI-wide tests are meaningful.
- `docs/migration/fortnite-orchestrator-removal.md` must be created before command removal is user-facing.
- `update-self --version <tag>` must be fixed before operator rollback can be claimed.
