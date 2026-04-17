# 006 — Housekeeping, Memory, and Bootstrap: Implementation Plan

**Date:** 2026-04-17
**Status:** Ready for Implementation
**Scope:** Turn Spec 006 approval artifacts into concrete code changes across the ai-setup repo
**Out of Scope:** No new spec/docs writing; no changes to qmd/codegraph internals; no database/store migration (future spec)

---

## Approved Defaults Carried Forward

| # | Default | Implementation Implication |
|---|---------|--------------------------|
| D-1 | Sync state path `.ai/housekeeping/sync-state.json` | Go code creates/reads `.ai/housekeeping/` beneath project root; orchestrator TS reads it |
| D-2 | Strict for new artifacts, warn/migrate for legacy | `internal/validation` gains metadata schema; doctor reports legacy gaps without blocking |
| D-3 | `compose_agent` API assumption first | No new fields on `ComposeAgentInput`; inject via existing `stepInstructions` |
| D-4 | Default task-scoped approval; session-scoped optional; standing 30-day hard-expire; in-flight may finish | Contract model in orchestrator TS; CLI wizard describes options |
| D-5 | Doctor report-only for stray `AGENTS.md` | Doctor check adds a finding section; no deletion flag wired |
| D-6 | qmd index project-local by default | Wizard and scaffold write project-local config; no global cache paths |

---

## Repo Architecture Snapshot (Implementation-Relevant)

```
ai-setup/
├── cmd/                    ← Go CLI commands (init, doctor, update, compile, …)
├── internal/
│   ├── frontmatter/        ← YAML frontmatter parsing (will gain schema validation)
│   ├── validation/         ← Artifact name/class validation (will gain metadata schema enforcement)
│   ├── scaffold/           ← File scaffolding engine (will stop generating sub-AGENTS.md)
│   ├── types/              ← Core domain types (will gain housekeeping/metadata types)
│   ├── compiler/           ← Template/compiler (unchanged, but reads from frontmatter)
│   ├── db/                 ← SQLite store (may gain housekeeping state tables later)
│   └── adapter/            ← Tool-specific config compilers (unchanged)
├── tui/
│   ├── wizard/             ← 4-phase interactive wizard (will gain Phase 5: optional tooling)
│   └── components/         ← Shared TUI primitives
├── orchestrator/
│   └── src/
│       ├── server.ts       ← MCP tool registration (will add housekeeping/bootstrap hooks)
│       ├── tool-handlers.ts ← Compose/start/advance (will inject bootstrap context)
│       ├── composer.ts      ← Prompt layer composition (unchanged API)
│       ├── chain-machine.ts ← State machine (will call bootstrap/housekeeping hooks)
│       ├── persistence.ts   ← State persistence (will add sync-state helpers)
│       └── types.ts         ← Orchestrator type defs (will add housekeeping types)
├── library/
│   └── specs-agents/       ← Replacement source for sub-AGENTS.md (already exists)
├── .ai/
│   └── orchestration/      ← Existing state dir (housekeeping state joins here)
└── specs/
    ├── adrs/AGENTS.md       ← Stray — will be removed after migration
    └── features/AGENTS.md   ← Stray — will be removed after migration
```

---

## Execution Waves

### Wave 0 — Foundation Types (Go + TS)

**Goal:** Introduce the Go types and TS types that all later waves depend on. No behavior changes yet.

| Task | Layer | Files Touched | Depends On | Acceptance |
|------|-------|-------------|------------|------------|
| **W0.1** Go: add `Spec006Metadata` struct | Go CLI | `internal/types/types.go` | — | Struct mirrors data-model.md §1.3 core fields; `ArtifactType` gains `memory_note`, `maintenance_contract`, `sync_state_snapshot` |
| **W0.2** Go: add `SyncState` struct | Go CLI | `internal/types/types.go` | W0.1 | Struct mirrors data-model.md §4.5 sync-state JSON schema |
| **W0.3** Go: add `MaintenanceContract` struct | Go CLI | `internal/types/types.go` | W0.1 | Struct mirrors data-model.md §3.3 contract shape; includes `ApprovalScope` enum |
| **W0.4** Go: add `MemoryEntry` struct | Go CLI | `internal/types/types.go` | W0.1 | Struct mirrors data-model.md §2.2 timeline entry shape |
| **W0.5** TS: add housekeeping types | Orchestrator | `orchestrator/src/types.ts` | — | `BootstrapReport`, `HousekeepingReport`, `MaintenanceContractRecord`, `SyncStateSnapshot` types added |
| **W0.6** Go: metadata validation function | Go CLI | `internal/validation/validation.go`, `internal/validation/validation_test.go` | W0.1 | `ValidateSpec006Metadata(fm map[string]any) []string` returns warnings/errors; strict for new, warn for legacy |
| **W0.7** Go: frontmatter helper for schema_version | Go CLI | `internal/frontmatter/frontmatter.go`, `internal/frontmatter/frontmatter_test.go` | — | `ExtractSchemaVersion(content []byte) (int, error)` and `HasSpec006Metadata(content []byte) bool` |

