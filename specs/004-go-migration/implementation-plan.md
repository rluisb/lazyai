# 004 — Go Migration: Implementation Plan (Items 1-4)

## Date: 2026-04-15

## Current State

The Go CLI is **skeleton-complete** with all 14 commands, SQLite store, Charm TUI, and 76 unit tests passing. However, critical functionality is not yet wired end-to-end:

- The `init` wizard collects input but does NOT call `ScaffoldAll()` — it just prints "Scaffolding not yet implemented"
- The `library/` directory is accessed via filesystem paths, not `go:embed` — the binary won't work outside the repo
- Fang CLI styling is not integrated (plain Cobra help is used instead)
- No end-to-end tests exist verifying the full init→compile→doctor flow

This plan covers the 4 items needed to make the Go CLI **production-usable**.

---

## Item 1: Integration Testing (End-to-End Flow Wiring)

### Problem
`cmd/init.go` has two TODO comments where scaffolding should happen:
- Line 101: `// TODO: Implement the actual scaffolding once the scaffold packages are ported.`
- Line 205: Same TODO in non-interactive path

The scaffold packages ARE ported (`internal/scaffold/`), but they're not called from the commands.

### Tasks

#### 1.1 Wire `init` command to scaffold pipeline
**Files:** `cmd/init.go`

After the wizard completes (both interactive and non-interactive paths):
1. Build a `scaffold.ScaffoldContext` from the wizard result
2. Detect the library directory (walk up from binary or current dir to find `library/`)
3. Call `scaffold.ScaffoldAll(ctx)`
4. Write the result to the SQLite store (create or update)
5. Print a styled summary using `tui/components.SummaryBox`

#### 1.2 Wire `update` command
**Files:** `cmd/update.go`

1. Read existing store
2. Re-compute scaffold plan from current selections
3. Detect file changes (hash comparison)
4. Show changes via styled output
5. Re-scaffold modified/new files
6. Update store

#### 1.3 Wire `add` command
**Files:** `cmd/add.go`

1. Read existing store
2. Update tools/agents/skills selections
3. Re-scaffold only the new artifacts
4. Update store

#### 1.4 Wire `create` command to generators
**Files:** `cmd/create.go`

1. Use `generator.Registry` to create new artifacts
2. Write the generated file to disk
3. Track in store

#### 1.5 End-to-end integration tests
**Files:** `cmd/integration_test.go` (new)

Test the full flow:
```go
func TestInitNonInteractive(t *testing.T) {
    // Create temp dir, run init, verify files created, verify store written
}
func TestInitAndDoctor(t *testing.T) {
    // Init, then doctor should show all healthy
}
func TestInitAndCompile(t *testing.T) {
    // Init, then compile should generate per-tool configs
}
```

#### 1.6 Test against real project
Run these manually and verify output:
```bash
./ai-setup init --non-interactive --scope project --tools opencode,claude-code --preset standard --name my-app
./ai-setup status
./ai-setup doctor
./ai-setup compile
./ai-setup server list
```

---

## Item 2: go:embed for Library Data

### Problem
The scaffold and generator packages currently access `library/` via relative filesystem paths. When the binary is installed elsewhere (e.g., `/usr/local/bin/ai-setup`), it won't find `library/`.

### Tasks

#### 2.1 Create `internal/library/embed.go`
Embed the entire `library/` directory:
```go
package library

import "embed"

//go:embed all:agents all:skills all:constitution all:fragments all:infra all:mcp all:orchestration all:prompts all:root all:rules all:specs-agents all:templates all:tool-agents all:tool-templates
var FS embed.FS
```

#### 2.2 Add library directory access functions
```go
func Root() string             // root path in embed.FS
func AgentsDir() string       // library/agents/
func SkillsDir() string        // library/skills/
func TemplatesDir() string     // library/templates/
// etc.
func FindLibraryDir() (string, error)  // fallback: walk up from CWD to find library/
```

#### 2.3 Update all packages to use embed.FS
Currently these packages read from filesystem:
- `internal/scaffold/*` — all scaffold modules read from `libraryDir`
- `internal/adapter/*` — adapters read from `libraryDir`
- `internal/compiler/*` — template compiler reads from `libraryDir`
- `internal/generator/*` — generators read from `libraryDir`
- `internal/orchestrator/catalog.go` — reads from `library/orchestration/`
- `cmd/server.go` — reads `library/mcp/catalog.json`

**Migration strategy:** Each function that currently takes `libraryDir string` should also accept `embed.FS` or `fs.FS`. Add a parameter that defaults to the embedded FS when empty.

