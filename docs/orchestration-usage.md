# Orchestration Usage

This guide describes the **currently shipped** orchestration experience in `ai-setup`.

It is intentionally scoped to what exists in Phases 1–4 today:

- opt-in enablement through `ai-setup init`
- project-local orchestration definitions under `.ai/orchestration/`
- CLI support for `create`, `list`, `info`, and `orchestration ...`
- MCP server registration for the `@ai-setup/orchestrator` package
- tool-specific orchestrator guidance files for supported adapters

It does **not** describe speculative future UX or a separate `ai-setup` runtime runner command, because `ai-setup` is the scaffolding/configuration layer, not the orchestration runner itself.

---

## Mental model

Think of orchestration in this repo as three layers:

1. **`ai-setup` CLI**
   - scaffolds definitions
   - copies the built-in orchestration library into your project
   - registers the optional MCP server
   - generates tool-specific orchestrator guidance
2. **`.ai/orchestration/`**
   - your editable project-local source of truth for chains, teams, workflows, domains, and modes
3. **`ai-setup-orchestrator` runtime**
   - the Go binary that implements catalog loading, composition, state handling, persistence, and MCP tool handlers
   - built from `packages/orchestrator-go/`
   - invoked via `connect` so multiple MCP clients share a single daemon process

That means the adoption path is:

- enable orchestration during setup
- inspect or customize the generated definitions
- use your host CLI tool with the configured MCP server and orchestrator guidance

---

## Enable orchestration in `ai-setup`

### Non-interactive setup

```bash
ai-setup init \
  --scope project \
  --tools opencode,claude-code,copilot \
  --enable-servers orchestrator \
  --name my-app \
  --preset standard \
  --no-interactive
```

### Interactive wizard path

Run:

```bash
ai-setup init
```

Then, when the wizard asks which optional MCP integrations to enable, include:

```text
orchestrator
```

---

## What files are created

When `orchestrator` is enabled, `ai-setup` copies the orchestration library into:

```text
.ai/orchestration/
```

The scaffolded directory layout is:

```text
.ai/orchestration/
├── chains/
├── teams/
├── workflows/
└── skills/
    ├── domains/
    └── modes/
```

More explicitly, the top-level paths you can rely on are:

- `.ai/orchestration/chains/`
- `.ai/orchestration/teams/`
- `.ai/orchestration/workflows/`
- `.ai/orchestration/skills/domains/`
- `.ai/orchestration/skills/modes/`

The bundled defaults currently include:

- chains such as `feature`, `bugfix`, `review`, `refactor`, `tdd`, `onboard`
- teams such as `review-team`, `feature-team`, `assessment-team`
- workflows such as `rpi`, `tdd`, `refactor`, `code-review`, `incident-response`
- domain skills such as `backend`, `frontend`, `data`, `devops`, `security`
- mode skills such as `autonomous`, `junior`, `senior`

Project-local files with the same name override the built-in library entries for `list` and `info` lookups.

---

## CLI commands available now

### Create custom orchestration artifacts

You can generate new orchestration definitions with the normal `create` command:

```bash
ai-setup create domain payments --description "Payments domain constraints" --no-interactive
ai-setup create mode strict-review --description "High-friction approval mode" --no-interactive
ai-setup create workflow payments-review --chain feature --team review-team --no-interactive
```

Generated file locations:

- domain → `.ai/orchestration/skills/domains/<name>.md`
- mode → `.ai/orchestration/skills/modes/<name>.md`
- workflow → `.ai/orchestration/workflows/<name>.json`

### List orchestration artifacts

You can list orchestration categories directly:

```bash
ai-setup list workflows
ai-setup list chains
ai-setup list teams
ai-setup list domains
ai-setup list modes
ai-setup list orchestration --json
```

`list orchestration` is the aggregate view. It returns all orchestration categories together.

### Inspect an orchestration item

Use `info` with a workflow, chain, team, domain, or mode name:

```bash
ai-setup info feature
ai-setup info review-team
ai-setup info backend --json
ai-setup info rpi
```

### Use the orchestration namespace

The focused namespace is useful if you want orchestration-only commands:

```bash
ai-setup orchestration list workflows --json
ai-setup orchestration create domain payments --description "Payments domain" --no-interactive
ai-setup orchestration create workflow payments-review --chain feature --team review-team --no-interactive
ai-setup orchestration status --json
```

`ai-setup orchestration status` reports whether orchestration has been scaffolded plus project/library counts for workflows, chains, teams, domains, and modes.

---

## How the MCP server fits in

When you enable:

```bash
--enable-servers orchestrator
```

`ai-setup` adds the optional orchestrator server entry to the canonical MCP catalog in `.ai/mcp.json` and compiles it into supported tool-specific MCP config output.