**Parallel:** W0.1-W0.4 [P] independent of W0.5. W0.6-W0.7 depend on W0.1.
**Risk:** Low. Additive types only; no breaking changes.

---

### Wave 1 — Doctor: Stray AGENTS.md Detection + Metadata Gap Reporting

**Goal:** Doctor reports two new finding categories without mutating anything.

| Task | Layer | Files Touched | Depends On | Acceptance |
|------|-------|-------------|------------|------------|
| **W1.1** Go: doctor stray AGENTS.md check | Go CLI | `cmd/doctor.go` | — | Walk `specs/**` for `AGENTS.md` files; report as `stray_agents_md` findings; root `AGENTS.md` and `.opencode/agents/` excluded; report-only always; JSON output gains `strayAgentsFiles` array |
| **W1.2** Go: doctor metadata gap check | Go CLI | `cmd/doctor.go`, `internal/frontmatter/frontmatter.go` | W0.6, W0.7 | For each `specs/**/*.md` file, check if spec006 metadata is absent or incomplete; report as `metadata_gaps` finding with severity `warning` for legacy, `error` for new artifacts (create-after-spec006 detection heuristic: check `created_at` year >= 2026 and `schema_version` absent); output in JSON and styled terminals |
| **W1.3** Go: doctor test coverage | Go CLI | `cmd/doctor_test.go` (new or extend existing) | W1.1, W1.2 | Tests for: stray detection with fixture files; metadata gap detection; report-only enforcement; JSON output format correctness |

**Parallel:** W1.1 and W1.2 [P] independent of each other within the wave.
**Risk:** Medium. Doctor currently only checks file integrity; adding structural checks is new territory. Existing test patterns in `cmd/*_test.go` show the style.

---

### Wave 2 — Scaffold: Stop Generating Subdirectory AGENTS.md

**Goal:** Prevent future sprawl at the source. Scaffold no longer writes `specs/**/AGENTS.md` for new installs.

| Task | Layer | Files Touched | Depends On | Acceptance |
|------|-------|-------------|------------|------------|
| **W2.1** Go: remove specs-agents copy from scaffold | Go CLI | `internal/scaffold/specs.go` | — | Remove the loop that copies `library/specs-agents/<dir>.md` → `specs/<dir>/AGENTS.md`; delete `SpecsAgents` field from `ScaffoldContext` if present; `ScaffoldSpecs` only creates directories, not AGENTS.md files |
| **W2.2** Go: update scaffold context types | Go CLI | `internal/scaffold/types.go` | W2.1 | Remove `SpecsAgents []string` from `ScaffoldContext` |
| **W2.3** Go: update wizard/planner references | Go CLI | `tui/wizard/planner.go` (or equivalent file computing `SpecsAgents`) | W2.1 | Remove any code that populates `SpecsAgents` |
| **W2.4** Go: update `cmd/init.go` buildScaffoldContext | Go CLI | `cmd/init.go` | W2.1 | Remove `SpecsAgents` from scaffold context construction |
| **W2.5** Go: scaffold test update | Go CLI | `internal/scaffold/scaffold_test.go` | W2.1-W2.4 | Existing scaffold tests pass; new test asserts `ScaffoldSpecs` does not create `AGENTS.md` files in specs subdirectories |
| **W2.6** Go: non-interactive init regression | Go CLI | `cmd/init_test.go` | W2.4 | Existing non-interactive init tests pass without `SpecsAgents` |

