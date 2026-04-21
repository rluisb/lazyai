# Research — 013: GitHub Copilot Deep Setup (VS Code + Standalone CLI)

**Date:** 2026-04-20
**Author:** Ricardo (with Scout agent)
**Phase:** Research (R of RPI) — awaiting HUMAN GATE before Plan

---

## 1. Goal

Make `ai-setup` produce a GitHub-Copilot configuration that is **100 % conformant** with *both* Copilot surfaces:

1. **VS Code GitHub Copilot extension** — repo-scoped files under `.github/` and `.vscode/`.
2. **Standalone `@github/copilot` CLI** — user-scoped `~/.copilot/` layout **plus** repo-scoped `.github/agents/` that the CLI also reads.

The existence of the standalone CLI is new information vs. the 2025 adapter assumption that "Copilot has no global scope." It does. Scope support for (Copilot × global) must be reconsidered.

Where a deterministic, scriptable bootstrap exists we use it; where the CLI is interactive-only (most of this surface), fall back to direct-write — mirroring the pattern specs 011/012 locked in.

Workspace is an `ai-setup` concept (no upstream equivalent). It is treated like project-scope but rooted at a user-chosen directory; still respects Copilot's file-layout contract.

---

## 2. Authoritative Copilot layout(s)

### 2.1 VS Code GitHub Copilot (repo/project level)

Sources: <https://code.visualstudio.com/docs/copilot/copilot-customization>, <https://code.visualstudio.com/docs/copilot/chat/mcp-servers>, <https://code.visualstudio.com/docs/copilot/customization/prompt-files>, `~/.copilot/pkg/universal/0.0.373/index.js` string table.

```
<project>/
├── .github/
│   ├── copilot-instructions.md          ← always-on repo instructions (primary)
│   ├── AGENTS.md                        ← fallback (VS Code reads it)
│   ├── CLAUDE.md                        ← fallback (VS Code reads it too)
│   ├── instructions/
│   │   └── <name>.instructions.md       ← targeted, glob-gated
│   ├── prompts/
│   │   └── <name>.prompt.md             ← slash-invokable prompt templates
│   ├── chatmodes/
│   │   └── <name>.chatmode.md           ← named chat modes
│   └── agents/                          ← shared surface with standalone CLI (§2.2)
│       └── <name>.agent.yaml
└── .vscode/
    └── mcp.json                         ← { "servers": { … }, "inputs": [ … ] }
```

**Frontmatter contracts (exact):**

| File                | Required keys            | Optional keys                                                                                              |
|---------------------|--------------------------|------------------------------------------------------------------------------------------------------------|
| `*.instructions.md` | — (frontmatter optional) | `applyTo` (glob, comma-sep or array — e.g. `"**/*.ts"`), `description`                                     |
| `*.prompt.md`       | —                        | `description`, `name`, `argument-hint`, `agent` (`ask`\|`agent`\|`plan`\|custom), `model`, `tools` (array) |
| `*.chatmode.md`     | `description`            | `tools` (array), `model`                                                                                   |

**`.vscode/mcp.json` schema** — top-level key is **`servers`** (confirmed by docs and by our own `toCopilotMcp` at `internal/adapter/mcp_compiler.go:363`). Server entries accept `type` (`http`\|`stdio`), `command`, `args`, `env`, `url`, `headers`, `sandboxEnabled`, `sandbox`. `"inputs"` is a sibling top-level key for secret prompts (VS Code prompts the user on first use).

**Settings toggles (user- or workspace-level `settings.json`):**

- `github.copilot.chat.codeGeneration.useInstructionFiles` (enables `.github/copilot-instructions.md`)
- `chat.promptFiles`, `chat.promptFilesLocations`
- `chat.instructionsFilesLocations`, `chat.modeFilesLocations`
- `chat.useCustomizationsInParentRepositories` (monorepo parent discovery)

### 2.2 Standalone `copilot` CLI

Sources: live binary at `~/.copilot/pkg/universal/0.0.373/npm-loader.js --help`, source inspection of minified `index.js`, `~/.copilot/pkg/universal/0.0.373/definitions/*.agent.yaml`, <https://docs.github.com/copilot/how-tos/use-copilot-agents/use-copilot-cli>.

