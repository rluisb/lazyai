# Research: Antigravity Broad Tool-Capability Treatment

**Issue:** #575
**Epic:** #568
**Date:** 2026-06-29
**Status:** research
**Decision recorded:** Not skills-only — cover every Antigravity surface that supports or mentions tool usage.

---

## Problem Statement

Issue #575 asked whether the Antigravity adapter should remain skills-only (Option A) or emit subagents via the `define_subagent` enable-flag model (Option B). The user has since expanded the scope: the correct stance is **not skills-only**, and the plan must address every Antigravity surface that supports or mentions tool usage — agents/subagents, skills, workflows, hooks, and commands.

The adapter currently emits infrastructure only (settings, hooks, skills, rules). No agent or subagent definitions are emitted.

---

## Current Adapter State

**File:** `packages/cli/internal/adapter/antigravity.go`

What `Install()` emits today:

| Output | Path | Source |
|---|---|---|
| Settings | `.gemini/settings.json` | `antigravity/settings.json` (library asset) merged via configmerge |
| Hook scripts | `.gemini/hooks/lazyai/*.sh` | `antigravity/hooks/` (library assets) |
| Hook config | `.agents/hooks.json` or `~/.gemini/config/hooks.json` (global) | `antigravity/hooks.json` (library asset) |
| Skills | `.agents/skills/<name>/SKILL.md` or `~/.gemini/config/skills/<name>/SKILL.md` (global) | `packages/cli/library/skills/*.md` via `CopyLibraryDirectory` |
| Rules bridge | `.agents/rules/lazyai.md` | Inline generated — imports `@/AGENTS.md` for IDE |

What `CompileMCP()` emits (separate phase):

| Output | Path |
|---|---|
| MCP config (native) | `~/.gemini/config/mcp_config.json` (`serverUrl` field) |

**Not emitted:** agent/subagent definitions, workflow definitions, commands, plugins, sidecars.

---

## Antigravity Tool-Capability Model

Source: `docs/ai-cli-tools/tool-systems/antigravity.md` (verified 2026-06-29).

### Built-in tools (fixed set, not togglable per agent except via enable flags)

| Category | Tool names |
|---|---|
| Files/dirs | `view_file`, `write_to_file`, `replace_file_content`, `multi_replace_file_content`, `list_dir`, `find_by_name` |
| Search/research | `grep_search`, `search_web`, `read_url_content` |
| System/exec | `run_command`, `manage_task`, `schedule`, `list_permissions`, `ask_permission` |
| Agent collaboration | `invoke_subagent`, `define_subagent`, `send_message`, `manage_subagents` |
| Interaction/media | `ask_question`, `generate_image` |

### Permissions engine (governs every sensitive op)

Evaluation order: **Deny > Ask > Allow** (strict). Every action is `action(target)`:

| Action | Scope |
|---|---|
| `read_file(path)` | File/dir read ops |
| `write_file(path)` | Write ops (implicitly grants read) |
| `read_url(domain)` | URL fetch |
| `execute_url(domain)` | Browser actuation |
| `command(prefix/regex/*)` | Shell execution |
| `mcp(server/tool)` | MCP tool access |
| `unsandboxed(prefix/regex/*)` | Outside sandbox |

### Subagent enable flags (per `define_subagent`)

| Flag | What it unlocks |
|---|---|
| `enable_mcp_tools` | MCP server access (`mcp(server/tool)` permission) |
| `enable_write_tools` | Write-tier ops (`write_to_file`, `replace_file_content`, `multi_replace_file_content`) |
| `enable_subagent_tools` | Spawning nested subagents (`invoke_subagent`) |

These flags are **capability gates**, not tool-name allowlists. A read-only agent omits `enable_write_tools`; a leaf agent omits `enable_subagent_tools`.

### Hook events and tool matchers

`hooks.json` event structure:

| Event | Has matcher | Tool names in matchers |
|---|---|---|
| `PreToolUse` | Yes | Any built-in tool name (e.g., `run_command`, `write_to_file`, `view_file`) |
| `PostToolUse` | Yes | Same |
| `PreInvocation` | No | n/a — runs before model call |
| `PostInvocation` | No | n/a — runs after tool calls |
| `Stop` | No | n/a — fires on termination |

`settings.json` (Gemini CLI OSS) uses **different event names**: `BeforeTool` / `AfterAgent`. Both files are currently written by LazyAI (two-file strategy).

---

## Per-Surface Analysis

### Surface 1: Agents / Subagents

**Native mechanism:** `define_subagent` built-in tool (runtime call, not a static file).

**Canonical → enable-flag mapping:**

