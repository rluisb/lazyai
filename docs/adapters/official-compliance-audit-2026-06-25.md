# LazyAI Adapters — Official Tool Compliance Audit (2026-06-25)

> **Question answered:** *Is LazyAI's compiled output fully aligned with each supported AI CLI tool's official documentation — paths, filenames, directory layout, frontmatter/JSON schemas — i.e. exactly what each tool expects?*
>
> **Method:** Independent re-verification. One agent per tool read the actual adapter code (`Install`, `CompileMCP`, scaffold, frontmatter transforms — not just `output_mapping.go`) and compared it against the **live** official docs (2026-06-25). The lead then independently corroborated the highest-risk findings (OpenCode dir, all MCP keys, Antigravity `serverUrl`, **Kiro agents/prompts**). This audit supersedes the self-verification dated 2026-06-21/23 in `capabilities.go` and `capability-matrix.md`.
>
> **Detailed per-tool reports:** `docs/adapters/snapshots/compliance-audit-2026-06-25-<tool>.md` (opencode, claude-code, copilot, pi, omp, antigravity, kiro). The Kiro snapshot carries a lead-review correction banner (see §4).

---

## 1. Headline answer

**Mostly yes — but not "exactly" everywhere.** All 7 adapters are correctly aligned on their **core surfaces** (root instructions, skills layout, primary MCP keys, hooks schemas, agent directories). No adapter is fundamentally misaligned. However, the audit found **2 dual-target MCP nuances**, **~6 medium-severity divergences** (mostly deprecated-but-working forms and scope-specific path bugs), and a **cluster of low-severity richness/hygiene gaps**. The recurring theme: agent **frontmatter** is copied verbatim or reduced to `name`+`description`, so tool-specific config fields are under-populated — bodies work, structured config is left on the table.

### Verdict table

| Tool | Verdict | Highest-severity issue |
|---|---|---|
| **OpenCode** | ✅ MOSTLY ALIGNED | MED: bundled mode files use deprecated `tools:` frontmatter (should be `permission:`) |
| **Claude Code** | ✅ MOSTLY ALIGNED | MED: global-scope hook script path mismatch; misleading internal MCP doc |
| **Copilot** | ✅ MOSTLY ALIGNED | MED: deprecated `.chatmode.md` (VS Code 2026 prefers `.agent.md`); CLI MCP uses deprecated `sse` |
| **Pi** | ✅ ALIGNED (core) | LOW: `.pi/agents` + MCP are *ecosystem/extension* surfaces, not core — optional |
| **OMP** | ✅ MOSTLY ALIGNED | MED: project `AGENTS.md` lands at bare root (priority 10) not `.omp/AGENTS.md` (priority 100) |
| **Antigravity / Gemini** | ⚠️ MOSTLY ALIGNED (dual-target caveat) | MED: MCP uses `serverUrl` + `mcp_config.json` — **not discoverable by open-source Gemini CLI** (which uses `url` in `settings.json`) |
| **Kiro** | ✅ MOSTLY ALIGNED *(corrected from PARTIAL)* | MED: MCP remote server emits `type:"http"` not in Kiro's documented `{url, headers}` shape |

> **Two findings were corrected during lead review** — see §4. They flip Kiro from "PARTIAL/misaligned" to "mostly aligned" and confirm OpenCode's directory is correct.

---

## 2. What we get exactly right (high-confidence, corroborated)

- **MCP top-level keys are correct per surface** — independently verified in `mcp_compiler.go`:
  - Claude `.mcp.json` / OMP `.omp/mcp.json` / Kiro `.kiro/settings/mcp.json` → `mcpServers` ✅
  - Copilot **VS Code** `.vscode/mcp.json` → `servers` + `inputs` ✅ (a common place to get wrong; we get it right)
  - Copilot **CLI** `~/.copilot/mcp-config.json` → `mcpServers` ✅
  - OpenCode `opencode.json` → `mcp` block with `type: local|remote`, `command[]`, `enabled`, `environment` ✅
