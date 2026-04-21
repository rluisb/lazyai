# Plan — 017: Gemini Deep Setup

**Date:** 2026-04-20
**Phase:** Plan (P of RPI) — awaiting HUMAN GATE before Implement
**Research:** `research.md` (decisions locked in §5)

---

## 1. Objective

Close the Gemini parity gap with specs 011/012/013 by:
1. Moving Gemini-specific library assets into a dedicated `library/gemini/` directory.
2. Adding `ai-setup build-gemini-extension` — a new generator subcommand analog to spec 016's `build-plugin` that produces a loadable Gemini CLI extension from the library.
3. Adding a minimal headless probe (`LookPath("gemini")`) so install feedback is consistent with other tools.

Exit: `ai-setup init --tools gemini` still emits the same config today (no regression); `ai-setup build-gemini-extension --out dist/gemini-extension` produces a `gemini extensions link`-ready directory.

---

## 2. Locked decisions (from research §5)

| # | Answer |
|---|---|
| Q1 | **Option C** — library restructure + extension generator |
| Q2 | Default output `./dist/gemini-extension/` (mirrors `dist/plugin/` from spec 016) |
| Q3 | Ship **canonical MCP catalog, static-only** — skip entries with `${VAR}` placeholders (matches spec 014) |
| Q4 | **Raw template** — bundle `library/root/GEMINI.template.md` as-is so recipients fill in their own `[YOUR_*]` markers |
| Q5 | **Keep** `library/commands/` as a fallback read path for one release; preferred path becomes `library/gemini/commands/` |
| Q6 | Extension name: **`"ai-setup-gemini"`** (disambiguate from the Claude plugin `"ai-setup"`) |
| Q7 | **Add** `LookPath("gemini")` probe; warn on stderr when absent, non-fatal |

---

## 3. Target extension layout

```
dist/gemini-extension/
├── gemini-extension.json       # Generated manifest
├── GEMINI.md                   # Verbatim from library/root/GEMINI.template.md
└── commands/
    ├── commit.toml             # Copied from library/gemini/commands/
    └── …
```

### Manifest payload

```json
{
  "name": "ai-setup-gemini",
  "version": "<cmd.Version>",
  "description": "ai-setup GEMINI.md template and custom commands for Gemini CLI",
  "contextFileName": "GEMINI.md",
  "mcpServers": {
    "<name>": { "command": "...", "args": [...], "cwd": "..." }
  }
}
```

`mcpServers` populated from canonical `.ai/mcp.json` **only** for servers with no `${VAR}` placeholder in `env`/`headers` (same rule spec 014 uses). Servers with placeholders are skipped with a stderr warning listing them (user can hand-edit `settings` field later).

If `.ai/mcp.json` is absent or empty, `mcpServers` is omitted entirely.

---

## 4. Phased breakdown

### Phase 1 — Library restructure (commands move)

**Goal:** Move `library/commands/*.toml` to `library/gemini/commands/` so Gemini has its own per-tool home. Adapter reads from the new path with a fallback to the old one for one release.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 001 | Move every `library/commands/*.toml` → `library/gemini/commands/`. Keep `library/commands/` empty in the repo (or delete if git plays nice; retain the **read path** in the adapter). | ~10 (git moves) |
| 002 | Update Gemini adapter: read from `library/gemini/commands/` first; if empty, fall back to `library/commands/`. Both paths compile-safe in both disk and embedded FS modes. Log a one-line note when the fallback activates. | ~40 |
| 003 | Update the generic commands selection wiring — any reference to `library/commands/` for Gemini should consult the new path via a centralized helper (`library.GeminiCommandsDir()`). | ~30 |

**Exit criteria:** `go test ./... -count=1` green; `ai-setup init --tools gemini` produces identical `.gemini/commands/*.toml` output as before the move.

---

### Phase 2 — Extension generator core

**Goal:** New package `internal/geminiext/` exposes `Build(libFS, catalog, outDir, version) (BuildResult, error)` that writes the extension layout from §3.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 004 | `internal/geminiext/geminiext.go`: `Build()` orchestrator; constants (`ExtensionName = "ai-setup-gemini"`, description, defaults); `writeManifest(outDir, version, mcpServers)` writes `gemini-extension.json`. | ~110 |
| 005 | `copyGeminiMd(libFS, outDir)` — copies `library/root/GEMINI.template.md` → `<outDir>/GEMINI.md` verbatim (no placeholder fill per locked Q4). | ~30 |
| 006 | `copyCommands(libFS, outDir)` — walks `library/gemini/commands/` (with fallback to `library/commands/`), copies each `*.toml` into `<outDir>/commands/`. Preserves subdirectory structure so namespaced commands stay namespaced. | ~60 |
| 007 | `buildMcpServers(catalog)` — filters out entries with `${VAR}` placeholders in env/headers; builds the map for `gemini-extension.json`'s `mcpServers` field. Logs skipped entries to stderr with reason. | ~70 |

**Exit criteria:** Package builds and `go vet` clean; each helper unit-testable in isolation.

---

### Phase 3 — CLI subcommand

