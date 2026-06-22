# MCP Servers

LazyAI maintains a canonical MCP catalog under `.ai/mcp.json` and compiles it into each
tool's native config format (see [MCP Integration](../integration/mcp.md) for the
compile flow and per-tool output files). This page documents the servers that ship in
the catalog.

The authoritative source is `packages/cli/library/mcp/catalog.json` (schema version `1.0`).
Each server records its transport, the tools it exposes, whether it requires a separate
install, and an optional setup hint. Servers also declare a `preferred_interface` and
`cli_equivalent` so the agent can choose the cheaper CLI path for bulk work where one
exists.

## Enabling servers

```bash
# At init
lazyai-cli init --enable-servers filesystem,ai-memory

# Later: edit .ai/mcp.json (set "enabled": true on a server) then recompile
lazyai-cli compile
```

LazyAI does not create or manage `.env` files. When an enabled server declares
environment variables, `lazyai-cli` prints shell-ready `export NAME=""` guidance; supply
real values through your shell, a secret manager, or a local env file per your project
policy.

## Cataloged servers

| Server | Transport | Requires install | Preferred interface | Purpose |
|---|---|---|---|---|
| `ai-memory` | HTTP (`http://127.0.0.1:49374/mcp`) | Yes | hybrid | Long-term project memory, handoffs, and durable annotations |
| `filesystem` | stdio (`npx @modelcontextprotocol/server-filesystem`) | No | cli | Local filesystem read/write access |
| `ripgrep` | stdio (`npx mcp-ripgrep`) | No | cli | Fast code search via ripgrep |
| `codegraph` | stdio (`codegraph serve --mcp`) | Yes | hybrid | Code knowledge graph for semantic exploration |
| `obsidian` | stdio (`ob mcp`) | Yes | cli-first | Obsidian vault read/write for persistent knowledge |

### ai-memory

- **Tools:** `memory_query`, `memory_recent`, `memory_handoff_accept`, `memory_write_page`, `memory_auto_improve`, `memory_consolidate`
- **CLI equivalent:** `ai-memory`
- **Install:** `curl -fsSL https://raw.githubusercontent.com/akitaonrails/ai-memory/main/bin/ai-memory -o ~/.local/bin/ai-memory && chmod +x ~/.local/bin/ai-memory`
- **Setup:** Start the ai-memory server, then point clients at `http://127.0.0.1:49374/mcp` or a remote `/mcp` endpoint. Add an `Authorization` header when bearer auth is enabled.

### filesystem

- **Tools:** `read_file`, `write_file`, `list_directory`, `search_files`
- **CLI equivalent:** native shell / `lazyai-cli` file ops
- **Install:** none (runs via `npx -y @modelcontextprotocol/server-filesystem .`)
- Enabled by default; prefer native shell/CLI for bulk file work.

### ripgrep

- **Tools:** `search`, `list-files`
- **CLI equivalent:** `rg`
- **Install:** none (runs via `npx -y mcp-ripgrep`)

### codegraph

- **Tools:** `codegraph_context`, `codegraph_search`, `codegraph_files`
- **CLI equivalent:** `codegraph`
- **Install:** `bun install -g codegraph`
- **Setup:** Initialize the repository index with `codegraph init`.

### obsidian

- **Tools:** `read_note`, `write_note`, `search_notes`, `list_notes`
- **CLI equivalent:** `ob`
- **Install:** `npm install -g obsidian-cli`
- **Setup:** Ensure the Obsidian "Advanced URI" plugin is installed and enabled.

## Per-tool compilation

The catalog compiles into each tool's native MCP config. See
[MCP Integration](../integration/mcp.md) and [Tool Outputs](tool-outputs.md) for the
exact output files per tool (for example `opencode.json` for OpenCode, `.mcp.json` for
Claude Code, `.vscode/mcp.json` for Copilot).

## L1 vs L3 validation

`lazyai-cli server doctor` performs **L1 config checks** only: it verifies that the server entry exists in `.ai/mcp.json`, is enabled, and is present in each per-tool compiled MCP config file.

**L3 stdio handshake** (spawning the server process and performing a `tools/list` JSON-RPC exchange) is not performed. The Go binary does not bundle an MCP client library for the handshake protocol. A TypeScript wrapper using `@modelcontextprotocol/sdk` could perform L3 checks; this is future work.

The `server doctor` output includes a `stdio handshake` check that is always skipped, with a message explaining the limitation.