**Root layout (`~/.copilot/`, overridable by `$XDG_CONFIG_HOME` or `$COPILOT_HOME`):**

```
~/.copilot/
├── config.json                  ← UI/UX + trusted_folders + auth state (schema §2.3)
├── mcp-config.json              ← MCP servers (schema §2.4)
├── command-history-state.json   ← CLI-managed
├── logs/                        ← CLI-managed
├── pkg/universal/<ver>/         ← CLI self-install (do not touch)
│   └── definitions/*.agent.yaml ← built-in agents: plan, task, explore, code-review
├── copilot-instructions.md      ← optional user-level custom instructions (per string table)
├── skills/                      ← user-level skills (per string table; /skills interactive cmd)
└── agents/                      ← USER-AUTHORED CUSTOM AGENTS (confirmed)
    └── <name>.agent.yaml
```

**Custom-agent discovery — confirmed via source (`index.js`):**

```js
function bf(){ let t=process.env.XDG_CONFIG_HOME, e=".copilot";
  return t ? tv.join(t,e) : tv.join(os.homedir(),e) }

function hOn(t,e,n,r){ …
  let { agents:a } = await x2e(r,n,t, $Ll.join(bf(),"agents"), l, !1); … }

function x2e(t,e,n,r,l,a=!1){
  let o = await eTn(r);                                   // ~/.copilot/agents
  if (l===null) return {agents:o.agents, warnings:o.warnings};
  let s = await eTn(nTn(l.path, ".github","agents"));     // <gitRoot>/.github/agents
  … // plus remote org/enterprise agents via GitHub API
}
```

So the three discovery layers — **user (`~/.copilot/agents/`) → repo (`<gitRoot>/.github/agents/`) → org (GitHub API)** — match the public docs' stated precedence exactly (docs list user as highest priority on name conflict).

**Agent YAML schema** — confirmed from `~/.copilot/pkg/universal/0.0.373/definitions/plan.agent.yaml`:

```yaml
name: plan                     # required, lowercase id
displayName: Plan Agent
description: >
  …
model: claude-sonnet-4.5       # optional; any value from --help choices
tools:
  - "*"                        # or an explicit list of tool ids / tool-set patterns
promptParts:                   # optional bag of boolean toggles
  includeAISafety: true
  includeToolInstructions: true
  includeParallelToolCalling: true
  includeCustomAgentInstructions: false
prompt: |                      # the system prompt
  You are …
```

**Custom instructions ("AGENTS.md and related files"):** the string table in `index.js` enumerates the exact set Copilot CLI reads:

- `AGENTS.md` (local, plus "AGENTS.md (nested)" — walks up the tree)
- `CLAUDE.md`
- `.github/copilot-instructions.md`
- `copilot-instructions.md`
- `.github/instructions/` (the whole targeted-instructions dir from the VS Code surface)
- `~/.copilot/copilot-instructions.md` (user-level)

`COPILOT_CUSTOM_INSTRUCTIONS_DIRS` env var adds extra search directories. `--no-custom-instructions` suppresses them all.

### 2.3 `~/.copilot/config.json`

User-editable keys (from `copilot help config`): `auto_update`, `banner`, `beep`, `custom_agents.default_local_only`, `log_level`, `model`, `parallel_tool_execution`, `render_markdown`, `screen_reader`, `stream`, `theme`, `trusted_folders`, `allowed_urls`, `denied_urls`. Schema is flat JSON. Safe to deep-merge.

### 2.4 `~/.copilot/mcp-config.json`

Confirmed (from zod schema in `index.js`):

```jsonc
{
  "mcpServers": {
    "<name>": {                         // [A-Za-z0-9_-]+, non-empty
      // Local (stdio) variant:
      "type": "stdio" | "local",        // optional; default stdio
      "command": "<bin>",               // required
      "args": ["…"],
      "env": { "KEY": "VAL" },          // optional
      "tools": ["*"] | ["tool_id", …],  // optional
      "timeout": 30000,                 // optional, positive int

      // OR Remote (http|sse) variant:
      "type": "http" | "sse",
      "url": "<url>",                   // required for http/sse
      "headers": { "Authorization": "Bearer …" },
      "tools": ["*"]
    }
  }
}
```

