# Spec 005 — Setup Flow Fixes

> Scope: Seven targeted fixes to the `ai-setup init` wizard pipeline and adapter layer. No net-new features — correctness, coverage, and ergonomics only.

---

## Acceptance Criteria

| # | AC | Verified By |
|---|----|-------------|
| AC-1 | `opencode.json` is written to `.opencode/opencode.json` (project scope), not to project root; `opencode.jsonc` MCP-compiled output moves to `.opencode/opencode.jsonc` | Unit test + manual `ls .opencode/` |
| AC-2 | Phase 1 wizard includes a CLI Tools multi-select step drawn from `library/mcp/catalog.json` → `cliTools` key; selected tools are persisted to `ScaffoldContext.CLITools` and surfaced in the Phase 4 summary | Integration test + manual run |
| AC-3 | Phase 1 wizard includes an MCP Servers multi-select step drawn from `library/mcp/catalog.json` → `servers` key; selected servers are persisted to `ScaffoldContext.EnableServers` and flow through `ScaffoldMcp` | Integration test + manual run |
| AC-4 | After Phase 2, a detection pass runs `package.json`, `go.mod`, `Cargo.toml`, `pom.xml` probes on `TargetDir` to auto-fill `PrimaryLanguage`, `Framework`, `PackageManager`, `WorkspaceType` in `FragmentContext` / `ScaffoldContext` when they are empty | Unit test with fixture dirs + manual run |
| AC-5 | Codex is a first-class tool in Phase 1 (already listed); if selected, the wizard offers to verify `codex` is on `$PATH` and prints install hints if not found | Unit test (PATH check mock) + manual run |
| AC-6 | Global scope installs to `~/.config/opencode` (not `~/.ai`); `globalpaths.ResolveGlobalToolTargetDir(opencode, …)` returns `~/.config/opencode`; `ComputePlan` global target dir matches | Unit test + manual `ls ~/.config/opencode/` |
| AC-7 | Provider-command strategy is HYBRID: scaffold structure and file generation uses the Go shared-library + adapters; specialized generation or validation (e.g. `claude -p`, `codex exec`) is invoked only when the tool is installed and the adapter explicitly opts in | ADR + adapter interface extension test |

---

## Affected Files

### Go source files

| File | Change Category |
|------|---------------|
| `tui/wizard/phase1.go` | Add CLI-tools multi-select; add MCP-servers multi-select; add Codex install check |
| `tui/wizard/wizard.go` | Extend `WizardResult` / `WizardConfig` to carry CLI tools, MCP servers, metadata detection result |
| `tui/wizard/planner.go` | Fix `ComputePlan` global target dir from `~/.ai` → `~/.config/opencode` |
| `internal/adapter/opencode.go` | Move `opencode.json` write from project root to `.opencode/opencode.json`; adjust MCP compilation target to `.opencode/opencode.jsonc` |
| `internal/adapter/mcp_compiler.go` | `compileOpenCodeMCP` writes to `.opencode/opencode.jsonc` instead of root `opencode.jsonc` |
| `internal/adapter/types.go` | Add `CliTools []string` and `EnableServers []string` to `AdapterContext` |
| `internal/adapter/codex.go` | Add `DetectInstallation()` method; add install-hint printing in `Install()` |
| `internal/globalpaths/globalpaths.go` | Change `ResolveGlobalToolTargetDir(opencode)` to return `~/.config/opencode`; update `GlobalSetupDir()` to return `~/.config/opencode` |
| `internal/scaffold/types.go` | Add `DetectedMetadata` field to `ScaffoldContext` |
| `internal/scaffold/root.go` | Consume `DetectedMetadata` to fill `FragmentContext` placeholders |
| `internal/scaffold/mcp.go` | Accept `EnableServers` from wizard Phase 1 result instead of only from CLI tools |
| `internal/scaffold/scaffold.go` | Pass `EnableServers` from context step 2 |
| `cmd/helpers.go` | Wire CLI tools + MCP servers + detected metadata from wizard result into `ScaffoldContext` |
| `cmd/init.go` | Add `--cli-tools` and `--enable-servers` flags for non-interactive mode |
| `cmd/init_test.go` | Update test expectations for `.opencode/opencode.json` path |
| `internal/adapter/adapter_test.go` | Update `opencode.json` path assertion; add Codex detection test |
| `internal/globalpaths/globalpaths.go` (tests) | Update expected global path values |

### New files

