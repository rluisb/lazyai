# LazyAI

A CLI-first AI operating system for software teams. Define your AI setup once in a canonical format, then compile it to any supported AI tool.

`lazyai-cli` uses a **canonical source → compile** model:

1. `init` scaffolds a tool-agnostic canonical layer under `.ai/`
2. You edit rules, agents, and templates in one place
3. `compile` generates tool-native files (`.opencode/`, `.claude/`, `.github/`, `.vscode/`)
4. `update` refreshes managed files from the bundled library
5. `doctor` checks health and drift

Learn more in [How It Works](docs/concepts/how-it-works.md).

---

## Commands

### Session Management
Track AI agent sessions with SQLite persistence:

```bash
lazyai-cli session start "Implement auth feature"
# → Session started: ses_1234567890

lazyai-cli session list
# → 🟢 ses_1234567890 | Implement auth feature | 2026-05-24T00:08:16Z

lazyai-cli session show ses_1234567890
# → Shows session details and dispatch history

lazyai-cli session end ses_1234567890
# → ✅ Session ended
```

### Health Checks
Validate environment before work:

```bash
lazyai-cli doctor
# → Checks: sqlite3, git, jq, bash, ollama, openai, disk space

lazyai-cli doctor --json
# → Machine-readable output for CI integration
```

### Audit Trail
Immutable hash-chained ledger for accountability:

```bash
lazyai-cli ledger init
# → Initializes .specify/ledger.jsonl

lazyai-cli ledger append dispatch "agent=builder task=auth"
# → Appends event with SHA-256 hash

lazyai-cli ledger verify
# → Verifies chain integrity

lazyai-cli ledger show 5
# → Shows last 5 entries
```

### Validation
Check agent and skill file structure:

```bash
lazyai-cli validate agents
# → Checks dispatch parameters, tool schemas, common mistakes

lazyai-cli validate skills
# → Checks skill structure
```

### Task Queue
SQLite-backed task queue with atomic claiming:

```bash
lazyai-cli task create "Implement login page"
# → ✅ Task created: task_1234567890

lazyai-cli task list
# → Shows all tasks with status

lazyai-cli task claim task_1234567890
# → ✅ Task claimed (atomic, prevents duplicates)

lazyai-cli task complete task_1234567890
# → ✅ Task completed
```

### Agent Message Bus
SQLite-based messaging between agents:

```bash
lazyai-cli message send builder "Need help" "Can you review the auth code?"
# → ✅ Message sent: msg_1234567890

lazyai-cli message recv builder
# → Shows unread messages for builder

lazyai-cli message broadcast "All hands" "System update at 2pm"
# → ✅ Broadcast sent to 5 agents
```

### Metrics Dashboard
Track performance and generate dashboards:

```bash
lazyai-cli metrics list
# → Shows recent quality metrics

lazyai-cli metrics export
# → Exports to Prometheus format (metrics.prom)

lazyai-cli metrics dashboard
# → Generates HTML dashboard (dashboard.html)
```

### Memory Vault
Long-term institutional memory:

```bash
lazyai-cli memory save "Always test database migrations" --type lesson --tags database
# → ✅ Memory saved: 20260523_225222_lesson.md

lazyai-cli memory list
# → Shows all saved memories

lazyai-cli memory search database
# → Searches memories by content
```

### Evaluation Harness
Measure agent quality over time:

```bash
lazyai-cli eval list
# → Shows available evaluation suites

lazyai-cli eval run agent-quality
# → Runs evaluation suite
```

### Workflow Execution
Execute structured workflows:

```bash
# Workflows are defined in .opencode/workflows/*.yaml
# See .opencode/workflows/rpi.yaml for example
```

---

## Sidecar (Optional)

LazyAI can keep your docs, specs, and plans in a dedicated **sidecar** directory instead of inside each project. This is useful when you want a single knowledge base shared across workspaces, or when you prefer to keep planning artifacts outside version control.

### What sidecar means in LazyAI

A sidecar is a separate directory on disk that stores:
- `docs/` — documentation and guides
- `specs/` — feature specifications and ADRs
- `plans/` — execution plans and task breakdowns

When a sidecar is configured, LazyAI resolves these directories from the sidecar path instead of the project/workspace root. If no sidecar is configured, LazyAI falls back to its default behavior (docs/specs/plans live in the current scope root).

