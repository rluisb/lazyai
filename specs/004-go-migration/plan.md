# 004 — TypeScript → Go CLI Migration: Plan

## Date
2026-04-15

## Context

ai-setup CLI is currently TypeScript (Commander + @clack/prompts + lowdb). We are migrating the CLI to Go for:
- Single binary deployment (no Node.js runtime)
- Charm Bracelet TUI ecosystem (bubbletea, lipgloss, huh, fang, bubbles)
- SQLite persistence with schema migrations
- Interactive conflict resolution (side-by-side diffs)
- Better UX with polished terminal UI

**Orchestrator MCP stays in TypeScript** — it's consumed as a subprocess via npx, and the TS SDK is the reference implementation.

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| MCP language | **TypeScript (keep)** | Most mature SDK, first to get new features, subprocess model already works |
| Persistence | **SQLite (modernc.org/sqlite)** | ACID, SQL queries, schema migrations, pure Go (no CGO) |
| CLI framework | **Cobra + Fang** | Cobra for commands, Fang for styled help/errors |
| TUI framework | **Bubble Tea v2** | Core interactive view engine |
| Forms/prompts | **Huh v2** | Wizard flows, accessible mode, dynamic forms |
| Styling/layout | **Lip Gloss v2** | Diff views, tables, trees, borders |
| Components | **Bubbles v2** | Viewport, list, table, spinner, progress |
| Schema migrations | **golang-migrate/migrate** | Industry-standard, supports SQLite, versioned SQL files |
| Branch | **feature/go-migration** | From main |
| Library data | Keep shipping YAML/JSON/MD files alongside binary | Embedded via `embed.FS` |

## Architecture Overview

```
ai-setup (Go binary)
├── cmd/                    # Cobra commands
│   ├── root.go             # Root command (fang.Execute)
│   ├── init.go             # ai-setup init
│   ├── add.go              # ai-setup add
│   ├── update.go           # ai-setup update
│   ├── doctor.go           # ai-setup doctor
│   ├── status.go           # ai-setup status
│   ├── create.go           # ai-setup create
│   ├── eject.go            # ai-setup eject
│   ├── compile.go          # ai-setup compile
│   ├── list.go             # ai-setup list
│   ├── info.go             # ai-setup info
│   ├── server.go           # ai-setup server (add/remove/list/doctor)
│   ├── orchestration.go    # ai-setup orchestration
│   ├── migrate.go          # ai-setup migrate
│   ├── import.go           # ai-setup import
│   └── completions.go     # Shell completion generation
│
├── internal/               # Internal packages
│   ├── adapter/            # Tool adapters (opencode, claude, gemini, copilot, codex, pi)
│   ├── compiler/           # Template compiler, fragment resolver
│   ├── conflict/           # Conflict resolution + interactive diff viewer
│   ├── db/                 # SQLite store + schema migrations
│   ├── diff/               # Diff computation
│   ├── error/              # Structured error system
│   ├── frontmatter/        # YAML frontmatter parsing
│   ├── generator/          # Artifact generators
│   ├── migration/          # Import/migrate from other AI setups
│   ├── orchestrator/       # MCP server management (add/remove/list/doctor)
│   ├── preset/             # Feature preset definitions
│   ├── scaffold/            # File scaffolding modules
│   ├── types/              # Core type definitions
│   └── wizard/             # Interactive setup wizard
│
├── tui/                    # TUI components (separate from internal for testability)
│   ├── diffviewer/         # Side-by-side conflict diff (bubbletea + lipgloss)
│   ├── wizard/             # Wizard phases (huh forms + bubbletea views)
│   ├── status/             # Status display components
│   ├── theme/              # Lip Gloss theme definitions
│   └── components/         # Reusable TUI components
│
├── library/                # Embedded library data (go:embed)
│   └── ...                 # Same structure as current library/
│
├── migrations/             # SQL schema migrations
│   ├── 001_initial.sql
│   ├── 002_*.sql
│   └── ...
│
├── go.mod
├── go.sum
├── main.go                 # Entry point
└── Makefile                # Build, test, install targets
```

