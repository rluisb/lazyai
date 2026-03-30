# @teachable/ai-setup

Scaffold an AI-ready repository in one command.

`@teachable/ai-setup` creates a consistent starting point for AI-assisted development with shared docs, templates, rules, root agent instructions, and tool-specific setup for **five AI coding tools**: Pi, OpenCode, Claude Code, Gemini CLI, and GitHub Copilot.

> Note: this project is currently intended for private/internal distribution, not public npm publishing.

## Why this exists

Setting up an AI-friendly repo by hand is repetitive:

- copying docs and templates
- creating the same folder structure over and over
- wiring tool-specific directories manually
- keeping shared files updated without clobbering custom edits

This CLI gives you a repeatable baseline and tracks managed files in `.ai-setup.json` so updates and integrity checks stay safe.

## Supported tools

### Pi (Claude Code)

**Tool ID:** `pi` · **Directory:** `.pi/` · **Root file:** `CLAUDE.md`

Pi is an internal Claude Code wrapper that uses a `.pi/` directory for agents, templates, and skills. When selected, the CLI creates:

- `.pi/agents/` — agent definition files
- `.pi/templates/` — prompt templates
- `.pi/skills/` — reusable skill definitions
- `CLAUDE.md` — root instruction file loaded by Claude Code at session start

### OpenCode

**Tool ID:** `opencode` · **Directory:** `.opencode/` · **Root file:** `AGENTS.md`

OpenCode uses an `.opencode/` directory for agents and commands. When selected, the CLI creates:

- `.opencode/agents/` — agent definition files
- `.opencode/commands/` — custom command definitions
- `AGENTS.md` — root instruction file for OpenCode

### Claude Code

**Tool ID:** `claude-code` · **Directory:** `.claude/` · **Root file:** `CLAUDE.md`

