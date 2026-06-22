# Suggested CLI Tools

LazyAI favors a **CLI-first** workflow: for bulk or deterministic work, a real command-line
tool is cheaper and more predictable than an MCP JSON-RPC round trip (see
[MCP vs CLI](../concepts/mcp-vs-cli.md)). During `lazyai-cli init` you can select
companion CLI tools to enable; the canonical catalog records each tool's purpose and
install hint so the agent knows the local command exists.

The catalog lives in `packages/cli/library/mcp/catalog.json` under the `cliTools` key.
LazyAI never installs these for you and never runs package managers on your behalf — it
records the selection and surfaces the install hint. You install the binary yourself.

## Selecting CLI tools

```bash
# Interactive: pick from the "Which CLI tools would you like to enable?" step
lazyai-cli init

# Non-interactive
lazyai-cli init --cli-tools gh,ai-memory,codegraph --no-interactive
```

## Cataloged tools

| Tool | Purpose | Install |
|---|---|---|
| `gh` | GitHub CLI for repository, issue, and PR operations. Never pushes without explicit user approval. | `brew install gh` |
| `ai-memory` | Thin client and installer for the ai-memory project-memory server. | `curl -fsSL https://raw.githubusercontent.com/akitaonrails/ai-memory/main/bin/ai-memory -o ~/.local/bin/ai-memory && chmod +x ~/.local/bin/ai-memory` |
| `ai-jail` | Sandbox wrapper for AI coding agents with per-project path constraints. | `brew tap akitaonrails/tap && brew install ai-jail` |
| `codegraph` | Code knowledge graph for semantic exploration. | `bun install -g codegraph` |
| `ob` | Obsidian vault read/write access for persistent knowledge. | `npm install -g obsidian-cli` |

## Relationship to MCP servers

Several CLI tools have an MCP counterpart with the same backing binary
(`ai-memory`, `codegraph`, `ob`/`obsidian`). The catalog marks each MCP server's
`preferred_interface` and `cli_equivalent` so the agent prefers the CLI for bulk or
deterministic work and reserves MCP for interactive, context-rich calls. See the
[MCP Servers](mcp-servers.md) reference for the server-side details.
