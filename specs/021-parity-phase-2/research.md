# Spec 021 — Runtime Parity Phase 2: Research

**Date:** 2026-04-24
**Status:** Research — executing during the same session as spec 020/022
**Predecessor:** Spec 020 (configmerge + opencode_validate ported Go→TS). Spec 022 restructured into a monorepo.

## Scope

Close the three remaining parity gaps identified in spec 020's §5 audit:

- **G3 — Housekeeping scaffold** (Go → TS) — port `packages/ai-setup-go/internal/scaffold/housekeeping.go` to TS. 55 LOC + type + wizard hook.
- **T1 — Migration parsers** (TS → Go) — port TS's Claude/Gemini/Copilot config importers to Go. 3 parsers × ~300 LOC each.
- **T2 — Diff3 three-way merge** (TS → Go) — port `packages/ai-setup-ts/src/migration/diff/diff3.ts` to Go. ~200 LOC.

## Ordering

Execute G3 first (smallest, lowest risk, self-contained). Then T2 (diff3 — utility function needed by T1 for merge conflicts). Then T1 (parsers — largest, depends on T2 if we want conflict-merging; otherwise independent).

## G3 — Housekeeping scaffold

### What Go has

`packages/ai-setup-go/internal/scaffold/housekeeping.go` scaffolds `.ai/housekeeping/sync-state.json` with a v1 schema tracking QMD/CodeGraph drift + repair proposals. Called once during `scaffold.Scaffold()` when `ctx.Housekeeping` is non-nil.

Dependencies:
- `types.HousekeepingConfig` struct with `MemoryPath`, `EnableQmd`, `QmdIndexPath`, `EnableCodegraph`, `CodegraphDataPath`
- Consumed nowhere else in Go runtime — purely a scaffold/emit step

### What TS needs

- New type `HousekeepingConfig` in `src/types.ts` (mirror Go shape)
- New function `scaffoldHousekeeping(opts)` in a new file `src/scaffold/housekeeping.ts`
- Wiring into `src/wizard/index.ts` — call when `housekeeping` is set in selections
- Wizard has no housekeeping phase currently; spec 021 does NOT add one. The Go runtime also doesn't expose it interactively — it's set via programmatic `WizardSelections` or non-interactive CLI flags. The TS port mirrors this (accept from selections, skip if absent).

### Q (to resolve during plan)

Does the scaffold alone (write the JSON file) have any real value without a consumer that writes updates to `sync-state.json`? In Go, the file is also just scaffolded — nothing in the Go runtime touches it after init. So the parity is literal: both runtimes scaffold an empty sync-state.json when a housekeeping config is provided; nothing else reads it yet. A future "housekeeping runtime" spec would add the drift-detection consumer to both runtimes.

## T2 — Diff3 three-way merge

### What TS has

`packages/ai-setup-ts/src/migration/diff/diff3.ts` — classic three-way merge (base + ours + theirs → merged with conflict markers). Used by the migration planner when reconciling user-authored config against a new managed baseline.

### What Go needs

- New package `packages/ai-setup-go/internal/diff3/` with a `Merge(base, ours, theirs []string) (MergeResult, error)` function returning merged content + conflict regions.
- Unit tests mirroring TS's cases.

### Complexity

The algorithm is tractable — it's the Myers diff + three-way reconciliation, well-documented. Go's `internal/diff/` already has two-way diff primitives we can build on; diff3 is a direct extension.

## T1 — Migration parsers (TS → Go)

### What TS has

`packages/ai-setup-ts/src/migration/parsers/`:
- `claude-parser.ts` — reads `~/.claude/`, `.claude/settings.json`, frontmatter-laden agents in `.claude/agents/`, `CLAUDE.md`
- `gemini-parser.ts` — reads `.gemini/settings.json`, `GEMINI.md`, custom commands in `.gemini/commands/`
- `copilot-parser.ts` — reads `.github/copilot-instructions.md`, `.github/copilot-chat-modes/`, VS Code `mcp.json`

Each parser:
1. Discovers existing config artifacts for the target tool
2. Extracts agents / rules / commands / MCP entries
3. Normalizes into the canonical `.ai/` format

Called by `ai-setup migrate` and `ai-setup import` commands.

### What Go has

`packages/ai-setup-go/internal/migration/parser.go` — only `parseOpenCodeSetup`. No parsers for Claude/Gemini/Copilot, so Go users can't import existing setups from those tools.

### Strategy for the Go port

One parser per tool, mirroring TS file-by-file:
- `packages/ai-setup-go/internal/migration/parsers/claude.go` + tests
- `packages/ai-setup-go/internal/migration/parsers/gemini.go` + tests
- `packages/ai-setup-go/internal/migration/parsers/copilot.go` + tests

Shared pieces (YAML frontmatter extraction, agent file walking) go into `packages/ai-setup-go/internal/migration/parsers/shared.go`.

The Go `migration.Parser` interface already exists; each new parser implements it. The cobra `migrate` / `import` commands in `packages/ai-setup-go/cmd/` should discover these parsers via the existing registry pattern.

### Scope caveat

T1 is the largest item here — realistically ~900 LOC of Go across 3 parsers plus shared helpers and tests. If context gets tight, G3 + T2 ship first; T1 gets its own follow-up commit within the same spec (spec 021 can have multiple commits).

## Acceptance criteria

- [ ] `packages/ai-setup-ts/src/scaffold/housekeeping.ts` exists with `scaffoldHousekeeping` exported
- [ ] `HousekeepingConfig` type added to TS `src/types.ts`
- [ ] Wizard calls `scaffoldHousekeeping` when a housekeeping config is provided
- [ ] `packages/ai-setup-go/internal/diff3/` exists with `Merge` exported + tests
- [ ] `packages/ai-setup-go/internal/migration/parsers/{claude,gemini,copilot}.go` exist + tests
- [ ] Each Go parser implements the `migration.Parser` interface
- [ ] `pnpm -r run test`, `pnpm -r run typecheck`, `go test ./...` all pass

## Out of scope (for spec 021)

- Housekeeping runtime (actually using `sync-state.json` for drift detection) — separate spec
- Cross-runtime parity harness (CI byte-diff on install outputs) — tracked as spec 023
- Wizard UI for housekeeping selection — deferred
