---
title: Agent Tools — Cross-CLI Compatibility Matrix
summary: How a canonical agent's tool capability maps onto each target CLI's native per-agent tool model, and what LazyAI emits today.
status: verified
verified_on: 2026-06-29
scope: LazyAI agent emission vs upstream per-agent tool models
applies_to: [opencode, claude-code, copilot, pi, omp, kiro, antigravity]
---

# Agent Tools — Cross-CLI Compatibility Matrix

This page maps **agent tool capability** across the seven LazyAI targets: how each CLI lets an agent/subagent restrict its tools, the native tool names/casing, and **what LazyAI emits today** vs. the correct shape. It is the reference for the agent-tools alignment work tracked in the epic linked at the bottom.

## Root finding

Canonical agents (`packages/cli/library/canonical/agents/*.md`) express capability via `mode:` and `role:` only — there is **no machine-readable per-tool capability** (no `tools:` allowlist, no `readonly` flag). Some canonical sources even contradict themselves (e.g. `researcher.md` declares `mode: all` while its description says "read-only codebase explorer"). Because no capability signal exists, every adapter improvises, and read-only roles (researcher, reviewer) are emitted **unrestricted** on most targets.

## Per-agent tool model by target

| Target | Native per-agent restriction | Built-in tool names (casing) | Allow/deny semantics |
|---|---|---|---|
| **Claude Code** | `tools:` (allow) + `disallowedTools:` (deny) in agent `.md` frontmatter | Capitalized: `Read`, `Write`, `Edit`, `Bash`, `Grep`, `Glob`, `Agent`, `WebFetch`, `WebSearch`, `Skill` | allowlist; `disallowedTools` wins |
| **OpenCode** | `tools:{bash:true,…}` gate map + `permission:` + `mode:` in agent frontmatter / `opencode.json` `agent.<name>` | lowercase built-in gates: `bash`, `edit`, `write`, plus `permission` keys | per-tool boolean + `ask`/`allow`/`deny` |
| **Copilot** | `tools:` list in `.agent.md` | lowercase set: `read`, `search`, `edit`, `shell` (+ `*`) | allowlist |
| **Pi** | **none** (CLI flags `--tools`/`--exclude-tools`; experimental skill `allowed-tools`) | lowercase: `read`, `bash`, `edit`, `write`, `grep`, `find`, `ls` | n/a per-agent |
| **OMP** | `tools:` (CSV/YAML subset) in `.omp/agents/<name>.md` + `spawns`, `thinkingLevel`, `autoloadSkills`, `read-summarize` | lowercase: `read`, `bash`, `edit`, `write`, `search`, `task`, … | allowlist (subset of built-ins) |
| **Kiro** | `tools` (whitelist) + `allowedTools` (auto-approved) in agent **JSON** | aliases: `fs_read`/`read`, `fs_write`/`write`, `execute_bash`/`shell`, `use_aws`/`aws` | allowlist + auto-approve |
| **Antigravity** | `define_subagent` enable flags (`enable_mcp_tools`, `enable_write_tools`, `enable_subagent_tools`) — not a tool-name list | distinct: `view_file`, `write_to_file`, `replace_file_content`, `run_command`, `grep_search`, `search_web`, `read_url_content`, `invoke_subagent` | capability flags + permissions engine |

## Canonical capability → target mechanism

