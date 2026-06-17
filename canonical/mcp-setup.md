# MCP Setup (Supported Terminal-First Tools)

> How to connect external tools through MCP for the surfaces vibe-lab actively supports: Claude Code, OpenCode, and OMP/Pi.

## The Four Points

- **WHAT** — Add an MCP server so the agent can use external tools (database, browser, GitHub, etc.).
- **HOW** — Declare the server in the correct tool-specific config; verify with that tool's observable MCP signal.
- **What I DON'T want** — Claude-only config copied into OpenCode or OMP/Pi; blind npm installs; committed secrets.
- **How we VALIDATE** — Confirm the tool exposes the MCP server, then ask a natural-language prompt that requires the server and confirm the agent uses the MCP call.

## Support Matrix

| Tool | Config Surface | Status |
|------|----------------|--------|
| Claude Code | Project `.mcp.json` | Supported example |
| OpenCode | `mcp` block in existing `opencode.json` or `opencode.jsonc` | Supported example |
| OMP/Pi | No project-local MCP config verified in vibe-lab | Unsupported here; document external setup separately |

## Claude Code Example

Project-local `.mcp.json`:

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "."]
    }
  }
}
```

Remote server:

```json
{
  "mcpServers": {
    "ai-memory": {
      "url": "http://127.0.0.1:49374/mcp"
    }
  }
}
```

Verify:

```bash
claude mcp list
```

## OpenCode Example

Use whichever project-local OpenCode config already exists: `opencode.json` or `opencode.jsonc`. Do not rename the file just for vibe-lab.

```json
{
  "mcp": {
    "codegraph": {
      "type": "local",
      "enabled": true,
      "command": [
        "codegraph",
        "serve",
        "--mcp"
      ]
    }
  }
}
```

Remote server:

```json
{
  "mcp": {
    "ai-memory": {
      "type": "remote",
      "enabled": true,
      "url": "http://127.0.0.1:49374/mcp"
    }
  }
}
```

Verify:

1. Restart OpenCode so it reloads config.
2. Ask the agent to perform a task that requires the server.
3. Confirm the response used an `mcp__<server>_*` tool instead of guessing.

## OMP/Pi Boundary

vibe-lab currently verifies OMP/Pi as shared `AGENTS.md` plus `.pi/skills` symlinks only. No project-local OMP/Pi MCP config path is verified here.

When a separate OMP/Pi MCP mechanism is available:

1. Document that tool's exact config file and server shape.
2. Add a dedicated OMP/Pi example before claiming support.
3. Add a verification command or manual scenario that proves the MCP server is exposed.

Until then, mark OMP/Pi MCP as **unsupported in vibe-lab** instead of reusing Claude or OpenCode config.

## Security Rules

1. **Never commit tokens.** Use environment references; inject via shell, `.env`, or the tool's secret mechanism.
2. **Pin versions.** Use `@version` in npx args or a lockfile. Do not use `-y` for untrusted sources.
3. **Scope per project where the tool supports it.** Global MCP servers affect every session.
4. **One server per concern.** Avoid kitchen-sink configs that slow agent startup.

## Verification Checklist

After adding a server:

- [ ] The configured tool exposes the server (`claude mcp list`, OpenCode tool call evidence, or a documented OMP/Pi equivalent).
- [ ] Agent responds to a natural prompt with the MCP tool, not hallucinated output.
- [ ] Token/credential is not visible in repo files (`git grep` check).
- [ ] Unsupported surfaces are explicitly marked unsupported instead of given borrowed config.

## Categories

| Category | Common Servers |
|----------|----------------|
| Context | ai-memory, filesystem, obsidian |
| Code Intelligence | ripgrep, codegraph |
| Code Quality | linter, test runner, coverage |
| Data | PostgreSQL, SQLite, Redis |
| External | none shipped in the current MCP catalog |