| Canonical capability (from #569) | `enable_write_tools` | `enable_mcp_tools` | `enable_subagent_tools` |
|---|---|---|---|
| read-only (researcher, reviewer, evidence-verifier) | ❌ omit | ✅ if MCP needed | ❌ omit |
| read-write (implementer, deployer) | ✅ | ✅ | optional |
| orchestrator (planner, guide, responder) | ✅ | ✅ | ✅ |

**Current gap:** `antigravity.go` emits no agent/subagent definitions. There is no static `.agents/agents/` discovery path in Antigravity (unlike Claude Code's `.claude/agents/` or OpenCode's agents frontmatter). Subagent creation is programmatic via `define_subagent` at runtime.

**Realistic emission options:**
1. **Rules-based blueprint**: Emit `.agents/rules/lazyai-subagents.md` that documents `define_subagent` call patterns per canonical role with correct enable flags — acts as an instruction to the orchestrator.
2. **Skill-as-orchestrator**: Emit orchestrator skills that programmatically call `define_subagent` with the correct flags derived from canonical agent descriptions.
3. **Plugin agents slot**: If Antigravity ever adds a static agent-definition slot in `plugin.json`, emit there. Currently the plugin manifest fields are: `skills`, `rules`, `mcp_config.json`, `hooks.json`, `sidecars` — no agents slot.

Options 1 and 2 are implementable without a native static file format. Option 3 requires an upstream format change.

**Blocker:** Canonical tool-capability model (#569) must define the `tools`/`readonly` field before any enable-flag mapping can be read from canonical agent frontmatter.

### Surface 2: Skills

**Native mechanism:** `SKILL.md` with YAML frontmatter (`description` required, `name` optional). Already emitted.

**Current gap:** LazyAI emits skills verbatim from `packages/cli/library/skills/*.md`. The SKILL.md frontmatter format does not include tool-capability fields. Skills have no `enable_write_tools` equivalent.

**Analysis:** Antigravity skills are context-injection mechanisms (instructions fed to the agent), not tool-restriction mechanisms. Tool capability for skills is enforced by the permissions engine and the wrapping agent's `define_subagent` flags — not by the skill file itself. Therefore:
- The SKILL.md format change needed is **documentation only**: skill descriptions should note whether the skill is read-only to help the orchestrator apply correct `define_subagent` flags when spawning skill-executing subagents.
- No structural change to SKILL.md emission is required to implement capability; the capability is enforced at the subagent-invocation level.

### Surface 3: Workflows

**Native mechanism:** None. Antigravity has no static workflow file format. There is no `.agents/workflows/` discovery path.

**Closest analog mechanisms:**
- `PreInvocation` hook event: Inject steps before every model call. Can express "before this session's workflow step, validate state".
- `PostInvocation` hook event: Inject steps after tool calls. Can express "after tool calls, run gate check".
- `Stop` hook event (already emitted as `lazyai-objective-workflow-gate`): Fires on termination.
- Orchestrator skills: Library workflow `.md` files (`packages/cli/library/workflows/*.md`) could be emitted as skills so the Antigravity orchestrator can load them as instructions.

**Current gap:** LazyAI library workflows (verified-research.md, bugfix.md, feature.md, etc.) are not emitted for Antigravity. They are emitted for Claude Code (`.claude/commands/`) and OpenCode (`.opencode/workflows/`).

**Assessment:** The most faithful Antigravity representation is to emit library workflows as **orchestrator skills** under `.agents/skills/workflow-<name>/SKILL.md`. This is different from the hook-injection approach and doesn't require upstream format changes.

**Tool-usage relevance:** Workflow skills that include shell steps imply `run_command` access; write-step workflows imply `enable_write_tools`. This mapping must be captured so the orchestrator knows to enable the right flags when invoking a workflow-executing subagent.

### Surface 4: Hooks

**Native mechanism:** `hooks.json` (`.agents/hooks.json` / `~/.gemini/config/hooks.json`). Already emitted.

**Current hook emission:**
- `lazyai-block-destructive-shell`: `PreToolUse` matcher on `run_command` — blocks destructive shell commands.
- `lazyai-objective-workflow-gate`: `Stop` handler — objective/workflow gate.

`settings.json` additionally wires the same hooks under `BeforeTool`/`AfterAgent` for Gemini CLI OSS users.

**Tool-capability gap:** Current hook config covers only `run_command` in `PreToolUse`. There is no coverage for:
- Write-tool blocking (`write_to_file`, `replace_file_content`, `multi_replace_file_content`) — relevant for enforcing read-only agent constraints.
- MCP tool gating (`mcp(server/tool)` via permissions engine) — not currently wired in hooks.json.
- `PreInvocation`/`PostInvocation` hooks for workflow step injection — unused.

**Planned additions:** Hook matchers for write-tier tools (to gate read-only subagents) and `PreInvocation` slot for workflow step injection.

### Surface 5: Commands

**Native mechanism:** None. Antigravity has no "commands" surface (no `/slash-command` registration, no `.agents/commands/` discovery path, no quickactions manifest field). The permissions engine and skills cover all extensibility.

**Closest analog:** The Antigravity CLI itself has a `/mcp` overlay and interactive prompts, but these are user-driven, not LazyAI-emittable. SDK `register_tool` API allows Python-level custom tools but requires code.

**Assessment:** Commands are N/A for Antigravity. Document in the matrix capability column as `—` (not applicable, not a gap).

---

## Capability → Antigravity Native Mapping (consolidated)

| Canonical capability | Antigravity mechanism | LazyAI emission |
|---|---|---|
| file read | `view_file`, `list_dir`, `find_by_name` (always available) | no explicit gate needed |
| file write/edit | `write_to_file`, `replace_file_content` (`enable_write_tools`) | omit flag for read-only roles |
| shell/exec | `run_command` (`command(*)` permission) | `PreToolUse` hook to gate |
| search/grep | `grep_search`, `find_by_name` (always available) | no gate needed |
| web fetch/search | `search_web`, `read_url_content` (`read_url(*)` permission) | permissions engine |
| MCP tools | `mcp(server/tool)` permission (`enable_mcp_tools`) | flag when MCP servers present |
| subagent spawn | `invoke_subagent` (`enable_subagent_tools`) | flag for orchestrator roles only |
| read-only role | omit `enable_write_tools`, `PreToolUse` deny on write tools | rules-blueprint + hook expansion |
| workflow | no native surface | emit as orchestrator skills |
| commands | no native surface | N/A |

---

## Affected Files

| File | Change type |
|---|---|
| `packages/cli/internal/adapter/antigravity.go` | New emission paths (subagent rules, workflow skills, hook expansion) |
| `packages/cli/library/antigravity/hooks.json` | Add write-tool matchers and `PreInvocation` slot |
| `packages/cli/library/antigravity/settings.json` | Align `BeforeTool` matchers with hooks.json additions |
| `packages/cli/library/antigravity/` | New: `subagents-blueprint.md` rule file; workflow skills |
| `docs/ai-cli-tools/tool-systems/antigravity.md` | Update "what LazyAI emits" section to reflect new surfaces |
| `docs/ai-cli-tools/antigravity.md` | Update generated structure and concepts table |
| `docs/ai-cli-tools/tool-systems/agent-tools-matrix.md` | Update Antigravity gap column to "planned" then "closed" |

---

## Blockers and Dependencies

| Dependency | Issue | Status | Impact |
|---|---|---|---|
| Canonical tool-capability model | #569 | Open (P2-high) | Implementation of enable-flag mapping from canonical agent frontmatter is blocked until #569 defines the `tools`/`readonly` field. Shape design can proceed. |
| Foundation PR (tool-systems docs + matrix) | #577 | Merged per assignment context | Research baseline (tool-system docs) available. |

---

## Open Questions

1. **Subagent file format**: Antigravity has no static agent definition file. Is emitting a rules-based subagent blueprint (Option 1 above) the accepted approach, or should LazyAI generate Python SDK scaffold files? SDK files would be outside the adapter's current write pattern (non-JSON/non-Markdown).
2. **Write-tool PreToolUse gate**: Should the hook block writes for a specific agent identity (no agent-identity field in hook context) or block writes globally and rely on the permissions engine for allow-listing? Hook matchers have no "agent name" dimension.
3. **Workflow skill naming**: Should workflow skills be prefixed `workflow-` to avoid collision with existing capability skills?
4. **settings.json BeforeTool matcher naming**: The existing `settings.json` uses `run_shell_command` as the matcher while `hooks.json` uses `run_command`. These are different tool names for different surfaces. Any new matchers must maintain this duality.

---

## Sources

- `packages/cli/internal/adapter/antigravity.go` (implementation ground truth)
- `packages/cli/library/antigravity/hooks.json` (current hook template)
- `packages/cli/library/antigravity/settings.json` (current settings template)
- `docs/ai-cli-tools/tool-systems/antigravity.md` (verified 2026-06-29)
- `docs/ai-cli-tools/tool-systems/agent-tools-matrix.md` (matrix row: Antigravity)
- `specs/031-cross-cli-agent-tools-alignment/research.md` (upstream research)
- Issue #575 body
- Issue #569 body (blocker)
