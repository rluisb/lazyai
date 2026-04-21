# Spec 014: Copilot `--drive-cli` — Research

**Date:** 2026-04-20
**Status:** Research — awaiting human gate

## §1 — The deferred item

From `specs/KNOWLEDGE_MAP.md` Pending/Follow-up:
> `--drive-cli` for OpenCode (interactive-only upstream) / Copilot (flag surface unverified)

Spec 013 added CLI probes (`LookupCopilotBinary`, `CopilotHomePresent`) but did **not** wire Copilot MCP registration through the CLI. Today Copilot only writes `.vscode/mcp.json` (VS Code extension) at project/workspace scope; the standalone `@github/copilot` CLI receives no MCP config from ai-setup.

## §2 — Upstream reality check

From the official docs (https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-mcp-servers):

| Surface | Path / Command | Scriptable? |
|---|---|---|
| Persistent MCP config | `~/.copilot/mcp-config.json` | File edit only |
| One-off (session) | `copilot --additional-mcp-config <path>` | Yes (flag) |
| Interactive `/mcp add`, `/mcp edit`, `/mcp delete` | slash commands inside a running session | **No — interactive only** |

**Schema** of `~/.copilot/mcp-config.json`:
```json
{
  "mcpServers": {
    "<name>": {
      "type": "local|http|stdio|sse",
      "command": "...", "args": [...], "env": {...},
      "url": "...", "headers": {...},
      "tools": ["*"] or ["tool1", ...]
    }
  }
}
```

**Finding:** There is **no** scriptable `copilot mcp add` command comparable to `claude mcp add-json` or `gemini mcp add`. The upstream model is **direct-file edit** (same as OpenCode).

## §3 — Name correction

The pending item is misnamed. `--drive-cli` implies "delegate to a CLI subcommand" (Claude/Gemini/Codex pattern). For Copilot, there is nothing to delegate to. The real work is:

> **Add `~/.copilot/mcp-config.json` as a Copilot compile target at global scope, deep-merge with backup-on-first-touch** (mirrors OpenCode's direct-write decision).

I propose renaming this spec from "Copilot --drive-cli" to **"Copilot global MCP compile (mcp-config.json)"** to match the actual work.

## §4 — Current state in ai-setup

`internal/adapter/mcp_compiler.go:compileCopilotMCP`:
- Writes only `.vscode/mcp.json` for VS Code extension
- Skips global scope (probe-gated; returns 0 records)
- `toCopilotMcp` emits `{ servers: {...}, inputs: [...] }` shape (VS Code schema)

Gap:
- At `scope=global`, if probe passes, we should emit `~/.copilot/mcp-config.json`
- Schema is **different** from VS Code — uses `mcpServers` key, no `inputs` (CLI reads env directly, no VS Code prompt UI)

## §5 — Options

| Option | Description | Pros | Cons |
|---|---|---|---|
| **A. Direct-write to `~/.copilot/mcp-config.json`** | At global scope + probe pass, emit mcp-config.json via deep-merge | Matches upstream (no CLI driving possible); consistent with OpenCode pattern | New transform function; new merge target |
| **B. Use `copilot --additional-mcp-config <path>` on CLI invocation** | Generate mcp-config.json fragment; rely on Copilot flag at runtime | Clean separation from user's persistent config | User has to remember the flag; doesn't address the core ask |
| **C. Do nothing, close the item** | Accept that Copilot global MCP is user-managed | Zero code | Leaves the feature gap; breaks parity with other tools |

**Recommendation: A.** It closes the parity gap, respects user-authored keys via deep-merge, and mirrors the pattern locked in spec 011.

## §6 — Decision interview (need human answers)

1. **Q1** — Confirm renaming from "drive-cli" → "global MCP compile (mcp-config.json)"?
2. **Q2** — Pick Option A (direct-write with deep-merge) vs alternatives?
3. **Q3** — Should project/workspace scope also emit to `~/.copilot/mcp-config.json` (so CLI works even in per-repo usage), or strictly global-only?
4. **Q4** — VS Code `.vscode/mcp.json` uses `servers` key; Copilot CLI uses `mcpServers`. Do we maintain **two separate transforms**, or add an abstraction?
5. **Q5** — Does probe gating still apply (only emit if `copilot` on PATH or `~/.copilot/` present), or always emit at global (user may install Copilot later)?

## §7 — Risks & constraints

- **R1:** Different schema per surface (VS Code vs CLI). Must keep both transforms in sync when MCP catalog changes.
- **R2:** Deep-merge needs to preserve user-authored `mcpServers` entries; managed keys win on collision (existing pattern).
- **R3:** No CLI validation available (no `copilot mcp list` equivalent); can't confirm registration like spec 012 did for Claude.

## §8 — Out of scope

- OpenCode `--drive-cli` (upstream permanently interactive — verified in spec 011)
- Copilot `copilot-instructions.md` changes (already shipped in spec 013)
- Copilot cloud/marketplace publishing (parked per spec 013)
