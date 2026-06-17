# MCP vs CLI — Token Efficiency Guide

LazyAI mixes MCP servers and companion CLIs to cover the same workflow surface where that makes sense. Choosing the right interface for each task saves tokens, reduces latency, and keeps agent context windows focused.

## The Rule of Thumb

> **Use the CLI for bulk, deterministic, or filesystem-heavy work. Use MCP for interactive, stateful, or agent-orchestrated work.**

## Interface Comparison

| Capability | MCP Server | CLI Equivalent | Token Policy |
|---|---|---|---|
| **Filesystem** | `filesystem` | `lazyai-cli` built-in file ops, `ls`, `cat`, `find` | **CLI-first** — MCP adds JSON wrapping + round-trip overhead for every file read/write. |
| **Code search** | `ripgrep` | `rg` | **CLI-first** — Large result sets serialize through MCP as JSON arrays; CLI streams directly. |
| **Project memory + handoffs** | `ai-memory` | `ai-memory` | **Hybrid** — MCP is best for in-session retrieval, handoffs, and durable notes; the CLI is best for bootstrap, install, status, and admin flows. |
| **Obsidian vault** | `obsidian` | `ob` CLI | **CLI-first** — For bulk note exports/imports; use MCP for live vault queries during a session. |
| **Code graph** | `codegraph` | `codegraph` CLI | **Hybrid** — CLI for initial index/build; MCP for semantic `codegraph_context` calls. |
| **GitHub workflows** | — | `gh` | **CLI-first** — Current repository, issue, and PR workflows stay on the GitHub CLI surface. |
| **Agent sandboxing** | — | `ai-jail` | **CLI-only** — Wrap the agent process before it starts; there is no MCP session surface for sandbox policy. |

## Why CLI Saves Tokens

1. **No JSON envelope** — MCP wraps every tool call and result in a JSON-RPC message. For 100 file reads, that's 100 extra envelopes.
2. **No server process** — MCP servers run as persistent stdio or SSE processes. CLI commands exit immediately, freeing memory and context.
3. **Streaming output** — CLI tools stream results; MCP buffers them into a single response.
4. **Native composition** — Shell pipelines (`rg pattern | wc -l`) are more efficient than chaining MCP tool calls.

## When MCP Is Worth the Overhead

- **Stateful project memory** — `ai-memory` keeps retrieval, handoffs, and durable annotations close to the active session.
- **Inline code intelligence** — `codegraph` can answer semantic questions without leaving the current session.
- **Vault-backed note access** — `obsidian` is useful when you need one-off note reads or writes mid-session.

The retired LazyAI orchestration runtime is not part of the active MCP catalog. See [Migration: Fortnite / orchestrator removal](../migration/fortnite-orchestrator-removal.md) for compatibility guidance.

## Recommended Preset Adjustments

If you are token-constrained (e.g., large monorepos, long-running agents):

1. **Omit `filesystem` and `ripgrep` from your MCP selection** — rely on CLI file ops and `rg` instead.
2. **Keep `ai-memory` selected when durable session memory is useful.**
3. **Use `codegraph` and `obsidian` via whichever interface is closer to the task** — CLI for batch, MCP for inline.

## Configuration

No special configuration is required for CLI-only companions. Install the local binaries you intend to use, keep `.ai/mcp.json` focused on MCP servers, and use wrappers like `ai-jail` at process start rather than through MCP.

## See Also

- [MCP Integration](../integration/mcp.md) — full server catalog and setup instructions.
- [Tools](tools.md) — CLI tool inventory and usage.
