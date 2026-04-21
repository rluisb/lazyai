# Plan — AI CLI Tool Structure Parity (Spec 008)

## Context

`ai-setup` scaffolds files for five AI CLI tools (`claude-code`, `opencode`, `gemini`, `copilot`, `codex`) across three scopes (`project`, `global`, `workspace`). Research at `specs/008-cli-tool-structure-parity/research.md` documents seven concrete gaps in today's behaviour:

- **Workspace scope is silently ignored** — all five adapters fall back to project layout.
- **Scope guards are missing** — `gemini.go` / `copilot.go` write project-shaped output at the wrong path when scope=global; `codex.go:37` uses a `filepath.Dir(ctx.TargetDir)/.agents/` fallback that matches no upstream convention.
- **Global-path registry is incomplete** — `internal/globalpaths/globalpaths.go` only resolves globals for OpenCode and Claude Code.
- **Codex never emits `config.toml`** — MCP server registration for Codex lives in `~/.codex/config.toml` `[mcp_servers.*]`; nothing writes it.
- **Codex conflates `.agents/` and `.codex/`** — skills live in `.agents/skills/`, settings in `.codex/config.toml`; current code emits only `.agents/`.
- **Root memory doc is scope-blind** — `scaffold/root.go:168` always writes under `<target>`.
- **Settings files are overwritten without merging** — user-authored `settings.json` / `opencode.json` / `.codex/config.toml` keys are clobbered.

The user wants each tool's canonical on-disk structure scaffolded correctly for the chosen scope, with "workspace" interpreted as "project-shaped layout rooted at the user-selected workspace directory" (no tool-native workspace concept exists upstream). Evaluated driving each tool's own CLI for bootstrapping — only Gemini exposes non-interactive subcommands; deferred.

### Locked decisions (user interview, 2026-04-17)

1. **Workspace memory doc** = root-only at workspace dir; no mirroring into `Config.Repos[i]`.
2. **Incompatible tool × scope** (e.g. Copilot × global) = interactive wizard filters the multi-select; non-interactive mode WARNs per skipped tool and proceeds, exits non-zero only if the result is empty.
3. **Codex global layout** = upstream split: `~/.codex/config.toml` + `~/.codex/AGENTS.md` + `~/.agents/skills/<name>/SKILL.md`.
4. **Merge policy** for shared config files = deep-merge with backup-on-first-touch (`.bak` written once; never overwritten).
5. **CLI-driven scaffolding** (`gemini mcp add` etc.) = out of Wave 1; direct-write only.
6. **Spec directory** = keep `specs/008-cli-tool-structure-parity/` (flat); reconcile CLAUDE.md mismatch via follow-up tech-debt.

---

## Acceptance Criteria

| # | AC | Verified by |
|---|---|---|
| AC-1 | `adapter.ResolveToolRoot(tool, scope, ctx)` returns the canonical path from research §2 for every (tool, scope) pair; returns `ErrScopeUnsupported` for Copilot × global. | Table test covering 15 pairs + Codex split-root helper |
| AC-2 | `adapter.IsScopeSupported(tool, scope)` returns `false` exactly for Copilot × global. | Same table test |
| AC-3 | `globalpaths.ResolveGlobalToolTargetDir` returns canonical roots for all five tools (or `""` for Copilot). | `globalpaths_test.go` |
| AC-4 | Each adapter's `Install()` writes only under its resolved root (+ documented siblings — Codex skills, Copilot `.github/`). No writes under `<project>/` when scope=global; none under `~/` when scope=project/workspace. | Per-adapter scope-parity tests with `t.TempDir()` |
| AC-5 | Codex adapter emits `config.toml` at the correct scope-specific path. | Codex scope test + fixture |
| AC-6 | Memory doc placement obeys scope: project/workspace → `<root>/<MemoryFile>`; global → tool's global root (or skip for Copilot). | `scaffold/root_test.go` |
| AC-7 | Re-running `Install()` against a pre-existing user-authored config file preserves user keys, overlays ai-setup's keys, writes `.bak` only on the first touch; subsequent runs are no-ops on `.bak` and on content. | `internal/configmerge/configmerge_test.go` |
| AC-8 | Interactive wizard's tool multi-select filters out scope-incompatible tools; re-entering scope refreshes the filter. | `tui/wizard/phase1_test.go` |
| AC-9 | Non-interactive `ai-setup init --scope global --tools copilot,claude-code …` prints one WARN per unsupported tool, installs the rest, exits 0. With only unsupported tools → exits non-zero, installs nothing. | `cmd/init_test.go` + manual smoke |
| AC-10 | No changes to `StoreData.Config.SetupScope` schema, `ToolAdapter` interface, or non-interactive flag surface. | Code review + `go test ./internal/types/... ./cmd/...` |
| AC-11 | `go vet ./...` and `go test ./... -count=1` green after every wave. | CI gate |