| File | Purpose |
|------|---------|
| `internal/detect/detect.go` | Project metadata detection (language, framework, package manager) |
| `internal/detect/detect_test.go` | Tests for detection heuristics |
| `internal/detect/codex.go` | Codex installation detection + hint logic |
| `internal/detect/codex_test.go` | Tests for Codex PATH check |
| `tui/wizard/phase1_cli_mcp.go` | CLI Tools + MCP Servers multi-select form constructors (extracted from phase1.go to keep files ≤ ~300 lines) |
| `specs/005-setup-flow-fixes/research.md` | Decision log |
| `specs/005-setup-flow-fixes/data-model.md` | Entity/contract changes |
| `specs/005-setup-flow-fixes/quickstart.md` | Verification runbook |
| `specs/005-setup-flow-fixes/tasks/*.md` | Ordered task files |

### Library / template files

| File | Change |
|------|--------|
| `library/mcp/catalog.json` | No structural changes; the `cliTools` key already exists and is the data source |

---

## Implementation Phases

### Phase A — Config Path Fix (AC-1, AC-6)

**Task A-1**: Move OpenCode project config to `.opencode/opencode.json`
- Touch: `internal/adapter/opencode.go`
- Change: In `OpenCodeAdapter.Install()`, write `opencode.json` to `.opencode/opencode.json` instead of `<targetDir>/opencode.json`
- Update MCP compilation: `compileOpenCodeMCP` → write to `.opencode/opencode.jsonc`
- Touch: `internal/adapter/mcp_compiler.go`
- Update file record paths from `opencode.json` → `.opencode/opencode.json` and `opencode.jsonc` → `.opencode/opencode.jsonc`

**Task A-2**: Fix global scope target directory
- Touch: `internal/globalpaths/globalpaths.go`
- Change: `ResolveGlobalToolTargetDir(opencode, homeDir)` returns `filepath.Join(homeDir, ".config", "opencode")`
- Change: `GlobalSetupDir()` returns `~/.config/opencode` (or the XDG variant) instead of `~/.config/ai-setup`
- Touch: `tui/wizard/planner.go`
- Change: `ComputePlan` uses `globalpaths.ResolveGlobalToolTargetDir` instead of hardcoded `~/.ai`
- Update tests for both `globalpaths` and `planner`

**depends_on**: none
**parallel**: `[P]` — independent of all other phases

---

### Phase B — CLI Tools & MCP Server Selection Steps (AC-2, AC-3)

**Task B-1**: Add CLI Tools multi-select to Phase 1
- New file: `tui/wizard/phase1_cli_mcp.go`
- Read `library/mcp/catalog.json` → `cliTools` key
- Build `huh.NewMultiSelect[string]` with each `cliTools` entry (name, description, installHint as subtitle)
- Add `CLITools []string` to `Phase1Result`
- Touch: `tui/wizard/phase1.go` — call the new form after tool selection

**Task B-2**: Add MCP Servers multi-select to Phase 1
- Same file: `tui/wizard/phase1_cli_mcp.go`
- Read `library/mcp/catalog.json` → `servers` key
- Build `huh.NewMultiSelect[string]` with each server name + description
- Add `EnableServers []string` to `Phase1Result`
- For servers with `requiresInstall: true`, show install hint inline

**Task B-3**: Wire selections through wizard & scaffold
- Touch: `tui/wizard/wizard.go` — extend `WizardConfig` with `CLITools []string`, `EnableServers []string`
- Touch: `cmd/helpers.go` — `buildScaffoldContext` flows `Phase1Result.CLITools` and `EnableServers` into `ScaffoldContext`
- Touch: `cmd/init.go` — add `--cli-tools` and `--enable-servers` flags for non-interactive mode
- Touch: `internal/scaffold/scaffold.go` — pass `EnableServers` to `ScaffoldMcp` step

**depends_on**: none
**parallel**: `[P]` — independent of Phase A and Phase C

---

### Phase C — Project Metadata Detection (AC-4)

**Task C-1**: Create `internal/detect/detect.go`
- `DetectProjectMetadata(targetDir string) *ProjectMetadata`
- Probes in order (stop on first match):
  - `package.json` → language: TypeScript/JavaScript, package manager: npm/yarn/pnpm (from `packageManager` field)
  - `go.mod` → language: Go, framework: (module name heuristic)
  - `Cargo.toml` → language: Rust
  - `pyproject.toml` / `setup.py` / `requirements.txt` → language: Python
  - `pom.xml` → language: Java, framework: Maven
  - `Gemfile` → language: Ruby, framework: Rails (if `config/routes.rb` exists)
- Returns `ProjectMetadata{PrimaryLanguage, Framework, PackageManager, WorkspaceType}`

**Task C-2**: Integrate detection into wizard pipeline
- Touch: `tui/wizard/phase2.go` — After Phase 2 completes and before git conventions, run `detect.DetectProjectMetadata(config.TargetDir)` and use results as defaults for an optional "Detected project info" confirmation step
- Alternatively, run detection in `cmd/helpers.go` → `buildScaffoldContext()` and always inject into `ScaffoldContext.DetectedMetadata`
- Touch: `internal/scaffold/types.go` — add `DetectedMetadata *detect.ProjectMetadata`
- Touch: `internal/scaffold/root.go` — `ScaffoldCompiledRoot` uses `DetectedMetadata` as fallback for empty `PrimaryLanguage`, `Framework`, etc.

