# Contributing

## Repository structure

```text
ai-setup/
├── packages/
│   ├── ai-setup-ts/          # TypeScript CLI (legacy bootstrap)
│   ├── ai-setup-go/          # Go CLI binary
│   ├── orchestrator-go/      # Go orchestrator runtime
│   ├── orchestrator/         # Shared orchestrator packages
│   └── diffviewer/           # Diff viewer utility
├── library/                  # Bundled agents, skills, templates, rules
├── docs/                     # Documentation site
├── .github/workflows/        # CI, release, docs
└── README.md
```

## Requirements

- **Go 1.26+** (for the binary)
- **Node.js** `>=20.12.0` and **pnpm** `>=9.0.0` (for the monorepo)

## Local development

```bash
git clone git@github.com:ricardoborges-teachable/ai-setup.git
cd ai-setup
pnpm install
pnpm run build
pnpm run test
pnpm run lint
pnpm run typecheck
```

## Useful commands

```bash
pnpm run build      # build all packages
pnpm run test       # run test suite
pnpm run lint       # biome checks
pnpm run typecheck  # TypeScript no-emit check
pnpm run clean      # clean build artifacts
```

For the Go binary:

```bash
cd packages/ai-setup-go
go build -o ai-setup .
```

## Linking the CLI locally

```bash
pnpm --filter @ricardoborges-teachable/ai-setup link --global
ai-setup --help
```

## Tests

- Tests are located in `packages/*/src/__tests__/`.
- Run with `pnpm run test` (Vitest).
- Minimum coverage expectations are enforced in CI.

## Code style

- **TypeScript**: Biome linting and formatting
- **Go**: standard `gofmt`
- Commit messages should follow the project commit pattern configured in `.ai-setup.json`

## CI

- `.github/workflows/ci.yml` — Node/pnpm CI
- `.github/workflows/go-ci.yml` — Go CI
- `.github/workflows/docs.yml` — MkDocs build and GitHub Pages deploy

## Opening issues and PRs

- Use the provided issue templates when available
- Ensure `pnpm run lint` and `pnpm run test` pass before submitting
- For large changes, open a discussion or draft PR first

## Security

- Never commit `.env` or files containing secrets
- Follow the agent-security and security rules in `library/rules/`
- Report security concerns privately to the maintainers
