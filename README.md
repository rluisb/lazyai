# @ricardoborges-teachable/ai-setup

Scaffold an AI-ready repository in one command.

`@ricardoborges-teachable/ai-setup` generates a consistent AI collaboration baseline: project docs, docs-scoped `AGENTS.md` guides, workflow templates, root instruction files, and adapter-specific folders for supported tools.

## What it covers

- Structured `docs/` foundation for features, bugfixes, refactors, tech debt, ADRs, memory, prompts, rules, standards, and templates
- Rich docs-scoped agent guidance under `docs/*/AGENTS.md`
- Pluggable Adapter Architecture supporting leading AI assistants
- Rich "Standard Library" of AI Artifacts (Rules, Skills, Agents, Templates)
- Safe update/merge behavior with conflict prompts, backup creation, and manifest tracking (`.ai-setup.json`).

## Supported adapters

| Tool ID | Adapter | Root instruction file | Adapter output |
|---|---|---|---|
| `pi` | Pi (Claude Code wrapper) | `CLAUDE.md` | `.pi/agents`, `.pi/skills`, `.pi/templates` |
| `opencode` | OpenCode | `AGENTS.md` | `.opencode/agents`, `.opencode/commands`, `.opencode/templates` |
| `claude-code` | Claude Code | `CLAUDE.md` | `.claude/` agent files, `.claude/commands`, `.claude/templates`, `.claude/rules/` |
| `gemini` | Gemini CLI | `GEMINI.md` | `.gemini/` agent files, `.gemini/skills`, `.gemini/templates` |
| `copilot` | GitHub Copilot | `.github/copilot-instructions.md` | `.github/` agent files, `.github/prompts`, `.github/templates`, `.github/instructions/` |

## Usage

### Interactive setup (Recommended)

```bash
npx @ricardoborges-teachable/ai-setup init
```

The interactive wizard will guide you through:
1. Identifying your project details.
2. Selecting which AI coding assistants you use.
3. Choosing which rules, agents, skills, and templates to install.
4. Handling any conflicts with existing files.

### Non-interactive init

```bash
npx @ricardoborges-teachable/ai-setup init --type project --tools opencode,pi --name my-repo --no-interactive
```

### Commands

| Command | Description |
|---|---|
| `init` | Run the setup wizard to scaffold the environment |
| `import [path]` | Detect and migrate an existing AI tool setup into `ai-setup` |
| `migrate [path]` | Alias for `import` |
| `add <tool>` | Add an adapter (`pi`, `opencode`, `claude-code`, `gemini`, `copilot`) to an existing setup |
| `update` | Reconcile managed files with current library templates (preserves/prompts on conflicts) |
| `doctor` | Verify tracked files against `.ai-setup.json` hashes to find drift or missing files |
| `create` | Scaffold a new template document (e.g., `npx ai-setup create adr`) |
| `eject` | Remove `.ai-setup.json` tracking but keep all generated files |
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

- Full guide: [`docs/features/002-migration-engine/README.md`](docs/features/002-migration-engine/README.md)
- Example migration: [`docs/features/002-migration-engine/examples/opencode-to-ai-setup.md`](docs/features/002-migration-engine/examples/opencode-to-ai-setup.md)
- Example migration: [`docs/features/002-migration-engine/examples/claude-code-to-ai-setup.md`](docs/features/002-migration-engine/examples/claude-code-to-ai-setup.md)
- Community parser guide: [`docs/features/002-migration-engine/COMMUNITY-PARSERS.md`](docs/features/002-migration-engine/COMMUNITY-PARSERS.md)

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
