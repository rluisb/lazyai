# MCP vs CLI — Token Efficiency Guide

LazyAI exposes the same capabilities through two interfaces: **MCP servers** (consumed by AI agents via the Model Context Protocol) and **CLI commands** (run directly in your terminal). Choosing the right interface for each task saves tokens, reduces latency, and keeps agent context windows focused.

## The Rule of Thumb

> **Use the CLI for bulk, deterministic, or filesystem-heavy work. Use MCP for interactive, stateful, or agent-orchestrated work.**

## Interface Comparison

| Capability | MCP Server | CLI Equivalent | Token Policy |
|---|---|---|---|
| **Filesystem** | `filesystem` | `lazyai-cli` built-in file ops, `ls`, `cat`, `find` | **CLI-first** — MCP adds JSON wrapping + round-trip overhead for every file read/write. |
| **Code search** | `ripgrep` | `rg` | **CLI-first** — Large result sets serialize through MCP as JSON arrays; CLI streams directly. |
| **Web fetch** | `fetch` | `curl`, `wget` | **CLI-first** — For single URLs, MCP is fine; for bulk scraping or large payloads, CLI avoids double serialization. |
| **Knowledge graph** | `graphify` | `graphify` CLI | **CLI-first** for batch ingestion; **MCP** for interactive exploration (`query_graph`, `shortest_path`). |
| **Obsidian vault** | `obsidian` | `ob` CLI | **CLI-first** for bulk note exports/imports; **MCP** for live vault queries during a session. |
| **Markdown search** | `qmd` | `qmd` CLI | **Hybrid** — CLI is faster for scripted indexing; MCP is convenient for inline agent queries. |
| **Code graph** | `codegraph` | `codegraph` CLI | **Hybrid** — CLI for initial index/build; MCP for semantic `codegraph_context` calls. |
| **Browser automation** | `playwright` | `npx playwright` | **MCP-only** — No stable CLI equivalent for agent-driven browser snapshots and clicks. |
| **Atlassian (Jira/Confluence)** | `atlassian` | `acli` | **MCP-only** — The Atlassian MCP server handles OAuth and remote APIs; `acli` is a separate surface. |
| **Memory / knowledge graph** | `memory` | None | **MCP-only** — Stateful graph operations require the persistent MCP server. |
| **Memoria (git history)** | `memoria` | `npx @byronwade/memoria` | **MCP-only** — No dedicated CLI wrapper; use MCP for `ask_history` and `search_memories`. |

## Why CLI Saves Tokens

1. **No JSON envelope** — MCP wraps every tool call and result in a JSON-RPC message. For 100 file reads, that's 100 extra envelopes.
2. **No server process** — MCP servers run as persistent stdio or SSE processes. CLI commands exit immediately, freeing memory and context.
3. **Streaming output** — CLI tools stream results; MCP buffers them into a single response.
4. **Native composition** — Shell pipelines (`rg pattern | wc -l`) are more efficient than chaining MCP tool calls.

## When MCP Is Worth the Overhead

- **Stateful knowledge** — `memory` maintains context across multiple agent turns.
- **Remote APIs** — `atlassian` and `brave-search` abstract OAuth and rate limits.
- **Interactive exploration** — `playwright` snapshots and clicks are naturally request/response.

The retired LazyAI orchestration runtime is not part of the active MCP catalog. See [Migration: Fortnite / orchestrator removal](../migration/fortnite-orchestrator-removal.md) for compatibility guidance.

## Recommended Preset Adjustments

If you are token-constrained (e.g., large monorepos, long-running agents):

1. **Disable `filesystem` and `ripgrep` MCP servers** — rely on CLI file ops and `rg` instead.
2. **Keep `memory`, `playwright`, and `atlassian` enabled when you need state, browser interaction, or SaaS APIs.**
3. **Use `qmd` and `codegraph` via whichever interface is closer to the task** — CLI for batch, MCP for inline.

## Configuration

No special configuration is required. The CLI tools are already listed in `.ai/mcp.json` under `cliTools`. Ensure they are installed locally (`brew install`, `npm install -g`, etc.) and the agent is instructed to prefer CLI for bulk operations.

## See Also

- [MCP Integration](../integration/mcp.md) — full server catalog and setup instructions.
- [Tools](tools.md) — CLI tool inventory and usage.
