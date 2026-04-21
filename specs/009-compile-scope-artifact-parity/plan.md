# Plan — Compile-time Scope Awareness & Artifact Parity (Spec 009)

**Date:** 2026-04-18
**Depends on:** Spec 008 merged (scope resolver + configmerge)
**Locked decisions (user, 2026-04-18):** break CompileMCP signature · bundle Gemini commands + Copilot chatmodes · defer snapshot tests · auto-fill mechanical CLAUDE.md fields · bundle Claude Code `--drive-cli`

---

## Context

`ai-setup compile` regenerates per-tool MCP configs from the canonical `.ai/mcp.json`. Post-spec-008, `init` routes per scope but `compile` still writes to project-relative paths at every scope — silent data loss at global scope. Two tool features are also unscaffolded: Gemini custom commands and Copilot chat modes. This spec closes both gaps and delivers three smaller items (Claude Code `--drive-cli`, CLAUDE.md placeholder fill, dead-code removal).

---

## Acceptance Criteria

| # | AC | Verified by |
|---|---|---|
| AC-1 | `ToolAdapter.CompileMCP` accepts a `CompileContext{TargetDir, HomeDir, SetupScope, FileRecords}`; all 5 adapters implement it. | compile-time type check |
| AC-2 | For each (tool, scope) pair, `CompileMCP` writes to the scope-correct path (`ResolveToolRoot` result) or skips cleanly (Copilot × global). No writes under project when scope=global; no writes under home when scope=project/workspace. | `mcp_compiler_scope_test.go` (15 pairs + Copilot skip) |
| AC-3 | Merging MCP servers into existing tool config preserves user keys via `configmerge.Merge*File`; `.bak` written once. | extend existing configmerge tests |
| AC-4 | `isGlobalOpenCodeDir` removed; no callers remain. | `go vet ./...` + grep |
| AC-5 | `AdapterSelections` gains `Commands []types.CommandId` and `ChatModes []types.ChatModeId`; `AdapterContext` threads both; scaffold/store layer persists them. | types test + db round-trip test |
| AC-6 | Gemini adapter installs commands from `library/commands/*.toml` to `<geminiRoot>/commands/<name>.toml` at every supported scope. | `gemini_commands_test.go` |
| AC-7 | Copilot adapter installs chatmodes from `library/chatmodes/*.chatmode.md` to `<githubDir>/chatmodes/<name>.chatmode.md` at project/workspace scope only. | `copilot_chatmodes_test.go` |
| AC-8 | When `DriveCLI=true` and `claude` binary is on PATH, ClaudeCodeAdapter calls `claude mcp add` for each enabled server. Falls back to direct-write otherwise. | `claudecode_drivecli_test.go` (stub binary) |
| AC-9 | `CLAUDE.md` has `[YOUR_*]` placeholders filled for: project name, description, tech stack, dev/test/build commands. Subjective fields (rules, session checks) remain as placeholders with a `<!-- fill-in -->` marker. | manual diff review |
| AC-10 | `go vet ./... && go test ./... -count=1` green after every wave. | CI gate |

---

## Approach

### Wave A — CompileContext + scope-aware MCP compile (foundation)

**New type** in `internal/adapter/types.go`:
```go
type CompileContext struct {
    TargetDir   string
    HomeDir     string
    SetupScope  types.SetupScope
    FileRecords []types.TrackedFile
}
```

**Interface change** (`ToolAdapter`):
```go
// Before:
CompileMCP(targetDir string, fileRecords []types.TrackedFile) ([]types.TrackedFile, error)
// After:
CompileMCP(ctx CompileContext) ([]types.TrackedFile, error)
```

**Per-tool target-path resolution** (reuse `ResolveToolRoot`):

| Tool | Project/Workspace | Global |
|---|---|---|
| claude-code | `<target>/.mcp.json` *(preserve existing user-managed file)* | skip compile step — init's `settings.json` merge already handles mcpServers via configmerge |
| opencode | `<target>/.opencode/opencode.jsonc` | `<home>/.config/opencode/opencode.jsonc` |
| gemini | `<target>/.gemini/settings.json` | `<home>/.gemini/settings.json` |
| copilot | `<target>/.vscode/mcp.json` | **skip** (Copilot × global unsupported) |
| codex | `<target>/.codex/config.toml` | `<home>/.codex/config.toml` |