## Migration Waves

### Wave 1: Go Project Scaffolding + Core Types (Foundation)

**Goal**: Bootable Go CLI with Cobra + Fang, core types, and SQLite store.

#### 1.1 — Initialize Go module + Cobra structure

- `go mod init github.com/ricardoborges-teachable/ai-setup`
- Install dependencies: cobra, fang, bubbletea v2, lipgloss v2, huh v2, bubbles v2
- Create `main.go` with `fang.Execute(ctx, rootCmd)`
- Create all 14 command stubs in `cmd/`
- Wire Cobra command structure (flags, descriptions, argument validation)

**Files created:**
- `main.go`
- `cmd/root.go`, `cmd/init.go`, `cmd/add.go`, `cmd/update.go`, `cmd/doctor.go`, `cmd/status.go`, `cmd/create.go`, `cmd/eject.go`, `cmd/compile.go`, `cmd/list.go`, `cmd/info.go`, `cmd/server.go`, `cmd/orchestration.go`, `cmd/migrate.go`, `cmd/import.go`, `cmd/completions.go`

**Verification:** `go build` produces binary, `./ai-setup --help` shows styled Fang output, all subcommands appear.

---

#### 1.2 — Port core types

- Port `src/types.ts` → `internal/types/types.go`
- Port `src/store/schema.ts` → Go structs
- Port `src/presets.ts` → `internal/preset/presets.go`
- Define all ID enums as Go string enums or typed strings
- Define `StoreData`, `TrackedFile`, `Operation`, `Config`, `Selections` structs

**Files created:**
- `internal/types/types.go`
- `internal/preset/presets.go`

**Verification:** `go build ./...`, `go vet ./...`

---

#### 1.3 — SQLite store with schema migrations

- Integrate `modernc.org/sqlite` (pure Go, no CGO)
- Integrate `golang-migrate/migrate` v4
- Design schema from current `.ai-setup.json` structure:

```sql
-- migrations/001_initial_schema.sql
CREATE TABLE meta (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  schema_version INTEGER NOT NULL DEFAULT 1,
  cli_version TEXT NOT NULL,
  installed_at TEXT NOT NULL,
  last_updated_at TEXT NOT NULL
);

CREATE TABLE config (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  scope TEXT NOT NULL,           -- global/project/workspace
  tools TEXT NOT NULL,           -- JSON array
  cli_tools TEXT NOT NULL DEFAULT '[]',  -- JSON array
  enable_servers TEXT NOT NULL DEFAULT '[]',
  project_name TEXT NOT NULL DEFAULT '',
  target_dir TEXT NOT NULL DEFAULT '',
  planning_dir TEXT NOT NULL DEFAULT 'specs',
  repos TEXT NOT NULL DEFAULT '[]'
);

CREATE TABLE selections (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  templates TEXT NOT NULL DEFAULT '[]',
  rules TEXT NOT NULL DEFAULT '[]',
  agents TEXT NOT NULL DEFAULT '[]',
  skills TEXT NOT NULL DEFAULT '[]',
  prompts TEXT NOT NULL DEFAULT '[]',
  infra TEXT NOT NULL DEFAULT '[]',
  constitution TEXT NOT NULL DEFAULT '[]',
  features TEXT NOT NULL DEFAULT '{}',  -- JSON object
  git_conventions TEXT NOT NULL DEFAULT '{}',
  preset TEXT NOT NULL DEFAULT 'standard'
);

CREATE TABLE tracked_files (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  path TEXT NOT NULL UNIQUE,
  hash TEXT NOT NULL,
  source TEXT NOT NULL,
  owner TEXT NOT NULL DEFAULT 'library',  -- library/user/migrated
  status TEXT NOT NULL DEFAULT 'installed',  -- installed/modified/missing/conflict
  installed_at TEXT NOT NULL,
  last_checked_at TEXT NOT NULL,
  FOREIGN KEY (config_id) REFERENCES config(id)
);

CREATE TABLE operations (
  id TEXT PRIMARY KEY,  -- op_{timestamp}_{random}
  type TEXT NOT NULL,   -- init/update/add/eject/create/migrate/import
  timestamp TEXT NOT NULL,
  files_affected TEXT NOT NULL DEFAULT '[]',  -- JSON array
  result TEXT NOT NULL,  -- success/partial/failed
  error TEXT,
  backup_paths TEXT NOT NULL DEFAULT '{}'  -- JSON object
);

CREATE TABLE feature_flags (
  key TEXT PRIMARY KEY,
  value INTEGER NOT NULL DEFAULT 0
);

-- Index for file lookups
CREATE INDEX idx_tracked_files_path ON tracked_files(path);
CREATE INDEX idx_operations_timestamp ON operations(timestamp);
```

