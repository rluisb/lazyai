# @ricardoborges-teachable/ai-setup

Scaffold an AI-ready development environment with a single command.

## Architecture

ai-setup uses a **canonical source → compile** model:

1. `ai-setup init` creates a tool-agnostic `.ai/` directory as the source of truth
2. `ai-setup compile` transforms `.ai/` into tool-native directories

### `.ai/` Structure

```text
.ai/
├── constitution/       # Project principles, constraints, quality gates
│   ├── constitution.md
│   ├── constraints.md
│   ├── quality-gates.md
│   └── uncertainty.md
├── memory/             # AI memory and context persistence
│   ├── decisions/      # (workspace only)
│   ├── handoffs/
│   ├── patterns/       # (workspace only)
│   └── projects/       # (workspace only)
└── [agents, skills, prompts, templates, rules — coming soon]
```

`.ai-setup.json` is the manifest file created in the project root that tracks managed files, hashes, and setup metadata.

## Scopes

| Scope | Target | Use Case |
|-------|--------|----------|
| `global` | `~/.ai/` + native tool global paths | Personal defaults across all projects |
| `workspace` | Planning repo only | Multi-repo team with shared planning |
| `project` | Current directory | Self-contained single repository |

### Global Scope

- Creates `~/.ai/` with canonical structure
- Compiles to `~/.config/opencode/` (opencode) and `~/.claude/` (claude-code)
- Copilot, Gemini, Pi don't support file-based global config — fallback to project scope

### Workspace Scope

- All setup goes ONLY in the planning repo
- Referenced repos are listed in config, never modified
- Scans referenced repos to detect type (Rails, Next.js, Go, etc.)

### Project Scope

- Self-contained in the current directory
- Layers on top of global `~/.ai/` if it exists

## Supported Tools

| Tool | Config Dir | Skills Dir | Root File | Global Support |
|------|-----------|------------|-----------|----------------|
| opencode | `.opencode/` | `.opencode/skills/<name>/SKILL.md` | `AGENTS.md` | ✅ |
| claude-code | `.claude/` | `.claude/skills/<name>/SKILL.md` | `CLAUDE.md` | ✅ |
| codex | `.codex/` | `.codex/skills/<name>/SKILL.md` | `AGENTS.md` | ❌ |
| copilot | `.github/` | `.github/prompts/<name>.prompt.md` | `copilot-instructions.md` | ❌ |
| gemini | `.gemini/` | `.gemini/skills/<name>/SKILL.md` | `GEMINI.md` | ❌ |
| pi | `.pi/` | `.pi/skills/` | `INSTRUCTIONS.md` | ❌ |

## Usage

### Interactive setup

```bash
npx @ricardoborges-teachable/ai-setup init
```

### Non-interactive

```bash
# Project scope
npx @ricardoborges-teachable/ai-setup init --scope project --tools opencode,claude-code --name my-repo --no-interactive

# Global scope
npx @ricardoborges-teachable/ai-setup init --scope global --tools opencode,claude-code --no-interactive

# Workspace scope
npx @ricardoborges-teachable/ai-setup init --scope workspace --tools opencode --name my-workspace --planning-repo ./planning --repos ../app1,../app2 --no-interactive
```

### Compile (re-generate tool dirs from .ai/)

```bash
npx @ricardoborges-teachable/ai-setup compile
npx @ricardoborges-teachable/ai-setup compile --tools opencode --force
npx @ricardoborges-teachable/ai-setup compile --scope global
```

### Commands

| Command | Description |
|---------|-------------|
| `init` | Run the setup wizard (creates `.ai/` + compiles to tool dirs) |
| `compile` | Re-compile `.ai/` artifacts to tool-native directories |
| `import [path]` | Detect and migrate an existing AI tool setup |
| `add <tool>` | Add a tool adapter to existing setup |
| `update` | Reconcile managed files with current library templates |
| `doctor` | Verify tracked files against manifest hashes |
| `create` | Scaffold a new template document |
| `eject` | Remove tracking but keep generated files |
| `status` | Show current setup status |

## Migration engine

The migration engine helps you adopt `ai-setup` without throwing away an existing AI-tooling workflow.

It currently supports detecting and importing setups from:

- OpenCode
- Claude Code
- Pi
- Gemini CLI
- GitHub Copilot

### Common migration flows

```bash
# Preview a migration plan for the current repo
npx @ricardoborges-teachable/ai-setup import --preview

# Migrate another repo into the current working tree
npx @ricardoborges-teachable/ai-setup import ../older-project --yes

# Use the alias command
npx @ricardoborges-teachable/ai-setup migrate --strategy preserve

# Run init, but import an existing setup first
npx @ricardoborges-teachable/ai-setup init --migrate --from ../legacy-ai-config

# Check drift after migration
npx @ricardoborges-teachable/ai-setup doctor --migration-check --verbose
```

### Merge strategies

| Strategy | When to use it | Behavior |
|---|---|---|
| `smart` | Default choice when you want merge help | Builds a plan, attempts a 3-way merge, and stops when manual review is needed |
| `preserve` | You want to keep current files wherever there is overlap | Prefers existing content and adds new `ai-setup` files around it |
| `replace` | You want a clean `ai-setup` baseline | Replaces overlapping files and creates backups first |
| `append` | You want combined content where the parser supports it | Appends or combines content for supported file types |

### What the CLI does for you

1. Scans for supported markers such as `AGENTS.md`, `CLAUDE.md`, `GEMINI.md`, `.opencode/`, `.claude/`, `.pi/`, `.gemini/`, and GitHub Copilot instruction files.
2. Builds a migration plan before applying changes.
3. Shows conflicts before execution.
4. Creates backups unless you pass `--skip-backup`.
5. Writes or updates `.ai-setup.json` so you can run drift checks afterwards.

### Documentation and examples

- Full guide: [`specs/features/002-migration-engine/README.md`](specs/features/002-migration-engine/README.md)
- Example migration: [`specs/features/002-migration-engine/examples/opencode-to-ai-setup.md`](specs/features/002-migration-engine/examples/opencode-to-ai-setup.md)
- Example migration: [`specs/features/002-migration-engine/examples/claude-code-to-ai-setup.md`](specs/features/002-migration-engine/examples/claude-code-to-ai-setup.md)
- Community parser guide: [`specs/features/002-migration-engine/COMMUNITY-PARSERS.md`](specs/features/002-migration-engine/COMMUNITY-PARSERS.md)

## Update + conflict behavior

`update` builds the expected managed file set from the current library and selected tools, then resolves conflicts file-by-file.

Behavior summary:

- **Tracked and unchanged** (hash matches): overwritten with latest template
- **Tracked and customized** (hash differs): prompt to overwrite; overwrite creates backup (`--force` will automatically overwrite and backup)
- **Existing untracked collision**: prompt to replace; replace creates backup
- **Newly expected files**: created and added to `.ai-setup.json`
- **Tracked but missing files**: reported as missing (not auto-recreated)

Backups are written under `.ai-setup-backup/` preserving relative paths.

## Requirements

- Node.js `>=18`
- npm

## Development

```bash
git clone git@github.com:ricardoborges-teachable/ai-setup.git
cd ai-setup
npm install
npm run build
npm run test
```

## License

MIT. See [LICENSE](LICENSE).
