# Spec 007 — Wizard Step-by-Step UX

> Scope: Split the Go `ai-setup init` wizard's bundled `huh.NewGroup` forms into sequential per-question screens so it feels step-by-step again (matches the prior TypeScript wizard's feel). UX-only. No new data fields, no workspace/planning-repo parity, no scaffold or adapter changes.

---

## Context

The live Go wizard (`tui/wizard/`) groups many unrelated questions into single `huh.NewGroup(...).Run()` calls:

- **Phase 1** (`tui/wizard/phase1.go:139`) bundles scope + AI tools + project name + CLI tools + MCP servers in one group.
- **Phase 5** (`tui/wizard/phase5.go:57-65`) bundles memory path + three `enable-X` toggles + three `X-path` inputs in one group (7 fields, no conditional skips).
- **Phase 2** (`tui/wizard/phase2.go:169-172`) bundles branch pattern + commit pattern + require-ticket in one git-conventions group.

The legacy TypeScript wizard (`src/wizard/phase1-context.ts`) asked these one prompt at a time, yielding an inline, incremental flow. The port lost that feel.

Re-reading `src/wizard/phase1-context.ts` also surfaces cheap parity wins the Go port dropped: `(previous: X)` pre-fill markers on re-run, pre-selection of already-installed CLI tools, filesystem-safe project-name validation, and a skip of the name prompt when scope is global.

This plan restores step-by-step sequencing and the cheap parity wins. It does **not** restore the workspace / planning-repo / repos flow — that is explicitly Wave 2.

---

## Goals

