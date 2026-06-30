# Research: Claude Code — agents drop tool restrictions (#570)

**Issue:** [#570](https://github.com/rluisb/lazyai/issues/570)
**Epic:** #568 (Cross-CLI Agent-Tools Alignment)
**Blocks:** Blocked until #569 merges (canonical capability model)
**Date:** 2026-06-29
**Status:** research-complete

---

## Problem

`RewriteAgentForClaudeCode` emits Claude subagent `.md` files with **name and description only**.
No `tools:` allowlist and no `disallowedTools:` deny list is emitted.
As a result, read-only canonical agents (researcher, reviewer, evidence-verifier) run with full
tool access on Claude Code — they can `Edit`, `Write`, and `Bash` despite their explicit read-only intent.

---

## Evidence — source code

### `packages/cli/internal/adapter/agent_transform.go` — `RewriteAgentForClaudeCode` (lines 96–126)

```go
// RewriteAgentForClaudeCode transforms a library agent into a Claude Code agent
// file. Output frontmatter contains only name and description; the source body is
// preserved verbatim after the vibe-lab managed marker.
func RewriteAgentForClaudeCode(source []byte, ctx *AdapterContext) ([]byte, error) {
    _ = ctx                             // <-- ctx is unused; no capability read
    fm, body, err := frontmatter.ExtractFrontmatter(source)
    ...
    b.WriteString("---\n")
    b.WriteString("name: ")             // only name
    ...
    b.WriteString("description: ")      // only description
    b.WriteString("\n---\n\n")
    ...
}
```

- `ctx` is explicitly discarded (`_ = ctx`) — no AgentSpec parsed, no tool capability consulted.
- Output frontmatter emits only `name:` and `description:`.
- No `tools:` or `disallowedTools:` key is written regardless of the agent's role.

### `packages/cli/internal/adapter/claudecode_frontmatter_test.go`

- `TestRewriteAgentForClaudeCode_EmitsDescription` (lines 15–40): asserts `description` is preserved.
  Guards against the pre-#208 regression. Does **not** assert `tools`.
- `validateAgentsSchemas` (lines 151–200): requires only `name` and `description`.
  Comments that if `tools:` is present it must be whitespace-separated (spec 012 task 004),
  but there is no assertion that `tools:` is emitted for any agent — the field is purely optional today.
- `TestClaudeCode_InstalledCanonicalAgentsHaveRequiredFields` (lines 269–357): installs
  agents and checks only `name`/`description` — no `tools` assertion.

**Gap confirmed:** no test currently verifies tool restriction on read-only agents.

---

## Evidence — canonical agents (capability source)

All canonical agents under `packages/cli/library/canonical/agents/` declare `mode: all`,
even those whose description and role are explicitly read-only:

| Agent | `role:` | `mode:` | Description |
|---|---|---|---|
| `researcher.md` | `researcher` | `all` | "read-only codebase explorer" — contradicts `mode: all` |
| `reviewer.md` | `reviewer` | `all` | "Read-only." — contradicts `mode: all` |
| `evidence-verifier.md` | `evidence-verifier` | `all` | "Verify claims" — no write need |
| `implementer.md` | `implementer` | `all` | Writes code — full access correct |
| `planner.md` | `planner` | `all` | Planning only — read-heavy but may need write |
| `deployer.md` | `deployer` | `all` | Executes commands — full access correct |
| `guide.md` | `guide` | `all` | Onboarding — read-only in practice |
| `responder.md` | `responder` | `all` | Fast Q&A — read-only in practice |

No canonical agent currently carries a `tools:` allowlist or `readonly:` flag.
This is the gap #569 will fix by introducing a machine-readable capability field.

---

## Evidence — Claude Code upstream tool model

Source: `docs/ai-cli-tools/tool-systems/claude-code.md` (verified 2026-06-29, from code.claude.com).

### Subagent frontmatter schema

Required: `name`, `description`.
Optional: `tools` (allowlist), `disallowedTools` (deny list), `model`, `permissionMode`, `maxTurns`, etc.

```yaml
---
name: researcher
description: "Scout agent — read-only codebase explorer."
tools:
  - Read
  - Grep
  - Glob
  - Agent
disallowedTools:
  - Edit
  - Write
  - Bash
---
```

### Built-in tool names — exact casing

Claude's built-in tools use **PascalCase**. Exact names matter for permission matching.

| Capability | Claude built-in name | Requires permission? |
|---|---|---|
| File read | `Read` | No |
| Grep/search | `Grep` | No |
| Glob listing | `Glob` | No |
| File edit | `Edit` | Yes |
| File write | `Write` | Yes |
| Shell execution | `Bash` | Yes |
| Web fetch | `WebFetch` | Yes |
| Web search | `WebSearch` | Yes |
| Subagent spawn | `Agent` | No |
| Skill invocation | `Skill` | No |
| LSP tools | `LSP` | No |
| Notebook edit | `NotebookEdit` | Yes |

Source: `docs/ai-cli-tools/tool-systems/claude-code.md` — "Built-in tools" table.

### Semantics

- `tools:` is an **allowlist**: only the listed tools are available.
- `disallowedTools:` is a **deny list**: the listed tools are removed from context.
- Both can appear together; `disallowedTools` wins on conflict.
- MCP tools referenced as `mcp__<server>__<tool>`.
- Tool names in `tools`/`disallowedTools` are **bare names** (no parenthesized rule format).

---

## Evidence — cross-CLI matrix

Source: `docs/ai-cli-tools/tool-systems/agent-tools-matrix.md` (verified 2026-06-29).

Canonical capability → Claude target:

| Canonical capability | Claude built-in | Casing |
|---|---|---|
| file read | `Read` | PascalCase |
| file write/edit | `Write`, `Edit` | PascalCase |
| shell/exec | `Bash` | PascalCase |
| search/grep | `Grep`, `Glob` | PascalCase |
| web | `WebFetch`, `WebSearch` | PascalCase |
| subagent spawn | `Agent` | PascalCase |
| skill invocation | `Skill` | PascalCase |

Gap status (matrix line): `Claude Code | agent .md with name + description only | ❌ no tools/disallowedTools`.

---

## Root cause

Two independent gaps that must both be closed:

1. **#569 (prerequisite):** Canonical agents carry no machine-readable capability. `mode:` is uniformly `all`
   even for read-only roles. Without a canonical signal, no adapter can generate correct tool restrictions.

2. **#570 (this issue):** `RewriteAgentForClaudeCode` does not read capability from the agent spec
   and emits no `tools:`/`disallowedTools:` regardless of role. The function signature accepts
   `*AdapterContext` but discards it (`_ = ctx`). It uses `frontmatter.ExtractFrontmatter`
   (generic parse) rather than `frontmatter.ParseAgentSpec` (spec parse that reads tier, role, etc.).

---

## Dependency chain

```
#568 (epic)
├── #569  ← adds canonical tools:/readonly field on all agents (PREREQUISITE for #570)
├── #570  ← translates canonical grants to Claude tools:/disallowedTools (THIS ISSUE)
│          ← depends on #569 for the canonical signal
├── #572  ← OpenCode adapter (shares agent_transform.go — must coordinate)
├── #571  ← Copilot adapter (independent file)
└── #573  ← OMP adapter (independent file)
```

> **Collision note:** `agent_transform.go` is the shared entry point for both #570 (Claude)
> and #572 (OpenCode). These must be implemented sequentially or in separate worktrees
> with coordinated merges to avoid conflicts.

---

## Canonical → Claude tool mapping (draft, pending #569 vocabulary)

The exact token names #569 introduces are unknown until that issue lands. This table is
**speculative** and must be reconciled with #569's actual vocabulary:

| Canonical token (#569, TBD) | Claude `tools:` entry | Notes |
|---|---|---|
| `read` | `Read`, `Grep`, `Glob`, `LSP` | All read-only built-ins |
| `edit` | `Edit`, `Write`, `NotebookEdit` | Destructive write tools |
| `shell` | `Bash` | Shell execution |
| `web` | `WebFetch`, `WebSearch` | Network tools |
| `spawn` | `Agent` | Subagent invocation |
| `skill` | `Skill` | Skill invocation |
| `mcp` | `mcp__<server>__<tool>` | Per-server wildcard |

Read-only agents (researcher, reviewer, evidence-verifier) should emit `disallowedTools: [Edit, Write, Bash]`
or an explicit `tools:` allowlist containing only `[Read, Grep, Glob, LSP, Agent, Skill]`.

---

## Risks

| Risk | Severity | Mitigation |
|---|---|---|
| #569 canonical vocabulary differs from draft mapping above | Medium | Plan defers final mapping until #569 merges; implementation reads #569's exported parser |
| `agent_transform.go` conflict with #572 (OpenCode) | Medium | Sequence or coordinate; both PRs touch `RewriteAgentForClaudeCode` / `RewriteAgentForOpenCode` in the same file |
| Test changes broaden scope of `claudecode_frontmatter_test.go` | Low | Add targeted assertions; do not refactor existing tests |
| Claude upstream tool name changes | Low | `claude-code.md` is `verified_on: 2026-06-29`; recheck on implementation |
| `mode: all` on agents that should be read-only | Low | Fixed by #569; this issue only consumes the output |
