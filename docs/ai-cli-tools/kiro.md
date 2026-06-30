# Kiro setup

Kiro is a stable LazyAI target for Kiro agent profiles, skills, prompt templates, native Kiro v3 hooks, permissions/global configuration metadata, and MCP.

## Generated structure

```text
.
├── AGENTS.md
└── .kiro/
    ├── agents/<agent>.json
    ├── skills/<skill>/SKILL.md
    ├── prompts/<prompt>.md
    ├── hooks/block-destructive-shell.json
    ├── hooks/block-destructive-shell.sh
    └── settings/mcp.json
```

```mermaid
flowchart LR
    AI[".ai/ canonical source"] --> A[".kiro/agents"]
    AI --> S[".kiro/skills"]
    AI --> P[".kiro/prompts"]
    MCP[".ai/mcp.json"] --> M[".kiro/settings/mcp.json"]
```

## Kiro concepts LazyAI uses

| Kiro concept | LazyAI source |
|---|---|
| Root instructions | `AGENTS.md` |
| Custom agent profiles | canonical agents transformed to JSON at `.kiro/agents/<name>.json`; `tools`/`allowedTools` from `tools:` frontmatter |
| Skills | Agent Skills-compatible `SKILL.md` directories |
| Prompts | prompt markdown under `.kiro/prompts/` |
| Hooks | native Kiro v3 hook JSON under `.kiro/hooks/` |
| MCP | `.ai/mcp.json` compiled to `.kiro/settings/mcp.json` |

## LazyAI options

| Use case | Command |
|---|---|
| Add Kiro during init | `lazyai-cli init --scope project --tools kiro --preset standard --no-interactive` |
| Add Kiro later | `lazyai-cli add --tools kiro --no-interactive` |
| Compile only Kiro MCP | `lazyai-cli compile --tool kiro` |
| Preview Kiro MCP output | `lazyai-cli compile --tool kiro --dry-run` |

## Example

```bash
lazyai-cli init \
  --scope project \
  --tools kiro \
  --preset standard \
  --enable-servers filesystem \
  --no-interactive

lazyai-cli compile --tool kiro
lazyai-cli status
```

## Readiness notes

- Support level: stable.
- Project, workspace, and global scopes are supported.
- LazyAI emits native Kiro v3 hooks at `.kiro/hooks/<name>.json` for source-verified trigger mappings only; today that is `block-destructive-shell` on `PreToolUse`.
- LazyAI intentionally emits no `.kiro/workflows`, specs, steering, commands, chat modes, templates, or output styles.
- Repo-local permissions are forbidden (`Permissions: true` is host-support metadata), and direct `.kiro/powers/` output is not emitted.