**Note on claude-code:** The existing `.mcp.json` output is a user-committed project file with a different schema than `settings.json`. Keep project-scope behaviour unchanged; at global scope, skip the `.mcp.json` write and let init's `settings.json` merge cover mcpServers. Documented in code comment.

**Dead code removal:** Drop `isGlobalOpenCodeDir` — no longer needed.

**Consumer update:** `cmd/compile.go:173` builds `CompileContext` from `storeData.Config.SetupScope` + `storeData.Config.HomeDir` (fallback `os.UserHomeDir()` if HomeDir is empty at non-global scopes).

### Wave B — Gemini commands

**New types:**
```go
// internal/types/types.go
type CommandId string
```

**AdapterSelections extension:**
```go
Commands []types.CommandId
```

**Library assets** — start with 3 commands to prove the pattern:
- `library/commands/rpi.toml` — "Start RPI workflow"
- `library/commands/review.toml` — "Review current work"
- `library/commands/plan.toml` — "Create implementation plan"

TOML schema (verified from Gemini CLI docs):
```toml
name = "rpi"
description = "Start RPI workflow"
prompt = """
Begin the Research → Plan → Implement workflow...
"""
```

**Adapter change** (`gemini.go` `Install`):
```go
// After skills copy:
if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
    Ctx:          ctx,
    SourceSubdir: "commands",
    SelectionKey: "commands",
    ToDestPath: func(file string) string {
        return filepath.Join(geminiDir, "commands", filepath.Base(file))
    },
}); err != nil { return nil, err }
```

**Catalog wiring:** add `Commands` to `AdapterContext.Selections` threading in `internal/scaffold/artifacts.go` and `tui/wizard/phase2.go` (same pattern as Skills). Presets (`internal/preset/`) default all commands to "enabled" for now — revisit in a future UX spec.

### Wave C — Copilot chatmodes

Same pattern as Wave B, different destination:

**Types:** `ChatModeId string` + `ChatModes []types.ChatModeId`

**Library assets** — start with 2 chatmodes:
- `library/chatmodes/architect.chatmode.md`
- `library/chatmodes/reviewer.chatmode.md`

Chatmode file format:
```markdown
---
description: "Architect mode — focused on system design"
tools: ['codebase', 'search']
---
You are an architect...
```

**Adapter change** (`copilot.go` `Install`):
```go
// Skip if scope=global (copilot unsupported there — already early-returned)
if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
    Ctx:          ctx,
    SourceSubdir: "chatmodes",
    SelectionKey: "chatmodes",
    ToDestPath: func(file string) string {
        return filepath.Join(githubDir, "chatmodes", filepath.Base(file))
    },
}); err != nil { return nil, err }
```

### Wave D — Claude Code `--drive-cli`

Mirror Gemini's pattern in `claudecode.go`:

```go
if ctx.DriveCLI {
    if ok := installClaudeMCPViaCLI(ctx, claudeDir); ok {
        log.Println("[claude] MCP servers registered via CLI")
    }
}
```

`installClaudeMCPViaCLI` — reads `.ai/mcp.json`, iterates enabled servers, calls `claude mcp add --name <name> --command <cmd> [args...]`. Falls back silently on binary missing or exec error. Test with stub binary (same pattern as `TestGeminiAdapter_DriveCLI_CallsGeminiBinary`).

### Wave E — CLAUDE.md placeholder fill

**Mechanical replacements** (from `package.json` + `go.mod`):
- `[YOUR_PROJECT_NAME]` → `ai-setup`
- `[YOUR_PROJECT_DESCRIPTION]` → "One-command AI development environment scaffold"
- `[YOUR_TECH_STACK]` → `Go 1.26 · Cobra CLI · SQLite · huh (TUI)`
- `[YOUR_DEV_COMMAND]` → `go run .`
- `[YOUR_TEST_COMMAND]` → `go test ./...`
- `[YOUR_BUILD_COMMAND]` → `go build ./...`