---

## Approach

**Chosen — scope resolver + adapter sweep + merge helper.** One pure function `ResolveToolRoot(tool, scope, ctx)` becomes the single source of truth every adapter writes into; each adapter's `if isGlobal { … }` branch collapses to a linear call. A companion `IsScopeSupported` exposes the same decision to wizard + non-interactive callers. A tiny `internal/configmerge/` package handles JSON/JSONC/TOML read-modify-write with backup-on-first-touch (~150 LOC, reuses `internal/jsonc/`, needs `github.com/BurntSushi/toml`).

Advantages: minimal interface churn (no new `ToolAdapter` method, no schema bump), localized blast radius per adapter, testable (both primitives are pure).

**Rejected alternative — populate `ScopePaths` struct on `AdapterContext`.** Would require a new interface method or breaking change, threads a new shape through `scaffold.go`'s dispatch site, and doesn't reduce adapter code enough to justify it. Revisit when a sixth tool arrives.

---

## Critical Files

### Modified

| File | Change |
|---|---|
| `internal/globalpaths/globalpaths.go` | Add gemini (`~/.gemini`) and codex (`~/.codex`) to `ResolveGlobalToolTargetDir`; add `ResolveCodexSkillsGlobalDir(homeDir)` → `~/.agents/skills`; `IsGlobalSupportedTool` returns `true` for all tools except Copilot. |
| `internal/adapter/claudecode.go` | Replace `:24-28` `isGlobal` branch with `ResolveToolRoot`; delete memory-doc override at `:134-137` (moves to scaffold/root.go); swap `WriteJSONFile` → `configmerge.MergeJSONFile`. |
| `internal/adapter/opencode.go` | Replace `:24-42` branch with `ResolveToolRoot`; remove `if !isGlobal` gate around `opencode.json` write (global also writes, via merge helper); swap to `MergeJSONFile`. |
| `internal/adapter/gemini.go` | Add scope guard; compute `geminiDir` via `ResolveToolRoot`; `settings.json` via merge helper; delete `InstallRootTemplateIfMissing("GEMINI.md", …)` at `:77-81`. |
| `internal/adapter/copilot.go` | Early-return `ErrScopeUnsupported` when `scope=global`; compute `githubDir` via `ResolveToolRoot`; delete `InstallRootTemplateIfMissing("AGENTS.md", …)` at `:61-65`. |
| `internal/adapter/codex.go` | Use new `ResolveCodexRoots(scope, ctx)` split-root helper; delete `:34-38`; emit `config.toml` via `configmerge.MergeTOMLFile`; write skills under the skills root; delete adapter-level memory-doc install at `:81-85`. |
| `internal/scaffold/root.go` | New `memoryDocDestPath(tool, scope, targetDir, homeDir)` helper replaces `filepath.Join(opts.TargetDir, outputFile)` at `:168`; global scope routes through `globalpaths.ResolveGlobalToolTargetDir`; Copilot × global skipped with a log line. |
| `internal/adapter/types.go` | Add `var ErrScopeUnsupported = errors.New(...)`. No interface change. |
| `tui/wizard/phase1.go` | `askTools` builds options from `registry.AllTools()` filtered by `IsScopeSupported(tool, scope)`. |
| `cmd/init.go` | After parsing `--tools`, filter non-interactively via `IsScopeSupported`; WARN per dropped tool; error only on empty result. |

### New

| File | Purpose |
|---|---|
| `internal/adapter/scope.go` | `ResolveToolRoot`, `ResolveCodexRoots`, `IsScopeSupported`, `ErrScopeUnsupported` |
| `internal/adapter/scope_test.go` | Table test for all 15 (tool, scope) pairs + Codex split |
| `internal/adapter/{claudecode,opencode,gemini,copilot,codex}_scope_test.go` | Per-adapter scope-parity tests |
| `internal/configmerge/configmerge.go` | `MergeJSONFile(path, patch)` and `MergeTOMLFile(path, patch)` with backup-on-first-touch; deep-merge maps recurse; slices replaced wholesale |
| `internal/configmerge/configmerge_test.go` | JSON + TOML fixtures; idempotency; user-keys-preserved; `.bak` appears exactly once |
| `internal/globalpaths/globalpaths_test.go` | Covers new entries + negative Copilot case |
| `internal/scaffold/root_test.go` | (tool × scope) → expected dest path, including Copilot × global skip |
| `cmd/init_test.go` | Non-interactive scope-filter behaviour |
| `tui/wizard/phase1_test.go` | Extend (from spec 007) with a scope-filter subtest |
| `specs/008-cli-tool-structure-parity/plan.md` | Full version of this plan, committed alongside the implementation |

