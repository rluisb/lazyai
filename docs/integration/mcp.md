# MCP Integration

`lazyai-cli` maintains a canonical MCP catalog under `.ai/mcp.json` and compiles it into each tool's native config format.

## Canonical source

The canonical catalog is written to:

```text
.ai/mcp.json
```

It contains bundled server definitions, enable/disable state, install hints, and environment-variable placeholders.

## Per-tool compilation

| Tool | Compiled MCP output | Notes |
|---|---|---|
| OpenCode | `.opencode/opencode.jsonc` | MCP entries live under the OpenCode config |
| Claude Code | `.mcp.json` or `.claude/settings.local.json` | Depends on scope and `--local-secrets` |
| Copilot | `.vscode/mcp.json` and optional `~/.copilot/mcp-config.json` | CLI probe decides whether the home config is emitted |

## Enabling servers

During setup:

```bash
lazyai-cli init --enable-servers filesystem,memory
```

Or edit `.ai/mcp.json` later, then recompile:

```bash
lazyai-cli compile
```

## Disabling servers

Set a server's `enabled` flag to `false` in `.ai/mcp.json`, then rerun `lazyai-cli compile`.

## Bundled MCP servers

| Server | Default status | Requires install | Notes |
|---|---|---|---|
| `memory` | enabled | No | Knowledge graph memory server |
| `filesystem` | enabled | No | Local filesystem read/write access |
| `ripgrep` | enabled | No | Fast code search |
| `memoria` | enabled | No | Git history + code memory |
| `codegraph` | enabled | No | Repository graph exploration |
| `qmd` | enabled | No | Markdown/document search |
| `graphify` | disabled | No | Knowledge-graph rendering |
| `obsidian` | disabled | No | Obsidian vault integration |
| `playwright` | disabled | No | Browser automation |
| `atlassian` | disabled | No | Jira + Confluence integration |
| `fetch` | disabled | No | General HTTP fetch MCP |

## Token-efficient usage

Many bundled servers have equivalent CLI tools. For bulk or deterministic work, prefer the CLI to avoid MCP JSON-RPC overhead and keep agent context windows small. See [MCP vs CLI](../concepts/mcp-vs-cli.md) for the broader comparison.

## Environment variables

If any enabled MCP server declares env vars, `lazyai-cli` generates:

```text
.env.example
```

`lazyai-cli` never writes real secrets into `.env.example`.

## Removed runtime note

The old orchestration runtime and its MCP server were removed from the active product surface during the runtime refactor. Use [the migration note](../migration/fortnite-orchestrator-removal.md) for compatibility and rollback guidance.