**Leave as TODO with inline comment** (subjective):
- `[YOUR_ORG]`, `[YOUR_TEAM]`, `[YOUR_RULE_*]`, `[YOUR_DO_NOT_*]`, `[YOUR_ARCHITECTURE_NOTES]`, `[YOUR_CODE_STYLE]`, `[YOUR_NAMING_CONVENTIONS]`, `[YOUR_TESTING_STRATEGY]`, `[YOUR_GIT_WORKFLOW]`, `[YOUR_SESSION_CHECK]`

Leave them with an explicit marker: `<!-- fill-in: one sentence -->` so users can grep and replace.

---

## Critical Files

### Modified

| File | Change |
|---|---|
| `internal/adapter/types.go` | Add `CompileContext`; break `CompileMCP` signature; add `Commands`, `ChatModes` to `AdapterSelections` |
| `internal/adapter/mcp_compiler.go` | Rewrite dispatch to accept `CompileContext`; all 5 compile funcs use `ResolveToolRoot`; delete `isGlobalOpenCodeDir` |
| `internal/adapter/claudecode.go` | CompileMCP signature + DriveCLI branch |
| `internal/adapter/opencode.go` | CompileMCP signature |
| `internal/adapter/gemini.go` | CompileMCP signature + copy `commands/` during Install |
| `internal/adapter/copilot.go` | CompileMCP signature + copy `chatmodes/` during Install |
| `internal/adapter/codex.go` | CompileMCP signature |
| `internal/types/types.go` | Add `CommandId`, `ChatModeId` string types |
| `internal/scaffold/artifacts.go` | Thread `Commands`, `ChatModes` into AdapterContext |
| `internal/scaffold/types.go` | Add `Commands`, `ChatModes` to ScaffoldContext |
| `cmd/compile.go` | Build `CompileContext` from storeData |
| `internal/db/store.go` | Serialize `Commands`, `ChatModes` in StoreData (additive) |
| `internal/preset/*.go` | Include `commands` + `chatmodes` in preset expansion |
| `tui/wizard/phase2.go` | (optional) Expose commands/chatmodes selection — default to preset if skipped |
| `CLAUDE.md` | Fill mechanical placeholders, mark subjective ones |

### New

| File | Purpose |
|---|---|
| `library/commands/rpi.toml` | Gemini command template |
| `library/commands/review.toml` | Gemini command template |
| `library/commands/plan.toml` | Gemini command template |
| `library/chatmodes/architect.chatmode.md` | Copilot chatmode template |
| `library/chatmodes/reviewer.chatmode.md` | Copilot chatmode template |
| `internal/adapter/mcp_compiler_scope_test.go` | 15 (tool, scope) pair table test + Copilot skip |
| `internal/adapter/gemini_commands_test.go` | Commands install at all scopes |
| `internal/adapter/copilot_chatmodes_test.go` | Chatmodes install + skip at global |
| `internal/adapter/claudecode_drivecli_test.go` | Stub-binary `claude mcp add` test |

### Not touched

- `ToolAdapter.Install` signature — unchanged
- `types.SetupScope` enum — unchanged
- `configmerge/` — reused, not modified
- `globalpaths/` — reused, not modified

---

## Implementation Waves

### Wave A — scope-aware compile (foundation) → AC-1..4, AC-10

- **A-1:** Add `CompileContext` type in `internal/adapter/types.go`; break `CompileMCP` signature.
- **A-2:** Rewrite `CompileMCPForTool` in `mcp_compiler.go` to accept `CompileContext`; route each adapter compile via `ResolveToolRoot`. Handle `ErrScopeUnsupported` gracefully (Copilot × global → no-op, no error).
- **A-3:** Update all 5 adapter `CompileMCP` wrappers.
- **A-4:** Update `cmd/compile.go` to build `CompileContext` from `storeData`.
- **A-5:** Delete `isGlobalOpenCodeDir`.
- **A-6:** `mcp_compiler_scope_test.go` — table test.
- **A-7:** Fix any existing compile tests that use the old signature.

### Wave B — Gemini commands → AC-5 (partial), AC-6

- **B-1:** Add `CommandId` to `internal/types/types.go`.
- **B-2:** Add `Commands` to `AdapterSelections` in `internal/adapter/types.go`.
- **B-3:** Thread `Commands` through scaffold/store/preset.
- **B-4:** Create `library/commands/{rpi,review,plan}.toml` and embed them.
- **B-5:** Extend `gemini.go` `Install` to copy `commands/*.toml` → `<geminiRoot>/commands/`.
- **B-6:** `gemini_commands_test.go`.