- Create `internal/db/store.go` — SQLite connection, store CRUD
- Create `internal/db/migrations.go` — migration runner
- Import existing `.ai-setup.json` data → SQLite on first run (migration bridge)
- Keep `.ai-setup.json` as fallback: if SQLite DB doesn't exist but JSON does, import it

**Files created:**
- `internal/db/db.go` — connection factory
- `internal/db/store.go` — CRUD operations
- `internal/db/migrations.go` — migration runner
- `migrations/001_initial_schema.sql`
- `migrations/002_add_preset_field.sql` (if needed)

**Verification:** Unit tests for store CRUD, migration from JSON → SQLite round-trip.

> **MILESTONE M1:** Go binary boots, connects to SQLite, can read/write store data.

---

### Wave 2: Core Commands (Non-Interactive)

**Goal**: Port all non-interactive commands that don't need TUI.

#### 2.1 — File utilities

- Port `src/utils/files.ts` → `internal/files/files.go`
- Port `src/utils/frontmatter.ts` → `internal/frontmatter/frontmatter.go`
- Port `src/utils/jsonc.ts` → `internal/jsonc/jsonc.go`
- Port `src/utils/validation.ts` → `internal/validation/validation.go`
- Port `src/utils/diff.ts` → `internal/diff/diff.go`
- Port `src/utils/manifest.ts` → `internal/manifest/manifest.go` (read .ai-setup.json bridge)

#### 2.2 — Error system

- Port `src/errors/types.ts` → `internal/error/errors.go`
- Port `src/errors/boundary.ts` → `internal/error/boundary.go`
- Port `src/errors/operation.ts` → `internal/error/operation.go`
- Use Fang's styled error output

#### 2.3 — Non-interactive commands

Port in order (simplest first):
1. `status` — reads store, displays state
2. `list` — lists installed artifacts
3. `info` — detailed artifact info
4. `doctor` — validates setup health
5. `eject` — removes library management
6. `compile` — manual MCP compilation
7. `server` (add/remove/list/doctor) — MCP server management

**Verification:** Each command runs via `./ai-setup <cmd>` and produces correct output. Doctor validates both JSON and SQLite stores.

> **MILESTONE M2:** All non-interactive commands work. `ai-setup doctor`, `ai-setup status`, `ai-setup list`, `ai-setup server list` all functional.

---

### Wave 3: Adapter + Compiler System

**Goal**: Port the multi-tool adapter layer and MCP compiler.

#### 3.1 — Adapter registry + shared helpers

- Port `src/adapters/registry.ts` → `internal/adapter/registry.go`
- Port `src/adapters/types.ts` → `internal/adapter/types.go`
- Port `src/adapters/shared.ts` → `internal/adapter/shared.go`
- Port `src/adapters/mcp-compiler.ts` → `internal/adapter/mcp_compiler.go`

#### 3.2 — Individual adapters