**Parallel:** Sequential within wave (W2.1 first, others follow).
**Risk:** Low-Medium. Removing a feature is simpler than adding one. The `library/specs-agents/` files remain — they just stop being copied into `specs/**/AGENTS.md`. Tests may break if they assert on those files.

**Clarification needed before coding:** Should existing installs that already have `specs/<dir>/AGENTS.md` tracked in their store have those records removed on `update`? Recommend: yes, add a migration step in Wave 5.

---

### Wave 3 — AGENTS.md Migration: Remove Live Stray Files

**Goal:** Delete the two known stray files after verifying library content covers them.

| Task | Layer | Files Touched | Depends On | Acceptance |
|------|-------|-------------|------------|------------|
| **W3.1** Compare stray content vs library | Manual/Doc | `specs/adrs/AGENTS.md` vs `library/specs-agents/adrs.md`; `specs/features/AGENTS.md` vs `library/specs-agents/features.md` | — | Unique guidance from stray files is reflected in library fragments; no normative content lost |
| **W3.2** Backfill any missing content into library | Go CLI / Library | `library/specs-agents/adrs.md`, `library/specs-agents/features.md` | W3.1 | Library files contain all unique normative rules from their stray counterparts |
| **W3.3** Delete stray files | Repo | Delete `specs/adrs/AGENTS.md`, `specs/features/AGENTS.md` | W3.2 | Files removed; doctor check (W1.1) should now report 0 stray files |
| **W3.4** Update store records for existing installs | Go CLI | `cmd/update.go` (or migration in `cmd/migrate.go`) | W2.1, W3.3 | `ai-setup update` removes `specs/<dir>/AGENTS.md` entries from tracked files in the store if they match the library source; this is a mutating action requiring `--confirm` or user approval in interactive mode |

**Parallel:** W3.1 must complete before W3.2. W3.4 depends on W2.1 (scaffold stops generating) and W3.3.
**Risk:** Low. Content preserved in library. Deletion is human-approved.

---

### Wave 4 — Orchestrator: Bootstrap Hook at Chain Start

**Goal:** When `start_chain` is called, the orchestrator runs a bootstrap sequence and attaches the report to chain state.

| Task | Layer | Files Touched | Depends On | Acceptance |
|------|-------|-------------|------------|------------|
| **W4.1** TS: add `runBootstrap` function | Orchestrator | `orchestrator/src/bootstrap.ts` (new) | W0.5 | Implements the 5-step bootstrap sequence from Spec 006 Task 002: discovery → drift check → approval eval → context load → report. Returns `BootstrapReport`. Pure function, no side effects until caller decides. |
| **W4.2** TS: add `readSyncState` helper | Orchestrator | `orchestrator/src/persistence.ts` | W0.5 | `readSyncState(projectRoot)` reads `.ai/housekeeping/sync-state.json` if present; returns `SyncStateSnapshot | null` |
| **W4.3** TS: add `writeSyncState` helper | Orchestrator | `orchestrator/src/persistence.ts` | W0.5 | `writeSyncState(projectRoot, state)` writes `.ai/housekeeping/sync-state.json`; creates `.ai/housekeeping/` dir if absent. **Note:** This is a mutating action — callers must check contract/approval before calling. |
| **W4.4** TS: add `readMaintenanceContracts` helper | Orchestrator | `orchestrator/src/persistence.ts` | W0.5 | Contracts stored in `.ai/housekeeping/contracts/` as JSON; `readMaintenanceContracts(projectRoot)` returns active contracts; expired ones filtered out |
| **W4.5** TS: wire `start_chain` to call bootstrap | Orchestrator | `orchestrator/src/tool-handlers.ts` | W4.1, W4.2, W4.4 | `startChain` handler calls `runBootstrap` before creating chain state; attaches `BootstrapReport` to chain state under `bootstrapReport` field; bootstrap failure is non-blocking (logged, chain still created) |
| **W4.6** TS: add `bootstrapReport` to `ChainState` type | Orchestrator | `orchestrator/src/types.ts` | W0.5 | Optional `bootstrapReport?: BootstrapReport` on `ChainState` |
| **W4.7** TS: bootstrap tests | Orchestrator | `orchestrator/src/__tests__/bootstrap.case.ts` (new) | W4.1 | Tests: missing memory path continues; sync-state absent treated as `unknown`; contract present skips approval; contract expired requests fresh; bootstrap report structure matches spec |

