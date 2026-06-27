# Tasks: LazyAI V2 (manifest-driven compile)

> **HARD GATE:** Implementation is blocked until the human approves `plan.md`. Do not spawn implementation subagents before approval.
>
> `[P]` = file-disjoint, runnable as an independent subagent. Each subagent runs ONLY its own focused test, never repo-wide build/lint; the lead runs full gates (`make build`, `cd packages/cli && go test ./... -count=1`, `make vet`, `make lint`, `mkdocs build --strict`) once over the union.

---

## Phase A — Manifest nucleus (V2 minimum)

- [x] **A.1** `internal/schema/`: embedded JSON schemas for `lazyai.json`/`lock.json`/`mcp-catalog.json` + accessors. Done in `internal/schema/`. [delegated]
- [x] **A.2** `internal/aimanifest/` (NOTE: package named `aimanifest`, not `manifest` — `internal/manifest` already exists as the legacy `.ai-setup.json` store manager): `Load`/`Save`/`Validate`/`ResolveTargets`/`EnabledTargets`/`Default`/`ForTools`. Rejects codex + unknown targets; `claude`↔`claude-code` alias.
- [x] **A.3** `internal/lockfile/`: `Load`/`Save`/`HashBytes`/`Find`/`Upsert`, sha256 + `generated[]`. [delegated]
- [x] **A.4** `internal/plan/`: pure `Build(desired, lock, DiskReader) Plan` — Create/Update/Skip/Drift; FR-004 idempotency; managed re-merge; untracked-whole-file drift.
- [x] **A.5** `internal/writer/`: managed-region `Apply` reusing `adapter.MergeManagedBlock` + atomic temp-rename; preserves out-of-region edits (EC-002); refuse-on-drift without `Force` (EC-003); records lock OutputHash of on-disk bytes.
- [x] **A.6** `internal/scaffold/ScaffoldManifest` wired into `ScaffoldAll` — `init` now writes `.ai/lazyai.json` (idempotent, from selected tools).
- [x] **A.7** `cmd/compile.go`: manifest is authoritative for target selection when present (overrides store), validated up front (rejects invalid/codex → EC-001-adjacent error), and writes `.ai/lock.json` (FR-003). SCOPE NOTE: this is the additive, non-breaking front-end. Full manifest-*required* compile + routing ALL asset families through `plan`/`writer` is deferred to Phase B once the adapter capabilities model (B.1) can enumerate per-adapter desired outputs. `plan`/`writer` are built + unit-proven, ready for B to feed.
- **Validation (A): ✅** `go build ./...`, `go vet ./...`, `go test ./...` (40 pkg ok, 0 fail), `make build` (binary) all green. `make lint` blocked by pre-existing staticcheck/go1.26 toolchain skew (not this change). Idempotency (FR-004) proven by `writer.TestSecondRunIsNoOp`.

## Phase B — Capabilities model & docs conformance

- [x] **B.1** `internal/adapter/capabilities.go`: `Capability` struct (16 surface bools + `SupportLevel`) + `Capabilities() Capability` on the `ToolAdapter` interface, implemented for all 7 targets, grounded in the 2026-06-21 compliance matrix. `IsBeta()` helper. Test: `capabilities_test.go` (every adapter reports MCP+root, support levels valid, Kiro specs/steering, Claude root).
- [x] **B.2** Codex removal: confirmed Codex was already absent as an adapter/target (no `codex.go`; not in `types.ToolId`, registry, or `IsValidToolId`). Cleared stray *target-context* references (workspace_compile + configmerge comments, `library/skills/populate` compat line, one stale test comment). KEPT (legit, not targets): `auth/probe.go` Codex-CLI→OpenAI provider probe, `models` `gpt-5.1-codex` model names, `library/hooks/pre-commit` AI-author detection grep, and the negative tests that assert codex is rejected (the V2 invariant).
- [x] **B.3** Conformance tests: FR-012 fixed a **real gap** — Claude target previously generated NO `CLAUDE.md` (only appended to a pre-existing one). `scaffold.ensureClaudeContextDoc` now generates a native `CLAUDE.md` importing `@AGENTS.md` (single-sourced), tracked + idempotent; `root_test`/`add_test` updated. FR-013 guards added: OpenCode config/agents must use `permission` not `tools`, never `maxSteps` (`opencode_adapter_test`). Capability conformance in `capabilities_test`. NOTE: `testdata/docs-snapshots/` official-doc capture is the **manual** review step that keeps OMP/Antigravity beta — deferred (cannot be fabricated without official doc text).
- [x] **B.4** OMP + Antigravity adapter capability model: compile output annotates adapter capability surfaces (EC-006), doctor reports adapter capabilities for configured tools; both OMP and Antigravity are stable (SupportStable). Tests: `TestCompileSurfacesBetaAdapters`, `TestBetaAdaptersAreOmpAndAntigravity`.
- **Validation (B): ✅** `go build ./...`, `go vet ./...`, `go test ./...` (40 pkg ok, 0 fail), `make build` all green. Codex fully removed from targets; FR-012/FR-013 conformance locked by tests. `make lint` still blocked by the pre-existing staticcheck/go1.26 toolchain skew (not this change).

