# Research — Compile-time Scope Awareness & Artifact Parity

**Date:** 2026-04-18
**Type:** Mixed (bugfix for §1, feature for §2, tech-debt for §3, docs for §4)
**Depends on:** Spec 008 (CLI tool structure parity — merged)

---

## Context

Spec 008 gave `ai-setup init` per-scope routing. But follow-up items remain:

1. **Compile path is scope-blind** — `ai-setup compile` writes per-tool MCP configs to project-relative paths at every scope. At `scope=global`, `~/.codex/config.toml` should receive `[mcp_servers.*]` enrichment; today it goes to `<project>/.codex/config.toml`.
2. **Missing artifact types** — Gemini custom commands (`.gemini/commands/*.toml`) and Copilot chat modes (`.github/chatmodes/*.chatmode.md`) are not scaffolded.
3. **No snapshot tests for library assets or compiled output** — assertions use existence checks and string contains.
4. **`CLAUDE.md` template placeholders** still present (18 `[YOUR_*]` fields) for this project itself.

Also scoped: **`--drive-cli` for tools other than Gemini** — narrow applicability; most tools lack non-interactive MCP CLIs. Low value.

---

## 1. Compile-time Scope Awareness — Findings

### Current flow

- `cmd/compile.go:39-42` — `targetDir` comes from `--dir` flag, default `os.Getwd()`. Always the project root. The store at `<targetDir>/.ai-setup.db` is opened.
- `cmd/compile.go:79` — `storeData.Config.SetupScope` is read into `storeData` but **never passed to adapters**.
- `cmd/compile.go:173` — `adapt.CompileMCP(dir, newFileRecords)` — interface carries no scope.

### Per-adapter compile functions (all project-relative)

| Function | Target path | Scope-aware? |
|---|---|---|
| `compileOpenCodeMCP` (mcp_compiler.go:107) | `<targetDir>/.opencode/opencode.jsonc` | Partial — `isGlobalOpenCodeDir(dir)` heuristic only triggers if caller runs `ai-setup compile` from inside `~/.config/opencode` (unused path) |
| `compileClaudeCodeMCP` (mcp_compiler.go:179) | `<targetDir>/.mcp.json` | No |
| `compileCopilotMCP` (mcp_compiler.go:221) | `<targetDir>/.vscode/mcp.json` | No |
| `compileGeminiMCP` (mcp_compiler.go:269) | `<targetDir>/.gemini/settings.json` | No |
| `compileCodexMCP` (mcp_compiler.go:287, just added) | `<targetDir>/.codex/config.toml` | No |

### `ToolAdapter` interface

```go
CompileMCP(targetDir string, fileRecords []types.TrackedFile) ([]types.TrackedFile, error)
```

No scope, no HomeDir. Extending it is the minimal fix. All implementations are internal (5 adapters), so the interface break is contained.

### Fix options

- **A. Extend interface** — `CompileMCP(ctx CompileContext) (...)` where `CompileContext` carries `TargetDir, HomeDir, SetupScope`. Mirror the `AdapterContext` pattern. ~50 LOC per adapter × 5 + interface change. Clean but churny.
- **B. Add a second method** — `CompileMCPWithScope(ctx CompileContext)` alongside `CompileMCP(targetDir)`; deprecate the old one. Two methods, same problem.
- **C. Reuse `ResolveToolRoot`** — adapters receive `(targetDir, homeDir, scope)` and reuse the scope resolver internally. Same as A but phrased around existing primitive.

**Recommend A or C** — they're equivalent functionally.

### Risk

- `isGlobalOpenCodeDir` becomes dead code. Remove or keep as safety net.
- `.mcp.json` at `<project>/` is the Claude project-level MCP file (committed by users). At global scope we want `~/.claude/settings.json` mcpServers merge instead. `claudecode.go` Install already handles settings.json via configmerge — CompileMCP duplicates logic. Need to reconcile.

---

## 2. New Artifact Types — Findings

### Gemini custom commands

Gemini CLI supports user-defined slash commands via TOML files:
- Global: `~/.gemini/commands/<name>.toml`
- Project: `.gemini/commands/<name>.toml`

Format:
```toml
name = "rpi"
description = "Start RPI workflow"
prompt = "Begin the RPI flow..."
```

**Current state:** no `library/commands/` directory; `gemini.go` doesn't create a `commands/` subdir; `AdapterSelections` has no `Commands` field.

### Copilot chat modes