**Parallel:** W4.1-W4.4 [P] can be done together. W4.5 depends on all of them. W4.6 and W4.7 follow.
**Risk:** Medium. This is the highest-risk wave — it changes the orchestrator runtime. Bootstrap must be non-blocking. Existing chain-machine tests must still pass.

**Clarification needed before coding:**
- Where should maintenance contracts be persisted? Proposed: `.ai/housekeeping/contracts/<id>.json`. This matches the existing persistence pattern in `.ai/orchestration/state/`. Confirm or suggest alternate location.
- Should `runBootstrap` shell out to qmd/codegraph commands, or should we add adapter interfaces? Proposed: define `DriftChecker` interface with `qmd` and `codegraph` implementations; start with stub/no-op implementations and wire real detection later.

---

### Wave 5 — Orchestrator: Housekeeping Hooks at Advance Chain

**Goal:** Pre-task and post-task housekeeping run during `advance_chain`.

| Task | Layer | Files Touched | Depends On | Acceptance |
|------|-------|-------------|------------|------------|
| **W5.1** TS: add `runPreTaskHousekeeping` function | Orchestrator | `orchestrator/src/housekeeping.ts` (new) | W0.5 | Implements §1 of Spec 006 Task 003: check contract → read-only drift → eval context needs → request approval → load. Returns `HousekeepingReport`. |
| **W5.2** TS: add `runPostTaskHousekeeping` function | Orchestrator | `orchestrator/src/housekeeping.ts` | W0.5 | Implements §3 of Spec 006 Task 003: mandatory memory extraction sweep → cleanup proposal → sync proposal → approval gate → report. Returns `HousekeepingReport`. |
| **W5.3** TS: add `runInlineMemoryExtraction` function | Orchestrator | `orchestrator/src/housekeeping.ts` | W0.5 | Implements §2 of Spec 006 Task 003: detects lessons/decisions from step output; stages or writes memory entries per contract. |
| **W5.4** TS: wire `advance_chain` to call housekeeping | Orchestrator | `orchestrator/src/tool-handlers.ts`, `orchestrator/src/chain-machine.ts` | W5.1, W5.2, W5.3 | Before next step: pre-task housekeeping. After step completes: inline extraction + post-task housekeeping. Housekeeping failure is non-blocking. |
| **W5.5** TS: add housekeeping report to step state | Orchestrator | `orchestrator/src/types.ts` | W0.5 | `StepState` gains optional `housekeepingReport?: HousekeepingReport` |
| **W5.6** TS: housekeeping tests | Orchestrator | `orchestrator/src/__tests__/housekeeping.case.ts` (new) | W5.1-W5.3 | Tests: contract present skips approval; contract expired requests fresh; rejected sync recorded as `staleAcked`; inline extraction stages entries when no write contract; post-task sweep merges staged entries |

**Parallel:** W5.1-W5.3 [P] independent. W5.4 depends on all three. W5.5-W5.6 follow.
**Risk:** Medium. Modifying `advance_chain` is the second-highest-risk change. Non-blocking requirement is critical.

---

### Wave 6 — Wizard Phase 5: Optional Tooling

**Goal:** Add a new wizard phase (or sub-phase) for optional qmd/codegraph/obsidian/memory-path configuration.