Port all 6 adapters:
1. `opencode.ts` → `internal/adapter/opencode.go`
2. `claude-code.ts` → `internal/adapter/claudecode.go`
3. `gemini.ts` → `internal/adapter/gemini.go`
4. `copilot.ts` → `internal/adapter/copilot.go`
5. `codex.ts` → `internal/adapter/codex.go`
6. `pi.ts` → `internal/adapter/pi.go`

#### 3.3 — Template compiler

- Port `src/compiler/index.ts` → `internal/compiler/compiler.go`
- Port `src/compiler/fragment-resolver.ts` → `internal/compiler/fragment.go`
- Port `src/compiler/template-compiler.ts` → `internal/compiler/template.go`

#### 3.4 — MCP compilation (compile command)

Port `src/commands/compile.ts` logic. This depends on 3.1-3.3.

**Verification:** `ai-setup compile` generates per-tool configs from `.ai/mcp.json`. Test with all 6 tool adapters.

> **MILESTONE M3:** Adapter system works. `ai-setup compile` generates correct per-tool configs. `ai-setup server add orchestrator` works.

---

### Wave 4: Scaffold + Generator System

**Goal**: Port all scaffolding and artifact generation.

#### 4.1 — Scaffold modules

Port all 13 scaffold functions:
1. `agents-skills-prompts.ts` → `internal/scaffold/artifacts.go`
2. `compiled-root.ts` → `internal/scaffold/root.go`
3. `constitution.ts` → `internal/scaffold/constitution.go`
4. `env-example.ts` → `internal/scaffold/env.go`
5. `gitignore.ts` → `internal/scaffold/gitignore.go`
6. `infra.ts` → `internal/scaffold/infra.go`
7. `mcp.ts` → `internal/scaffold/mcp.go`
8. `orchestration.ts` → `internal/scaffold/orchestration.go`
9. `repo-roots.ts` → `internal/scaffold/repos.go`
10. `root-file-map.ts` → `internal/scaffold/filemap.go`
11. `root-files.ts` → `internal/scaffold/rootfiles.go`
12. `specs.ts` → `internal/scaffold/specs.go`
13. `templates-rules.ts` → `internal/scaffold/templates.go`

#### 4.2 — Generator registry

- Port `src/generators/registry.ts` → `internal/generator/registry.go`
- Port all 8 generator types (agent, skill, workflow, domain, mode, prompt, command, template)

#### 4.3 — Embedded library

- Use `go:embed` to embed the `library/` directory
- Port library access from filesystem reads → embedded FS reads
- Keep `library/` directory for development; embed for release builds

**Verification:** `ai-setup create agent test-agent` generates a valid agent file. Scaffold functions produce correct file trees.

> **MILESTONE M4:** Scaffold and generator systems work. `ai-setup create` commands functional.

---

### Wave 5: Interactive Commands — TUI (The Big One)

**Goal**: Port all interactive commands using Charm TUI stack. This is where the UX upgrade happens.

#### 5.1 — TUI theme + shared components

- Create `tui/theme/theme.go` — Lip Gloss style definitions (colors, borders, spacing)
- Create `tui/components/` — Reusable components:
  - `spinner.go` — Bubbles spinner with lipgloss styling
  - `progress.go` — Bubbles progress bar
  - `summary.go` — Post-operation summary box
  - `table.go` — Styled table display
  - `tree.go` — Styled tree display

#### 5.2 — Wizard system (Huh + Bubble Tea)

Port the 4-phase wizard with TUI upgrade:

**Phase 1: Context** (`tui/wizard/phase1.go`)
- Replace @clack/prompts → `huh.NewForm` with `huh.NewSelect` (scope), `huh.NewMultiSelect` (tools), `huh.NewInput` (project name)
- Auto-detect: scan for git repo, package.json, existing `.ai-setup.json`
- Dynamic: if project scope detected, pre-fill and ask confirmation

