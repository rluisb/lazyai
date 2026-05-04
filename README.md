# @ricardoborges-teachable/ai-setup

![Node >=20.12.0](https://img.shields.io/badge/node-%3E%3D20.12.0-339933?logo=node.js&logoColor=white)

Scaffold a canonical, multi-tool AI development environment from one CLI, with optional orchestration scaffolding and MCP runtime integration.

`ai-setup` initializes and manages tool-agnostic project rules, agents, skills, and templates, then compiles them into native formats for OpenCode, Claude Code, and GitHub Copilot.

---

## Quick Start

```bash
npx github:ricardoborges-teachable/ai-setup init
```

This launches an interactive wizard that asks for scope, tools, preset, and optional MCP servers.

For a non-interactive project setup:

```bash
npx github:ricardoborges-teachable/ai-setup init \
  --scope project \
  --tools opencode,claude-code,copilot \
  --name my-app \
  --preset standard \
  --no-interactive
```

Read the full [Quick Start guide](docs/getting-started/quick-start.md) for workspace and global scope examples.

---

## Installation

- **From GitHub via npx** (recommended): `npx github:ricardoborges-teachable/ai-setup init`
- **Clone for development**: `git clone git@github.com:ricardoborges-teachable/ai-setup.git && npm install && npm run build && npm link`

See [Installation](docs/getting-started/installation.md) for details.

---

## How It Works

`ai-setup` uses a **canonical source → compile** model:

1. `init` scaffolds a tool-agnostic canonical layer under `.ai/`
2. You edit rules, agents, and templates in one place
3. `compile` generates tool-native files (`.opencode/`, `.claude/`, `.github/`, `.vscode/`)
4. `update` refreshes managed files from the bundled library
5. `doctor` checks health and drift

Learn more in [How It Works](docs/concepts/how-it-works.md).

---

## Supported Tools

- [OpenCode](docs/concepts/tools.md#opencode)
- [Claude Code](docs/concepts/tools.md#claude-code)
- [GitHub Copilot](docs/concepts/tools.md#github-copilot)

---

## Documentation

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

- Go 1.26+ (for the binary)
- Node.js >=20.12.0 and pnpm >=9.0.0 (for the monorepo)

```bash
pnpm install
pnpm run build
pnpm run test
pnpm run lint
```

Read the full [Contributing guide](docs/development/contributing.md).

---

## License

MIT. See [LICENSE](LICENSE).
