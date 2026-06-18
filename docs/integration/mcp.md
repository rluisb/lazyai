# MCP Integration

`lazyai-cli` maintains a canonical MCP catalog under `.ai/` and compiles it into each tool's native config format.

## Canonical source

The canonical catalog is written to:

```text
.ai/mcp.json
```

It contains the full catalog of bundled servers, including:

- server definitions
- whether each server is enabled or disabled
- install requirements
- command/URL configuration
- environment variable placeholders

## Per-tool compilation

| Tool | Compiled MCP output | Notes |
|---|---|---|
| OpenCode | `.opencode/opencode.jsonc` | Includes enabled + disabled servers |
| Claude Code | `.mcp.json` | Includes only enabled servers |
| GitHub Copilot | `.vscode/mcp.json` | Includes only enabled servers |

## Enabling servers

During setup:

```bash
lazyai-cli init --enable-servers atlassian,playwright,orchestrator
```

Or edit `.ai/mcp.json` later, then recompile:

```bash
lazyai-cli compile
```

## Disabling servers

Edit `.ai/mcp.json` and set the server's `enabled` flag to `false`, then rerun `lazyai-cli compile`.

## Bundled MCP servers

| Server | Default status | Requires install | Notes |
|---|---|---|---|
| `memory` | enabled | No | Knowledge graph memory server |
| `filesystem` | enabled | No | Local filesystem read/write access |
| `ripgrep` | enabled | No | Fast code search |
| `memoria` | enabled | No | Git history + code memory |
| `codegraph` | disabled | Yes | Semantic code graph (`go install` or project-specific install) |
| `qmd` | disabled | Yes | Markdown knowledge search (`brew install qmd`) |
| `playwright` | disabled | No | Browser automation and testing |
| `context7` | disabled | No | Remote docs lookup; uses `CONTEXT7_API_KEY` |
| `atlassian` | disabled | No | Jira/Confluence remote access |
| `brave-search` | disabled | No | Web search; needs `BRAVE_API_KEY` |
| `fetch` | disabled | No | General HTTP fetch MCP |
| `orchestrator` | disabled | Yes | Optional LazyAI orchestration runtime (`lazyai-orchestrator`) |

## Token-efficient usage — CLI-first vs MCP

Many bundled servers have equivalent CLI tools. For bulk or deterministic work, prefer the CLI to avoid MCP JSON-RPC overhead and keep agent context windows small. See [MCP vs CLI](../concepts/mcp-vs-cli.md) for the full comparison.

Quick reference:

| Server | Recommended interface | Reason |
|---|---|---|
| `filesystem` | **CLI-first** | Bulk file ops are cheaper via `lazyai-cli` or native shell commands. |
| `ripgrep` | **CLI-first** | Large search results stream faster through `rg` than via MCP JSON. |
| `fetch` | **CLI-first** | `curl` / `wget` avoid double serialization for large payloads. |
| `graphify` | **CLI-first** for batch; **MCP** for interactive | Batch ingestion via CLI; live queries via MCP. |
| `obsidian` | **CLI-first** for bulk; **MCP** for live queries | Bulk exports via `ob`; session queries via MCP. |
| `qmd` | **Hybrid** | CLI for scripted indexing; MCP for inline agent queries. |
| `codegraph` | **Hybrid** | CLI for initial index/build; MCP for semantic context calls. |
| `playwright` | **MCP-only** | No stable CLI equivalent for agent-driven browser automation. |
| `atlassian` | **MCP-only** | OAuth and remote API abstraction require the MCP server. |
| `memory` | **MCP-only** | Stateful graph operations need the persistent server. |
| `memoria` | **MCP-only** | No dedicated CLI wrapper for git-history queries. |
| `orchestrator` | **MCP-only** | Workflow control plane is only exposed through MCP. |

To apply a token-efficient preset, disable `filesystem`, `ripgrep`, and `fetch` in `.ai/mcp.json`, then recompile:

```bash
lazyai-cli compile
```

## Environment variables

If any enabled MCP server declares env vars, `lazyai-cli` generates:

```text
.env.example
```

Example:

```dotenv
# Required by: brave-search
BRAVE_API_KEY=
```

`lazyai-cli` never writes real secrets into `.env.example`.

## Orchestrator MCP server

When `orchestrator` is enabled, `lazyai-cli` scaffolds `.ai/orchestration/` and generates per-tool orchestrator guidance files. Install the runtime with:

```bash
go install github.com/rluisb/lazyai/packages/orchestrator/cmd/lazyai-orchestrator@latest
```

See [Orchestration](orchestration.md).