**Phase 2: Features** (`tui/wizard/phase2.go`)
- Preset select: `huh.NewSelect[Presets]` → Minimal / Standard / Full / Custom
- Custom path: `huh.NewMultiSelect[Feature]` with toggle
- Scope-aware defaults

**Phase 3: Conflicts** (`tui/wizard/phase3.go`) — **KEY UX UPGRADE**
- **Side-by-side diff viewer** built with Bubble Tea:
  - Two lipgloss-styled panes showing "ours" vs "theirs"
  - Color-coded: green for additions, red for deletions, yellow for conflicts
  - Keyboard nav: ↑/↓ to scroll, ←/→ to switch pane
  - Per-conflict action bar: `[a] Accept  [d] Deny  [s] Skip  [?] Help`
  - Bubbles viewport for scrollable content
  - Lip Gloss `JoinHorizontal` for side-by-side layout
- **Huh confirmation** after reviewing all conflicts
- **Back navigation** across wizard phases (preserve bubbletea model)

**Phase 4: Confirm** (`tui/wizard/phase4.go`)
- Lip Gloss styled summary table of what will be installed
- `huh.NewConfirm` to proceed
- Back navigation to any previous phase

**Wizard orchestrator** (`tui/wizard/wizard.go`)
- Bubble Tea main model managing phase transitions
- State machine: phase1 → phase2 → phase3 (if conflicts) → phase4 → execute
- Back navigation stack

#### 5.3 — Interactive commands port

1. `init` — calls wizard flow
2. `add` — huh form for artifact selection
3. `update` — conflict resolution TUI if files modified
4. `create` — huh form for artifact type + name
5. `migrate` / `import` — wizard-style importer

#### 5.4 — Conflict resolution TUI (standalone)

Create `tui/diffviewer/viewer.go` — reusable side-by-side diff component:
- Input: list of conflicts with "ours" and "theirs" content
- Bubble Tea model with:
  - Two viewports (bubbles viewport) showing side-by-side content
  - Lip Gloss styled borders, headers (File Path, Status)
  - Action bar: Accept (keep ours), Deny (keep theirs), Skip, Help
  - Key bindings displayed via bubbles help component
  - Summary bar at bottom: X accepted, Y denied, Z skipped, remaining
- Returns: list of resolution decisions per conflict

**Verification:**
- `ai-setup init` runs full 4-phase wizard with back navigation
- Conflict resolution shows side-by-side diffs with keyboard controls
- All interactive commands work with both TTY and `--non-interactive` flag fallback
- All non-interactive flags still work (`--scope project --tools opencode,claude --preset standard`)

> **MILESTONE M5:** All interactive commands work with Charm TUI. The wizard is polished and conflict resolution is interactive.

---

### Wave 6: Migration System + Orchestration Integration

**Goal**: Port the import/migration system and ensure orchestrator MCP integration works.

#### 6.1 — Migration system

Port `src/migration/` → `internal/migration/`:
1. `detector.go` — detect existing AI setups
2. `executor.go` — run migration
3. `plan.go` — compute migration plan
4. `parsers/` — per-tool migration parsers
5. `canonical-writer.go` — write parsed data into `.ai/` format
6. `diff/` — diff computation for migration preview

#### 6.2 — Orchestration commands

Port `src/commands/orchestration.ts` → `cmd/orchestration.go`:
- List catalogs (chains, teams, workflows, domains, modes)
- Create orchestration artifacts
- Show status

This does NOT rewrite the MCP server — it just ensures the CLI can manage the `.ai/orchestration/` directory structure and the MCP server (still TypeScript) can read from it.

#### 6.3 — Orchestrator MCP integration verification

Ensure the Go CLI correctly:
1. Adds orchestrator to `.ai/mcp.json` (`ai-setup server add orchestrator`)
2. Compiles per-tool configs that reference `npx -y @ai-setup/orchestrator`
3. The TypeScript MCP server can read `.ai/orchestration/` files written by the Go CLI
4. All three scopes (Global, Project, Workspace) produce correct MCP config paths