1. Each wizard phase presents one focused question per screen (conditional follow-ups only when the prior answer requires them).
2. Re-runs pre-fill prior choices and signal `(previous: X)` so users can press Enter to accept.
3. Non-interactive (`--scope … --tools … …`) behavior is byte-for-byte unchanged.
4. Each phase exposes a pure `buildXResult(...)` builder unit-testable without TUI.
5. Phase titling is internally consistent (no "Phase 4/4" / "Phase 5/5" drift against wizard.go's own comments).
6. Every sub-screen displays a `<step>/<total>` counter accurate for the user's current branch.
7. Selecting `↩ Back` on any Select prompt returns to the prior sub-screen within the same phase; the outer `PhaseBack` contract (re-entering prior phases) is preserved.

---

## Acceptance Criteria

| #   | AC                                                                                                                                                                                          | Verified By                                    |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------- |
| AC-1  | Phase 1 presents: scope → AI tools → (conditionally) project name → CLI tools → MCP servers as separate sequential screens                                                                 | Manual smoke + unit test of sequencing         |
| AC-2  | Project-name screen is skipped when `scope == global`; name is set to `"global"` implicitly                                                                                                 | Unit test                                      |
| AC-3  | Project-name validator rejects path separators, empty, leading-dot, and whitespace-only inputs (not just empty)                                                                            | Unit test for `validateProjectName`            |
| AC-4  | Already-installed CLI tools are pre-selected by default in the CLI-tools multi-select                                                                                                       | Manual smoke + unit test of `NewCliToolsSelect`|
| AC-5  | Scope / tools / project-name prompts show `(previous: X)` marker in the title when defaults are pre-filled (re-run case)                                                                    | Manual smoke                                   |
| AC-6  | Phase 2 presents: preset → (only on custom) features → branch pattern → commit pattern → require-ticket as separate sequential screens; custom-pattern sub-prompt remains contextual        | Manual smoke                                   |
| AC-7  | Phase 5 presents: memory path → enable Obsidian → (if enabled) Obsidian vault path → enable qmd → (if enabled) qmd index path → enable codegraph → (if enabled) codegraph data path         | Manual smoke + unit test of `buildPhase5Result`|
| AC-8  | Wizard screen titles no longer display `Phase X/4` / `Phase X/5`; replaced with neutral section labels (`Setup Context`, `Features & Conventions`, `Optional Tooling`, `Review & Confirm`) | Manual smoke                                   |
| AC-9  | `wizard.go` doc comments no longer contradict the real phase count (today: `:1-2` says 5-phase, `:84` says 4-phase)                                                                         | Code review                                    |
| AC-10 | Each phase exposes a pure `buildPhaseXResult(...) *PhaseXResult` function callable without a TTY                                                                                            | Unit tests for each builder                    |
| AC-11 | Non-interactive behavior unchanged: same flags, same defaults, same result structs; existing `TestRunPhase5NonInteractiveDefaults` still passes; equivalents added for Phases 1 and 2       | `go test ./tui/wizard/...`                     |
| AC-12 | Ctrl-C / Esc at any sub-screen returns `PhaseCancel` cleanly; partial state is not persisted                                                                                                | Manual smoke                                   |
| AC-13 | Each Select / MultiSelect prompt (except the first of each phase) offers an `↩ Back` option that returns `PhaseBack`; the caller resumes at the prior sub-screen within the same phase. Inputs / Confirms remain cancel-only via Esc (matches prior TS behavior). | Manual smoke + unit test of in-phase back routing |
| AC-14 | Every sub-screen title uses the format `<PhaseTitle> — <n>/<N>: <StepTitle>`. `N` is the count of screens that will render for the user's current branch (e.g., Phase 5 with Obsidian off / qmd on / codegraph off shows `N = 3`); `n` updates as branching answers change. | Manual smoke |

---

## Affected Files

### Modified

| File                                | Change                                                                                                                                                                                                                       |
| ----------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `tui/wizard/phase1.go`              | Split single `huh.NewGroup` into sequential `askScope` / `askTools` / `askProjectName` / `askCliTools` / `askMcpServers`; skip name on `global` scope; add safe-name validator; add `(previous: X)` markers; extract `buildPhase1Result` |
| `tui/wizard/phase1_cli_mcp.go`      | Pre-select installed CLI tools as `MultiSelect` initial value; decouple catalog read from presentation so the builder stays pure                                                                                              |
| `tui/wizard/phase2.go`              | Split git-conventions group into sequential `askBranchPattern` / `askCommitPattern` / `askRequireTicket`; keep existing custom-pattern sub-prompt; extract `buildPhase2Result`                                               |
| `tui/wizard/phase4.go`              | Replace `Phase 4/4: Review & Confirm` title with `Review & Confirm` (summary render unchanged)                                                                                                                              |
| `tui/wizard/phase5.go`              | Replace single `huh.NewGroup` with sequential per-toggle / per-path screens with conditional skips; extract `buildPhase5Result`                                                                                              |
| `tui/wizard/wizard.go`              | Remove `4-phase` / `5-phase` doc-comment drift; update phase-flow comment to reflect actual order (1 → 2 → 3 → 5 → 4)                                                                                                        |

### New

| File                                                 | Purpose                                                                                                    |
| ---------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| `tui/wizard/phase1_test.go`                          | Unit tests for `buildPhase1Result`, non-interactive defaults, `validateProjectName`, global-scope skip      |
| `tui/wizard/phase2_test.go`                          | Unit tests for `buildPhase2Result`, preset resolution, git-convention defaults                             |
| `specs/007-wizard-step-by-step-ux/plan.md`           | This plan                                                                                                  |

### Explicitly Not Touched (see Out of Scope)

- `internal/scaffold/**`, `internal/adapter/**`, `cmd/helpers.go`, `cmd/init.go`, `internal/globalpaths/**`, any `library/**` asset.
- `Phase1Result`, `Phase2Result`, `Phase5Result`, `WizardConfig`, `WizardResult` struct shapes.
- Non-interactive CLI-flag surface.

---

## Approach

### Option A — Sequential screens via Huh (chosen)

- Keep `huh` as the TUI library.
- Replace each bundled group with a sequence of short single-question `huh.NewForm(huh.NewGroup(...)).Run()` calls.
- Conditional fields (Obsidian vault path, qmd index path, codegraph data path, Phase 1 project name on global scope, Phase 2 custom-pattern prompts) become skipped early-returns in the sequential flow.
- Separate "ask" (TUI) from "build" (pure mapping) functions so the builders are unit-testable.

### Option B — Single form with group-level `WithHideFunc()` guards (fallback)

- Keep a single `huh.NewForm(g1, g2, g3, ...)` but hide groups based on prior-group answers via `huh.Group.WithHideFunc()`.
- Preserves one alt-screen session, avoids any flicker between screens.

### Option C — Custom Bubble Tea state machine (rejected)

- Full rewrite of the wizard shell. Highest UX ceiling but effort is disproportionate to the goal.

### Decision

**Option A** is selected for Wave 1 (locked 2026-04-17). Option B is retained only as a fallback mitigation for R-1 should Wave 1a smoke reveal visible alt-screen flicker. Option C is rejected.

**Tradeoff accepted**: more small helper functions and more separate alt-screen renders than the single-form variant. Not recording as ADR — this is UX plumbing, not an architecture decision.

---

## Scope

### In

- Phase 1, 2, 5 screen splitting.
- Phase 4 title neutralization; wizard.go comment drift fix.
- Cheap TS-parity wins listed in AC-2, AC-3, AC-4, AC-5.
- Extraction of pure `buildXResult` builders and their tests (AC-10, AC-11).

### Out (Wave 2 — separate spec)

- TypeScript-parity restoration for workspace scope: `workspaceName` prompt, `planningRepoPath` picker, `repos` scan / multiselect.
- `--planning-repo` / `--repos` CLI flags.
- Re-run "existing manifest" note (`p.note('Re-running setup — previous selections will be pre-filled as defaults')`) — requires reading manifest state into the wizard, out of current scope.
- "Global AI setup detected at `~/.ai/`" note on project/workspace scope — requires filesystem probe, out of current scope.

---

## Resolved Decisions

Locked 2026-04-17 at plan review:

1. **In-phase back nav** — include in Wave 1. Each Select / MultiSelect sub-screen (except the first of its phase) appends an `↩ Back` option that returns `PhaseBack`; inputs / confirms remain cancel-only on Esc. Captured as AC-13.
2. **Title strategy** — neutral titles only. Function names `RunPhase1` / `RunPhase2` / `RunPhase4` / `RunPhase5` remain as-is.
3. **Progress counter** — yes. Format: `<PhaseTitle> — <n>/<N>: <StepTitle>`. `N` is the live count of screens that will render for the user's current branch; `n` updates as branching answers change. Captured as AC-14.
4. **Rendering** — Option A (sequential `huh.NewForm(group).Run()` calls). Option B retained only as a contingency mitigation for R-1.

---

## Implementation Waves

### Wave 1a — Phase 1 split (AC-1, AC-2, AC-3, AC-4, AC-5, AC-12, AC-13, AC-14)

**Task 1a-1**: Extract `buildPhase1Result(scope, tools, name, cliTools, servers) *Phase1Result` pure builder in `tui/wizard/phase1.go`. Move the codex-removal side-effect out of the builder and into a separate `filterUninstalledCodex(...)` helper.

**Task 1a-2**: Replace the single `huh.NewForm(group).Run()` with sequential sub-step functions, each returning `(value, PhaseAction, error)`:

1. `askScope(defaults, stepInfo) (types.SetupScope, PhaseAction, error)` — first of phase; no `↩ Back`.
2. `askTools(defaults, stepInfo) ([]types.ToolId, PhaseAction, error)` — followed by codex-install-hint confirm if codex is selected and not installed.
3. `askProjectName(defaults, scope, stepInfo) (string, PhaseAction, error)` — early-return `"global", PhaseContinue, nil` when `scope == types.SetupScopeGlobal`; in that branch the total (`N`) drops by 1.
4. `askCliTools(defaults, stepInfo) ([]string, PhaseAction, error)` — reuses `NewCliToolsSelect` with installed tools pre-selected (Task 1a-7).
5. `askMcpServers(defaults, stepInfo) ([]string, PhaseAction, error)` — reuses `NewMcpServersSelect`.

Every `askX` that uses a Select / MultiSelect (and is not the first of its phase) appends an `↩ Back` option; selecting it returns `PhaseBack`. Every title is formatted as `"Setup Context — <n>/<N>: <StepTitle>"` via a shared `stepInfo` struct carrying the live counter.

**Task 1a-3**: Add `validateProjectName(name string) error` in `tui/wizard/phase1.go`. Reject: empty string, whitespace-only, names containing `/`, `\`, `..`, leading `.`, trailing whitespace. If a Go port of `src/utils/validation.ts#validateFilesystemSafeName` exists in `internal/`, reuse it; otherwise add this validator locally and note the duplication for a future consolidation task.

**Task 1a-4**: Add `(previous: X)` marker to each `askX` title when `defaults` is non-nil and the relevant field is populated. Apply to scope, tools, project name. Pattern: `fmt.Sprintf("Setup Context — %d/%d: Scope (previous: %s)", n, total, priorLabel)`.

**Task 1a-5**: Rewrite `runPhase1Interactive` as a sub-step dispatcher. Maintain `currentStep := 1`; dispatch to the `askX` for that step; on `PhaseBack` decrement `currentStep` (clamped at 1); on `PhaseContinue` increment; on first-step `PhaseBack`, bubble up to the outer `runPhase12Loop` (no-op today, per `wizard.go:208-211`).

**Task 1a-6**: Add `tui/wizard/phase1_test.go` covering: non-interactive defaults; `buildPhase1Result` with each scope; `validateProjectName` valid/invalid cases; global-scope implicit name; in-phase back routing (step dispatcher returning to prior step on `PhaseBack`).

**Task 1a-7**: In `tui/wizard/phase1_cli_mcp.go`, extend `NewCliToolsSelect` to accept a `preSelected []string` derived from installed CLI tools (resolve each catalog entry via `exec.LookPath`; errors treated as "not installed"). Keep the lookup non-blocking; on panic / unexpected error fall back to an empty pre-selection.

---

### Wave 1b — Phase 2 split (AC-6, AC-13, AC-14)

**Task 1b-1**: Extract `buildPhase2Result(scope, preset, features, branch, commit, requireTicket) *Phase2Result` pure builder in `tui/wizard/phase2.go`.

**Task 1b-2**: Replace the git-conventions group (`phase2.go:169-172`) with sequential `ask*` sub-steps, each returning `(value, PhaseAction, error)`:

1. `askBranchPattern(defaults, stepInfo)` — if `"custom"`, recurse into the existing custom-branch input (same step; counter does not advance).
2. `askCommitPattern(defaults, stepInfo)` — same pattern for custom.
3. `askRequireTicket(defaults, stepInfo)`.

Preset and feature-multiselect screens already render as separate forms; re-wrap them under the same dispatcher so the counter reflects the full phase (e.g., `Features & Conventions — 3/4: Branch Pattern`; on `custom` preset, `N=5`). Append `↩ Back` to every Select except the preset screen (first of phase). Add `(previous: X)` markers where priors exist.

**Task 1b-3**: Rewrite `runPhase2Interactive` as a sub-step dispatcher mirroring Task 1a-5.

**Task 1b-4**: Add `tui/wizard/phase2_test.go` covering: non-interactive defaults (each scope → each preset); `buildPhase2Result` with custom preset + custom branch/commit patterns; git-convention default fallback; in-phase back routing.

---

### Wave 1c — Phase 5 split (AC-7, AC-13, AC-14)

**Task 1c-1**: Extract `buildPhase5Result(memoryPath, enableObsidian, obsidianVaultPath, enableQmd, qmdIndexPath, enableCodegraph, codegraphDataPath) *Phase5Result` pure builder in `tui/wizard/phase5.go`. Move the default-fallback block (`phase5.go:71-79`) into the builder.

**Task 1c-2**: Replace the single `huh.NewGroup` at `phase5.go:57-65` with sequential sub-steps:

1. `askMemoryPath(default, stepInfo)` → input.
2. `askEnableObsidian(default, stepInfo)` → confirm. If false, skip (3); `N` decrements by 1.
3. `askObsidianVaultPath(default, stepInfo)` → input.
4. `askEnableQmd(default, stepInfo)` → confirm. If false, skip (5); `N` decrements by 1.
5. `askQmdIndexPath(default, stepInfo)` → input (placeholder `.qmd-index`).
6. `askEnableCodegraph(default, stepInfo)` → confirm. If false, skip (7); `N` decrements by 1.
7. `askCodegraphDataPath(default, stepInfo)` → input (placeholder `.codegraph`).

Titles use `"Optional Tooling — <n>/<N>: <StepTitle>"`. `N` is recomputed after each confirm (max 7 when all toggles are on; min 4 when all are off). Phase 5 has **no** `↩ Back` options because all prompts are Input or Confirm (per Resolved Decision 1). If a future refactor switches any of these to Select, append `↩ Back` then.

**Task 1c-3**: Rewrite `runPhase5Interactive` as a sub-step dispatcher. Because Phase 5 has no Select-based back nav, the dispatcher only handles `PhaseContinue` and `PhaseCancel` (Esc). Retain the dispatcher pattern for consistency with 1a / 1b.

**Task 1c-4**: Extend `tui/wizard/phase5_test.go` with `buildPhase5Result` cases covering each toggle combination and the default fallback when a path is empty.

---

### Wave 1d — Title & comment cleanup (AC-8, AC-9, AC-14)

**Task 1d-1**: Most title rewriting is already folded into Waves 1a / 1b / 1c via the `<PhaseTitle> — <n>/<N>: <StepTitle>` format. Remaining:
- **Phase 3** (if present, `phase3.go`): align any hard-coded `Phase 3/N` references with the neutral `Conflict Resolution` title. Counter not required — Phase 3 is a single prompt.
- **Phase 4** (`phase4.go:101`): change title to `Review & Confirm`. Single screen, no counter.

**Task 1d-2**: In `tui/wizard/wizard.go`, reconcile the doc-comment drift:
- `:1-2` currently says `5-phase`; `:84` says `4-phase`. Update both to reflect the real flow: `// RunWizardWithDefaults executes the wizard: Phase 1 (context) → Phase 2 (features) → Phase 3 (conflicts, conditional) → Phase 5 (optional tooling) → Phase 4 (review & confirm).`
- Add a short comment documenting the `<PhaseTitle> — <n>/<N>: <StepTitle>` convention so future edits preserve it.

---

### Wave 1e — Testability wrap-up (AC-10, AC-11)

**Task 1e-1**: Verify each phase's `buildXResult` is pure (no filesystem, no exec, no `huh` reference). Re-run `go test ./tui/wizard/... -count=1`.

**Task 1e-2**: Run full `go vet ./...` and `go test ./... -count=1` before each Wave 1[a-d] merge to confirm no regressions in `internal/scaffold`, `internal/adapter`, `cmd/`.

---

## Verification Strategy

### Automated

| Layer       | What                                                                   | How                                                                                 |
| ----------- | ---------------------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| Unit        | `buildPhase1Result`, `buildPhase2Result`, `buildPhase5Result`          | Table-driven tests covering default, pre-fill, and each branch                      |
| Unit        | `validateProjectName`                                                  | Valid / invalid name cases                                                          |
| Unit        | Non-interactive paths for Phases 1 and 2                                | New `phase1_test.go` and `phase2_test.go` mirroring `TestRunPhase5NonInteractiveDefaults` |
| Regression  | Full `go test ./... -count=1`                                          | Run after each wave; must remain green                                              |
| Regression  | `go vet ./...`                                                         | Same                                                                                |

### Manual smoke (`ai-setup init`)

1. Interactive run, `scope = project` → confirm Phase 1 advances one screen at a time (scope → tools → name → CLI tools → MCP servers).
2. Interactive run, `scope = global` → confirm project-name screen is **skipped**; resulting name is `global`.
3. Interactive run, enter project name `bad/name` → confirm validator rejects.
4. Interactive run, on a host where `gh` is installed and `rg` is not → confirm `gh` is pre-selected and `rg` is not in the CLI-tools multi-select.
5. Re-run against a directory with a persisted `Phase1Result` default → confirm `(previous: X)` markers render on the scope / tools / name prompts.
6. Phase 2 → pick preset `custom` → confirm features screen appears; pick preset `standard` → confirm features screen is skipped; branch/commit pattern screens each render separately.
7. Phase 5 → leave Obsidian, qmd, codegraph toggles `false` → confirm the three corresponding path screens are skipped.
8. Ctrl-C at each sub-screen → wizard returns non-zero cleanly; no partial config written.
9. In Phase 1 `askTools`, select `↩ Back` → wizard re-renders `askScope` with the current selection pre-filled. Repeat at `askCliTools` → re-renders `askProjectName`.
10. In Phase 2 `askBranchPattern`, select `↩ Back` → wizard re-renders the feature multi-select (custom preset) or the preset select. Counter (`<n>/<N>`) updates accordingly.
11. In Phase 5, answer `No` to `Enable qmd?` → verify the qmd index-path screen is skipped and counter recomputes (`N` decrements by 1).

### Commands

```bash
go vet ./...
go test ./tui/wizard/... -count=1 -v
go test ./... -count=1
go build ./...
```

---

## Risks

| #   | Risk                                                                                         | Severity | Mitigation                                                                                                                                            |
| --- | -------------------------------------------------------------------------------------------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| R-1 | Multiple sequential `form.Run()` calls cause alt-screen flicker                              | Medium   | Smoke-test after Wave 1a. If flicker is visible, fall back to Option B (single form with `WithHideFunc()` group guards). Contingency only — not expected to trigger.            |
| R-2 | Splitting breaks back-nav (`PhaseBack`) plumbing                                             | Low      | Each sub-step returns `PhaseBack` / `PhaseContinue` explicitly. In-phase back nav is in scope (AC-13); the outer `runPhase12Loop` behavior remains unchanged.                   |
| R-3 | Safe-name validator rejects names the prior wizard accepted                                   | Medium   | First port TS `validateFilesystemSafeName` behavior exactly; extend only with additions explicitly agreed by user. Test with prior wizard's accepted set. |
| R-4 | Pre-select-installed-CLI-tools introduces shell-exec in Phase 1                              | Low      | Reuse `internal/detect` if it covers `which`/`exec.LookPath`; keep non-blocking — on LookPath error fall back to "no pre-selection".                 |
| R-5 | Neutral titles lose progress orientation without a counter                                    | Low      | Counter format `<PhaseTitle> — <n>/<N>: <StepTitle>` mandated by AC-14 mitigates this directly.                                                       |
| R-6 | Non-interactive CI path regresses                                                            | High     | AC-11 gate: run full `go test ./...` before each wave merge.                                                                                          |
| R-7 | Scope creep into workspace parity                                                            | High     | Explicit "Out of Scope" list; any workspace / planning-repo / repos prompt introduced by this spec violates the plan — reject in review.             |
| R-8 | `huh` v2 form compositions reset local `Value(&x)` variables between runs                     | Low      | Each `askX` owns its local value variable; no shared mutable state between screens. Covered by unit tests on `buildXResult`.                          |
| R-9 | Pure builder extraction leaves the TUI paths uncovered                                        | Low      | Accept: TUI paths remain manually smoke-tested. Existing test coverage gap is pre-existing and out of scope to close fully here.                       |

---

## Dependency Graph

```
Wave 1a (Phase 1 split) ──┐
Wave 1b (Phase 2 split) ──┼── Wave 1d (titles/comments) ──→ Wave 1e (testability wrap-up)
Wave 1c (Phase 5 split) ──┘
```

- 1a, 1b, 1c are independent and can be implemented as separate PRs in any order.
- 1d is cosmetic and can land at any point; easiest to batch with the last of 1a/1b/1c.
- 1e is the final sweep: verify builders stayed pure and run full quality gates.

---

## Out of Scope (explicit)

- Workspace setup flow: `workspaceName`, `planningRepoPath`, `repos` scan / multiselect — tracked as Wave 2 in a separate spec.
- `--planning-repo`, `--repos` CLI flags for non-interactive mode.
- Rewrite wizard as custom Bubble Tea state machine (Option C above).
- Changes to `WizardResult`, `Phase1Result`, `Phase2Result`, `Phase5Result` struct shapes.
- Changes to non-interactive CLI flag surface or semantics.
- Edits to `internal/scaffold/**`, `internal/adapter/**`, `cmd/helpers.go`, `internal/globalpaths/**`.
- Edits to `library/**` assets or catalog schema.
- Adding "existing manifest detected" / "global `.ai/` detected" notes (TS parity but requires new filesystem probes — defer to Wave 2).

---

## Notes

- Research pass that produced this plan cross-read `tui/wizard/{wizard,phase1,phase2,phase4,phase5,phase1_cli_mcp}.go`, `tui/wizard/phase5_test.go`, and `src/wizard/phase1-context.ts`.
- Codex install-hint side-effect (`phase1.go:157-176`) already renders as a secondary screen; splitting slots it naturally between steps 2 and 3 of Wave 1a.
- Current wizard runs **Phase 5 before Phase 4** intentionally (Phase 4 is the confirm/summary). This plan preserves that order and only fixes the labels.