GitHub Copilot supports user-defined chat modes via markdown files:
- Project: `.github/chatmodes/<name>.chatmode.md`
- No global concept (chatmodes are project-scoped in Copilot)

Format: markdown file with YAML frontmatter describing the mode.

**Current state:** no `library/chatmodes/`; `copilot.go` handles `instructions/` and `prompts/` only; `AdapterSelections` has no `ChatModes` field.

### `AdapterSelections` (types.go:46-51)

```go
type AdapterSelections struct {
    Agents  []types.AgentId
    Skills  []types.SkillId
    Prompts []types.PromptId
}
```

Adding `Commands []types.CommandId` and `ChatModes []types.ChatModeId` is additive — no existing code breaks.

### Scope interaction

- Gemini commands: project + workspace + global (all supported)
- Copilot chatmodes: project + workspace only (no upstream global concept). Copilot global scope is already unsupported, so no edge case here.

---

## 3. Snapshot Tests — Findings

- No `testdata/`, `golden`, `cupaloy`, `goldie` in the repo.
- `internal/library/embed_test.go:1-138` uses `fs.ReadDir` walking + string-contains assertions for asset shape.
- Introducing snapshots is a full-on testing-infra decision. Options:
  - **stdlib only** — write `testdata/*.golden` manually, compare strings, `-update` flag
  - **`bradleyjkemp/cupaloy`** — minimal lib, golden files, `-update` flag
  - **`sebdah/goldie/v2`** — popular, more features
- **Recommendation:** defer unless a concrete churn-driver appears. The existing tests catch regressions fine; snapshots add update-friction.

---

## 4. CLAUDE.md Placeholders — Findings

18 remaining `[YOUR_*]` fields across 9 sections (Persona, Overview, Stack, Architecture, Conventions, Rules, DoNot, Testing, Commands, Session Checks).

Available sources:
- `package.json` — name, version, description, scripts (for `[YOUR_*_COMMAND]` fields)
- `go.mod` — module name, Go version
- `README.md` (48k) — likely has project description, architecture
- `AGENTS.md` (10k, gitignored) — may have testing/workflow details

**Recommendation:** lightweight fill from package.json + README; leave subjective `[YOUR_RULE_*]` for user to decide.

---

## 5. `--drive-cli` for Other Tools

| Tool | Has non-interactive MCP CLI? | Worth driving? |
|---|---|---|
| Gemini | Yes — `gemini mcp add` (already done) | ✅ |
| Claude Code | Yes — `claude mcp add` | Maybe — different output (settings.json, not a CLI-managed file) |
| OpenCode | Unknown / limited | ⛔ |
| Copilot | No | ⛔ |
| Codex | No | ⛔ |

**Recommendation:** add Claude Code `claude mcp add` support, skip the rest.

---

## Summary — What's Worth Doing

| # | Item | Value | Effort | Recommend |
|---|---|---|---|---|
| 1 | Compile-time scope awareness | **High** — silent data loss at global scope today | Medium (~300 LOC, 5 adapters) | **Do** |
| 2a | Gemini commands artifact type | Medium — genuine feature gap | Medium (~150 LOC + templates) | **Do** |
| 2b | Copilot chatmodes artifact type | Medium — genuine feature gap | Medium (~100 LOC + templates) | **Do** |
| 3 | Snapshot tests | Low — current tests sufficient | Medium | **Defer** |
| 4 | CLAUDE.md placeholders | Low — this is a template | Small | **Do partially** |
| 5 | `--drive-cli` for Claude Code | Low — plumbing already exists | Small (~40 LOC) | **Do** |

**Proposed scope for spec 009:** items 1, 2a, 2b, 4 (partial), 5. Defer 3.

---

## Open Questions (for HUMAN GATE)

1. **Compile interface:** break `ToolAdapter.CompileMCP` signature (clean, one-shot migration) or add parallel method? → recommend break (5 internal implementers).
2. **New artifact types:** commit to Gemini commands + Copilot chatmodes now, or split into its own spec? → recommend bundle (they share patterns with existing skill/prompt copying).
3. **Snapshot tests:** defer? → recommend defer.
4. **CLAUDE.md placeholders:** auto-fill from package.json/README, or leave for user? → recommend auto-fill mechanical fields, leave `[YOUR_RULE_*]` for user.
5. **Claude Code `--drive-cli`:** bundle with spec 009, or defer? → recommend bundle (reuses Gemini scaffolding).

---

⛔ **HUMAN GATE — approve research before planning.**
