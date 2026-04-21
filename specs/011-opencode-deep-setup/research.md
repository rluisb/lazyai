# Research — 011: OpenCode Deep Setup (Global / Project / Workspace)

**Date:** 2026-04-19
**Author:** Ricardo (with Scout agent)
**Phase:** Research (R of RPI) — awaiting HUMAN GATE before Plan

---

## 1. Goal

Make `ai-setup` produce an OpenCode configuration that is **100% conformant** with opencode's own expectations at all three scopes (project, workspace, global). Where possible, **delegate structural bootstrap to the `opencode` CLI itself** (so the scaffold converges on whatever layout opencode canonically creates), then layer our library assets on top of that structure.

Workspace is an ai-setup concept (no upstream equivalent), so it is handled like project-scope but at a user-chosen directory, and still must respect opencode's file layout contract.

---

## 2. Authoritative OpenCode Layout (from `opencode debug paths` + docs)

`opencode debug paths` (local install, v1.4.9) reports:

| Path | Purpose |
|---|---|
| `~/.config/opencode`         | **Config root** (global scope target) |
| `~/.local/share/opencode`    | Data (sessions, storage) |
| `~/.local/state/opencode`    | State (DB, WAL) |
| `~/.cache/opencode/bin`      | Cache + embedded binaries |

