# @teachable/ai-setup

Scaffold an AI-ready repository in one command.

`@teachable/ai-setup` creates a consistent starting point for AI-assisted development with shared docs, templates, rules, root agent instructions, and tool-specific setup for Pi and OpenCode.

## Why this exists

Setting up an AI-friendly repo by hand is repetitive:

- copying docs and templates
- creating the same folder structure over and over
- wiring tool-specific directories manually
- keeping shared files updated without clobbering custom edits

This CLI gives you a repeatable baseline and tracks managed files in `.ai-setup.json` so updates and integrity checks stay safe.

## What it does

- `init` — scaffold a fresh AI setup
- `add <tool>` — add Pi or OpenCode after initial setup
- `update` — refresh managed files while skipping customized ones
- `doctor` — verify tracked file integrity
- `status` — reserved for future implementation

## Requirements

- Node.js 18+
- npm

## Install

### Global install

```bash
npm install -g @teachable/ai-setup
```

### Local development

```bash
npm install
npm run build
node ./dist/index.js --help
```

## Quick start

### Interactive setup

```bash
ai-setup init
```

### Non-interactive setup

```bash
ai-setup init --type project --tools pi,opencode --name my-repo --no-interactive
```

### Add a tool later

```bash
ai-setup add opencode
```

### Refresh managed files

```bash
ai-setup update
```

### Verify setup integrity

```bash
ai-setup doctor
```

## Commands

```bash
ai-setup --help
```

Available commands:

- `init` — scaffold docs and tool files
- `add <tool>` — add `pi` or `opencode` to an existing setup
- `update` — refresh managed files while skipping customized files
- `doctor` — verify tracked file integrity
- `status` — placeholder command

## Example output

```text
.
├── .ai-setup.json
├── AGENTS.md              # when opencode is selected
├── CLAUDE.md              # when pi is selected
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
├── .opencode/
│   ├── agents/
│   └── commands/
└── .pi/
    ├── agents/
    ├── skills/
    └── templates/
```

## What `init` creates

- `docs/` structure for features, bugfixes, refactors, tech debt, ADRs, memory, standards, templates, rules, and context
- shared files such as `CODEOWNERS`, compliance docs, and a knowledge map
- root instructions:
  - `CLAUDE.md` when `pi` is selected
  - `AGENTS.md` when `opencode` is selected
- tool folders:
  - `.pi/agents`, `.pi/templates`, `.pi/skills`
  - `.opencode/agents`, `.opencode/commands`
- `.ai-setup.json` manifest with tracked file hashes

## Safe updates

`update` uses `.ai-setup.json` to determine what can be refreshed:

- unchanged tracked files are updated
- customized tracked files are skipped
- newly expected managed files are added

This lets the library evolve without overwriting local edits.

## Development

```bash
npm run test
npm run typecheck
npm run build
```

## Contributing

1. Create a branch for your change.
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

- implement `status`
- expand negative-path CLI coverage
- add `pi`-only and `opencode`-only integration tests
- add more tool adapters over time

## Architecture decisions

- [ADR-001: TypeScript CLI with @clack/prompts](docs/adrs/001-typescript-clack-cli.md)

## License

MIT. See [LICENSE](LICENSE).