| Task | Layer | Files Touched | Depends On | Acceptance |
|------|-------|-------------|------------|------------|
| **W6.1** Go: add `WizardResult.Phase5` type | Go CLI | `tui/wizard/wizard.go` | — | `Phase5Result` struct: `MemoryPath string`, `EnableObsidian bool`, `ObsidianVaultPath string`, `EnableQmd bool`, `QmdIndexPath string`, `EnableCodegraph bool`, `CodegraphDataPath string` |
| **W6.2** Go: implement `RunPhase5` wizard step | Go CLI | `tui/wizard/phase5.go` (new) | W6.1 | Interactive prompts: memory path (default `specs/memory`), Obsidian enable, qmd enable, codegraph enable. Each prompt explains read-only vs mutating approval model per Spec 006 Task 005. Non-interactive path uses defaults. |
| **W6.3** Go: wire Phase 5 into wizard flow | Go CLI | `tui/wizard/wizard.go` | W6.2 | After Phase 4 confirmation, run Phase 5 (or before Phase 4 — confirm sequence). Update `RunWizardWithDefaults` to include Phase 5. |
| **W6.4** Go: scaffold `.ai/housekeeping/` setup | Go CLI | `internal/scaffold/` (new file or extend `specs.go`) | W6.1 | If any housekeeping feature is enabled, scaffold `specs/memory/` dir and `.ai/housekeeping/` dir. Optionally write initial `sync-state.json` with disabled tools. |
| **W6.5** Go: persist Phase 5 config to store | Go CLI | `cmd/init.go`, `internal/types/types.go` | W6.1, W6.4 | `StoreData` gains optional `Housekeeping HousekeepingConfig` field; `writeStoreFromScaffoldResult` writes it. |
| **W6.6** Go: update `buildScaffoldContext` | Go CLI | `cmd/init.go` | W6.1, W6.4 | Pass Phase 5 results into scaffold context so directories/config are created. |
| **W6.7** Go: non-interactive init support for Phase 5 | Go CLI | `cmd/init.go` | W6.1 | New CLI flags: `--memory-path`, `--enable-qmd`, `--enable-codegraph`, `--enable-obsidian`. Wired to Phase 5 defaults in non-interactive mode. |
| **W6.8** Go: wizard Phase 5 tests | Go CLI | `tui/wizard/phase5_test.go` (new) | W6.2 | Non-interactive phase 5 with defaults; interactive flow skipped in CI but manually verifiable. |

**Parallel:** W6.1 first, then W6.2 [P] with W6.4 [P]. W6.3, W6.5-W6.7 follow.
**Risk:** Medium. Adding a wizard phase touches the TUI flow and requires careful back-navigation handling. The existing 4-phase wizard pattern is well-established.

**Clarification needed before coding:**
- Phase 5 placement: before or after Phase 4 (confirm)? **Proposed:** Before Phase 4, so the user sees optional tooling choices in the final summary before confirming. This matches the install-ux task 005 recommendation that memory path is confirmed before final config is written.
- Should `--enable-qmd` etc. also appear on `ai-setup add` and `ai-setup update`? **Proposed:** Yes for `update`, no for `add` (add is for agents/skills/templates, not infrastructure). Add to `update` command flags as a separate small task.

---

### Wave 7 — Compose Agent: Inject Library Context via stepInstructions

**Goal:** When `compose_agent` is called during a chain, the orchestrator injects the relevant `library/specs-agents/*.md` content into `stepInstructions` based on task type.

| Task | Layer | Files Touched | Depends On | Acceptance |
|------|-------|-------------|------------|------------|
| **W7.1** TS: add `resolveSpecAgentContent` function | Orchestrator | `orchestrator/src/loader.ts` | — | Given a task-type hint (e.g., `adr`, `feature`, `bugfix`), read the matching `library/specs-agents/<type>.md` from the library directory and return its content as a string. Falls back to empty string if not found. |
| **W7.2** TS: wire `composeAgent` handler to inject spec-agent content | Orchestrator | `orchestrator/src/tool-handlers.ts` | W7.1 | When `composeAgent` is called during a chain step, and the step has a task-type classification, prepend or append the relevant `library/specs-agents/*.md` content to `stepInstructions`. Uses existing `stepInstructions` field — no new API surface. |
| **W7.3** TS: add step-level `taskType` hint to `ChainStepDefinition` | Orchestrator | `orchestrator/src/types.ts` | — | Optional `taskType?: string` on `ChainStepDefinition`. Values: `feature`, `bugfix`, `refactor`, `adr`, etc. Used to resolve which spec-agent file to inject. |
| **W7.4** TS: composer integration tests | Orchestrator | `orchestrator/src/__tests__/composer.case.ts` (extend) | W7.1, W7.2 | Test that `stepInstructions` includes spec-agent content when `taskType` is set; test fallback when taskType is absent or library file missing |

**Parallel:** W7.1 [P] with W7.3. W7.2 depends on both. W7.4 follows.
**Risk:** Low. No API change; just adding content to an existing field. But need to verify that `stepInstructions` length doesn't cause context bloat.

---

### Wave 8 — Integration Verification