**Verification:**
- End-to-end: `ai-setup init → ai-setup server add orchestrator → ai-setup compile`
- Verify `.mcp.json` and `opencode.jsonc` contain correct orchestrator config
- Start an AI tool with the orchestrator MCP enabled — it connects successfully

> **MILESTONE M6:** Migration system works. Orchestrator MCP integration verified end-to-end.

---

### Wave 7: Polish + Testing + Documentation

**Goal**: Production-ready Go CLI that fully replaces the TypeScript version.

#### 7.1 — Comprehensive test suite

- Unit tests for all `internal/` packages
- Integration tests for all commands
- TUI component tests (bubbletea test harness)
- SQLite store tests (in-memory SQLite for speed)
- Migration bridge tests (JSON → SQLite)
- Adapter output comparison tests (Go vs TypeScript output)

#### 7.2 — Documentation + VHS demos

- README update with Go installation instructions
- VHS recordings of key flows:
  - `ai-setup init` wizard
  - Conflict resolution
  - `ai-setup doctor`
  - `ai-setup server add orchestrator`
- Man page generation (via Fang + mango)

#### 7.3 — Build + Distribution

- Makefile targets: `build`, `test`, `lint`, `install`, `release`
- Cross-compilation: `GOOS=linux GOARCH=amd64`, `GOOS=darwin GOARCH=arm64`, `GOOS=windows GOARCH=amd64`
- goreleaser config for GitHub Releases
- Homebrew tap (optional)

#### 7.4 — GitHub Actions CI

- Go test + vet + build on PR
- goreleaser on tag
- Keep existing TypeScript CI running in parallel during migration

#### 7.5 — Remove TypeScript CLI

After the Go binary is production-ready:
- Remove `src/`, `bin/`, `tsconfig.json`, `tsup.config.ts`, `biome.json`
- Update `package.json` to only contain the orchestrator subpackage
- Keep `orchestrator/` directory unchanged
- Update `library/` references if needed

> **MILESTONE M7:** Go CLI is production-ready. TypeScript CLI code removed. Only orchestrator remains as TypeScript.

---

## Dependency Graph

```
Wave 1 (Types + SQLite)
  └─► Wave 2 (Non-interactive commands)
        └─► Wave 3 (Adapters + Compiler)
              └─► Wave 4 (Scaffold + Generators)
                    └─► Wave 5 (TUI + Interactive commands)
                          └─► Wave 6 (Migration + Orchestration)
                                └─► Wave 7 (Polish + Testing + Cleanup)
```

All waves are **sequential** — each builds on the previous.

---

## Risks

| Risk | Likelihood | Impact | Mitigation |
|------|:----------:|:------:|-----------|
| SQLite schema migrations break existing installs | Medium | High | JSON → SQLite bridge import on first run; keep `.ai-setup.json` readable; test with real existing setups |
| Charm TUI libraries have breaking API changes (v2 is new) | Low | Medium | Pin versions in go.mod; follow Charm upgrade guides; v2 is stable per repo READMEs |
| Side-by-side diff viewer is complex to build | High | Medium | Build incrementally — start with simple accept/reject, upgrade to side-by-side in steps; wizard-tutorial code as reference |
| Orchestrator MCP incompatible with Go-written config files | Low | High | Orchestrator reads JSON/MD files from `.ai/orchestration/` — Go CLI writes same format; verify with integration test |
| Feature parity gaps — Go CLI misses TypeScript features | Medium | High | Track all 14 commands + flags in a checklist; cross-test every flag; keep TypeScript CLI available until parity confirmed |
| `go:embed` increases binary size significantly | Low | Low | Library is ~500KB of text files; binary stays under 20MB; acceptable |
| Windows terminal compatibility with Charm TUI | Medium | Low | Ultraviolet handles Windows; test on Windows Terminal + PowerShell; CI on Windows if needed |