### Explicitly Not Touched

- `StoreData` / `types.Config` / `types.SetupScope` schema.
- `ToolAdapter` interface — no new methods.
- `internal/scaffold/repos.go` — workspace multi-repo scaffolding stays as-is.
- `library/**` — no asset changes.
- Non-interactive flag surface — no new flags.
- `CompileMCP` implementations — Codex `config.toml` `[mcp_servers.*]` enrichment via `CompileMCP` is a follow-up, not Wave 1.

### Existing Primitives to Reuse

- `internal/jsonc/` — JSONC read/write (Claude/OpenCode/Gemini settings files).
- `internal/files/` — `CopyFile` for `.bak` creation.
- `internal/globalpaths/` — extended, not replaced.
- `internal/conflict/` — existing conflict-strategy types used by adapter conflict handling; merge helper complements but does not replace.

---

## Implementation Waves

### Wave 1a — Scope resolver (AC-1, AC-2, AC-3, AC-10)

- **1a-1:** Extend `globalpaths.go` with Gemini/Codex globals + Codex-skills helper; update `IsGlobalSupportedTool`.
- **1a-2:** Create `internal/adapter/scope.go` with `ResolveToolRoot`, `ResolveCodexRoots` (split), `IsScopeSupported`, `ErrScopeUnsupported`.
- **1a-3:** Add `scope_test.go` covering all 15 (tool, scope) pairs plus Codex split.
- **1a-4:** Add `globalpaths_test.go`.
- Exit: new files compile; targeted tests green. No adapter behaviour change yet.

### Wave 1b — Adapter parity sweep (AC-4, AC-5)

One task per adapter. Each is an independent PR-sized commit. All gate behind 1a + 1d.

- **1b-1 claudecode.go:** swap to `ResolveToolRoot`; delete `:134-137` override; `settings.json` via `MergeJSONFile`.
- **1b-2 opencode.go:** swap to `ResolveToolRoot`; `opencode.json` now always writes (global included) via `MergeJSONFile`.
- **1b-3 gemini.go:** swap to `ResolveToolRoot`; `settings.json` via `MergeJSONFile`; drop memory-doc install.
- **1b-4 copilot.go:** add `ErrScopeUnsupported` early-return; swap to `ResolveToolRoot`; drop memory-doc install.
- **1b-5 codex.go:** use `ResolveCodexRoots`; emit `config.toml` via `MergeTOMLFile` with minimal `[mcp_servers]`; skills under skills root; drop memory-doc install.
- Exit: per-adapter `*_scope_test.go` green.

### Wave 1c — Memory doc via scaffold/root.go (AC-6)

- **1c-1:** Add `memoryDocDestPath(tool, scope, targetDir, homeDir)` helper.
- **1c-2:** Replace `:168` `filepath.Join(opts.TargetDir, outputFile)` with the helper; skip Copilot × global with a log line.
- **1c-3:** Delete now-redundant per-adapter `InstallRootTemplateIfMissing` memory-doc calls (tracked by the 1b tasks).
- **1c-4:** Add `root_test.go` covering the path table.
- Exit: single emitter; no adapter writes memory docs.

### Wave 1d — Config merge helper (AC-7)

- **1d-1:** Create `internal/configmerge/configmerge.go` with `MergeJSONFile` (uses `jsonc`) and `MergeTOMLFile` (uses `BurntSushi/toml`). Deep-merge: maps recurse, slices replaced wholesale. Backup: if `.bak` missing, copy first; if present, do not overwrite. Idempotent sorted output.
- **1d-2:** Fixture tests per AC-7.
- **1d-3:** Adapters in 1b call the helpers.
- Exit: merge tests green.

### Wave 1e — Wizard + non-interactive gating (AC-8, AC-9)

- **1e-1:** `tui/wizard/phase1.go` `askTools` filters options by `IsScopeSupported(tool, scope)`.
- **1e-2:** Verify filter refreshes when user backs out and changes scope.
- **1e-3:** `cmd/init.go` filters CLI `--tools` by `IsScopeSupported`, WARNs per drop, errors only on empty.
- **1e-4:** Unit tests for non-interactive filter (mixed list + all-unsupported).
- **1e-5:** Wizard scope-filter unit test.