## Phase C — Validation hardening & doctor security

- [x] **C.1** `internal/validate` (new package): `All(Options{Root,Profile}) Report` runs skill/agent/hook/MCP validators over the canonical `.ai/` tree, wired behind `lazyai-cli validate --all` (FR-009). Skills require frontmatter + valid `name` (via `internal/validation.ValidateArtifactName`) + non-empty `description`. MCP entries must declare `command`|`url`. Hooks: dangerous-pattern + shebang checks. Tests: bad skill name fails, missing description fails, MCP-missing-command fails, plus `validate --all` CLI integration (fail/clean).
- [x] **C.2** `internal/validate/secrets.go`: named-pattern scanner (AWS/private-key/OpenAI/GitHub/Slack/Google) over markdown/script/config assets + MCP `env` literals. `${VAR}`/`$VAR` references are always safe. Profile-aware severity: **error under `team`, warning under `personal`** (FR-010). Tests: inline secret fails team / warns personal, env-ref passes.
- [x] **C.3** `internal/validate/paths.go`: walks `.ai/`, lstat-resolves every symlink; targets escaping the repo root are **errors**, internal symlinks **warn**. Tests: escaping symlink rejected, internal symlink warns (both `t.Skip` when the OS lacks symlink support).
- [x] **C.4** `cmd/doctor_security.go`: `buildSecurityReport(root, tools, profile)` → MCP inventory (transport + env ref/inline/none), hook/secret/path risks (reuses `validate.All`), and trust/sandbox caveats for Pi+Kiro (`noSandboxTools`, FR-011). Rendered by `printSecurityReport` in `runDoctor` (advisory; never changes doctor health). Tests: pi/kiro caveats present, hook+secret risks flagged, clean repo = no findings.
- **Validation (C): ✅** `gofmt -l` clean, `go vet ./...`, `go test ./...` (41 pkg ok — new `internal/validate` — 0 fail), `make build` all green. Negative-case suite (`internal/validate/validate_test.go`) + doctor caveats covered. `make lint` still blocked by the pre-existing staticcheck/go1.26 toolchain skew. NOTE: profile resolves from `--profile` flag > manifest `profile` field > `personal` default.

## Phase D — Migration & eject

- [x] **D.1** `internal/migration` (existing engine, extended in lieu of a parallel `internal/migrate/` package): detector now covers `AGENTS.md`/`CLAUDE.md`/`.opencode`/`opencode.json`/`.claude`/`.github`/`.pi`/`.omp`/`.kiro`/`.agents`; import preserves every detected native file under `.ai/adapters/<target>/raw/` and never deletes originals (FR-015). Canonical extraction remains **OpenCode-only** for now; other targets are preserved as raw adapter assets and marked low/medium confidence in the report per TechSpec §16.2/§16.3. Tests: detector coverage for pi/omp/kiro/antigravity/copilot sources; import fixture repo produces `.ai/` plus raw-preserved copies.
- [x] **D.2** `cmd/import.go`: rewired to the canonical path (`ParseDetectedSetups` → `BuildCanonicalPlan` → `ExecuteToCanonical`), bootstraps `.ai/lazyai.json` from detected targets, preserves raw native inputs under `.ai/adapters/<target>/raw/`, and writes `.ai/migration-report.md` with confidence-scored findings / zero-delete guarantee. Tests: report emitted, manifest emitted, native files retained.
- [x] **D.3** `internal/eject/` + `cmd/eject.go`: `eject.Inspect`/`eject.Run` strip LazyAI management metadata (`.ai/lazyai.json`, `.ai/lock.json`, `.ai/migration-report.md`, legacy `.ai-setup.json`, `.ai-setup.db`) while leaving native files untouched (FR-016). Tests: metadata removed, native files still present after eject.
- **Validation (D): ✅** `gofmt -l` clean, `go vet ./...`, `go test ./...` (**42 pkg ok**, +`internal/eject`, 0 fail), `make build` all green. Round-trip command test (`runImport` → `runEject`) proves **no native loss**; detector test covers the widened source set. `make lint` remains blocked by the pre-existing staticcheck/go1.26 toolchain skew.

## Phase E — Bundles & eval (optional, gated)