The `--additional-mcp-config <json|@file>` flag **augments** this file for the session (can be repeated). Built-in default: a `github` server on `https://api.githubcopilot.com/mcp`, tools gated by CLI allowlist unless `--enable-all-github-mcp-tools`.

**No project-scope `mcp-config.json`** — the CLI reads only `~/.copilot/mcp-config.json` plus per-session `--additional-mcp-config`. There is no `.copilot/mcp-config.json` discovery.

### 2.5 Project-scope dir for the standalone CLI?

**Answer: essentially no.** The CLI reads from the project tree only three things:

- `.github/agents/` (custom agents)
- `.github/copilot-instructions.md`, `.github/instructions/`, `AGENTS.md`, `CLAUDE.md`, `copilot-instructions.md` (custom instructions)
- Plus whatever the user puts under `COPILOT_CUSTOM_INSTRUCTIONS_DIRS`.

There is no `.copilot/` project dir. The standalone CLI **reuses the same `.github/` layout** as the VS Code extension. That convergence is the key architectural insight for this spec.

---

## 3. CLI surface available for non-interactive orchestration

From `~/.copilot/pkg/universal/0.0.373/npm-loader.js --help` (captured 2026-04-20):

| Affordance | Mode | Verdict for ai-setup |
|---|---|---|
| `copilot -p "<prompt>" --allow-all-tools [--agent <name>] [--model …]` | **Scriptable** | Use for validation — e.g. confirm a custom agent parses. |
| `copilot --additional-mcp-config <json-or-@file>` | **Scriptable** | Can inject MCP servers per-session, but **does not persist** — useless for install. |
| `copilot --add-dir <dir>` | **Scriptable** | Adds trusted dir for the session only. |
| `/mcp [show\|add\|edit\|delete\|disable\|enable]` | **Interactive only** | Cannot drive from ai-setup. |
| `/skills [list\|info\|add\|remove\|reload]` | **Interactive only** | Cannot drive. |
| `/agent` | **Interactive only** | Browser UI. |
| `copilot help config\|commands\|environment\|logging\|permissions` | Read-only | Useful for user docs; not for install. |
| No `copilot init`, no `copilot mcp add <json>` at the CLI top-level | — | **Hard limit**: no subcommand exists to *write* an MCP server into `mcp-config.json` non-interactively. |