#### 2.4 Update `libraryFindDir` fallback
When running from source (development), prefer the filesystem `library/` directory for live editing. When running from installed binary, use `embed.FS`.

Pattern:
```go
func GetLibraryFS() fs.FS {
    // If library/ exists on filesystem (dev mode), use it
    if dir, err := FindLibraryDir(); err == nil {
        return os.DirFS(dir)
    }
    // Otherwise use embedded FS (production)
    return library.FS
}
```

#### 2.5 Test embed in production mode
```bash
go build -o /tmp/ai-setup-test .
cd /tmp/empty-dir
/tmp/ai-setup-test init --non-interactive --scope project --tools opencode --preset standard --name test
# Should work without access to repository library/ directory
```

---

## Item 3: Fang Integration (CLI Styling)

### Problem
Root command uses plain Cobra help output. Fang provides styled help pages, error formatting, and manpage generation.

### Tasks

#### 3.1 Install Fang dependency
```bash
go get github.com/charmbracelet/fang/v2
```

Note: Check exact import path — Fang may use `charm.land/fang/v2` or `github.com/charmbracelet/fang/v2`.

#### 3.2 Update `cmd/root.go`
Replace `rootCmd.Execute()` with `fang.Execute(ctx, rootCmd)`:
```go
func Execute(ctx context.Context) error {
    return fang.Execute(ctx, rootCmd)
}
```

#### 3.3 Configure Fang styling
Customize Fang colors and formatting to match the ai-setup brand (purple primary, teal secondary).

#### 3.4 Test styled help
```bash
./ai-setup --help          # Should show Fang-styled help
./ai-setup init --help     # Should show Fang-styled help for init
./ai-setup nonexistent     # Should show Fang-styled error
```

---

## Item 4: End-to-End Verification Against Real Project

### Problem
The Go CLI has not been tested against a real project with real `.ai/` configs. We need to verify that Go-written configs match TypeScript-written configs byte-for-byte.

### Tasks

#### 4.1 Create comparison test suite
**Files:** `cmd/comparison_test.go` (new)

For each tool adapter:
1. Run TypeScript `ai-setup compile --tool opencode`
2. Run Go `./ai-setup compile --tool opencode`
3. Compare the output files
4. Assert they are semantically equivalent (JSON field order may differ)

#### 4.2 Test all 6 adapters
| Adapter | Config file | Verified |
|---------|------------|-----------|
| OpenCode | `.opencode/opencode.jsonc` | ☐ |
| Claude Code | `.mcp.json` | ☐ |
| Gemini | `.gemini/settings.json` | ☐ |
| Copilot | `.vscode/mcp.json` | ☐ |
| Codex | `.agents/AGENTS.md` | ☐ |
| Pi | (minimal) | ☐ |

#### 4.3 Test MCP server add/remove flow
```bash
./ai-setup server add memory
./ai-setup server add ripgrep
./ai-setup server list
./ai-setup server remove memory
./ai-setup compile
# Verify .ai/mcp.json and per-tool configs updated correctly
```

#### 4.4 Test JSON bridge migration
```bash
# Create a .ai-setup.json from TypeScript
npx ai-setup init --scope project --tools opencode
# Run Go binary on same directory
./ai-setup doctor
# Should detect .ai-setup.json, auto-import to SQLite, and show healthy
```

#### 4.5 Test all command flags
Verify every flag works:
```bash
./ai-setup init --scope project --tools opencode,claude-code --preset full --name test --non-interactive
./ai-setup add --tools gemini
./ai-setup update --dry-run
./ai-setup doctor --fix
./ai-setup status --json
./ai-setup list --type agents --verbose
./ai-setup info builder
./ai-setup create agent test-agent --force
./ai-setup orchestration list
./ai-setup compile --tool opencode
./ai-setup migrate --from v0 --non-interactive
./ai-setup import /path/to/existing/setup --non-interactive
```

---

## Execution Order

| Order | Item | Dependencies | Estimated Effort |
|-------|------|-------------|-----------------|
| 1 | 1.1-1.4: Wire commands to scaffold | None | Medium |
| 2 | 1.5-1.6: Integration tests | Item 1.1-1.4 | Medium |
| 3 | 2.1-2.5: go:embed library | None | Medium |
| 4 | 3.1-3.4: Fang integration | None | Small |
| 5 | 4.1-4.2: Comparison tests | Items 1-3 | Medium |
| 6 | 4.3-4.5: Full E2E verification | Items 1-4 | Medium |

Items 1, 2, and 3 can be worked on in parallel. Item 4 depends on all three.