| Canonical capability | Claude | OpenCode | Copilot | Pi | OMP | Kiro | Antigravity |
|---|---|---|---|---|---|---|---|
| file read | `Read` | (default) | `read` | `read` | `read` | `read`/`fs_read` | `view_file` |
| file write/edit | `Write`/`Edit` | `edit:true` / `permission.edit` | `edit` | `write`/`edit` | `write`/`edit` | `write`/`fs_write` | `write_to_file`/`replace_file_content` (`enable_write_tools`) |
| shell/exec | `Bash` | `bash:true` / `permission.bash` | `shell` | `bash` | `bash` | `execute_bash`/`shell` | `run_command` |
| search/grep | `Grep`/`Glob` | (default) | `search` | `grep`/`find` | `search` | (built-in) | `grep_search`/`find_by_name` |
| web fetch/search | `WebFetch`/`WebSearch` | (built-in) | (n/a) | extension only | `web_search` | (built-in) | `search_web`/`read_url_content` |
| MCP tools | `mcp__<srv>__<tool>` | `.mcp.json` servers | VS Code/CLI MCP | **none** | `mcp__<srv>_<tool>` | `mcpServers`/`autoApprove` | `mcp(server/tool)` (`enable_mcp_tools`) |
| subagent spawn | `Agent` | (built-in) | (n/a) | subagent ext. | `task`/`spawns` | (agent profiles) | `invoke_subagent` (`enable_subagent_tools`) |
| **read-only role** (researcher/reviewer) | `tools:` minus Write/Edit/Bash, or `disallowedTools:[Write,Edit,Bash]` | `permission:{edit:deny,bash:deny}` + `mode` | `tools:["read","search"]` | `--no-builtin-tools` / runtime | `tools:[read,search,...]` | `tools` minus write/exec | omit `enable_write_tools` |

## What LazyAI emits today (gap status)

| Target | Current emission | Respects per-agent tools? | Gap |
|---|---|---|---|
| Claude Code | agent `.md` with **name + description only** (`RewriteAgentForClaudeCode`) | ❌ no `tools`/`disallowedTools` | read-only agents unrestricted |
| OpenCode | canonical agents → description + marker only; only built-in `plan`/`build`/`explore` get `permission` in `opencode.json` | ⚠️ partial (named built-ins only) | canonical agents get no `permission`/`mode`/`tools` |
| Copilot | `.agent.md` with **hardcoded** `tools: ["read","search","edit","shell"]` for every agent; skills `tools:["*"]` | ❌ blanket allowlist | read-only agents get `edit`+`shell` |
| Pi | agents copied; no tools field (Pi has no mechanism) | ✅ correct by design | none (document intentional non-mapping) |
| OMP | `RewriteAgentForOMP` transform: `tools` (OMP allowlist from canonical grants), `thinkingLevel`, `autoloadSkills` (from `skills:`); LazyAI-only fields dropped | ✅ read-only agents restricted (`tools: ["read","search"]`); full-capability agents get OMP equivalents | ✅ closed by #573 |
| Kiro | canonical agents copied verbatim as `.md` | ❌ no `tools`/`allowedTools` | doc says JSON (`.kiro/agents/<name>.json`); spec 030 says `.md` tolerated — reconcile |
| Antigravity | **no agent/subagent files** (skills-only) | n/a | decide & document subagent stance |

## Evidence (file:line)

- Canonical: `packages/cli/library/canonical/agents/researcher.md:5-7` (`tools: [read, search]`); all canonical agents carry `tools:` grants as of #569.
- Claude: `packages/cli/internal/adapter/agent_transform.go` `RewriteAgentForClaudeCode` (name+description only); `claudecode_frontmatter_test.go`.
- OpenCode: `opencode.go:73-101` (hardcoded `plan`/`build`/`explore` permissions), `opencode.go:134-141` + `agent_transform.go` `RewriteAgentForOpenCode` (description-only).
- Copilot: `copilot.go:322` (`tools: ["read","search","edit","shell"]`).
- OMP: `agent_transform.go` `RewriteAgentForOMP`; `omp.go` (transform-based copy via `CopyLibraryDirectoryOption.Transform`); `omp_frontmatter_test.go` (#573).
- Kiro: `kiro.go:39-44`; `docs/ai-cli-tools/tool-systems/kiro.md` (JSON) vs `specs/030-kiro-cli-v3-output-gaps/spec.md:35` (`.md` tolerated, no transform).
- Pi: `pi.go`, `docs/ai-cli-tools/tool-systems/pi.md` (no per-agent mechanism).
- Antigravity: `antigravity.go` (no agent emission).

## Tracking

This matrix backs the cross-CLI agent-tools alignment epic [#568](https://github.com/rluisb/lazyai/issues/568) and its per-target child issues (#569 canonical model, #570 Claude, #571 Copilot, #572 OpenCode, #573 OMP, #574 Kiro, #575 Antigravity). Update the "gap status" column as each child closes.