**Goal:** End-to-end verification that all waves work together correctly.

| Task | Layer | Files Touched | Depends On | Acceptance |
|------|-------|-------------|------------|------------|
| **W8.1** Go: `ai-setup init` produces housekeeping config | Manual/E2E | — | W6 | Run `ai-setup init --non-interactive --enable-qmd --memory-path specs/memory`; verify `.ai/housekeeping/` created; verify no `specs/**/AGENTS.md` generated |
| **W8.2** Go: `ai-setup doctor` reports stray AGENTS + metadata gaps | Manual/E2E | — | W1 | After W3 (files deleted), doctor reports 0 stray; before W3, reports 2 stray; adds metadata gap warnings for spec files |
| **W8.3** TS: `start_chain` → bootstrap report present | Manual/E2E | — | W4, W5 | Start a chain via MCP; verify chain state contains `bootstrapReport`; verify pre/post-task housekeeping reports present on steps |
| **W8.4** TS: compose agent injects spec-agent content | Manual/E2E | — | W7 | Compose an agent with `taskType: "adr"`; verify `stepInstructions` includes ADR guidance from `library/specs-agents/adrs.md` |
| **W8.5** Full regression | CI | — | All | `go test ./...` passes; `cd orchestrator && npx vitest run` passes; no new lint errors |

**Parallel:** W8.1-W8.4 [P]. W8.5 is final gate.
**Risk:** Low. Verification only, no new code.

---

## Dependency Graph (Simplified)

```
W0 (Foundation Types)
 ├──► W1 (Doctor Checks) ──────────────────────────────────────────────► W8
 ├──► W2 (Stop Sub-AGENTS.md Generation) ──► W3 (Remove Stray Files) ──► W8
 ├──► W4 (Bootstrap Hook) ──► W5 (Housekeeping Hooks) ───────────────► W8
 └──► W6 (Wizard Phase 5) ────────────────────────────────────────────► W8

W7 (Compose Agent Injection) ─── [P] independent ────────────────────► W8

W8 (Integration Verification) depends on all
```

Parallelism opportunities:
- W0 first (everyone depends on it)
- Then W1, W2, W4, W6, W7 can all proceed in parallel
- W3 depends on W2; W5 depends on W4
- W8 is the final verification wave

---

## Risk Register

| # | Risk | Severity | Wave | Mitigation |
|---|------|----------|------|------------|
| R-1 | Bootstrap/housekeeping hooks make `start_chain`/`advance_chain` slower or flaky | High | W4, W5 | All hooks are non-blocking; failures are logged and swallowed; timeout guards on external tool calls; extensive unit tests for error paths |
| R-2 | Removing sub-AGENTS.md generation breaks existing users on `update` | Medium | W2, W3 | `update` command migration step removes tracked records; `doctor` reports stray files so users can self-serve before updating; no silent deletion |
| R-3 | Wizard Phase 5 adds friction to `init` flow | Medium | W6 | Phase 5 is quick (3 boolean toggles + 1 path); all optional; non-interactive mode uses defaults without extra prompts; `init` summary shows enabled integrations clearly |
| R-4 | `stepInstructions` bloat from spec-agent content injection | Medium | W7 | Content is advisory and short (each `specs-agents/*.md` is <100 lines); injection is opt-in per step `taskType`; no injection when type is absent |
| R-5 | `.ai/housekeeping/` directory creation race with existing `.ai/orchestration/` | Low | W4, W6 | `persistence.ts` already has `ensureStateDir`; housekeeping dir creation follows same pattern; no conflict since paths differ |
| R-6 | Maintenance contract persistence format unclear | Low | W4 | Proposed JSON files in `.ai/housekeeping/contracts/`; matches existing pattern; if rejected, alternative is SQLite table in future spec |

---

## Unresolved Decisions (DECISION NEEDED)

