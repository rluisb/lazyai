# Research — 012: Claude Code Deep Setup (Global / Project / Workspace)

**Date:** 2026-04-19
**Author:** Ricardo (with Scout + claude-code-guide agents)
**Phase:** Research (R of RPI) — awaiting HUMAN GATE before Plan

---

## 1. Goal

Make `ai-setup` produce a Claude Code configuration that is **100 % conformant** with Claude Code's own expectations at all three scopes (project, workspace, global). Where the `claude` CLI offers a deterministic, scriptable bootstrap (notably `mcp add`, `plugin install`, `plugin marketplace add`, `plugin validate`), **delegate to it** so the on-disk layout converges on whatever Claude itself canonically writes; layer our library assets on top of that structure for everything the CLI will not create for us.

Workspace is an `ai-setup` concept (no upstream equivalent). It is treated like project-scope but rooted at a user-chosen directory; it must still respect Claude's file-layout contract.

---

## 2. Authoritative Claude Code Layout

Sources: <https://code.claude.com/docs/en/settings.md>, `/claude-directory.md`, `/sub-agents.md`, `/skills.md`, `/commands.md`, `/hooks.md`, `/plugins-reference.md`, `/cli-reference.md`. Cross-checked with `claude --help` and a live inspection of `~/.claude/`.

### 2.1 Per-scope file layout (canonical)

```
PROJECT (committed)               WORKSPACE (= project at user dir)   GLOBAL / USER
<project>/                        <workspace>/                        ~/
├── CLAUDE.md                     ├── CLAUDE.md                       ├── .claude.json   ← user-state (auto-managed)
├── .mcp.json                     ├── .mcp.json                       ├── .claude/
└── .claude/                      └── .claude/                            ├── settings.json   (mcpServers live here for user-scope)
    ├── settings.json                  ├── settings.json                  ├── CLAUDE.md       (personal conventions)
    ├── settings.local.json (gitignored) ├── settings.local.json          ├── agents/<name>.md
    ├── CLAUDE.md (project conventions)  ├── CLAUDE.md                    ├── commands/<name>.md
    ├── agents/<name>.md                 ├── agents/<name>.md             ├── skills/<name>/SKILL.md
    ├── commands/<name>.md               ├── commands/<name>.md           ├── rules/<name>.md
    ├── skills/<name>/SKILL.md           ├── skills/<name>/SKILL.md       ├── output-styles/<name>.md
    ├── rules/<name>.md                  ├── rules/<name>.md              ├── hooks/  (optional, also inline)
    ├── output-styles/<name>.md          ├── output-styles/<name>.md      ├── statusline.json
    ├── hooks/                           ├── hooks/                       ├── keybindings.json
    └── (no plugins dir — mgr-owned)     └── (idem)                       └── plugins/  (CLI-managed)
```

### 2.2 Settings precedence

`Managed (enterprise)` > `User (~/.claude/settings.json)` > `Project (.claude/settings.json)` > `Local (.claude/settings.local.json)`.

Per-key wins are **last-wins** for scalars; directories (`agents/`, `skills/`, …) **merge by union** across scopes.

### 2.3 MCP storage matrix (the part we get wrong today)

| Scope | Where Claude expects MCP servers | How `claude` CLI writes them |
|---|---|---|
| user (our "global") | `~/.claude/settings.json` ➜ `mcpServers` key (and/or `~/.claude.json`; CLI is the source of truth) | `claude mcp add --scope user …` |
| project (committed) | `<project>/.mcp.json` ➜ `mcpServers` | `claude mcp add --scope project …` |
| local (gitignored, per-user) | `<project>/.claude/settings.local.json` ➜ `mcpServers` | `claude mcp add --scope local …` (default) |

`.mcp.json` at **global** scope is wrong — Claude does not read it from `~/`. We currently skip it correctly at compile time, but we hand-merge into `settings.json`; the CLI would do it for us.

### 2.4 Frontmatter contracts (must match exactly)