Per [opencode.ai/docs](https://opencode.ai/docs/config/), both global and project configs share the same directory schema (plural dir names):

```
<root>/
├── opencode.json              ← main config ($schema, mcp, mode, permission, ...)
├── opencode.jsonc             ← alt form; if present, overrides .json
├── AGENTS.md                  ← explicit instructions (via config.instructions)
├── agents/<name>.md           ← custom agents (frontmatter: description, mode, tools, model)
├── commands/<name>.md         ← slash commands (frontmatter: description, agent, model)
├── modes/<name>.md            ← chat modes (alt: defined inline in opencode.json)
├── skills/<name>/SKILL.md     ← skills (directory-per-skill)
├── plugins/                   ← JS/TS plugins (installed via `opencode plugin <module>`)
└── themes/                    ← custom themes
```

- **Global root:** `~/.config/opencode/`
- **Project root:** `<project>/.opencode/`
- **Workspace root (our concept):** `<workspace-dir>/.opencode/` — same shape as project.

**Discovery rules** (per docs):
- `opencode.json` is merged across global → project (keys deep-merge, arrays replace).
- `AGENTS.md` is **not** auto-discovered from project root — it must be declared in `config.instructions`.
- `.opencode/agents/`, `commands/`, `modes/`, `skills/` subdirs **are** auto-discovered.

---

## 3. Current ai-setup Implementation (inventory)

### 3.1 Adapter: `internal/adapter/opencode.go`

| Responsibility | Behavior | Gap? |
|---|---|---|
| Path resolution | Via `ResolveToolRoot(ToolIdOpenCode, scope, ctx)` — project/workspace → `<target>/.opencode`, global → `~/.config/opencode` (`scope.go:49-62`) | ✅ Matches canonical |
| `opencode.json` | Merges `{$schema, instructions:["AGENTS.md"], permission:{edit:"ask",bash:"ask"}}` via `configmerge.MergeJSONFile` (`opencode.go:40-61`) | ⚠️  `instructions` is the *documented* way to pull in AGENTS.md — need to verify key name matches current opencode schema |
| `agents/` | Copies library agents (minus orchestrator unless enabled), applies `StripFrontmatterAndInjectModel` (`opencode.go:64-91`) | ✅ Present |
| `skills/` | Directory-per-skill layout `<skill>/SKILL.md` (`opencode.go:94-105`) | ✅ Matches docs |
| `commands/` | **Directory is created** (`opencode.go:31`) but **no commands are installed** | ❌ **Gap** — library has no opencode command assets; Gemini has `library/commands/*.toml`, Copilot has `library/chatmodes/*.chatmode.md`; opencode has nothing |
| `modes/` | **Not created, not populated** | ❌ **Gap** |
| `AGENTS.md` (root) | Installed via `InstallToolContextFiles` → `.opencode/AGENTS.md`, `.opencode/agents/AGENTS.md`, `.opencode/skills/AGENTS.md` (`opencode.go:108-117`) | ✅ Present — but only root one is referenced by `instructions` |
| `plugins/` | Not handled | 🟡 Out of scope for MVP — opencode has its own `opencode plugin <module>` installer |
| `themes/` | Not handled | 🟡 Out of scope |

### 3.2 MCP compile: `internal/adapter/mcp_compiler.go#compileOpenCodeMCP` (lines 115-181)

- Writes to `<root>/opencode.jsonc` (not `.json`) — this creates **two config files** if default-path pre-exists. Needs reconciliation (see §5).
- Sets `existingConfig["mcp"] = ocMcp` — **overwrites** the whole `mcp` key rather than deep-merging individual servers. Acceptable for now (compile-time regen), but users who hand-authored an `mcp` entry will lose it.
- Remote/local schema transformation matches docs (`type`, `command`/`url`, `environment`/`headers`).

### 3.3 Tests (adapter_test.go, scope_test.go, mcp_compiler_test.go, scaffold_test.go)

All 3 scopes are covered for `Install()` and `CompileMCPForTool`. No tests exist for:
- Correct frontmatter schema on installed agents (no assertion that opencode can *parse* the files).
- Roundtrip validation via `opencode debug config` or `opencode debug agent <name>`.

---

## 4. What the `opencode` CLI Can Do For Us

`opencode --help` + subcommand help (v1.4.9):

| Subcommand | Scriptable? | Useful for scaffold? |
|---|---|---|
| `opencode agent create` | **Partially** — flags exist: `--path`, `--description`, `--mode`, `--tools`, `--model`. No `--name` flag surfaced — likely still prompts for name interactively | 🟡 Too narrow: we copy templates verbatim, not interactive wizards. Would still need file-write fallback. |
| `opencode agent list` | ✅ Yes (read-only) | ✅ **Validation** — confirm our scaffolded agents are discovered |
| `opencode mcp add` | ❌ **No flags** — fully interactive | ❌ Cannot orchestrate; stick with file-write |
| `opencode mcp list` | ✅ Yes | ✅ Validation |
| `opencode plugin <module>` | ✅ Fully scriptable (`-g` for global, `-f` for force) | 🟡 Optional — not in scope for 011 |
| `opencode debug config` | ✅ Yes (stdout JSON) | ✅ **Validation** — authoritative resolved config |
| `opencode debug paths` | ✅ Yes | ✅ Done once; paths are stable |
| `opencode debug agent <name>` | ✅ Yes | ✅ **Validation** — confirm each agent parses |
| `opencode debug skill` | ✅ Yes | ✅ Validation for skills |

**Conclusion on CLI delegation:** The CLI is **better as a validator than as a writer**. `agent create` and `mcp add` are interactive-heavy. The pragmatic pattern is:

1. Write files directly (as today), matching the canonical layout exactly.
2. Optionally run `opencode debug config` / `opencode debug agent --list` **as a post-install smoke test** and surface mismatches to the user.

This matches the approach the user asked for — "respect the structure for the agent CLI tool" — without depending on interactive prompts the scaffold can't drive.

---

## 5. Gaps & Issues Identified

| # | Gap | Severity | Evidence |
|---|---|---|---|
| G1 | `commands/` dir created but never populated — no opencode command library assets exist | High | `opencode.go:31`, `library/commands/` (Gemini-only TOML) |
| G2 | `modes/` not handled at all | High | Not in `opencode.go`; no `library/modes/` |
| G3 | `opencode.json` at install vs. `opencode.jsonc` at MCP compile creates two configs | Medium | `opencode.go:37` vs `mcp_compiler.go:122` |
| G4 | `instructions: ["AGENTS.md"]` — need to verify (a) the key is still supported and (b) the path is relative to config file (so `AGENTS.md` resolves to `<root>/AGENTS.md`) | Medium | `opencode.go:42` — unverified against current opencode schema |
| G5 | No post-install validation via `opencode debug config` / `opencode debug agent` | Medium | No test or runtime check |
| G6 | Agent frontmatter: our `StripFrontmatterAndInjectModel` strips the original frontmatter. We need to confirm the **re-injected** frontmatter matches opencode's expected schema (`description`, `mode`, `tools`, `model`) | High | `opencode.go:72` — not cross-checked against opencode spec |
| G7 | `mcp` key overwrite on compile loses user-authored servers | Low | `mcp_compiler.go:138` — `existingConfig["mcp"] = ocMcp` |
| G8 | Workspace scope uses same shape as project — confirmed fine. But our "extras" (e.g., top-level workspace memory docs) need explicit rules so they don't pollute `.opencode/` | Low | `scope.go:52` — project/workspace share path |

---

## 6. Workspace-Scope Specifics

Workspace = user-selected directory that is **not** the git project root (e.g., `~/Workspaces/client-a/`). Per spec 008/009:

- We treat it as **project-shaped layout** at the chosen dir (same `.opencode/` tree inside).
- Extras: workspace-level memory docs, shared rules, selection-store entries — these live outside `.opencode/` and must not be confused with opencode's own config files.

No opencode-side change needed here — workspace is transparent to opencode. But we must ensure the adapter receives the right `ctx.TargetDir` for workspace scope (currently handled by scope resolver; spec 009 confirmed this works).

---

## 7. Decisions (locked via interview, 2026-04-19)

1. **Asset layout — D1:** Add dedicated `library/opencode/commands/` and `library/opencode/modes/` dirs with opencode-specific frontmatter. Keep Gemini's `library/commands/*.toml` separate (different schemas).
2. **Post-install validation — D2:** Opt-in, gated on `opencode` being on `PATH`. Run `opencode debug config` and `opencode debug agent <name>` after install; surface mismatches as non-fatal warnings.
3. **Agent frontmatter — D3:** Write a dedicated opencode frontmatter emitter that produces `{description, mode, tools, model, permission}` per opencode's schema. Strip source frontmatter, inject clean block.
4. **MCP merge — D4:** Switch from whole-key overwrite to per-server deep-merge keyed on server name. ai-setup-managed servers update/insert; user-authored servers preserved.
5. **Config file format — D5:** Standardize on `opencode.jsonc` only. Install-time and compile-time both write to `.jsonc`. Migration: if `.json` exists at install, rename with a `.bak` sidecar and write `.jsonc`.
6. **Plugins (expanded scope) — D6:** Include an optional plugin-selection step in the wizard and shell out to `opencode plugin <module>` when binary is on PATH. Gated the same way as validation (no-op if opencode missing).
7. **Workspace extras — D7:** Workspace memory docs (AGENTS.md, specs/) live outside `.opencode/` at the workspace root. `opencode.jsonc` `instructions` points at `../AGENTS.md` (or absolute path) so opencode sees them via the config key, not via mirroring.

---

## 8. Assumptions (to confirm in Plan phase)

- [ ] **A1:** opencode's `instructions` config key is still the canonical way to pull AGENTS.md into context (verify against current opencode.ai/docs/config/).
- [ ] **A2:** Installing `commands/*.md` and `modes/*.md` with proper frontmatter is all that's needed for opencode to auto-discover them (no registration step).
- [ ] **A3:** `opencode agent list` reads from both global and project `.opencode/agents/` simultaneously (confirms merged discovery).
- [ ] **A4:** Validation via `opencode debug *` is safe to run without side effects on user state.

---

## 9. Confidence

- **Medium** overall.
- **High** on canonical paths (confirmed via `opencode debug paths` on the local box).
- **Medium** on frontmatter schema — docs describe it but our library templates haven't been cross-checked key-by-key.
- **Medium** on CLI orchestration viability — `mcp add` being interactive-only is a concrete blocker; `agent create` is likely only partially scriptable without deeper testing.

---

## 10. Next (after HUMAN GATE)

Move to **Plan** phase. Plan should cover, in order:
1. Decide CLI-delegation scope (validation-only vs. selective use of `opencode agent create`).
2. Define library asset additions (`library/opencode/commands/`, `library/opencode/modes/`).
3. Adapter refactor to populate `commands/` and `modes/`.
4. Reconcile `.json` vs `.jsonc` config target.
5. Agent frontmatter schema validation.
6. Post-install `opencode debug config` validation (opt-in, gated on binary presence).
7. MCP deep-merge on compile.
8. Test additions: schema-parse tests + `opencode debug *` roundtrip (if opencode on PATH in CI).
