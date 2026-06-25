# Pi setup

Pi is a stable LazyAI target for Pi agents, skills, prompt templates, TypeScript extensions, system prompts, and settings.

## Generated structure

```text
.
├── AGENTS.md
└── .pi/
    ├── agents/<agent>.md
    ├── skills/<skill>/SKILL.md
    ├── prompts/<prompt>.md
    ├── extensions/<extension>.ts
    ├── extensions/<extension>/index.ts
    ├── SYSTEM.md
    ├── APPEND_SYSTEM.md
    └── settings.json

```mermaid
flowchart LR
    AI[".ai/ canonical source"] --> A[".pi/agents"]
    AI --> S[".pi/skills"]
    AI --> P[".pi/prompts"]
    LIB["Pi library assets"] --> E[".pi/extensions"]
    LIB --> Y[".pi/SYSTEM.md + APPEND_SYSTEM.md"]
    AI --> Z[".pi/settings.json"]
    MCP[".ai/mcp.json"] -. no emitted config .-> N["Pi MCP no-op"]

## Pi concepts LazyAI uses

| Pi concept | LazyAI source |
|---|---|
| Root instructions | `AGENTS.md` |
| Agent profiles | canonical agent markdown under `.pi/agents/` |
| Skills | Agent Skills-compatible `SKILL.md` directories |
| Prompts | prompt markdown under `.pi/prompts/` |
| Extensions | TypeScript extensions under `.pi/extensions/` (flat files or directory `index.ts` layouts) |
| System prompts | `.pi/SYSTEM.md` and `.pi/APPEND_SYSTEM.md` |
| Settings | `.pi/settings.json` / `~/.pi/agent/settings.json` resource references |
| MCP | no capability declared; no Pi MCP config is emitted |

## LazyAI options

| Use case | Command |
|---|---|
| Add Pi during init | `lazyai-cli init --scope project --tools pi --preset standard --no-interactive` |
| Add Pi later | `lazyai-cli add --tools pi --no-interactive` |
| Include prompts/skills from full preset | `lazyai-cli init --scope project --tools pi --preset full --no-interactive` |
| Build a Pi bundle | `lazyai-cli build-plugin --target pi --out ./dist/pi` |

## Example

```bash
lazyai-cli init \
  --scope project \
  --tools pi \
  --preset full \
  --name my-app \
  --no-interactive

lazyai-cli doctor
```

## Readiness notes

- Support level: stable.
- Project, workspace, and global (`~/.pi/`) scopes are supported.
- Pi has no `.pi/hooks` directory; LazyAI emits safety behavior as `.pi/extensions/`.
- `lazyai-cli compile --tool pi` reads canonical MCP state but emits no MCP file for Pi.
- Pi settings compilation emits resource references for `extensions`, `skills`, `prompts`, and a local package-root reference; theme preferences remain user-owned.
- Pi can load `.pi/SYSTEM.md` and `.pi/APPEND_SYSTEM.md`; LazyAI emits both when present in the Pi library.