### Wave C — Copilot chatmodes → AC-5 (partial), AC-7

- **C-1..3:** `ChatModeId`, `ChatModes` field, threading (parallel with B).
- **C-4:** Create `library/chatmodes/{architect,reviewer}.chatmode.md` + embed.
- **C-5:** Extend `copilot.go` `Install` to copy chatmodes.
- **C-6:** `copilot_chatmodes_test.go`.

### Wave D — Claude Code `--drive-cli` → AC-8

- **D-1:** Add `installClaudeMCPViaCLI` helper in `claudecode.go`.
- **D-2:** Wire behind `ctx.DriveCLI` check in `Install`.
- **D-3:** `claudecode_drivecli_test.go` — stub binary + missing-binary cases.

### Wave E — CLAUDE.md fill → AC-9

- **E-1:** Mechanical placeholder replacements + `<!-- fill-in -->` markers.

### Dependency graph

```
A ──────┬──▶ D
        ├──▶ B
        └──▶ C
E is independent (docs only)
```

Merge order: A → {B, C, D} parallel → E.

---

## Verification

### Automated

```bash
go vet ./...
go test ./internal/adapter/... ./cmd/... ./internal/scaffold/... -count=1 -v
go build ./...
```

### Manual smoke

1. **Global compile, all tools:**
   ```bash
   cd /tmp/test-project && ai-setup init --scope global --tools claude-code,opencode,gemini,codex --non-interactive
   echo '{"servers":{"filesystem":{"command":"npx","args":["-y","@modelcontextprotocol/server-filesystem"]}}}' > .ai/mcp.json
   ai-setup compile
   ```
   Assert: `~/.config/opencode/opencode.jsonc` has filesystem entry; `~/.codex/config.toml` has `[mcp_servers.filesystem]`; `~/.gemini/settings.json` has mcpServers.filesystem; Copilot skipped with log.
2. **Project compile unchanged:** no drift.
3. **Gemini commands:** `ls .gemini/commands/` shows rpi.toml / review.toml / plan.toml after `ai-setup init`.
4. **Copilot chatmodes:** `ls .github/chatmodes/` shows architect.chatmode.md / reviewer.chatmode.md.
5. **Claude `--drive-cli`:** with stub `claude` binary on PATH, run `ai-setup init --drive-cli --tools claude-code`; assert stub invoked with `mcp add ...`.
6. **CLAUDE.md:** `grep '\[YOUR_' CLAUDE.md` shows only subjective fields remain, each marked `<!-- fill-in -->`.

---

## Risks

| # | Risk | Severity | Mitigation |
|---|---|---|---|
| R-1 | Breaking `CompileMCP` signature breaks internal callers | Low | Type-checker flags all 5 implementations + 1 caller; no external consumers |
| R-2 | Claude project-scope `.mcp.json` behaviour change | Medium | Keep project-scope path identical; only skip at global (comment justifying) |
| R-3 | Commands/chatmodes TOML + frontmatter format drift | Medium | Start with canonical format from Gemini/Copilot docs; library content is regeneratable |
| R-4 | Store schema change for `Commands`/`ChatModes` | Low | Additive fields, JSON marshalling handles missing keys |
| R-5 | Preset expansion may not include new artifact types | Medium | Audit `internal/preset/*.go`; add explicit handling or rely on "all" defaults |
| R-6 | Wizard phase 2 refactor could disrupt spec 007 | Medium | Keep phase2 selection flow intact; default commands/chatmodes to preset when phase is skipped |

---

## Out of Scope (explicit)

- Snapshot tests for library assets (deferred per research decision #3).
- `--drive-cli` for OpenCode, Copilot, Codex (no viable CLI surface).
- Subjective CLAUDE.md fields (rules, architecture, team).
- Gemini `commands/` user-facing wizard UI for selection (defer to spec 010).
- MCP schema evolution (e.g. streaming transports) — current shape is sufficient.
- Project-scope `.mcp.json` format changes — preserve user-committed file shape.

---

⛔ **HUMAN GATE — approve plan before implementing.**