**Goal:** Surface the generator via cobra.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 008 | `cmd/build_gemini_extension.go` — `ai-setup build-gemini-extension [--out <path>] [--force]`. Reads `.ai/mcp.json` via existing helpers (nil-safe if absent); resolves `libFS`; calls `geminiext.Build`. Preflight uses the same `preflightOutDir` logic already in `cmd/build_plugin.go` (factor it into a helper in `cmd/build_helpers.go` shared between the two subcommands). | ~80 |

**Exit criteria:** `ai-setup build-gemini-extension --help` shows the flags; running from the repo root produces `dist/gemini-extension/` with the expected files.

---

### Phase 4 — LookPath validation

**Goal:** Give Gemini the same courtesy post-install signal Claude/OpenCode already have. Minimal, non-fatal.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 009 | `GeminiAdapter.RunHeadlessValidation`: `exec.LookPath("gemini")`; if absent, log an info line (not a warning — Gemini is optional until the user runs it); `CanRunHeadless()` continues to return `false` since there's no real non-interactive validation we can perform. | ~20 |

**Exit criteria:** `ai-setup init --tools gemini` prints a single info line about Gemini's CLI status at the end of install; unit test asserts the behavior.

---

### Phase 5 — Tests + docs

**Goal:** Golden coverage of every new path and knowledge-map updates.

| Task | Deliverable | Rough LOC |
|---|---|---|
| 010 | Unit tests for `internal/geminiext/`: `TestBuild_WritesManifest`, `TestBuild_CopiesGeminiMdVerbatim`, `TestBuild_CopiesCommandsWithNamespacing`, `TestBuild_SkipsPlaceholderMcp`, `TestBuild_OmitsMcpWhenCatalogEmpty`. Use an in-memory `fstest.MapFS` like spec 016's plugin tests. | ~180 |
| 011 | Integration test: parse the generated `gemini-extension.json`, assert `name == "ai-setup-gemini"`, `contextFileName == "GEMINI.md"`, `mcpServers` present when catalog has entries. | ~60 |
| 012 | `cmd/build_gemini_extension_test.go`: preflight tests (reuse the shared helper) + golden output tests. | ~80 |
| 013 | Update Gemini adapter tests to cover the new commands path + fallback (assert fallback activates only when `library/gemini/commands/` is absent from the test FS). | ~60 |
| 014 | Update `specs/KNOWLEDGE_MAP.md`: mark spec 017 complete; add Packages Reference rows for `internal/geminiext/` and `cmd/build_gemini_extension.go`. | ~10 |

**Exit criteria:** `go test ./... -count=1` green; knowledge map reflects ship.

---

## 5. Non-goals

- **No Gemini theme emission** — `gemini-extension.json` supports themes, but ai-setup has no theme content today.
- **No auto-publishing** — just like spec 016, publishing an extension is a manual `gemini extensions link` by the user.
- **No rewrite of existing Gemini adapter** — Phase 1's move is path-swap only; no behavior change.
- **No Gemini `/init` replication** — `ai-setup init` already writes `GEMINI.md`; don't duplicate.
- **No codex deep setup** — out of scope; separate spec 018 candidate.

---

## 6. Risk register

| # | Risk | Mitigation |
|---|---|---|
| R1 | File move breaks embedded FS (`go:embed`) pattern if the root embed directive doesn't include the new dir | Ensure `library/gemini/` is under the existing embed root; test that `library.GetLibraryFS()` can read `gemini/commands/commit.toml`. |
| R2 | Fallback logic accidentally double-reads, producing duplicates | Helper returns ONE path; caller never reads both. Unit test covers the fallback-active case. |
| R3 | Generator ships MCP with placeholders, breaking Gemini's extension load | Static-only filter in `buildMcpServers` (spec 014 pattern); unit test asserts placeholder servers are skipped. |
| R4 | User already has `~/.gemini/extensions/ai-setup-gemini/` from a prior run | Generator writes to `dist/` not `~/.gemini/`; installation is the user's choice via `gemini extensions link`. Not our problem. |
| R5 | `preflightOutDir` helper move to a shared file breaks the spec 016 tests | Keep the function signature identical; cobra test re-runs should pass with the move. |

---

## 7. Sequencing & sizing

| Phase | Tasks | LOC |
|---|---|---|
| 1. Library restructure | 001–003 | ~80 |
| 2. Extension generator core | 004–007 | ~270 |
| 3. CLI subcommand | 008 | ~80 |
| 4. LookPath validation | 009 | ~20 |
| 5. Tests + docs | 010–014 | ~390 |
| **Total** | **14 tasks** | **~840 LOC** |

Fits a single implementation session; each phase leaves the tree buildable and tested.

---

## 8. Follow-ups queued for later specs

- Spec 018 — Codex deep setup (analogous treatment for Codex). Separate research required.
- Follow-up to 017 — delete `library/commands/` fallback path once one release ships.
- Follow-up to 017 — auto-derive Gemini extension `settings[]` from `${VAR}` placeholders so MCP entries with secrets can ship.
- Gemini theme bundling (if contributor interest emerges).