---

## Acceptance Criteria

| AC | Description | Verification |
|----|-------------|-------------|
| AC-1 | All 14 CLI commands work with identical flags and behavior | Run each command with same flags as TS version, compare output |
| AC-2 | Wizard flows are interactive with Charm TUI | `ai-setup init` shows huh-based wizard with back navigation |
| AC-3 | Conflict resolution shows side-by-side diffs | Create a conflict scenario, verify diff viewer renders correctly |
| AC-4 | SQLite store persists data with schema migrations | Create v1 store, run migration, verify v2 schema works; JSON bridge import works |
| AC-5 | Orchestrator MCP works with Go CLI | `ai-setup server add orchestrator → compile` → AI tool connects to MCP |
| AC-6 | All scopes work (Global, Project, Workspace) | Run init with each scope, verify correct file placement |
| AC-7 | All 6 tool adapters generate correct configs | Compare Go output vs TypeScript output for each adapter |
| AC-8 | `--non-interactive` flag works for all commands | Run `ai-setup init --non-interactive --scope project --tools opencode --preset standard` |
| AC-9 | Single binary deploys with no runtime dependencies | `go build` → binary; run on clean machine without Node.js |
| AC-10 | TypeScript CLI code removed, only orchestrator remains | `src/` directory deleted; `package.json` only has orchestrator |

---

## Go Dependencies

```
require (
    // CLI
    github.com/spf13/cobra v1.8+
    github.com/charmbracelet/fang v2.0+

    // TUI
    charm.land/bubbletea/v2     // Core TUI framework
    charm.land/lipgloss/v2      // Styling/layout
    charm.land/huh/v2           // Forms/wizard

    // TUI Components
    github.com/charmbracelet/bubbles/v2  // Viewport, list, table, spinner

    // Persistence
    modernc.org/sqlite          // Pure Go SQLite
    github.com/golang-migrate/migrate/v4  // Schema migrations

    // Utilities
    github.com/yaml/go-yaml     // YAML parsing
    github.com/twpwayne/go-shell // Shell completions (or use cobra built-in)
)
```

---

## File Touch Map

### New Go files (~40+)
All files in `cmd/`, `internal/`, `tui/`, `migrations/` — see architecture tree above.

### Deleted TypeScript files (Wave 7)
- Entire `src/` directory (31 files + 31 test files)
- `bin/ai-setup.js`
- `tsconfig.json`, `tsup.config.ts`, `biome.json`
- Most `package.json` deps (keep only orchestrator-related)

### Preserved unchanged
- `orchestrator/` — entire directory stays TypeScript
- `library/` — same content, accessed via go:embed
- `.ai/`, `.opencode/`, `.claude/`, specs, etc. — runtime/output directories
- `AGENTS.md`, `CLAUDE.md` — project-level config
- GitHub Actions (extend, don't replace)

### Modified
- `package.json` — simplified to orchestrator-only
- `.github/workflows/ci.yml` — add Go CI
- `.github/workflows/publish.yml` — add goreleaser
- `README.md` — update installation instructions

---

## Estimated Complexity

| Wave | Tasks | Complexity | Key Risk |
|------|-------|------------|----------|
| Wave 1 | 3 | M | SQLite schema design |
| Wave 2 | 3 | M | Error system parity |
| Wave 3 | 4 | M | Adapter output must match TS exactly |
| Wave 4 | 3 | M | Embedded FS access pattern |
| Wave 5 | 4 | **H** | Side-by-side diff viewer, wizard state machine |
| Wave 6 | 3 | M | Orchestrator integration testing |
| Wave 7 | 5 | M | Comprehensive testing, distribution setup |

**Total estimated effort**: 7 waves, ~25 tasks, with Wave 5 being the highest-risk highest-reward area.

---

## Next Step

Create feature branch `feature/go-migration` from `main` and begin Wave 1.