At a high level, the configured server runs the managed Go binary:

```bash
ai-setup-orchestrator
```

The runtime source lives in:

```text
packages/orchestrator-go/
```

That package is where the repo keeps the orchestration runtime logic, including:

- catalog loading
- agent prompt composition
- chain/team/workflow runtime state machines
- persistence and handoff records
- MCP tool handlers and stdio server bootstrap

Important boundary:

- **`ai-setup`** scaffolds definitions and config
- **`ai-setup-orchestrator`** owns runtime behavior
- **your host CLI tool** is what actually calls the MCP tools during a session

> **A2A protocol:** Agent-to-Agent (A2A) orchestration is optional and currently deferred. The default execution model uses the native host CLI (e.g., Claude Code, OpenCode) directly. A2A can be enabled opt-in when it becomes available.

> **Packaging caveat:** Released installs can download, verify, and cache the matching prebuilt `ai-setup-orchestrator` binary when local source and a PATH binary are unavailable. Local-source managed setups still build from `packages/orchestrator-go/` and require `go` to be installed.

So there is no separate `ai-setup` command that “runs a workflow” end-to-end. The MCP server is the runtime boundary.

---

## Supported tool integration notes

When orchestration is enabled, `ai-setup` generates additional user-facing guidance files:

| Tool | Generated orchestration guidance |
|---|---|
| OpenCode | `.opencode/agents/orchestrator.md` |
| Claude Code | `.claude/agents/orchestrator.md` |
| GitHub Copilot | `.github/prompts/orchestrator.prompt.md` |

Notes:

- OpenCode, Claude Code, and GitHub Copilot receive compiled project-local MCP config when the server is enabled.
- Copilot receives a prompt artifact rather than an agent or skill directory.

---

## How to use the orchestrator MCP in supported tools

### Universal pattern

No matter which CLI tool you use, the practical flow is the same:

1. Run `ai-setup init --enable-servers orchestrator`
2. Open the project root in your coding-agent CLI so it can see the generated MCP config and orchestration guidance files
3. Ask the tool to use the orchestrator for a specific task
4. Let the host tool do the execution while the orchestrator MCP server tracks chain/team/workflow state

Example prompts that should work conceptually across tools:

- `Use the orchestrator and start the feature chain for auth middleware.`
- `Use the orchestrator to build a review team for this PR.`
- `Use the orchestrator to show the current workflow status and budget.`
- `Use the orchestrator to hand off this failed chain.`

### Claude Code

**What ai-setup generates**
- MCP config: `.mcp.json`
- Orchestrator guidance: `.claude/agents/orchestrator.md`

**How to use it**
- Start Claude Code from the project root
- Use the generated orchestrator agent naturally in chat
- Ask for chains, teams, workflow status, retries, escalation, or handoff

**Good fit for**
- orchestrator agent + native subagent dispatch
- more structured multi-step execution

**Example**

```text
Use the orchestrator agent and start the feature chain for the payments refactor.
```

### OpenCode

**What ai-setup generates**
- MCP config: `.opencode/opencode.jsonc`
- Orchestrator guidance: `.opencode/agents/orchestrator.md`

**How to use it**
- Start OpenCode from the project root
- Ask it to use the orchestrator for a specific chain/team/workflow action

**Good fit for**
- orchestrator agent + task-based coordination

**Example**

```text
Use the orchestrator to start the bugfix chain for the login failure.
```

### GitHub Copilot

**What ai-setup generates**
- MCP config: `.vscode/mcp.json`
- Orchestrator guidance: `.github/prompts/orchestrator.prompt.md`

**How to use it**
- Open the workspace in VS Code
- Use Copilot Chat / prompt flows that can read the generated prompt and workspace MCP config
- Ask it to use the orchestrator prompt for a specific operation

**Important limitation**
- This is prompt-oriented guidance, not a native agent runtime

**Example**

```text
Use the orchestrator prompt and walk me through the code-review workflow.
```

---

## Demo

For a reproducible walkthrough, see:

- [`demo/13-orchestration.tape`](../demo/13-orchestration.tape)

That demo covers:

1. `init --enable-servers orchestrator`
2. listing orchestration artifacts
3. creating a custom domain and workflow

---

## Contributor note

If you are contributing to orchestration internals rather than just using the feature, the design doc recommends keeping orchestration work isolated on a dedicated branch line:

```text
main
└── v2/orchestrator
    ├── feature/definitions-and-scaffold
    ├── feature/mcp-chains
    ├── feature/mcp-teams-workflows
    ├── feature/cli-agent-enhancements
    └── feature/docs-and-demos
```

See [`docs/orchestrator-design.md`](./orchestrator-design.md) §22 for the full rollout and migration guidance.