**depends_on**: none
**parallel**: `[P]` — independent of Phases A and B

---

### Phase D — Codex Coverage (AC-5)

**Task D-1**: Create `internal/detect/codex.go`
- `IsCodexInstalled() bool` — `exec.LookPath("codex")`
- `CodexInstallHint() string` — returns platform-specific install instructions
- `EnsureCodexOrPrompt() error` — if not installed, print hint and return nil (non-fatal)

**Task D-2**: Integrate Codex check into wizard
- Touch: `tui/wizard/phase1.go` — after tool multi-select, if `codex` is selected, run `detect.IsCodexInstalled()` and present a huh Confirm with install hint
- Touch: `internal/adapter/codex.go` — at top of `Install()`, call `EnsureCodexOrPrompt()` and log a warning if not found (non-blocking)

**depends_on**: Phase C (reuse `internal/detect/` package)
**parallel**: sequential after C

---

### Phase E — Global OpenCode Path (AC-6, continuation)

**Task E-1**: Wire `~/.config/opencode` through adapter
- Touch: `internal/adapter/opencode.go` — when `SetupScope == Global`, set `ocDir = filepath.Join(homeDir, ".config", "opencode")` using `globalpaths.ResolveGlobalToolTargetDir`
- Touch: `internal/adapter/mcp_compiler.go` — `compileOpenCodeMCP` global scope path: `~/.config/opencode/opencode.jsonc`

**depends_on**: Phase A (global path constants)
**parallel**: sequential after A

---

### Phase F — Hybrid Provider-Command Strategy (AC-7)

**Task F-1**: Extend ToolAdapter interface with optional provider-command hooks
- Touch: `internal/adapter/types.go`
  - Add `CanRunHeadless() bool` — returns true if this adapter supports headless CLI mode
  - Add `RunHeadlessValidation(ctx *AdapterContext) error` — no-op by default; adapters that support it (claude-code, codex) can run `claude -p` or `codex exec` for validation
  - Implement in `ClaudeCodeAdapter` and `CodexAdapter` as opt-in, no-op fallback for others
- This is the **HYBRID** approach: structure = Go shared-library + adapters (source of truth for files, paths, content); generation/validation = headless CLI mode only where natively supported and adapter opts in

**Task F-2**: Implement optional headless validation for Claude Code
- Touch: `internal/adapter/claudecode.go`
  - `CanRunHeadless() bool { return true }`
  - `RunHeadlessValidation(ctx *AdapterContext) error` — runs `claude -p "verify setup structure"` if `claude` is on PATH; logs warning and returns nil if not available

**Task F-3**: Implement optional headless validation for Codex
- Touch: `internal/adapter/codex.go`
  - `CanRunHeadless() bool { return true }`
  - `RunHeadlessValidation(ctx *AdapterContext) error` — runs `codex exec "check .agents/ structure"` if `codex` is on PATH; logs warning and returns nil if not available

**depends_on**: Phase D (Codex detection)
**parallel**: sequential after D

---

## Decision Protocol: Hybrid Provider-Command Strategy

### Approaches Considered

**Option A: Pure Library (Go only)**
- Approach: Every adapter generates all files through Go code reading the library FS. No CLI subprocess calls.
- Pros: Deterministic, testable offline, no tool dependency at setup time
- Cons: Cannot validate that generated files are actually valid for the target tool; drift between adapter output and tool expectations goes undetected

**Option B: Pure CLI (Delegated)**
- Approach: Invoke `claude init`, `codex setup`, etc. directly; wrapper scripts only.
- Pros: Always produces tool-native output
- Cons: Requires every tool installed before setup; no offline generation; tools have no standardized init interface; fragile across versions

**Option C: Hybrid (Library + Selective CLI)**
- Approach: Go shared-library + adapters are the source of truth for structure (files, paths, content); headless CLI modes invoked only for validation/generation where a tool natively supports it (`claude -p`, `codex exec`)
- Pros: Deterministic core with optional validation; works offline; catches structural drift when tools are present; progressive enhancement
- Cons: Two code paths means two things to test; adapter interface slightly more complex

### Decision: Option C (Hybrid)

**Rationale**: The Go library must be the source of truth because (1) users may not have every tool installed at setup time, (2) headless CLI modes are not available for all tools (copilot, gemini), and (3) deterministic generation is required for `--dry-run` and conflict resolution. However, Claude Code and Codex natively support headless modes that can validate or supplement generated output — ignoring that capability wastes a free correctness check.