| # | Item | Options | Impact if Postponed | Default Assumption for Coding |
|---|------|---------|---------------------|-------------------------------|
| UD-1 | Where to persist maintenance contracts | (a) `.ai/housekeeping/contracts/<id>.json` (b) SQLite table (c) In-memory only per session | (c) means contracts don't survive restart; (b) requires DB schema work outside scope | **(a)** — JSON files, matching existing persistence pattern |
| UD-2 | How does the orchestrator call external tools (qmd, codegraph) for drift checks? | (a) Shell out via `child_process.exec` (b) Define adapter interfaces with stub implementations (c) MCP tool calls | (c) requires MCP tools that don't exist yet; (a) is simplest but tightly coupled | **(b)** — `DriftChecker` interface, start with no-op/stub; real implementations in future PRs |
| UD-3 | Should `advance_chain` always run housekeeping, or only when a flag/config is set? | (a) Always (b) Only when housekeeping is enabled in project config | (b) means some projects never get housekeeping — contradicts spec intent | **(a)** — Always run; housekeeping degrades gracefully when tools are absent |
| UD-4 | Phase 5 wizard placement: before or after Phase 4 confirmation? | (a) Before Phase 4 (b) After Phase 4 as a post-confirm step | Minor UX difference | **(a)** — User sees all choices in summary before confirming |

---

## Acceptance Criteria Mapping

| AC | Wave(s) | How Verified |
|----|---------|-------------|
| AC-1 | W4 | `start_chain` produces `BootstrapReport` with all 5 sections; tested in W4.7 and W8.3 |
| AC-2 | W5 | Pre-task + post-task + inline extraction hooks in `advance_chain`; tested in W5.6 and W8.3 |
| AC-3 | W0, W4 | Memory file format types and compaction thresholds in data model; bootstrap/housekeeping emit line-count warnings |
| AC-4 | W0, W1 | Metadata schema in `Spec006Metadata` type; doctor reports gaps; validation function enforces required fields |
| AC-5 | W0, W4 | Contract types and permitted-action matrix in `MaintenanceContract` type; bootstrap and housekeeping evaluate contracts before mutation |
| AC-6 | W4, W5 | Drift detection via `DriftChecker` interface; proposal flow in `runPostTaskHousekeeping`; rejection recorded in `staleAcked` |
| AC-7 | W6 | Wizard Phase 5 prompts for `ob`/qmd/codegraph with approval model descriptions |
| AC-8 | W2, W3, W1 | Scaffold stops generating; stray files deleted; doctor reports remaining strays |
| AC-9 | All | Architecture boundaries enforced: wizard owns install UX; orchestrator owns runtime hooks; types/specs own schemas |
| AC-10 | — | All open decisions preserved above with `DECISION NEEDED` markers |

---

## Quickstart Verification Checklist (for Reviewer)

```bash
# 1. Foundation types compile
go build ./...

# 2. Doctor reports stray AGENTS (before W3 deletion)
./ai-setup doctor --json | jq '.strayAgentsFiles'

# 3. Init does not create sub-AGENTS.md
rm -rf /tmp/test-init && mkdir /tmp/test-init && cd /tmp/test-init
/path/to/ai-setup init --non-interactive --scope project --tools opencode --preset standard --name test
find specs/ -name "AGENTS.md"   # Should only find root AGENTS.md, no specs/**/AGENTS.md

# 4. Init with housekeeping
/path/to/ai-setup init --non-interactive --scope project --tools opencode --preset standard --name test \
  --memory-path specs/memory --enable-qmd
ls -la .ai/housekeeping/        # Should see sync-state.json

# 5. Orchestrator tests
cd orchestrator && npx vitest run

# 6. Doctor after W3: no stray
./ai-setup doctor --json | jq '.strayAgentsFiles'  # Should be []

# 7. Full regression
go test ./...
cd orchestrator && npx vitest run
```

---

## Task Summary

| Wave | Tasks | Estimated Effort | Layer |
|------|-------|-----------------|-------|
| W0 | 7 | Medium | Go + TS |
| W1 | 3 | Small-Medium | Go |
| W2 | 6 | Small | Go |
| W3 | 4 | Small (content + deletion) | Library + Repo |
| W4 | 7 | Medium-Large | TS |
| W5 | 6 | Medium | TS |
| W6 | 8 | Medium | Go |
| W7 | 4 | Small-Medium | TS |
| W8 | 5 | Small | E2E/Manual |

**Total:** 50 implementation tasks across 8 waves.
**Critical path:** W0 → W4 → W5 → W8 (orchestrator housekeeping is the main value delivery).
**Fastest path to visible value:** W0 → W1 → W2 → W3 (doctor + scaffold cleanup is self-contained and user-visible immediately).