[Claude Code](https://docs.anthropic.com/en/docs/claude-code) reads a `CLAUDE.md` file at project root for persistent instructions. It also supports `.claude/rules/` for path-specific rules using YAML frontmatter with glob patterns. When selected, the CLI creates:

- `.claude/` — Claude Code configuration directory
- `.claude/rules/` — path-specific instruction rules (e.g., `*.test.ts` files get testing conventions)
- `CLAUDE.md` — root instruction file with project context, conventions, and `@file` imports

Key features of the Claude Code setup:
- `@docs/rules/`, `@docs/standards/`, `@docs/context/` imports for structured project knowledge
- Path-specific rules in `.claude/rules/` with `paths:` YAML frontmatter
- Conflict detection — if `CLAUDE.md` already exists (e.g., from Pi), you'll be prompted to replace it

### Gemini CLI

**Tool ID:** `gemini` · **Directory:** `.gemini/` · **Root file:** `GEMINI.md`

[Gemini CLI](https://github.com/google-gemini/gemini-cli) reads a `GEMINI.md` file at project root for persistent context. It supports hierarchical context loading and `@file.md` imports. When selected, the CLI creates:

- `.gemini/` — Gemini CLI configuration directory
- `GEMINI.md` — root instruction file with project context and `@file.md` imports

Key features of the Gemini CLI setup:
- `@docs/rules/`, `@docs/standards/`, `@docs/context/` imports for structured project knowledge
- Hierarchical context loading (global → workspace → project)
- Compatible with `.geminiignore` for excluding content from context
- `/memory` command support for session memory management

### GitHub Copilot

**Tool ID:** `copilot` · **Directory:** `.github/` · **Root file:** `.github/copilot-instructions.md`

[GitHub Copilot](https://docs.github.com/en/copilot) reads `.github/copilot-instructions.md` for repository-wide instructions. It also supports path-specific instructions and reusable prompt files. When selected, the CLI creates:

- `.github/copilot-instructions.md` — repository-wide instruction file
- `.github/instructions/` — directory for path-specific instruction files (using `applyTo:` YAML frontmatter)

Key features of the Copilot setup:
- Repository-wide instructions applied to all Copilot interactions
- Path-specific rules via `.github/instructions/NAME.instructions.md` with `applyTo:` glob patterns
- Compatible with `.github/prompts/` for reusable prompt files (`.prompt.md`)

## Requirements

- Node.js 18+
- npm

## Install

### Internal/private use (recommended for now)

Install directly from the private GitHub repository:

```bash
npm install -g git+ssh://git@github.com/ricardoborges-teachable/ai-setup.git
```

You can also run it without a publish step using the GitHub repo reference:

```bash
npx github:ricardoborges-teachable/ai-setup init
```

### Local development

```bash
git clone git@github.com:ricardoborges-teachable/ai-setup.git
cd ai-setup
npm install
npm run build
node ./dist/index.js --help
```

## Step-by-step setup guide

### 1. Install the CLI

```bash
npm install -g git+ssh://git@github.com/ricardoborges-teachable/ai-setup.git
```

### 2. Initialize your project (interactive)

Navigate to your project root and run:

```bash
ai-setup init
```

You'll be guided through three prompts:

1. **Setup type** — choose `project` (single repo) or `workspace` (monorepo)
2. **Tool selection** — pick one or more AI tools:
   - Pi (Claude Code) — `.pi/` + `CLAUDE.md`
   - OpenCode — `.opencode/` + `AGENTS.md`
   - Claude Code — `.claude/` + `CLAUDE.md`
   - Gemini CLI — `.gemini/` + `GEMINI.md`
   - GitHub Copilot — `.github/` + `copilot-instructions.md`
3. **Project name** — used in generated templates (defaults to directory name)

### 3. Initialize your project (non-interactive)

For CI or scripted setups, pass all options as flags:

```bash
# Single tool
ai-setup init --type project --tools claude-code --name my-repo --no-interactive

# Multiple tools
ai-setup init --type project --tools claude-code,gemini,copilot --name my-repo --no-interactive
```

Available tool IDs: `pi`, `opencode`, `claude-code`, `gemini`, `copilot`

### 4. Fill in placeholders

After scaffolding, open the generated root instruction files and replace `[YOUR_*]` placeholders with your project-specific information:

- `[YOUR_PROJECT_NAME]` — your project name
- `[YOUR_TECH_STACK]` — e.g., "TypeScript, React, Node.js"
- `[YOUR_KEY_DIRECTORIES]` — e.g., "src/ — application code, tests/ — test files"
- `[YOUR_ARCHITECTURE_PATTERNS]` — e.g., "MVC, Repository pattern"
- `[YOUR_CODING_CONVENTIONS]` — e.g., "ESLint + Prettier, camelCase for variables"
- `[YOUR_TEST_COMMANDS]` — e.g., "npm test, npm run test:e2e"
- `[YOUR_BUILD_COMMANDS]` — e.g., "npm run build"

### 5. Add a tool later

Already initialized? Add more tools at any time:

```bash
ai-setup add claude-code
ai-setup add gemini
ai-setup add copilot
```

If a root file already exists (e.g., `CLAUDE.md`), you'll be prompted to confirm replacement.

### 6. Keep files up to date

When the library templates evolve, refresh your managed files:

```bash
ai-setup update
```

This updates unchanged tracked files, skips customized ones, and adds any newly expected files.

### 7. Verify setup integrity

Check that managed files haven't been corrupted or accidentally deleted:

```bash
ai-setup doctor
```

## Commands

```bash
ai-setup --help
```

Available commands:

| Command | Description |
|---------|-------------|
| `init` | Scaffold docs, templates, and tool files |
| `add <tool>` | Add a tool to an existing setup (`pi`, `opencode`, `claude-code`, `gemini`, `copilot`) |
| `update` | Refresh managed files while skipping customized files |
| `doctor` | Verify tracked file integrity |
| `status` | Reserved for future implementation |

## Example output

Running `ai-setup init` with all tools selected produces:

```text
.
├── .ai-setup.json
├── AGENTS.md                          # opencode
├── CLAUDE.md                          # pi or claude-code
├── GEMINI.md                          # gemini
├── CODEOWNERS
├── docs/
│   ├── adrs/
│   ├── bugfixes/
│   ├── context/
│   ├── features/
│   ├── memory/
│   ├── refactors/
│   ├── rules/
│   ├── standards/
│   ├── tech-debt/
│   └── templates/
├── .claude/                           # claude-code
│   └── rules/
├── .gemini/                           # gemini
├── .github/                           # copilot
│   ├── copilot-instructions.md
│   └── instructions/
├── .opencode/                         # opencode
│   ├── agents/
│   └── commands/
└── .pi/                               # pi
    ├── agents/
    ├── skills/
    └── templates/
```

## What `init` creates

- **`docs/`** structure for features, bugfixes, refactors, tech debt, ADRs, memory, standards, templates, rules, and context
- **Shared files** such as `CODEOWNERS`, compliance docs, and a knowledge map
- **Root instructions** (based on selected tools):
  - `CLAUDE.md` when `pi` or `claude-code` is selected
  - `AGENTS.md` when `opencode` is selected
  - `GEMINI.md` when `gemini` is selected
  - `.github/copilot-instructions.md` when `copilot` is selected
- **Tool folders** (based on selected tools):
  - `.pi/agents`, `.pi/templates`, `.pi/skills`
  - `.opencode/agents`, `.opencode/commands`
  - `.claude/`, `.claude/rules/`
  - `.gemini/`
  - `.github/`, `.github/instructions/`
- **`.ai-setup.json`** manifest with tracked file hashes

## Safe updates

`update` uses `.ai-setup.json` to determine what can be refreshed:

- unchanged tracked files are updated
- customized tracked files are skipped
- newly expected managed files are added

This lets the library evolve without overwriting local edits.

## File conflict handling

When adding a tool that creates a file that already exists:

- **Root instruction files** (e.g., `CLAUDE.md`, `GEMINI.md`) — you'll be prompted: _"File already exists. Replace?"_
- **Tool directory files** (e.g., agents) — you'll be prompted to confirm replacement

This prevents accidental overwrites when multiple tools share root files (e.g., both Pi and Claude Code create `CLAUDE.md`).

## Development

```bash
npm install          # install dependencies
npm run build        # compile TypeScript
npm run typecheck    # type-check without emitting
npm run test         # run tests
npm run test:watch   # run tests in watch mode
```

## Contributing

1. Create a branch from `main` for your change.
2. Keep changes scoped to one behavior when possible.
3. Add or update tests for any behavioral change.
4. Run:
   - `npm run test`
   - `npm run typecheck`
   - `npm run build`
5. Open a PR with:
   - problem statement
   - approach
   - verification output

## Roadmap

Good next improvements:

- implement `status` command
- expand negative-path CLI coverage
- add per-tool integration tests
- add `.geminiignore` and `.github/prompts/` scaffolding
- support `remove <tool>` to cleanly uninstall a tool's files

## Architecture decisions

- [ADR-001: TypeScript CLI with @clack/prompts](docs/adrs/001-typescript-clack-cli.md)

## License

MIT. See [LICENSE](LICENSE).