- **agents/*.md** — `name`, `description`, optional: `model`, `effort`, `maxTurns`, `tools` (whitespace-separated), `disallowedTools`, `skills`, `memory`, `background`, `isolation`, `hooks`.
- **skills/*/SKILL.md** — `name` (or dir name), `description`, `when_to_use`, `argument-hint`, `disable-model-invocation`, `user-invocable`, `allowed-tools`, `model`, `effort`, `context: fork`, `agent`, `hooks`, `paths`, `shell`.
- **commands/*.md** — same shape as skills (commands are the legacy flat form of skills).
- **output-styles/*.md** — `name`, `description`, `keep-coding-instructions`.

**Whitespace-separated tools list** is the canonical form for agents. Our existing `NormalizeToolsFrontmatter("comma")` may emit the wrong delimiter — to verify in the Plan phase.

---

## 3. CLI surface available for non-interactive orchestration

From `claude --help` (live, current install):

| Command | Useful flags | Verdict for ai-setup |
|---|---|---|
| `claude mcp add <name> <commandOrUrl> [args...]` | `-s/--scope local\|user\|project`, `-t/--transport stdio\|sse\|http`, `-e/--env KEY=VAL`, `-H/--header`, `--client-id` | **USE** — replaces our hand-rolled JSON merge for all three scopes. |
| `claude mcp add-json <name> <json>` | `-s/--scope`, `--client-secret` | **USE** as fallback when args are too rich for `add` (e.g. `cwd`, `disabled`, `alwaysAllow`). |
| `claude mcp list` / `mcp get <name>` | — | **USE** for post-install verification. |
| `claude mcp remove <name>` | `-s/--scope` | Use for clean re-runs. |
| `claude plugin install <plugin>[@marketplace]` | `-s/--scope user\|project\|local` | **USE** for any "shipped as plugin" assets. |
| `claude plugin list [--json]` | `--available`, `--json` | **USE** for post-install verification. |
| `claude plugin marketplace add <source>` | URL / path / GitHub repo | **USE** if we ever publish ai-setup assets as a marketplace. |
| `claude plugin validate <path>` | — | **USE** in tests against any plugin manifest we emit. |
| `claude agents` | `--setting-sources user,project,local` | List-only — **read-only verification**. |
| `claude doctor` | — | Auto-updater health, not config validation. Limited use. |
| `claude --bare -p "…" --settings <file> --mcp-config <file>` | full programmatic surface | **USE** for headless smoke validation in CI later. |

**Hard limits — no CLI exists for:**

- creating agents, skills, commands, output-styles, rules, hooks, statusline, keybindings, CLAUDE.md, settings.json contents (other than `mcp add` and plugin install).

➜ For everything in that list we **must write files directly**. The CLI buys us correctness for MCP and plugins only.

**Bootstrap reality:** there is no `claude init`. Claude Code does **not** auto-create `.claude/` in a project; it lazily creates `~/.claude/` on first run. We must produce both.

---

## 4. Current `ai-setup` Claude Code state (inventory + bugs)

### 4.1 Files in scope

| Concern | File |
|---|---|
| Adapter | `internal/adapter/claudecode.go` (≈230 LOC) |
| MCP compile | `internal/adapter/mcp_compiler.go#compileClaudeCodeMCP` (lines 215-260) |
| Scope resolver | `internal/adapter/scope.go` |
| Shared helpers | `internal/adapter/shared.go` (`InstallToolContextFiles`, `NormalizeToolsFrontmatter`, …) |
| Root memory | `internal/scaffold/root.go` (`memoryDocDestPath`, `fillClaudeMdPlaceholders`) |
| Library assets | `library/agents/*.md`, `library/skills/*.md`, `library/tool-agents/*.md`, `library/root/CLAUDE.template.md` |
| Tests | `adapter_scope_test.go`, `mcp_compiler_test.go` (one global-skips test only) |

### 4.2 Per-scope behaviour today vs expected

| Scope | What we write today | What Claude expects | Verdict |
|---|---|---|---|
| project | `.claude/settings.json`, `.claude/agents/*.md`, `.claude/skills/<n>/SKILL.md`, `.claude/rules/typescript.md`, `.claude/CLAUDE.md`, `.claude/agents/CLAUDE.md`, `.claude/skills/CLAUDE.md`, `.mcp.json`, root `CLAUDE.md` | Same paths, plus `commands/`, `output-styles/` directories | ✅ Mostly correct. ❌ Missing `commands/`, `output-styles/`. ⚠ MCP via direct write instead of `mcp add --scope project`. |
| workspace | Identical to project, rooted at `<workspace>/.claude/` | Identical to project at workspace dir | ✅ Layout matches. Same gaps as project. |
| global | `~/.claude/settings.json` (with `mcpServers` merged in), agents written **flat** at `~/.claude/builder.md` etc., skills under `~/.claude/skills/<n>/SKILL.md`, `~/.claude/rules/typescript.md`, `~/.claude/CLAUDE.md`, **no orchestrator agent** | `~/.claude/agents/<name>.md` (subdir!), skills as we do, `~/.claude/commands/<name>.md`, `~/.claude/CLAUDE.md` for personal conventions, MCP via `mcp add --scope user` | ❌ **Major bug** — agents are in the wrong directory, will not be discovered. Confirmed live in author's `~/.claude/`. |

### 4.3 Confirmed defects (from explorer report + live filesystem)

1. **Global agents written flat at `~/.claude/`** instead of `~/.claude/agents/`. (`claudecode.go:82-83`) — Claude will not enumerate them. **Verified live**: my `~/.claude/` has `builder.md`, `planner.md`, `scout.md` etc. at the root; `agents/` subdir does not exist.
2. **Tool-context placeholder file (`agents-dir.md`) lands at `~/.claude/CLAUDE.md` at global scope** — collides with the personal-conventions CLAUDE.md and references an `agents/` dir that doesn't exist (because of #1). (`claudecode.go:124-129`)
3. **Orchestrator agent silently dropped at global scope** with no rationale or warning. (`claudecode.go:95`) — Either intentional (then document) or a bug.
4. **`commands/` directory not produced at any scope.** Library has no Claude-targeted command assets. Slash commands are a first-class Claude feature; our scaffold ignores them.
5. **`output-styles/` directory not produced.** No assets, no scaffold.
6. **Hard-coded `rules/typescript.md`** instead of sourcing from `library/rules/`. Brittle and inconsistent with how every other artifact is loaded.
7. **MCP at user scope merged via raw JSON deep-merge** instead of `claude mcp add --scope user`. Two implications: (a) Claude's CLI may rewrite the file in a way that doesn't round-trip through our merge, and (b) we duplicate logic the CLI already owns.
8. **MCP at project scope written by hand to `.mcp.json`** instead of `claude mcp add --scope project`. Same concern.
9. **No post-install verification.** We don't run `claude mcp list`, `claude agents`, or `claude plugin list` to confirm that what we wrote is what Claude can see.
10. **No `claude plugin validate` step** — we have nothing to validate today, but if we ever ship our agents/skills as a plugin manifest, we'd want this.
11. **Tests assert directory existence only** — no schema validation of frontmatter, no MCP content shape test, no parity assertion that "global agents/ and project agents/ contain the same files". Easy class of regressions.
12. **`NormalizeToolsFrontmatter("comma")` for orchestrator** — Claude expects whitespace-separated tools in agent frontmatter. Comma may still parse, but it's not the documented form.

---

## 5. Orchestration strategy — three options

### Option A: "Direct-write only" (status quo, fix bugs)
Keep writing every file ourselves; just fix the global agents path, add `commands/`/`output-styles/`, source the typescript rule from library, drop comma normalization. **No CLI orchestration.**

- ✅ Pure-Go, deterministic, no external dependency on `claude` binary.
- ✅ Matches how every other adapter works today.
- ❌ We continue to own a contract that Claude controls (settings.json shape, MCP server precedence). Future Claude updates can break us silently.
- ❌ Doesn't address the user's stated goal: *"orchestrate agent CLI tool commands … so we start creating files in correct format and place"*.

### Option B: "CLI-orchestrated where it exists, direct-write for the rest" (recommended)
Use `claude mcp add(-json)` and `claude plugin install` to delegate the parts the CLI owns; write files ourselves for agents, skills, commands, output-styles, rules, hooks, statusline, keybindings, CLAUDE.md, the parts of settings.json the CLI does not touch.

- ✅ Forces convergence on Claude's canonical layout for MCP and plugins.
- ✅ Adds `claude mcp list` / `claude plugin list --json` post-checks essentially for free.
- ✅ Falls back to direct-write when `claude` is missing (so we don't break offline / fresh installs).
- ⚠ Adds a runtime dependency on `claude` being on PATH; needs a probe + graceful fallback (we already have this pattern via `ctx.DriveCLI`).
- ⚠ Two write paths (CLI vs direct) ➜ test matrix doubles for MCP code paths.

### Option C: "Ship as a Claude plugin"
Repackage our agents/skills/commands as a Claude marketplace plugin (`.claude-plugin/plugin.json` + `agents/`, `skills/`, …). `ai-setup` would then run `claude plugin marketplace add <our-marketplace>` then `claude plugin install ai-setup --scope user|project`.

- ✅ One CLI call installs everything; Claude owns the on-disk layout entirely.
- ✅ User can `claude plugin uninstall` to fully remove us.
- ❌ Workspace scope has no plugin equivalent — would still need direct-write fallback.
- ❌ Big architectural pivot; existing scope-resolver / configmerge / library structure all become "alternative path".
- ❌ Multi-tool support (Gemini, Codex, Copilot, OpenCode) doesn't benefit; Claude becomes an outlier.
- ➜ Park for a follow-up spec; not the right move now.

**Recommendation:** **Option B**, with Option C kept on the roadmap as a v2 if the plugin authoring story matures.

---

## 6. Open questions for the human gate

1. **CLI dependency policy.** Are we OK requiring `claude` on PATH for the "happy path" and falling back to direct-write only when it's missing — or do we want CLI use behind an opt-in flag (extending today's `ctx.DriveCLI`)?
2. **Orchestrator at global scope.** Was the exclusion intentional? If yes, please state the reason so I can document it in the plan. If no, I'll add it.
3. **`.claude/CLAUDE.md` at user scope vs `~/.claude/CLAUDE.md`.** Today we write the tool-context (`agents-dir.md` content) to that path. Claude treats `~/.claude/CLAUDE.md` as the user's *personal-conventions* file. We almost certainly want our context page to live at `~/.claude/agents/CLAUDE.md` (next to the agents) and leave the personal-conventions file alone (or only template-fill it with placeholders). Confirm.
4. **Commands & output-styles content.** We have **zero** Claude-targeted `commands/` or `output-styles/` assets in `library/`. Should this spec ship a small starter set (e.g. mirror the OpenCode `library/opencode/commands/{review,test,commit}.md` set), or just create the directories and document them as user-extensible?
5. **Settings.local.json.** We currently never write `.claude/settings.local.json`. Anything user-machine-specific (e.g. `defaultMode`, machine-specific hooks) is a candidate. Out of scope for this spec, or in?
6. **Plugin path (Option C).** Confirm "park for later" so I don't blend it into this plan.
7. **Workspace special-case.** Confirm: workspace = project layout rooted at user-chosen dir; the only "extra" is whatever ai-setup metadata we already write at workspace scope (which I haven't itemized yet — flag if there's something specific I should preserve).

---

## 7. Proposed scope for the Plan phase (preview, not committed)

If Option B + the answers above land roughly as I expect, the plan would phase as:

- **Phase 1 — Fix structural bugs (no CLI yet):** correct global `agents/` subdir, correct context-file placement, source rules from library, fix tools-frontmatter delimiter, add `commands/` + `output-styles/` directories.
- **Phase 2 — CLI orchestration for MCP:** route MCP install through `claude mcp add(-json) --scope {local|user|project}` with direct-write fallback, then `claude mcp list` post-check.
- **Phase 3 — CLI orchestration for plugins (if applicable):** wire `claude plugin install` for any plugins we want to opt users into.
- **Phase 4 — Starter library assets:** add a minimal `library/claudecode/{commands,output-styles}/` set mirroring the OpenCode bundle.
- **Phase 5 — Tests + verification:** frontmatter-schema tests, MCP round-trip tests, post-install `claude mcp list` / `claude agents` smoke, and a `claude plugin validate` test if we author a manifest.

Each phase leaves the codebase shippable.

---

## 8. Decisions (from human gate, 2026-04-19)

| # | Question | Decision | Planning implication |
|---|---|---|---|
| Q1 | CLI dependency policy | **Try CLI, silent fallback to direct-write** | Plan must probe `claude` presence, use it when available, fall back transparently on error. Both code paths need tests. Emit a single-line warning on fallback so users notice; don't make it an error. |
| Q2 | Orchestrator at global scope | **Include it — was a bug** | Remove the `!isGlobal` gate at `claudecode.go:95`. Orchestrator ships at global whenever `EnableServers` includes it. |
| Q3 | `~/.claude/CLAUDE.md` collision | **Move tool-context to `agents/CLAUDE.md`** | Relocate `agents-dir.md` render target at global scope to `~/.claude/agents/CLAUDE.md`. Leave `~/.claude/CLAUDE.md` alone on re-run; only template-fill on first install if absent. |
| Q4 | Starter `commands/` + `output-styles/` | **Port the OpenCode starter set** | Create `library/claudecode/commands/{review,test,commit}.md` + ~2 output-styles (`terse.md`, `explanatory.md`) with Claude frontmatter. Match OpenCode parity. |
| Q5 | `.claude/settings.local.json` | **Out of scope — follow-up ticket** | Do not write or stub in 012. Add a line in Pending/Follow-up in `specs/KNOWLEDGE_MAP.md`. |
| Q6 | Ship-as-Claude-plugin (Option C) | **Park for later spec** | 012 commits to Option B. Do not author a plugin manifest in this spec. |
| Q7 | Workspace scope extras | **Project-identical for Claude Code** | No workspace-only branches for Claude. Workspace = project at workspace dir. Preserve any existing cross-tool workspace metadata (not a Claude concern). |

---

## 9. References


- Claude Code docs (settings, claude-directory, sub-agents, skills, commands, hooks, plugins-reference, cli-reference): <https://code.claude.com/docs/en/>
- Live `claude` CLI surface: `claude --help`, `claude mcp --help`, `claude plugin --help` (captured 2026-04-19)
- Spec 011 (OpenCode deep setup) for parallel structural pattern: `specs/011-opencode-deep-setup/`
- Current adapter: `internal/adapter/claudecode.go`, `internal/adapter/mcp_compiler.go:215-260`
