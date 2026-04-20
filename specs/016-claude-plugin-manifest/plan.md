# Plan — 016: `ai-setup build-plugin` Command

**Date:** 2026-04-20
**Phase:** Plan (P of RPI) — awaiting HUMAN GATE before Implement
**Research:** `research.md` (decisions locked in §5)

---

## 1. Objective

Ship a `build-plugin` subcommand that generates a Claude Code plugin directory from ai-setup's embedded `library/` so users can install the Claude-facing content (agents + skills + commands + output styles) via the official plugin system (`claude --plugin-dir <path>` or a future marketplace entry).

Exit: running `ai-setup build-plugin` in any directory produces a `dist/plugin/` tree that passes `claude plugin validate`, contains every source agent/skill/command/output-style, and installs cleanly with `claude --plugin-dir ./dist/plugin`.

---

## 2. Locked decisions (from research §5)

| # | Answer |
|---|---|
| Q1 | **Option B** — `ai-setup build-plugin` generator command |
| Q2 | **b** — agents + skills + commands + output styles (no MCP in MVP) |
| Q3 | Default `./dist/plugin/`; override via `--out <path>` |
| Q4 | Hardcoded plugin metadata (name, author, homepage, license) |
| Q5 | Defer `marketplace.json` generation to a future spec |
| Q6 | Plugin name: `"ai-setup"` |
| Q7 | Plugin version **synced with `cmd.Version`** (ldflag-driven binary version) |

---

## 3. Target plugin layout

```
dist/plugin/
├── .claude-plugin/
│   └── plugin.json              # Generated manifest (metadata only; no custom paths)
├── agents/
│   ├── builder.md               # Verbatim copies from library/agents/
│   ├── planner.md
│   ├── scout.md
│   └── …
├── skills/
│   ├── implement/
│   │   └── SKILL.md             # Restructured from library/skills/implement.md
│   ├── plan/
│   │   └── SKILL.md
│   └── …
├── commands/
│   ├── commit.md                # Verbatim copies from library/claudecode/commands/
│   ├── review.md
│   └── test.md
└── output-styles/
    ├── explanatory.md           # Verbatim copies from library/claudecode/output-styles/
    └── terse.md
```

### `plugin.json` payload

```json
{
  "name": "ai-setup",
  "version": "<cmd.Version>",
  "description": "ai-setup agents, skills, commands, and output styles for Claude Code",
  "author": {
    "name": "Ricardo Borges",
    "url": "https://github.com/ricardoborges-teachable/ai-setup"
  },
  "homepage": "https://github.com/ricardoborges-teachable/ai-setup",
  "repository": "https://github.com/ricardoborges-teachable/ai-setup",
  "license": "MIT",
  "keywords": ["ai-setup", "claude-code", "agents", "skills"]
}
```

No custom `skills`/`agents`/`commands` paths — we emit at the default locations so Claude Code auto-discovers them.

---

## 4. Phased breakdown

### Phase 1 — Generator core (`internal/plugin/`)

**Goal:** Pure-Go library that reads `library/*` via `fs.FS` and writes a plugin tree.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 001 | New package `internal/plugin/`: `Build(libFS fs.FS, outDir, version string) error`. Orchestrates the copy/restructure steps below, returns the count of emitted files and any error. | ~120 |
| 002 | `buildManifest(outDir, version)` — writes `.claude-plugin/plugin.json` with the schema from §3. Uses hardcoded constants; version comes from the arg. | ~50 |
| 003 | `copyFlat(libFS, subdir, outDir)` — copies `library/<subdir>/*.md` files verbatim to `<outDir>/<subdir>/` (used for agents, commands, output-styles). Preserves mode bits via `fs.ReadFile` + `os.WriteFile`. | ~60 |
| 004 | `restructureSkills(libFS, outDir)` — walks `library/skills/*.md`; for each, extracts frontmatter name (fallback: basename), writes to `<outDir>/skills/<name>/SKILL.md`. Ensures the frontmatter `name` matches the dir name (rewriting if the source used a different `name` field). | ~90 |

**Exit criteria:** package builds and passes `go vet`; unit tests for each helper pass in isolation.

---

### Phase 2 — CLI subcommand

