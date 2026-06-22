# Tool Outputs

This page documents what `lazyai-cli` writes for each configured companion tool, grounded in `packages/cli/internal/adapter`.

## Comparison

| Tool | Root file | Config dir | Agents | Skills | Hooks | MCP |
|---|---|---|---|---|---|---|
| OpenCode | `AGENTS.md` | `.opencode/` | Yes (`.opencode/agents`) | Yes (`.opencode/skills`) | Yes (`.opencode/plugins/vibe-lab-hooks.js`) | `opencode.json` |
| Claude Code | `AGENTS.md` (plus `CLAUDE.md` compat handling) | `.claude/` | Yes (`.claude/agents`) | Yes (`.claude/skills`) | Yes (`.claude/hooks/*`) | `.mcp.json` (+ `settings.local` when local-secrets) |
| GitHub Copilot | `.github/copilot-instructions.md` | `.github/` | Yes (`.github/agents`) | Yes (`.github/skills`) | Yes (`.github/hooks/*`) | `.vscode/mcp.json`, optional `~/.copilot/mcp-config.json` |
| Pi | `AGENTS.md` | `.pi/` | Yes (`.pi/agents`) | Yes (`.pi/skills`) | none (`.pi/extensions/*.ts`) | none |
| OMP (beta) | `AGENTS.md` | `.omp/` | Yes (`.omp/agents`) | Yes (`.omp/skills`) | Yes (`.omp/hooks/pre/*.ts`) | `.omp/mcp.json` |
| Antigravity (beta) | `AGENTS.md` | `.gemini/` | none | Yes (`.agents/skills`) | Yes (`.gemini/hooks/lazyai/*` and `.agents/hooks.json`) | `~/.gemini/config/mcp_config.json` |
| Kiro | `AGENTS.md` | `.kiro/` | Yes (`.kiro/agents`) | Yes (`.kiro/skills`) | none | `.kiro/settings/mcp.json` |

## OpenCode

**Surface summary:** emits OpenCode-native agents, skills, commands/modes, MCP, and managed hook plugin.

- **Root instructions:** `AGENTS.md`.
- **Config directory:** `.opencode/`
- **Config tree (key files):**

```text
.
├── AGENTS.md
└── .opencode/
    ├── agents/guide.md
    ├── agents/<agent>.md
    ├── skills/<skill>/SKILL.md
    ├── commands/<command>.md
    ├── modes/<mode>.md
    ├── plugins/vibe-lab-hooks.js
    └── opencode.json
```

- **Agents surface:** `.opencode/agents/<name>.md`.
- **Skills surface:** `.opencode/skills/<name>/SKILL.md`.
- **Commands / prompts / chat modes:** commands in `.opencode/commands`; chat modes in `.opencode/modes`; no prompt surface.
- **Hook runtime surface:** `.opencode/plugins/vibe-lab-hooks.js`.
- **MCP output files:** `opencode.json` (with merged managed `mcp`).
- **Global scope support:** `yes`.
- **Maturity note:** stable.

## Claude Code

**Surface summary:** emits root instructions plus AGENTS, agents, skills, commands, output styles, hooks, and MCP config.

- **Root instructions:** `AGENTS.md` + project/workspace `CLAUDE.md` compatibility handling.
- **Config directory:** `.claude/`
- **Config tree (key files):**

```text
.
├── AGENTS.md
└── .claude/
    ├── agents/<agent>.md
    ├── skills/<skill>/SKILL.md
    ├── commands/<command>.md
    ├── output-styles/<style>.md
    ├── hooks/block-destructive-shell.sh
    ├── hooks/objective-workflow-gate.sh
    ├── rules/typescript.md
    └── settings.json
```

- **Agents surface:** `.claude/agents/<name>.md`.
- **Skills surface:** `.claude/skills/<name>/SKILL.md`.
- **Commands / prompts / chat modes:** commands in `.claude/commands`; output styles in `.claude/output-styles`; no prompt/chat-mode directories.
- **Hook runtime surface:** `.claude/hooks/*.sh` and hook wiring in `.claude/settings.json`.
- **MCP output files:** `.mcp.json` (project/workspace), optionally `.claude/settings.local.json` when local-secrets mode is enabled.
- **Global scope support:** `yes`.

## GitHub Copilot

**Surface summary:** emits project instructions, agents, skills, prompts, chat modes, and MCP outputs.

- **Root instructions:** `.github/copilot-instructions.md`.
- **Config directory:** `.github/`
- **Config tree (key files):**

```text
.
├── .github/
│   ├── copilot-instructions.md
│   ├── agents/<agent>.agent.md
│   ├── instructions/<instruction>.md
│   ├── prompts/<prompt>.prompt.md
│   ├── hooks/*.json
│   ├── hooks/*.sh
│   ├── chatmodes/<chatmode>.chatmode.md
│   └── skills/<skill>/SKILL.md
└── .vscode/mcp.json

# if probe-gated global install succeeds:
#   ~/.copilot/mcp-config.json
```

