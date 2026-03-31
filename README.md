# @teachable/ai-setup

Scaffold an AI-ready repository in one command.

`@teachable/ai-setup` generates a consistent AI collaboration baseline: project docs, docs-scoped `AGENTS.md` guides, workflow templates, root instruction files, and adapter-specific folders for supported tools.

> Internal/private distribution for now (not published as a public npm package).

## What it covers

- Structured `docs/` foundation for features, bugfixes, refactors, tech debt, ADRs, memory, prompts, rules, standards, and templates
- Rich docs-scoped agent guidance under `docs/*/AGENTS.md`
- Upgraded root templates for:
  - `AGENTS.md` (OpenCode)
  - `CLAUDE.md` (Pi / Claude Code)
  - `GEMINI.md` (Gemini CLI)
  - `.github/copilot-instructions.md` (GitHub Copilot)
- Skills + prompt templates scaffolded across all adapters
- Safe update/merge behavior with conflict prompts, backup creation, and `--force`

Managed files are tracked in `.ai-setup.json` with content hashes.

## Supported adapters

| Tool ID | Adapter | Root instruction file | Adapter output |
|---|---|---|---|
| `pi` | Pi (Claude Code wrapper) | `CLAUDE.md` | `.pi/agents`, `.pi/skills`, `.pi/templates` |
| `opencode` | OpenCode | `AGENTS.md` | `.opencode/agents`, `.opencode/commands`, `.opencode/templates` |
| `claude-code` | Claude Code | `CLAUDE.md` | `.claude/` agent files, `.claude/commands`, `.claude/templates`, `.claude/rules/` |
| `gemini` | Gemini CLI | `GEMINI.md` | `.gemini/` agent files, `.gemini/skills`, `.gemini/templates` |
| `copilot` | GitHub Copilot | `.github/copilot-instructions.md` | `.github/` agent files, `.github/prompts`, `.github/templates`, `.github/instructions/` |

### Notes on skills/prompts per adapter

- `library/skills/*.md` are copied as native skill/command files for Pi, OpenCode, Claude Code, and Gemini.
- For Copilot, skills are transformed into `.github/prompts/*.prompt.md` files with `mode: agent` frontmatter.
- `library/prompts/*.md` are copied to each adapter’s template directory.

## What `init` generates

### Shared docs and references

`init` creates and populates:

- `docs/AGENTS.md`
- `docs/features/AGENTS.md`
- `docs/bugfixes/AGENTS.md`
- `docs/refactors/AGENTS.md`
- `docs/tech-debt/AGENTS.md`
- `docs/rules/AGENTS.md`
- `docs/standards/AGENTS.md`
- `docs/templates/AGENTS.md`
- `docs/memory/AGENTS.md`
- `docs/adrs/AGENTS.md`
- `docs/memory/handoffs/` (handoff destination)
- `docs/prompts/local-examples/`:
  - `preflight-task-framing.md`
  - `react-trace-and-handoff.md`
  - `commit-message-pattern.md`
- Reference docs:
  - `docs/KNOWLEDGE_MAP.md`
  - `docs/compliance.md`
- Plus workflow/rules/template libraries from:
  - `library/templates/*` → `docs/templates/*`
  - `library/rules/*` → `docs/rules/*`

### Infra and tracking

- `CODEOWNERS`
- `.git/hooks/pre-commit` (if target already has a `.git` directory)
- `.ai-setup.json` manifest with tracked file hashes and file sources

## Prompt + context-engineering baseline

The generated root instruction templates include:

- Decision-tree style context loading
- Session start checks
- Reasoning / architecture-decision / trace protocols
- Confidence + verification gates
- Session compaction guidance
- Multi-session handoff protocol (`docs/memory/handoffs/`)
- Local prompt example references (`docs/prompts/local-examples/`)

This gives teams a consistent, docs-first workflow for Research → Plan → Implement loops without hard-coding tool-specific behavior into project code.

## Install

### Internal install (recommended)

```bash
npm install -g git+ssh://git@github.com/ricardoborges-teachable/ai-setup.git
```

### Run without global install

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

## Usage

### Interactive init

```bash
ai-setup init
```

### Non-interactive init

```bash
ai-setup init --type project --tools claude-code,gemini,copilot --name my-repo --no-interactive
```

### Force overwrite (with backups)

```bash
ai-setup init --type project --tools pi --name my-repo --no-interactive --force
```

### Add adapter to existing setup

```bash
ai-setup add opencode
```

### Refresh managed files

```bash
ai-setup update
ai-setup update --force
```

### Integrity check

```bash
ai-setup doctor
```

## Commands

| Command | Description |
|---|---|
| `init` | Scaffold docs, root instructions, and adapter files |
| `add <tool>` | Add one adapter (`pi`, `opencode`, `claude-code`, `gemini`, `copilot`) |
| `update` | Reconcile managed files with current library templates |
| `doctor` | Verify tracked files against `.ai-setup.json` hashes |
| `status` | Placeholder command (`coming soon`) |

## Update + conflict behavior

`update` builds the expected managed file set from the current library and selected tools, then resolves conflicts file-by-file.

Behavior summary:

- **Tracked and unchanged** (hash matches): overwritten with latest template
- **Tracked and customized** (hash differs): prompt to overwrite; overwrite creates backup
- **Existing untracked collision**: prompt to replace; replace creates backup
- **`--force`**: overwrite all existing managed targets, always creating backups first
- **Newly expected files**: created and added to `.ai-setup.json`
- **Tracked but missing files**: reported as missing (not auto-recreated)

Backups are written under `.ai-setup-backup/` preserving relative paths.

## Requirements

- Node.js `>=18`
- npm

## Development

```bash
npm run test
npm run typecheck
npm run build
```

## Architecture decisions

- [ADR-001: TypeScript CLI with @clack/prompts](docs/adrs/001-typescript-clack-cli.md)

## License

MIT. See [LICENSE](LICENSE).