**Goal:** Expose the generator through cobra.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 005 | `cmd/build_plugin.go` — `ai-setup build-plugin [--out <path>]`. Resolves `libFS` via `library.GetLibraryFS()`, defaults `out` to `./dist/plugin`, calls `plugin.Build(libFS, out, Version)`, prints file count + path. Registered in `root.go` via `rootCmd.AddCommand`. | ~70 |
| 006 | Add `--force` flag: when unset, abort if `outDir` exists and is non-empty. When set, wipe the dir first (`os.RemoveAll`). Protects users from accidentally overwriting a working plugin. | ~30 |

**Exit criteria:** `ai-setup build-plugin --help` shows the flag; running it from the repo root produces `dist/plugin/` with all expected files; second run without `--force` errors cleanly.

---

### Phase 3 — Tests + docs

**Goal:** Lock the generator's behavior against library drift and document usage.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 007 | Unit tests for `internal/plugin/`: golden-file assertions — generate into `t.TempDir()`, walk the result, assert every `library/agents/*.md` → `agents/*.md`, every `library/skills/*.md` → `skills/<name>/SKILL.md` with matching body bytes, every `library/claudecode/commands/*.md` → `commands/*.md`, every `library/claudecode/output-styles/*.md` → `output-styles/*.md`, and `.claude-plugin/plugin.json` has `name == "ai-setup"` and the expected version. | ~150 |
| 008 | Integration test: run the generator + parse every `SKILL.md` frontmatter, assert `name` field matches containing dir; assert every agent file has `name` frontmatter. Mirrors the existing `internal/library/copilot_schema_test.go` style. | ~80 |
| 009 | `cmd/build_plugin_test.go` — invokes the cobra command through a test harness; verifies `--out` path works; verifies `--force` behavior (existing dir with content → error without flag, wiped with flag). | ~80 |
| 010 | Knowledge map: mark spec 016 complete; add Packages Reference entry for `internal/plugin/`. `--help` text in `cmd/build_plugin.go` serves as the user-facing docs; no separate doc file. | ~10 |

**Exit criteria:** `go test ./... -count=1` green; manually running `ai-setup build-plugin` followed by `claude --plugin-dir ./dist/plugin` loads all agents/skills/commands under the `ai-setup:` namespace.

---

## 5. Non-goals (for clarity)

- **No MCP server emission** — requires handling `${VAR}` placeholders which plugins don't support; deferred to a follow-up if demand emerges.
- **No `marketplace.json`** — publishing workflow is a separate spec.
- **No auto-update / release automation** — generating into `dist/` is manual today.
- **No rule/CLAUDE.md emission** — plugins have no `rules/` slot; those stay in `ai-setup init`.
- **No plugin install side effects from `ai-setup init`** — the two surfaces stay independent.

---

## 6. Risk register

| # | Risk | Mitigation |
|---|---|---|
| R1 | Flat-skill → `SKILL.md` restructure silently loses content | Unit test (task 007) asserts byte-identical body between source and generated SKILL.md |
| R2 | Agent frontmatter fields not supported by plugin schema (e.g. `hooks`, `mcpServers`, `permissionMode` are forbidden for plugin-shipped agents per upstream docs) | Generator validates: if any forbidden field is present in a source agent, log a warning and strip it from the output (spec 013's agents don't use these; risk is preventive). |
| R3 | `version` not passed through at build time (stays `"0.0.0-dev"`) | Document that published plugins should be built with the release binary. Tests pin to whatever `cmd.Version` happens to be at test time. |
| R4 | `dist/` accidentally committed | Add `dist/` to `.gitignore` as part of task 006. |
| R5 | Skill's frontmatter `name` collides with filename (e.g. source file `implement.md` has `name: something-else`) | Dir name wins. Generator rewrites frontmatter `name` to match directory (Claude plugin docs: when SKILL.md is in a directory, frontmatter `name` determines invocation; we align both to keep behavior deterministic). |

---

## 7. Sequencing & sizing

- Phase 1 — 4 tasks, ~320 LOC. Generator core.
- Phase 2 — 2 tasks, ~100 LOC. CLI surface.
- Phase 3 — 4 tasks, ~320 LOC. Tests + knowledge map.

Total: 10 tasks, ~740 LOC. Fits a single implementation session; each phase leaves the tree buildable and tested.

---

## 8. Follow-ups queued for later specs

- Marketplace repo + `marketplace.json` generator (Option C from research)
- MCP server emission with a `--skip-placeholders` flag
- CI: build plugin on every ai-setup release, upload as an artifact
- Rules/CLAUDE.md shipping surface (needs upstream support or a convention)