**Tradeoff accepted**: The `CanRunHeadless` / `RunHeadlessValidation` extension adds ~2 methods to the adapter interface. The validation path must be explicitly non-fatal (log + continue) to avoid breaking setups where the tool isn't installed or the headless mode is unavailable.

**Record as ADR**: Yes — recorded in `specs/005-setup-flow-fixes/research.md`

---

## Verification Strategy

### Automated tests

| Layer | What | How |
|-------|------|-----|
| Unit | `internal/detect/detect.go` | Table-driven tests with fixture dirs containing `package.json`, `go.mod`, etc. |
| Unit | `internal/detect/codex.go` | Mock `exec.LookPath` via build-tag injection or `os.Getenv` override |
| Unit | `internal/globalpaths/globalpaths.go` | Assert `ResolveGlobalToolTargetDir(opencode, "/home/u")` == `/home/u/.config/opencode` |
| Unit | `internal/adapter/opencode.go` | Assert `opencode.json` written to `.opencode/opencode.json` not root |
| Unit | `internal/adapter/mcp_compiler.go` | Assert OpenCode MCP compilation writes to `.opencode/opencode.jsonc` |
| Unit | `tui/wizard/phase1.go` + `phase1_cli_mcp.go` | Non-interactive path: verify `Phase1Result.CLITools` and `EnableServers` populated from defaults |
| Integration | Full `init` pipeline | Non-interactive run with `--scope project --tools opencode --cli-tools gh --enable-servers memory,ripgrep --name test-proj` and assert file tree |

### Manual checks

1. **Interactive run**: `ai-setup init` — verify Phase 1 now shows CLI tools and MCP servers multi-selects
2. **Config location**: After project scope run, `ls .opencode/opencode.json` exists; `ls opencode.json` does NOT exist at root
3. **Global scope**: `ai-setup init --scope global --tools opencode` → files appear in `~/.config/opencode/`
4. **Metadata detection**: Run `init` in a directory with `package.json` — verify `PrimaryLanguage` appears in `AGENTS.md`
5. **Codex check**: With codex NOT on PATH, select it in Phase 1 — verify install hint prints
6. **Headless validation**: With `claude` on PATH, run `init` selecting `claude-code` — verify `RunHeadlessValidation` logs a validation attempt

### Test commands

```bash
# Go quality gates
go vet ./...
go test ./internal/... ./tui/... ./cmd/... -count=1 -v

# Specific test targets
go test ./internal/detect/... -v
go test ./internal/globalpaths/... -v
go test ./internal/adapter/... -v
go test ./tui/wizard/... -v
```

---

## Risks & Mitigations

| # | Risk | Severity | Mitigation |
|---|------|----------|------------|
| R-1 | Moving `opencode.json` breaks existing projects that reference it at root | High | Add a migration path: if root `opencode.json` exists and `.opencode/opencode.json` does not, copy it during upgrade and log a deprecation notice. `ai-setup doctor` detects the stale root file. |
| R-2 | MCP catalog JSON format changes break wizard selection | Medium | Parse defensively — if `cliTools` or `servers` key is missing, skip the step gracefully. Add schema version field to catalog in a follow-up. |
| R-3 | Project detection false positives (e.g., a `package.json` in a submodule) | Low | Detection only sets defaults; the wizard confirmation step lets users override. Detection stops on first match in target dir (non-recursive). |
| R-4 | `exec.LookPath("codex")` varies across OS / install methods | Medium | Fall back to `which codex` / `where codex` on error; make detection non-blocking — install hint is always shown. |
| R-5 | Headless CLI validation hangs or prompts interactively | Medium | `RunHeadlessValidation` must use a timeout (30s) and pipe `/dev/null` to stdin. Non-fatal on any error. |
| R-6 | Global scope path change from `~/.ai` to `~/.config/opencode` breaks existing global setups | High | Add `ai-setup migrate --scope global` command in a follow-up spec. For now, `doctor` detects stale `~/.ai` and reports a recommendation. Both paths are checked during `init`. |
| R-7 | Phase 1 form becomes too long with 3 additional multi-selects | Medium | Group CLI tools and MCP servers into a sub-screen (separate huh form group) with a "Configure integrations" prompt. Use scrollable multi-select. |

---

## Dependency Graph

```
Phase A ──→ Phase E
Phase B ──┐
Phase C ──→ Phase D ──→ Phase F
           │
           └─── (all phases complete) → manual verification → done
```

- A and B and C can run in parallel (`[P]`)
- D depends on C (reuse `internal/detect/`)
- E depends on A (global path constants)
- F depends on D (Codex detection)

---

## Out of Scope

- Full migration tooling for existing global installs (separate spec)
- MCP catalog schema versioning
- Custom MCP server entry UI (just select from catalog)
- Non-catalog CLI tool support (only catalog entries)
- Headless generation (not just validation) — deferred until tools stabilize their CLI interfaces