### Scope behavior

Sidecar configuration can live at three levels, with **workspace** as the primary use case:

| Scope | Config file | Priority |
|---|---|---|
| **Workspace** | `~/.lazyai/workspaces.yaml` (active workspace entry) | Highest |
| **Project** | `<project-root>/.lazyai-sidecar.yaml` | Middle |
| **Global** | `~/.lazyai/sidecar.yaml` | Lowest |

Resolution follows the chain: **workspace → project → global → default**. A workspace sidecar always wins over a project sidecar; a project sidecar wins over global; if none are configured, LazyAI uses the scope default.

**Workspace scope (recommended):**
- Best for multi-repo teams with a planning repo
- The active workspace entry in `workspaces.yaml` carries the sidecar block
- All projects in that workspace share the same sidecar by default

**Project scope:**
- Best when one repo needs its own isolated docs/specs/plans
- Create `.lazyai-sidecar.yaml` in the project root

**Global scope:**
- Best for personal defaults across all projects
- Set once in `~/.lazyai/sidecar.yaml`

### Commands

```bash
# Initialize a sidecar at a scope
lazyai-cli sidecar init --scope workspace --path /Users/me/kb/my-workspace

# Show resolved paths for the current scope
lazyai-cli sidecar status
# → Scope: workspace | Config Level: workspace
# → Docs:  /Users/me/kb/my-workspace/docs
# → Specs: /Users/me/kb/my-workspace/specs
# → Plans: /Users/me/kb/my-workspace/plans

# Attach a sidecar to the active workspace or project
lazyai-cli sidecar attach --path /tmp/kb

# Detach (remove) the sidecar configuration
lazyai-cli sidecar detach

# Validate sidecar paths exist and are writable
lazyai-cli sidecar doctor
```

### Optional fallback behavior

Sidecar is **always optional**. If you never run `sidecar init`, LazyAI behaves exactly as it does today:
- `project` scope → docs/specs/plans live in the project root
- `workspace` scope → docs/specs/plans live in the workspace (planning repo) root
- `global` scope → docs/specs/plans live in `~/.lazyai/`

No sidecar configured = no errors, no warnings, no behavior change.

### Explicit exclusions

- **No Skeeper integration.** The sidecar is purely local. There is no `skeeper` field, no provider abstraction, and no remote sync.
- **No content migration.** `sidecar init` does not move existing docs/specs/plans.
- **No multi-sidecar.** One sidecar per scope level.
- **No auto-discovery.** Sidecars are explicitly configured, not detected from parent directories or environment variables.

---

## Supported Tools

- [OpenCode](docs/concepts/tools.md#opencode)
- [Claude Code](docs/concepts/tools.md#claude-code)
- [GitHub Copilot](docs/concepts/tools.md#github-copilot)

---

## Documentation

- **Official docs:** <https://rluisb.github.io/lazyai/>
- **GitHub Wiki:** <https://github.com/rluisb/lazyai/wiki>

| Topic | Link |
|---|---|
| Quick Start | [docs/getting-started/quick-start.md](docs/getting-started/quick-start.md) |
| Installation | [docs/getting-started/installation.md](docs/getting-started/installation.md) |
| How It Works | [docs/concepts/how-it-works.md](docs/concepts/how-it-works.md) |
| Scopes | [docs/concepts/scopes.md](docs/concepts/scopes.md) |
| Presets | [docs/concepts/presets.md](docs/concepts/presets.md) |
| Tools | [docs/concepts/tools.md](docs/concepts/tools.md) |
| CLI Reference | [docs/cli/reference.md](docs/cli/reference.md) |
| MCP Integration | [docs/integration/mcp.md](docs/integration/mcp.md) |
| Orchestration | [docs/integration/orchestration.md](docs/integration/orchestration.md) |
| Contributing | [docs/development/contributing.md](docs/development/contributing.md) |
| Release Process | [docs/development/release.md](docs/development/release.md) |
| FAQ | [docs/troubleshooting/faq.md](docs/troubleshooting/faq.md) |

---

## Development

Requirements:

- Go 1.26+

```bash
cd packages/cli && go test ./...
cd ../orchestrator && go test ./...
cd ../diffviewer && go test ./...
```

Read the full [Contributing guide](docs/development/contributing.md).

---

## License

MIT. See [LICENSE](LICENSE).