**Consequence:** unlike Claude (which gave us `claude mcp add-json --scope …`), Copilot has **no scriptable MCP writer**. If we want to install MCP servers at user scope, we must **write `~/.copilot/mcp-config.json` directly** (deep-merge with backup, mirroring spec 011's OpenCode pattern). Session-level `--additional-mcp-config` is useless for persistence.

---

## 4. Current `ai-setup` Copilot state (inventory + defects)

### 4.1 Files in scope

| Concern | File |
|---|---|
| Adapter | `internal/adapter/copilot.go` (≈330 LOC) |
| MCP compile | `internal/adapter/mcp_compiler.go:321-364` (`compileCopilotMCP`, `toCopilotMcp`) |
| Scope resolver | `internal/adapter/scope.go:28` (Copilot × global blocked), `:110` (project subdir = `.github`) |
| Global-path flag | `internal/globalpaths/globalpaths.go:71-82` ("Copilot doesn't support file-based global config. Use project scope instead.") |
| Root-memory emitter | `internal/scaffold/root.go:35`; `internal/scaffold/filemap.go:11` (`.github/copilot-instructions.md`) |
| Chatmodes library | `library/chatmodes/architect.chatmode.md`, `reviewer.chatmode.md` |
| Prompts library | `library/prompts/{compact,implement,plan,research,local-example}.md` |
| Instructions library | **Does not exist.** |
| Tests | `internal/adapter/copilot_chatmodes_test.go`, `scope_test.go:29-31,72-74`, `mcp_compiler_test.go:183`, `mcp_compiler_scope_test.go:78` |

### 4.2 What we produce today vs what Copilot expects

| Scope | Current output | Expected | Verdict |
|---|---|---|---|
| project/workspace | `.github/copilot-instructions.md`, `.github/prompts/<name>.prompt.md` (skills reformatted with `mode: agent` frontmatter), `.github/chatmodes/<name>.chatmode.md`, `.vscode/mcp.json` (`{"servers":…}`) | all of the above **plus** `.github/instructions/*.instructions.md`, `.github/agents/*.agent.yaml`, `AGENTS.md` via scaffold root emitter | ⚠ Partial. Missing **instructions** dir entirely. Missing **agents** dir entirely. Chatmodes + prompts + MCP are correct-shape. |
| global | **Nothing written** (`IsScopeSupported` returns false) | `~/.copilot/{config.json (merge), mcp-config.json (merge), agents/*.agent.yaml, copilot-instructions.md}` | ❌ **Coverage gap now that standalone CLI exists.** Assumption "Copilot has no global scope" is stale. |

### 4.3 Confirmed defects

| # | Gap | Severity | Evidence |
|---|---|---|---|
| G1 | Global scope unsupported. `scope.go:28` hard-blocks Copilot × global; standalone CLI now provides a real user-scope surface. | **High** | `scope.go:28`, `globalpaths.go:73-82`, live `~/.copilot/` |
| G2 | No `.github/agents/*.agent.yaml` emission. Key asset the standalone CLI reads from a repo. We emit skills-as-prompts instead, which Copilot CLI doesn't map to agents. | **High** | `copilot.go:43-78` — only `prompts/` and `chatmodes/` populated |
| G3 | No `.github/instructions/*.instructions.md` emission. First-class Copilot customization surface; CLI's "related files" explicitly enumerate it. Zero library assets. | **Medium** | `library/` has no `instructions/` dir; `copilot.go:36-37` creates dir but nothing fills it |
| G4 | MCP at user scope not written. `compileCopilotMCP` short-circuits to project scope (`.vscode/mcp.json` only). | **High** (gated by G1) | `mcp_compiler.go:321` |
| G5 | `.vscode/mcp.json` missing `inputs` support — secret-requiring servers get stringly-typed `env` values that are either committed or empty. | Low | `toCopilotMcp` at `mcp_compiler.go:339-364` emits only `{servers:…}` |
| G6 | `copilot-instructions.md` at user scope never emitted. | Medium (gated by G1) | String table confirms path; no mention in `scaffold/root.go` or `filemap.go` |
| G7 | Skills → `.prompt.md` transform is semantically wrong for the CLI. `EnsureModeAgentFrontmatter` writes `mode: agent` which is VS-Code-only metadata; the standalone CLI ignores `.prompt.md` entirely. | Medium | `copilot.go:262`, `EnsureModeAgentFrontmatter` in `shared.go` |
| G8 | No `AGENTS.md` at user scope. Scaffold emits AGENTS.md only at project scope (`scaffold/root.go:35`). | Low | — |
| G9 | No post-install validation. No `copilot --agent <shipped-name> -p "ping" --allow-all-tools -s` smoke. | Low | — |
| G10 | Frontmatter not schema-validated. Chatmodes files ship with `tools: [...]` but not validated against Copilot's tool id catalog. Same risk for `applyTo` globs we'll ship under G3. | Low | `library/chatmodes/architect.chatmode.md` |
| G11 | No `chatmode.md` ↔ CLI mapping. Chat modes remain VS-Code-only. Expected but worth documenting. | Note | — |

---

## 5. Orchestration strategy — options

### Option A — "Fix VS-Code surface only" (incremental)
Keep global unsupported. Fill the VS-Code gaps: `.github/instructions/*` starter set, `.github/agents/*.agent.yaml` starter set (VS Code will also read these at repo scope). Add `inputs` support in `.vscode/mcp.json`.

- ✅ Small blast radius; no new scope.
- ❌ Leaves the standalone CLI's user-scope surface entirely unowned.
- ❌ Users who run `copilot` from the terminal get zero ai-setup leverage.

### Option B — "Treat standalone CLI as first-class global scope" (recommended)
Lift the Copilot × global block. Add `globalpaths.ResolveCopilotGlobalRoot()` → `~/.copilot`. Write agents, MCP config, user-level `copilot-instructions.md` at global scope; write VS Code repo-layout at project/workspace. Library gets `library/copilot/agents/*.agent.yaml` + `library/copilot/instructions/*.instructions.md` sibling dirs (mirroring spec 011's `library/opencode/`).

- ✅ Unified surface: same agents work in VS Code (repo) and `copilot` CLI (user).
- ✅ Matches the architectural pattern from spec 011 (OpenCode) and 012 (Claude Code).
- ✅ Deep-merge with backup-on-first-touch for `~/.copilot/{config,mcp-config}.json` fits our existing `configmerge` package.
- ⚠ No CLI-assisted MCP writes available — we own the JSON round-trip. Same risk OpenCode already accepts.
- ⚠ Adds scope-parity tests: (copilot, global) was previously always "skip".

### Option C — "Publish as a Copilot-cloud custom agent marketplace entry"
Author agents as org-level GitHub-cloud agents (`.github-private/agents/…`). No local-install needed.

- ✅ Zero scaffolding at client side.
- ❌ Requires a GitHub org we don't have; requires GitHub-cloud auth; invisible to solo users.
- ❌ Orthogonal to ai-setup's value proposition.
- ➜ Park.

**Recommendation: Option B**, phased so Phase 1 is "VS-Code gaps only" (Option-A-shaped) and Phase 2 flips on the global scope.

---

## 6. Open questions for the human gate

1. **Global scope for Copilot.** Confirm we lift the Copilot × global block and treat `~/.copilot/` as a first-class scope. If yes: do we gate behind a probe (detect `copilot` binary on PATH) or write unconditionally?
2. **Agents starter set.** How many starter `library/copilot/agents/*.agent.yaml`? Minimum parity with OpenCode's modes and Claude's agents would be ~3 (e.g. `plan`, `review`, `test`). Or mirror the CLI's built-ins (plan, task, explore, code-review) with ai-setup overrides.
3. **Instructions starter set.** `library/copilot/instructions/` starter files? Candidates: `typescript.instructions.md` (applyTo `**/*.ts`), `go.instructions.md` (applyTo `**/*.go`), `tests.instructions.md` (applyTo `**/*_test.*`). Or ship empty dir + doc.
4. **Skills-as-prompts (G7).** Keep today's "skills → `.prompt.md`" transform (VS Code benefit) or drop it (CLI-invisible, noise)? Or migrate: skills → `.github/agents/<skill>.agent.yaml` so both surfaces see them.
5. **MCP user-scope writing.** Write `~/.copilot/mcp-config.json` via deep-merge when SetupScope=global? Or keep project-scope `.vscode/mcp.json` only and document "the CLI uses its own config; use `/mcp add` interactively"?
6. **`.vscode/mcp.json` inputs.** Emit `"inputs": []` scaffold for secret-requiring servers, or park?
7. **Validation.** Ship a `copilot --agent <x> -p ping --allow-all-tools -s` smoke for each shipped agent? Adds a runtime dependency on the `copilot` binary.
8. **User-scope `copilot-instructions.md` (G6, G8).** Emit one at global scope (with CLAUDE.md-style placeholder template), or leave untouched?
9. **Chatmodes.** Keep VS-Code-only (current) or try to promote chatmode content into agents so the CLI sees them too?
10. **Plugin path (Option C).** Confirm "park" so I don't blend it into the plan.

---

## 7. Proposed scope for Plan phase (preview, not committed)

Pending Option B + reasonable answers to §6:

- **Phase 1 — VS-Code surface parity (no scope change).** Ship `library/copilot/agents/*.agent.yaml` starter set; emit `.github/agents/` at project/workspace. Ship `library/copilot/instructions/*.instructions.md` starter set; emit `.github/instructions/`. Add `inputs` sibling to `.vscode/mcp.json`.
- **Phase 2 — Global scope: lift the block.** Update `IsScopeSupported`, add `ResolveCopilotGlobalRoot`, `globalpaths.IsGlobalInstallSupported` flip for Copilot. Scope-parity tests gain (copilot, global).
- **Phase 3 — Global-scope emitters.** Write `~/.copilot/agents/*.agent.yaml` (library parity with repo), `~/.copilot/mcp-config.json` (deep-merge with backup), `~/.copilot/copilot-instructions.md` (template-filled).
- **Phase 4 — Tests + validation.** Frontmatter-schema tests for `.agent.yaml` (zod-shape parity), `.instructions.md` (applyTo present), `.chatmode.md`. MCP round-trip tests for both `.vscode/mcp.json` and `~/.copilot/mcp-config.json`. Optional `copilot --agent … -p ping` smoke behind a build tag or env guard.
- **Phase 5 — Docs + knowledge map.** Update `specs/KNOWLEDGE_MAP.md` Packages Reference and Pending/Follow-up. Add decision row(s) for global-scope policy.

Each phase leaves the codebase shippable.

---

## 8. Confidence

- **Medium-high** overall.
- **High** on §2.1 (VS Code frontmatter + `.vscode/mcp.json`) — cross-verified against docs and existing `copilot_chatmodes_test.go`.
- **Medium-high** on §2.2 (standalone CLI custom-agent discovery) — source-confirmed via `bf()`/`x2e()` in bundled `index.js`, aligns with public docs.
- **Medium** on §2.4 MCP schema — zod-extracted; `type: "local"` appears as an undocumented alias for `"stdio"` (emit `"stdio"` canonically).
- **Medium** on "related files" precedence — string-table enumeration, not loader-code trace.

---

## 9. References

- VS Code Copilot customization: <https://code.visualstudio.com/docs/copilot/copilot-customization>
- VS Code MCP servers: <https://code.visualstudio.com/docs/copilot/chat/mcp-servers>
- VS Code prompt files: <https://code.visualstudio.com/docs/copilot/customization/prompt-files>
- GitHub Copilot CLI about: <https://docs.github.com/copilot/concepts/agents/about-copilot-cli>
- GitHub Copilot CLI custom agents: <https://docs.github.com/copilot/how-tos/use-copilot-agents/use-copilot-cli>
- GitHub Copilot CLI configuration: <https://docs.github.com/en/copilot/how-tos/copilot-cli/set-up-copilot-cli/configure-copilot-cli>
- Live binary: `~/.copilot/pkg/universal/0.0.373/npm-loader.js --help`, `copilot help config|commands|environment`
- Source inspection (for `bf()`, `hOn()`, `x2e()`, MCP zod schema): `~/.copilot/pkg/universal/0.0.373/index.js`
- Agent YAML sample: `~/.copilot/pkg/universal/0.0.373/definitions/plan.agent.yaml`
- Spec 011 (OpenCode): `specs/011-opencode-deep-setup/`
- Spec 012 (Claude Code): `specs/012-claude-code-deep-setup/`
- Current adapter: `internal/adapter/copilot.go`, `internal/adapter/mcp_compiler.go:321-364`, `internal/adapter/scope.go:28,110`, `internal/globalpaths/globalpaths.go:71-82`

---

## 10. Decisions (from human gate, 2026-04-20)

| #   | Question                                 | Decision                                                                                                                                                                     | Planning implication                                                                                                                                                                                                          |
| --- | ---------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Q1  | Copilot × global scope                   | **Lift block, probe-gated**                                                                                                                                                  | Detect `copilot` on PATH or `~/.copilot/` presence before emitting global files; silent fallback (no-op + single-line warning) if neither. Update `IsScopeSupported`, `globalpaths.IsGlobalInstallSupported`. Scope-parity tests gain (copilot, global). |
| Q2  | Agents starter set                       | **Port ai-setup roles**                                                                                                                                                      | `library/copilot/agents/` ships `planner.agent.yaml`, `builder.agent.yaml`, `scout.agent.yaml`, `reviewer.agent.yaml`, plus `orchestrator.agent.yaml` gated on `EnableServers`. Schema = confirmed zod shape in §2.2.          |
| Q3  | Instructions starter set                 | **Language-specific trio**                                                                                                                                                   | `library/copilot/instructions/` ships `typescript.instructions.md` (applyTo `**/*.{ts,tsx}`), `go.instructions.md` (applyTo `**/*.go`), `tests.instructions.md` (applyTo `**/*_test.*,**/*.test.*,**/*.spec.*`).                |
| Q4  | Skills transform reconciliation (G7)     | **Migrate skills → agents**                                                                                                                                                  | Drop `EnsureModeAgentFrontmatter` / `.prompt.md` transform path for skills. Emit each selected skill as `.github/agents/<skill>.agent.yaml` (and at global scope, `~/.copilot/agents/<skill>.agent.yaml`). One write path, two surfaces. Prompts (`.prompt.md`) continue to be emitted for the `library/prompts/` set only. |
| Q5  | MCP user-scope write policy              | **Deep-merge `~/.copilot/mcp-config.json`**                                                                                                                                  | `compileCopilotMCP` at global scope writes `~/.copilot/mcp-config.json` with `configmerge` + backup-on-first-touch. Managed servers win on key collision; user-authored preserved. At project/workspace scope, keep `.vscode/mcp.json` behavior. |
| Q6  | `.vscode/mcp.json` inputs (G5)           | **Scaffold `inputs` when env placeholders detected**                                                                                                                         | When an MCP server's `env` contains `${VAR}`-style placeholders, emit matching `inputs: [{type:'promptString',id:'VAR',password:true}]` entries. Otherwise omit the `inputs` key.                                              |
| Q7  | Post-install validation                  | **Smoke test each shipped agent**                                                                                                                                            | Probe-gated on `copilot` binary presence. For each shipped `.agent.yaml`, run `copilot --agent <name> -p "ai-setup validation ping" --allow-all-tools -s` with a short timeout. Non-zero exit = warning, not error. Follows spec 011's post-install pattern. |
| Q8  | User-scope `copilot-instructions.md` (G6, G8) | **Template-filled, first-install only**                                                                                                                                  | Emit `~/.copilot/copilot-instructions.md` only when absent (mirror spec 010's hybrid CLAUDE.md fill). Re-runs leave it alone. Placeholders include org/team fill-in markers + mechanical auto-infer of project metadata where applicable. |
| Q9  | `.chatmode.md` fate                      | **Keep VS-Code-only, document it**                                                                                                                                           | No code change for chatmodes. Add a README note (or section in `specs/KNOWLEDGE_MAP.md`) stating chatmodes are a VS Code concept; standalone CLI users should use agents instead.                                              |
| Q10 | Option C (cloud/marketplace)             | **Park**                                                                                                                                                                     | Add a `Pending / Follow-up` line to `specs/KNOWLEDGE_MAP.md` for a future spec. Spec 013 = local files only.                                                                                                                  |

---

## 11. Next (after HUMAN GATE)

Move to **Plan** phase. Plan will phase the work as:

- **Phase 1 — Library assets (no behavior change yet).** Ship `library/copilot/agents/*.agent.yaml` (Q2) and `library/copilot/instructions/*.instructions.md` (Q3). Add schema unit tests.
- **Phase 2 — VS-Code surface parity at project/workspace scope.** Emit `.github/agents/` (from library + migrated skills per Q4) and `.github/instructions/`. Update `.vscode/mcp.json` emitter to scaffold `inputs` on env-placeholder detection (Q6).
- **Phase 3 — Lift Copilot × global block (Q1).** Update `IsScopeSupported`, add `ResolveCopilotGlobalRoot` in `globalpaths`, probe for `copilot` binary / `~/.copilot/` dir, silent no-op when neither present.
- **Phase 4 — Global-scope emitters (Q5, Q8).** Write `~/.copilot/agents/*.agent.yaml`, `~/.copilot/mcp-config.json` via deep-merge, `~/.copilot/copilot-instructions.md` template-filled first-install.
- **Phase 5 — Post-install validation (Q7).** Add `CopilotCLIRunner` interface + `LookupCopilotBinary()` (mirror spec 012's `ClaudeCLIRunner`). Run per-agent smoke test, surface warnings.
- **Phase 6 — Tests + docs.** Scope-parity tests including (copilot, global). Frontmatter-schema tests. MCP round-trip tests for both `.vscode/mcp.json` and `~/.copilot/mcp-config.json`. Integration test for probe-gated global behavior. Update `specs/KNOWLEDGE_MAP.md` decisions + packages reference + pending/follow-up.

Each phase leaves the codebase shippable.
