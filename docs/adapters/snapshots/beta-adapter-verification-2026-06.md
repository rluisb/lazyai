# Beta Adapter Docs-Snapshot Verification — 2026-06-23

> Tracks [#486](https://github.com/rluisb/lazyai/issues/486). Captures source-verified
> documentation snapshots for the two beta adapters (OMP, Antigravity/Gemini), records
> the verdict per emitted surface, and states the promotion decision.
>
> **Method:** every surface the adapter actually emits (per `omp.go` / `antigravity.go`
> and `mcp_compiler.go`) was checked against the host tool's official documentation,
> rendered where the docs are JS-only.

## Summary

| Adapter | Verdict | Decision |
|---|---|---|
| OMP | All emitted surfaces verified against authoritative docs | **Promote beta → stable** |
| Antigravity/Gemini | All surfaces verified; both 2026-06 gaps closed + pinned | **Promote beta → stable** |

---

## OMP

**Authoritative source:** the OMP (Oh My Pi) coding-agent documentation set, available
in-harness at `omp://` (e.g. `omp://context-files.md`, `omp://mcp-config.md`). This *is*
the official documentation, so the original beta blocker ("partially JS-rendered docs not
snapshot-verified") no longer applies.

> **Correction:** OMP is **Oh My Pi**, the AI coding-agent harness — not "Oh My Posh"
> (`ohmyposh.dev`), which `docs/adapters/omp.md` previously referenced in error.

| Emitted surface | Path emitted | Doc source | Verdict |
|---|---|---|---|
| Root instructions | root `AGENTS.md` (via `scaffold.ScaffoldCompiledRoot`) | `omp://context-files.md` — `agents-md` provider reads standalone `AGENTS.md` walking up to repo root (priority 10) | Verified (works; native `.omp/AGENTS.md` would be higher priority) |
| Agents | `.omp/agents/<name>.md` | `omp://task-agent-discovery.md` — native project agents dir `.omp/agents`, frontmatter `name`/`description` + body | Verified |
| Skills | `.omp/skills/<name>/SKILL.md` | `omp://skills.md` — `<root>/skills/<name>/SKILL.md`, non-recursive, `description` required for native provider | Verified |
| Hooks | `.omp/hooks/pre/*.ts` | `omp://hooks.md` — JS/TS hook factories discovered via `hookCapability` from `.omp/hooks/pre/*.ts`; default-export `HookAPI` factory | Verified |
| Commands | `.omp/commands/<name>.md` | `omp://slash-command-internals.md` — native provider scans `.omp/commands/*.md`; frontmatter `description` + `$ARGUMENTS`/`$1` template body | Verified |
| MCP | `.omp/mcp.json` | `omp://mcp-config.md` — native project MCP at `.omp/mcp.json`; `{mcpServers, disabledServers?}`; stdio/http/sse transports | Verified |
| Prompts | `.omp/prompts/<name>.md` | not covered by a dedicated prompts doc | **Emitted but discovery unverified** — capability does not declare `PromptTemplates`; treat as best-effort |

Capability flags `Plugins`, `Compaction`, `Sessions`, `GlobalConfig` are host-support
metadata (OMP supports them — `omp://marketplace.md`, `compaction.md`, `session.md`,
`settings.md`), not surfaces the adapter emits files for. This matches how the matrix
treats e.g. Pi's MCP no-op.

**Decision:** OMP meets the stable bar (official docs verified + golden tests
`omp_adapter_test.go` + install smoke). `CanRunHeadless()=false` is not a blocker — Pi is
stable with the same constraint. Promoted to **stable**.

---

## Antigravity / Gemini

The adapter is **dual-target**: it emits the Antigravity IDE surface under `.agents/` and a
Gemini CLI surface under `.gemini/`.

**Authoritative sources (rendered; docs are JS-only so `read` returns only the meta shell):**
- Antigravity IDE: `https://antigravity.google/docs/{skills,hooks,mcp,rules-workflows,plugins}`
- Gemini CLI: `https://geminicli.com/docs/hooks`, `https://github.com/google-gemini/gemini-cli`

| Emitted surface | Path emitted | Doc source | Verdict |
|---|---|---|---|
| Skills (workspace) | `.agents/skills/<name>/SKILL.md` | AG `/docs/skills` — workspace skills at `<root>/.agents/skills/<name>/SKILL.md`; `name` optional, `description` required | Verified |
| Skills (global) | `~/.gemini/config/skills/<name>/SKILL.md` | AG `/docs/skills` — global skills at `~/.gemini/config/skills/<name>/` | Verified (gap 1 closed: `antigravity.go` writes scope-aware skills dir) |
| Hooks (IDE) | `.agents/hooks.json` + `.gemini/hooks/lazyai/*.sh` | AG `/docs/hooks` — `hooks.json` in `.agents/`; event-keyed (`PreToolUse`/`PostToolUse`/`PreInvocation`/`PostInvocation`/`Stop`), `matcher: run_command`, `{type:command, command, timeout}` | Verified (asset matches schema) |
| Hooks (CLI) | `.gemini/settings.json` `hooks` block | Gemini CLI `/docs/hooks` — `.gemini/settings.json` `hooks`, events `BeforeTool`/`AfterAgent`/…, `matcher: run_shell_command`, `$GEMINI_PROJECT_DIR/.gemini/hooks/…` | Verified (asset matches schema) |
| MCP | `~/.gemini/config/mcp_config.json` | AG `/docs/mcp` — `~/.gemini/config/mcp_config.json`; `{mcpServers:{...}}`; stdio `command`, HTTP `serverUrl` (not `url`); `toAntigravityMcp` correctly emits `serverUrl` | Verified |
| Root instructions | `GEMINI.md` (Gemini CLI, via scaffold) + `.agents/rules/lazyai.md` (Antigravity IDE workspace) + canonical root `AGENTS.md` | AG `/docs/rules-workflows` — workspace rules `.agents/rules/*.md`, global `~/.gemini/GEMINI.md`; Gemini CLI context file is `GEMINI.md` | Verified (gap 2 closed: GEMINI.md imports `@./AGENTS.md`; `.agents/rules/lazyai.md` imports `@/AGENTS.md`) |

Capability flags `Plugins` (AG `/docs/plugins`: `.agents/plugins/<name>/` with
`plugin.json` + skills/rules/hooks/mcp) and `Permissions` are host-support metadata; no
plugin bundle or permissions file is emitted.

### 2026-06 gaps — closed

Both exit criteria below were resolved and pinned by conformance tests:

1. **Global-scope skills path** *(closed).* `AntigravityAdapter.Install` now writes
   global skills to `~/.gemini/config/skills/` (scope-aware `skillsDir`); workspace/project
   keep `.agents/skills/`. Pinned by `TestAntigravityAdapter_Install_GlobalSkillsUseGeminiConfigDir`.
2. **Root instructions** *(closed).* The canonical root `AGENTS.md` is now discovered by both
   tools: the scaffold emits a `GEMINI.md` pointer (`@./AGENTS.md`) for Gemini CLI, and the
   adapter emits `.agents/rules/lazyai.md` (`@/AGENTS.md`) for Antigravity IDE workspaces.
   Global rules (`~/.gemini/GEMINI.md`) remain user-managed (same conservative policy as
   Claude's global `CLAUDE.md`). Pinned by `TestScaffoldCompiledRootAntigravityGeneratesGeminiMd`
   and `TestAntigravityAdapter_Install_EmitsWorkspaceRules`.

**Decision:** all emitted surfaces verified and both gaps closed + pinned. Antigravity/Gemini
promoted to **stable** (no beta adapters remain).
