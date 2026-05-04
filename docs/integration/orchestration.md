# Orchestration

`ai-setup` can optionally scaffold orchestration definitions and register the `@ai-setup/orchestrator` MCP server. This is **opt-in**; if you never enable `orchestrator`, nothing about the existing `ai-setup` flow changes.

## Execution model

- **Local native agents** (Claude Code, OpenCode, Copilot) are the intended execution path.
- A2A is a config/seam only and is **not** remote/network execution by default.
- The `ai-setup-orchestrator` is a Go runtime invoked via `connect` so multiple MCP clients share a single daemon process.
- Released installs can download, verify, and cache matching prebuilt `ai-setup-orchestrator-*` assets from GitHub Releases.

## Enable during `init`

**Non-interactive**

```bash
ai-setup init \
  --scope project \
  --tools opencode,claude-code,copilot \
  --enable-servers orchestrator \
  --name my-app \
  --preset standard \
  --no-interactive
```

**Interactive wizard**

Run `ai-setup init` and select `orchestrator` when asked for optional MCP integrations.

## What gets scaffolded

When enabled, `ai-setup` copies the orchestration library into:

```text
.ai/orchestration/
├── chains/
├── teams/
├── workflows/
└── skills/
    ├── domains/
    └── modes/
```

Bundled defaults include:

- **Chains**: `feature`, `bugfix`, `review`, `refactor`, `tdd`, `onboard`
- **Teams**: `review-team`, `feature-team`, `assessment-team`
- **Workflows**: `rpi`, `tdd`, `refactor`, `code-review`, `incident-response`, `system-design`
- **Domains**: `backend`, `frontend`, `data`, `devops`, `security`
- **Modes**: `autonomous`, `junior`, `senior`

Project-local files with the same name override built-in library entries for `list` and `info`.

## CLI commands

### Create custom artifacts

```bash
ai-setup create domain payments --description "Payments domain constraints" --no-interactive
ai-setup create mode strict-review --description "High-friction approval mode" --no-interactive
ai-setup create workflow payments-review --chain feature --team review-team --no-interactive
```

### List orchestration artifacts

```bash
ai-setup list workflows
ai-setup list chains
ai-setup list teams
ai-setup list domains
ai-setup list modes
ai-setup list orchestration --json
```

### Inspect an item

```bash
ai-setup info feature
ai-setup info review-team
ai-setup info backend --json
```

### Orchestration namespace

```bash
ai-setup orchestration list workflows --json
ai-setup orchestration create domain payments --description "Payments domain" --no-interactive
ai-setup orchestration status --json
```

## Tool integration

When orchestration is enabled, `ai-setup` generates additional guidance files:

| Tool | Generated guidance |
|---|---|
| OpenCode | `.opencode/agents/orchestrator.md` |
| Claude Code | `.claude/agents/orchestrator.md` |
| GitHub Copilot | `.github/prompts/orchestrator.prompt.md` |

## Using the orchestrator in your CLI tool

1. Run `ai-setup init --enable-servers orchestrator`
2. Open the project root in your coding-agent CLI so it can see the generated MCP config and orchestration guidance files
3. Ask the tool to use the orchestrator for a specific task
4. The host tool does the execution while the orchestrator MCP server tracks chain/team/workflow state

Example prompts:

- `Use the orchestrator and start the feature chain for auth middleware.`
- `Use the orchestrator to build a review team for this PR.`
- `Use the orchestrator to show the current workflow status and budget.`

## Runtime details

- The orchestrator runtime source lives in `packages/orchestrator-go/`
- `ai-setup` scaffolds definitions and config; `ai-setup-orchestrator` owns runtime behavior
- The MCP server is the runtime boundary; there is no separate `ai-setup` command that runs a workflow end-to-end

## Legacy orchestration usage doc

The older standalone guide remains in the repository for historical reference. The page you are reading now is the canonical orchestration documentation for the docs site.