- [x] **E.1** `cmd/build_plugin.go` + `internal/plugin/bundles.go`: `build-plugin --target {claude,copilot-cli,omp,pi}` now supports four bundle targets (FR-017). Claude keeps the existing `.claude-plugin` layout. Copilot-CLI bundle emits `plugin.json`, transformed `agents/*.agent.md`, `skills/<name>/SKILL.md`, merged `hooks.json` + hook scripts, and `.mcp.json`. OMP bundle emits `skills/`, `commands/`, `hooks/pre/`, and `mcp.json` (compiled when available, otherwise bundled from `library/mcp/catalog.json`). Pi package bundle emits `agents/`, `skills/`, `prompts/`, and `extensions/` from the existing adapter shapes. Tests: `internal/plugin/bundles_test.go` covers target normalization + bundle shape for copilot-cli/omp/pi; `cmd/build_plugin_test.go` covers command wiring + invalid target rejection.
- [x] **E.2** `internal/schema` + `internal/evals` + `validate evals` (FR-018): embedded schema artifacts added for `eval-case`, `eval-holdout`, and `eval-rubric`; `validate evals` validates `.ai/evals/cases/*.yaml`, `.ai/evals/holdouts/*.yaml`, and `.ai/evals/rubrics/*.md` locally with no cloud/runtime dependency. Conservative rules: YAML files must parse and contain `id`, `title`, `input`, `expected`; optional `holdout` must be boolean; rubric markdown must be non-empty. Tests: valid eval case passes, invalid eval case fails, missing `.ai/evals` errors clearly; schema accessors/names updated and tested.
- **Validation (E): ✅** `gofmt -l` clean, `go vet ./...`, `go test ./...` (**42 pkg ok, 11 no tests**, 0 fail), `make build` all green. Bundle-shape tests pass for copilot-cli/omp/pi; command wiring test proves `build-plugin --target copilot-cli`; `validate evals` passes a valid case and fails invalid input. `make lint` remains blocked by the pre-existing staticcheck/go1.26 toolchain skew.

## Phase F — Cleanup (LAST; gated on A–D working)

- [x] **F.1** Spec-pack binary-name normalization complete (FR-019): pack command examples now use `lazyai-cli`; root `README.md` updated for the manifest/lockfile V2 contract, seven-target surface, and multi-target `build-plugin`; `CHANGELOG.md` records the V2 manifest-driven compile, validation hardening, migration/eject, bundles/evals, and Codex-drop cleanup.
- [x] **F.2** Completion registered in docs/agent notes: `specs/KNOWLEDGE_MAP.md` marks spec 029 complete and links a new architecture decision; `packages/cli/KNOWLEDGE_MAP.md` replaced the placeholder template with the setup-core module map and V2 terminology; `.github/copilot-instructions.md` now records the canonical `.ai/` contract for future agents; ADR-006 documents manifest-driven compile + the seven-target/Codex-removal contract.
- [x] **F.3** `.ai/` v1 schema freeze complete: `internal/aimanifest` now enforces manifest schema version `1.0`; `internal/lockfile` + `internal/writer` always emit lockfile schema version `1.0`; scaffolded canonical `.ai/mcp.json` now carries MCP catalog version `1.0`; embedded JSON schemas/examples updated to `1.0`; compile lock metadata now records the actual `cmd.Version` instead of a stale hardcoded `0.1.0`. Focused tests cover manifest-version rejection, lock defaults, scaffolded MCP versioning, and compile lock emission.
- **Validation (F): ✅** `make test` (**42 pkg ok, 11 no tests**), `make vet`, `make lint`, and `mkdocs build --strict` all green. MkDocs strict-mode cleanup explicitly sets `validation.nav.omitted_files: info` for intentionally linkable-but-unlisted docs and excludes root `docs/README.md` to avoid the `index.md` collision warning.

---

## Dependency Order

1. A.1 → A.2 → A.3 → A.4 → A.5 → A.6 → A.7
2. A.7 → B.1 → {B.2, B.3, B.4}
3. B.1 → C.1 → {C.2, C.3} → C.4
4. C.* → D.1 → D.2 → D.3
5. (optional) D.* → {E.1, E.2}
6. A–D green → F.1 → F.2 → F.3

## Parallel lanes (after their gate task lands)
- Lane 1 [P]: A.1, A.3 (independent foundation files)
- Lane 2 [P]: B.3, B.4 (after B.1)
- Lane 3 [P]: C.2, C.3 (after C.1)
- Lane 4 [P]: D.3 (after D.1); E.1, E.2 (after D)
- Sequential single-owner: A.7, B.2, C.4, D.1/D.2, and all of Phase F (shared files / cross-cutting).