- **Agents surface:** `.github/agents/<name>.agent.md`.
- **Skills surface:** `.github/skills/<name>/SKILL.md`.
- **Commands / prompts / chat modes:** no command surface; prompts at `.github/prompts/*.prompt.md`; chat modes at `.github/chatmodes/*.chatmode.md`.
- **Hook runtime surface:** `.github/hooks/*.json` and `.github/hooks/*.sh` (project/workspace only).
- **MCP output files:** `.vscode/mcp.json`, optional `~/.copilot/mcp-config.json` (global only when probe passes: `copilot` binary or `~/.copilot/` present).
- **Global scope support:** `yes` (probe-gated).

## Pi

**Surface summary:** emits agents, skills, and prompt templates under `.pi/`, plus extension-based safety hooks. No MCP config is written.

- **Root instructions:** `AGENTS.md`.
- **Config directory:** `.pi/`
- **Config tree (key files):**

```text
.
├── AGENTS.md
└── .pi/
    ├── agents/<agent>.md
    ├── skills/<skill>/SKILL.md
    ├── prompts/<prompt>.md
    └── extensions/*.ts
```

- **Agents surface:** `.pi/agents/<name>.md` (Pi subagent extension reads markdown agent definitions).
- **Skills surface:** `.pi/skills/<name>/SKILL.md`.
- **Commands / prompts / chat modes:** no commands or chat modes; prompt templates at `.pi/prompts/*.md`.
- **Hook runtime surface:** none as a `.pi/hooks` directory; safety hooks ship as Pi extensions at `.pi/extensions/*.ts`.
- **MCP output files:** none.
- **Global scope support:** `no`.


## OMP (beta)

**Surface summary:** emits agents, skills, commands, prompts, hook factories, and MCP.

- **Root instructions:** `AGENTS.md`.
- **Config directory:** `.omp/`
- **Config tree (key files):**

```text
.
├── AGENTS.md
└── .omp/
    ├── agents/<agent>.md
    ├── skills/<skill>/SKILL.md
    ├── commands/<command>.md
    ├── prompts/<prompt>.md
    ├── hooks/pre/<factory>.ts
    └── mcp.json
```

- **Agents surface:** `.omp/agents/<name>.md`.
- **Skills surface:** `.omp/skills/<name>/SKILL.md`.
- **Commands / prompts / chat modes:** commands in `.omp/commands`; prompts in `.omp/prompts`; no chat modes.
- **Hook runtime surface:** `.omp/hooks/pre/<name>.ts`.
- **MCP output files:** `.omp/mcp.json`.
- **Global scope support:** `yes`.
- **Maturity note:** **beta** (docs snapshot verification incomplete).

## Antigravity (beta)

**Surface summary:** emits minimal `.gemini` settings + hook surface and `.agents`-scoped Agent Skills; MCP is written to user config.

- **Root instructions:** `AGENTS.md`.
- **Config directory:** `.gemini/`
- **Config tree (key files):**

```text
.
├── AGENTS.md
└── .gemini/
    ├── settings.json
    └── hooks/lazyai/
        ├── block-destructive-shell.sh
        └── objective-workflow-gate.sh

.agents/
├── hooks.json
└── skills/<skill>/SKILL.md
```

- **Agents surface:** none.
- **Skills surface:** `.agents/skills/<name>/SKILL.md`.
- **Commands / prompts / chat modes:** none.
- **Hook runtime surface:** `.gemini/hooks/lazyai/*.sh` and `.agents/hooks.json`.
- **MCP output files:** `~/.gemini/config/mcp_config.json`.
- **Global scope support:** `no`.
- **Maturity note:** **beta** (partially JS-rendered docs snapshot incomplete).

## Kiro

**Surface summary:** emits agents, skills, and prompt templates under `.kiro/`, plus MCP config; no commands, chat modes, or hook runtime.

- **Root instructions:** `AGENTS.md`.
- **Config directory:** `.kiro/`
- **Config tree (key files):**

```text
.
├── AGENTS.md
└── .kiro/
    ├── agents/<agent>.md
    ├── skills/<skill>/SKILL.md
    ├── prompts/<prompt>.md
    └── settings/mcp.json
```

- **Agents surface:** `.kiro/agents/<name>.md` (Kiro CLI v3 custom agent profiles).
- **Skills surface:** `.kiro/skills/<name>/SKILL.md`.
- **Commands / prompts / chat modes:** prompts at `.kiro/prompts/*.md`; no commands or chat modes.
- **Hook runtime surface:** none.
- **MCP output files:** `.kiro/settings/mcp.json`.
- **Global scope support:** `yes`.

## How outputs are generated

`lazyai-cli compile` regenerates MCP output from `./.ai/mcp.json` for configured targets. The emit and surface rules come from:

- `packages/cli/internal/adapter/output_mapping.go`
- `packages/cli/internal/adapter/output_contract.go`
- `packages/cli/internal/adapter/capabilities.go`
- `packages/cli/internal/adapter/<tool>.go`

For baseline target-level behavior, see [../concepts/tools.md](../concepts/tools.md).