### Wave 1f — Test sweep (AC-11)

- **1f-1:** Per-adapter `TestInstall_ScopeParity` (temp dirs for `TargetDir` and `HomeDir`; assert expected paths exist; no unexpected paths; idempotent on re-run).
- **1f-2:** Confirm `configmerge` fixture coverage.
- **1f-3/1f-4:** Wizard + non-interactive tests per 1e.
- Exit: `go vet ./... && go test ./... -count=1` green.

### Dependency graph

```
1a ──┬──▶ 1b ◀── 1d
     ├──▶ 1c
     └──▶ 1e
                   ▼
                   1f
```

Suggested merge order: 1a → 1d → 1b-{1..5} (parallel OK) → 1c → 1e → 1f.

---

## Verification

### Automated

```bash
go vet ./...
go test ./internal/adapter/... ./internal/globalpaths/... ./internal/configmerge/... ./internal/scaffold/... -count=1 -v
go test ./... -count=1
go build ./...
```

### Manual smoke (`ai-setup init`)

1. **Project, all tools:** `ai-setup init --scope project --tools claude-code,opencode,gemini,codex,copilot --non-interactive` → under `$PWD`: `.claude/`, `.opencode/` + `opencode.json`, `.gemini/`, `.codex/config.toml` + `.agents/skills/`, `.github/copilot-instructions.md`. Memory docs at repo root.
2. **Global, mixed tools:** `ai-setup init --scope global --tools claude-code,opencode,gemini,codex,copilot --non-interactive` → writes under `~/.claude/`, `~/.config/opencode/`, `~/.gemini/`, `~/.codex/` + `~/.agents/skills/`. Prints `WARN: skipping tool "copilot" for scope "global"`. Exits 0.
3. **Global, only-copilot:** exits non-zero; writes nothing.
4. **Workspace:** `--scope workspace --name my-ws` → same structure as project, rooted at `$PWD`; no mirroring into `Config.Repos[i]`.
5. **Merge preservation:** hand-author `~/.claude/settings.json` with a custom `experimental` key; run (2); assert key survives and `~/.claude/settings.json.bak` exists. Re-run; assert `.bak` content unchanged.
6. **Wizard filter:** interactive `scope=global` → Copilot hidden from tool multi-select. Back up, switch to `project` → Copilot reappears.

---

## Risks

| # | Risk | Severity | Mitigation |
|---|---|---|---|
| R-1 | New direct dep `github.com/BurntSushi/toml` | Low | `go mod why` to check if transitive; add as direct dep with justification if needed. It's the de-facto Go TOML lib. |
| R-2 | Deep-merge "slices replaced wholesale" surprises users with large MCP lists | Medium | Document in `configmerge` godoc; AC-7 covers the expected behaviour. |
| R-3 | Scope tests read real `os.UserHomeDir()` instead of `ctx.HomeDir` | Medium | Reviewer checklist enforces `t.TempDir()` for HomeDir; grep gate in CI. |
| R-4 | Memory-doc move leaves orphaned files from prior installs on global scope | Low | One-shot orphan, not deleted; document in release notes. |
| R-5 | Wizard scope filter regresses spec-007 tests | Medium | Extend, don't rewrite; `go test ./tui/wizard/... -count=1` before every merge. |
| R-6 | Non-interactive WARN-then-proceed masks typos | Medium | Filter only drops *recognized* tools; unknown IDs still error out via pre-existing `types.IsValidToolId`. |
| R-7 | Codex `config.toml` upstream schema drift | Low | Wave 1 emits a minimal skeleton; merge helper preserves user additions. |
| R-8 | `ResolveToolRoot` placement (`adapter/` vs `globalpaths/`) | Low | `adapter/` owns it because it needs `AdapterContext`; `scaffold/root.go` uses `globalpaths` directly. |
| R-9 | Deleting adapter-level memory-doc writes breaks users who rely on them | Low | `scaffold/root.go` becomes the single emitter; regression covered by AC-6 test table + manual smoke #1. |

---

## Out of Scope (explicit)

- CLI-driven scaffolding via each tool's own subcommands (`gemini mcp add`, etc.).
- New artifact types per adapter (e.g. Gemini `.toml` commands, Copilot `chatmodes/`).
- Multi-repo workspace scaffolding (already in `scaffold/repos.go`).
- CLAUDE.md ↔ `specs/` directory-convention reconciliation (follow-up tech-debt task).
- Snapshot tests for library content.
- Codex `config.toml` `[mcp_servers.*]` enrichment via `CompileMCP` (follow-up).
- `AGENTS.override.md` creation for Codex (user-authored surface).
