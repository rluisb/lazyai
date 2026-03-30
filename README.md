# @teachable/ai-setup

One-command scaffold for an AI-ready repository layout (docs, templates, rules, and tool-specific directories for Pi / OpenCode).

## Requirements

- Node.js 18+
- npm

## Install

### Global (typical CLI usage)

```bash
npm install -g @teachable/ai-setup
```

### Local development

```bash
npm install
npm run build
node ./dist/index.js --help
```

## Commands

```bash
ai-setup --help
```

Available commands:

- `init` — scaffold docs + tool files
- `add <tool>` — add `pi` or `opencode` to existing setup
- `update` — refresh managed files while skipping customized ones
- `doctor` — verify tracked file integrity
- `status` — placeholder command (coming soon)

## Usage

### Interactive init

```bash
ai-setup init
```

### Non-interactive init (CI/script friendly)

```bash
ai-setup init --type project --tools pi,opencode --name my-repo --no-interactive
```

### Add a new tool later

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

## What `init` creates

- `docs/` structure (`features`, `bugfixes`, `refactors`, `tech-debt`, `adrs`, `memory`, `standards`, `templates`, `rules`, `context`)
- shared files (`CODEOWNERS`, compliance and knowledge map docs)
- root file(s):
  - `CLAUDE.md` when `pi` is selected
  - `AGENTS.md` when `opencode` is selected
- tool folders:
  - `.pi/agents`, `.pi/templates`, `.pi/skills`
  - `.opencode/agents`, `.opencode/commands`
- manifest file: `.ai-setup.json` (tracks installed files + hashes)

## Development

```bash
npm run test
npm run typecheck
npm run build
```

## Contributing

1. Create a branch for your change.
2. Keep changes scoped to one task/behavior when possible.
3. Add or update tests for any behavioral change.
4. Run:
   - `npm run test`
   - `npm run typecheck`
   - `npm run build`
5. Open a PR with:
   - problem statement
   - approach
   - verification output

## Architecture decisions

- [ADR-001: TypeScript CLI with @clack/prompts](docs/adrs/001-typescript-clack-cli.md)
