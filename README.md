# LazyAI

![Go >=1.26](https://img.shields.io/badge/go-%3E%3D1.26-00ADD8?logo=go&logoColor=white)

Scaffold a canonical, multi-tool AI development environment from one CLI, with optional orchestration scaffolding and MCP runtime integration.

`lazyai-cli` initializes and manages tool-agnostic project rules, agents, skills, and templates, then compiles them into native formats for OpenCode, Claude Code, and GitHub Copilot. LazyAI is distributed as Go modules under `github.com/rluisb/lazyai`.

---

## Quick Start

```bash
go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@latest
lazyai-cli init
```

This launches an interactive wizard that asks for scope, tools, preset, and optional MCP servers.

For a non-interactive project setup:

```bash
lazyai-cli init \
  --scope project \
  --tools opencode,claude-code,copilot \
  --name my-app \
  --preset standard \
  --no-interactive
```

Read the full [Quick Start guide](docs/getting-started/quick-start.md) for workspace and global scope examples.

---

## Installation

- **CLI**: `go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@latest`
- **Orchestrator MCP runtime**: `go install github.com/rluisb/lazyai/packages/orchestrator/cmd/lazyai-orchestrator@latest`
- **Diff viewer utility**: `go install github.com/rluisb/lazyai/packages/diffviewer/cmd/lazyai-diffviewer@latest`
- **Clone for development**: `git clone git@github.com:rluisb/lazyai.git`

See [Installation](docs/getting-started/installation.md) for details.

---

## How It Works

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
# → Checks skill structure (coming soon)
```


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