- **OpenCode agents at `.opencode/agents/` (plural)** — matches current official docs (the singular `agent/` is stale); `steps` (not deprecated `maxSteps`), `permission` (not deprecated `tools`) ✅
- **Skills** as `<name>/SKILL.md` dir-per-item across all tools that support skills; correct project/global roots (incl. Copilot `~/.copilot/skills`, Antigravity `~/.gemini/config/skills`) ✅
- **Claude** `CLAUDE.md` with `@AGENTS.md` import; hooks `settings.json` schema (`matcher`/`hooks[]`/`type`/`command`) with correct event names; permissions `allow`/`deny` ✅
- **Antigravity** `GEMINI.md` `@./AGENTS.md` import; `settings.json` hook event names `BeforeTool`/`AfterAgent`; global skills/hooks roots (post #486/#497) ✅
- **Kiro** `.kiro/agents/<name>.md` (markdown + YAML frontmatter) and `.kiro/prompts/<name>.md` — **both are real, current Kiro CLI v3 surfaces** (see §4); hooks `version:v1` schema; MCP at `.kiro/settings/mcp.json` ✅
- **Pi** core surfaces (`.pi/skills`, `.pi/prompts`, `.pi/SYSTEM.md`/`APPEND_SYSTEM.md`, `.pi/extensions/*.ts`, `.pi/settings.json` project + `~/.pi/agent/settings.json` global) all exact ✅

---

## 3. Confirmed divergences (cross-tool, prioritized)

### 3.1 MEDIUM — fix or explicitly document

| # | Tool | Divergence | Code | Official says | Recommendation |
|---|---|---|---|---|---|
| M1 | **Antigravity/Gemini** | HTTP MCP emits `serverUrl` into `~/.gemini/config/mcp_config.json`. Open-source **Gemini CLI** reads MCP from `settings.json` `mcpServers` with key **`url`** (SSE) / `httpUrl` (HTTP). Our output is **not discoverable by Gemini CLI**; the Antigravity-desktop path/key is only verifiable via JS-rendered docs. | `mcp_compiler.go:493-518`; pinned by `antigravity_install_test.go:175-179` | `google-gemini/gemini-cli/schemas/settings.schema.json`: `MCPServerConfig.url` (no `serverUrl`) | Make serialization transport-/target-aware: emit `url` into `settings.json` for Gemini CLI, keep `serverUrl`/`mcp_config.json` only if a citable Antigravity-desktop doc confirms it. At minimum, document this as Antigravity-desktop-only. |
| M2 | **Kiro** | Remote MCP server emits `{type:"http", url, headers}` (reuses `toClaudeCodeMcp`). Kiro's documented remote shape is `{url, headers}` with **no `type`**. | `mcp_compiler.go:434-456,400-428` | `kiro.dev/docs/cli/v3/agent-config`, `kiro.dev/docs/mcp/configuration` | Add `toKiroMcp()` emitting url-only remote entries (drop `type`). Likely tolerated today, but undocumented. |
| M3 | **Copilot** | Chat modes emitted as `.github/chatmodes/<name>.chatmode.md`. VS Code (2026) treats chat modes as **deprecated** in favor of custom agents `.github/agents/<name>.agent.md`. | `copilot.go:124-134` | `code.visualstudio.com/docs/agent-customization/custom-agents` | Migrate library chatmodes → `.agent.md` under `.github/agents/`. Keep chatmodes only for back-compat. |
| M4 | **Copilot CLI** | Remote MCP transport serialized as `sse` (deprecated). | `mcp_compiler.go:698` (`toCopilotCLIMcp(..., "sse")`) | Copilot CLI MCP doc example uses `type:"http"`; SSE deprecated in MCP spec | Change CLI remote transport `sse` → `http`. (Sourced from CLI doc, the correct surface for `~/.copilot/mcp-config.json`.) |
| M5 | **Claude Code** | At **global scope**, `settings.json` hook commands reference `${CLAUDE_PROJECT_DIR:-$PWD}/.claude/hooks/<x>.sh`, but scripts install to `~/.claude/hooks/`. Global-only installs point at a non-existent path. (Project scope is fine.) | `claudecode.go:65,76` vs `:37,177-185` | `code.claude.com/docs/en/hooks` | Scope-aware default settings: at global scope use `${HOME}/.claude/hooks/<x>.sh`. |
| M6 | **Claude Code** | Global/user MCP correct only via DriveCLI (`claude mcp add-json -s user` → `~/.claude.json`); with DriveCLI off / CLI absent **nothing is written**. Internal docs/comment claim "mcpServers live in settings.json" — **incorrect**. | `mcp_compiler.go:284,292-296`; `docs/adapters/claude-code.md:73` | `code.claude.com/docs/en/mcp` (scope table) | Fix the internal doc/comment; add a warning when no global MCP is written; consider direct-write fallback to `~/.claude.json`. |
| M7 | **OpenCode** | Bundled mode files use deprecated `tools:` map frontmatter and omit `mode:`. | `library/opencode/modes/plan.md`, `audit.md` | `opencode.ai/docs/agents` (`tools` deprecated → `permission`) | Convert mode files to `permission:`; add `mode: primary`. |
| M8 | **OMP** | Project-scope root instructions land at `<target>/AGENTS.md` (OMP `agents-md` provider, **priority 10**) instead of `.omp/AGENTS.md` (native provider, **priority 100**). Works, but lower discovery precedence. | `scaffold/filemap.go:12`, `scaffold/root.go:81` | `omp://context-files.md` | Map `RootFileByTool[omp]` → `.omp/AGENTS.md` for project scope. |

### 3.2 LOW — richness, hygiene, optional

- **Agent frontmatter under-population (cross-cutting).** Claude & OpenCode reduce agents to `name`+`description`; OMP & Kiro copy canonical frontmatter verbatim. None map to the host's richer agent schema (Claude `tools`/`model`/`color`; OpenCode `mode`; OMP `spawns`/`autoloadSkills`/`thinkingLevel`; Kiro v3 `tools` category tags / `permissions` / `resources` / `model` / `welcomeMessage`). Bodies (system prompts) are honored; structured config is left unset. (`agent_transform.go`, `omp.go:48-56`, `kiro.go`)
- **Copilot** SKILL.md carries LazyAI-internal frontmatter (`tier`, `temperature`, `thinking`, `risk`) not in the official SKILL.md schema — strip before emit or document. (`copilot.go:331-344`)
- **Copilot CLI** stdio MCP uses `type:"stdio"` vs example `"local"` — functional non-issue (docs say stdio is cross-compatible).
- **Claude** commands at legacy `.claude/commands/<name>.md` — still valid; skills are the newer recommended form (modernization, not misalignment).
- **OpenCode** agent `mode:` not propagated from source; `reasoningEffort`/`textVerbosity` emitted as top-level YAML keys (not in `AgentConfig` schema); dead `OpenCodeAgentOpts.Tools` emit path exists.
- **OMP** `.omp/instructions/*.md` and `.omp/RULES.md` native surfaces not emitted/acknowledged; all library hooks placed in `hooks/pre/` regardless of event.
- **Pi** `.pi/agents/<name>.md` and MCP (`pi-mcp-extension` → `.pi/mcp.json`) are **ecosystem/extension** surfaces, not Pi core — currently `.pi/agents` is emitted (extension-dependent consumer), MCP is a no-op. Treat as optional compatibility, not core failures.
- **Kiro** `Manual` trigger in `validKiroTriggers` test map is removed in current docs (test cleanup); `.kiro/steering/lazyai.md` could be emitted with `inclusion: always` (optional — `AGENTS.md` already works).

---

## 4. Corrections made during lead review (important)

Two sub-agent findings were **wrong** and were corrected by independently reading the authoritative docs:

1. **Kiro `.kiro/agents/` and `.kiro/prompts/` are REAL, current surfaces — NOT undocumented.** The Kiro agent flagged both as HIGH "no official docs / files likely ignored" and rated Kiro **PARTIAL**. It had only audited the Kiro **IDE** docs tree (`kiro.dev/docs/`) and missed the entire **Kiro CLI** tree:
   - `kiro.dev/docs/cli/v3/agent-config` (updated 2026-06-17): *"Write a **Markdown file** with your system prompt as the body... Configuration lives in YAML frontmatter."* Locations: `.kiro/agents/` (workspace), `~/.kiro/agents/` (user). JSON is "equivalent."
   - `kiro.dev/docs/cli/chat/manage-prompts`: the Kiro CLI prompt manager stores reusable prompts in `.kiro/prompts/`, referenced via `@name`.
   → LazyAI's `.kiro/agents/<name>.md` and `.kiro/prompts/<name>.md` are **aligned**. Kiro verdict revised to **MOSTLY ALIGNED**. (Residual: MCP `type` field M2 + frontmatter richness.)
   *Lesson: LazyAI's Kiro target is the **Kiro CLI**, so the `kiro.dev/docs/cli/` tree is authoritative — not the IDE docs.*

2. **OpenCode `.opencode/agents/` (plural) is correct.** A pre-check suspected the documented dir was singular `agent/`; the live docs (`opencode.ai/docs/agents`) place markdown agents in `.opencode/agents/` and `~/.config/opencode/agents/` (plural). Adapter is aligned.

Also reconciled: the **internal compliance matrix is stale in spots** but code is correct — e.g. Pi `.pi/settings.json` *is* emitted (`writePiSettings`) with `GlobalConfig:true` (#532), contradicting older matrix prose; the `capability-matrix.md` (2026-06-23) already reflects the fix.

---

## 5. Recommended action list (by priority)

1. **M1 (Antigravity MCP):** make MCP serialization target-aware (Gemini CLI `url`/`settings.json` vs Antigravity-desktop `serverUrl`/`mcp_config.json`), or document the desktop-only scope. Highest user-visible risk (Gemini CLI users get non-functional MCP).
2. **M2 (Kiro MCP `type`), M4 (Copilot CLI `sse`→`http`):** small serializer fixes; both are dedicated `toXMcp` functions.
3. **M3 (Copilot chatmodes→agents):** library asset migration; chat modes are deprecating in VS Code.
4. **M5/M6 (Claude global hooks/MCP):** scope-aware settings + fix misleading internal doc; add warnings.
5. **M7 (OpenCode mode frontmatter), M8 (OMP root path):** one-line-ish library/mapping fixes.
6. **Low cluster:** decide policy on agent-frontmatter richness (map host-specific fields vs intentionally minimal), strip non-standard Copilot SKILL.md fields, Kiro test cleanup.
7. **Process:** add the corrections in §4 to the per-tool research notes so future audits check the **Kiro CLI** docs tree and Gemini-CLI-vs-Antigravity-desktop MCP split.

---

## 6. Sources

Per-tool official sources (with retrieval method + date) are listed in each `docs/adapters/snapshots/compliance-audit-2026-06-25-<tool>.md`. Primary references used:
`opencode.ai/docs/{agents,config,mcp-servers,skills,commands,plugins}` + `opencode.ai/config.json`; `code.claude.com/docs/en/{sub-agents,skills,hooks,mcp,settings}`; `docs.github.com/copilot/...` + `code.visualstudio.com/docs/agent-customization/*`; `pi.dev/docs/latest/*` + `skill://pi-agent-specialist`; `omp://` (context-files, config-usage, mcp-config, task-agent-discovery, skills, slash-command-internals); `raw.githubusercontent.com/google-gemini/gemini-cli/main/schemas/settings.schema.json` + `docs/hooks/*` + `skill://antigravity-platform`; `kiro.dev/docs/cli/v3/{agent-config,hooks,permissions}`, `kiro.dev/docs/cli/chat/manage-prompts`, `kiro.dev/docs/{mcp/configuration,skills,steering}`.
